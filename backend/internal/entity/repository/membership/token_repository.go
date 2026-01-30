package membership

import (
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

type TokenAccountRepository struct {
	*repo.BaseRepository[membershipModel.TokenAccount]
}

func NewTokenAccountRepository(db *gorm.DB) *TokenAccountRepository {
	return &TokenAccountRepository{BaseRepository: repo.NewBaseRepository[membershipModel.TokenAccount](db)}
}

type TokenTransactionRepository struct {
	*repo.BaseRepository[membershipModel.TokenTransaction]
}

func NewTokenTransactionRepository(db *gorm.DB) *TokenTransactionRepository {
	return &TokenTransactionRepository{BaseRepository: repo.NewBaseRepository[membershipModel.TokenTransaction](db)}
}
