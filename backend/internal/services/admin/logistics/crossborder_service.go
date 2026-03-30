package logistics

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CrossborderService struct {
	docRepo   *LogisticsRepo.CrossborderDocumentRepository
	quoteRepo *LogisticsRepo.CrossborderTaxQuoteRepository
	mapRepo   *LogisticsRepo.CrossborderTrackingMapRepository
}

type UpsertCrossborderDocumentRequest struct {
	ID          string         `json:"id,omitempty"`
	WaybillID   string         `json:"waybill_id,omitempty"`
	WaybillNo   string         `json:"waybill_no"`
	DocType     string         `json:"doc_type"`
	DocNo       string         `json:"doc_no"`
	CountryFrom string         `json:"country_from,omitempty"`
	CountryTo   string         `json:"country_to,omitempty"`
	Status      string         `json:"status,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type QuoteCrossborderTaxRequest struct {
	RequestKey    string  `json:"request_key,omitempty"`
	WaybillID     string  `json:"waybill_id,omitempty"`
	WaybillNo     string  `json:"waybill_no,omitempty"`
	Destination   string  `json:"destination_country"`
	Currency      string  `json:"currency,omitempty"`
	DeclaredValue float64 `json:"declared_value"`
	ShippingFee   float64 `json:"shipping_fee,omitempty"`
	InsuranceFee  float64 `json:"insurance_fee,omitempty"`
}

type UpsertCrossborderTrackingMapRequest struct {
	ID               string         `json:"id,omitempty"`
	Provider         string         `json:"provider"`
	ProviderStatus   string         `json:"provider_status"`
	NormalizedStatus string         `json:"normalized_status"`
	Description      string         `json:"description,omitempty"`
	Priority         int            `json:"priority,omitempty"`
	Enabled          *bool          `json:"enabled,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

type NormalizeCrossborderTrackingRequest struct {
	Provider       string `json:"provider"`
	ProviderStatus string `json:"provider_status"`
}

func NewCrossborderService(deps *app.Deps) *CrossborderService {
	if deps == nil || deps.DB == nil {
		return &CrossborderService{}
	}
	return &CrossborderService{
		docRepo:   LogisticsRepo.NewCrossborderDocumentRepository(deps.DB),
		quoteRepo: LogisticsRepo.NewCrossborderTaxQuoteRepository(deps.DB),
		mapRepo:   LogisticsRepo.NewCrossborderTrackingMapRepository(deps.DB),
	}
}

func (s *CrossborderService) ListDocuments(ctx context.Context, tenantUUID, waybillNo string, limit int) ([]LogisticsModel.CrossborderDocument, error) {
	if s == nil || s.docRepo == nil {
		return nil, errors.New("crossborder service unavailable")
	}
	return s.docRepo.List(withTenantContext(ctx, tenantUUID), waybillNo, limit)
}

func (s *CrossborderService) UpsertDocument(ctx context.Context, tenantUUID string, req UpsertCrossborderDocumentRequest) (*LogisticsModel.CrossborderDocument, error) {
	if s == nil || s.docRepo == nil {
		return nil, errors.New("crossborder service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	waybillNo := strings.TrimSpace(req.WaybillNo)
	docType := normalizeCrossborderDocType(req.DocType)
	docNo := strings.TrimSpace(req.DocNo)
	if waybillNo == "" || docType == "" || docNo == "" {
		return nil, errors.New("waybill_no/doc_type/doc_no is required")
	}
	status := normalizeCrossborderDocStatus(req.Status)
	if status == "" {
		status = "pending"
	}
	meta, err := jsonBytes(req.Metadata, []byte("{}"))
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var row *LogisticsModel.CrossborderDocument
	if id := strings.TrimSpace(req.ID); id != "" {
		row = &LogisticsModel.CrossborderDocument{
			ID:        id,
			WaybillID: strings.TrimSpace(req.WaybillID),
		}
	} else {
		existing, getErr := s.docRepo.GetByUniq(ctx, waybillNo, docType)
		if getErr == nil && existing != nil {
			row = existing
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		}
	}
	if row == nil {
		row = &LogisticsModel.CrossborderDocument{
			ID:        utils.NewUUID(),
			WaybillID: strings.TrimSpace(req.WaybillID),
		}
	}
	row.WaybillNo = waybillNo
	row.DocType = docType
	row.DocNo = docNo
	row.CountryFrom = strings.ToUpper(strings.TrimSpace(req.CountryFrom))
	row.CountryTo = strings.ToUpper(strings.TrimSpace(req.CountryTo))
	row.Status = status
	row.Metadata = datatypes.JSON(meta)
	if status == "validated" {
		row.ValidatedAt = &now
	}
	if err := s.docRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *CrossborderService) QuoteTax(ctx context.Context, tenantUUID string, req QuoteCrossborderTaxRequest) (*LogisticsModel.CrossborderTaxQuote, string, error) {
	if s == nil || s.quoteRepo == nil {
		return nil, "", errors.New("crossborder service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if req.DeclaredValue < 0 || req.ShippingFee < 0 || req.InsuranceFee < 0 {
		return nil, "", errors.New("declared/shipping/insurance must be >= 0")
	}
	dest := strings.ToUpper(strings.TrimSpace(req.Destination))
	if dest == "" {
		return nil, "", errors.New("destination_country is required")
	}
	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = "crossborder-tax#" + utils.NewUUID()
	}
	existing, err := s.quoteRepo.GetByRequestKey(ctx, requestKey)
	if err == nil && existing != nil {
		return existing, "replayed", nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	dutyRate, vatRate, exemption := crossborderTaxPolicy(dest)
	base := req.DeclaredValue + req.ShippingFee + req.InsuranceFee - exemption
	if base < 0 {
		base = 0
	}
	base = round2Crossborder(base)
	duty := round2Crossborder(base * dutyRate)
	vat := round2Crossborder((base + duty) * vatRate)
	total := round2Crossborder(duty + vat)
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	row := &LogisticsModel.CrossborderTaxQuote{
		ID:               utils.NewUUID(),
		RequestKey:       requestKey,
		WaybillID:        strings.TrimSpace(req.WaybillID),
		WaybillNo:        strings.TrimSpace(req.WaybillNo),
		Destination:      dest,
		Currency:         currency,
		DeclaredValue:    round2Crossborder(req.DeclaredValue),
		ShippingFee:      round2Crossborder(req.ShippingFee),
		InsuranceFee:     round2Crossborder(req.InsuranceFee),
		ExemptionAmount:  round2Crossborder(exemption),
		DutyRate:         dutyRate,
		VatRate:          vatRate,
		DutyAmount:       duty,
		VatAmount:        vat,
		TotalTaxAmount:   total,
		QuoteProvider:    "rule-engine",
		NormalizedStatus: "estimated",
		Metadata:         datatypes.JSON([]byte("{}")),
	}
	if err := s.quoteRepo.Save(ctx, row); err != nil {
		return nil, "", err
	}
	return row, "created", nil
}

func (s *CrossborderService) ListTrackingMaps(ctx context.Context, tenantUUID, provider string, enabled *bool, limit int) ([]LogisticsModel.CrossborderTrackingMap, error) {
	if s == nil || s.mapRepo == nil {
		return nil, errors.New("crossborder service unavailable")
	}
	return s.mapRepo.List(withTenantContext(ctx, tenantUUID), strings.ToLower(strings.TrimSpace(provider)), enabled, limit)
}

func (s *CrossborderService) UpsertTrackingMap(ctx context.Context, tenantUUID string, req UpsertCrossborderTrackingMapRequest) (*LogisticsModel.CrossborderTrackingMap, error) {
	if s == nil || s.mapRepo == nil {
		return nil, errors.New("crossborder service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	providerStatus := strings.ToLower(strings.TrimSpace(req.ProviderStatus))
	normalized := normalizeCrossborderTrackingStatus(req.NormalizedStatus)
	if provider == "" || providerStatus == "" || normalized == "" {
		return nil, errors.New("provider/provider_status/normalized_status is required")
	}
	priority := req.Priority
	if priority <= 0 {
		priority = 100
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	meta, err := jsonBytes(req.Metadata, []byte("{}"))
	if err != nil {
		return nil, err
	}

	var row *LogisticsModel.CrossborderTrackingMap
	if id := strings.TrimSpace(req.ID); id != "" {
		row = &LogisticsModel.CrossborderTrackingMap{ID: id}
	} else {
		existing, getErr := s.mapRepo.GetByUniq(ctx, provider, providerStatus)
		if getErr == nil && existing != nil {
			row = existing
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		}
	}
	if row == nil {
		row = &LogisticsModel.CrossborderTrackingMap{ID: utils.NewUUID()}
	}
	row.Provider = provider
	row.ProviderStatus = providerStatus
	row.NormalizedStatus = normalized
	row.Description = strings.TrimSpace(req.Description)
	row.Priority = priority
	row.Enabled = enabled
	row.Metadata = datatypes.JSON(meta)
	if err := s.mapRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *CrossborderService) NormalizeTrackingStatus(ctx context.Context, tenantUUID string, req NormalizeCrossborderTrackingRequest) (string, string, error) {
	if s == nil || s.mapRepo == nil {
		return "", "", errors.New("crossborder service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	providerStatus := strings.ToLower(strings.TrimSpace(req.ProviderStatus))
	if provider == "" || providerStatus == "" {
		return "", "", errors.New("provider/provider_status is required")
	}
	row, err := s.mapRepo.GetByUniq(ctx, provider, providerStatus)
	if err == nil && row != nil && row.Enabled {
		return row.NormalizedStatus, "mapping_table", nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", err
	}
	switch providerStatus {
	case "pending_pickup", "label_created":
		return "created", "builtin", nil
	case "in_transit", "customs_clearing", "arrived_destination":
		return "in_transit", "builtin", nil
	case "exception", "customs_hold", "returning":
		return "exception", "builtin", nil
	case "delivered", "signed":
		return "delivered", "builtin", nil
	default:
		return "created", "fallback", nil
	}
}

func normalizeCrossborderDocType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "invoice", "customs_form", "certificate", "packing_list":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeCrossborderDocStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pending", "validated", "rejected":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeCrossborderTrackingStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "created", "in_transit", "exception", "delivered":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func crossborderTaxPolicy(destination string) (dutyRate float64, vatRate float64, exemption float64) {
	switch strings.ToUpper(strings.TrimSpace(destination)) {
	case "US":
		return 0.05, 0, 800
	case "EU":
		return 0.06, 0.20, 0
	case "JP":
		return 0.04, 0.10, 1000
	default:
		return 0.08, 0.13, 50
	}
}

func round2Crossborder(v float64) float64 {
	return math.Round(v*100) / 100
}
