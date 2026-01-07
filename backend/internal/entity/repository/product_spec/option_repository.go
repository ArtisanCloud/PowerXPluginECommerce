package product_spec

import (
	"context"
	"errors"
	"strings"

	productspecmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_spec"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// OptionRepository exposes CRUD helpers for ProductSpecOption entities.
type OptionRepository struct {
	*repo.BaseRepository[productspecmodel.ProductSpecOption]
}

func NewOptionRepository(db *gorm.DB) *OptionRepository {
	return &OptionRepository{BaseRepository: repo.NewBaseRepository[productspecmodel.ProductSpecOption](db)}
}

func (r *OptionRepository) ListByGroupIDs(ctx context.Context, tenantID string, groupIDs []string) ([]productspecmodel.ProductSpecOption, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("spec option repository is not initialized")
	}
	if strings.TrimSpace(tenantID) == "" || len(groupIDs) == 0 {
		return []productspecmodel.ProductSpecOption{}, nil
	}
	var rows []productspecmodel.ProductSpecOption
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND group_id IN ? AND deleted_at IS NULL", tenantID, groupIDs).
		Order("sort_order ASC, code ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

