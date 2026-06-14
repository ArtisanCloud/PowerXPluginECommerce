package promotion

import (
	"context"
	"encoding/json"
	"strings"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	promotionrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditLogService struct {
	deps *app.Deps
	repo *promotionrepo.AuditLogRepository
}

func NewAuditLogService(deps *app.Deps) *AuditLogService {
	if deps == nil || deps.DB == nil {
		return &AuditLogService{deps: deps}
	}
	return &AuditLogService{deps: deps, repo: promotionrepo.NewAuditLogRepository(deps.DB)}
}

func (s *AuditLogService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *AuditLogService) Log(ctx context.Context, db *gorm.DB, tenantUUID, promotionID, orderID, action, reason, actor, requestID string, payload any) error {
	if !s.Ready() {
		return nil
	}
	if db == nil {
		db = s.deps.DB
	}
	raw := datatypes.JSON([]byte("{}"))
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err == nil {
			raw = datatypes.JSON(encoded)
		}
	}
	var orderPtr *string
	if v := strings.TrimSpace(orderID); v != "" {
		orderPtr = &v
	}
	row := &promotionmodel.AuditLog{
		TenantUUID: strings.TrimSpace(tenantUUID), PromotionID: strings.TrimSpace(promotionID), OrderID: orderPtr,
		Action: strings.TrimSpace(action), ActionReason: strings.TrimSpace(reason), RequestID: strings.TrimSpace(requestID),
		CreatedBy: strings.TrimSpace(actor), Payload: raw,
	}
	return db.WithContext(ctx).Create(row).Error
}

func (s *AuditLogService) List(ctx context.Context, tenantUUID, promotionID string, page, pageSize int) ([]promotionmodel.AuditLog, int64, error) {
	if !s.Ready() {
		return nil, 0, gorm.ErrInvalidDB
	}
	return s.repo.ListByPromotion(ctx, tenantUUID, promotionID, page, pageSize)
}
