package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type CustomsRuleService struct {
	packRepo    *LogisticsRepo.CustomsRulePackRepository
	versionRepo *LogisticsRepo.CustomsRuleVersionRepository
}

type UpsertCustomsRulePackRequest struct {
	ID               string         `json:"id,omitempty"`
	Name             string         `json:"name"`
	CountryCode      string         `json:"country_code"`
	Status           string         `json:"status,omitempty"`
	Strategy         string         `json:"strategy,omitempty"`
	DefaultRiskLevel string         `json:"default_risk_level,omitempty"`
	Description      string         `json:"description,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

type PublishCustomsRuleVersionRequest struct {
	PackID      string                  `json:"pack_id"`
	VersionNo   int                     `json:"version_no,omitempty"`
	Status      string                  `json:"status,omitempty"`
	HitStrategy string                  `json:"hit_strategy,omitempty"`
	Rules       []CustomsRuleDefinition `json:"rules"`
	RiskConfig  map[string]any          `json:"risk_config,omitempty"`
}

type CustomsRuleDefinition struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Field      string `json:"field"`
	Operator   string `json:"operator"`
	Value      any    `json:"value"`
	RiskLevel  string `json:"risk_level"`
	Suggestion string `json:"suggestion"`
	Enabled    bool   `json:"enabled"`
}

type CustomsPrecheckRequest struct {
	PackID        string         `json:"pack_id,omitempty"`
	CountryCode   string         `json:"country_code,omitempty"`
	WaybillNo     string         `json:"waybill_no,omitempty"`
	DeclaredValue float64        `json:"declared_value,omitempty"`
	TaxNo         string         `json:"tax_no,omitempty"`
	HSCode        string         `json:"hs_code,omitempty"`
	DocumentType  string         `json:"document_type,omitempty"`
	DocumentCount int            `json:"document_count,omitempty"`
	ManualRelease bool           `json:"manual_release,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
}

type CustomsPrecheckMatchedRule struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	RiskLevel  string `json:"risk_level"`
	Suggestion string `json:"suggestion"`
	Reason     string `json:"reason"`
}

type CustomsPrecheckResult struct {
	PackID        string                       `json:"pack_id"`
	VersionID     string                       `json:"version_id"`
	VersionNo     int                          `json:"version_no"`
	CountryCode   string                       `json:"country_code"`
	Decision      string                       `json:"decision"`
	RiskLevel     string                       `json:"risk_level"`
	Suggestion    string                       `json:"suggestion"`
	MatchedRules  []CustomsPrecheckMatchedRule `json:"matched_rules"`
	ManualRelease bool                         `json:"manual_release"`
}

func NewCustomsRuleService(deps *app.Deps) *CustomsRuleService {
	if deps == nil || deps.DB == nil {
		return &CustomsRuleService{}
	}
	return &CustomsRuleService{
		packRepo:    LogisticsRepo.NewCustomsRulePackRepository(deps.DB),
		versionRepo: LogisticsRepo.NewCustomsRuleVersionRepository(deps.DB),
	}
}

func (s *CustomsRuleService) ListPacks(ctx context.Context, tenantUUID, countryCode, status string, limit int) ([]LogisticsModel.CustomsRulePack, error) {
	if s == nil || s.packRepo == nil {
		return nil, errors.New("customs rule service unavailable")
	}
	return s.packRepo.List(withTenantContext(ctx, tenantUUID), countryCode, normalizeCustomsPackStatus(status), limit)
}

func (s *CustomsRuleService) UpsertPack(ctx context.Context, tenantUUID string, req UpsertCustomsRulePackRequest) (*LogisticsModel.CustomsRulePack, error) {
	if s == nil || s.packRepo == nil {
		return nil, errors.New("customs rule service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	countryCode := strings.ToUpper(strings.TrimSpace(req.CountryCode))
	if countryCode == "" {
		return nil, errors.New("country_code is required")
	}
	status := normalizeCustomsPackStatus(req.Status)
	if status == "" {
		status = "draft"
	}
	strategy := normalizeCustomsHitStrategy(req.Strategy)
	if strategy == "" {
		strategy = "first_hit"
	}
	riskLevel := normalizeCustomsRiskLevel(req.DefaultRiskLevel)
	if riskLevel == "" {
		riskLevel = "low"
	}
	meta, err := jsonBytes(req.Metadata, []byte("{}"))
	if err != nil {
		return nil, err
	}

	row := &LogisticsModel.CustomsRulePack{ID: strings.TrimSpace(req.ID)}
	if row.ID == "" {
		row.ID = utils.NewUUID()
	}
	row.Name = name
	row.CountryCode = countryCode
	row.Status = status
	row.Strategy = strategy
	row.DefaultRiskLevel = riskLevel
	row.Description = strings.TrimSpace(req.Description)
	row.Metadata = datatypes.JSON(meta)
	if err := s.packRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *CustomsRuleService) ListVersions(ctx context.Context, tenantUUID, packID string, limit int) ([]LogisticsModel.CustomsRuleVersion, error) {
	if s == nil || s.versionRepo == nil {
		return nil, errors.New("customs rule service unavailable")
	}
	if strings.TrimSpace(packID) == "" {
		return nil, errors.New("pack_id is required")
	}
	return s.versionRepo.ListByPackID(withTenantContext(ctx, tenantUUID), packID, limit)
}

func (s *CustomsRuleService) PublishVersion(ctx context.Context, tenantUUID string, req PublishCustomsRuleVersionRequest) (*LogisticsModel.CustomsRuleVersion, error) {
	if s == nil || s.packRepo == nil || s.versionRepo == nil {
		return nil, errors.New("customs rule service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	packID := strings.TrimSpace(req.PackID)
	if packID == "" {
		return nil, errors.New("pack_id is required")
	}
	if _, err := s.packRepo.GetByID(ctx, packID); err != nil {
		return nil, err
	}
	if len(req.Rules) == 0 {
		return nil, errors.New("rules is required")
	}
	versionNo := req.VersionNo
	if versionNo <= 0 {
		latest, err := s.versionRepo.GetLatestByPackID(ctx, packID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if latest != nil {
			versionNo = latest.VersionNo + 1
		} else {
			versionNo = 1
		}
	}
	status := normalizeCustomsVersionStatus(req.Status)
	if status == "" {
		status = "published"
	}
	hitStrategy := normalizeCustomsHitStrategy(req.HitStrategy)
	if hitStrategy == "" {
		hitStrategy = "first_hit"
	}
	for i := range req.Rules {
		if strings.TrimSpace(req.Rules[i].Code) == "" {
			req.Rules[i].Code = fmt.Sprintf("RULE-%d", i+1)
		}
		if !req.Rules[i].Enabled {
			continue
		}
		if strings.TrimSpace(req.Rules[i].Field) == "" || strings.TrimSpace(req.Rules[i].Operator) == "" {
			return nil, errors.New("enabled rule field/operator is required")
		}
		if normalizeCustomsRiskLevel(req.Rules[i].RiskLevel) == "" {
			req.Rules[i].RiskLevel = "medium"
		}
	}
	rulesJSON, err := jsonBytes(req.Rules, []byte("[]"))
	if err != nil {
		return nil, err
	}
	riskJSON, err := jsonBytes(req.RiskConfig, []byte("{}"))
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	row := &LogisticsModel.CustomsRuleVersion{
		ID:           utils.NewUUID(),
		PackID:       packID,
		VersionNo:    versionNo,
		Status:       status,
		Rules:        datatypes.JSON(rulesJSON),
		HitStrategy:  hitStrategy,
		RiskSnapshot: datatypes.JSON(riskJSON),
	}
	if status == "published" {
		row.PublishedAt = &now
	}
	if err := s.versionRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *CustomsRuleService) Precheck(ctx context.Context, tenantUUID string, req CustomsPrecheckRequest) (*CustomsPrecheckResult, error) {
	if s == nil || s.packRepo == nil || s.versionRepo == nil {
		return nil, errors.New("customs rule service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	pack, version, err := s.resolvePackAndVersion(ctx, req)
	if err != nil {
		return nil, err
	}

	var rules []CustomsRuleDefinition
	if err := decodeJSONInto(version.Rules, &rules); err != nil {
		return nil, err
	}
	matched := make([]CustomsPrecheckMatchedRule, 0)
	maxRisk := normalizeCustomsRiskLevel(pack.DefaultRiskLevel)
	if maxRisk == "" {
		maxRisk = "low"
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		hit, reason := evaluateCustomsRule(rule, req)
		if !hit {
			continue
		}
		risk := normalizeCustomsRiskLevel(rule.RiskLevel)
		if risk == "" {
			risk = "medium"
		}
		if customsRiskRank(risk) > customsRiskRank(maxRisk) {
			maxRisk = risk
		}
		matched = append(matched, CustomsPrecheckMatchedRule{
			Code:       strings.TrimSpace(rule.Code),
			Name:       strings.TrimSpace(rule.Name),
			RiskLevel:  risk,
			Suggestion: strings.TrimSpace(rule.Suggestion),
			Reason:     reason,
		})
		if strings.EqualFold(version.HitStrategy, "first_hit") {
			break
		}
	}

	decision := "pass"
	suggestion := "auto_release"
	if len(matched) > 0 {
		suggestion = matched[0].Suggestion
		switch maxRisk {
		case "critical", "high":
			decision = "block"
		case "medium":
			decision = "review"
		default:
			decision = "pass"
		}
	}
	if req.ManualRelease && decision == "block" {
		decision = "manual_release"
		suggestion = "manual_release_with_audit"
	}

	return &CustomsPrecheckResult{
		PackID:        pack.ID,
		VersionID:     version.ID,
		VersionNo:     version.VersionNo,
		CountryCode:   pack.CountryCode,
		Decision:      decision,
		RiskLevel:     maxRisk,
		Suggestion:    suggestion,
		MatchedRules:  matched,
		ManualRelease: req.ManualRelease,
	}, nil
}

func (s *CustomsRuleService) resolvePackAndVersion(ctx context.Context, req CustomsPrecheckRequest) (*LogisticsModel.CustomsRulePack, *LogisticsModel.CustomsRuleVersion, error) {
	packID := strings.TrimSpace(req.PackID)
	if packID != "" {
		pack, err := s.packRepo.GetByID(ctx, packID)
		if err != nil {
			return nil, nil, err
		}
		version, err := s.versionRepo.GetLatestPublishedByPackID(ctx, pack.ID)
		if err != nil {
			return nil, nil, err
		}
		return pack, version, nil
	}

	countryCode := strings.ToUpper(strings.TrimSpace(req.CountryCode))
	if countryCode == "" {
		return nil, nil, errors.New("pack_id or country_code is required")
	}
	packs, err := s.packRepo.List(ctx, countryCode, "active", 1)
	if err != nil {
		return nil, nil, err
	}
	if len(packs) == 0 {
		return nil, nil, errors.New("no active customs rule pack found")
	}
	version, err := s.versionRepo.GetLatestPublishedByPackID(ctx, packs[0].ID)
	if err != nil {
		return nil, nil, err
	}
	return &packs[0], version, nil
}

func evaluateCustomsRule(rule CustomsRuleDefinition, req CustomsPrecheckRequest) (bool, string) {
	field := strings.ToLower(strings.TrimSpace(rule.Field))
	op := strings.ToLower(strings.TrimSpace(rule.Operator))
	actual := readCustomsField(field, req)
	expected := strings.TrimSpace(fmt.Sprintf("%v", rule.Value))

	switch op {
	case "exists":
		hit := strings.TrimSpace(actual) != ""
		return hit, fmt.Sprintf("%s exists=%t", field, hit)
	case "missing":
		hit := strings.TrimSpace(actual) == ""
		return hit, fmt.Sprintf("%s missing=%t", field, hit)
	case "eq":
		hit := strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(expected))
		return hit, fmt.Sprintf("%s=%s", field, actual)
	case "neq":
		hit := !strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(expected))
		return hit, fmt.Sprintf("%s=%s", field, actual)
	case "contains":
		hit := strings.Contains(strings.ToLower(actual), strings.ToLower(expected))
		return hit, fmt.Sprintf("%s contains %s", field, expected)
	case "gt", "gte", "lt", "lte":
		left, leftErr := strconv.ParseFloat(strings.TrimSpace(actual), 64)
		right, rightErr := strconv.ParseFloat(strings.TrimSpace(expected), 64)
		if leftErr != nil || rightErr != nil {
			return false, fmt.Sprintf("%s numeric parse failed", field)
		}
		switch op {
		case "gt":
			return left > right, fmt.Sprintf("%s=%v", field, left)
		case "gte":
			return left >= right, fmt.Sprintf("%s=%v", field, left)
		case "lt":
			return left < right, fmt.Sprintf("%s=%v", field, left)
		default:
			return left <= right, fmt.Sprintf("%s=%v", field, left)
		}
	default:
		return false, "unsupported operator"
	}
}

func readCustomsField(field string, req CustomsPrecheckRequest) string {
	switch field {
	case "country_code":
		return strings.ToUpper(strings.TrimSpace(req.CountryCode))
	case "waybill_no":
		return strings.TrimSpace(req.WaybillNo)
	case "tax_no":
		return strings.TrimSpace(req.TaxNo)
	case "hs_code":
		return strings.TrimSpace(req.HSCode)
	case "document_type":
		return strings.TrimSpace(req.DocumentType)
	case "document_count":
		return strconv.Itoa(req.DocumentCount)
	case "declared_value":
		return fmt.Sprintf("%.2f", req.DeclaredValue)
	default:
		if req.Payload != nil {
			if v, ok := req.Payload[field]; ok {
				return strings.TrimSpace(fmt.Sprintf("%v", v))
			}
		}
		return ""
	}
}

func decodeJSONInto(data datatypes.JSON, target any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, target)
}

func normalizeCustomsPackStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "draft", "active", "disabled":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeCustomsVersionStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "draft", "published", "archived":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeCustomsHitStrategy(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "first_hit", "all_hit":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeCustomsRiskLevel(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "low", "medium", "high", "critical":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func customsRiskRank(level string) int {
	switch normalizeCustomsRiskLevel(level) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}
