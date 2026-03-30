package logistics

import (
	"context"
	"errors"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	LogisticsObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

// RateTemplateService handles freight-template operations.
type RateTemplateService struct {
	tplRepo  *LogisticsRepo.RateTemplateRepository
	zoneRepo *LogisticsRepo.RateZoneRepository
	emitter  *LogisticsObs.Emitter
}

func NewRateTemplateService(deps *app.Deps) *RateTemplateService {
	if deps == nil || deps.DB == nil {
		return &RateTemplateService{}
	}
	return &RateTemplateService{
		tplRepo:  LogisticsRepo.NewRateTemplateRepository(deps.DB),
		zoneRepo: LogisticsRepo.NewRateZoneRepository(deps.DB),
		emitter:  LogisticsObs.NewEmitter(deps.RuntimeLogger(context.Background(), "logistics-rate-template", nil)),
	}
}

type UpsertRateTemplateRequest struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name"`
	Currency string         `json:"currency,omitempty"`
	Channels map[string]any `json:"channels,omitempty"`
	Rules    map[string]any `json:"rules,omitempty"`
}

func (s *RateTemplateService) List(ctx context.Context, tenantUUID string) ([]LogisticsModel.RateTemplate, error) {
	if s == nil || s.tplRepo == nil {
		return nil, errors.New("rate template service unavailable")
	}
	return s.tplRepo.List(withTenantContext(ctx, tenantUUID))
}

func (s *RateTemplateService) Upsert(ctx context.Context, tenantUUID string, req UpsertRateTemplateRequest) (*LogisticsModel.RateTemplate, error) {
	if s == nil || s.tplRepo == nil {
		return nil, errors.New("rate template service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name required")
	}
	channels, _ := jsonBytes(req.Channels, []byte("[]"))
	rules, _ := jsonBytes(req.Rules, []byte("{}"))
	if strings.TrimSpace(req.ID) != "" {
		tpl, err := s.tplRepo.GetByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(tpl.Status, "published") {
			return nil, errors.New("published template is read-only")
		}
		tpl.Name = name
		tpl.Currency = defaultString(req.Currency, tpl.Currency)
		tpl.Channels = datatypes.JSON(channels)
		tpl.Rules = datatypes.JSON(rules)
		if err := s.tplRepo.Save(ctx, tpl); err != nil {
			return nil, err
		}
		s.emitAudit(tenantUUID, "rate_template.update", tpl.ID, "updated")
		return tpl, nil
	}

	tpl := &LogisticsModel.RateTemplate{
		ID:       utils.NewUUID(),
		Name:     name,
		Currency: defaultString(req.Currency, "CNY"),
		Status:   "draft",
		Version:  1,
		Channels: datatypes.JSON(channels),
		Rules:    datatypes.JSON(rules),
	}
	if err := s.tplRepo.Create(ctx, tpl); err != nil {
		return nil, err
	}
	s.emitAudit(tenantUUID, "rate_template.create", tpl.ID, "created")
	return tpl, nil
}

func (s *RateTemplateService) Publish(ctx context.Context, tenantUUID, templateID string) (*LogisticsModel.RateTemplate, error) {
	if s == nil || s.tplRepo == nil {
		return nil, errors.New("rate template service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	tpl, err := s.tplRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	tpl.Status = "published"
	tpl.Version = tpl.Version + 1
	if err := s.tplRepo.Save(ctx, tpl); err != nil {
		return nil, err
	}
	s.emitAudit(tenantUUID, "rate_template.publish", tpl.ID, "published")
	return tpl, nil
}

func (s *RateTemplateService) emitAudit(tenantUUID, action, targetID, result string) {
	if s == nil || s.emitter == nil {
		return
	}
	s.emitter.EmitAudit(LogisticsObs.AuditEvent{
		Action:     action,
		TenantUUID: strings.TrimSpace(tenantUUID),
		TargetID:   targetID,
		Result:     result,
		EmittedAt:  time.Now().UTC(),
	})
}
