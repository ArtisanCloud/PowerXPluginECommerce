package product

import (
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// VersionRepository stores SPU version snapshots.
type VersionRepository struct {
	*repo.BaseRepository[productmodel.SPUVersion]
}

// NewVersionRepository instantiates the repository with the shared DB connection.
func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{BaseRepository: repo.NewBaseRepository[productmodel.SPUVersion](db)}
}
