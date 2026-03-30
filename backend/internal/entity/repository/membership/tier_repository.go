package membership

import (
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

type TierRepository struct {
	*repo.BaseRepository[membershipModel.MembershipTier]
}

func NewTierRepository(db *gorm.DB) *TierRepository {
	return &TierRepository{BaseRepository: repo.NewBaseRepository[membershipModel.MembershipTier](db)}
}
