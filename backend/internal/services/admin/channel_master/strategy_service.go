package channel_master

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelStrategyInput describes strategy fields accepted from HTTP layer.
type ChannelStrategyInput struct {
	PricebookID         string     `json:"pricebook_id"`
	InventoryStrategyID string     `json:"inventory_strategy_id"`
	LogisticsStrategyID string     `json:"logistics_strategy_id"`
	CSSLAID             string     `json:"cs_sla_id"`
	FeeRate             *float64   `json:"fee_rate"`
	SettlementCycle     string     `json:"settlement_cycle"`
	PaymentTerms        string     `json:"payment_terms"`
	Notes               string     `json:"notes"`
	EffectiveAt         *time.Time `json:"effective_at"`
}

// ChannelTeamInput captures team assignment payload.
type ChannelTeamInput struct {
	OwnerUUID    string   `json:"owner_uuid"`
	ApproverUUID string   `json:"approver_uuid"`
	Operators    []string `json:"operators"`
}

// StrategyUpsertInput combines strategy + team updates.
type StrategyUpsertInput struct {
	Strategy ChannelStrategyInput `json:"strategy"`
	Team     ChannelTeamInput     `json:"team"`
}

// ChannelStrategySnapshot returned to callers.
type ChannelStrategySnapshot struct {
	Config ChannelStrategyOutput `json:"config"`
	Team   ChannelTeamSnapshot   `json:"team"`
}

// ChannelStrategyOutput is serializable representation.
type ChannelStrategyOutput struct {
	PricebookID         string     `json:"pricebook_id,omitempty"`
	InventoryStrategyID string     `json:"inventory_strategy_id,omitempty"`
	LogisticsStrategyID string     `json:"logistics_strategy_id,omitempty"`
	CSSLAID             string     `json:"cs_sla_id,omitempty"`
	FeeRate             float64    `json:"fee_rate"`
	SettlementCycle     string     `json:"settlement_cycle,omitempty"`
	PaymentTerms        string     `json:"payment_terms,omitempty"`
	Notes               string     `json:"notes,omitempty"`
	EffectiveAt         *time.Time `json:"effective_at,omitempty"`
}

// ChannelTeamSnapshot returns assignment info.
type ChannelTeamSnapshot struct {
	OwnerUUID    string   `json:"owner_uuid"`
	ApproverUUID string   `json:"approver_uuid"`
	Operators    []string `json:"operators"`
}

// StrategyService manages strategy/team persistence.
type StrategyService struct {
	configRepo  *channelrepo.ChannelConfigRepository
	channelRepo *channelrepo.ChannelMasterRepository
	logger      *logrus.Entry
	audit       AuditEmitter
}

// NewStrategyService wires repository dependencies.
func NewStrategyService(deps *app.Deps, audit AuditEmitter) *StrategyService {
	if deps == nil || deps.DB == nil {
		panic("strategy service requires database dependency")
	}
	logger := deps.RuntimeLogger(context.TODO(), "channel-strategy-service", nil)
	if audit == nil {
		audit = channelobs.NewAuditEmitter(logger)
	}
	return &StrategyService{
		configRepo:  channelrepo.NewChannelConfigRepository(deps.DB),
		channelRepo: channelrepo.NewChannelMasterRepository(deps.DB),
		logger:      logger,
		audit:       audit,
	}
}

// GetStrategy returns persisted config + team snapshot.
func (s *StrategyService) GetStrategy(ctx context.Context, channelID string) (*ChannelStrategySnapshot, error) {
	channel, err := s.channelRepo.FindByID(ctx, strings.TrimSpace(channelID))
	if err != nil {
		return nil, err
	}
	var cfg *channelmodel.ChannelConfig
	if entity, err := s.configRepo.FindByChannel(ctx, channel.ID); err == nil {
		cfg = entity
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.buildSnapshot(channel, cfg), nil
}

// UpsertStrategy validates and writes strategy/team data.
func (s *StrategyService) UpsertStrategy(ctx context.Context, channelID string, input StrategyUpsertInput) (*ChannelStrategySnapshot, error) {
	channel, err := s.channelRepo.FindByID(ctx, strings.TrimSpace(channelID))
	if err != nil {
		return nil, err
	}
	if err := input.Team.normalize(); err != nil {
		return nil, err
	}
	if err := input.Strategy.normalize(); err != nil {
		return nil, err
	}

	channel.OwnerUUID = input.Team.OwnerUUIDOrFallback(channel.OwnerUUID)
	channel.ApproverUUID = input.Team.ApproverPointer(channel.ApproverUUID)
	channel.UpdatedBy = actorFromContext(ctx)
	if _, err := s.channelRepo.Save(ctx, channel); err != nil {
		return nil, err
	}

	cfg := &channelmodel.ChannelConfig{
		ChannelID:           channel.ID,
		PricebookID:         input.Strategy.PricebookID,
		InventoryStrategyID: input.Strategy.InventoryStrategyID,
		LogisticsStrategyID: input.Strategy.LogisticsStrategyID,
		CSSLAID:             input.Strategy.CSSLAID,
		SettlementCycle:     input.Strategy.SettlementCycle,
		PaymentTerms:        input.Strategy.PaymentTerms,
		Notes:               input.Strategy.Notes,
		EffectiveAt:         input.Strategy.EffectiveAt,
		TeamMetadata:        encodeTeamMetadata(input.Team.Operators),
		CreatedBy:           actorFromContext(ctx),
		UpdatedBy:           actorFromContext(ctx),
	}
	if input.Strategy.FeeRate != nil {
		cfg.FeeRate = *input.Strategy.FeeRate
	}
	saved, err := s.configRepo.Upsert(ctx, cfg)
	if err != nil {
		return nil, err
	}
	s.emitAudit(ctx, channel.ID, input)
	return s.buildSnapshot(channel, saved), nil
}

func (s *StrategyService) buildSnapshot(channel *channelmodel.ChannelMaster, cfg *channelmodel.ChannelConfig) *ChannelStrategySnapshot {
	var strategy ChannelStrategyOutput
	if cfg != nil {
		strategy = ChannelStrategyOutput{
			PricebookID:         cfg.PricebookID,
			InventoryStrategyID: cfg.InventoryStrategyID,
			LogisticsStrategyID: cfg.LogisticsStrategyID,
			CSSLAID:             cfg.CSSLAID,
			FeeRate:             cfg.FeeRate,
			SettlementCycle:     cfg.SettlementCycle,
			PaymentTerms:        cfg.PaymentTerms,
			Notes:               cfg.Notes,
			EffectiveAt:         cfg.EffectiveAt,
		}
	}
	team := ChannelTeamSnapshot{
		OwnerUUID: channel.OwnerUUID,
		Operators: decodeOperators(cfg),
	}
	if channel.ApproverUUID != nil {
		team.ApproverUUID = *channel.ApproverUUID
	}
	return &ChannelStrategySnapshot{Config: strategy, Team: team}
}

func (s *StrategyService) emitAudit(ctx context.Context, channelID string, input StrategyUpsertInput) {
	if s.audit == nil {
		return
	}
	payload := map[string]any{
		"channel_id":            channelID,
		"owner_uuid":            input.Team.OwnerUUID,
		"approver_uuid":         input.Team.ApproverUUID,
		"pricebook_id":          input.Strategy.PricebookID,
		"inventory_strategy_id": input.Strategy.InventoryStrategyID,
		"logistics_strategy_id": input.Strategy.LogisticsStrategyID,
		"cs_sla_id":             input.Strategy.CSSLAID,
		"settlement_cycle":      input.Strategy.SettlementCycle,
		"payment_terms":         input.Strategy.PaymentTerms,
		"operators_count":       len(input.Team.Operators),
	}
	_ = s.audit.EmitChannelAudit(ctx, "channel.strategy.updated", payload)
}

func encodeTeamMetadata(operators []string) datatypes.JSON {
	if len(operators) == 0 {
		return datatypes.JSON([]byte(`{"operators":[]}`))
	}
	payload := map[string]any{"operators": operators}
	data, err := json.Marshal(payload)
	if err != nil {
		return datatypes.JSON([]byte(`{"operators":[]}`))
	}
	return datatypes.JSON(data)
}

func decodeOperators(cfg *channelmodel.ChannelConfig) []string {
	if cfg == nil || len(cfg.TeamMetadata) == 0 {
		return nil
	}
	var payload struct {
		Operators []string `json:"operators"`
	}
	if err := json.Unmarshal(cfg.TeamMetadata, &payload); err != nil {
		return nil
	}
	clean := make([]string, 0, len(payload.Operators))
	set := map[string]struct{}{}
	for _, item := range payload.Operators {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			lower := strings.ToLower(trimmed)
			if _, exists := set[lower]; exists {
				continue
			}
			set[lower] = struct{}{}
			clean = append(clean, trimmed)
		}
	}
	return clean
}

func (in *ChannelStrategyInput) normalize() error {
	in.PricebookID = strings.TrimSpace(in.PricebookID)
	in.InventoryStrategyID = strings.TrimSpace(in.InventoryStrategyID)
	in.LogisticsStrategyID = strings.TrimSpace(in.LogisticsStrategyID)
	in.CSSLAID = strings.TrimSpace(in.CSSLAID)
	in.SettlementCycle = strings.TrimSpace(strings.ToLower(in.SettlementCycle))
	in.PaymentTerms = strings.TrimSpace(in.PaymentTerms)
	in.Notes = strings.TrimSpace(in.Notes)
	if in.FeeRate != nil {
		if *in.FeeRate < 0 || *in.FeeRate > 100 {
			return fmt.Errorf("fee_rate must be between 0 and 100")
		}
	}
	if in.SettlementCycle != "" && !isSupportedSettlementCycle(in.SettlementCycle) {
		return fmt.Errorf("unsupported settlement_cycle: %s", in.SettlementCycle)
	}
	return nil
}

func (in *ChannelTeamInput) normalize() error {
	in.OwnerUUID = strings.TrimSpace(in.OwnerUUID)
	in.ApproverUUID = strings.TrimSpace(in.ApproverUUID)
	if in.OwnerUUID == "" {
		return errors.New("owner_uuid is required")
	}
	if len(in.Operators) > 0 {
		clean := make([]string, 0, len(in.Operators))
		seen := map[string]struct{}{}
		for _, op := range in.Operators {
			if trimmed := strings.TrimSpace(op); trimmed != "" {
				lower := strings.ToLower(trimmed)
				if _, ok := seen[lower]; ok {
					continue
				}
				seen[lower] = struct{}{}
				clean = append(clean, trimmed)
			}
		}
		in.Operators = clean
	}
	return nil
}

func (in ChannelTeamInput) OwnerUUIDOrFallback(current string) string {
	if in.OwnerUUID != "" {
		return in.OwnerUUID
	}
	return current
}

func (in ChannelTeamInput) ApproverPointer(current *string) *string {
	if in.ApproverUUID == "" {
		return current
	}
	return ptrString(in.ApproverUUID)
}

func isSupportedSettlementCycle(value string) bool {
	if value == "" {
		return true
	}
	switch value {
	case "dd", "net7", "net14", "net30", "net45", "net60":
		return true
	default:
		return false
	}
}
