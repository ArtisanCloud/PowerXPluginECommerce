package logistics

import (
	"context"
	"errors"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AddressValidationService struct {
	repo *LogisticsRepo.AddressValidationRepository
}

type CheckAddressRequest struct {
	RequestKey string         `json:"request_key,omitempty"`
	WaybillID  string         `json:"waybill_id,omitempty"`
	WaybillNo  string         `json:"waybill_no,omitempty"`
	Address    string         `json:"address"`
	Context    map[string]any `json:"context,omitempty"`
}

func NewAddressValidationService(deps *app.Deps) *AddressValidationService {
	if deps == nil || deps.DB == nil {
		return &AddressValidationService{}
	}
	return &AddressValidationService{
		repo: LogisticsRepo.NewAddressValidationRepository(deps.DB),
	}
}

func (s *AddressValidationService) Check(ctx context.Context, tenantUUID string, req CheckAddressRequest) (*LogisticsModel.AddressValidation, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("address validation service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	address := strings.TrimSpace(req.Address)
	if address == "" {
		return nil, errors.New("address required")
	}
	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = "addr#" + strings.ToLower(strings.ReplaceAll(address, " ", ""))
	}
	if cached, err := s.repo.GetByRequestKey(ctx, requestKey); err == nil && cached != nil {
		return cached, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	normalized := normalizeAddress(address)
	reachable := true
	risk := "low"
	suggestion := ""
	needManual := false
	lower := strings.ToLower(normalized)
	switch {
	case strings.Contains(lower, "测试"), strings.Contains(lower, "test"):
		reachable = false
		risk = "high"
		suggestion = "请确认收货地址，避免测试地址导致妥投失败"
		needManual = true
	case strings.Contains(lower, "未知"), strings.Contains(lower, "unknown"):
		reachable = false
		risk = "medium"
		suggestion = "建议补充门牌号或联系电话"
		needManual = true
	case len(normalized) < 8:
		risk = "medium"
		suggestion = "地址信息偏短，建议补充详细门牌"
	}
	meta, _ := jsonBytes(req.Context, []byte("{}"))
	row := &LogisticsModel.AddressValidation{
		ID:               utils.NewUUID(),
		RequestKey:       requestKey,
		WaybillID:        strings.TrimSpace(req.WaybillID),
		WaybillNo:        strings.TrimSpace(req.WaybillNo),
		RawAddress:       address,
		Normalized:       normalized,
		Reachable:        reachable,
		RiskLevel:        risk,
		Suggestion:       suggestion,
		NeedManualReview: needManual,
		Metadata:         datatypes.JSON(meta),
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *AddressValidationService) List(ctx context.Context, tenantUUID, waybillNo string, limit int) ([]LogisticsModel.AddressValidation, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("address validation service unavailable")
	}
	return s.repo.List(withTenantContext(ctx, tenantUUID), strings.TrimSpace(waybillNo), limit)
}

func normalizeAddress(address string) string {
	normalized := strings.TrimSpace(address)
	normalized = strings.ReplaceAll(normalized, "，", ",")
	normalized = strings.ReplaceAll(normalized, "。", ".")
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}
