package master

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CredentialChecker periodically scans credential expiry.
type CredentialChecker struct {
	db           *gorm.DB
	alertRepo    *channelrepo.ChannelAlertRepository
	logger       *logrus.Entry
	alertEmitter *channelobs.AlertEmitter
	lead         time.Duration
	interval     time.Duration
}

// NewCredentialChecker constructs job.
func NewCredentialChecker(deps *app.Deps, lead, interval time.Duration, alertEmitter *channelobs.AlertEmitter) *CredentialChecker {
	logger := deps.RuntimeLogger(context.TODO(), "channel-credential-checker", nil)
	return &CredentialChecker{
		db:           deps.DB,
		alertRepo:    channelrepo.NewChannelAlertRepository(deps.DB),
		logger:       logger,
		alertEmitter: alertEmitter,
		lead:         lead,
		interval:     interval,
	}
}

// Run starts ticker until context cancelled.
func (c *CredentialChecker) Run(ctx context.Context) {
	if c == nil || c.db == nil {
		return
	}
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	c.scan(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.scan(ctx)
		}
	}
}

func (c *CredentialChecker) scan(ctx context.Context) {
	tenantIDs, err := c.listTenantUUIDs(ctx)
	if err != nil {
		c.logger.WithError(err).Warn("failed to list tenants for credential checker")
		return
	}
	for _, tenantID := range tenantIDs {
		tenantCtx := authx.ContextWithTenantUUID(ctx, tenantID)
		repo := channelrepo.NewChannelCredentialRepository(c.db)
		creds, err := repo.ListExpiring(tenantCtx, c.lead)
		if err != nil {
			c.logger.WithError(err).WithField("tenant", tenantID).Warn("credential scan failed")
			continue
		}
		for _, cred := range creds {
			c.raiseAlert(tenantCtx, cred)
		}
	}
}

func (c *CredentialChecker) raiseAlert(ctx context.Context, cred *channelmodel.ChannelCredential) {
	if cred == nil {
		return
	}
	expiry := ""
	if cred.ExpiresAt != nil {
		expiry = cred.ExpiresAt.Format(time.RFC3339)
	}
	alert := &channelmodel.ChannelAlert{
		ChannelID:   cred.ChannelID,
		Type:        channelmodel.AlertTypeCredentialExpiring,
		Title:       "渠道凭证即将到期",
		Description: "渠道凭证即将到期，请尽快刷新授权。",
		Severity:    "warning",
		TriggeredAt: time.Now().UTC(),
		Metadata:    datatypesJSON(map[string]any{"credential_id": cred.ID, "expires_at": expiry}),
		Status:      "open",
	}
	if _, err := c.alertRepo.UpsertStatus(ctx, alert); err != nil {
		c.logger.WithError(err).WithField("channel_id", cred.ChannelID).Warn("failed to persist credential alert")
	} else if c.alertEmitter != nil {
		c.alertEmitter.Emit(ctx, fmt.Sprintf("channel.%s", channelmodel.AlertTypeCredentialExpiring), map[string]any{
			"channel_id":    cred.ChannelID,
			"credential_id": cred.ID,
			"type":          channelmodel.AlertTypeCredentialExpiring,
			"tenant_uuid":   alert.TenantUUID,
		})
	}
}

func (c *CredentialChecker) listTenantUUIDs(ctx context.Context) ([]string, error) {
	var ids []string
	if err := c.db.WithContext(ctx).
		Model(&channelmodel.ChannelMaster{}).
		Where("tenant_uuid IS NOT NULL").
		Distinct().
		Pluck("tenant_uuid", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func datatypesJSON(m map[string]any) datatypes.JSON {
	if len(m) == 0 {
		return nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return datatypes.JSON(raw)
}
