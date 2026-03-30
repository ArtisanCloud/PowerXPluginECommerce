package logistics

import (
	"context"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type BillingFilter struct {
	CarrierID string
	From      *time.Time
	To        *time.Time
}

type BillingCarrierSummary struct {
	CarrierID     string  `json:"carrier_id"`
	CarrierName   string  `json:"carrier_name"`
	WaybillCount  int64   `json:"waybill_count"`
	EstimatedFee  float64 `json:"estimated_fee"`
	ActualFee     float64 `json:"actual_fee"`
	DiffFee       float64 `json:"diff_fee"`
	AbnormalCount int64   `json:"abnormal_count"`
}

type BillingWaybillItem struct {
	WaybillID     string    `json:"waybill_id"`
	OrderID       string    `json:"order_id"`
	WaybillNo     string    `json:"waybill_no"`
	CarrierID     string    `json:"carrier_id"`
	CarrierName   string    `json:"carrier_name"`
	Status        string    `json:"status"`
	EstimatedFee  float64   `json:"estimated_fee"`
	ActualFee     float64   `json:"actual_fee"`
	DiffFee       float64   `json:"diff_fee"`
	BillingStatus string    `json:"billing_status"`
	CreatedAt     time.Time `json:"created_at"`
}

// BillingRepository provides read models for reconciliation views.
type BillingRepository struct {
	waybillRepo *WaybillRepository
}

func NewBillingRepository(waybillRepo *WaybillRepository) *BillingRepository {
	return &BillingRepository{waybillRepo: waybillRepo}
}

func (r *BillingRepository) ListWaybills(ctx context.Context, filter BillingFilter) ([]BillingWaybillItem, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	db := r.waybillRepo.DB.WithContext(ctx).
		Table(LogisticsModel.Waybill{}.TableName()+" AS wb").
		Select(`
			wb.id AS waybill_id,
			wb.order_id AS order_id,
			wb.waybill_no AS waybill_no,
			wb.carrier_id AS carrier_id,
			c.name AS carrier_name,
			wb.status AS status,
			wb.fee_amount AS estimated_fee,
			wb.actual_fee_amount AS actual_fee,
			wb.fee_diff_amount AS diff_fee,
			wb.billing_status AS billing_status,
			wb.created_at AS created_at
		`).
		Joins("LEFT JOIN "+LogisticsModel.Carrier{}.TableName()+" AS c ON c.id = wb.carrier_id").
		Where("wb.tenant_uuid = ?", tenantUUID)
	db = applyBillingFilter(db, filter)

	var rows []BillingWaybillItem
	err = db.Order("wb.created_at DESC").Scan(&rows).Error
	return rows, err
}

func (r *BillingRepository) SummarizeByCarrier(ctx context.Context, filter BillingFilter) ([]BillingCarrierSummary, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	db := r.waybillRepo.DB.WithContext(ctx).
		Table(LogisticsModel.Waybill{}.TableName()+" AS wb").
		Select(`
			wb.carrier_id AS carrier_id,
			COALESCE(c.name, wb.carrier_id) AS carrier_name,
			COUNT(1) AS waybill_count,
			COALESCE(SUM(wb.fee_amount), 0) AS estimated_fee,
			COALESCE(SUM(wb.actual_fee_amount), 0) AS actual_fee,
			COALESCE(SUM(wb.fee_diff_amount), 0) AS diff_fee,
			SUM(CASE WHEN ABS(wb.fee_diff_amount) > 0.0001 THEN 1 ELSE 0 END) AS abnormal_count
		`).
		Joins("LEFT JOIN "+LogisticsModel.Carrier{}.TableName()+" AS c ON c.id = wb.carrier_id").
		Where("wb.tenant_uuid = ?", tenantUUID)
	db = applyBillingFilter(db, filter)
	db = db.Group("wb.carrier_id, c.name")

	var rows []BillingCarrierSummary
	err = db.Order("estimated_fee DESC").Scan(&rows).Error
	return rows, err
}

func applyBillingFilter(db *gorm.DB, filter BillingFilter) *gorm.DB {
	if strings.TrimSpace(filter.CarrierID) != "" {
		db = db.Where("wb.carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if filter.From != nil {
		db = db.Where("wb.created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		db = db.Where("wb.created_at <= ?", *filter.To)
	}
	return db
}
