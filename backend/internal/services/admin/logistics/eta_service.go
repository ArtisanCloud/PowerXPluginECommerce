package logistics

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	defaultETATimezone = "UTC"
	defaultCutoffHour  = 18
)

type ETAService struct {
	waybillRepo *LogisticsRepo.WaybillRepository
	trackRepo   *LogisticsRepo.TrackingEventRepository
	policyRepo  *LogisticsRepo.ETAPolicyRepository
	recordRepo  *LogisticsRepo.ETARecordRepository
}

func NewETAService(deps *app.Deps) *ETAService {
	if deps == nil || deps.DB == nil {
		return &ETAService{}
	}
	return &ETAService{
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
		trackRepo:   LogisticsRepo.NewTrackingEventRepository(deps.DB),
		policyRepo:  LogisticsRepo.NewETAPolicyRepository(deps.DB),
		recordRepo:  LogisticsRepo.NewETARecordRepository(deps.DB),
	}
}

type ETAQuery struct {
	DestinationZone string `json:"destination_zone,omitempty"`
	Timezone        string `json:"timezone,omitempty"`
	ForceRecompute  bool   `json:"force_recompute,omitempty"`
}

type ETAResult struct {
	WaybillID          string     `json:"waybill_id"`
	WaybillNo          string     `json:"waybill_no"`
	CarrierID          string     `json:"carrier_id"`
	ServiceCode        string     `json:"service_code"`
	Timezone           string     `json:"timezone"`
	PickupDeadlineAt   *time.Time `json:"pickup_deadline_at,omitempty"`
	DeliveryDeadlineAt *time.Time `json:"delivery_deadline_at,omitempty"`
	PromisedAt         *time.Time `json:"promised_at,omitempty"`
	EstimatedAt        *time.Time `json:"estimated_at,omitempty"`
	Delayed            bool       `json:"delayed"`
	Source             string     `json:"source"`
	LastComputedAt     time.Time  `json:"last_computed_at"`
}

func (s *ETAService) GetByWaybillID(ctx context.Context, tenantUUID, waybillID string, query ETAQuery) (*ETAResult, error) {
	if s == nil || s.waybillRepo == nil || s.recordRepo == nil {
		return nil, errors.New("eta service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	waybillID = strings.TrimSpace(waybillID)
	if waybillID == "" {
		return nil, errors.New("waybill_id is required")
	}
	if !query.ForceRecompute {
		record, err := s.recordRepo.GetByWaybillID(ctx, waybillID)
		if err == nil {
			return toETAResult(record), nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	wb, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	computed, err := s.computeAndStore(ctx, wb, query)
	if err != nil {
		return nil, err
	}
	return toETAResult(computed), nil
}

func (s *ETAService) ListByWaybillIDs(ctx context.Context, tenantUUID string, waybillIDs []string, query ETAQuery) ([]ETAResult, error) {
	if s == nil || s.waybillRepo == nil || s.recordRepo == nil {
		return nil, errors.New("eta service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	ids := make([]string, 0, len(waybillIDs))
	for _, id := range waybillIDs {
		if v := strings.TrimSpace(id); v != "" {
			ids = append(ids, v)
		}
	}
	if len(ids) == 0 {
		return []ETAResult{}, nil
	}
	results := make([]ETAResult, 0, len(ids))
	recordsByID := map[string]*LogisticsModel.ETARecord{}
	if !query.ForceRecompute {
		records, err := s.recordRepo.ListByWaybillIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		for i := range records {
			r := records[i]
			recordsByID[r.WaybillID] = &r
		}
	}
	for _, id := range ids {
		if cached := recordsByID[id]; cached != nil {
			results = append(results, *toETAResult(cached))
			continue
		}
		wb, err := s.waybillRepo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		computed, err := s.computeAndStore(ctx, wb, query)
		if err != nil {
			return nil, err
		}
		results = append(results, *toETAResult(computed))
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].LastComputedAt.After(results[j].LastComputedAt)
	})
	return results, nil
}

func (s *ETAService) computeAndStore(ctx context.Context, wb *LogisticsModel.Waybill, query ETAQuery) (*LogisticsModel.ETARecord, error) {
	if wb == nil {
		return nil, gorm.ErrRecordNotFound
	}
	policy := s.resolvePolicy(ctx, wb, query.DestinationZone)
	locName := normalizeETATimezone(query.Timezone, policy)
	loc := mustLoadLocation(locName)
	pickupSLAHours := policy.PickupSLAHours
	if pickupSLAHours <= 0 {
		pickupSLAHours = 24
	}
	deliverySLAHours := policy.DeliverySLAHours
	if deliverySLAHours <= 0 {
		deliverySLAHours = 72
	}
	cutoffHour := policy.CutoffHourLocal
	if cutoffHour < 0 || cutoffHour > 23 {
		cutoffHour = defaultCutoffHour
	}

	createdLocal := wb.CreatedAt.In(loc)
	pickupBase := createdLocal
	if createdLocal.Hour() >= cutoffHour {
		nextDay := createdLocal.Add(24 * time.Hour)
		pickupBase = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, loc)
	}
	pickupDeadline := pickupBase.Add(time.Duration(pickupSLAHours) * time.Hour)
	deliveryDeadline := pickupDeadline.Add(time.Duration(deliverySLAHours) * time.Hour)
	estimatedAt := deliveryDeadline
	promisedAt := deliveryDeadline

	pickedAt, deliveredAt := s.resolveTimeline(ctx, wb.ID)
	if pickedAt != nil {
		estimatedAt = pickedAt.In(loc).Add(time.Duration(deliverySLAHours) * time.Hour)
	}
	if deliveredAt != nil {
		estimatedAt = deliveredAt.In(loc)
	}

	now := time.Now().UTC()
	status := strings.ToLower(strings.TrimSpace(wb.Status))

	meta, _ := jsonBytes(map[string]any{
		"status":             status,
		"destination_zone":   strings.ToUpper(strings.TrimSpace(query.DestinationZone)),
		"pickup_sla_hours":   pickupSLAHours,
		"delivery_sla_hours": deliverySLAHours,
		"cutoff_hour_local":  cutoffHour,
	}, []byte("{}"))

	record := &LogisticsModel.ETARecord{
		ID:                 utils.NewUUID(),
		WaybillID:          wb.ID,
		WaybillNo:          wb.WaybillNo,
		CarrierID:          wb.CarrierID,
		ServiceCode:        wb.ServiceCode,
		Timezone:           locName,
		PickupDeadlineAt:   ptrTimeInUTC(pickupDeadline),
		DeliveryDeadlineAt: ptrTimeInUTC(deliveryDeadline),
		PromisedAt:         ptrTimeInUTC(promisedAt),
		EstimatedAt:        ptrTimeInUTC(estimatedAt),
		Source:             "calculated",
		Version:            1,
		Metadata:           datatypes.JSON(meta),
		LastComputedAt:     now,
	}
	if err := s.recordRepo.UpsertByWaybillID(ctx, record); err != nil {
		return nil, err
	}
	fresh, err := s.recordRepo.GetByWaybillID(ctx, wb.ID)
	if err == nil {
		return fresh, nil
	}
	return record, nil
}

func (s *ETAService) resolvePolicy(ctx context.Context, wb *LogisticsModel.Waybill, zone string) LogisticsModel.ETAPolicy {
	policy := LogisticsModel.ETAPolicy{
		PickupSLAHours:   24,
		DeliverySLAHours: 72,
		CutoffHourLocal:  defaultCutoffHour,
		Timezone:         defaultETATimezone,
	}
	if s == nil || s.policyRepo == nil || wb == nil {
		return policy
	}
	row, err := s.policyRepo.FindByCarrierService(ctx, wb.CarrierID, wb.ServiceCode, zone)
	if err != nil || row == nil {
		return policy
	}
	return *row
}

func (s *ETAService) resolveTimeline(ctx context.Context, waybillID string) (*time.Time, *time.Time) {
	if s == nil || s.trackRepo == nil || strings.TrimSpace(waybillID) == "" {
		return nil, nil
	}
	events, err := s.trackRepo.ListByWaybillID(ctx, waybillID)
	if err != nil {
		return nil, nil
	}
	var pickedAt *time.Time
	var deliveredAt *time.Time
	for _, evt := range events {
		if evt.OccurredAt == nil {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(evt.Status))
		if status == "picked" || status == "in_transit" || status == "in-transit" {
			if pickedAt == nil || evt.OccurredAt.Before(*pickedAt) {
				t := *evt.OccurredAt
				pickedAt = &t
			}
		}
		if status == "delivered" || status == "signed" {
			if deliveredAt == nil || evt.OccurredAt.Before(*deliveredAt) {
				t := *evt.OccurredAt
				deliveredAt = &t
			}
		}
	}
	return pickedAt, deliveredAt
}

func normalizeETATimezone(queryTZ string, policy LogisticsModel.ETAPolicy) string {
	if tz := strings.TrimSpace(queryTZ); tz != "" {
		return tz
	}
	if tz := strings.TrimSpace(policy.Timezone); tz != "" {
		return tz
	}
	return defaultETATimezone
}

func mustLoadLocation(name string) *time.Location {
	if strings.TrimSpace(name) == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

func ptrTimeInUTC(v time.Time) *time.Time {
	u := v.UTC()
	return &u
}

func toETAResult(record *LogisticsModel.ETARecord) *ETAResult {
	if record == nil {
		return nil
	}
	now := time.Now().UTC()
	delayed := record.PromisedAt != nil && record.PromisedAt.UTC().Before(now)
	return &ETAResult{
		WaybillID:          record.WaybillID,
		WaybillNo:          record.WaybillNo,
		CarrierID:          record.CarrierID,
		ServiceCode:        record.ServiceCode,
		Timezone:           record.Timezone,
		PickupDeadlineAt:   record.PickupDeadlineAt,
		DeliveryDeadlineAt: record.DeliveryDeadlineAt,
		PromisedAt:         record.PromisedAt,
		EstimatedAt:        record.EstimatedAt,
		Delayed:            delayed,
		Source:             record.Source,
		LastComputedAt:     record.LastComputedAt,
	}
}
