package membership

import (
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

type EntitlementRepository struct {
	*repo.BaseRepository[membershipModel.Entitlement]
}

func NewEntitlementRepository(db *gorm.DB) *EntitlementRepository {
	return &EntitlementRepository{BaseRepository: repo.NewBaseRepository[membershipModel.Entitlement](db)}
}
