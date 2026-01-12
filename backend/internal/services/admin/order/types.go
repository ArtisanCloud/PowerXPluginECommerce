package order

import "time"

type CreateOrderItemInput struct {
	SKUID string `json:"skuId"`
	Qty   int64  `json:"qty"`
}

type CreateOrderRequest struct {
	CustomerID string                 `json:"customerId"`
	Channel    string                 `json:"channel"`
	Items      []CreateOrderItemInput `json:"items"`
	Note       string                 `json:"note,omitempty"`
}

type MoneyDTO struct {
	Currency string `json:"currency"`
	Subtotal int64  `json:"subtotal"`
	Total    int64  `json:"total"`
}

type OrderSummaryDTO struct {
	OrderID   string    `json:"orderId"`
	OrderNo   string    `json:"orderNo"`
	Status    string    `json:"status"`
	Amounts   MoneyDTO  `json:"amounts"`
	CreatedAt time.Time `json:"createdAt"`
}

type OrderListResponse struct {
	Items    []OrderSummaryDTO `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	Total    int64             `json:"total"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason,omitempty"`
}
