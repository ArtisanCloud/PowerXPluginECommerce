package logistics

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

type QualityAuditReportService struct {
	reportRepo         *LogisticsRepo.QualityAuditReportRepository
	waybillRepo        *LogisticsRepo.WaybillRepository
	forecastRepo       *LogisticsRepo.CapacityForecastRepository
	rootCauseRepo      *LogisticsRepo.TrackingRootCauseRepository
	interwarehouseRepo *LogisticsRepo.InterwarehouseAllocationRepository
}

type QualityAuditReportQuery struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	Status          string `json:"status,omitempty"`
	WindowHours     int    `json:"window_hours,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type GenerateQualityAuditReportRequest struct {
	CarrierID        string `json:"carrier_id,omitempty"`
	WarehouseID      string `json:"warehouse_id,omitempty"`
	DestinationZone  string `json:"destination_zone,omitempty"`
	ReportPeriodFrom string `json:"report_period_from,omitempty"`
	ReportPeriodTo   string `json:"report_period_to,omitempty"`
	WindowHours      int    `json:"window_hours,omitempty"`
	OperatorID       string `json:"operator_id,omitempty"`
}

func NewQualityAuditReportService(deps *app.Deps) *QualityAuditReportService {
	if deps == nil || deps.DB == nil {
		return &QualityAuditReportService{}
	}
	return &QualityAuditReportService{
		reportRepo:         LogisticsRepo.NewQualityAuditReportRepository(deps.DB),
		waybillRepo:        LogisticsRepo.NewWaybillRepository(deps.DB),
		forecastRepo:       LogisticsRepo.NewCapacityForecastRepository(deps.DB),
		rootCauseRepo:      LogisticsRepo.NewTrackingRootCauseRepository(deps.DB),
		interwarehouseRepo: LogisticsRepo.NewInterwarehouseAllocationRepository(deps.DB),
	}
}

func (s *QualityAuditReportService) List(ctx context.Context, tenantUUID string, query QualityAuditReportQuery) ([]LogisticsModel.QualityAuditReport, error) {
	if s == nil || s.reportRepo == nil {
		return nil, errors.New("quality audit report service unavailable")
	}
	return s.reportRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.QualityAuditReportFilter{
		CarrierID:       strings.TrimSpace(query.CarrierID),
		WarehouseID:     strings.TrimSpace(query.WarehouseID),
		DestinationZone: strings.TrimSpace(query.DestinationZone),
		Status:          normalizeQualityAuditStatus(query.Status),
		Limit:           query.Limit,
	})
}

func (s *QualityAuditReportService) Get(ctx context.Context, tenantUUID, reportID string) (*LogisticsModel.QualityAuditReport, error) {
	if s == nil || s.reportRepo == nil {
		return nil, errors.New("quality audit report service unavailable")
	}
	id := strings.TrimSpace(reportID)
	if id == "" {
		return nil, errors.New("report id is required")
	}
	return s.reportRepo.GetByID(withTenantContext(ctx, tenantUUID), id)
}

func (s *QualityAuditReportService) Generate(ctx context.Context, tenantUUID string, req GenerateQualityAuditReportRequest) (*LogisticsModel.QualityAuditReport, error) {
	if s == nil || s.reportRepo == nil || s.waybillRepo == nil || s.forecastRepo == nil || s.rootCauseRepo == nil || s.interwarehouseRepo == nil {
		return nil, errors.New("quality audit report service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	windowHours := normalizeWindowHours(req.WindowHours)
	periodFrom, periodTo, err := resolveReportPeriod(req.ReportPeriodFrom, req.ReportPeriodTo, windowHours)
	if err != nil {
		return nil, err
	}
	if !periodTo.After(periodFrom) {
		return nil, errors.New("report period is invalid")
	}

	waybills, err := s.waybillRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	filteredWaybills := make([]LogisticsModel.Waybill, 0, len(waybills))
	for _, row := range waybills {
		if row.CreatedAt.UTC().Before(periodFrom) || row.CreatedAt.UTC().After(periodTo) {
			continue
		}
		meta := parseWaybillMetadata(row.Metadata)
		if !matchesControlTowerScope(row, meta, req.CarrierID, req.WarehouseID, req.DestinationZone) {
			continue
		}
		filteredWaybills = append(filteredWaybills, row)
	}
	kpi := summarizeKPI(filteredWaybills, windowHours, "quality_audit")

	forecastRows, err := s.forecastRepo.List(ctx, LogisticsRepo.CapacityForecastFilter{
		CarrierID:       strings.TrimSpace(req.CarrierID),
		WarehouseID:     strings.TrimSpace(req.WarehouseID),
		DestinationZone: strings.TrimSpace(req.DestinationZone),
		Limit:           500,
	})
	if err != nil {
		return nil, err
	}
	forecastCount := 0
	for _, row := range forecastRows {
		if row.CreatedAt.UTC().Before(periodFrom) || row.CreatedAt.UTC().After(periodTo) {
			continue
		}
		forecastCount++
	}

	rootCauseRows, err := s.rootCauseRepo.List(ctx, LogisticsRepo.TrackingRootCauseFilter{
		CarrierID:       strings.TrimSpace(req.CarrierID),
		WarehouseID:     strings.TrimSpace(req.WarehouseID),
		DestinationZone: strings.TrimSpace(req.DestinationZone),
		Limit:           500,
	})
	if err != nil {
		return nil, err
	}
	rootCauseCount := 0
	for _, row := range rootCauseRows {
		if row.CreatedAt.UTC().Before(periodFrom) || row.CreatedAt.UTC().After(periodTo) {
			continue
		}
		rootCauseCount++
	}

	interRows, err := s.interwarehouseRepo.List(ctx, LogisticsRepo.InterwarehouseAllocationFilter{
		CarrierID: strings.TrimSpace(req.CarrierID),
		Limit:     500,
	})
	if err != nil {
		return nil, err
	}
	interwarehouseCount := 0
	for _, row := range interRows {
		if row.CreatedAt.UTC().Before(periodFrom) || row.CreatedAt.UTC().After(periodTo) {
			continue
		}
		if strings.TrimSpace(req.DestinationZone) != "" && row.DestinationZone != strings.TrimSpace(req.DestinationZone) {
			continue
		}
		interwarehouseCount++
	}

	conclusion, actions := summarizeQualityAuditConclusionAndActions(
		kpi,
		forecastCount,
		rootCauseCount,
		interwarehouseCount,
	)
	actionJSON, _ := jsonBytes(actions, []byte("[]"))
	metaJSON, _ := jsonBytes(map[string]any{
		"report_period_from": periodFrom.Format(time.RFC3339),
		"report_period_to":   periodTo.Format(time.RFC3339),
	}, []byte("{}"))
	now := time.Now().UTC()

	row := &LogisticsModel.QualityAuditReport{
		ID:                  utils.NewUUID(),
		ReportPeriodFrom:    periodFrom,
		ReportPeriodTo:      periodTo,
		WindowHours:         windowHours,
		CarrierID:           strings.TrimSpace(req.CarrierID),
		WarehouseID:         strings.TrimSpace(req.WarehouseID),
		DestinationZone:     strings.TrimSpace(req.DestinationZone),
		TotalWaybills:       kpi.TotalWaybills,
		DeliveredCount:      kpi.DeliveredCount,
		ExceptionCount:      kpi.ExceptionCount,
		TimeoutCount:        kpi.TimeoutCount,
		OnTimeRate:          kpi.OnTimeRate,
		DeliverySuccessRate: kpi.DeliverySuccessRate,
		TotalCost:           kpi.TotalCost,
		AvgCost:             kpi.AvgCost,
		ForecastCount:       forecastCount,
		RootCauseCount:      rootCauseCount,
		InterwarehouseCount: interwarehouseCount,
		Conclusion:          conclusion,
		ActionItems:         datatypes.JSON(actionJSON),
		Status:              "generated",
		GeneratedBy:         strings.TrimSpace(req.OperatorID),
		GeneratedAt:         &now,
		Metadata:            datatypes.JSON(metaJSON),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.reportRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *QualityAuditReportService) ExportCSV(ctx context.Context, tenantUUID string, query QualityAuditReportQuery) (string, error) {
	rows, err := s.List(ctx, tenantUUID, query)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	writer := csv.NewWriter(&b)
	_ = writer.Write([]string{
		"id", "period_from", "period_to", "carrier_id", "warehouse_id", "destination_zone",
		"total_waybills", "delivered_count", "exception_count", "timeout_count",
		"on_time_rate", "delivery_success_rate", "total_cost", "avg_cost",
		"forecast_count", "root_cause_count", "interwarehouse_count", "status", "conclusion",
	})
	for _, row := range rows {
		_ = writer.Write([]string{
			row.ID,
			row.ReportPeriodFrom.UTC().Format(time.RFC3339),
			row.ReportPeriodTo.UTC().Format(time.RFC3339),
			row.CarrierID,
			row.WarehouseID,
			row.DestinationZone,
			strconv.Itoa(row.TotalWaybills),
			strconv.Itoa(row.DeliveredCount),
			strconv.Itoa(row.ExceptionCount),
			strconv.Itoa(row.TimeoutCount),
			fmt.Sprintf("%.2f", row.OnTimeRate),
			fmt.Sprintf("%.2f", row.DeliverySuccessRate),
			fmt.Sprintf("%.2f", row.TotalCost),
			fmt.Sprintf("%.2f", row.AvgCost),
			strconv.Itoa(row.ForecastCount),
			strconv.Itoa(row.RootCauseCount),
			strconv.Itoa(row.InterwarehouseCount),
			row.Status,
			row.Conclusion,
		})
	}
	writer.Flush()
	return b.String(), writer.Error()
}

func resolveReportPeriod(fromRaw, toRaw string, windowHours int) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	to := now
	if strings.TrimSpace(toRaw) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(toRaw))
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("report_period_to must be RFC3339")
		}
		to = parsed.UTC()
	}
	from := to.Add(-time.Duration(normalizeWindowHours(windowHours)) * time.Hour)
	if strings.TrimSpace(fromRaw) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(fromRaw))
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("report_period_from must be RFC3339")
		}
		from = parsed.UTC()
	}
	return from, to, nil
}

func summarizeQualityAuditConclusionAndActions(
	kpi *KPIDashboardOverview,
	forecastCount int,
	rootCauseCount int,
	interwarehouseCount int,
) (string, []map[string]any) {
	if kpi == nil {
		kpi = &KPIDashboardOverview{}
	}
	actions := make([]map[string]any, 0)
	exceptionRate := 0.0
	if kpi.TotalWaybills > 0 {
		exceptionRate = float64(kpi.ExceptionCount) * 100 / float64(kpi.TotalWaybills)
	}
	if kpi.OnTimeRate < 90 {
		actions = append(actions, map[string]any{
			"area":     "timeliness",
			"priority": "high",
			"action":   "针对延误路由启用加急策略并复核 SLA 守卫阈值",
		})
	}
	if exceptionRate >= 8 {
		actions = append(actions, map[string]any{
			"area":     "risk",
			"priority": "high",
			"action":   "按根因类型拆分异常专案，优先处理高频异常责任方",
		})
	}
	if interwarehouseCount == 0 && kpi.TimeoutCount > 0 {
		actions = append(actions, map[string]any{
			"area":     "collaboration",
			"priority": "medium",
			"action":   "增加跨仓协同调拨策略覆盖，降低超时单",
		})
	}
	if forecastCount == 0 {
		actions = append(actions, map[string]any{
			"area":     "capacity",
			"priority": "medium",
			"action":   "补齐容量预测生成频次，建立每周配额复盘",
		})
	}
	if len(actions) == 0 {
		actions = append(actions, map[string]any{
			"area":     "stability",
			"priority": "low",
			"action":   "保持当前策略并持续监控关键指标波动",
		})
	}

	conclusion := "履约质量整体稳定"
	switch {
	case kpi.TotalWaybills == 0:
		conclusion = "周期内无有效履约样本"
	case kpi.OnTimeRate < 85 || exceptionRate >= 12:
		conclusion = "履约质量存在显著风险，需立即执行专项整改"
	case kpi.OnTimeRate < 92 || exceptionRate >= 6 || rootCauseCount >= 5:
		conclusion = "履约质量存在波动，建议按行动项持续优化"
	}
	return conclusion, actions
}

func normalizeQualityAuditStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "generated", "published", "archived":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}
