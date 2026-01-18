package repository

import (
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

type PaymentTransactionRepository struct {
	*BaseRepository[models.PaymentTransaction]
}

func NewPaymentTransactionRepository(db *gorm.DB) *PaymentTransactionRepository {
	return &PaymentTransactionRepository{BaseRepository: NewBaseRepository[models.PaymentTransaction](db)}
}
