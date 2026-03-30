package fulfillment

import (
	"context"
	"strings"
	"time"

	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	"gorm.io/gorm"
)

// ExceptionRepository manages fulfillment exception persistence.
type ExceptionRepository struct {
	*Repository[FulfillmentModel.Exception]
}

func NewExceptionRepository(db *gorm.DB) *ExceptionRepository {
	return &ExceptionRepository{Repository: NewRepository[FulfillmentModel.Exception](db)}
}

func (r *ExceptionRepository) List(ctx context.Context, status string) ([]FulfillmentModel.Exception, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []FulfillmentModel.Exception
	err = query.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *ExceptionRepository) GetByID(ctx context.Context, id string) (*FulfillmentModel.Exception, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row FulfillmentModel.Exception
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ExceptionRepository) Create(ctx context.Context, ex *FulfillmentModel.Exception) error {
	if ex == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(ex.TenantUUID) == "" {
		ex.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(ex).Error
}

func (r *ExceptionRepository) Save(ctx context.Context, ex *FulfillmentModel.Exception) error {
	if ex == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(ex.TenantUUID) == "" {
		ex.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Save(ex).Error
}

// EscalateOpenBefore escalates open exceptions that have no first action before cutoff.
func (r *ExceptionRepository) EscalateOpenBefore(ctx context.Context, cutoff time.Time, now time.Time) ([]FulfillmentModel.Exception, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []FulfillmentModel.Exception
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND status = ? AND first_action_at IS NULL AND created_at <= ?", tenantUUID, "open", cutoff).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for idx := range rows {
		rows[idx].Status = "escalated"
		rows[idx].EscalatedAt = &now
		if saveErr := r.DB.WithContext(ctx).Save(&rows[idx]).Error; saveErr != nil {
			return nil, saveErr
		}
	}
	return rows, nil
}
