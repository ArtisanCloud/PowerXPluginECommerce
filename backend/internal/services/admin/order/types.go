package order

import "time"

type CreateOrderItemInput struct {
	SKUID string `json:"skuId"`
	Qty   int64  `json:"qty"`
}

type CreateOrderRequest struct {
	CustomerID        string                 `json:"customerId"`
	Channel           string                 `json:"channel"`
	ShippingAddressID string                 `json:"shippingAddressId,omitempty"`
	ShippingAddress   *ShippingAddress       `json:"shippingAddress,omitempty"`
	Items             []CreateOrderItemInput `json:"items"`
	Note              string                 `json:"note,omitempty"`
}

type ShippingAddress struct {
	Label          string         `json:"label,omitempty"`
	RecipientName  string         `json:"recipientName"`
	RecipientPhone string         `json:"recipientPhone"`
	CountryCode    string         `json:"countryCode,omitempty"`
	Province       string         `json:"province,omitempty"`
	City           string         `json:"city,omitempty"`
	District       string         `json:"district,omitempty"`
	Address1       string         `json:"address1"`
	Address2       string         `json:"address2,omitempty"`
	PostalCode     string         `json:"postalCode,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type MoneyDTO struct {
	Currency string `json:"currency"`
	Subtotal int64  `json:"subtotal"`
	Total    int64  `json:"total"`
}

type OrderSummaryDTO struct {
	OrderID                 string           `json:"orderId"`
	OrderNo                 string           `json:"orderNo"`
	CustomerID              string           `json:"customerId"`
	Channel                 string           `json:"channel"`
	CreatedByType           string           `json:"createdByType"`
	Status                  string           `json:"status"`
	Amounts                 MoneyDTO         `json:"amounts"`
	ShippingAddressSnapshot *ShippingAddress `json:"shippingAddressSnapshot,omitempty"`
	CreatedAt               time.Time        `json:"createdAt"`
}

type OrderListResponse struct {
	Items    []OrderSummaryDTO `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	Total    int64             `json:"total"`
}

type OrderItemDTO struct {
	SKUID      string `json:"skuId"`
	Qty        int64  `json:"qty"`
	UnitPrice  int64  `json:"unitPrice"`
	LineAmount int64  `json:"lineAmount"`
}

type OrderEventDTO struct {
	EventType    string    `json:"eventType"`
	OperatorType string    `json:"operatorType"`
	Operator     string    `json:"operator,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type OrderDetailDTO struct {
	Summary OrderSummaryDTO `json:"summary"`
	Items   []OrderItemDTO  `json:"items"`
	Events  []OrderEventDTO `json:"events"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason,omitempty"`
}
