package product_sku

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

const inventoryStaleThreshold = 5 * time.Minute
const defaultWarehouseID = "default"

var (
	ErrInventoryDeltaRequired   = errors.New("inventory delta is required")
	ErrInventoryWouldBeNegative = errors.New("inventory would be negative")
)

// SnapshotInventory will load the latest inventory state for a SKU.
func (s *Service) SnapshotInventory(ctx context.Context, skuID string) (*InventorySnapshot, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}

	rows, err := s.InventoryRepo.ListBySKU(ctx, tenantID, skuID)
	if err != nil {
		return nil, err
	}

	snapshot := &InventorySnapshot{
		Warehouses: make([]InventoryWarehouseSnapshot, 0, len(rows)),
	}
	var total InventoryWarehouseSnapshot
	var latest *time.Time

	for _, row := range rows {
		alert := ""
		if row.AlertThreshold > 0 && row.AvailableQty < row.AlertThreshold {
			alert = "alert"
		} else if row.SafetyStock > 0 && row.AvailableQty < row.SafetyStock {
			alert = "warning"
		}
		item := InventoryWarehouseSnapshot{
			WarehouseID:  row.WarehouseID,
			AvailableQty: row.AvailableQty,
			LockedQty:    row.LockedQty,
			InTransitQty: row.InTransitQty,
			SafetyStock:  row.SafetyStock,
			AlertLevel:   alert,
			LastSyncedAt: row.LastSyncedAt,
		}
		snapshot.Warehouses = append(snapshot.Warehouses, item)
		total.AvailableQty += row.AvailableQty
		total.LockedQty += row.LockedQty
		total.InTransitQty += row.InTransitQty
		total.SafetyStock += row.SafetyStock

		if row.LastSyncedAt != nil {
			if latest == nil || row.LastSyncedAt.After(*latest) {
				ts := *row.LastSyncedAt
				latest = &ts
			}
		}
	}
	snapshot.Total = total
	snapshot.LastSyncedAt = latest
	if latest != nil && time.Since(*latest) > inventoryStaleThreshold {
		snapshot.IsStale = true
	}
	return snapshot, nil
}

// AdjustInventory applies an incremental delta adjustment to the available inventory for default warehouse.
// Delta MUST be an integer unit count and MUST NOT result in negative available inventory.
func (s *Service) AdjustInventory(ctx context.Context, skuID string, delta int64) (*InventorySnapshot, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	if delta == 0 {
		return nil, ErrInventoryDeltaRequired
	}

	actor := ""
	if tc, ok := authx.TenantContextFromContext(ctx); ok && tc.UserID > 0 {
		actor = fmt.Sprintf("%d", tc.UserID)
	}
	requestID, _ := ctx.Value("request_id").(string)

	err = s.InventoryRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var sku productskumodel.ProductSKU
		if err := tx.WithContext(ctx).
			Where("tenant_uuid = ? AND id = ?", tenantID, skuID).
			First(&sku).Error; err != nil {
			return err
		}

		row, err := s.InventoryRepo.EnsureWarehouseRowForUpdate(ctx, tx, tenantID, skuID, defaultWarehouseID)
		if err != nil {
			return err
		}

		before := row.AvailableQty
		after := before + delta
		if after < 0 {
			return ErrInventoryWouldBeNegative
		}

		if err := tx.WithContext(ctx).
			Model(&productskumodel.ProductSKUInventory{}).
			Where("id = ?", row.ID).
			Update("available_qty", after).Error; err != nil {
			return err
		}

		if s.AuditLogRepo != nil {
			log := &productskumodel.ProductSKUAuditLog{
				TenantUUID:      tenantID,
				SKUId:           skuID,
				WarehouseID:     defaultWarehouseID,
				Action:          "inventory.adjust",
				Actor:           actor,
				Delta:           delta,
				BeforeAvailable: before,
				AfterAvailable:  after,
				RequestID:       strings.TrimSpace(requestID),
			}
			if err := s.AuditLogRepo.CreateWithTx(ctx, tx, log); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.SnapshotInventory(ctx, skuID)
}
