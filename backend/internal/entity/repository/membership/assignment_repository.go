package membership

import (
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

type AssignmentRepository struct {
	*repo.BaseRepository[membershipModel.MembershipAssignment]
}

func NewAssignmentRepository(db *gorm.DB) *AssignmentRepository {
	return &AssignmentRepository{BaseRepository: repo.NewBaseRepository[membershipModel.MembershipAssignment](db)}
}
