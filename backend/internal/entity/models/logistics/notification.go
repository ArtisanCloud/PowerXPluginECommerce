package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// NotificationTemplate stores logistics-message templates by event/channel.
type NotificationTemplate struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_notify_tpl,priority:1" json:"tenant_uuid"`
	Name       string         `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Event      string         `gorm:"column:event;type:varchar(32);not null;index;uniqueIndex:uk_logistics_notify_tpl,priority:2" json:"event"`
	Channel    string         `gorm:"column:channel;type:varchar(32);not null;default:'sms';uniqueIndex:uk_logistics_notify_tpl,priority:3" json:"channel"`
	Title      string         `gorm:"column:title;type:varchar(256)" json:"title,omitempty"`
	Body       string         `gorm:"column:body;type:text;not null" json:"body"`
	Enabled    bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Metadata   datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (NotificationTemplate) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsNotificationTemplates)
}

// NotificationRecord stores delivery attempts and idempotent send logs.
type NotificationRecord struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_notify_idem,priority:1" json:"tenant_uuid"`
	TemplateID     string         `gorm:"column:template_id;type:uuid;not null;index" json:"template_id"`
	WaybillID      string         `gorm:"column:waybill_id;type:uuid;not null;index" json:"waybill_id"`
	Event          string         `gorm:"column:event;type:varchar(32);not null;index" json:"event"`
	Channel        string         `gorm:"column:channel;type:varchar(32);not null" json:"channel"`
	Status         string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	AttemptCount   int            `gorm:"column:attempt_count;type:int;not null;default:0" json:"attempt_count"`
	MaxAttempts    int            `gorm:"column:max_attempts;type:int;not null;default:3" json:"max_attempts"`
	IdempotencyKey string         `gorm:"column:idempotency_key;type:varchar(191);not null;index;uniqueIndex:uk_logistics_notify_idem,priority:2" json:"idempotency_key"`
	LastError      string         `gorm:"column:last_error;type:text" json:"last_error,omitempty"`
	RenderedTitle  string         `gorm:"column:rendered_title;type:varchar(256)" json:"rendered_title,omitempty"`
	RenderedBody   string         `gorm:"column:rendered_body;type:text" json:"rendered_body,omitempty"`
	Payload        datatypes.JSON `gorm:"column:payload;type:jsonb" json:"payload,omitempty"`
	SentAt         *time.Time     `gorm:"column:sent_at" json:"sent_at,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (NotificationRecord) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsNotificationRecords)
}
