package membership

import (
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

type BenefitRepository struct {
	*repo.BaseRepository[membershipModel.MembershipBenefit]
}

func NewBenefitRepository(db *gorm.DB) *BenefitRepository {
	return &BenefitRepository{BaseRepository: repo.NewBaseRepository[membershipModel.MembershipBenefit](db)}
}
