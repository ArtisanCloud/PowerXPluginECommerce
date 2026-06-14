package promotion

import (
	"context"
	"strings"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// AuditLogRepository handles promotion audit log persistence.
type AuditLogRepository struct {
	*repository.BaseRepository[promotionmodel.AuditLog]
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{BaseRepository: repository.NewBaseRepository[promotionmodel.AuditLog](db)}
}

func (r *AuditLogRepository) ListByPromotion(ctx context.Context, tenantUUID, promotionID string, page, pageSize int) ([]promotionmodel.AuditLog, int64, error) {
	if r == nil || r.DB == nil {
		return nil, 0, gorm.ErrInvalidDB
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	promotionID = strings.TrimSpace(promotionID)
	if tenantUUID == "" {
		return nil, 0, repository.ErrTenantUuidRequired
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	query := r.DB.WithContext(ctx).Model(&promotionmodel.AuditLog{}).Where("tenant_uuid = ?", tenantUUID)
	if promotionID != "" {
		query = query.Where("promotion_id = ?", promotionID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []promotionmodel.AuditLog
	err := query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error
	return rows, total, err
}
