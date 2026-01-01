package product_sku

import (
	"context"
	"errors"
	"strings"
	"time"
)

const inventoryStaleThreshold = 5 * time.Minute

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
