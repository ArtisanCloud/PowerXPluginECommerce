package migrations

import couponModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"

// CouponTables enumerates coupon tables for automated migrations.
var CouponTables = []interface{}{
	&couponModel.CouponTemplate{},
	&couponModel.CouponAsset{},
	&couponModel.CouponUsageLog{},
	&couponModel.OrderCouponSnapshot{},
}
