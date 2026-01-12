package sellability

// ReasonCode is a stable string enum returned to mini-app clients.
// New values must remain backward compatible (clients should treat unknown values as UNKNOWN).
type ReasonCode string

const (
	ReasonSPUNotPublished         ReasonCode = "SPU_NOT_PUBLISHED"
	ReasonSKUNotOnline            ReasonCode = "SKU_NOT_ONLINE"
	ReasonNoPublicPrice           ReasonCode = "NO_PUBLIC_PRICE"
	ReasonOutOfStock              ReasonCode = "OUT_OF_STOCK"
	ReasonChannelDisabled         ReasonCode = "CHANNEL_DISABLED"
	ReasonNotInAvailabilityWindow ReasonCode = "NOT_IN_AVAILABILITY_WINDOW"
	ReasonChannelStatusBlocked    ReasonCode = "CHANNEL_STATUS_BLOCKED"
	ReasonUnknown                 ReasonCode = "UNKNOWN"
)

// MoneyDTO represents the display price chosen by pricing rules.
type MoneyDTO struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// ItemDTO represents sellability for a single SKU under a specific channel.
type ItemDTO struct {
	SKUID        string    `json:"skuId"`
	Sellable     bool      `json:"sellable"`
	Reasons      []string  `json:"reasons"`
	Price        *MoneyDTO `json:"price"`
	AvailableQty int       `json:"availableQty"`
}

// ResultDTO represents sellability evaluation for an SPU under a specific channel.
type ResultDTO struct {
	SPUID   string    `json:"spuId"`
	Channel string    `json:"channel"`
	Items   []ItemDTO `json:"items"`
}

// SummaryDTO is a compact per-SPU summary suitable for list pages.
type SummaryDTO struct {
	Sellable bool     `json:"sellable"`
	Reasons  []string `json:"reasons"`
}
