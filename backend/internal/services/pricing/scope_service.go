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

type ScopeService struct {
	*Service
	audit *AuditService
}

func NewScopeService(deps *app.Deps) *ScopeService {
	svc := NewService(deps)
	return &ScopeService{Service: svc, audit: &AuditService{Service: svc}}
}

type ReplaceScopesInput struct {
	PricebookID string
	Scopes      *PricebookScopesInput // nil or empty => 全量适用（删除所有 scope 记录）
	Actor       string
}

func (s *ScopeService) ReplacePricebookScopes(ctx context.Context, tx *gorm.DB, in ReplaceScopesInput) error {
	if !s.Ready() {
		return E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return E(CodeTenantMissing, err)
	}
	pricebookID := strings.TrimSpace(in.PricebookID)
	if pricebookID == "" {
		return E(CodeInvalidArgument, errors.New("pricebook_id is required"))
	}
	db := tx
	if db == nil {
		var txErr error
		db, txErr = s.PricebookScopeRepo.BeginTenantTx(ctx)
		if txErr != nil {
			return txErr
		}
		defer func() { _ = db.Rollback() }()
	}

	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND pricebook_id = ?", tenantUUID, pricebookID).
		Delete(&pricingModel.PricebookScope{}).Error; err != nil {
		return err
	}

	// nil => keep deleted (全量适用)
	if in.Scopes == nil {
		if tx == nil {
			if err := db.Commit().Error; err != nil {
				return err
			}
		}
		return nil
	}

	now := time.Now().UTC()
	add := func(dimension string, ids []string) error {
		for _, raw := range ids {
			id := strings.TrimSpace(raw)
			if id == "" {
				continue
			}
			scope := &pricingModel.PricebookScope{
				ID:          uuid.NewString(),
				TenantUUID:  tenantUUID,
				PricebookID: pricebookID,
				Dimension:   dimension,
				DimensionID: id,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := db.WithContext(ctx).Create(scope).Error; err != nil {
				return err
			}
		}
		return nil
	}
	if err := add("channel", in.Scopes.ChannelIDs); err != nil {
		return err
	}
	if err := add("customer_group", in.Scopes.CustomerGroupIDs); err != nil {
		return err
	}
	if err := add("supplier", in.Scopes.SupplierIDs); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, db, "pricebook", pricebookID, "scopes_replace", in.Actor, map[string]any{
		"channel_ids":        len(in.Scopes.ChannelIDs),
		"customer_group_ids": len(in.Scopes.CustomerGroupIDs),
		"supplier_ids":       len(in.Scopes.SupplierIDs),
	})

	if tx == nil {
		if err := db.Commit().Error; err != nil {
			return err
		}
	}
	return nil
}
