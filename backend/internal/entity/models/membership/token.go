package membership

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TokenAccount stores customer token balance.
type TokenAccount struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index;uniqueIndex:uk_token_account,priority:1" json:"tenant_uuid"`
	CustomerID string         `gorm:"type:uuid;not null;index;uniqueIndex:uk_token_account,priority:2" json:"customer_id"`
	TokenCode  string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_token_account,priority:3" json:"token_code"`
	Balance    int64          `gorm:"not null;default:0" json:"balance"`
	Metadata   datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TokenAccount) TableName() string { return basemodels.S(basemodels.TableTokenAccounts) }

// TokenTransaction records token grants/consumption.
type TokenTransaction struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string    `gorm:"type:uuid;not null;index;uniqueIndex:uk_token_txn_grant,priority:1" json:"tenant_uuid"`
	CustomerID string    `gorm:"type:uuid;not null;index" json:"customer_id"`
	TokenCode  string    `gorm:"type:varchar(64);not null;index;uniqueIndex:uk_token_txn_grant,priority:4" json:"token_code"`
	Delta      int64     `gorm:"not null" json:"delta"`
	SourceType string    `gorm:"type:varchar(32);not null;uniqueIndex:uk_token_txn_grant,priority:2" json:"source_type"`
	SourceID   string    `gorm:"type:varchar(128);not null;uniqueIndex:uk_token_txn_grant,priority:3" json:"source_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (TokenTransaction) TableName() string { return basemodels.S(basemodels.TableTokenTransactions) }
