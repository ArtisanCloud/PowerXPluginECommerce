package pricing

import (
	"context"
	"errors"
	"strings"
	"time"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PricebookListFilters struct {
	Keyword  string
	Type     string
	Currency string
	Status   string
	Page     int
	PageSize int
}

type PricebookListResult struct {
	Items    []*pricingModel.Pricebook
	Page     int
	PageSize int
	Total    int64
}

type PricebookService struct {
	*Service
	audit *AuditService
}

func NewPricebookService(deps *app.Deps) *PricebookService {
	svc := NewService(deps)
	return &PricebookService{Service: svc, audit: &AuditService{Service: svc}}
}

func (s *PricebookService) List(ctx context.Context, f PricebookListFilters) (*PricebookListResult, error) {
	if !s.Ready() {
		return nil, E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, E(CodeTenantMissing, err)
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.PageSize > 200 {
		f.PageSize = 200
	}

	query := s.deps.DB.WithContext(ctx).Model(&pricingModel.Pricebook{}).Where("tenant_uuid = ?", tenantUUID)
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		needle := "%" + strings.ToLower(kw) + "%"
		query = query.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ?", needle, needle)
	}
	if v := strings.TrimSpace(f.Type); v != "" {
		query = query.Where("type = ?", v)
	}
	if v := strings.TrimSpace(f.Currency); v != "" {
		query = query.Where("currency = ?", v)
	}
	if v := strings.TrimSpace(f.Status); v != "" {
		query = query.Where("status = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []*pricingModel.Pricebook
	if err := query.Order("created_at desc").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &PricebookListResult{Items: rows, Page: f.Page, PageSize: f.PageSize, Total: total}, nil
}

type CreatePricebookInput struct {
	Code        string
	Name        string
	Type        string
	Currency    string
	Description string
	Scopes      *PricebookScopesInput
	Actor       string
}

type PricebookScopesInput struct {
	ChannelIDs       []string
	CustomerGroupIDs []string
	SupplierIDs      []string
}

func (s *PricebookService) Create(ctx context.Context, in CreatePricebookInput) (*pricingModel.Pricebook, error) {
	if !s.Ready() {
		return nil, E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, E(CodeTenantMissing, err)
	}
	code := strings.TrimSpace(in.Code)
	name := strings.TrimSpace(in.Name)
	currency := strings.TrimSpace(in.Currency)
	if code == "" || name == "" || currency == "" {
		return nil, E(CodeInvalidArgument, errors.New("code/name/currency are required"))
	}
	typ := strings.TrimSpace(in.Type)
	if typ == "" {
		typ = "sales"
	}
	if typ != "sales" && typ != "purchase" {
		return nil, E(CodeInvalidArgument, errors.New("type must be sales or purchase"))
	}

	tx, err := s.PricebookRepo.BeginTenantTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Guard uniqueness by tenant + code
	var exists int64
	if err := tx.WithContext(ctx).Model(&pricingModel.Pricebook{}).
		Where("tenant_uuid = ? AND code = ?", tenantUUID, code).
		Count(&exists).Error; err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, E(CodeDuplicatePricebook, errors.New("pricebook code already exists"))
	}

	now := time.Now().UTC()
	pb := &pricingModel.Pricebook{
		ID:          uuid.NewString(),
		TenantUUID:  tenantUUID,
		Code:        code,
		Name:        name,
		Type:        typ,
		Currency:    currency,
		Description: strings.TrimSpace(in.Description),
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := tx.WithContext(ctx).Create(pb).Error; err != nil {
		return nil, err
	}

	v1 := &pricingModel.PricebookVersion{
		ID:          uuid.NewString(),
		TenantUUID:  tenantUUID,
		PricebookID: pb.ID,
		Version:     1,
		State:       "draft",
		EffectiveAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := tx.WithContext(ctx).Create(v1).Error; err != nil {
		return nil, err
	}
	if err := tx.WithContext(ctx).Model(&pricingModel.Pricebook{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, pb.ID).
		Update("current_version_id", v1.ID).Error; err != nil {
		return nil, err
	}
	pb.CurrentVersionID = &v1.ID

	if in.Scopes != nil {
		scopeSvc := NewScopeService(s.deps)
		if err := scopeSvc.ReplacePricebookScopes(ctx, tx, ReplaceScopesInput{
			PricebookID: pb.ID,
			Scopes:      in.Scopes,
			Actor:       in.Actor,
		}); err != nil {
			return nil, err
		}
	}
	_ = s.audit.Log(ctx, tx, "pricebook", pb.ID, "create", in.Actor, map[string]any{
		"code":     pb.Code,
		"type":     pb.Type,
		"currency": pb.Currency,
	})

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return pb, nil
}

type UpdatePricebookInput struct {
	PricebookID string
	Name        *string
	Description *string
	Status      *string
	Scopes      *PricebookScopesInput
	Actor       string
}

func (s *PricebookService) Update(ctx context.Context, in UpdatePricebookInput) (*pricingModel.Pricebook, error) {
	if !s.Ready() {
		return nil, E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, E(CodeTenantMissing, err)
	}
	if strings.TrimSpace(in.PricebookID) == "" {
		return nil, E(CodeInvalidArgument, errors.New("pricebook_id is required"))
	}

	tx, err := s.PricebookRepo.BeginTenantTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var pb pricingModel.Pricebook
	if err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, in.PricebookID).
		First(&pb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, E(CodePricebookNotFound, err)
		}
		return nil, err
	}

	updates := map[string]any{}
	if in.Name != nil {
		updates["name"] = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		updates["description"] = strings.TrimSpace(*in.Description)
	}
	if in.Status != nil {
		v := strings.TrimSpace(*in.Status)
		if v != "active" && v != "archived" {
			return nil, E(CodeInvalidArgument, errors.New("status must be active or archived"))
		}
		updates["status"] = v
	}
	if len(updates) > 0 {
		if err := tx.WithContext(ctx).Model(&pricingModel.Pricebook{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, pb.ID).
			Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	if in.Scopes != nil {
		scopeSvc := NewScopeService(s.deps)
		if err := scopeSvc.ReplacePricebookScopes(ctx, tx, ReplaceScopesInput{
			PricebookID: pb.ID,
			Scopes:      in.Scopes,
			Actor:       in.Actor,
		}); err != nil {
			return nil, err
		}
	}

	_ = s.audit.Log(ctx, tx, "pricebook", pb.ID, "update", in.Actor, updates)

	if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, pb.ID).First(&pb).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &pb, nil
}
