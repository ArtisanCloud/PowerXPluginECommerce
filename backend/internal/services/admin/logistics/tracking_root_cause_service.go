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

type TrackingRootCauseService struct {
	rootRepo    *LogisticsRepo.TrackingRootCauseRepository
	waybillRepo *LogisticsRepo.WaybillRepository
	eventRepo   *LogisticsRepo.TrackingEventRepository
}

type TrackingRootCauseQuery struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	AnomalyType     string `json:"anomaly_type,omitempty"`
	Status          string `json:"status,omitempty"`
	WindowHours     int    `json:"window_hours,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type AnalyzeTrackingRootCauseRequest struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	WindowHours     int    `json:"window_hours,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type HandleTrackingRootCauseRequest struct {
	RootCauseID string `json:"root_cause_id"`
	Action      string `json:"action"`
	Status      string `json:"status,omitempty"`
	OperatorID  string `json:"operator_id,omitempty"`
	ResultNote  string `json:"result_note,omitempty"`
}

type TrackingRootCauseDistribution struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type TrackingRootCauseSummary struct {
	WindowHours         int                             `json:"window_hours"`
	TotalCases          int                             `json:"total_cases"`
	OpenCases           int                             `json:"open_cases"`
	ResolvedCases       int                             `json:"resolved_cases"`
	AnomalyDistribution []TrackingRootCauseDistribution `json:"anomaly_distribution"`
	OwnerDistribution   []TrackingRootCauseDistribution `json:"owner_distribution"`
	SuggestedActions    []string                        `json:"suggested_actions"`
}

func NewTrackingRootCauseService(deps *app.Deps) *TrackingRootCauseService {
	if deps == nil || deps.DB == nil {
		return &TrackingRootCauseService{}
	}
	return &TrackingRootCauseService{
		rootRepo:    LogisticsRepo.NewTrackingRootCauseRepository(deps.DB),
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
		eventRepo:   LogisticsRepo.NewTrackingEventRepository(deps.DB),
	}
}

func (s *TrackingRootCauseService) List(ctx context.Context, tenantUUID string, query TrackingRootCauseQuery) ([]LogisticsModel.TrackingRootCause, error) {
	if s == nil || s.rootRepo == nil {
		return nil, errors.New("tracking root cause service unavailable")
	}
	return s.rootRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.TrackingRootCauseFilter{
		CarrierID:       strings.TrimSpace(query.CarrierID),
		WarehouseID:     strings.TrimSpace(query.WarehouseID),
		DestinationZone: strings.TrimSpace(query.DestinationZone),
		AnomalyType:     strings.TrimSpace(query.AnomalyType),
		Status:          normalizeRootCauseStatus(query.Status),
		Limit:           query.Limit,
	})
}

func (s *TrackingRootCauseService) Analyze(ctx context.Context, tenantUUID string, req AnalyzeTrackingRootCauseRequest) ([]LogisticsModel.TrackingRootCause, error) {
	if s == nil || s.rootRepo == nil || s.waybillRepo == nil || s.eventRepo == nil {
		return nil, errors.New("tracking root cause service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	windowHours := normalizeWindowHours(req.WindowHours)
	limit := req.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	windowStart := time.Now().UTC().Add(-time.Duration(windowHours) * time.Hour)

	waybills, err := s.waybillRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]LogisticsModel.TrackingRootCause, 0, limit)

	for _, wb := range waybills {
		if len(out) >= limit {
			break
		}
		if wb.CreatedAt.UTC().Before(windowStart) {
			continue
		}
		meta := parseWaybillMetadata(wb.Metadata)
		if !matchesControlTowerScope(wb, meta, req.CarrierID, req.WarehouseID, req.DestinationZone) {
			continue
		}

		anomalyType, ownerType, severity, suggested, shouldCreate := detectRootCause(wb, time.Now().UTC())
		if !shouldCreate {
			continue
		}
		events, _ := s.eventRepo.ListByWaybillID(ctx, wb.ID)
		evidence, _ := jsonBytes(buildRootCauseEvidence(wb, events), []byte("{}"))

		row, err := s.rootRepo.GetByWaybillAndAnomaly(ctx, wb.ID, anomalyType)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			now := time.Now().UTC()
			row = &LogisticsModel.TrackingRootCause{
				ID:              utils.NewUUID(),
				WaybillID:       wb.ID,
				WaybillNo:       wb.WaybillNo,
				CarrierID:       wb.CarrierID,
				WarehouseID:     meta.WarehouseID,
				DestinationZone: meta.DestinationZone,
				AnomalyType:     anomalyType,
				OwnerType:       ownerType,
				Severity:        severity,
				SuggestedAction: suggested,
				Status:          "open",
				EvidenceChain:   datatypes.JSON(evidence),
				DetectedAt:      &now,
				Metadata:        datatypes.JSON([]byte("{}")),
			}
		} else {
			row.OwnerType = ownerType
			row.Severity = severity
			row.SuggestedAction = suggested
			row.EvidenceChain = datatypes.JSON(evidence)
			if row.Status == "" {
				row.Status = "open"
			}
		}
		if err := s.rootRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		out = append(out, *row)
	}
	return out, nil
}

func (s *TrackingRootCauseService) Summary(ctx context.Context, tenantUUID string, query TrackingRootCauseQuery) (*TrackingRootCauseSummary, error) {
	if s == nil || s.rootRepo == nil {
		return nil, errors.New("tracking root cause service unavailable")
	}
	rows, err := s.List(ctx, tenantUUID, query)
	if err != nil {
		return nil, err
	}
	summary := &TrackingRootCauseSummary{
		WindowHours:         normalizeWindowHours(query.WindowHours),
		AnomalyDistribution: make([]TrackingRootCauseDistribution, 0),
		OwnerDistribution:   make([]TrackingRootCauseDistribution, 0),
		SuggestedActions:    make([]string, 0),
	}
	anomalyMap := map[string]int{}
	ownerMap := map[string]int{}
	actionMap := map[string]struct{}{}
	for _, row := range rows {
		summary.TotalCases++
		if row.Status == "resolved" || row.Status == "closed" {
			summary.ResolvedCases++
		} else {
			summary.OpenCases++
		}
		anomalyMap[row.AnomalyType]++
		ownerMap[row.OwnerType]++
		if strings.TrimSpace(row.SuggestedAction) != "" {
			actionMap[row.SuggestedAction] = struct{}{}
		}
	}
	for key, count := range anomalyMap {
		summary.AnomalyDistribution = append(summary.AnomalyDistribution, TrackingRootCauseDistribution{Key: key, Count: count})
	}
	for key, count := range ownerMap {
		summary.OwnerDistribution = append(summary.OwnerDistribution, TrackingRootCauseDistribution{Key: key, Count: count})
	}
	for key := range actionMap {
		summary.SuggestedActions = append(summary.SuggestedActions, key)
	}
	return summary, nil
}

func (s *TrackingRootCauseService) Handle(ctx context.Context, tenantUUID string, req HandleTrackingRootCauseRequest) (*LogisticsModel.TrackingRootCause, error) {
	if s == nil || s.rootRepo == nil {
		return nil, errors.New("tracking root cause service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	rootID := strings.TrimSpace(req.RootCauseID)
	if rootID == "" {
		return nil, errors.New("root_cause_id is required")
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		return nil, errors.New("action is required")
	}
	status := normalizeRootCauseHandleStatus(req.Status)
	if status == "" {
		status = "resolved"
	}
	row, err := s.rootRepo.GetByID(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(row.SelectedAction), action) && strings.EqualFold(strings.TrimSpace(row.Status), status) {
		return row, nil
	}
	now := time.Now().UTC()
	row.SelectedAction = action
	row.Status = status
	row.ResultNote = strings.TrimSpace(req.ResultNote)
	row.HandledBy = strings.TrimSpace(req.OperatorID)
	row.HandledAt = &now
	if err := s.rootRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func detectRootCause(row LogisticsModel.Waybill, now time.Time) (anomalyType, ownerType, severity, suggestedAction string, shouldCreate bool) {
	status := strings.ToLower(strings.TrimSpace(row.Status))
	switch status {
	case "delay":
		return "delay", "carrier", "high", "expedite_route", true
	case "exception":
		return "tracking_exception", "warehouse", "high", "manual_review", true
	case "cancelled":
		return "cancelled", "system", "medium", "reroute_recreate", true
	case "in_transit", "created", "pending":
		if row.CreatedAt.Add(72 * time.Hour).Before(now) {
			return "timeout", "carrier", "high", "priority_dispatch", true
		}
	}
	return "", "", "", "", false
}

func buildRootCauseEvidence(row LogisticsModel.Waybill, events []LogisticsModel.TrackingEvent) map[string]any {
	eventStatuses := make([]string, 0, len(events))
	for i, evt := range events {
		if i >= 5 {
			break
		}
		eventStatuses = append(eventStatuses, evt.Status)
	}
	elapsedHours := int(time.Now().UTC().Sub(row.CreatedAt).Hours())
	return map[string]any{
		"waybill_status":     row.Status,
		"elapsed_hours":      elapsedHours,
		"recent_event_count": len(events),
		"recent_statuses":    eventStatuses,
	}
}

func normalizeRootCauseStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "open", "processing", "resolved", "closed", "ignored":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func normalizeRootCauseHandleStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "processing", "resolved", "closed", "ignored":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}
