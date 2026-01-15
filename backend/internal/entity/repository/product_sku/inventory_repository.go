package product_sku

import (
	"context"
	"errors"
	"strings"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InventoryRepository manages per-warehouse inventory snapshots.
type InventoryRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUInventory]
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUInventory](db)}
}

// ListBySKU returns all warehouse snapshots for the sku.
func (r *InventoryRepository) ListBySKU(ctx context.Context, tenantID, skuID string) ([]productskumodel.ProductSKUInventory, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("inventory repository is not initialized")
	}
	var rows []productskumodel.ProductSKUInventory
	err := r.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		return tx.
			Where("tenant_uuid = ? AND sku_id = ?", tenantID, skuID).
			Order("warehouse_id ASC").
			Find(&rows).Error
	})
	return rows, err
}

// EnsureWarehouseRowForUpdate ensures the inventory row exists and returns it locked FOR UPDATE.
// Caller MUST execute this within a tenant transaction (BeginTenantTx/WithTenantTx).
func (r *InventoryRepository) EnsureWarehouseRowForUpdate(
	ctx context.Context,
	tx *gorm.DB,
	tenantID, skuID, warehouseID string,
) (*productskumodel.ProductSKUInventory, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("inventory repository is not initialized")
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	skuID = strings.TrimSpace(skuID)
	warehouseID = strings.TrimSpace(warehouseID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	if warehouseID == "" {
		return nil, errors.New("warehouse id is required")
	}

	seed := &productskumodel.ProductSKUInventory{
		ID:           uuid.NewString(),
		TenantUUID:   tenantID,
		SKUId:        skuID,
		WarehouseID:  warehouseID,
		AvailableQty: 0,
		LockedQty:    0,
		InTransitQty: 0,
		SafetyStock:  0,
	}
	if err := tx.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(seed).Error; err != nil {
		return nil, err
	}

	var row productskumodel.ProductSKUInventory
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_uuid = ? AND sku_id = ? AND warehouse_id = ?", tenantID, skuID, warehouseID).
		First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// LockInventory increases locked_qty for a SKU in the given warehouse within the provided transaction.
// Caller MUST execute this within a tenant transaction (BeginTenantTx/WithTenantTx).
func (r *InventoryRepository) LockInventory(ctx context.Context, tx *gorm.DB, tenantID, skuID, warehouseID string, qty int64) (*productskumodel.ProductSKUInventory, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("inventory repository is not initialized")
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	if qty <= 0 {
		return nil, errors.New("qty must be positive")
	}
	row, err := r.EnsureWarehouseRowForUpdate(ctx, tx, tenantID, skuID, warehouseID)
	if err != nil {
		return nil, err
	}
	after := row.LockedQty + qty
	if after < 0 {
		after = 0
	}
	if err := tx.WithContext(ctx).
		Model(&productskumodel.ProductSKUInventory{}).
		Where("id = ?", row.ID).
		Update("locked_qty", after).Error; err != nil {
		return nil, err
	}
	row.LockedQty = after
	return row, nil
}

// UnlockInventory decreases locked_qty for a SKU in the given warehouse within the provided transaction.
// Caller MUST execute this within a tenant transaction (BeginTenantTx/WithTenantTx).
func (r *InventoryRepository) UnlockInventory(ctx context.Context, tx *gorm.DB, tenantID, skuID, warehouseID string, qty int64) (*productskumodel.ProductSKUInventory, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("inventory repository is not initialized")
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	if qty <= 0 {
		return nil, errors.New("qty must be positive")
	}
	row, err := r.EnsureWarehouseRowForUpdate(ctx, tx, tenantID, skuID, warehouseID)
	if err != nil {
		return nil, err
	}
	after := row.LockedQty - qty
	if after < 0 {
		after = 0
	}
	if err := tx.WithContext(ctx).
		Model(&productskumodel.ProductSKUInventory{}).
		Where("id = ?", row.ID).
		Update("locked_qty", after).Error; err != nil {
		return nil, err
	}
	row.LockedQty = after
	return row, nil
}
