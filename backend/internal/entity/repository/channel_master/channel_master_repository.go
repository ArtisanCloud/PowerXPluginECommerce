package repository

import (
	"context"
	"errors"
	"strings"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelListFilters enumerates filters for channel listing endpoints.
type ChannelListFilters struct {
	Keyword   string
	Platform  string
	Status    []string
	Owner     string
	Region    string
	Tags      []string
	MinHealth int
	MinGMV    float64
	Page      int
	PageSize  int
}

// ChannelMasterRepository wraps BaseRepository with tenant-aware helpers.
type ChannelMasterRepository struct {
	*repo.BaseRepository[channelmodel.ChannelMaster]
}

// NewChannelMasterRepository constructs a repository backed by the shared DB connection.
func NewChannelMasterRepository(db *gorm.DB) *ChannelMasterRepository {
	return &ChannelMasterRepository{BaseRepository: repo.NewBaseRepository[channelmodel.ChannelMaster](db)}
}

// UniqueConstraintColumns returns the composite keys enforced via upsert logic.
func (r *ChannelMasterRepository) UniqueConstraintColumns() []clause.Column {
	return []clause.Column{{Name: "tenant_uuid"}, {Name: "store_id"}}
}

// BeginTenantTx ensures downstream operations run with SET LOCAL app.tenant_uuid.
func (r *ChannelMasterRepository) BeginTenantTx(ctx context.Context) (*gorm.DB, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	return r.BaseRepository.BeginTenantTx(ctx, tenantUUID)
}

// Create inserts a channel record after verifying tenant scope.
func (r *ChannelMasterRepository) Create(ctx context.Context, channel *channelmodel.ChannelMaster) (*channelmodel.ChannelMaster, error) {
	if channel == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(channel.TenantUUID) == "" {
		channel.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, channel)
}

// Upsert persists a channel using tenant+store_id as uniqueness guard.
func (r *ChannelMasterRepository) Upsert(ctx context.Context, channel *channelmodel.ChannelMaster) (*channelmodel.ChannelMaster, error) {
	if channel == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(channel.TenantUUID) == "" {
		channel.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Upsert(ctx, channel, r.UniqueConstraintColumns())
}

// FindPage returns paginated channel masters given filters.
func (r *ChannelMasterRepository) FindPage(ctx context.Context, filters ChannelListFilters) (*repo.Page[[]*channelmodel.ChannelMaster], error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	conditions := map[string]interface{}{
		"tenant_uuid = ?": tenantUUID,
	}
	return r.BaseRepository.FindByCondition(ctx, conditions, filters.Page, filters.PageSize, func(db *gorm.DB, opt interface{}) *gorm.DB {
		f := opt.(ChannelListFilters)
		if kw := strings.TrimSpace(f.Keyword); kw != "" {
			pattern := "%" + kw + "%"
			db = db.Where("name ILIKE ? OR owner_uuid ILIKE ? OR store_id ILIKE ?", pattern, pattern, pattern)
		}
		if strings.TrimSpace(f.Platform) != "" {
			db = db.Where("platform = ?", f.Platform)
		}
		if len(f.Status) > 0 {
			db = db.Where("status IN ?", f.Status)
		}
		if strings.TrimSpace(f.Owner) != "" {
			db = db.Where("owner_uuid = ?", f.Owner)
		}
		if strings.TrimSpace(f.Region) != "" {
			db = db.Where("region = ?", f.Region)
		}
		if arr := normalizeTags(f.Tags); len(arr) > 0 {
			db = db.Where("tags && ?", pq.StringArray(arr))
		}
		if f.MinHealth > 0 {
			db = db.Where("health_score >= ?", f.MinHealth)
		}
		if f.MinGMV > 0 {
			db = applyMinGMVFilter(db, f.MinGMV)
		}
		return db.Order("updated_at DESC")
	}, filters)
}

// FindByID returns a channel scoped to the current tenant.
func (r *ChannelMasterRepository) FindByID(ctx context.Context, id string) (*channelmodel.ChannelMaster, error) {
	if strings.TrimSpace(id) == "" {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var entity channelmodel.ChannelMaster
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Save persists modifications to a channel entity under the tenant scope.
func (r *ChannelMasterRepository) Save(ctx context.Context, channel *channelmodel.ChannelMaster) (*channelmodel.ChannelMaster, error) {
	if channel == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(channel.TenantUUID) == "" {
		channel.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Update(ctx, channel)
}

// FindByStoreID returns the channel with the given store identifier within a tenant.
func (r *ChannelMasterRepository) FindByStoreID(ctx context.Context, storeID string) (*channelmodel.ChannelMaster, error) {
	if strings.TrimSpace(storeID) == "" {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var entity channelmodel.ChannelMaster
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND store_id = ?", tenantUUID, strings.TrimSpace(storeID)).
		First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func normalizeTags(values []string) []string {
	clean := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	return clean
}

func applyMinGMVFilter(db *gorm.DB, minGmv float64) *gorm.DB {
	if db == nil || minGmv <= 0 {
		return db
	}
	dialect := strings.ToLower(db.Dialector.Name())
	switch dialect {
	case "sqlite":
		return db.Where("COALESCE(json_extract(metadata, '$.gmv_30d'), 0) >= ?", minGmv)
	case "postgres":
		return db.Where("(metadata->>'gmv_30d')::numeric >= ?", minGmv)
	default:
		return db.Where("(metadata->>'gmv_30d')::numeric >= ?", minGmv)
	}
}
