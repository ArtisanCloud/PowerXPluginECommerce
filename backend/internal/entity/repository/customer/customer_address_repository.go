package customer

import (
	"context"
	"errors"
	"strings"

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

type CustomerAddressRepository struct {
	*repository.BaseRepository[customermodel.CustomerAddress]
}

func NewCustomerAddressRepository(db *gorm.DB) *CustomerAddressRepository {
	return &CustomerAddressRepository{BaseRepository: repository.NewBaseRepository[customermodel.CustomerAddress](db)}
}

func (r *CustomerAddressRepository) ListByCustomer(ctx context.Context, tenantUUID, customerID string) ([]customermodel.CustomerAddress, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("customer address repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if customerID == "" {
		return nil, errors.New("customer id is required")
	}
	var rows []customermodel.CustomerAddress
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, customerID).
		Order("is_default DESC, updated_at DESC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CustomerAddressRepository) GetByID(ctx context.Context, tenantUUID, customerID, id string) (*customermodel.CustomerAddress, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("customer address repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	id = strings.TrimSpace(id)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if customerID == "" {
		return nil, errors.New("customer id is required")
	}
	if id == "" {
		return nil, errors.New("address id is required")
	}
	var row customermodel.CustomerAddress
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND id = ?", tenantUUID, customerID, id).
		First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CustomerAddressRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, addr *customermodel.CustomerAddress) error {
	if r == nil || r.DB == nil {
		return errors.New("customer address repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if addr == nil {
		return errors.New("address is required")
	}
	return tx.WithContext(ctx).Create(addr).Error
}

func (r *CustomerAddressRepository) UpdateWithTx(ctx context.Context, tx *gorm.DB, addr *customermodel.CustomerAddress) error {
	if r == nil || r.DB == nil {
		return errors.New("customer address repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if addr == nil {
		return errors.New("address is required")
	}
	return tx.WithContext(ctx).Save(addr).Error
}

func (r *CustomerAddressRepository) DeleteByIDWithTx(ctx context.Context, tx *gorm.DB, tenantUUID, customerID, id string) error {
	if r == nil || r.DB == nil {
		return errors.New("customer address repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	id = strings.TrimSpace(id)
	if tenantUUID == "" {
		return repository.ErrTenantUuidRequired
	}
	if customerID == "" {
		return errors.New("customer id is required")
	}
	if id == "" {
		return errors.New("address id is required")
	}
	return tx.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND id = ?", tenantUUID, customerID, id).
		Delete(&customermodel.CustomerAddress{}).Error
}

func (r *CustomerAddressRepository) ClearDefaultWithTx(ctx context.Context, tx *gorm.DB, tenantUUID, customerID string) error {
	if r == nil || r.DB == nil {
		return errors.New("customer address repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if tenantUUID == "" {
		return repository.ErrTenantUuidRequired
	}
	if customerID == "" {
		return errors.New("customer id is required")
	}
	return tx.WithContext(ctx).
		Model(&customermodel.CustomerAddress{}).
		Where("tenant_uuid = ? AND customer_id = ? AND is_default = true", tenantUUID, customerID).
		Update("is_default", false).Error
}
