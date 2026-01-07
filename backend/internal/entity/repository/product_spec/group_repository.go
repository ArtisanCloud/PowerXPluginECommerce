package product_spec

import (
	"context"
	"errors"
	"strings"

	productspecmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_spec"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// GroupRepository exposes CRUD helpers for ProductSpecGroup entities.
type GroupRepository struct {
	*repo.BaseRepository[productspecmodel.ProductSpecGroup]
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{BaseRepository: repo.NewBaseRepository[productspecmodel.ProductSpecGroup](db)}
}

func (r *GroupRepository) ListBySPU(ctx context.Context, tenantID, spuID string) ([]productspecmodel.ProductSpecGroup, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("spec group repository is not initialized")
	}
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(spuID) == "" {
		return []productspecmodel.ProductSpecGroup{}, nil
	}
	var rows []productspecmodel.ProductSpecGroup
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND spu_id = ? AND deleted_at IS NULL", tenantID, strings.TrimSpace(spuID)).
		Order("sort_order ASC, code ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

