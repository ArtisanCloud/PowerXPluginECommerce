package repository

import (
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

type PaymentRefundRepository struct {
	*BaseRepository[models.PaymentRefund]
}

func NewPaymentRefundRepository(db *gorm.DB) *PaymentRefundRepository {
	return &PaymentRefundRepository{BaseRepository: NewBaseRepository[models.PaymentRefund](db)}
}
