package cart

import "time"

type CartItemInput struct {
	SKUID string `json:"skuId"`
	Qty   int64  `json:"qty"`
}

type CartDTO struct {
	Items     []CartItemInput `json:"items"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type CartSyncRequest struct {
	Items    []CartItemInput `json:"items"`
	Strategy string          `json:"strategy,omitempty"` // max | overwrite
}
