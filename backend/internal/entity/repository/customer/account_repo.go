package customer

import (
	"context"
	"strings"
	"time"

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// AccountRepository 提供 customer_accounts CRUD。
type AccountRepository struct {
	*repository.BaseRepository[customermodel.CustomerAccount]
}

// NewAccountRepository 构造账户仓储。
func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{
		BaseRepository: repository.NewBaseRepository[customermodel.CustomerAccount](db),
	}
}

// FindByIdentifier 返回匹配账户。
func (r *AccountRepository) FindByIdentifier(ctx context.Context, identifier string) (*customermodel.CustomerAccount, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var entity customermodel.CustomerAccount
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Where("LOWER(identifier) = ?", strings.ToLower(strings.TrimSpace(identifier))).
		First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// CreateAccount 插入记录。
func (r *AccountRepository) CreateAccount(ctx context.Context, entity *customermodel.CustomerAccount) (*customermodel.CustomerAccount, error) {
	if entity == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	entity.TenantUUID = tenantUUID
	entity.Identifier = strings.ToLower(strings.TrimSpace(entity.Identifier))
	if err := r.DB.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

// UpdateAccount 保存账户。
func (r *AccountRepository) UpdateAccount(ctx context.Context, entity *customermodel.CustomerAccount, updates map[string]any) error {
	if entity == nil || entity.ID == 0 {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	setMap := map[string]any{"updated_at": time.Now()}
	for k, v := range updates {
		setMap[k] = v
	}
	return r.DB.WithContext(ctx).
		Model(&customermodel.CustomerAccount{}).
		Where("id = ?", entity.ID).
		Where("tenant_uuid = ?", tenantUUID).
		Updates(setMap).Error
}
