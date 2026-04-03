package coupon

import "strings"

const (
	AssetStatusAvailable = "available"
	AssetStatusReserved  = "reserved"
	AssetStatusRedeemed  = "redeemed"
	AssetStatusRefunded  = "refunded"
	AssetStatusExpired   = "expired"
)

func CanTransitAssetStatus(from, to string) bool {
	from = strings.TrimSpace(strings.ToLower(from))
	to = strings.TrimSpace(strings.ToLower(to))
	if from == "" || to == "" {
		return false
	}
	switch from {
	case AssetStatusAvailable:
		return to == AssetStatusReserved || to == AssetStatusExpired
	case AssetStatusReserved:
		return to == AssetStatusAvailable || to == AssetStatusRedeemed || to == AssetStatusExpired
	case AssetStatusRedeemed:
		return to == AssetStatusRefunded
	default:
		return false
	}
}
