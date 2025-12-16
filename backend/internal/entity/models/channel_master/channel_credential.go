package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelCredential stores encrypted credential material per channel.
type ChannelCredential struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:凭证ID" json:"id"`
	TenantUUID       string         `gorm:"type:uuid;not null;index:idx_channel_cred_tenant,priority:1" json:"tenant_uuid"`
	ChannelID        string         `gorm:"type:uuid;not null;index:idx_channel_cred_tenant,priority:2" json:"channel_id"`
	Type             string         `gorm:"type:varchar(32);not null;index;comment:凭证类型" json:"type"`
	Status           string         `gorm:"type:varchar(32);not null;default:'valid';index;comment:凭证状态" json:"status"`
	Scope            pq.StringArray `gorm:"type:text[];comment:权限范围" json:"scope"`
	SecretCiphertext []byte         `gorm:"column:secret_ciphertext;type:bytea;not null" json:"-"`
	SecretNonce      []byte         `gorm:"column:secret_nonce;type:bytea;not null" json:"-"`
	DekCiphertext    []byte         `gorm:"column:dek_ciphertext;type:bytea;not null" json:"-"`
	DekNonce         []byte         `gorm:"column:dek_nonce;type:bytea;not null" json:"-"`
	Algorithm        string         `gorm:"type:varchar(32);not null;default:'AES-GCM'" json:"algorithm"`
	KeyVersion       string         `gorm:"type:varchar(32);not null;default:'static-v1'" json:"key_version,omitempty"`
	ExpiresAt        *time.Time     `json:"expires_at,omitempty"`
	LastRefreshedAt  *time.Time     `json:"last_refreshed_at,omitempty"`
	LastTestedAt     *time.Time     `json:"last_tested_at,omitempty"`
	LastRotatedAt    *time.Time     `json:"last_rotated_at,omitempty"`
	TestResult       datatypes.JSON `gorm:"type:jsonb;comment:最近一次测试结果" json:"test_result,omitempty"`
	Metadata         datatypes.JSON `gorm:"type:jsonb;comment:额外字段" json:"metadata,omitempty"`
	AttachmentURL    string         `gorm:"type:text;comment:线下凭证附件" json:"attachment_url,omitempty"`
	CreatedBy        string         `gorm:"type:varchar(64);comment:创建人" json:"created_by,omitempty"`
	UpdatedBy        string         `gorm:"type:varchar(64);comment:更新人" json:"updated_by,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChannelCredential) TableName() string {
	return models.S(models.TableChannelCredentials)
}
