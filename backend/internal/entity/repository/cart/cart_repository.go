package cart

import (
	"context"
	"errors"
	"strings"
	"time"

	cartmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/cart"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	*repository.BaseRepository[cartmodel.Cart]
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{BaseRepository: repository.NewBaseRepository[cartmodel.Cart](db)}
}

func (r *Repository) FindByCustomerID(ctx context.Context, tenantUUID, customerID string) (*cartmodel.Cart, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("cart repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if customerID == "" {
		return nil, errors.New("customer id is required")
	}
	var row cartmodel.Cart
	err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, customerID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *Repository) LockByCustomerID(ctx context.Context, tx *gorm.DB, tenantUUID, customerID string) (*cartmodel.Cart, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("cart repository is not initialized")
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if customerID == "" {
		return nil, errors.New("customer id is required")
	}
	var row cartmodel.Cart
	query := tx.WithContext(ctx)
	if query.Dialector != nil && query.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, customerID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *Repository) SaveWithTx(ctx context.Context, tx *gorm.DB, cart *cartmodel.Cart) error {
	if r == nil || r.DB == nil {
		return errors.New("cart repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if cart == nil {
		return errors.New("cart is required")
	}
	cart.UpdatedAt = time.Now().UTC()
	return tx.WithContext(ctx).Save(cart).Error
}
