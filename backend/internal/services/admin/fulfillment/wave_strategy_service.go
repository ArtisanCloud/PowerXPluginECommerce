package fulfillment

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	FulfillmentRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/fulfillment"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type WaveStrategyService struct {
	strategyRepo *FulfillmentRepo.WaveStrategyRepository
	taskRepo     *FulfillmentRepo.TaskRepository
}

func NewWaveStrategyService(deps *app.Deps) *WaveStrategyService {
	if deps == nil || deps.DB == nil {
		return &WaveStrategyService{}
	}
	return &WaveStrategyService{
		strategyRepo: FulfillmentRepo.NewWaveStrategyRepository(deps.DB),
		taskRepo:     FulfillmentRepo.NewTaskRepository(deps.DB),
	}
}

type CreateWaveStrategyRequest struct {
	Name            string         `json:"name"`
	WarehouseID     string         `json:"warehouse_id,omitempty"`
	CarrierCode     string         `json:"carrier_code,omitempty"`
	TimeWindow      string         `json:"time_window,omitempty"`
	PriorityBand    string         `json:"priority_band,omitempty"`
	MaxTasksPerWave int            `json:"max_tasks_per_wave,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
	Rules           map[string]any `json:"rules,omitempty"`
}

type PreviewWaveStrategyRequest struct {
	StrategyID string `json:"strategy_id"`
}

type WaveStrategyPreviewGroup struct {
	GroupKey     string   `json:"group_key"`
	WarehouseID  string   `json:"warehouse_id"`
	CarrierCode  string   `json:"carrier_code,omitempty"`
	TimeSlot     string   `json:"time_slot,omitempty"`
	PriorityBand string   `json:"priority_band,omitempty"`
	TaskIDs      []string `json:"task_ids"`
}

type WaveStrategyPreviewResult struct {
	StrategyID     string                     `json:"strategy_id"`
	StrategyName   string                     `json:"strategy_name"`
	TotalCandidate int                        `json:"total_candidate"`
	Groups         []WaveStrategyPreviewGroup `json:"groups"`
	SkippedTaskIDs []string                   `json:"skipped_task_ids"`
}

func (s *WaveStrategyService) List(ctx context.Context, tenantUUID string) ([]FulfillmentModel.WaveStrategy, error) {
	if s == nil || s.strategyRepo == nil {
		return nil, errors.New("wave strategy service unavailable")
	}
	return s.strategyRepo.List(withTenantContext(ctx, tenantUUID), false)
}

func (s *WaveStrategyService) Create(ctx context.Context, tenantUUID string, req CreateWaveStrategyRequest) (*FulfillmentModel.WaveStrategy, error) {
	if s == nil || s.strategyRepo == nil {
		return nil, errors.New("wave strategy service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name required")
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	maxTasks := req.MaxTasksPerWave
	if maxTasks <= 0 {
		maxTasks = 50
	}
	rules, _ := jsonBytes(req.Rules, []byte("{}"))
	row := &FulfillmentModel.WaveStrategy{
		ID:              utils.NewUUID(),
		Name:            name,
		WarehouseID:     strings.TrimSpace(req.WarehouseID),
		CarrierCode:     strings.TrimSpace(req.CarrierCode),
		TimeWindow:      normalizeTimeWindow(req.TimeWindow),
		PriorityBand:    strings.TrimSpace(req.PriorityBand),
		MaxTasksPerWave: maxTasks,
		Enabled:         enabled,
		Rules:           datatypes.JSON(rules),
	}
	if err := s.strategyRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *WaveStrategyService) Preview(ctx context.Context, tenantUUID string, req PreviewWaveStrategyRequest) (*WaveStrategyPreviewResult, error) {
	if s == nil || s.strategyRepo == nil || s.taskRepo == nil {
		return nil, errors.New("wave strategy service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	strategy, err := s.strategyRepo.GetByID(ctx, req.StrategyID)
	if err != nil {
		return nil, err
	}
	if !strategy.Enabled {
		return nil, errors.New("strategy disabled")
	}

	tasks, err := s.taskRepo.List(ctx, "pending")
	if err != nil {
		return nil, err
	}
	groups := make(map[string]*WaveStrategyPreviewGroup)
	skipped := make([]string, 0)

	for _, task := range tasks {
		if !matchWarehouse(strategy.WarehouseID, task.WarehouseID) {
			continue
		}
		meta := decodeTaskMetadata(task.Metadata)
		carrier := readString(meta, "carrier_code")
		priority := readString(meta, "priority")
		slot := resolveTimeSlot(meta, time.Now().UTC())
		if !matchCarrier(strategy.CarrierCode, carrier) {
			continue
		}
		if !matchPriorityBand(strategy.PriorityBand, priority) {
			continue
		}
		if !matchTimeWindow(strategy.TimeWindow, slot) {
			continue
		}
		if carrier == "" {
			skipped = append(skipped, task.ID)
			continue
		}
		key := strings.Join([]string{
			strings.TrimSpace(task.WarehouseID),
			strings.TrimSpace(carrier),
			strings.TrimSpace(slot),
			strings.TrimSpace(priority),
		}, "|")
		group, ok := groups[key]
		if !ok {
			group = &WaveStrategyPreviewGroup{
				GroupKey:     key,
				WarehouseID:  strings.TrimSpace(task.WarehouseID),
				CarrierCode:  carrier,
				TimeSlot:     slot,
				PriorityBand: priority,
				TaskIDs:      make([]string, 0),
			}
			groups[key] = group
		}
		group.TaskIDs = append(group.TaskIDs, task.ID)
	}

	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	resultGroups := make([]WaveStrategyPreviewGroup, 0, len(keys))
	total := 0
	for _, key := range keys {
		group := groups[key]
		sort.Strings(group.TaskIDs)
		if strategy.MaxTasksPerWave > 0 && len(group.TaskIDs) > strategy.MaxTasksPerWave {
			group.TaskIDs = group.TaskIDs[:strategy.MaxTasksPerWave]
		}
		total += len(group.TaskIDs)
		resultGroups = append(resultGroups, *group)
	}
	sort.Strings(skipped)

	return &WaveStrategyPreviewResult{
		StrategyID:     strategy.ID,
		StrategyName:   strategy.Name,
		TotalCandidate: total,
		Groups:         resultGroups,
		SkippedTaskIDs: skipped,
	}, nil
}

func decodeTaskMetadata(raw datatypes.JSON) map[string]any {
	out := map[string]any{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func readString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	value, ok := m[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(strings.ReplaceAll(toString(v), "\n", ""))
	}
}

func resolveTimeSlot(meta map[string]any, now time.Time) string {
	if slot := readString(meta, "time_slot"); slot != "" {
		return slot
	}
	hour := now.Hour()
	switch {
	case hour < 12:
		return "morning"
	case hour < 18:
		return "afternoon"
	default:
		return "evening"
	}
}

func normalizeTimeWindow(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	switch value {
	case "morning", "afternoon", "evening", "all":
		return value
	default:
		return "all"
	}
}

func matchWarehouse(expected, actual string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return true
	}
	return strings.EqualFold(expected, strings.TrimSpace(actual))
}

func matchCarrier(expected, actual string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return true
	}
	return strings.EqualFold(expected, strings.TrimSpace(actual))
}

func matchPriorityBand(expected, actual string) bool {
	expected = strings.TrimSpace(strings.ToLower(expected))
	if expected == "" {
		return true
	}
	return expected == strings.TrimSpace(strings.ToLower(actual))
}

func matchTimeWindow(expected, slot string) bool {
	expected = normalizeTimeWindow(expected)
	if expected == "" || expected == "all" {
		return true
	}
	return expected == strings.TrimSpace(strings.ToLower(slot))
}

func toString(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case int:
		return strconv.Itoa(value)
	case int64:
		return strconv.FormatInt(value, 10)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		if value {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}
