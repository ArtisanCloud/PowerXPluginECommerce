package logistics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type ReconciliationService struct {
	batchRepo   *LogisticsRepo.ReconciliationBatchRepository
	recordRepo  *LogisticsRepo.ReconciliationRecordRepository
	caseRepo    *LogisticsRepo.ReconciliationCaseRepository
	billingRepo *LogisticsRepo.BillingRepository
}

type ReconciliationInputRecord struct {
	WaybillID     string  `json:"waybill_id,omitempty"`
	WaybillNo     string  `json:"waybill_no"`
	CarrierID     string  `json:"carrier_id,omitempty"`
	BillAmount    float64 `json:"bill_amount"`
	BankAmount    float64 `json:"bank_amount"`
	InvoiceAmount float64 `json:"invoice_amount"`
}

type CreateReconciliationBatchRequest struct {
	CarrierID string                      `json:"carrier_id,omitempty"`
	From      string                      `json:"from,omitempty"`
	To        string                      `json:"to,omitempty"`
	Records   []ReconciliationInputRecord `json:"records,omitempty"`
}

type HandleReconciliationCaseRequest struct {
	Action     string `json:"action"`
	OperatorID string `json:"operator_id,omitempty"`
	Note       string `json:"note,omitempty"`
}

func NewReconciliationService(deps *app.Deps) *ReconciliationService {
	if deps == nil || deps.DB == nil {
		return &ReconciliationService{}
	}
	wbRepo := LogisticsRepo.NewWaybillRepository(deps.DB)
	return &ReconciliationService{
		batchRepo:   LogisticsRepo.NewReconciliationBatchRepository(deps.DB),
		recordRepo:  LogisticsRepo.NewReconciliationRecordRepository(deps.DB),
		caseRepo:    LogisticsRepo.NewReconciliationCaseRepository(deps.DB),
		billingRepo: LogisticsRepo.NewBillingRepository(wbRepo),
	}
}

func (s *ReconciliationService) ListBatches(ctx context.Context, tenantUUID, carrierID, status string, limit int) ([]LogisticsModel.ReconciliationBatch, error) {
	if s == nil || s.batchRepo == nil {
		return nil, errors.New("reconciliation service unavailable")
	}
	return s.batchRepo.List(withTenantContext(ctx, tenantUUID), carrierID, normalizeReconciliationBatchStatus(status), limit)
}

func (s *ReconciliationService) ListRecords(ctx context.Context, tenantUUID, batchID, status string, limit int) ([]LogisticsModel.ReconciliationRecord, error) {
	if s == nil || s.recordRepo == nil {
		return nil, errors.New("reconciliation service unavailable")
	}
	return s.recordRepo.List(withTenantContext(ctx, tenantUUID), batchID, normalizeReconciliationRecordStatus(status), limit)
}

func (s *ReconciliationService) ListCases(ctx context.Context, tenantUUID, batchID, status string, limit int) ([]LogisticsModel.ReconciliationCase, error) {
	if s == nil || s.caseRepo == nil {
		return nil, errors.New("reconciliation service unavailable")
	}
	return s.caseRepo.List(withTenantContext(ctx, tenantUUID), batchID, normalizeReconciliationCaseStatus(status), limit)
}

func (s *ReconciliationService) CreateBatch(ctx context.Context, tenantUUID string, req CreateReconciliationBatchRequest) (*LogisticsModel.ReconciliationBatch, error) {
	if s == nil || s.batchRepo == nil || s.recordRepo == nil || s.caseRepo == nil || s.billingRepo == nil {
		return nil, errors.New("reconciliation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	records, err := s.resolveInputRecords(ctx, req)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	batch := &LogisticsModel.ReconciliationBatch{
		ID:                 utils.NewUUID(),
		BatchNo:            buildReconciliationBatchNo(),
		CarrierID:          strings.TrimSpace(req.CarrierID),
		Status:             "open",
		Summary:            datatypes.JSON([]byte("{}")),
		Metadata:           datatypes.JSON([]byte("{}")),
		ExecutedAt:         &now,
		RecordCount:        len(records),
		MatchedCount:       0,
		ExceptionCount:     0,
		TotalBillAmount:    0,
		TotalBankAmount:    0,
		TotalInvoiceAmount: 0,
	}

	recordRows := make([]LogisticsModel.ReconciliationRecord, 0, len(records))
	caseRows := make([]LogisticsModel.ReconciliationCase, 0)
	summary := map[string]int{}

	for _, item := range records {
		batch.TotalBillAmount += item.BillAmount
		batch.TotalBankAmount += item.BankAmount
		batch.TotalInvoiceAmount += item.InvoiceAmount

		matchType, suggestion, status := evaluateReconciliation(item.BillAmount, item.BankAmount, item.InvoiceAmount)
		summary[matchType]++
		if status == "resolved" {
			batch.MatchedCount++
		} else {
			batch.ExceptionCount++
		}
		recordID := utils.NewUUID()
		caseID := ""
		if status == "pending" {
			caseID = utils.NewUUID()
			caseRows = append(caseRows, LogisticsModel.ReconciliationCase{
				ID:         caseID,
				CaseNo:     buildReconciliationCaseNo(),
				BatchID:    batch.ID,
				RecordID:   recordID,
				CarrierID:  strings.TrimSpace(item.CarrierID),
				WaybillNo:  strings.TrimSpace(item.WaybillNo),
				Status:     "open",
				Reason:     matchType,
				Suggestion: suggestion,
				Metadata:   datatypes.JSON([]byte("{}")),
			})
		}

		recordRows = append(recordRows, LogisticsModel.ReconciliationRecord{
			ID:            recordID,
			BatchID:       batch.ID,
			WaybillID:     strings.TrimSpace(item.WaybillID),
			WaybillNo:     strings.TrimSpace(item.WaybillNo),
			CarrierID:     strings.TrimSpace(item.CarrierID),
			BillAmount:    item.BillAmount,
			BankAmount:    item.BankAmount,
			InvoiceAmount: item.InvoiceAmount,
			DiffAmount:    roundReconciliation2(item.BillAmount - item.BankAmount),
			MatchType:     matchType,
			Suggestion:    suggestion,
			Status:        status,
			CaseID:        caseID,
			Metadata:      datatypes.JSON([]byte("{}")),
		})
	}

	summaryJSON, _ := jsonBytes(summary, []byte("{}"))
	batch.Summary = datatypes.JSON(summaryJSON)
	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, err
	}
	if err := s.recordRepo.CreateBatch(ctx, recordRows); err != nil {
		return nil, err
	}
	if err := s.caseRepo.CreateBatch(ctx, caseRows); err != nil {
		return nil, err
	}
	return batch, nil
}

func (s *ReconciliationService) HandleCase(ctx context.Context, tenantUUID, caseID string, req HandleReconciliationCaseRequest) (*LogisticsModel.ReconciliationCase, error) {
	if s == nil || s.caseRepo == nil || s.recordRepo == nil {
		return nil, errors.New("reconciliation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	action := normalizeReconciliationCaseAction(req.Action)
	if action == "" {
		return nil, errors.New("action must be confirm, appeal, close or manual_review")
	}
	now := time.Now().UTC()
	row.HandledBy = strings.TrimSpace(req.OperatorID)
	row.HandledAt = &now
	row.ActionNote = strings.TrimSpace(req.Note)
	switch action {
	case "confirm":
		row.Status = "confirmed"
	case "appeal":
		row.Status = "appealed"
	case "close":
		row.Status = "closed"
	default:
		row.Status = "open"
	}
	if err := s.caseRepo.Save(ctx, row); err != nil {
		return nil, err
	}

	record, err := s.recordRepo.GetByID(ctx, row.RecordID)
	if err == nil && record != nil {
		record.HandledAction = action
		record.HandledBy = row.HandledBy
		record.HandledAt = &now
		record.HandledNote = row.ActionNote
		if row.Status == "confirmed" || row.Status == "closed" {
			record.Status = "resolved"
		} else {
			record.Status = "pending"
		}
		_ = s.recordRepo.Save(ctx, record)
	}
	return row, nil
}

func (s *ReconciliationService) resolveInputRecords(ctx context.Context, req CreateReconciliationBatchRequest) ([]ReconciliationInputRecord, error) {
	records := make([]ReconciliationInputRecord, 0)
	if len(req.Records) > 0 {
		for _, row := range req.Records {
			if strings.TrimSpace(row.WaybillNo) == "" {
				continue
			}
			records = append(records, ReconciliationInputRecord{
				WaybillID:     strings.TrimSpace(row.WaybillID),
				WaybillNo:     strings.TrimSpace(row.WaybillNo),
				CarrierID:     strings.TrimSpace(row.CarrierID),
				BillAmount:    row.BillAmount,
				BankAmount:    row.BankAmount,
				InvoiceAmount: row.InvoiceAmount,
			})
		}
		return records, nil
	}

	filter, err := parseBillingFilter(BillingQuery{CarrierID: req.CarrierID, From: req.From, To: req.To})
	if err != nil {
		return nil, err
	}
	items, err := s.billingRepo.ListWaybills(ctx, filter)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		invoiceAmount := item.ActualFee
		if !strings.EqualFold(item.BillingStatus, "settled") {
			invoiceAmount = 0
		}
		records = append(records, ReconciliationInputRecord{
			WaybillID:     item.WaybillID,
			WaybillNo:     item.WaybillNo,
			CarrierID:     item.CarrierID,
			BillAmount:    chooseAmount(item.ActualFee, item.EstimatedFee),
			BankAmount:    item.ActualFee,
			InvoiceAmount: invoiceAmount,
		})
	}
	if len(records) == 0 {
		records = []ReconciliationInputRecord{
			{WaybillNo: "MOCK-RCN-001", CarrierID: strings.TrimSpace(req.CarrierID), BillAmount: 120, BankAmount: 120, InvoiceAmount: 120},
			{WaybillNo: "MOCK-RCN-002", CarrierID: strings.TrimSpace(req.CarrierID), BillAmount: 90, BankAmount: 90, InvoiceAmount: 0},
		}
	}
	return records, nil
}

func chooseAmount(primary, fallback float64) float64 {
	if math.Abs(primary) > 0.0001 {
		return primary
	}
	return fallback
}

func evaluateReconciliation(bill, bank, invoice float64) (matchType, suggestion, status string) {
	if closeEnough(bill, bank, 0.01) && closeEnough(bill, invoice, 0.01) {
		return "matched", "auto_archive", "resolved"
	}
	if math.Abs(invoice) <= 0.0001 && math.Abs(bank) > 0.0001 {
		return "missing_invoice", "request_invoice", "pending"
	}
	if math.Abs(bank) <= 0.0001 && math.Abs(invoice) > 0.0001 {
		return "missing_bank_flow", "check_bank_flow", "pending"
	}
	if closeEnough(bill, bank, 1.0) && closeEnough(bill, invoice, 1.0) {
		return "amount_tolerance", "confirm", "pending"
	}
	return "amount_mismatch", "manual_review", "pending"
}

func closeEnough(left, right, tolerance float64) bool {
	return math.Abs(left-right) <= tolerance
}

func normalizeReconciliationBatchStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "open":
		return "open"
	case "confirmed":
		return "confirmed"
	default:
		return ""
	}
}

func normalizeReconciliationRecordStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pending":
		return "pending"
	case "resolved":
		return "resolved"
	default:
		return ""
	}
}

func normalizeReconciliationCaseStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "open":
		return "open"
	case "confirmed":
		return "confirmed"
	case "appealed":
		return "appealed"
	case "closed":
		return "closed"
	default:
		return ""
	}
}

func normalizeReconciliationCaseAction(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "confirm":
		return "confirm"
	case "appeal":
		return "appeal"
	case "close":
		return "close"
	case "manual_review":
		return "manual_review"
	default:
		return ""
	}
}

func roundReconciliation2(value float64) float64 {
	return math.Round(value*100) / 100
}

func buildReconciliationBatchNo() string {
	return fmt.Sprintf("RCN%s", time.Now().UTC().Format("20060102150405"))
}

func buildReconciliationCaseNo() string {
	return fmt.Sprintf("RC%s%s", time.Now().UTC().Format("20060102150405"), strings.ToUpper(utils.NewUUID()[:6]))
}
