package repository

import (
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

type PaymentRiskEventRepository struct {
	*BaseRepository[models.PaymentRiskEvent]
}

func NewPaymentRiskEventRepository(db *gorm.DB) *PaymentRiskEventRepository {
	return &PaymentRiskEventRepository{BaseRepository: NewBaseRepository[models.PaymentRiskEvent](db)}
}
