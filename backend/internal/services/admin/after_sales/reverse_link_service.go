package after_sales

import (
	"context"
	"errors"
	"strings"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	reversemodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/reverse"
	AfterSalesRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/after_sales"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	afterSalesObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrReverseLinkUnavailable      = errors.New("after-sales reverse link service unavailable")
	ErrReverseLinkNotAllowed       = errors.New("reverse logistics is only allowed for return_refund or exchange case")
	ErrReverseWaybillRequired      = errors.New("reverse waybill id or waybill no is required")
	ErrReverseWaybillNotFound      = errors.New("reverse waybill not found")
	ErrReverseWaybillOrderMismatch = errors.New("reverse waybill does not belong to the same order")
)

type ReverseLogisticsLinkInput struct {
	ReverseWaybillID string
	ReverseWaybillNo string
	CarrierCode      string
	Note             string
}

type ReverseLogisticsLinkDTO struct {
	CaseID           string     `json:"caseId"`
	CaseNo           string     `json:"caseNo"`
	ReverseWaybillID string     `json:"reverseWaybillId,omitempty"`
	ReverseWaybillNo string     `json:"reverseWaybillNo,omitempty"`
	CarrierCode      string     `json:"carrierCode,omitempty"`
	ReceiveStatus    string     `json:"receiveStatus"`
	ReceivedAt       *time.Time `json:"receivedAt,omitempty"`
	LinkedAt         time.Time  `json:"linkedAt"`
}

type ReverseLinkService struct {
	deps     *app.Deps
	caseRepo *AfterSalesRepo.CaseRepository
	linkRepo *AfterSalesRepo.ReverseLogisticsLinkRepository
	emitter  *afterSalesObs.Emitter
}

func NewReverseLinkService(deps *app.Deps) *ReverseLinkService {
	if deps == nil || deps.DB == nil {
		return &ReverseLinkService{deps: deps}
	}
	return &ReverseLinkService{
		deps:     deps,
		caseRepo: AfterSalesRepo.NewCaseRepository(deps.DB),
		linkRepo: AfterSalesRepo.NewReverseLogisticsLinkRepository(deps.DB),
		emitter:  afterSalesObs.NewEmitter(deps.RuntimeLogger(deps.Ctx, "after-sales", nil)),
	}
}

func (s *ReverseLinkService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.caseRepo != nil && s.linkRepo != nil
}

func (s *ReverseLinkService) Link(ctx context.Context, tenantUUID, caseID, operatorID string, req ReverseLogisticsLinkInput) (*ReverseLogisticsLinkDTO, error) {
	if !s.Ready() {
		return nil, ErrReverseLinkUnavailable
	}
	ctx = withTenantContext(ctx, tenantUUID)
	tenantUUID = strings.TrimSpace(tenantUUID)
	caseID = strings.TrimSpace(caseID)
	req.ReverseWaybillID = strings.TrimSpace(req.ReverseWaybillID)
	req.ReverseWaybillNo = strings.TrimSpace(req.ReverseWaybillNo)
	req.CarrierCode = strings.TrimSpace(req.CarrierCode)
	req.Note = strings.TrimSpace(req.Note)

	if req.ReverseWaybillID == "" && req.ReverseWaybillNo == "" {
		return nil, ErrReverseWaybillRequired
	}

	var linked *AfterSalesModel.ReturnLogisticsLink
	var caseRow *AfterSalesModel.AfterSaleCase
	err := s.caseRepo.WithTenantTx(ctx, func(tx *gorm.DB) error {
		query := tx.WithContext(ctx)
		if query.Dialector != nil && query.Dialector.Name() != "sqlite" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var current AfterSalesModel.AfterSaleCase
		if err := query.Where("tenant_uuid = ? AND id = ?", tenantUUID, caseID).First(&current).Error; err != nil {
			return err
		}
		if !IsCaseTypeAllowsReverseLink(current.CaseType) {
			return ErrReverseLinkNotAllowed
		}
		caseRow = &current

		waybill, err := s.findWaybill(ctx, tx, tenantUUID, req)
		if err != nil {
			return err
		}
		if strings.TrimSpace(waybill.OrderID) != strings.TrimSpace(current.OrderID) {
			return ErrReverseWaybillOrderMismatch
		}

		now := time.Now().UTC()
		receiveStatus := normalizeReceiveStatus(waybill.Status)
		var receivedAt *time.Time
		if receiveStatus == "received" {
			receivedAt = &now
		}

		linkTable := models.S(models.TableAfterSalesReverseLogisticsLinks)
		var existing AfterSalesModel.ReturnLogisticsLink
		err = tx.WithContext(ctx).
			Table(linkTable).
			Where("tenant_uuid = ? AND case_id = ?", tenantUUID, current.ID).
			Order("created_at DESC").
			First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			existing = AfterSalesModel.ReturnLogisticsLink{
				ID:               utils.NewUUID(),
				TenantUUID:       tenantUUID,
				CaseID:           current.ID,
				ReverseWaybillID: waybill.ID,
				ReverseWaybillNo: strings.TrimSpace(waybill.WaybillNo),
				CarrierCode:      req.CarrierCode,
				ReceiveStatus:    receiveStatus,
				ReceivedAt:       receivedAt,
			}
			if err := tx.WithContext(ctx).Create(&existing).Error; err != nil {
				return err
			}
		} else {
			existing.ReverseWaybillID = waybill.ID
			existing.ReverseWaybillNo = strings.TrimSpace(waybill.WaybillNo)
			if req.CarrierCode != "" {
				existing.CarrierCode = req.CarrierCode
			}
			existing.ReceiveStatus = receiveStatus
			existing.ReceivedAt = receivedAt
			if err := tx.WithContext(ctx).
				Model(&AfterSalesModel.ReturnLogisticsLink{}).
				Where("tenant_uuid = ? AND id = ?", tenantUUID, existing.ID).
				Updates(map[string]any{
					"reverse_waybill_id": waybill.ID,
					"reverse_waybill_no": existing.ReverseWaybillNo,
					"carrier_code":       existing.CarrierCode,
					"receive_status":     existing.ReceiveStatus,
					"received_at":        existing.ReceivedAt,
				}).Error; err != nil {
				return err
			}
		}
		if err := tx.WithContext(ctx).Create(&AfterSalesModel.AfterSaleTimeline{
			TenantUUID:   tenantUUID,
			CaseID:       current.ID,
			Action:       "link_reverse_logistics",
			FromStatus:   current.Status,
			ToStatus:     current.Status,
			OperatorType: "operator",
			OperatorID:   strings.TrimSpace(operatorID),
			Note:         req.Note,
		}).Error; err != nil {
			return err
		}
		linked = &existing
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminCaseNotFound
		}
		return nil, err
	}
	if linked == nil || caseRow == nil {
		return nil, gorm.ErrInvalidData
	}
	if s.emitter != nil {
		reqID, _ := authx.RequestIDFromContext(ctx)
		s.emitter.Emit(afterSalesObs.Event{
			Action:     "after_sales.reverse_logistics.linked",
			TenantUUID: strings.TrimSpace(tenantUUID),
			RequestID:  reqID,
			OperatorID: strings.TrimSpace(operatorID),
			CaseID:     strings.TrimSpace(caseRow.ID),
			CaseNo:     strings.TrimSpace(caseRow.CaseNo),
			CaseType:   strings.TrimSpace(caseRow.CaseType),
			OrderID:    strings.TrimSpace(caseRow.OrderID),
			Status:     strings.TrimSpace(caseRow.Status),
			Result:     "success",
			Metadata: map[string]any{
				"reverse_waybill_id": linked.ReverseWaybillID,
				"reverse_waybill_no": linked.ReverseWaybillNo,
				"receive_status":     linked.ReceiveStatus,
			},
			EmittedAt: time.Now().UTC(),
		})
	}
	return &ReverseLogisticsLinkDTO{
		CaseID:           caseRow.ID,
		CaseNo:           caseRow.CaseNo,
		ReverseWaybillID: linked.ReverseWaybillID,
		ReverseWaybillNo: linked.ReverseWaybillNo,
		CarrierCode:      linked.CarrierCode,
		ReceiveStatus:    linked.ReceiveStatus,
		ReceivedAt:       linked.ReceivedAt,
		LinkedAt:         linked.CreatedAt,
	}, nil
}

func (s *ReverseLinkService) findWaybill(ctx context.Context, tx *gorm.DB, tenantUUID string, req ReverseLogisticsLinkInput) (*reversemodel.Waybill, error) {
	var row reversemodel.Waybill
	q := tx.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if req.ReverseWaybillID != "" {
		q = q.Where("id = ?", req.ReverseWaybillID)
	} else {
		q = q.Where("waybill_no = ?", req.ReverseWaybillNo)
	}
	if err := q.First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReverseWaybillNotFound
		}
		return nil, err
	}
	return &row, nil
}

func normalizeReceiveStatus(waybillStatus string) string {
	status := strings.TrimSpace(strings.ToLower(waybillStatus))
	if status == "received" || status == "closed" {
		return "received"
	}
	return "pending"
}
