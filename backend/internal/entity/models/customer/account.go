package customer

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// CustomerAccount 存储 mini-app 客户登录凭证。
type CustomerAccount struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:varchar(64);not null;index:idx_customer_account_tenant_identifier,priority:1" json:"tenant_uuid"`
	CustomerID   string         `gorm:"type:varchar(64);not null;index" json:"customer_id"`
	Identifier   string         `gorm:"type:varchar(255);not null;index:idx_customer_account_tenant_identifier,priority:2" json:"identifier"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Status       string         `gorm:"type:varchar(16);not null;default:'active'" json:"status"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CustomerAccount) TableName() string {
	return models.S("customer_accounts")
}
