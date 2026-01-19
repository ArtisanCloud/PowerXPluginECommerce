package payments

import "errors"

var (
	ErrPaymentServiceUnavailable = errors.New("payment service unavailable")
	ErrInvalidArgument           = errors.New("invalid argument")
	ErrOrderNotFound              = errors.New("order not found")
	ErrOrderNotPayable            = errors.New("order is not payable")
	ErrProviderUnavailable        = errors.New("payment provider unavailable")
	ErrTransactionNotFound        = errors.New("payment transaction not found")
)

type CreateTransactionRequest struct {
	OrderID        string `json:"orderId"`
	OrderNo        string `json:"orderNo"`
	AmountMinor    int64  `json:"amountMinor"`
	Currency       string `json:"currency"`
	PayMethod      string `json:"payMethod"`
	OpenID         string `json:"openid"`
	Client         string `json:"client"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type WechatPayParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

type CreateTransactionResponse struct {
	TransactionID string           `json:"transactionId"`
	Status        string           `json:"status"`
	Wechat        *WechatPayParams `json:"wechat,omitempty"`
}

type TransactionStatusResponse struct {
	TransactionID string `json:"transactionId"`
	OrderID       string `json:"orderId"`
	OrderNo       string `json:"orderNo"`
	Status        string `json:"status"`
	FailureReason string `json:"failureReason,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}
