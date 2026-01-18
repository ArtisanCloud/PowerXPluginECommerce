package repository

import (
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

type PaymentReconciliationRepository struct {
	*BaseRepository[models.PaymentReconciliation]
}

func NewPaymentReconciliationRepository(db *gorm.DB) *PaymentReconciliationRepository {
	return &PaymentReconciliationRepository{BaseRepository: NewBaseRepository[models.PaymentReconciliation](db)}
}

type PaymentReconciliationItemRepository struct {
	*BaseRepository[models.PaymentReconciliationItem]
}

func NewPaymentReconciliationItemRepository(db *gorm.DB) *PaymentReconciliationItemRepository {
	return &PaymentReconciliationItemRepository{BaseRepository: NewBaseRepository[models.PaymentReconciliationItem](db)}
}
