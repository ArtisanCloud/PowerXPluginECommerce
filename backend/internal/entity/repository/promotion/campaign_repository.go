package promotion

import (
	"context"
	"strings"
	"time"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// CampaignRepository handles promotion campaign persistence.
type CampaignRepository struct {
	*repository.BaseRepository[promotionmodel.Campaign]
}

func NewCampaignRepository(db *gorm.DB) *CampaignRepository {
	return &CampaignRepository{BaseRepository: repository.NewBaseRepository[promotionmodel.Campaign](db)}
}

func (r *CampaignRepository) ListActive(ctx context.Context, tenantUUID, channel string, at time.Time) ([]promotionmodel.Campaign, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	var rows []promotionmodel.Campaign
	err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Where("status = ?", promotionmodel.StatusActive).
		Where("valid_from <= ? AND valid_to >= ?", at, at).
		Order("updated_at DESC, created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *CampaignRepository) GetByID(ctx context.Context, tenantUUID, id string) (*promotionmodel.Campaign, error) {
	if r == nil || r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	id = strings.TrimSpace(id)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	var row promotionmodel.Campaign
	if err := r.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}
