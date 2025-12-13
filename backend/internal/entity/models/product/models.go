package product

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SPU captures the tenant-scoped catalog entity that groups SKUs.
type SPU struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:SPU 主键" json:"id"`
	TenantUUID       string         `gorm:"type:uuid;not null;index:idx_spu_tenant_code,priority:1;comment:租户" json:"tenant_uuid"`
	Code             string         `gorm:"type:varchar(120);not null;index:idx_spu_tenant_code,priority:2;comment:业务编码" json:"code"`
	Name             string         `gorm:"type:varchar(255);not null;comment:默认语言名称" json:"name"`
	Type             string         `gorm:"type:varchar(32);not null;comment:商品类型" json:"type"`
	CategoryID       string         `gorm:"type:varchar(64);not null;comment:类目ID" json:"category_id"`
	CategoryPath     string         `gorm:"type:text;not null;comment:完整类目路径" json:"category_path"`
	BrandID          string         `gorm:"type:varchar(64);comment:品牌ID" json:"brand_id,omitempty"`
	DefaultLocale    string         `gorm:"type:varchar(16);not null;comment:默认语言" json:"default_locale"`
	Status           string         `gorm:"type:varchar(32);not null;index;comment:状态" json:"status"`
	CurrentVersionID *string        `gorm:"type:uuid;comment:当前发布版本" json:"current_version_id,omitempty"`
	Tags             pq.StringArray `gorm:"type:text[];comment:标签" json:"tags,omitempty"`
	ResponsibleUser  string         `gorm:"type:varchar(64);comment:负责人" json:"responsible_user,omitempty"`
	ChannelsSummary  datatypes.JSON `gorm:"type:jsonb;comment:渠道摘要" json:"channels_summary,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SPU) TableName() string { return models.S(models.TableProductSpus) }

// SPUVersion stores full snapshots for draft/reviewed releases.
type SPUVersion struct {
	ID                    string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID            string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SPUID                 string         `gorm:"column:spu_id;type:uuid;not null;index" json:"spu_id"`
	VersionNumber         int            `gorm:"not null;comment:版本号" json:"version_number"`
	Status                string         `gorm:"type:varchar(32);not null;index" json:"status"`
	Payload               datatypes.JSON `gorm:"type:jsonb;not null;comment:版本快照" json:"payload"`
	DiffSummary           datatypes.JSON `gorm:"type:jsonb;comment:字段差异" json:"diff_summary,omitempty"`
	SubmittedBy           string         `gorm:"type:varchar(64);comment:提交人" json:"submitted_by,omitempty"`
	SubmittedAt           *time.Time     `json:"submitted_at,omitempty"`
	ApprovedBy            string         `gorm:"type:varchar(64);comment:审批人" json:"approved_by,omitempty"`
	ApprovedAt            *time.Time     `json:"approved_at,omitempty"`
	RollbackSourceVersion *string        `gorm:"type:uuid;comment:回滚来源" json:"rollback_source_version,omitempty"`
	AuditLogID            *string        `gorm:"type:uuid;comment:审计ID" json:"audit_log_id,omitempty"`
	CreatedAt             time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SPUVersion) TableName() string { return models.S(models.TableProductSpuVersions) }

// SPULocale persists per-language content authored for an SPU.
type SPULocale struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SPUID       string         `gorm:"column:spu_id;type:uuid;not null;index" json:"spu_id"`
	Locale      string         `gorm:"type:varchar(16);not null;index:idx_spu_locale_unique,priority:1" json:"locale"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Subtitle    string         `gorm:"type:varchar(255);comment:副标题" json:"subtitle,omitempty"`
	Description string         `gorm:"type:text;comment:描述" json:"description,omitempty"`
	Attributes  datatypes.JSON `gorm:"type:jsonb;comment:多语言属性" json:"attributes,omitempty"`
	Status      string         `gorm:"type:varchar(16);default:'active'" json:"status"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (SPULocale) TableName() string { return models.S(models.TableProductSpuLocales) }

// ChannelVisibility describes per-channel publishing windows.
type ChannelVisibility struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SPUID           string         `gorm:"column:spu_id;type:uuid;not null;index" json:"spu_id"`
	Channel         string         `gorm:"type:varchar(64);not null;index:idx_spu_channel_unique,priority:1" json:"channel"`
	Availability    string         `gorm:"type:varchar(32);not null;index" json:"availability"`
	PublishAt       *time.Time     `json:"publish_at,omitempty"`
	WithdrawAt      *time.Time     `json:"withdraw_at,omitempty"`
	ContentOverride datatypes.JSON `gorm:"type:jsonb;comment:内容差异" json:"content_override,omitempty"`
	AuditState      string         `gorm:"type:varchar(32);default:'pending'" json:"audit_state"`
	LastFeedback    datatypes.JSON `gorm:"type:jsonb" json:"last_feedback,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ChannelVisibility) TableName() string { return models.S(models.TableProductSpuChannels) }

// SubscriptionPlan configures recurring billing details.
type SubscriptionPlan struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SPUID        string         `gorm:"column:spu_id;type:uuid;not null;index" json:"spu_id"`
	PlanCode     string         `gorm:"type:varchar(64);not null;index:idx_spu_plan_unique,priority:1" json:"plan_code"`
	Name         string         `gorm:"type:varchar(120);not null" json:"name"`
	BillingCycle string         `gorm:"type:varchar(32);not null" json:"billing_cycle"`
	BillingValue int            `gorm:"comment:自定义周期天数" json:"billing_value,omitempty"`
	Price        float64        `gorm:"type:numeric(18,4);not null" json:"price"`
	Currency     string         `gorm:"type:varchar(16);default:'CNY'" json:"currency"`
	TrialDays    int            `gorm:"comment:试用天数" json:"trial_days,omitempty"`
	AutoRenew    bool           `gorm:"not null;default:true" json:"auto_renew"`
	CancelPolicy string         `gorm:"type:varchar(32);not null" json:"cancel_policy"`
	EffectScope  string         `gorm:"type:varchar(32);not null;default:'new_only'" json:"effect_scope"`
	Status       string         `gorm:"type:varchar(32);not null;default:'active'" json:"status"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SubscriptionPlan) TableName() string {
	return models.S(models.TableProductSpuSubscriptionPlans)
}

// SPUImportTask records asynchronous import executions.
type SPUImportTask struct {
	TaskID       string         `gorm:"primaryKey;type:varchar(64)" json:"task_id"`
	TenantUUID   string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	Template     string         `gorm:"type:varchar(64)" json:"template"`
	SuccessRows  int            `json:"success_rows"`
	FailedRows   int            `json:"failed_rows"`
	FailedReport string         `gorm:"type:text" json:"failed_report,omitempty"`
	Status       string         `gorm:"type:varchar(32);not null" json:"status"`
	Initiator    string         `gorm:"type:varchar(64);not null" json:"initiator"`
	Detail       datatypes.JSON `gorm:"type:jsonb" json:"detail,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
}

func (SPUImportTask) TableName() string {
	return models.S(models.TableProductSpuImportTasks)
}

// SPUExportTask mirrors import tasks for export jobs.
type SPUExportTask struct {
	TaskID      string         `gorm:"primaryKey;type:varchar(64)" json:"task_id"`
	TenantUUID  string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	Filters     datatypes.JSON `gorm:"type:jsonb" json:"filters,omitempty"`
	Fields      pq.StringArray `gorm:"type:text[]" json:"fields,omitempty"`
	Status      string         `gorm:"type:varchar(32);not null" json:"status"`
	Initiator   string         `gorm:"type:varchar(64);not null" json:"initiator"`
	DownloadURL string         `gorm:"type:text" json:"download_url,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

func (SPUExportTask) TableName() string {
	return models.S(models.TableProductSpuExportTasks)
}

// SPUApprovalRecord stores stage-level approval metadata.
type SPUApprovalRecord struct {
	ID         string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string     `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SPUID      string     `gorm:"column:spu_id;type:uuid;not null;index" json:"spu_id"`
	VersionID  string     `gorm:"type:uuid;not null;index" json:"version_id"`
	ChainOrder int        `gorm:"not null" json:"chain_order"`
	Role       string     `gorm:"type:varchar(32);not null" json:"role"`
	Status     string     `gorm:"type:varchar(32);not null" json:"status"`
	Comment    string     `gorm:"type:text" json:"comment,omitempty"`
	SLADueAt   *time.Time `json:"sla_due_at,omitempty"`
	ActedBy    string     `gorm:"type:varchar(64)" json:"acted_by,omitempty"`
	ActedAt    *time.Time `json:"acted_at,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SPUApprovalRecord) TableName() string {
	return models.S(models.TableProductSpuApprovals)
}

// SPUAuditLog captures immutable actions for compliance.
type SPUAuditLog struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SPUID      string         `gorm:"column:spu_id;type:uuid;not null;index" json:"spu_id"`
	EventType  string         `gorm:"type:varchar(64);not null" json:"event_type"`
	Payload    datatypes.JSON `gorm:"type:jsonb" json:"payload,omitempty"`
	Operator   string         `gorm:"type:varchar(64)" json:"operator"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (SPUAuditLog) TableName() string {
	return models.S(models.TableProductSpuAuditLogs)
}
