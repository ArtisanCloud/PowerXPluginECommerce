package fulfillment

import (
	"context"
	"errors"
	"strings"

	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	FulfillmentRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/fulfillment"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WarehouseBridgeService struct {
	outboundRepo *FulfillmentRepo.OutboundRepository
	pickRepo     *FulfillmentRepo.PickItemRepository
	packRepo     *FulfillmentRepo.PackOrderRepository
	taskRepo     *FulfillmentRepo.TaskRepository
	logRepo      *FulfillmentRepo.TaskLogRepository
}

func NewWarehouseBridgeService(deps *app.Deps) *WarehouseBridgeService {
	if deps == nil || deps.DB == nil {
		return &WarehouseBridgeService{}
	}
	return &WarehouseBridgeService{
		outboundRepo: FulfillmentRepo.NewOutboundRepository(deps.DB),
		pickRepo:     FulfillmentRepo.NewPickItemRepository(deps.DB),
		packRepo:     FulfillmentRepo.NewPackOrderRepository(deps.DB),
		taskRepo:     FulfillmentRepo.NewTaskRepository(deps.DB),
		logRepo:      FulfillmentRepo.NewTaskLogRepository(deps.DB),
	}
}

type CreateOutboundRequest struct {
	TaskID    string         `json:"task_id"`
	WaybillID string         `json:"waybill_id,omitempty"`
	Items     []PickLineItem `json:"items,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type PickLineItem struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type ExecuteOutboundRequest struct {
	OperatorID string         `json:"operator_id,omitempty"`
	PackageNo  int            `json:"package_no,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type OutboundExecutionResult struct {
	Outbound *FulfillmentModel.Outbound `json:"outbound"`
	Task     *FulfillmentModel.Task     `json:"task"`
}

func (s *WarehouseBridgeService) ListOutbounds(ctx context.Context, tenantUUID, status string) ([]FulfillmentModel.Outbound, error) {
	if s == nil || s.outboundRepo == nil {
		return nil, errors.New("warehouse bridge service unavailable")
	}
	return s.outboundRepo.List(withTenantContext(ctx, tenantUUID), status)
}

func (s *WarehouseBridgeService) CreateOutbound(ctx context.Context, tenantUUID string, req CreateOutboundRequest) (*FulfillmentModel.Outbound, error) {
	if s == nil || s.outboundRepo == nil || s.taskRepo == nil {
		return nil, errors.New("warehouse bridge service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		return nil, errors.New("task_id required")
	}
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if existing, getErr := s.outboundRepo.GetByTaskID(ctx, taskID); getErr == nil && existing != nil {
		return existing, nil
	}
	metadata, _ := jsonBytes(req.Metadata, []byte("{}"))
	row := &FulfillmentModel.Outbound{
		ID:          utils.NewUUID(),
		TaskID:      task.ID,
		OrderID:     task.OrderID,
		WarehouseID: task.WarehouseID,
		WaybillID:   strings.TrimSpace(req.WaybillID),
		Status:      "reserved",
		Metadata:    datatypes.JSON(metadata),
	}
	if err := s.outboundRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	picks := make([]FulfillmentModel.PickItem, 0, len(req.Items))
	for _, item := range req.Items {
		sku := strings.TrimSpace(item.SKU)
		if sku == "" {
			continue
		}
		qty := item.Qty
		if qty <= 0 {
			qty = 1
		}
		picks = append(picks, FulfillmentModel.PickItem{
			ID:         utils.NewUUID(),
			OutboundID: row.ID,
			SKU:        sku,
			Qty:        qty,
			Status:     "pending",
			Metadata:   datatypes.JSON([]byte("{}")),
		})
	}
	if len(picks) > 0 && s.pickRepo != nil {
		if err := s.pickRepo.CreateBatch(ctx, picks); err != nil {
			return nil, err
		}
	}
	task.Status = "picking"
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, err
	}
	_ = s.appendTaskLog(ctx, task.ID, "task.warehouse.reserve", "system", map[string]any{
		"outbound_id": row.ID,
		"status":      row.Status,
	})
	return row, nil
}

func (s *WarehouseBridgeService) ExecuteOutbound(ctx context.Context, tenantUUID, outboundID string, req ExecuteOutboundRequest) (*OutboundExecutionResult, error) {
	if s == nil || s.outboundRepo == nil || s.taskRepo == nil {
		return nil, errors.New("warehouse bridge service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	outbound, err := s.outboundRepo.GetByID(ctx, outboundID)
	if err != nil {
		return nil, err
	}
	task, err := s.taskRepo.GetByID(ctx, outbound.TaskID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(outbound.Status, "shipped") {
		return &OutboundExecutionResult{Outbound: outbound, Task: task}, nil
	}
	packageNo := req.PackageNo
	if packageNo <= 0 {
		packageNo = 1
	}
	if s.packRepo != nil {
		if _, getErr := s.packRepo.GetByOutboundID(ctx, outbound.ID); getErr != nil {
			if !errors.Is(getErr, gorm.ErrRecordNotFound) {
				return nil, getErr
			}
			meta, _ := jsonBytes(req.Metadata, []byte("{}"))
			if err := s.packRepo.Create(ctx, &FulfillmentModel.PackOrder{
				ID:         utils.NewUUID(),
				OutboundID: outbound.ID,
				PackageNo:  packageNo,
				Status:     "packed",
				Metadata:   datatypes.JSON(meta),
			}); err != nil {
				return nil, err
			}
		}
	}
	outbound.Status = "shipped"
	if err := s.outboundRepo.Save(ctx, outbound); err != nil {
		return nil, err
	}
	task.Status = "completed"
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, err
	}
	_ = s.appendTaskLog(ctx, task.ID, "task.warehouse.ship", req.OperatorID, map[string]any{
		"outbound_id": outbound.ID,
		"status":      outbound.Status,
	})
	return &OutboundExecutionResult{Outbound: outbound, Task: task}, nil
}

func (s *WarehouseBridgeService) RollbackOutbound(ctx context.Context, tenantUUID, outboundID, reason string) (*OutboundExecutionResult, error) {
	if s == nil || s.outboundRepo == nil || s.taskRepo == nil {
		return nil, errors.New("warehouse bridge service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	outbound, err := s.outboundRepo.GetByID(ctx, outboundID)
	if err != nil {
		return nil, err
	}
	task, err := s.taskRepo.GetByID(ctx, outbound.TaskID)
	if err != nil {
		return nil, err
	}
	outbound.Status = "rollback"
	if err := s.outboundRepo.Save(ctx, outbound); err != nil {
		return nil, err
	}
	task.Status = "exception"
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, err
	}
	_ = s.appendTaskLog(ctx, task.ID, "task.warehouse.rollback", "system", map[string]any{
		"outbound_id": outbound.ID,
		"reason":      strings.TrimSpace(reason),
	})
	return &OutboundExecutionResult{Outbound: outbound, Task: task}, nil
}

func (s *WarehouseBridgeService) appendTaskLog(ctx context.Context, taskID, action, operatorID string, detail map[string]any) error {
	if s == nil || s.logRepo == nil {
		return nil
	}
	payload, _ := jsonBytes(detail, []byte("{}"))
	return s.logRepo.Create(ctx, &FulfillmentModel.TaskLog{
		ID:         utils.NewUUID(),
		TaskID:     strings.TrimSpace(taskID),
		Action:     strings.TrimSpace(action),
		OperatorID: strings.TrimSpace(operatorID),
		Detail:     datatypes.JSON(payload),
	})
}
