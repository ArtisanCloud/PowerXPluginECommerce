package repository

import (
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

type PaymentSplitRuleRepository struct {
	*BaseRepository[models.PaymentSplitRule]
}

func NewPaymentSplitRuleRepository(db *gorm.DB) *PaymentSplitRuleRepository {
	return &PaymentSplitRuleRepository{BaseRepository: NewBaseRepository[models.PaymentSplitRule](db)}
}
