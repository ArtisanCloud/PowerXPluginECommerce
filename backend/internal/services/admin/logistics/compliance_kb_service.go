package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ComplianceKBService struct {
	kbRepo      *LogisticsRepo.ComplianceKBVersionRepository
	packRepo    *LogisticsRepo.CustomsRulePackRepository
	versionRepo *LogisticsRepo.CustomsRuleVersionRepository
}

type ComplianceKBQuery struct {
	CountryCode string `json:"country_code,omitempty"`
	Status      string `json:"status,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

type SyncComplianceKBPolicyRequest struct {
	CountryCode     string `json:"country_code,omitempty"`
	PackID          string `json:"pack_id,omitempty"`
	SourceVersionID string `json:"source_version_id,omitempty"`
	PolicyVersion   string `json:"policy_version,omitempty"`
	EffectiveFrom   string `json:"effective_from,omitempty"`
	EffectiveTo     string `json:"effective_to,omitempty"`
	Notes           string `json:"notes,omitempty"`
	OperatorID      string `json:"operator_id,omitempty"`
}

type DiffComplianceKBPolicyRequest struct {
	BasePolicyID   string `json:"base_policy_id"`
	TargetPolicyID string `json:"target_policy_id"`
}

type ComplianceKBDiff struct {
	BasePolicyID      string   `json:"base_policy_id"`
	TargetPolicyID    string   `json:"target_policy_id"`
	AddedRuleCodes    []string `json:"added_rule_codes"`
	RemovedRuleCodes  []string `json:"removed_rule_codes"`
	ChangedRuleCodes  []string `json:"changed_rule_codes"`
	UnchangedRuleCode int      `json:"unchanged_rule_count"`
}

type PublishComplianceKBPolicyRequest struct {
	RolloutPercent int      `json:"rollout_percent,omitempty"`
	RolloutTenants []string `json:"rollout_tenants,omitempty"`
	EffectiveFrom  string   `json:"effective_from,omitempty"`
	EffectiveTo    string   `json:"effective_to,omitempty"`
	OperatorID     string   `json:"operator_id,omitempty"`
}

func NewComplianceKBService(deps *app.Deps) *ComplianceKBService {
	if deps == nil || deps.DB == nil {
		return &ComplianceKBService{}
	}
	return &ComplianceKBService{
		kbRepo:      LogisticsRepo.NewComplianceKBVersionRepository(deps.DB),
		packRepo:    LogisticsRepo.NewCustomsRulePackRepository(deps.DB),
		versionRepo: LogisticsRepo.NewCustomsRuleVersionRepository(deps.DB),
	}
}

func (s *ComplianceKBService) ListPolicies(ctx context.Context, tenantUUID string, query ComplianceKBQuery) ([]LogisticsModel.ComplianceKBVersion, error) {
	if s == nil || s.kbRepo == nil {
		return nil, errors.New("compliance kb service unavailable")
	}
	return s.kbRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.ComplianceKBVersionFilter{
		CountryCode: strings.TrimSpace(query.CountryCode),
		Status:      normalizeComplianceKBStatus(query.Status),
		Limit:       query.Limit,
	})
}

func (s *ComplianceKBService) SyncPolicy(ctx context.Context, tenantUUID string, req SyncComplianceKBPolicyRequest) (*LogisticsModel.ComplianceKBVersion, error) {
	if s == nil || s.kbRepo == nil || s.packRepo == nil || s.versionRepo == nil {
		return nil, errors.New("compliance kb service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	pack, version, err := s.resolveSyncSource(ctx, req)
	if err != nil {
		return nil, err
	}
	countryCode := strings.ToUpper(strings.TrimSpace(req.CountryCode))
	if countryCode == "" {
		countryCode = strings.ToUpper(strings.TrimSpace(pack.CountryCode))
	}
	if countryCode == "" {
		return nil, errors.New("country_code is required")
	}
	mapping := map[string]any{
		"country_code":      countryCode,
		"source_pack_id":    pack.ID,
		"source_pack_name":  pack.Name,
		"source_version_id": version.ID,
		"source_version_no": version.VersionNo,
		"rules":             decodeRulesAsMap(version.Rules),
	}
	mappingJSON, err := jsonBytes(mapping, []byte("{}"))
	if err != nil {
		return nil, err
	}
	effectiveFrom, effectiveTo, err := parseEffectiveWindow(req.EffectiveFrom, req.EffectiveTo)
	if err != nil {
		return nil, err
	}

	versionLabel := strings.TrimSpace(req.PolicyVersion)
	if versionLabel == "" {
		latest, lerr := s.kbRepo.GetLatestByCountry(ctx, countryCode)
		if lerr != nil && !errors.Is(lerr, gorm.ErrRecordNotFound) {
			return nil, lerr
		}
		if latest == nil {
			versionLabel = fmt.Sprintf("v%d", version.VersionNo)
		} else {
			versionLabel = nextCompliancePolicyVersion(latest.PolicyVersion)
		}
	}

	row := &LogisticsModel.ComplianceKBVersion{
		ID:                 utils.NewUUID(),
		CountryCode:        countryCode,
		PolicyVersion:      versionLabel,
		SourcePackID:       pack.ID,
		SourceVersionID:    version.ID,
		SourceVersionNo:    version.VersionNo,
		CountryRuleMapping: datatypes.JSON(mappingJSON),
		EffectiveFrom:      effectiveFrom,
		EffectiveTo:        effectiveTo,
		Status:             "draft",
		Notes:              strings.TrimSpace(req.Notes),
		PublishedBy:        strings.TrimSpace(req.OperatorID),
	}
	if err := s.kbRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ComplianceKBService) DiffPolicies(ctx context.Context, tenantUUID string, req DiffComplianceKBPolicyRequest) (*ComplianceKBDiff, error) {
	if s == nil || s.kbRepo == nil {
		return nil, errors.New("compliance kb service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	base, err := s.kbRepo.GetByID(ctx, strings.TrimSpace(req.BasePolicyID))
	if err != nil {
		return nil, err
	}
	target, err := s.kbRepo.GetByID(ctx, strings.TrimSpace(req.TargetPolicyID))
	if err != nil {
		return nil, err
	}
	baseRules := extractComplianceRuleMap(base.CountryRuleMapping)
	targetRules := extractComplianceRuleMap(target.CountryRuleMapping)

	added := make([]string, 0)
	removed := make([]string, 0)
	changed := make([]string, 0)
	unchanged := 0

	for code, targetRaw := range targetRules {
		baseRaw, ok := baseRules[code]
		if !ok {
			added = append(added, code)
			continue
		}
		if normalizeJSON(baseRaw) != normalizeJSON(targetRaw) {
			changed = append(changed, code)
		} else {
			unchanged++
		}
	}
	for code := range baseRules {
		if _, ok := targetRules[code]; !ok {
			removed = append(removed, code)
		}
	}

	return &ComplianceKBDiff{
		BasePolicyID:      base.ID,
		TargetPolicyID:    target.ID,
		AddedRuleCodes:    added,
		RemovedRuleCodes:  removed,
		ChangedRuleCodes:  changed,
		UnchangedRuleCode: unchanged,
	}, nil
}

func (s *ComplianceKBService) PublishPolicy(ctx context.Context, tenantUUID, policyID string, req PublishComplianceKBPolicyRequest) (*LogisticsModel.ComplianceKBVersion, error) {
	if s == nil || s.kbRepo == nil {
		return nil, errors.New("compliance kb service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.kbRepo.GetByID(ctx, strings.TrimSpace(policyID))
	if err != nil {
		return nil, err
	}
	rolloutPercent := req.RolloutPercent
	if rolloutPercent <= 0 {
		rolloutPercent = 100
	}
	if rolloutPercent < 1 || rolloutPercent > 100 {
		return nil, errors.New("rollout_percent must be within 1~100")
	}
	effectiveFrom, effectiveTo, err := parseEffectiveWindow(req.EffectiveFrom, req.EffectiveTo)
	if err != nil {
		return nil, err
	}
	if effectiveFrom != nil {
		row.EffectiveFrom = effectiveFrom
	}
	if req.EffectiveTo != "" || effectiveTo != nil {
		row.EffectiveTo = effectiveTo
	}
	scopeJSON, err := jsonBytes(map[string]any{
		"rollout_percent": rolloutPercent,
		"rollout_tenants": sanitizeRolloutTenants(req.RolloutTenants),
	}, []byte("{}"))
	if err != nil {
		return nil, err
	}
	row.RolloutScope = datatypes.JSON(scopeJSON)
	if op := strings.TrimSpace(req.OperatorID); op != "" {
		row.PublishedBy = op
	}
	now := time.Now().UTC()
	row.PublishedAt = &now
	if rolloutPercent < 100 {
		row.Status = "grayscale"
		if err := s.kbRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		return row, nil
	}

	rows, err := s.kbRepo.ListByCountry(ctx, row.CountryCode, 200)
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		if item.ID == row.ID {
			continue
		}
		if item.Status == "active" || item.Status == "grayscale" {
			item.Status = "superseded"
			if err := s.kbRepo.Save(ctx, &item); err != nil {
				return nil, err
			}
		}
	}
	row.Status = "active"
	if err := s.kbRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ComplianceKBService) resolveSyncSource(ctx context.Context, req SyncComplianceKBPolicyRequest) (*LogisticsModel.CustomsRulePack, *LogisticsModel.CustomsRuleVersion, error) {
	packID := strings.TrimSpace(req.PackID)
	if packID == "" {
		if strings.TrimSpace(req.CountryCode) == "" {
			return nil, nil, errors.New("pack_id or country_code is required")
		}
		packs, err := s.packRepo.List(ctx, strings.TrimSpace(req.CountryCode), "active", 1)
		if err != nil {
			return nil, nil, err
		}
		if len(packs) == 0 {
			return nil, nil, gorm.ErrRecordNotFound
		}
		packID = packs[0].ID
	}
	pack, err := s.packRepo.GetByID(ctx, packID)
	if err != nil {
		return nil, nil, err
	}
	versionID := strings.TrimSpace(req.SourceVersionID)
	if versionID != "" {
		version, err := s.versionRepo.GetByID(ctx, versionID)
		return pack, version, err
	}
	version, err := s.versionRepo.GetLatestPublishedByPackID(ctx, pack.ID)
	if err != nil {
		return nil, nil, err
	}
	return pack, version, nil
}

func normalizeComplianceKBStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "draft", "grayscale", "active", "superseded", "inactive":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func parseEffectiveWindow(from, to string) (*time.Time, *time.Time, error) {
	var fromPtr *time.Time
	var toPtr *time.Time
	if strings.TrimSpace(from) != "" {
		v, err := time.Parse(time.RFC3339, strings.TrimSpace(from))
		if err != nil {
			return nil, nil, errors.New("effective_from must be RFC3339")
		}
		v = v.UTC()
		fromPtr = &v
	}
	if strings.TrimSpace(to) != "" {
		v, err := time.Parse(time.RFC3339, strings.TrimSpace(to))
		if err != nil {
			return nil, nil, errors.New("effective_to must be RFC3339")
		}
		v = v.UTC()
		toPtr = &v
	}
	if fromPtr != nil && toPtr != nil && toPtr.Before(*fromPtr) {
		return nil, nil, errors.New("effective_to must be later than effective_from")
	}
	return fromPtr, toPtr, nil
}

func decodeRulesAsMap(raw datatypes.JSON) map[string]any {
	out := map[string]any{}
	var rules []map[string]any
	if err := json.Unmarshal(raw, &rules); err != nil {
		return out
	}
	for _, rule := range rules {
		code := strings.TrimSpace(asString(rule["code"]))
		if code == "" {
			continue
		}
		out[code] = rule
	}
	return out
}

func extractComplianceRuleMap(raw datatypes.JSON) map[string]any {
	wrap := map[string]any{}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return map[string]any{}
	}
	rules, _ := wrap["rules"].(map[string]any)
	if rules == nil {
		return map[string]any{}
	}
	return rules
}

func normalizeJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func sanitizeRolloutTenants(input []string) []string {
	out := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, item := range input {
		v := strings.TrimSpace(item)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func nextCompliancePolicyVersion(prev string) string {
	prev = strings.TrimSpace(prev)
	if prev == "" {
		return "v1"
	}
	if strings.HasPrefix(strings.ToLower(prev), "v") {
		numeric := strings.TrimPrefix(strings.ToLower(prev), "v")
		var n int
		if _, err := fmt.Sscanf(numeric, "%d", &n); err == nil && n > 0 {
			return fmt.Sprintf("v%d", n+1)
		}
	}
	return prev + "-next"
}

func asString(v any) string {
	switch vv := v.(type) {
	case string:
		return vv
	default:
		data, _ := json.Marshal(vv)
		return string(data)
	}
}
