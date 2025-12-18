package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// ChannelMetric keeps KPI snapshots per time window.
type ChannelMetric struct {
	ID                string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID        string         `gorm:"type:uuid;not null;index:idx_channel_metric_tenant,priority:1" json:"tenant_uuid"`
	ChannelID         string         `gorm:"type:uuid;not null;index:idx_channel_metric_tenant,priority:2" json:"channel_id"`
	Window            string         `gorm:"type:varchar(8);not null;comment:d1/d7/d30" json:"window"`
	GMV               float64        `gorm:"type:numeric(20,2);default:0" json:"gmv"`
	Orders            int64          `gorm:"type:bigint;default:0" json:"orders"`
	GMVGrowthRate     float64        `gorm:"type:numeric(10,4);default:0" json:"gmv_growth_rate"`
	InventoryCoverage float64        `gorm:"type:numeric(10,4);default:0" json:"inventory_coverage"`
	ErrorRate         float64        `gorm:"type:numeric(10,4);default:0" json:"error_rate"`
	SyncSuccessRate   float64        `gorm:"type:numeric(10,4);default:0" json:"sync_success_rate"`
	HealthScore       int            `gorm:"type:int;default:0" json:"health_score"`
	SourceTimestamp   *time.Time     `json:"source_timestamp,omitempty"`
	SourceJobID       string         `gorm:"type:varchar(128)" json:"source_job_id,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChannelMetric) TableName() string {
	return models.S(models.TableChannelMetrics)
}
