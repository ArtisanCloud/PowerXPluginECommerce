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

// IdentityRepository 提供 customer_identities CRUD。
type IdentityRepository struct {
	*repository.BaseRepository[customermodel.CustomerIdentity]
}

// NewIdentityRepository 构造身份仓储。
func NewIdentityRepository(db *gorm.DB) *IdentityRepository {
	return &IdentityRepository{
		BaseRepository: repository.NewBaseRepository[customermodel.CustomerIdentity](db),
	}
}

// FindByProviderAppSubject 返回匹配身份记录。
func (r *IdentityRepository) FindByProviderAppSubject(ctx context.Context, provider, appID, subject string) (*customermodel.CustomerIdentity, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var entity customermodel.CustomerIdentity
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Where("provider = ?", strings.TrimSpace(provider)).
		Where("app_id = ?", strings.TrimSpace(appID)).
		Where("subject = ?", strings.TrimSpace(subject)).
		First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindByProviderSubjectLegacy 查找旧记录（app_id 为空）。
func (r *IdentityRepository) FindByProviderSubjectLegacy(ctx context.Context, provider, subject string) (*customermodel.CustomerIdentity, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var entity customermodel.CustomerIdentity
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Where("provider = ?", strings.TrimSpace(provider)).
		Where("subject = ?", strings.TrimSpace(subject)).
		Where("(app_id IS NULL OR app_id = '')").
		First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindByCustomerProviderApp 返回匹配客户、应用的身份记录。
func (r *IdentityRepository) FindByCustomerProviderApp(ctx context.Context, customerID, provider, appID string) (*customermodel.CustomerIdentity, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var entity customermodel.CustomerIdentity
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Where("customer_id = ?", strings.TrimSpace(customerID)).
		Where("provider = ?", strings.TrimSpace(provider)).
		Where("app_id = ?", strings.TrimSpace(appID)).
		Where("deleted_at IS NULL").
		First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// CreateIdentity 插入身份记录。
func (r *IdentityRepository) CreateIdentity(ctx context.Context, entity *customermodel.CustomerIdentity) (*customermodel.CustomerIdentity, error) {
	if entity == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	entity.TenantUUID = tenantUUID
	entity.Provider = strings.TrimSpace(entity.Provider)
	entity.AppID = strings.TrimSpace(entity.AppID)
	entity.Subject = strings.TrimSpace(entity.Subject)
	if err := r.DB.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

// UpsertIdentity 按 provider+app_id+subject Upsert 身份记录。
func (r *IdentityRepository) UpsertIdentity(ctx context.Context, entity *customermodel.CustomerIdentity) (*customermodel.CustomerIdentity, error) {
	if entity == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	entity.TenantUUID = tenantUUID
	entity.Provider = strings.TrimSpace(entity.Provider)
	entity.AppID = strings.TrimSpace(entity.AppID)
	entity.Subject = strings.TrimSpace(entity.Subject)
	var out *customermodel.CustomerIdentity
	err = r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customermodel.CustomerIdentity
		findErr := tx.
			Where("tenant_uuid = ?", tenantUUID).
			Where("provider = ?", entity.Provider).
			Where("app_id = ?", entity.AppID).
			Where("subject = ?", entity.Subject).
			Where("deleted_at IS NULL").
			First(&existing).Error
		if findErr != nil && findErr != gorm.ErrRecordNotFound {
			return findErr
		}
		if findErr == nil {
			if err := tx.Model(&customermodel.CustomerIdentity{}).
				Where("id = ?", existing.ID).
				Where("tenant_uuid = ?", tenantUUID).
				Updates(map[string]any{
					"customer_id": entity.CustomerID,
					"union_id":    entity.UnionID,
					"metadata":    entity.Metadata,
					"updated_at":  time.Now(),
				}).Error; err != nil {
				return err
			}
			existing.CustomerID = entity.CustomerID
			existing.UnionID = entity.UnionID
			existing.Metadata = entity.Metadata
			existing.UpdatedAt = time.Now()
			out = &existing
			return nil
		}
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		out = entity
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateIdentity 保存身份记录。
func (r *IdentityRepository) UpdateIdentity(ctx context.Context, entity *customermodel.CustomerIdentity, updates map[string]any) error {
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
		Model(&customermodel.CustomerIdentity{}).
		Where("id = ?", entity.ID).
		Where("tenant_uuid = ?", tenantUUID).
		Updates(setMap).Error
}
