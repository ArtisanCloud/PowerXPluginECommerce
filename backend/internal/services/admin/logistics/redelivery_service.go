package logistics

import (
	"context"
	"errors"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RedeliveryService struct {
	taskRepo    *LogisticsRepo.RedeliveryTaskRepository
	waybillRepo *LogisticsRepo.WaybillRepository
}

func NewRedeliveryService(deps *app.Deps) *RedeliveryService {
	if deps == nil || deps.DB == nil {
		return &RedeliveryService{}
	}
	return &RedeliveryService{
		taskRepo:    LogisticsRepo.NewRedeliveryTaskRepository(deps.DB),
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
	}
}

type InitiateRedeliveryRequest struct {
	WaybillID  string         `json:"waybill_id"`
	RequestKey string         `json:"request_key,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	OperatorID string         `json:"operator_id,omitempty"`
	Address    map[string]any `json:"address,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type UpdateRedeliveryAddressRequest struct {
	Address    map[string]any `json:"address"`
	OperatorID string         `json:"operator_id,omitempty"`
	Reason     string         `json:"reason,omitempty"`
}

type RedispatchRedeliveryRequest struct {
	RequestKey string         `json:"request_key,omitempty"`
	OperatorID string         `json:"operator_id,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type CloseRedeliveryRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func (s *RedeliveryService) List(ctx context.Context, tenantUUID, waybillID, status string) ([]LogisticsModel.RedeliveryTask, error) {
	if s == nil || s.taskRepo == nil {
		return nil, errors.New("redelivery service unavailable")
	}
	return s.taskRepo.List(withTenantContext(ctx, tenantUUID), waybillID, status)
}

func (s *RedeliveryService) Initiate(ctx context.Context, tenantUUID string, req InitiateRedeliveryRequest) (*LogisticsModel.RedeliveryTask, string, error) {
	if s == nil || s.taskRepo == nil || s.waybillRepo == nil {
		return nil, "", errors.New("redelivery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	waybillID := strings.TrimSpace(req.WaybillID)
	if waybillID == "" {
		return nil, "", errors.New("waybill_id required")
	}
	wb, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, "", err
	}
	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = "init:" + wb.ID + ":" + utils.NewUUID()
	}
	if existed, err := s.taskRepo.GetByRequestKey(ctx, requestKey); err == nil && existed != nil {
		return existed, "replayed", nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	meta, _ := jsonBytes(req.Metadata, []byte("{}"))
	addr, _ := jsonBytes(req.Address, []byte("{}"))
	row := &LogisticsModel.RedeliveryTask{
		ID:              utils.NewUUID(),
		WaybillID:       wb.ID,
		WaybillNo:       wb.WaybillNo,
		RequestKey:      requestKey,
		Status:          "initiated",
		AttemptNo:       1,
		AddressSnapshot: datatypes.JSON(addr),
		LastReason:      strings.TrimSpace(req.Reason),
		OperatorID:      strings.TrimSpace(req.OperatorID),
		Metadata:        datatypes.JSON(meta),
	}
	if err := s.taskRepo.Create(ctx, row); err != nil {
		return nil, "", err
	}
	return row, "created", nil
}

func (s *RedeliveryService) UpdateAddress(ctx context.Context, tenantUUID, taskID string, req UpdateRedeliveryAddressRequest) (*LogisticsModel.RedeliveryTask, error) {
	if s == nil || s.taskRepo == nil {
		return nil, errors.New("redelivery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(task.Status, "closed") {
		return nil, errors.New("closed task cannot update address")
	}
	addr, _ := jsonBytes(req.Address, []byte("{}"))
	task.AddressSnapshot = datatypes.JSON(addr)
	task.Status = "address_updated"
	task.OperatorID = strings.TrimSpace(req.OperatorID)
	task.LastReason = strings.TrimSpace(req.Reason)
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *RedeliveryService) Redispatch(ctx context.Context, tenantUUID, taskID string, req RedispatchRedeliveryRequest) (*LogisticsModel.RedeliveryTask, string, error) {
	if s == nil || s.taskRepo == nil {
		return nil, "", errors.New("redelivery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, "", err
	}
	if strings.EqualFold(task.Status, "closed") {
		return nil, "", errors.New("closed task cannot redispatch")
	}
	reqKey := strings.TrimSpace(req.RequestKey)
	if reqKey == "" {
		reqKey = "redispatch:" + task.ID + ":" + utils.NewUUID()
	}
	if reqKey == strings.TrimSpace(task.RequestKey) {
		return task, "replayed", nil
	}
	task.AttemptNo++
	task.RequestKey = reqKey
	task.Status = "redispatched"
	task.OperatorID = strings.TrimSpace(req.OperatorID)
	task.LastReason = strings.TrimSpace(req.Reason)
	if len(req.Metadata) > 0 {
		meta, _ := jsonBytes(req.Metadata, []byte("{}"))
		task.Metadata = datatypes.JSON(meta)
	}
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, "", err
	}
	return task, "created", nil
}

func (s *RedeliveryService) Close(ctx context.Context, tenantUUID, taskID string, req CloseRedeliveryRequest) (*LogisticsModel.RedeliveryTask, error) {
	if s == nil || s.taskRepo == nil {
		return nil, errors.New("redelivery service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(task.Status, "closed") {
		return task, nil
	}
	now := time.Now().UTC()
	task.Status = "closed"
	task.OperatorID = strings.TrimSpace(req.OperatorID)
	task.LastReason = strings.TrimSpace(req.Reason)
	task.ClosedAt = &now
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}
