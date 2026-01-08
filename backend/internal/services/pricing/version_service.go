package pricing

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	pricingRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VersionService struct {
	*Service
	audit *AuditService
}

func NewVersionService(deps *app.Deps) *VersionService {
	svc := NewService(deps)
	return &VersionService{Service: svc, audit: &AuditService{Service: svc}}
}

type CreateVersionInput struct {
	PricebookID       string
	CopyFromVersionID *string
	Actor             string
}

type PublishVersionInput struct {
	PricebookID string
	VersionID   string
	EffectiveAt *time.Time
	ExpiresAt   *time.Time
	Note        *string
	Actor       string
}

type ArchiveVersionInput struct {
	PricebookID string
	VersionID   string
	Note        *string
	Actor       string
}

var publishLocks sync.Map // key string -> *sync.Mutex

func lockForPublish(key string) func() {
	v, _ := publishLocks.LoadOrStore(key, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (s *VersionService) CreateDraft(ctx context.Context, in CreateVersionInput) (*pricingModel.PricebookVersion, error) {
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

	tx, err := s.PricebookVersionRepo.BeginTenantTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var pb pricingModel.Pricebook
	if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, in.PricebookID).First(&pb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, E(CodePricebookNotFound, err)
		}
		return nil, err
	}

	repo := pricingRepo.NewPricebookVersionRepository(tx)
	maxVer, err := repo.MaxVersion(ctx, in.PricebookID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	v := &pricingModel.PricebookVersion{
		ID:          uuid.NewString(),
		TenantUUID:  tenantUUID,
		PricebookID: pb.ID,
		Version:     maxVer + 1,
		State:       "draft",
		EffectiveAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := tx.WithContext(ctx).Create(v).Error; err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, tx, "pricebook_version", v.ID, "create", in.Actor, map[string]any{
		"pricebook_id": v.PricebookID,
		"version":      v.Version,
		"copy_from":    in.CopyFromVersionID,
	})

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VersionService) Publish(ctx context.Context, in PublishVersionInput) (*pricingModel.PricebookVersion, error) {
	if !s.Ready() {
		return nil, E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, E(CodeTenantMissing, err)
	}
	pricebookID := strings.TrimSpace(in.PricebookID)
	versionID := strings.TrimSpace(in.VersionID)
	if pricebookID == "" || versionID == "" {
		return nil, E(CodeInvalidArgument, errors.New("pricebook_id/version_id are required"))
	}

	unlock := lockForPublish(tenantUUID + ":" + pricebookID)
	defer unlock()

	tx, err := s.PricebookVersionRepo.BeginTenantTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	pbRepo := pricingRepo.NewPricebookRepository(tx)
	if _, err := pbRepo.LockByID(ctx, pricebookID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, E(CodePricebookNotFound, err)
		}
		return nil, err
	}
	verRepo := pricingRepo.NewPricebookVersionRepository(tx)
	if err := verRepo.LockActiveVersions(ctx, pricebookID); err != nil {
		return nil, err
	}

	var v pricingModel.PricebookVersion
	if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ? AND pricebook_id = ?", tenantUUID, versionID, pricebookID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, E(CodeVersionNotFound, err)
		}
		return nil, err
	}
	if v.State != "draft" {
		return nil, E(CodeVersionNotEditable, errors.New("only draft version can be published"))
	}

	now := time.Now().UTC()
	effectiveAt := now
	if in.EffectiveAt != nil {
		effectiveAt = in.EffectiveAt.UTC()
	}
	if in.ExpiresAt != nil {
		expiresAt := in.ExpiresAt.UTC()
		if !expiresAt.After(effectiveAt) {
			return nil, E(CodeInvalidVersionRange, errors.New("expires_at must be after effective_at"))
		}
	}

	// Reject publish if it would overlap a future active version.
	var nextActive pricingModel.PricebookVersion
	nextActiveRes := tx.WithContext(ctx).Model(&pricingModel.PricebookVersion{}).
		Where("tenant_uuid = ? AND pricebook_id = ? AND state = ? AND id <> ? AND effective_at >= ?",
			tenantUUID, pricebookID, "active", versionID, effectiveAt).
		Order("effective_at asc").
		Limit(1).
		Select("id", "effective_at").
		Find(&nextActive)
	if nextActiveRes.Error != nil {
		return nil, nextActiveRes.Error
	}
	if nextActiveRes.RowsAffected > 0 {
		if in.ExpiresAt == nil || in.ExpiresAt.UTC().After(nextActive.EffectiveAt) {
			return nil, E(CodePublishConflict, errors.New("publish overlaps with an existing active version"))
		}
	}

	// End any currently active version at effectiveAt (without creating gaps).
	update := map[string]any{
		"expires_at": effectiveAt,
		"updated_at": now,
	}
	if !effectiveAt.After(now) {
		update["state"] = "expired"
	}
	if err := tx.WithContext(ctx).Model(&pricingModel.PricebookVersion{}).
		Where("tenant_uuid = ? AND pricebook_id = ? AND state = ? AND id <> ? AND effective_at < ? AND (expires_at IS NULL OR expires_at > ?)",
			tenantUUID, pricebookID, "active", versionID, effectiveAt, effectiveAt).
		Updates(update).Error; err != nil {
		return nil, err
	}

	note := ""
	if in.Note != nil {
		note = strings.TrimSpace(*in.Note)
	}
	actor := strings.TrimSpace(in.Actor)
	if actor == "" {
		actor = defaultActorFromTenant(tenantUUID)
	}
	if err := tx.WithContext(ctx).Model(&pricingModel.PricebookVersion{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, versionID).
		Updates(map[string]any{
			"state":        "active",
			"effective_at": effectiveAt,
			"expires_at":   in.ExpiresAt,
			"published_at": now,
			"published_by": actor,
			"note":         note,
			"updated_at":   now,
		}).Error; err != nil {
		return nil, err
	}

	if err := tx.WithContext(ctx).Model(&pricingModel.Pricebook{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, pricebookID).
		Update("current_version_id", versionID).Error; err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, tx, "pricebook_version", versionID, "publish", actor, map[string]any{
		"pricebook_id": pricebookID,
		"effective_at": effectiveAt,
		"expires_at":   in.ExpiresAt,
		"published_by": actor,
		"ended_active": true,
	})

	if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, versionID).First(&v).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *VersionService) Archive(ctx context.Context, in ArchiveVersionInput) (*pricingModel.PricebookVersion, error) {
	if !s.Ready() {
		return nil, E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, E(CodeTenantMissing, err)
	}
	pricebookID := strings.TrimSpace(in.PricebookID)
	versionID := strings.TrimSpace(in.VersionID)
	if pricebookID == "" || versionID == "" {
		return nil, E(CodeInvalidArgument, errors.New("pricebook_id/version_id are required"))
	}

	tx, err := s.PricebookVersionRepo.BeginTenantTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var v pricingModel.PricebookVersion
	if err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ? AND pricebook_id = ?", tenantUUID, versionID, pricebookID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, E(CodeVersionNotFound, err)
		}
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := v.ExpiresAt
	if expiresAt == nil || expiresAt.After(now) {
		expiresAt = &now
	}
	note := ""
	if in.Note != nil {
		note = strings.TrimSpace(*in.Note)
	}
	actor := strings.TrimSpace(in.Actor)
	if actor == "" {
		actor = defaultActorFromTenant(tenantUUID)
	}

	if err := tx.WithContext(ctx).Model(&pricingModel.PricebookVersion{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, versionID).
		Updates(map[string]any{
			"state":      "archived",
			"expires_at": expiresAt,
			"note":       note,
			"updated_at": now,
		}).Error; err != nil {
		return nil, err
	}

	// Clear current_version_id if it points to this version.
	if err := tx.WithContext(ctx).Model(&pricingModel.Pricebook{}).
		Where("tenant_uuid = ? AND id = ? AND current_version_id = ?", tenantUUID, pricebookID, versionID).
		Update("current_version_id", nil).Error; err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, tx, "pricebook_version", versionID, "archive", actor, map[string]any{
		"pricebook_id": pricebookID,
		"expires_at":   expiresAt,
	})

	if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, versionID).First(&v).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func defaultActorFromTenant(tenantUUID string) string {
	if strings.TrimSpace(tenantUUID) == "" {
		return "admin"
	}
	return "tenant:" + tenantUUID
}
