package customer

import (
	"context"
	"strings"

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListQueryOptions 描述仓储层的列表筛选条件。
type ListQueryOptions struct {
	Keyword   string
	Tier      string
	Source    string
	Region    string
	RiskLevel string
	Type      string
	Tags      []string
	Sort      string
	Page      int
	PageSize  int
}

// Repository 提供客户实体的持久化访问能力。
type Repository struct {
	*repository.BaseRepository[customermodel.Customer]
}

// NewRepository 构造客户仓储。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepository: repository.NewBaseRepository[customermodel.Customer](db),
	}
}

// List 根据筛选条件返回客户分页数据。
func (r *Repository) List(
	ctx context.Context,
	opts ListQueryOptions,
) (*repository.Page[[]*customermodel.Customer], error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	conditions := map[string]interface{}{
		"tenant_uuid = ?": tenantUUID,
	}
	callback := func(db *gorm.DB, arg interface{}) *gorm.DB {
		filters, _ := arg.(ListQueryOptions)
		db = applyListFilters(db, filters)
		order := buildOrderClause(filters.Sort)
		return db.Order(order)
	}
	return r.BaseRepository.FindByCondition(ctx, conditions, opts.Page, opts.PageSize, callback, opts)
}

// FindByCustomerID 根据租户与业务 ID 查询客户。
func (r *Repository) FindByCustomerID(ctx context.Context, customerID string) (*customermodel.Customer, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var entity customermodel.Customer
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Where("customer_id = ?", customerID).
		First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// CreateCustomer 插入客户记录。
func (r *Repository) CreateCustomer(ctx context.Context, entity *customermodel.Customer) (*customermodel.Customer, error) {
	if entity == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(entity.TenantUUID) == "" {
		entity.TenantUUID = tenantUUID
	} else if !strings.EqualFold(strings.TrimSpace(entity.TenantUUID), tenantUUID) {
		return nil, gorm.ErrInvalidData
	}
	if err := r.DB.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

// SaveCustomer 保存客户记录。
func (r *Repository) SaveCustomer(ctx context.Context, entity *customermodel.Customer) (*customermodel.Customer, error) {
	if entity == nil {
		return nil, gorm.ErrInvalidData
	}
	if err := r.DB.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

// DeleteByCustomerID 删除客户记录。
func (r *Repository) DeleteByCustomerID(ctx context.Context, customerID string) error {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	return r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Where("customer_id = ?", customerID).
		Delete(&customermodel.Customer{}).Error
}

// DetectConflict 检查邮箱 / 手机唯一性冲突，返回冲突字段名。
func (r *Repository) DetectConflict(ctx context.Context, email, phone, excludeID string) (string, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return "", err
	}
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	normalizedPhone := strings.TrimSpace(phone)
	if normalizedEmail != "" {
		query := r.DB.WithContext(ctx).Model(&customermodel.Customer{}).
			Where("tenant_uuid = ?", tenantUUID).
			Where("LOWER(email) = ?", normalizedEmail)
		if excludeID != "" {
			query = query.Where("customer_id <> ?", excludeID)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return "", err
		}
		if count > 0 {
			return "email", nil
		}
	}
	if normalizedPhone != "" {
		query := r.DB.WithContext(ctx).Model(&customermodel.Customer{}).
			Where("tenant_uuid = ?", tenantUUID).
			Where("phone = ?", normalizedPhone)
		if excludeID != "" {
			query = query.Where("customer_id <> ?", excludeID)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return "", err
		}
		if count > 0 {
			return "phone", nil
		}
	}
	return "", nil
}

func applyListFilters(db *gorm.DB, filters ListQueryOptions) *gorm.DB {
	if keyword := strings.ToLower(strings.TrimSpace(filters.Keyword)); keyword != "" {
		pattern := "%" + keyword + "%"
		db = db.Where(
			"(LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone) LIKE ? OR LOWER(account_manager) LIKE ?)",
			pattern, pattern, pattern, pattern,
		)
	}
	if tier := strings.TrimSpace(filters.Tier); tier != "" {
		db = db.Where("LOWER(membership_tier) = ?", strings.ToLower(tier))
	}
	if source := strings.TrimSpace(filters.Source); source != "" {
		db = db.Where("LOWER(source) = ?", strings.ToLower(source))
	}
	if region := strings.TrimSpace(filters.Region); region != "" {
		pattern := "%" + strings.ToLower(region) + "%"
		db = db.Where("LOWER(region) LIKE ?", pattern)
	}
	if risk := strings.TrimSpace(filters.RiskLevel); risk != "" {
		db = db.Where("LOWER(risk_level) = ?", strings.ToLower(risk))
	}
	if customerType := strings.TrimSpace(filters.Type); customerType != "" {
		db = db.Where("LOWER(type) = ?", strings.ToLower(customerType))
	}
	if len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			pattern := "%\"" + tag + "\"%"
			db = db.Where("CAST(tags AS TEXT) LIKE ?", pattern)
		}
	}
	return db
}

func buildOrderClause(raw string) clause.OrderByColumn {
	descending := true
	field := strings.TrimSpace(raw)
	if field == "" {
		return clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}
	}
	if strings.HasPrefix(field, "-") {
		field = strings.TrimPrefix(field, "-")
		descending = true
	} else {
		descending = false
	}
	column := customerSortColumns[strings.ToLower(field)]
	if column == "" {
		column = "created_at"
		descending = true
	}
	return clause.OrderByColumn{Column: clause.Column{Name: column}, Desc: descending}
}

var customerSortColumns = map[string]string{
	"createdat":      "created_at",
	"updatedat":      "updated_at",
	"name":           "name",
	"lastorderat":    "last_order_at",
	"status":         "status",
	"tier":           "membership_tier",
	"membershiptier": "membership_tier",
}
