package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CarrierProfileService struct {
	carrierRepo *LogisticsRepo.CarrierRepository
	profileRepo *LogisticsRepo.CarrierProfileRepository
}

type CarrierProfileQuery struct {
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type EvaluateCarrierProfilesRequest struct {
	CarrierID  string `json:"carrier_id,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
}

type ConfirmCarrierProfileRatingRequest struct {
	Rating     string `json:"rating"`
	OperatorID string `json:"operator_id,omitempty"`
}

type RetireCarrierProfileRequest struct {
	Reason     string `json:"reason,omitempty"`
	Force      bool   `json:"force,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
}

type RestoreCarrierProfileRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
}

func NewCarrierProfileService(deps *app.Deps) *CarrierProfileService {
	if deps == nil || deps.DB == nil {
		return &CarrierProfileService{}
	}
	return &CarrierProfileService{
		carrierRepo: LogisticsRepo.NewCarrierRepository(deps.DB),
		profileRepo: LogisticsRepo.NewCarrierProfileRepository(deps.DB),
	}
}

func (s *CarrierProfileService) List(ctx context.Context, tenantUUID string, query CarrierProfileQuery) ([]LogisticsModel.CarrierProfile, error) {
	if s == nil || s.profileRepo == nil {
		return nil, errors.New("carrier profile service unavailable")
	}
	return s.profileRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.CarrierProfileFilter{
		Status: normalizeCarrierProfileStatus(query.Status),
		Limit:  query.Limit,
	})
}

func (s *CarrierProfileService) Evaluate(ctx context.Context, tenantUUID string, req EvaluateCarrierProfilesRequest) ([]LogisticsModel.CarrierProfile, error) {
	if s == nil || s.carrierRepo == nil || s.profileRepo == nil {
		return nil, errors.New("carrier profile service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	targetCarrierID := strings.TrimSpace(req.CarrierID)

	carriers, err := s.carrierRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	if targetCarrierID != "" {
		filtered := make([]LogisticsModel.Carrier, 0, 1)
		for _, item := range carriers {
			if item.ID == targetCarrierID {
				filtered = append(filtered, item)
				break
			}
		}
		carriers = filtered
	}
	out := make([]LogisticsModel.CarrierProfile, 0, len(carriers))
	for _, carrier := range carriers {
		profile, err := s.profileRepo.GetByCarrierID(ctx, carrier.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) || profile == nil {
			profile = &LogisticsModel.CarrierProfile{
				ID:        utils.NewUUID(),
				CarrierID: carrier.ID,
				Status:    "active",
			}
		}
		config := mapFromJSON(carrier.Config)
		onTimeRate := numberFromRaw(config, "onTimeRate", numberFromRaw(config, "on_time_rate", 90))
		avgHours := numberFromRaw(config, "avgHours", numberFromRaw(config, "avg_hours", 48))
		costIndex := numberFromRaw(config, "costIndex", numberFromRaw(config, "cost_index", 1))
		if avgHours <= 0 {
			avgHours = 48
		}
		if costIndex <= 0 {
			costIndex = 1
		}

		stability := clampProfileScore(onTimeRate*0.72 + (100-math.Min(avgHours*1.1, 100))*0.28)
		costScore := clampProfileScore(100 - (costIndex-1)*35)
		composite := clampProfileScore(stability*0.6 + costScore*0.4)

		profile.StabilityScore = roundProfile2(stability)
		profile.CostScore = roundProfile2(costScore)
		profile.CompositeScore = roundProfile2(composite)
		profile.ServiceRating = ratingFromComposite(composite)
		profile.Suggestion = buildCarrierProfileSuggestion(profile.ServiceRating, profile.Status)
		if profile.Status == "" {
			profile.Status = "active"
		}
		trend := appendCarrierProfileTrend(sliceFromJSON(profile.ScoreTrend), profile.CompositeScore)
		trendJSON, _ := jsonBytes(trend, []byte("[]"))
		profile.ScoreTrend = datatypes.JSON(trendJSON)
		now := time.Now().UTC()
		profile.EvaluatedAt = &now
		if strings.TrimSpace(req.OperatorID) != "" {
			profile.ConfirmedBy = strings.TrimSpace(req.OperatorID)
		}
		if err := s.profileRepo.Save(ctx, profile); err != nil {
			return nil, err
		}
		out = append(out, *profile)
	}
	return out, nil
}

func (s *CarrierProfileService) ConfirmRating(ctx context.Context, tenantUUID, profileID string, req ConfirmCarrierProfileRatingRequest) (*LogisticsModel.CarrierProfile, error) {
	if s == nil || s.profileRepo == nil {
		return nil, errors.New("carrier profile service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	rating := strings.ToUpper(strings.TrimSpace(req.Rating))
	if !isAllowedCarrierRating(rating) {
		return nil, errors.New("rating must be A|B|C|D")
	}
	row.ServiceRating = rating
	if op := strings.TrimSpace(req.OperatorID); op != "" {
		row.ConfirmedBy = op
	}
	now := time.Now().UTC()
	row.ConfirmedAt = &now
	row.Suggestion = buildCarrierProfileSuggestion(row.ServiceRating, row.Status)
	if err := s.profileRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *CarrierProfileService) Retire(ctx context.Context, tenantUUID, profileID string, req RetireCarrierProfileRequest) (*LogisticsModel.CarrierProfile, error) {
	if s == nil || s.profileRepo == nil {
		return nil, errors.New("carrier profile service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(row.Status, "retired") {
		return nil, errors.New("carrier profile already retired")
	}
	if !req.Force && (strings.EqualFold(row.ServiceRating, "A") || strings.EqualFold(row.ServiceRating, "B")) {
		return nil, errors.New("high-rated carrier cannot be retired without force")
	}
	row.Status = "retired"
	row.RetireReason = strings.TrimSpace(req.Reason)
	if op := strings.TrimSpace(req.OperatorID); op != "" {
		row.ConfirmedBy = op
	}
	now := time.Now().UTC()
	row.ConfirmedAt = &now
	row.Suggestion = buildCarrierProfileSuggestion(row.ServiceRating, row.Status)
	if err := s.profileRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *CarrierProfileService) Restore(ctx context.Context, tenantUUID, profileID string, req RestoreCarrierProfileRequest) (*LogisticsModel.CarrierProfile, error) {
	if s == nil || s.profileRepo == nil {
		return nil, errors.New("carrier profile service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(row.Status, "retired") {
		return nil, errors.New("carrier profile is not retired")
	}
	row.Status = "active"
	row.RetireReason = ""
	if op := strings.TrimSpace(req.OperatorID); op != "" {
		row.ConfirmedBy = op
	}
	now := time.Now().UTC()
	row.ConfirmedAt = &now
	row.Suggestion = buildCarrierProfileSuggestion(row.ServiceRating, row.Status)
	if err := s.profileRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func normalizeCarrierProfileStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "active", "retired", "monitor":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func isAllowedCarrierRating(rating string) bool {
	switch strings.ToUpper(strings.TrimSpace(rating)) {
	case "A", "B", "C", "D":
		return true
	default:
		return false
	}
}

func ratingFromComposite(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	default:
		return "D"
	}
}

func clampProfileScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func roundProfile2(v float64) float64 {
	return math.Round(v*100) / 100
}

func buildCarrierProfileSuggestion(rating, status string) string {
	rating = strings.ToUpper(strings.TrimSpace(rating))
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "retired" {
		return "已淘汰，建议暂停新增分单并观察恢复窗口"
	}
	switch rating {
	case "A":
		return "建议提升分单权重，作为优先承运商"
	case "B":
		return "建议维持当前策略并持续观察波动"
	case "C":
		return "建议降低分单比例并制定整改计划"
	default:
		return "建议进入淘汰评估，必要时启用替代承运商"
	}
}

func appendCarrierProfileTrend(history []float64, score float64) []float64 {
	next := make([]float64, 0, len(history)+1)
	for _, item := range history {
		next = append(next, roundProfile2(item))
	}
	next = append(next, roundProfile2(score))
	if len(next) > 8 {
		next = next[len(next)-8:]
	}
	return next
}

func mapFromJSON(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}
	}
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

func sliceFromJSON(raw datatypes.JSON) []float64 {
	if len(raw) == 0 || string(raw) == "null" {
		return []float64{}
	}
	var out []float64
	_ = json.Unmarshal(raw, &out)
	return out
}

func numberFromRaw(data map[string]any, key string, fallback float64) float64 {
	if data == nil {
		return fallback
	}
	raw, ok := data[key]
	if !ok || raw == nil {
		return fallback
	}
	switch vv := raw.(type) {
	case float64:
		return vv
	case float32:
		return float64(vv)
	case int:
		return float64(vv)
	case int64:
		return float64(vv)
	case int32:
		return float64(vv)
	case uint:
		return float64(vv)
	case uint64:
		return float64(vv)
	case uint32:
		return float64(vv)
	case string:
		parsed := strings.TrimSpace(vv)
		if parsed == "" {
			return fallback
		}
		num, err := strconv.ParseFloat(parsed, 64)
		if err != nil {
			return fallback
		}
		return num
	default:
		return fallback
	}
}
