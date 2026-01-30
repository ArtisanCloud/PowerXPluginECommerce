package customer

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CustomerIdentity binds third-party identities (wechat/dingtalk/alipay) to customers.
type CustomerIdentity struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:varchar(64);not null;index:idx_customer_identity_tenant_provider_subject,priority:1;index:idx_customer_identity_tenant_customer,priority:1;comment:租户UUID" json:"tenant_uuid"`
	CustomerID string         `gorm:"type:varchar(64);not null;index:idx_customer_identity_tenant_customer,priority:2;comment:客户业务ID" json:"customer_id"`
	Provider   string         `gorm:"type:varchar(32);not null;index:idx_customer_identity_tenant_provider_subject,priority:2;comment:第三方来源(wechat/dingtalk/alipay)" json:"provider"`
	AppID      string         `gorm:"type:varchar(64);not null;index:idx_customer_identity_tenant_provider_subject,priority:3;comment:第三方应用ID(小程序AppID等)" json:"app_id"`
	Subject    string         `gorm:"type:varchar(128);not null;index:idx_customer_identity_tenant_provider_subject,priority:4;comment:第三方主体ID(微信openid等)" json:"subject"`
	UnionID    string         `gorm:"type:varchar(128);comment:微信unionid(同主体多应用统一ID)" json:"union_id"`
	Metadata   datatypes.JSON `gorm:"type:json" json:"metadata"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CustomerIdentity) TableName() string {
	return models.S(models.TableCustomerIdentity)
}
