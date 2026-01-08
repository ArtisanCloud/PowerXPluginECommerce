package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

type ActiveCandidate struct {
	Pricebook pricingModel.Pricebook
	Version   pricingModel.PricebookVersion
}

// PriceQueryRepository encapsulates read queries for pricing/query.
type PriceQueryRepository struct {
	db *gorm.DB
}

func NewPriceQueryRepository(db *gorm.DB) *PriceQueryRepository { return &PriceQueryRepository{db: db} }

func (r *PriceQueryRepository) ListActiveCandidates(ctx context.Context, currency string, asOf time.Time) ([]ActiveCandidate, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository database is not initialized")
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(currency) == "" {
		return nil, errors.New("currency is required")
	}
	type row struct {
		PBID             string
		PBCode           string
		PBName           string
		PBType           string
		PBCurrency       string
		PBStatus         string
		PBCurrentVersion *string
		PBDescription    string
		VId              string
		VPricebookID     string
		VVersion         int
		VState           string
		VEffectiveAt     time.Time
		VExpiresAt       *time.Time
		VPublishedAt     *time.Time
		VPublishedBy     string
		VNote            string
		VCreatedAt       time.Time
		VUpdatedAt       time.Time
	}
	var rows []row
	q := r.db.WithContext(ctx).Table(pricingModel.Pricebook{}.TableName()+" AS pb").
		Select(`
			pb.id as pb_id,
			pb.code as pb_code,
			pb.name as pb_name,
			pb.type as pb_type,
			pb.currency as pb_currency,
			pb.status as pb_status,
			pb.current_version_id as pb_current_version,
			COALESCE(pb.description,'') as pb_description,
			v.id as v_id,
			v.pricebook_id as v_pricebook_id,
			v.version as v_version,
			v.state as v_state,
			v.effective_at as v_effective_at,
			v.expires_at as v_expires_at,
			v.published_at as v_published_at,
			COALESCE(v.published_by,'') as v_published_by,
			COALESCE(v.note,'') as v_note,
			v.created_at as v_created_at,
			v.updated_at as v_updated_at
		`).
		Joins("JOIN "+pricingModel.PricebookVersion{}.TableName()+" AS v ON v.pricebook_id = pb.id").
		Where("pb.tenant_uuid = ? AND pb.status = ? AND pb.currency = ?", tenantUUID, "active", strings.TrimSpace(currency)).
		Where("v.tenant_uuid = ? AND v.state = ? AND v.effective_at <= ? AND (v.expires_at IS NULL OR v.expires_at > ?)",
			tenantUUID, "active", asOf, asOf)
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]ActiveCandidate, 0, len(rows))
	for _, r := range rows {
		pb := pricingModel.Pricebook{
			ID:               r.PBID,
			TenantUUID:       tenantUUID,
			Code:             r.PBCode,
			Name:             r.PBName,
			Type:             r.PBType,
			Currency:         r.PBCurrency,
			Description:      r.PBDescription,
			Status:           r.PBStatus,
			CurrentVersionID: r.PBCurrentVersion,
		}
		v := pricingModel.PricebookVersion{
			ID:          r.VId,
			TenantUUID:  tenantUUID,
			PricebookID: r.VPricebookID,
			Version:     r.VVersion,
			State:       r.VState,
			EffectiveAt: r.VEffectiveAt,
			ExpiresAt:   r.VExpiresAt,
			PublishedAt: r.VPublishedAt,
			PublishedBy: r.VPublishedBy,
			Note:        r.VNote,
			CreatedAt:   r.VCreatedAt,
			UpdatedAt:   r.VUpdatedAt,
		}
		out = append(out, ActiveCandidate{Pricebook: pb, Version: v})
	}
	return out, nil
}

func (r *PriceQueryRepository) ListScopesByPricebook(ctx context.Context, pricebookID string) ([]*pricingModel.PricebookScope, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository database is not initialized")
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(pricebookID) == "" {
		return nil, errors.New("pricebook_id is required")
	}
	var scopes []*pricingModel.PricebookScope
	if err := r.db.WithContext(ctx).
		Where("tenant_uuid = ? AND pricebook_id = ?", tenantUUID, strings.TrimSpace(pricebookID)).
		Find(&scopes).Error; err != nil {
		return nil, err
	}
	return scopes, nil
}

func (r *PriceQueryRepository) FindItem(ctx context.Context, versionID, skuID string) (*pricingModel.PricebookItem, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository database is not initialized")
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(versionID) == "" || strings.TrimSpace(skuID) == "" {
		return nil, errors.New("version_id and sku_id are required")
	}
	var item pricingModel.PricebookItem
	res := r.db.WithContext(ctx).
		Where("tenant_uuid = ? AND version_id = ? AND sku_id = ?", tenantUUID, strings.TrimSpace(versionID), strings.TrimSpace(skuID)).
		First(&item)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return &item, nil
}
