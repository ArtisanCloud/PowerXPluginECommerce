package payments

import "errors"

var (
	ErrPaymentServiceUnavailable              = errors.New("payment service unavailable")
	ErrInvalidArgument                        = errors.New("invalid argument")
	ErrOrderNotFound                          = errors.New("order not found")
	ErrOrderNotPayable                        = errors.New("order is not payable")
	ErrOrderCustomerMismatch                  = errors.New("order customer mismatch")
	ErrCustomerRequired                       = errors.New("customer required")
	ErrCustomerIdentityNotFound               = errors.New("customer identity not found")
	ErrProviderUnavailable                    = errors.New("payment provider unavailable")
	ErrProviderSelectorRequired               = errors.New("payment provider selector required")
	ErrProviderCredentialsDecryptFailed       = errors.New("payment provider credentials decrypt failed")
	ErrTransactionNotFound                    = errors.New("payment transaction not found")
	ErrOrderNoMismatch                        = errors.New("order number mismatch")
	ErrOpenIDRequired                         = errors.New("openid is required")
	ErrWechatCredentialsMissingRequiredFields = errors.New("wechat credentials missing required fields")
	ErrWechatCredentialsMissingPrivateKey     = errors.New("wechat credentials missing private key")
	ErrWechatCredentialsMissingNotifyURL      = errors.New("wechat credentials missing notify url")
)

type CreateTransactionRequest struct {
	OrderID        string `json:"orderId"`
	PayMethod      string `json:"payMethod"`
	Client         string `json:"client,omitempty"`
	IdempotencyKey string `json:"idempotencyKey"`
	ProviderID     uint64 `json:"providerId"`
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
