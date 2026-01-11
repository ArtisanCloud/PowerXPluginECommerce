package product_sku

import (
	"context"
	"errors"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// AuditLogRepository persists immutable audit events for SKU operations.
type AuditLogRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUAuditLog]
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUAuditLog](db)}
}

// CreateWithTx persists an audit log row using the provided transaction.
func (r *AuditLogRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, log *productskumodel.ProductSKUAuditLog) error {
	if r == nil || r.DB == nil {
		return errors.New("audit log repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if log == nil {
		return errors.New("audit log is required")
	}
	return tx.WithContext(ctx).Create(log).Error
}

