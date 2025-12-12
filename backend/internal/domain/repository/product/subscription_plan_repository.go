package product

import (
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// SubscriptionPlanRepository handles CRUD for recurring plan records.
type SubscriptionPlanRepository struct {
	*repo.BaseRepository[productmodel.SubscriptionPlan]
}

// NewSubscriptionPlanRepository wires the repository to the shared DB connection.
func NewSubscriptionPlanRepository(db *gorm.DB) *SubscriptionPlanRepository {
	return &SubscriptionPlanRepository{BaseRepository: repo.NewBaseRepository[productmodel.SubscriptionPlan](db)}
}
