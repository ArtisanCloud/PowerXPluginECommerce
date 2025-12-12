package product

import (
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// ApprovalRepository exposes CRUD helpers for SPU approval records.
type ApprovalRepository struct {
	*repo.BaseRepository[productmodel.SPUApprovalRecord]
}

// NewApprovalRepository builds a repository backed by BaseRepository.
func NewApprovalRepository(db *gorm.DB) *ApprovalRepository {
	return &ApprovalRepository{BaseRepository: repo.NewBaseRepository[productmodel.SPUApprovalRecord](db)}
}
