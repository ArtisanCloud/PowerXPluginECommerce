package product_sku

import "errors"

var (
	// ErrServiceUnavailable indicates SKU service is not ready.
	ErrServiceUnavailable = errors.New("product SKU service unavailable")
)
