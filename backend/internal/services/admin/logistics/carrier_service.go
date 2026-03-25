package logistics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	LogisticsObs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

// CarrierService handles carrier administration operations.
type CarrierService struct {
	deps        *app.Deps
	carrierRepo *LogisticsRepo.CarrierRepository
	emitter     *LogisticsObs.Emitter
}

func NewCarrierService(deps *app.Deps) *CarrierService {
	if deps == nil || deps.DB == nil {
		return &CarrierService{}
	}
	return &CarrierService{
		deps:        deps,
		carrierRepo: LogisticsRepo.NewCarrierRepository(deps.DB),
		emitter:     LogisticsObs.NewEmitter(deps.RuntimeLogger(context.Background(), "logistics-carrier", nil)),
	}
}

type UpsertCarrierRequest struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name"`
	Code         string         `json:"code"`
	Type         string         `json:"type,omitempty"`
	Status       string         `json:"status,omitempty"`
	ContactName  string         `json:"contact_name,omitempty"`
	ContactPhone string         `json:"contact_phone,omitempty"`
	Capabilities map[string]any `json:"capabilities,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

type TestCarrierResult struct {
	CarrierID string `json:"carrier_id"`
	Reachable bool   `json:"reachable"`
	Message   string `json:"message"`
}

func (s *CarrierService) List(ctx context.Context, tenantUUID string) ([]LogisticsModel.Carrier, error) {
	if s == nil || s.carrierRepo == nil {
		return nil, errors.New("carrier service unavailable")
	}
	return s.carrierRepo.List(withTenantContext(ctx, tenantUUID))
}

func (s *CarrierService) Upsert(ctx context.Context, tenantUUID string, req UpsertCarrierRequest) (*LogisticsModel.Carrier, error) {
	if s == nil || s.carrierRepo == nil {
		return nil, errors.New("carrier service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	code := strings.TrimSpace(strings.ToLower(req.Code))
	if name == "" || code == "" {
		return nil, errors.New("name/code required")
	}
	capabilitiesBytes, _ := jsonBytes(req.Capabilities, []byte("{}"))
	configBytes, _ := jsonBytes(req.Config, []byte("{}"))

	if strings.TrimSpace(req.ID) != "" {
		carrier, err := s.carrierRepo.GetByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		carrier.Name = name
		carrier.Code = code
		if strings.TrimSpace(req.Type) != "" {
			carrier.Type = strings.TrimSpace(req.Type)
		}
		if strings.TrimSpace(req.Status) != "" {
			carrier.Status = strings.TrimSpace(req.Status)
		}
		carrier.ContactName = strings.TrimSpace(req.ContactName)
		carrier.ContactPhone = strings.TrimSpace(req.ContactPhone)
		carrier.Capabilities = datatypes.JSON(capabilitiesBytes)
		carrier.Config = datatypes.JSON(configBytes)
		if err := s.carrierRepo.Save(ctx, carrier); err != nil {
			return nil, err
		}
		s.emitAudit(ctx, tenantUUID, "carrier.update", carrier.ID, "updated")
		return carrier, nil
	}

	carrier := &LogisticsModel.Carrier{
		ID:           utils.NewUUID(),
		Name:         name,
		Code:         code,
		Type:         defaultString(req.Type, "self"),
		Status:       defaultString(req.Status, "active"),
		ContactName:  strings.TrimSpace(req.ContactName),
		ContactPhone: strings.TrimSpace(req.ContactPhone),
		Capabilities: datatypes.JSON(capabilitiesBytes),
		Config:       datatypes.JSON(configBytes),
	}
	if err := s.carrierRepo.Create(ctx, carrier); err != nil {
		return nil, err
	}
	s.emitAudit(ctx, tenantUUID, "carrier.create", carrier.ID, "created")
	return carrier, nil
}

func (s *CarrierService) Disable(ctx context.Context, tenantUUID, carrierID string) (*LogisticsModel.Carrier, error) {
	if s == nil || s.carrierRepo == nil {
		return nil, errors.New("carrier service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	carrier, err := s.carrierRepo.GetByID(ctx, carrierID)
	if err != nil {
		return nil, err
	}
	carrier.Status = "disabled"
	if err := s.carrierRepo.Save(ctx, carrier); err != nil {
		return nil, err
	}
	s.emitAudit(ctx, tenantUUID, "carrier.disable", carrier.ID, "disabled")
	return carrier, nil
}

func (s *CarrierService) TestConnectivity(ctx context.Context, tenantUUID, carrierID string) (*TestCarrierResult, error) {
	if s == nil || s.carrierRepo == nil {
		return nil, errors.New("carrier service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	carrier, err := s.carrierRepo.GetByID(ctx, carrierID)
	if err != nil {
		return nil, err
	}
	reachable := strings.EqualFold(carrier.Status, "active")
	msg := "carrier reachable"
	if !reachable {
		msg = "carrier is not active"
	}
	result := &TestCarrierResult{CarrierID: carrier.ID, Reachable: reachable, Message: msg}
	s.emitAudit(ctx, tenantUUID, "carrier.test", carrier.ID, fmt.Sprintf("reachable=%v", reachable))
	return result, nil
}

func (s *CarrierService) emitAudit(ctx context.Context, tenantUUID, action, targetID, result string) {
	if s == nil || s.emitter == nil {
		return
	}
	s.emitter.EmitAudit(LogisticsObs.AuditEvent{
		Action:     action,
		TenantUUID: strings.TrimSpace(tenantUUID),
		TargetID:   targetID,
		Result:     result,
		EmittedAt:  time.Now().UTC(),
	})
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
