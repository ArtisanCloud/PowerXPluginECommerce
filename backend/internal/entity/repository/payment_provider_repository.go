package repository

import (
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

type PaymentProviderRepository struct {
	*BaseRepository[models.PaymentProvider]
}

func NewPaymentProviderRepository(db *gorm.DB) *PaymentProviderRepository {
	return &PaymentProviderRepository{BaseRepository: NewBaseRepository[models.PaymentProvider](db)}
}
