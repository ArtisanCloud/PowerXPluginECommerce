package logistics

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type RiskService struct {
	ruleRepo      *LogisticsRepo.RiskRuleRepository
	blacklistRepo *LogisticsRepo.BlacklistRepository
	hitRepo       *LogisticsRepo.RiskHitRepository
}

func NewRiskService(deps *app.Deps) *RiskService {
	if deps == nil || deps.DB == nil {
		return &RiskService{}
	}
	return &RiskService{
		ruleRepo:      LogisticsRepo.NewRiskRuleRepository(deps.DB),
		blacklistRepo: LogisticsRepo.NewBlacklistRepository(deps.DB),
		hitRepo:       LogisticsRepo.NewRiskHitRepository(deps.DB),
	}
}

type UpsertRiskRuleRequest struct {
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name"`
	MatchField  string         `json:"match_field,omitempty"`
	MatchMode   string         `json:"match_mode,omitempty"`
	Pattern     string         `json:"pattern"`
	Decision    string         `json:"decision,omitempty"`
	RiskLevel   string         `json:"risk_level,omitempty"`
	Priority    int            `json:"priority,omitempty"`
	Enabled     *bool          `json:"enabled,omitempty"`
	Description string         `json:"description,omitempty"`
	RuleConfig  map[string]any `json:"rule_config,omitempty"`
}

type UpsertBlacklistEntryRequest struct {
	ID             string         `json:"id,omitempty"`
	EntryType      string         `json:"entry_type,omitempty"`
	RecipientName  string         `json:"recipient_name,omitempty"`
	RecipientPhone string         `json:"recipient_phone,omitempty"`
	AddressLine    string         `json:"address_line,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	Status         string         `json:"status,omitempty"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type EvaluateRiskRequest struct {
	WaybillID       string         `json:"waybill_id,omitempty"`
	WaybillNo       string         `json:"waybill_no,omitempty"`
	RecipientName   string         `json:"recipient_name,omitempty"`
	RecipientPhone  string         `json:"recipient_phone,omitempty"`
	DestinationLine string         `json:"destination_line,omitempty"`
	OperatorID      string         `json:"operator_id,omitempty"`
	Context         map[string]any `json:"context,omitempty"`
}

type RiskEvaluateResult struct {
	Blocked      bool                            `json:"blocked"`
	Decision     string                          `json:"decision"`
	Score        int                             `json:"score"`
	ReleasedBy   string                          `json:"released_by,omitempty"`
	Fingerprint  string                          `json:"fingerprint,omitempty"`
	MatchedRules []LogisticsModel.RiskRule       `json:"matched_rules,omitempty"`
	MatchedList  []LogisticsModel.BlacklistEntry `json:"matched_blacklist,omitempty"`
	Hits         []LogisticsModel.RiskHit        `json:"hits,omitempty"`
}

type ReleaseRiskHitRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func (s *RiskService) ListRules(ctx context.Context, tenantUUID string) ([]LogisticsModel.RiskRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("risk service unavailable")
	}
	return s.ruleRepo.List(withTenantContext(ctx, tenantUUID), false)
}

func (s *RiskService) UpsertRule(ctx context.Context, tenantUUID string, req UpsertRiskRuleRequest) (*LogisticsModel.RiskRule, error) {
	if s == nil || s.ruleRepo == nil {
		return nil, errors.New("risk service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name required")
	}
	pattern := strings.TrimSpace(req.Pattern)
	if pattern == "" {
		return nil, errors.New("pattern required")
	}
	matchField := firstNonEmpty(strings.ToLower(strings.TrimSpace(req.MatchField)), "address")
	if !isAllowed(matchField, "address", "recipient", "phone") {
		return nil, errors.New("match_field must be one of address|recipient|phone")
	}
	matchMode := firstNonEmpty(strings.ToLower(strings.TrimSpace(req.MatchMode)), "contains")
	if !isAllowed(matchMode, "contains", "exact", "prefix", "regex") {
		return nil, errors.New("match_mode must be one of contains|exact|prefix|regex")
	}
	decision := firstNonEmpty(strings.ToLower(strings.TrimSpace(req.Decision)), "review")
	if !isAllowed(decision, "review", "block") {
		return nil, errors.New("decision must be review|block")
	}
	priority := req.Priority
	if priority == 0 {
		priority = 100
	}
	level := firstNonEmpty(strings.ToLower(strings.TrimSpace(req.RiskLevel)), "medium")
	if !isAllowed(level, "low", "medium", "high") {
		return nil, errors.New("risk_level must be low|medium|high")
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	cfg, _ := jsonBytes(req.RuleConfig, []byte("{}"))

	id := strings.TrimSpace(req.ID)
	if id != "" {
		row, err := s.ruleRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		row.Name = name
		row.MatchField = matchField
		row.MatchMode = matchMode
		row.Pattern = pattern
		row.Decision = decision
		row.RiskLevel = level
		row.Priority = priority
		row.Enabled = enabled
		row.Description = strings.TrimSpace(req.Description)
		row.RuleConfig = datatypes.JSON(cfg)
		if err := s.ruleRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		return row, nil
	}

	row := &LogisticsModel.RiskRule{
		ID:          utils.NewUUID(),
		Name:        name,
		MatchField:  matchField,
		MatchMode:   matchMode,
		Pattern:     pattern,
		Decision:    decision,
		RiskLevel:   level,
		Priority:    priority,
		Enabled:     enabled,
		Description: strings.TrimSpace(req.Description),
		RuleConfig:  datatypes.JSON(cfg),
	}
	if err := s.ruleRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *RiskService) ListBlacklist(ctx context.Context, tenantUUID, status string) ([]LogisticsModel.BlacklistEntry, error) {
	if s == nil || s.blacklistRepo == nil {
		return nil, errors.New("risk service unavailable")
	}
	return s.blacklistRepo.List(withTenantContext(ctx, tenantUUID), status)
}

func (s *RiskService) UpsertBlacklist(ctx context.Context, tenantUUID string, req UpsertBlacklistEntryRequest) (*LogisticsModel.BlacklistEntry, error) {
	if s == nil || s.blacklistRepo == nil {
		return nil, errors.New("risk service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	entryType := firstNonEmpty(strings.ToLower(strings.TrimSpace(req.EntryType)), "recipient")
	if !isAllowed(entryType, "recipient", "address", "phone") {
		return nil, errors.New("entry_type must be recipient|address|phone")
	}
	status := firstNonEmpty(strings.ToLower(strings.TrimSpace(req.Status)), "active")
	if !isAllowed(status, "active", "inactive") {
		return nil, errors.New("status must be active|inactive")
	}
	if strings.TrimSpace(req.RecipientName) == "" && strings.TrimSpace(req.RecipientPhone) == "" && strings.TrimSpace(req.AddressLine) == "" {
		return nil, errors.New("recipient_name/recipient_phone/address_line at least one required")
	}
	meta, _ := jsonBytes(req.Metadata, []byte("{}"))
	id := strings.TrimSpace(req.ID)
	if id != "" {
		row, err := s.blacklistRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		row.EntryType = entryType
		row.RecipientName = strings.TrimSpace(req.RecipientName)
		row.RecipientPhone = strings.TrimSpace(req.RecipientPhone)
		row.AddressLine = strings.TrimSpace(req.AddressLine)
		row.Reason = strings.TrimSpace(req.Reason)
		row.Status = status
		row.ExpiresAt = req.ExpiresAt
		row.Metadata = datatypes.JSON(meta)
		if err := s.blacklistRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		return row, nil
	}
	row := &LogisticsModel.BlacklistEntry{
		ID:             utils.NewUUID(),
		EntryType:      entryType,
		RecipientName:  strings.TrimSpace(req.RecipientName),
		RecipientPhone: strings.TrimSpace(req.RecipientPhone),
		AddressLine:    strings.TrimSpace(req.AddressLine),
		Reason:         strings.TrimSpace(req.Reason),
		Status:         status,
		ExpiresAt:      req.ExpiresAt,
		Metadata:       datatypes.JSON(meta),
	}
	if err := s.blacklistRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *RiskService) ListHits(ctx context.Context, tenantUUID, waybillID, status string) ([]LogisticsModel.RiskHit, error) {
	if s == nil || s.hitRepo == nil {
		return nil, errors.New("risk service unavailable")
	}
	return s.hitRepo.List(withTenantContext(ctx, tenantUUID), waybillID, status)
}

func (s *RiskService) Evaluate(ctx context.Context, tenantUUID string, req EvaluateRiskRequest) (*RiskEvaluateResult, error) {
	if s == nil || s.ruleRepo == nil || s.blacklistRepo == nil || s.hitRepo == nil {
		return nil, errors.New("risk service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	now := time.Now().UTC()
	fingerprint := buildRiskFingerprint(req)
	if fingerprint != "" {
		releasedRows, err := s.hitRepo.ListReleasedByFingerprint(ctx, fingerprint)
		if err != nil {
			return nil, err
		}
		if len(releasedRows) > 0 {
			return &RiskEvaluateResult{
				Blocked:     false,
				Decision:    "allow",
				Score:       0,
				ReleasedBy:  releasedRows[0].ReleasedBy,
				Fingerprint: fingerprint,
			}, nil
		}
	}

	rules, err := s.ruleRepo.List(ctx, true)
	if err != nil {
		return nil, err
	}
	blacklist, err := s.blacklistRepo.ListActive(ctx, now)
	if err != nil {
		return nil, err
	}

	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority > rules[j].Priority })
	matchedRules := make([]LogisticsModel.RiskRule, 0)
	matchedList := make([]LogisticsModel.BlacklistEntry, 0)
	hits := make([]LogisticsModel.RiskHit, 0)
	blocked := false
	score := 0

	for _, item := range blacklist {
		if !blacklistMatch(item, req) {
			continue
		}
		matchedList = append(matchedList, item)
		score += 80
		blocked = true
		payload, _ := jsonBytes(req.Context, []byte("{}"))
		hit := &LogisticsModel.RiskHit{
			ID:          utils.NewUUID(),
			WaybillID:   strings.TrimSpace(req.WaybillID),
			WaybillNo:   strings.TrimSpace(req.WaybillNo),
			BlacklistID: item.ID,
			Source:      "blacklist",
			Decision:    "block",
			RiskLevel:   "high",
			Fingerprint: fingerprint,
			Description: firstNonEmpty(strings.TrimSpace(item.Reason), "matched blacklist"),
			Status:      "open",
			Payload:     datatypes.JSON(payload),
		}
		if err := s.hitRepo.Create(ctx, hit); err != nil {
			return nil, err
		}
		hits = append(hits, *hit)
	}

	for _, rule := range rules {
		if !matchRiskRule(rule, req) {
			continue
		}
		matchedRules = append(matchedRules, rule)
		if strings.EqualFold(rule.Decision, "block") {
			blocked = true
			score += 60
		} else {
			score += 30
		}
		payload, _ := jsonBytes(req.Context, []byte("{}"))
		hit := &LogisticsModel.RiskHit{
			ID:          utils.NewUUID(),
			WaybillID:   strings.TrimSpace(req.WaybillID),
			WaybillNo:   strings.TrimSpace(req.WaybillNo),
			RuleID:      rule.ID,
			Source:      "rule",
			Decision:    normalizeRiskDecision(rule.Decision),
			RiskLevel:   firstNonEmpty(strings.TrimSpace(rule.RiskLevel), "medium"),
			Fingerprint: fingerprint,
			Description: firstNonEmpty(strings.TrimSpace(rule.Description), "matched risk rule"),
			Status:      "open",
			Payload:     datatypes.JSON(payload),
		}
		if err := s.hitRepo.Create(ctx, hit); err != nil {
			return nil, err
		}
		hits = append(hits, *hit)
	}

	decision := "allow"
	if blocked {
		decision = "block"
	} else if len(matchedRules) > 0 {
		decision = "review"
	}
	if score > 100 {
		score = 100
	}

	return &RiskEvaluateResult{
		Blocked:      blocked,
		Decision:     decision,
		Score:        score,
		Fingerprint:  fingerprint,
		MatchedRules: matchedRules,
		MatchedList:  matchedList,
		Hits:         hits,
	}, nil
}

func (s *RiskService) Release(ctx context.Context, tenantUUID, hitID string, req ReleaseRiskHitRequest) (*LogisticsModel.RiskHit, error) {
	if s == nil || s.hitRepo == nil {
		return nil, errors.New("risk service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.hitRepo.GetByID(ctx, hitID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	row.Status = "released"
	row.ReleasedBy = strings.TrimSpace(req.OperatorID)
	row.ReleaseReason = strings.TrimSpace(req.Reason)
	row.ReleasedAt = &now
	if err := s.hitRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func buildRiskFingerprint(req EvaluateRiskRequest) string {
	raw := strings.ToLower(strings.Join([]string{
		strings.TrimSpace(req.RecipientName),
		strings.TrimSpace(req.RecipientPhone),
		strings.TrimSpace(req.DestinationLine),
	}, "|"))
	if strings.Trim(raw, "|") == "" {
		return ""
	}
	h := sha1.Sum([]byte(raw))
	return hex.EncodeToString(h[:])
}

func normalizeRiskDecision(decision string) string {
	if strings.EqualFold(strings.TrimSpace(decision), "block") {
		return "block"
	}
	return "review"
}

func isAllowed(input string, allowed ...string) bool {
	for _, item := range allowed {
		if strings.EqualFold(strings.TrimSpace(input), strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

func matchRiskRule(rule LogisticsModel.RiskRule, req EvaluateRiskRequest) bool {
	candidate := ""
	switch strings.ToLower(strings.TrimSpace(rule.MatchField)) {
	case "recipient":
		candidate = strings.TrimSpace(req.RecipientName)
	case "phone":
		candidate = strings.TrimSpace(req.RecipientPhone)
	default:
		candidate = strings.TrimSpace(req.DestinationLine)
	}
	if candidate == "" {
		return false
	}
	pattern := strings.TrimSpace(rule.Pattern)
	if pattern == "" {
		return false
	}
	candidate = strings.ToLower(candidate)
	pattern = strings.ToLower(pattern)
	switch strings.ToLower(strings.TrimSpace(rule.MatchMode)) {
	case "exact":
		return candidate == pattern
	case "prefix":
		return strings.HasPrefix(candidate, pattern)
	case "regex":
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(candidate)
	default:
		return strings.Contains(candidate, pattern)
	}
}

func blacklistMatch(row LogisticsModel.BlacklistEntry, req EvaluateRiskRequest) bool {
	if !strings.EqualFold(strings.TrimSpace(row.Status), "active") {
		return false
	}
	if row.ExpiresAt != nil && row.ExpiresAt.Before(time.Now().UTC()) {
		return false
	}
	name := strings.ToLower(strings.TrimSpace(req.RecipientName))
	phone := strings.ToLower(strings.TrimSpace(req.RecipientPhone))
	addr := strings.ToLower(strings.TrimSpace(req.DestinationLine))

	if v := strings.ToLower(strings.TrimSpace(row.RecipientPhone)); v != "" && phone != "" && phone == v {
		return true
	}
	if v := strings.ToLower(strings.TrimSpace(row.RecipientName)); v != "" && name != "" && strings.Contains(name, v) {
		return true
	}
	if v := strings.ToLower(strings.TrimSpace(row.AddressLine)); v != "" && addr != "" && strings.Contains(addr, v) {
		return true
	}
	return false
}
