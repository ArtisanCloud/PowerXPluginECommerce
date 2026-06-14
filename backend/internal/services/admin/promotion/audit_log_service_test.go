package promotion

import (
	"context"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

func TestAuditLogService_ListByPromotion(t *testing.T) {
	db := setupPromotionDB(t, "file:promotion_audit_service?mode=memory&cache=shared")
	svc := NewAuditLogService(&app.Deps{DB: db})
	ctx := context.Background()

	if err := svc.Log(ctx, nil, "tenant-1", "promo-1", "order-1", "quote", "applied", "user:1", "req-1", map[string]any{"discount": 100}); err != nil {
		t.Fatalf("log audit: %v", err)
	}
	if err := svc.Log(ctx, nil, "tenant-2", "promo-1", "order-2", "quote", "applied", "user:2", "req-2", nil); err != nil {
		t.Fatalf("log tenant2 audit: %v", err)
	}

	rows, total, err := svc.List(ctx, "tenant-1", "promo-1", 1, 20)
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("expected tenant filtered audit row, total=%d rows=%d", total, len(rows))
	}
	if rows[0].OrderID == nil || *rows[0].OrderID != "order-1" || rows[0].RequestID != "req-1" {
		t.Fatalf("unexpected audit row: %+v", rows[0])
	}
}
