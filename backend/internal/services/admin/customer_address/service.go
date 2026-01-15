package customer_address

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	customerrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/customer"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrAddressServiceUnavailable = errors.New("address service unavailable")
	ErrCustomerRequired          = errors.New("customer id is required")
	ErrAddressNotFound           = errors.New("address not found")
	ErrInvalidAddress            = errors.New("invalid shipping address")
	ErrAdminRequired             = errors.New("admin id is required")
)

type Service struct {
	deps        *app.Deps
	AddressRepo *customerrepo.CustomerAddressRepository
}

func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{deps: deps}
	}
	return &Service{
		deps:        deps,
		AddressRepo: customerrepo.NewCustomerAddressRepository(deps.DB),
	}
}

func (s *Service) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.AddressRepo != nil
}

func (s *Service) ListAddresses(ctx context.Context, tenantUUID, adminID, customerID string) ([]CustomerAddressDTO, error) {
	if !s.Ready() {
		return nil, ErrAddressServiceUnavailable
	}
	adminID = strings.TrimSpace(adminID)
	customerID = strings.TrimSpace(customerID)
	if adminID == "" {
		return nil, ErrAdminRequired
	}
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	rows, err := s.AddressRepo.ListByCustomer(ctx, tenantUUID, customerID)
	if err != nil {
		return nil, err
	}
	out := make([]CustomerAddressDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDTO(customerID, &row))
	}
	return out, nil
}

func (s *Service) CreateAddress(ctx context.Context, tenantUUID, adminID, customerID string, req AddressUpsertRequest) (*CustomerAddressDTO, error) {
	if !s.Ready() {
		return nil, ErrAddressServiceUnavailable
	}
	adminID = strings.TrimSpace(adminID)
	customerID = strings.TrimSpace(customerID)
	if adminID == "" {
		return nil, ErrAdminRequired
	}
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	if !isValidShippingAddress(&req.ShippingAddress) {
		return nil, ErrInvalidAddress
	}

	metadataRaw, _ := json.Marshal(req.ShippingAddress.Metadata)
	now := time.Now().UTC()
	entity := &customermodel.CustomerAddress{
		ID:             uuid.NewString(),
		TenantUUID:     strings.TrimSpace(tenantUUID),
		CustomerID:     customerID,
		IsDefault:      req.IsDefault != nil && *req.IsDefault,
		Label:          strings.TrimSpace(req.ShippingAddress.Label),
		RecipientName:  strings.TrimSpace(req.ShippingAddress.RecipientName),
		RecipientPhone: strings.TrimSpace(req.ShippingAddress.RecipientPhone),
		CountryCode:    strings.TrimSpace(req.ShippingAddress.CountryCode),
		Province:       strings.TrimSpace(req.ShippingAddress.Province),
		City:           strings.TrimSpace(req.ShippingAddress.City),
		District:       strings.TrimSpace(req.ShippingAddress.District),
		Address1:       strings.TrimSpace(req.ShippingAddress.Address1),
		Address2:       strings.TrimSpace(req.ShippingAddress.Address2),
		PostalCode:     strings.TrimSpace(req.ShippingAddress.PostalCode),
		Metadata:       datatypes.JSON(metadataRaw),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	var created *CustomerAddressDTO
	if err := s.AddressRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		var count int64
		if err := tx.WithContext(ctx).
			Model(&customermodel.CustomerAddress{}).
			Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, customerID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 && req.IsDefault == nil {
			entity.IsDefault = true
		}
		if entity.IsDefault {
			if err := s.AddressRepo.ClearDefaultWithTx(ctx, tx, tenantUUID, customerID); err != nil {
				return err
			}
		}
		if err := s.AddressRepo.CreateWithTx(ctx, tx, entity); err != nil {
			return err
		}
		created = ptr(toDTO(customerID, entity))
		return nil
	}); err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) UpdateAddress(ctx context.Context, tenantUUID, adminID, customerID, id string, req AddressUpsertRequest) (*CustomerAddressDTO, error) {
	if !s.Ready() {
		return nil, ErrAddressServiceUnavailable
	}
	adminID = strings.TrimSpace(adminID)
	customerID = strings.TrimSpace(customerID)
	id = strings.TrimSpace(id)
	if adminID == "" {
		return nil, ErrAdminRequired
	}
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	if id == "" {
		return nil, errors.New("address id is required")
	}
	if !isValidShippingAddress(&req.ShippingAddress) {
		return nil, ErrInvalidAddress
	}

	var updated *CustomerAddressDTO
	err := s.AddressRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		var row customermodel.CustomerAddress
		if err := tx.WithContext(ctx).
			Where("tenant_uuid = ? AND customer_id = ? AND id = ?", tenantUUID, customerID, id).
			First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAddressNotFound
			}
			return err
		}

		nextDefault := row.IsDefault
		if req.IsDefault != nil {
			nextDefault = *req.IsDefault
		}
		if nextDefault {
			if err := s.AddressRepo.ClearDefaultWithTx(ctx, tx, tenantUUID, customerID); err != nil {
				return err
			}
		}

		metadataRaw, _ := json.Marshal(req.ShippingAddress.Metadata)
		row.IsDefault = nextDefault
		row.Label = strings.TrimSpace(req.ShippingAddress.Label)
		row.RecipientName = strings.TrimSpace(req.ShippingAddress.RecipientName)
		row.RecipientPhone = strings.TrimSpace(req.ShippingAddress.RecipientPhone)
		row.CountryCode = strings.TrimSpace(req.ShippingAddress.CountryCode)
		row.Province = strings.TrimSpace(req.ShippingAddress.Province)
		row.City = strings.TrimSpace(req.ShippingAddress.City)
		row.District = strings.TrimSpace(req.ShippingAddress.District)
		row.Address1 = strings.TrimSpace(req.ShippingAddress.Address1)
		row.Address2 = strings.TrimSpace(req.ShippingAddress.Address2)
		row.PostalCode = strings.TrimSpace(req.ShippingAddress.PostalCode)
		row.Metadata = datatypes.JSON(metadataRaw)
		row.UpdatedAt = time.Now().UTC()

		if err := s.AddressRepo.UpdateWithTx(ctx, tx, &row); err != nil {
			return err
		}
		updated = ptr(toDTO(customerID, &row))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) DeleteAddress(ctx context.Context, tenantUUID, adminID, customerID, id string) error {
	if !s.Ready() {
		return ErrAddressServiceUnavailable
	}
	adminID = strings.TrimSpace(adminID)
	customerID = strings.TrimSpace(customerID)
	id = strings.TrimSpace(id)
	if adminID == "" {
		return ErrAdminRequired
	}
	if customerID == "" {
		return ErrCustomerRequired
	}
	if id == "" {
		return errors.New("address id is required")
	}

	return s.AddressRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		var row customermodel.CustomerAddress
		if err := tx.WithContext(ctx).
			Where("tenant_uuid = ? AND customer_id = ? AND id = ?", tenantUUID, customerID, id).
			First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAddressNotFound
			}
			return err
		}
		wasDefault := row.IsDefault
		if err := s.AddressRepo.DeleteByIDWithTx(ctx, tx, tenantUUID, customerID, id); err != nil {
			return err
		}
		if wasDefault {
			var next customermodel.CustomerAddress
			if err := tx.WithContext(ctx).
				Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, customerID).
				Order("updated_at DESC, id ASC").
				First(&next).Error; err == nil {
				_ = s.AddressRepo.ClearDefaultWithTx(ctx, tx, tenantUUID, customerID)
				_ = tx.WithContext(ctx).
					Model(&customermodel.CustomerAddress{}).
					Where("tenant_uuid = ? AND customer_id = ? AND id = ?", tenantUUID, customerID, next.ID).
					Update("is_default", true).Error
			}
		}
		return nil
	})
}

func (s *Service) SetDefault(ctx context.Context, tenantUUID, adminID, customerID, id string) (*CustomerAddressDTO, error) {
	if !s.Ready() {
		return nil, ErrAddressServiceUnavailable
	}
	adminID = strings.TrimSpace(adminID)
	customerID = strings.TrimSpace(customerID)
	id = strings.TrimSpace(id)
	if adminID == "" {
		return nil, ErrAdminRequired
	}
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	if id == "" {
		return nil, errors.New("address id is required")
	}

	var updated *CustomerAddressDTO
	err := s.AddressRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		var row customermodel.CustomerAddress
		if err := tx.WithContext(ctx).
			Where("tenant_uuid = ? AND customer_id = ? AND id = ?", tenantUUID, customerID, id).
			First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAddressNotFound
			}
			return err
		}
		if err := s.AddressRepo.ClearDefaultWithTx(ctx, tx, tenantUUID, customerID); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Model(&customermodel.CustomerAddress{}).
			Where("tenant_uuid = ? AND customer_id = ? AND id = ?", tenantUUID, customerID, id).
			Update("is_default", true).Error; err != nil {
			return err
		}
		row.IsDefault = true
		updated = ptr(toDTO(customerID, &row))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func isValidShippingAddress(addr *ShippingAddress) bool {
	if addr == nil {
		return false
	}
	if strings.TrimSpace(addr.RecipientName) == "" {
		return false
	}
	if strings.TrimSpace(addr.RecipientPhone) == "" {
		return false
	}
	if strings.TrimSpace(addr.Address1) == "" {
		return false
	}
	return true
}

func toDTO(customerID string, row *customermodel.CustomerAddress) CustomerAddressDTO {
	var metadata map[string]any
	if row != nil && len(row.Metadata) > 0 {
		_ = json.Unmarshal([]byte(row.Metadata), &metadata)
	}
	addr := ShippingAddress{}
	if row != nil {
		addr = ShippingAddress{
			Label:          strings.TrimSpace(row.Label),
			RecipientName:  strings.TrimSpace(row.RecipientName),
			RecipientPhone: strings.TrimSpace(row.RecipientPhone),
			CountryCode:    strings.TrimSpace(row.CountryCode),
			Province:       strings.TrimSpace(row.Province),
			City:           strings.TrimSpace(row.City),
			District:       strings.TrimSpace(row.District),
			Address1:       strings.TrimSpace(row.Address1),
			Address2:       strings.TrimSpace(row.Address2),
			PostalCode:     strings.TrimSpace(row.PostalCode),
			Metadata:       metadata,
		}
	}
	dto := CustomerAddressDTO{
		ID:         "",
		CustomerID: customerID,
		IsDefault:  false,
		Address:    addr,
	}
	if row != nil {
		dto.ID = row.ID
		dto.IsDefault = row.IsDefault
		dto.CreatedAt = row.CreatedAt
		dto.UpdatedAt = row.UpdatedAt
	}
	return dto
}

func ptr[T any](v T) *T { return &v }
