package http

import (
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	admincoupon "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/coupon"
)

func TestMarketplaceRBACEntriesContainListings(t *testing.T) {
	entries := marketplacePublicRBACEntries("/api/v1")
	perm, ok := entries["GET:/api/v1/marketplace/listings"]
	if !ok {
		t.Fatalf("expected listings GET entry in RBAC map, got %+v", entries)
	}
	if perm.Resource != "marketplace.listings" || perm.Action != "read" {
		t.Fatalf("unexpected permission %+v", perm)
	}
}

func TestRegistryIncludeCouponAdminRBAC(t *testing.T) {
	r := &Registry{rbac: map[string]Permission{}}
	r.mergeRBAC(couponRBACEntriesForTest("/api/v1"))
	perm, ok := r.rbac["GET:/api/v1/admin/coupons/templates"]
	if !ok {
		t.Fatalf("expected coupon template rbac entry in registry, got %+v", r.rbac)
	}
	if perm.Resource != "com.powerx.plugins.ecommerce:pricing.coupon.template" || perm.Action != "read" {
		t.Fatalf("unexpected permission %+v", perm)
	}
}

type Permission = authx.Permission

func couponRBACEntriesForTest(prefix string) map[string]Permission {
	return admincoupon.RBACEntries(prefix)
}
