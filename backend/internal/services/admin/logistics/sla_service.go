package logistics

import (
	"context"
	"errors"
	"strings"
	"time"

	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

type SLAService struct {
	slaRepo   *LogisticsRepo.SLARepository
	trackRepo *LogisticsRepo.TrackingEventRepository
}

func NewSLAService(deps *app.Deps) *SLAService {
	if deps == nil || deps.DB == nil {
		return &SLAService{}
	}
	waybillRepo := LogisticsRepo.NewWaybillRepository(deps.DB)
	return &SLAService{
		slaRepo:   LogisticsRepo.NewSLARepository(waybillRepo),
		trackRepo: LogisticsRepo.NewTrackingEventRepository(deps.DB),
	}
}

type SLAQuery struct {
	CarrierID        string `json:"carrier_id,omitempty"`
	From             string `json:"from,omitempty"`
	To               string `json:"to,omitempty"`
	PickupSLAHours   int    `json:"pickup_sla_hours,omitempty"`
	DeliverySLAHours int    `json:"delivery_sla_hours,omitempty"`
}

type SLACarrierSummary struct {
	CarrierID         string  `json:"carrier_id"`
	CarrierName       string  `json:"carrier_name"`
	WaybillCount      int64   `json:"waybill_count"`
	PickupOnTimeCount int64   `json:"pickup_on_time_count"`
	PickupOnTimeRate  float64 `json:"pickup_on_time_rate"`
	SignOnTimeCount   int64   `json:"sign_on_time_count"`
	SignOnTimeRate    float64 `json:"sign_on_time_rate"`
	ExceptionCount    int64   `json:"exception_count"`
	ExceptionRate     float64 `json:"exception_rate"`
	PickupSLAHours    int     `json:"pickup_sla_hours"`
	DeliverySLAHours  int     `json:"delivery_sla_hours"`
}

type SLASnapshot struct {
	Summary []SLACarrierSummary `json:"summary"`
	Total   SLACarrierSummary   `json:"total"`
}

func (s *SLAService) Snapshot(ctx context.Context, tenantUUID string, query SLAQuery) (*SLASnapshot, error) {
	if s == nil || s.slaRepo == nil {
		return nil, errors.New("sla service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	filter, pickupSLAHours, deliverySLAHours, err := parseSLAFilter(query)
	if err != nil {
		return nil, err
	}
	rows, err := s.slaRepo.ListWaybillMetrics(ctx, filter)
	if err != nil {
		return nil, err
	}
	acc := map[string]*SLACarrierSummary{}
	total := SLACarrierSummary{
		CarrierID:        "all",
		CarrierName:      "全部承运商",
		PickupSLAHours:   pickupSLAHours,
		DeliverySLAHours: deliverySLAHours,
	}
	pickupSLA := time.Duration(pickupSLAHours) * time.Hour
	deliverySLA := time.Duration(deliverySLAHours) * time.Hour
	for _, row := range rows {
		key := strings.TrimSpace(row.CarrierID)
		if key == "" {
			key = "unknown"
		}
		item := acc[key]
		if item == nil {
			item = &SLACarrierSummary{
				CarrierID:        key,
				CarrierName:      row.CarrierName,
				PickupSLAHours:   pickupSLAHours,
				DeliverySLAHours: deliverySLAHours,
			}
			acc[key] = item
		}
		item.WaybillCount++
		total.WaybillCount++

		pickupAt, deliveredAt := s.resolveSLAEventTimes(ctx, row.WaybillID)
		if pickupAt != nil && !pickupAt.Before(row.CreatedAt) && pickupAt.Sub(row.CreatedAt) <= pickupSLA {
			item.PickupOnTimeCount++
			total.PickupOnTimeCount++
		}
		if deliveredAt != nil && !deliveredAt.Before(row.CreatedAt) && deliveredAt.Sub(row.CreatedAt) <= deliverySLA {
			item.SignOnTimeCount++
			total.SignOnTimeCount++
		}
		if row.HasException {
			item.ExceptionCount++
			total.ExceptionCount++
		}
	}
	summary := make([]SLACarrierSummary, 0, len(acc))
	for _, item := range acc {
		item.PickupOnTimeRate = percent(item.PickupOnTimeCount, item.WaybillCount)
		item.SignOnTimeRate = percent(item.SignOnTimeCount, item.WaybillCount)
		item.ExceptionRate = percent(item.ExceptionCount, item.WaybillCount)
		summary = append(summary, *item)
	}
	total.PickupOnTimeRate = percent(total.PickupOnTimeCount, total.WaybillCount)
	total.SignOnTimeRate = percent(total.SignOnTimeCount, total.WaybillCount)
	total.ExceptionRate = percent(total.ExceptionCount, total.WaybillCount)
	return &SLASnapshot{
		Summary: summary,
		Total:   total,
	}, nil
}

func parseSLAFilter(query SLAQuery) (LogisticsRepo.SLAFilter, int, int, error) {
	filter := LogisticsRepo.SLAFilter{CarrierID: strings.TrimSpace(query.CarrierID)}
	pickupSLAHours := query.PickupSLAHours
	if pickupSLAHours <= 0 {
		pickupSLAHours = 24
	}
	deliverySLAHours := query.DeliverySLAHours
	if deliverySLAHours <= 0 {
		deliverySLAHours = 72
	}
	if strings.TrimSpace(query.From) != "" {
		from, err := time.Parse(time.RFC3339, strings.TrimSpace(query.From))
		if err != nil {
			return filter, 0, 0, errors.New("from must be RFC3339")
		}
		filter.From = &from
	}
	if strings.TrimSpace(query.To) != "" {
		to, err := time.Parse(time.RFC3339, strings.TrimSpace(query.To))
		if err != nil {
			return filter, 0, 0, errors.New("to must be RFC3339")
		}
		filter.To = &to
	}
	return filter, pickupSLAHours, deliverySLAHours, nil
}

func percent(part, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(part) * 100 / float64(total)
}

func (s *SLAService) resolveSLAEventTimes(ctx context.Context, waybillID string) (*time.Time, *time.Time) {
	if s == nil || s.trackRepo == nil || strings.TrimSpace(waybillID) == "" {
		return nil, nil
	}
	events, err := s.trackRepo.ListByWaybillID(ctx, waybillID)
	if err != nil {
		return nil, nil
	}
	var pickupAt *time.Time
	var deliveredAt *time.Time
	for _, event := range events {
		if event.OccurredAt == nil {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(event.Status))
		if status == "picked" || status == "in_transit" || status == "in-transit" {
			if pickupAt == nil || event.OccurredAt.Before(*pickupAt) {
				t := *event.OccurredAt
				pickupAt = &t
			}
		}
		if status == "delivered" || status == "signed" {
			if deliveredAt == nil || event.OccurredAt.Before(*deliveredAt) {
				t := *event.OccurredAt
				deliveredAt = &t
			}
		}
	}
	return pickupAt, deliveredAt
}
