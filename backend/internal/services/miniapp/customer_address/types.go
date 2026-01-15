package customer_address

import "time"

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

type CustomerAddressDTO struct {
	ID         string          `json:"id"`
	CustomerID string          `json:"customerId"`
	IsDefault  bool            `json:"isDefault"`
	Address    ShippingAddress `json:"shippingAddress"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

type AddressUpsertRequest struct {
	IsDefault       *bool           `json:"isDefault,omitempty"`
	ShippingAddress ShippingAddress `json:"shippingAddress"`
}
