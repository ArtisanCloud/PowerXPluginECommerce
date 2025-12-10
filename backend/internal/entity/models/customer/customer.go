package customer

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Customer represents the persisted CRM customer record.
type Customer struct {
	ID                  uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantUUID          string         `gorm:"column:tenant_uuid;type:varchar(64);not null;index;comment:租户UUID" json:"tenant_uuid"`
	CustomerID          string         `gorm:"type:varchar(64);not null;uniqueIndex;comment:业务侧客户ID" json:"customer_id"`
	Name                string         `gorm:"type:varchar(255);not null;comment:客户名称" json:"name"`
	Type                string         `gorm:"type:varchar(32);not null;comment:客户类型" json:"type"`
	Email               string         `gorm:"type:varchar(255);comment:邮箱" json:"email"`
	Phone               string         `gorm:"type:varchar(64);comment:电话" json:"phone"`
	Country             string         `gorm:"type:varchar(32);comment:国家" json:"country"`
	Region              string         `gorm:"type:varchar(255);comment:区域" json:"region"`
	MembershipTier      string         `gorm:"type:varchar(64);comment:会员等级" json:"membership_tier"`
	MembershipTierLabel string         `gorm:"type:varchar(64);comment:会员等级标签" json:"membership_tier_label"`
	GrowthValue         int            `gorm:"comment:成长值" json:"growth_value"`
	Points              int            `gorm:"comment:积分" json:"points"`
	LastOrderAmount     float64        `gorm:"type:numeric(18,2);comment:最近订单金额" json:"last_order_amount"`
	LastOrderAt         *time.Time     `gorm:"comment:最近下单时间" json:"last_order_at"`
	Status              string         `gorm:"type:varchar(32);comment:状态" json:"status"`
	RiskLevel           string         `gorm:"type:varchar(32);comment:风险等级" json:"risk_level"`
	Source              string         `gorm:"type:varchar(64);comment:来源渠道" json:"source"`
	AccountManager      string         `gorm:"type:varchar(128);comment:负责人" json:"account_manager"`
	Tags                datatypes.JSON `gorm:"type:json" json:"tags"`
	MaskedFields        datatypes.JSON `gorm:"type:json" json:"masked_fields"`
	MembershipSnapshot  datatypes.JSON `gorm:"type:json" json:"membership_snapshot"`
	Metadata            datatypes.JSON `gorm:"type:json" json:"metadata"`
	Notes               string         `gorm:"type:text;comment:备注" json:"notes"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Customer) TableName() string {
	return models.S(models.TableCustomer)
}
