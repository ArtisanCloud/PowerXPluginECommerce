package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	pricingRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditService struct {
	*Service
}

func NewAuditService(deps *app.Deps) *AuditService {
	return &AuditService{Service: NewService(deps)}
}

func (s *AuditService) Log(ctx context.Context, tx *gorm.DB, resourceType, resourceID, action, actor string, payload any) error {
	if !s.Ready() {
		return E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	if strings.TrimSpace(resourceType) == "" || strings.TrimSpace(resourceID) == "" || strings.TrimSpace(action) == "" {
		return E(CodeInvalidArgument, errors.New("audit log requires resource_type/resource_id/action"))
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return E(CodeTenantMissing, err)
	}

	var raw datatypes.JSON
	if payload != nil {
		if buf, marshalErr := json.Marshal(payload); marshalErr == nil {
			raw = datatypes.JSON(buf)
		}
	}

	entry := &pricingModel.PricebookAuditLog{
		ID:           uuid.NewString(),
		TenantUUID:   tenantUUID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		Actor:        actor,
		Payload:      raw,
	}

	db := tx
	if db == nil {
		db = s.deps.DB
	}
	repo := pricingRepo.NewPricebookAuditLogRepository(db)
	_, err = repo.Create(ctx, entry)
	return err
}
