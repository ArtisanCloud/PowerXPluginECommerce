package product

import (
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// ChannelRepository manages ChannelVisibility records for SPUs.
type ChannelRepository struct {
	*repo.BaseRepository[productmodel.ChannelVisibility]
}

// NewChannelRepository returns a repository bound to the provided DB handle.
func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{BaseRepository: repo.NewBaseRepository[productmodel.ChannelVisibility](db)}
}
