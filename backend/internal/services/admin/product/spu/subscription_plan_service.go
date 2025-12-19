package spu

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

// SubscriptionPlanInput captures the fields accepted from HTTP for create/update.
type SubscriptionPlanInput struct {
	PlanCode     string         `json:"planCode"`
	Name         string         `json:"name"`
	BillingCycle string         `json:"billingCycle"`
	BillingValue int            `json:"billingValue"`
	Price        float64        `json:"price"`
	Currency     string         `json:"currency"`
	TrialDays    int            `json:"trialDays"`
	AutoRenew    bool           `json:"autoRenew"`
	CancelPolicy string         `json:"cancelPolicy"`
	EffectScope  string         `json:"effectScope"`
	Metadata     map[string]any `json:"metadata"`
	Status       string         `json:"status"`
}

// SubscriptionPlanEntry represents a persisted plan returned to API consumers.
type SubscriptionPlanEntry struct {
	ID           string         `json:"id"`
	PlanCode     string         `json:"planCode"`
	Name         string         `json:"name"`
	BillingCycle string         `json:"billingCycle"`
	BillingValue int            `json:"billingValue"`
	Price        float64        `json:"price"`
	Currency     string         `json:"currency"`
	TrialDays    int            `json:"trialDays"`
	AutoRenew    bool           `json:"autoRenew"`
	CancelPolicy string         `json:"cancelPolicy"`
	EffectScope  string         `json:"effectScope"`
	Status       string         `json:"status"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	CreatedAt    time.Time      `json:"createdAt"`
}

// SubscriptionPlanService manages CRUD for recurring plans scoped to an SPU.
type SubscriptionPlanService struct {
	deps     *app.Deps
	spuRepo  *productrepo.SPURepository
	planRepo *productrepo.SubscriptionPlanRepository
}

// NewSubscriptionPlanService wires repositories required for plan flows.
func NewSubscriptionPlanService(deps *app.Deps) *SubscriptionPlanService {
	if deps == nil || deps.DB == nil {
		return nil
	}
	return &SubscriptionPlanService{
		deps:     deps,
		spuRepo:  productrepo.NewSPURepository(deps.DB),
		planRepo: productrepo.NewSubscriptionPlanRepository(deps.DB),
	}
}

// List returns all non-archived subscription plans for the provided SPU.
func (s *SubscriptionPlanService) List(ctx context.Context, spuID string) ([]SubscriptionPlanEntry, error) {
	if s == nil {
		return nil, errors.New("subscription plan service unavailable")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var plans []productmodel.SubscriptionPlan
	if err := s.planRepo.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND spu_id = ? AND status <> ?", tenantID, spuID, "archived").
		Order("created_at ASC").
		Find(&plans).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []SubscriptionPlanEntry{}, nil
		}
		return nil, err
	}
	return s.toEntries(plans), nil
}

// Create writes a new plan entry.
func (s *SubscriptionPlanService) Create(ctx context.Context, spuID string, input SubscriptionPlanInput) (*SubscriptionPlanEntry, error) {
	if s == nil {
		return nil, errors.New("subscription plan service unavailable")
	}
	if err := s.validateInput(input, true); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var entry *SubscriptionPlanEntry
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		spu, err := s.ensureSubscriptionSPU(tx, tenantID, spuID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		record := productmodel.SubscriptionPlan{
			ID:           uuidString(),
			TenantUUID:   tenantID,
			SPUID:        spu.ID,
			PlanCode:     strings.TrimSpace(input.PlanCode),
			Name:         strings.TrimSpace(input.Name),
			BillingCycle: strings.ToLower(strings.TrimSpace(input.BillingCycle)),
			BillingValue: input.BillingValue,
			Price:        input.Price,
			Currency:     normalizeCurrency(input.Currency),
			TrialDays:    input.TrialDays,
			AutoRenew:    input.AutoRenew,
			CancelPolicy: strings.ToLower(strings.TrimSpace(input.CancelPolicy)),
			EffectScope:  normalizeEffectScope(input.EffectScope),
			Status:       normalizePlanStatus(input.Status),
			Metadata:     encodeGenericJSON(input.Metadata),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		entries := s.toEntries([]productmodel.SubscriptionPlan{record})
		if len(entries) > 0 {
			entry = &entries[0]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// Update edits mutable fields for a plan.
func (s *SubscriptionPlanService) Update(ctx context.Context, spuID, planID string, input SubscriptionPlanInput) (*SubscriptionPlanEntry, error) {
	if s == nil {
		return nil, errors.New("subscription plan service unavailable")
	}
	if err := s.validateInput(input, false); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var entry *SubscriptionPlanEntry
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		if _, err := s.ensureSubscriptionSPU(tx, tenantID, spuID); err != nil {
			return err
		}
		var plan productmodel.SubscriptionPlan
		if err := tx.Where("tenant_uuid = ? AND id = ? AND spu_id = ?", tenantID, planID, spuID).
			First(&plan).Error; err != nil {
			return err
		}
		updates := map[string]any{
			"name":          strings.TrimSpace(input.Name),
			"billing_cycle": strings.ToLower(strings.TrimSpace(input.BillingCycle)),
			"billing_value": input.BillingValue,
			"price":         input.Price,
			"currency":      normalizeCurrency(input.Currency),
			"trial_days":    input.TrialDays,
			"auto_renew":    input.AutoRenew,
			"cancel_policy": strings.ToLower(strings.TrimSpace(input.CancelPolicy)),
			"effect_scope":  normalizeEffectScope(input.EffectScope),
			"metadata":      encodeGenericJSON(input.Metadata),
			"status":        normalizePlanStatus(input.Status),
			"updated_at":    time.Now().UTC(),
		}
		if strings.TrimSpace(input.PlanCode) != "" {
			updates["plan_code"] = strings.TrimSpace(input.PlanCode)
		}
		if err := tx.Model(&plan).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, plan.ID).First(&plan).Error; err != nil {
			return err
		}
		entries := s.toEntries([]productmodel.SubscriptionPlan{plan})
		if len(entries) > 0 {
			entry = &entries[0]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// Archive marks the plan as archived without deleting the record.
func (s *SubscriptionPlanService) Archive(ctx context.Context, spuID, planID string) error {
	if s == nil {
		return errors.New("subscription plan service unavailable")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return err
	}
	return s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		if _, err := s.ensureSubscriptionSPU(tx, tenantID, spuID); err != nil {
			return err
		}
		return tx.Model(&productmodel.SubscriptionPlan{}).
			Where("tenant_uuid = ? AND id = ? AND spu_id = ?", tenantID, planID, spuID).
			Update("status", "archived").Error
	})
}

func (s *SubscriptionPlanService) ensureSubscriptionSPU(tx *gorm.DB, tenantID, spuID string) (*productmodel.SPU, error) {
	var spu productmodel.SPU
	if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, spuID).First(&spu).Error; err != nil {
		return nil, err
	}
	if strings.ToLower(spu.Type) != "subscription" {
		return nil, fmt.Errorf("spu %s is not subscription type", spuID)
	}
	return &spu, nil
}

func (s *SubscriptionPlanService) validateInput(input SubscriptionPlanInput, requireCode bool) error {
	if requireCode && strings.TrimSpace(input.PlanCode) == "" {
		return errors.New("planCode is required")
	}
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(input.BillingCycle) == "" {
		return errors.New("billingCycle is required")
	}
	if input.Price <= 0 {
		return errors.New("price must be greater than 0")
	}
	if strings.TrimSpace(input.CancelPolicy) == "" {
		return errors.New("cancelPolicy is required")
	}
	if eff := strings.ToLower(strings.TrimSpace(input.EffectScope)); eff != "" {
		switch eff {
		case "new_only", "new_and_existing":
		default:
			return fmt.Errorf("invalid effectScope: %s", input.EffectScope)
		}
	}
	return nil
}

func (s *SubscriptionPlanService) toEntries(plans []productmodel.SubscriptionPlan) []SubscriptionPlanEntry {
	if len(plans) == 0 {
		return []SubscriptionPlanEntry{}
	}
	result := make([]SubscriptionPlanEntry, 0, len(plans))
	for _, plan := range plans {
		result = append(result, SubscriptionPlanEntry{
			ID:           plan.ID,
			PlanCode:     plan.PlanCode,
			Name:         plan.Name,
			BillingCycle: plan.BillingCycle,
			BillingValue: plan.BillingValue,
			Price:        plan.Price,
			Currency:     plan.Currency,
			TrialDays:    plan.TrialDays,
			AutoRenew:    plan.AutoRenew,
			CancelPolicy: plan.CancelPolicy,
			EffectScope:  plan.EffectScope,
			Status:       plan.Status,
			Metadata:     decodeGenericJSON(plan.Metadata),
			CreatedAt:    plan.CreatedAt,
			UpdatedAt:    plan.UpdatedAt,
		})
	}
	return result
}

func (s *SubscriptionPlanService) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil || s == nil {
		return "", errors.New("subscription plan service unavailable")
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return tid, nil
	}
	return "", ErrMissingTenant
}

func normalizeCurrency(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "CNY"
	}
	return value
}

func normalizeEffectScope(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "new_and_existing":
		return "new_and_existing"
	default:
		return "new_only"
	}
}

func normalizePlanStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "archived", "inactive":
		return "archived"
	default:
		return "active"
	}
}
