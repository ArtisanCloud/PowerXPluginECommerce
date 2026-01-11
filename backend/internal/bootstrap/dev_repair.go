package bootstrap

import (
	"context"
	"fmt"
	"strings"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"gorm.io/gorm"
)

// RepairDraftSKUsForPublishedSPUs 将已发布 SPU 下的 draft SKU 纠正为 online（仅用于本地/开发环境自愈）。
func RepairDraftSKUsForPublishedSPUs(ctx context.Context, db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, nil
	}
	if strings.EqualFold(strings.TrimSpace(db.Dialector.Name()), "sqlite") {
		return 0, nil
	}

	spuTable := productmodel.SPU{}.TableName()
	skuTable := productskumodel.ProductSKU{}.TableName()
	sql := fmt.Sprintf(
		`UPDATE %s AS s
SET status = 'online', updated_at = NOW()
FROM %s AS p
WHERE s.tenant_uuid = p.tenant_uuid
  AND s.spu_id = p.id
  AND s.deleted_at IS NULL
  AND p.deleted_at IS NULL
  AND p.status = 'published'
  AND s.status = 'draft'`,
		skuTable,
		spuTable,
	)
	res := db.WithContext(ctx).Exec(sql)
	return res.RowsAffected, res.Error
}
