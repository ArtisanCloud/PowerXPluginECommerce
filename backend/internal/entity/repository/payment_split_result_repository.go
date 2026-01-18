package repository

import (
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

type PaymentSplitResultRepository struct {
	*BaseRepository[models.PaymentSplitResult]
}

func NewPaymentSplitResultRepository(db *gorm.DB) *PaymentSplitResultRepository {
	return &PaymentSplitResultRepository{BaseRepository: NewBaseRepository[models.PaymentSplitResult](db)}
}
