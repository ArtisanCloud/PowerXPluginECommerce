package payments

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerLibs/v3/object"
	wxmodels "github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/payment/notify/request"
	orderrequest "github.com/ArtisanCloud/PowerWeChat/v3/src/payment/order/request"
	pxmodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	customerrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/customer"
	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	paymentslogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TransactionService struct {
	deps         *app.Deps
	transactions *paymentrepo.PaymentTransactionRepository
	providers    *paymentrepo.PaymentProviderRepository
	orders       *orderrepo.OrderRepository
	identities   *customerrepo.IdentityRepository
}

func NewTransactionService(deps *app.Deps) *TransactionService {
	if deps == nil || deps.DB == nil {
		return &TransactionService{deps: deps}
	}
	return &TransactionService{
		deps:         deps,
		transactions: paymentrepo.NewPaymentTransactionRepository(deps.DB),
		providers:    paymentrepo.NewPaymentProviderRepository(deps.DB),
		orders:       orderrepo.NewOrderRepository(deps.DB),
		identities:   customerrepo.NewIdentityRepository(deps.DB),
	}
}

func (s *TransactionService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.transactions != nil && s.orders != nil && s.identities != nil
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tenantUUID string, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	if !s.Ready() {
		return nil, ErrPaymentServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, ErrInvalidArgument
	}
	ctx = authx.ContextWithTenantUUID(ctx, tenantUUID)
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.PayMethod = strings.TrimSpace(req.PayMethod)
	req.Client = strings.TrimSpace(req.Client)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.OrderID == "" || req.PayMethod == "" || req.IdempotencyKey == "" {
		return nil, ErrInvalidArgument
	}
	if req.ProviderID == 0 {
		return nil, ErrProviderSelectorRequired
	}
	if req.Client == "" {
		req.Client = "miniapp"
	}
	customerCtx, ok := authx.CustomerFromContext(ctx)
	if !ok || strings.TrimSpace(customerCtx.CustomerID) == "" {
		return nil, ErrCustomerRequired
	}

	tx := s.deps.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() { _ = tx.Rollback() }()

	orderRow, err := s.orders.LockByID(ctx, tx, tenantUUID, req.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(orderRow.CustomerID) == "" || strings.TrimSpace(orderRow.CustomerID) != strings.TrimSpace(customerCtx.CustomerID) {
		return nil, ErrOrderCustomerMismatch
	}
	if strings.TrimSpace(orderRow.Status) != "pending_payment" {
		return nil, ErrOrderNotPayable
	}

	provider, err := s.resolveProvider(ctx, tx, tenantUUID, req.PayMethod, req.ProviderID, "", "", "")
	if err != nil {
		return nil, err
	}
	providerID := provider.ID
	providerType := strings.TrimSpace(provider.ProviderType)

	openID := ""
	if mapProviderType(req.PayMethod) == "wechat" {
		identity, err := s.identities.FindByCustomerProviderApp(ctx, strings.TrimSpace(customerCtx.CustomerID), "wechat", strings.TrimSpace(provider.AppID))
		if err != nil {
			return nil, err
		}
		if identity == nil || strings.TrimSpace(identity.Subject) == "" {
			return nil, ErrCustomerIdentityNotFound
		}
		openID = strings.TrimSpace(identity.Subject)
	}

	amountMinor := orderRow.TotalAmount
	if amountMinor <= 0 {
		return nil, ErrInvalidArgument
	}
	if mapProviderType(req.PayMethod) == "wechat" {
		// 微信调试最小金额：强制改为 0.03 元（3 分）
		amountMinor = 3
	}

	metadata := map[string]any{
		"client":          req.Client,
		"openid":          openID,
		"idempotency_key": req.IdempotencyKey,
		"provider_type":   providerType,
		"provider_mch_id": strings.TrimSpace(provider.MchID),
		"provider_app_id": strings.TrimSpace(provider.AppID),
	}
	metadataRaw, _ := json.Marshal(metadata)
	transaction := &pxmodels.PaymentTransaction{
		BaseModel:      pxmodels.BaseModel{TenantUuid: tenantUUID},
		TransactionNo:  newTransactionNo(orderRow.OrderNo),
		OrderID:        req.OrderID,
		OrderNo:        orderRow.OrderNo,
		ProviderID:     providerID,
		PayMethod:      req.PayMethod,
		AmountTotal:    amountMinor,
		AmountCurrency: strings.TrimSpace(orderRow.Currency),
		FeeAmount:      0,
		Status:         "pending_payment",
		Metadata:       datatypes.JSON(metadataRaw),
	}

	if err := tx.WithContext(ctx).Create(transaction).Error; err != nil {
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	resp, err := s.buildCreateResponse(ctx, transaction, provider, openID)
	if err != nil {
		_ = s.markTransactionFailed(ctx, tenantUUID, transaction.ID, err)
		return nil, err
	}
	return resp, nil
}

func (s *TransactionService) GetTransactionStatus(ctx context.Context, tenantUUID string, transactionID string) (*TransactionStatusResponse, error) {
	if !s.Ready() {
		return nil, ErrPaymentServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, ErrInvalidArgument
	}
	id := strings.TrimSpace(transactionID)
	if id == "" {
		return nil, ErrInvalidArgument
	}
	idNum, err := strconv.ParseUint(id, 10, 64)
	if err != nil || idNum == 0 {
		return nil, ErrInvalidArgument
	}

	var row pxmodels.PaymentTransaction
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, idNum).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	updatedAt := row.UpdatedAt.UTC().Format(time.RFC3339)
	return &TransactionStatusResponse{
		TransactionID: strconv.FormatUint(row.ID, 10),
		OrderID:       strings.TrimSpace(row.OrderID),
		OrderNo:       strings.TrimSpace(row.OrderNo),
		Status:        strings.TrimSpace(row.Status),
		FailureReason: strings.TrimSpace(row.FailureReason),
		UpdatedAt:     updatedAt,
	}, nil
}

func (s *TransactionService) HandleProviderCallback(ctx context.Context, tenantUUID string, providerID uint64, payload map[string]any) error {
	if !s.Ready() {
		return ErrPaymentServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return ErrInvalidArgument
	}
	orderNo := pickString(payload, "out_trade_no", "order_no", "orderNo", "order_no", "order_no")
	transactionNo := pickString(payload, "transaction_no", "transactionNo", "transaction_id")
	tradeState := strings.ToUpper(strings.TrimSpace(pickString(payload, "trade_state", "status", "tradeState")))

	if orderNo == "" && transactionNo == "" {
		return ErrInvalidArgument
	}

	var row pxmodels.PaymentTransaction
	query := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if transactionNo != "" {
		query = query.Where("transaction_no = ?", transactionNo)
	} else {
		query = query.Where("order_no = ?", orderNo)
	}
	if providerID > 0 {
		query = query.Where("provider_id = ?", providerID)
	}
	if err := query.First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTransactionNotFound
		}
		return err
	}

	if isTerminalStatus(row.Status) {
		s.emitCallbackEvent(ctx, tenantUUID, &row, providerID, "ignored", "")
		return nil
	}

	status := mapTradeState(tradeState)
	if status == "" {
		s.emitCallbackEvent(ctx, tenantUUID, &row, providerID, "ignored", "unknown_state")
		return nil
	}

	reason := pickString(payload, "failure_reason", "reason", "trade_state_desc")
	if err := s.applyStatusUpdate(ctx, tenantUUID, &row, status, reason); err != nil {
		s.emitCallbackEvent(ctx, tenantUUID, &row, providerID, "failed", reason)
		return err
	}
	s.emitCallbackEvent(ctx, tenantUUID, &row, providerID, "updated", reason)
	return nil
}

func (s *TransactionService) HandleProviderCallbackRequest(ctx context.Context, tenantUUID string, providerID uint64, req *http.Request) (*http.Response, error) {
	if !s.Ready() {
		return nil, ErrPaymentServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" || req == nil {
		if req == nil {
			return nil, ErrInvalidArgument
		}
	}
	provider, err := s.loadProvider(ctx, nil, providerID)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, ErrProviderUnavailable
	}
	if tenantUUID == "" {
		tenantUUID = strings.TrimSpace(provider.TenantUuid)
	}
	if strings.TrimSpace(provider.TenantUuid) != "" && provider.TenantUuid != tenantUUID {
		return nil, ErrProviderUnavailable
	}
	if provider != nil && strings.EqualFold(strings.TrimSpace(provider.ProviderType), "wechat") {
		return s.handleWechatCallback(ctx, tenantUUID, provider, req)
	}
	var payload map[string]any
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidArgument
	}
	if err := s.HandleProviderCallback(ctx, tenantUUID, providerID, payload); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *TransactionService) HandleProviderCallbackRequestBySelector(ctx context.Context, tenantUUID, providerType, mchID, appID string, req *http.Request) (*http.Response, error) {
	if !s.Ready() {
		return nil, ErrPaymentServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if req == nil {
		return nil, ErrInvalidArgument
	}
	provider, err := s.loadProviderBySelector(ctx, nil, tenantUUID, providerType, mchID, appID)
	if err != nil {
		return nil, err
	}
	return s.HandleProviderCallbackRequest(ctx, tenantUUID, provider.ID, req)
}

func (s *TransactionService) loadProviderBySelector(ctx context.Context, tx *gorm.DB, tenantUUID, providerType, mchID, appID string) (*pxmodels.PaymentProvider, error) {
	if strings.TrimSpace(providerType) == "" || strings.TrimSpace(mchID) == "" || strings.TrimSpace(appID) == "" {
		return nil, ErrProviderSelectorRequired
	}
	query := tx
	if query == nil {
		query = s.deps.DB
	}
	if query == nil {
		return nil, ErrPaymentServiceUnavailable
	}
	var row pxmodels.PaymentProvider
	err := query.WithContext(ctx).
		Where(
			"tenant_uuid = ? AND provider_type = ? AND status = ? AND mch_id = ? AND app_id = ?",
			strings.TrimSpace(tenantUUID),
			strings.TrimSpace(providerType),
			"active",
			strings.TrimSpace(mchID),
			strings.TrimSpace(appID),
		).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProviderUnavailable
		}
		return nil, err
	}
	return &row, nil
}

func (s *TransactionService) findActiveTransaction(ctx context.Context, tx *gorm.DB, tenantUUID, orderID string) (*pxmodels.PaymentTransaction, error) {
	var row pxmodels.PaymentTransaction
	err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ? AND status IN ?", tenantUUID, orderID, []string{"pending_payment", "paying"}).
		Order("created_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *TransactionService) resolveProvider(ctx context.Context, tx *gorm.DB, tenantUUID, payMethod string, providerID uint64, providerType, mchID, appID string) (*pxmodels.PaymentProvider, error) {
	requestedType := strings.TrimSpace(providerType)
	if requestedType == "" {
		requestedType = mapProviderType(payMethod)
	}
	var row pxmodels.PaymentProvider
	query := tx.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if providerID > 0 {
		query = query.Where("id = ? AND status = ?", providerID, "active")
	} else {
		if strings.TrimSpace(requestedType) == "" {
			requestedType = "wechat"
		}
		if strings.TrimSpace(mchID) == "" || strings.TrimSpace(appID) == "" {
			return nil, ErrProviderSelectorRequired
		}
		query = query.Where(
			"provider_type = ? AND status = ? AND mch_id = ? AND app_id = ?",
			strings.TrimSpace(requestedType),
			"active",
			strings.TrimSpace(mchID),
			strings.TrimSpace(appID),
		)
	}
	err := query.First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProviderUnavailable
		}
		return nil, err
	}
	return &row, nil
}

func (s *TransactionService) loadProvider(ctx context.Context, tx *gorm.DB, providerID uint64) (*pxmodels.PaymentProvider, error) {
	if providerID == 0 {
		return nil, ErrProviderUnavailable
	}
	var row pxmodels.PaymentProvider
	query := tx
	if query == nil {
		query = s.deps.DB
	}
	if query == nil {
		return nil, ErrPaymentServiceUnavailable
	}
	if err := query.WithContext(ctx).Where("id = ?", providerID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *TransactionService) buildCreateResponse(ctx context.Context, txn *pxmodels.PaymentTransaction, provider *pxmodels.PaymentProvider, openID string) (*CreateTransactionResponse, error) {
	if txn == nil {
		return nil, ErrInvalidArgument
	}
	resp := &CreateTransactionResponse{
		TransactionID: strconv.FormatUint(txn.ID, 10),
		Status:        strings.TrimSpace(txn.Status),
	}
	if strings.EqualFold(strings.TrimSpace(txn.PayMethod), "wechat_jsapi") || strings.EqualFold(strings.TrimSpace(txn.PayMethod), "wechat") {
		params, err := s.buildWechatParams(ctx, provider, txn, openID)
		if err != nil {
			return nil, err
		}
		resp.Wechat = params
	}
	return resp, nil
}

func (s *TransactionService) applyStatusUpdate(ctx context.Context, tenantUUID string, row *pxmodels.PaymentTransaction, status, reason string) error {
	if row == nil {
		return ErrInvalidArgument
	}
	return s.deps.DB.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		updates := map[string]any{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}
		if reason = strings.TrimSpace(reason); reason != "" {
			updates["failure_reason"] = reason
		}
		if status == "paid" {
			now := time.Now().UTC()
			updates["completed_at"] = &now
			if err := db.WithContext(ctx).
				Model(&ordermodel.Order{}).
				Where("tenant_uuid = ? AND id = ?", tenantUUID, row.OrderID).
				Updates(map[string]any{"status": "paid", "updated_at": now}).Error; err != nil {
				return err
			}
		}
		if err := db.WithContext(ctx).
			Model(&pxmodels.PaymentTransaction{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, row.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		s.emitStatusEvent(ctx, tenantUUID, row, status, reason)
		return nil
	})
}

func newTransactionNo(orderNo string) string {
	// WeChat JSAPI out_trade_no must be <= 32 bytes.
	ts := time.Now().Format("060102150405")
	suffix := randomHex(4)
	if suffix == "" {
		suffix = fmt.Sprintf("%06d", time.Now().UnixNano()%1_000_000)
	}
	return fmt.Sprintf("TX%s%s", ts, suffix)
}

func mapProviderType(payMethod string) string {
	pm := strings.ToLower(strings.TrimSpace(payMethod))
	if strings.Contains(pm, "wechat") || strings.Contains(pm, "wx") {
		return "wechat"
	}
	return ""
}

func mapTradeState(state string) string {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "SUCCESS":
		return "paid"
	case "USERPAYING", "PROCESSING":
		return "paying"
	case "NOTPAY":
		return "pending_payment"
	case "CLOSED", "REVOKED":
		return "canceled"
	case "PAYERROR", "FAILED", "FAIL":
		return "failed"
	default:
		return ""
	}
}

func isTerminalStatus(status string) bool {
	s := strings.TrimSpace(strings.ToLower(status))
	return s == "paid" || s == "failed" || s == "canceled" || s == "timeout"
}

func pickString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			default:
				b, err := json.Marshal(t)
				if err == nil {
					value := strings.Trim(string(b), "\"")
					if strings.TrimSpace(value) != "" {
						return strings.TrimSpace(value)
					}
				}
			}
		}
	}
	return ""
}

func (s *TransactionService) buildWechatParams(ctx context.Context, provider *pxmodels.PaymentProvider, txn *pxmodels.PaymentTransaction, openID string) (*WechatPayParams, error) {
	if provider == nil || txn == nil {
		return nil, ErrProviderUnavailable
	}
	app, _, err := getWechatPaymentAppCachedForProvider(s.deps.Config, provider)
	if err != nil {
		return nil, err
	}

	prepayID := getMetadataString(txn.Metadata, "wechat_prepay_id", "prepay_id")
	if prepayID == "" {
		if openID == "" {
			openID = getMetadataString(txn.Metadata, "openid", "open_id", "openId")
		}
		if openID == "" {
			return nil, ErrInvalidArgument
		}
		prepayReq := &orderrequest.RequestJSAPIPrepay{
			Description: fmt.Sprintf("订单%s支付", strings.TrimSpace(txn.OrderNo)),
			OutTradeNo:  strings.TrimSpace(txn.TransactionNo),
			Amount: &orderrequest.JSAPIAmount{
				Total:    int(txn.AmountTotal),
				Currency: strings.TrimSpace(txn.AmountCurrency),
			},
			Payer: &orderrequest.JSAPIPayer{OpenID: strings.TrimSpace(openID)},
		}
		prepayResp, err := app.Order.JSAPITransaction(ctx, prepayReq)
		if err != nil {
			return nil, err
		}
		if prepayResp == nil {
			return nil, errors.New("wechat prepay response empty")
		}
		prepayID = strings.TrimSpace(prepayResp.PrepayID)
		if prepayID == "" {
			return nil, fmt.Errorf("wechat prepay id missing: %s", compactJSON(prepayResp))
		}
		if err := s.updateTransactionMetadata(ctx, txn, map[string]any{
			"wechat_prepay_id": prepayID,
		}); err != nil {
			return nil, err
		}
	}

	bridge, err := app.JSSDK.BridgeConfig(prepayID, true)
	if err != nil {
		return nil, err
	}
	params, err := bridgeToWechatParams(bridge)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func bridgeToWechatParams(bridge interface{}) (*WechatPayParams, error) {
	switch v := bridge.(type) {
	case *object.StringMap:
		return &WechatPayParams{
			AppID:     vValue(v, "appId"),
			TimeStamp: vValue(v, "timeStamp"),
			NonceStr:  vValue(v, "nonceStr"),
			Package:   vValue(v, "package"),
			SignType:  vValue(v, "signType"),
			PaySign:   vValue(v, "paySign"),
		}, nil
	case []byte:
		var raw map[string]string
		if err := json.Unmarshal(v, &raw); err != nil {
			return nil, err
		}
		return &WechatPayParams{
			AppID:     strings.TrimSpace(raw["appId"]),
			TimeStamp: strings.TrimSpace(raw["timeStamp"]),
			NonceStr:  strings.TrimSpace(raw["nonceStr"]),
			Package:   strings.TrimSpace(raw["package"]),
			SignType:  strings.TrimSpace(raw["signType"]),
			PaySign:   strings.TrimSpace(raw["paySign"]),
		}, nil
	default:
		return nil, errors.New("wechat bridge config invalid")
	}
}

func vValue(m *object.StringMap, key string) string {
	if m == nil {
		return ""
	}
	return strings.TrimSpace((*m)[key])
}

func getMetadataString(raw datatypes.JSON, keys ...string) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			default:
				s := strings.TrimSpace(fmt.Sprintf("%v", t))
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

func compactJSON(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func randomHex(n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", buf)
}

func (s *TransactionService) updateTransactionMetadata(ctx context.Context, txn *pxmodels.PaymentTransaction, updates map[string]any) error {
	if s == nil || s.deps == nil || s.deps.DB == nil || txn == nil {
		return ErrPaymentServiceUnavailable
	}
	if len(updates) == 0 {
		return nil
	}
	payload := map[string]any{}
	if len(txn.Metadata) > 0 {
		_ = json.Unmarshal(txn.Metadata, &payload)
	}
	for k, v := range updates {
		if strings.TrimSpace(k) != "" {
			payload[k] = v
		}
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := s.deps.DB.WithContext(ctx).
		Model(&pxmodels.PaymentTransaction{}).
		Where("tenant_uuid = ? AND id = ?", txn.TenantUuid, txn.ID).
		Update("metadata", datatypes.JSON(buf)).Error; err != nil {
		return err
	}
	txn.Metadata = datatypes.JSON(buf)
	return nil
}

func (s *TransactionService) markTransactionFailed(ctx context.Context, tenantUUID string, id uint64, err error) error {
	if s == nil || s.deps == nil || s.deps.DB == nil {
		return ErrPaymentServiceUnavailable
	}
	reason := strings.TrimSpace(err.Error())
	return s.deps.DB.WithContext(ctx).
		Model(&pxmodels.PaymentTransaction{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, id).
		Updates(map[string]any{
			"status":         "failed",
			"failure_reason": reason,
			"updated_at":     time.Now().UTC(),
		}).Error
}

func (s *TransactionService) handleWechatCallback(ctx context.Context, tenantUUID string, provider *pxmodels.PaymentProvider, req *http.Request) (*http.Response, error) {
	if provider == nil {
		return nil, ErrProviderUnavailable
	}
	if req != nil && req.Body != nil {
		raw, err := io.ReadAll(req.Body)
		if err == nil && len(raw) > 0 && s.deps != nil && s.deps.Config != nil && s.deps.Config.Logging != nil && s.deps.Config.Logging.PaymentCallbackNotifyDebug {
			logger.WithFields(logger.Fields{
				"tenant_uuid": tenantUUID,
				"provider_id": provider.ID,
				"request_id":  requestIDFromHTTPRequest(req),
			}).Infof("wechat pay notify raw body: %s", string(raw))
		}
		req.Body = io.NopCloser(bytes.NewBuffer(raw))
	}
	app, _, err := getWechatPaymentAppCachedForProvider(s.deps.Config, provider)
	if err != nil {
		return nil, err
	}

	return app.HandlePaidNotify(req, func(message *request.RequestNotify, transaction *wxmodels.Transaction, fail func(message string)) interface{} {
		if transaction == nil {
			reqID, _ := authx.RequestIDFromContext(ctx)
			logger.WithFields(logger.Fields{
				"tenant_uuid": tenantUUID,
				"provider_id": provider.ID,
				"request_id":  reqID,
			}).Warn("wechat pay notify missing transaction")
			fail("missing transaction")
			return "missing transaction"
		}
		reqID, _ := authx.RequestIDFromContext(ctx)
		logger.WithFields(logger.Fields{
			"tenant_uuid":    tenantUUID,
			"provider_id":    provider.ID,
			"trade_state":    safeString(transaction.TradeState),
			"out_trade_no":   safeString(transaction.OutTradeNo),
			"transaction_id": safeString(transaction.TransactionID),
			"request_id":     reqID,
		}).Info("wechat pay notify received")
		if err := s.applyWechatNotification(ctx, tenantUUID, provider, transaction); err != nil {
			fail(err.Error())
			return err.Error()
		}
		return true
	})
}

func safeString(v string) string {
	return strings.TrimSpace(v)
}

func requestIDFromHTTPRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	if v := strings.TrimSpace(req.Header.Get("X-Request-ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(req.Header.Get("Request-ID")); v != "" {
		return v
	}
	return ""
}

func (s *TransactionService) applyWechatNotification(ctx context.Context, tenantUUID string, provider *pxmodels.PaymentProvider, notice *wxmodels.Transaction) error {
	if notice == nil {
		return ErrInvalidArgument
	}
	outTradeNo := strings.TrimSpace(notice.OutTradeNo)
	if outTradeNo == "" {
		return ErrInvalidArgument
	}
	var row pxmodels.PaymentTransaction
	query := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND transaction_no = ?", tenantUUID, outTradeNo)
	if provider != nil && provider.ID > 0 {
		query = query.Where("provider_id = ?", provider.ID)
	}
	if err := query.First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTransactionNotFound
		}
		return err
	}
	updates := map[string]any{
		"wechat_trade_state":    strings.TrimSpace(notice.TradeState),
		"wechat_transaction_id": strings.TrimSpace(notice.TransactionID),
	}
	if notice.Payer != nil && strings.TrimSpace(notice.Payer.OpenID) != "" {
		updates["wechat_payer_openid"] = strings.TrimSpace(notice.Payer.OpenID)
	}
	if err := s.updateTransactionMetadata(ctx, &row, updates); err != nil {
		return err
	}
	if isTerminalStatus(row.Status) {
		return nil
	}
	status := mapTradeState(notice.TradeState)
	if status == "" {
		return nil
	}
	reason := strings.TrimSpace(notice.TradeStateDesc)
	return s.applyStatusUpdate(ctx, tenantUUID, &row, status, reason)
}

func (s *TransactionService) emitCallbackEvent(ctx context.Context, tenantUUID string, row *pxmodels.PaymentTransaction, providerID uint64, result, reason string) {
	logger := s.logger(ctx)
	if logger == nil || row == nil {
		return
	}
	logger.EmitEvent(paymentslogger.Event{
		Action:        "provider_callback",
		TenantID:      tenantUUID,
		RequestID:     requestIDFromContext(ctx),
		OrderID:       strings.TrimSpace(row.OrderID),
		OrderNo:       strings.TrimSpace(row.OrderNo),
		TransactionID: strconv.FormatUint(row.ID, 10),
		ProviderID:    strconv.FormatUint(providerID, 10),
		Amount:        row.AmountTotal,
		Currency:      strings.TrimSpace(row.AmountCurrency),
		Status:        strings.TrimSpace(row.Status),
		Result:        strings.TrimSpace(result),
		Reason:        strings.TrimSpace(reason),
		EmittedAt:     time.Now().UTC(),
	})
}

func (s *TransactionService) emitStatusEvent(ctx context.Context, tenantUUID string, row *pxmodels.PaymentTransaction, status, reason string) {
	logger := s.logger(ctx)
	if logger == nil || row == nil {
		return
	}
	logger.EmitEvent(paymentslogger.Event{
		Action:        "status_update",
		TenantID:      tenantUUID,
		RequestID:     requestIDFromContext(ctx),
		OrderID:       strings.TrimSpace(row.OrderID),
		OrderNo:       strings.TrimSpace(row.OrderNo),
		TransactionID: strconv.FormatUint(row.ID, 10),
		ProviderID:    strconv.FormatUint(row.ProviderID, 10),
		Amount:        row.AmountTotal,
		Currency:      strings.TrimSpace(row.AmountCurrency),
		Status:        strings.TrimSpace(status),
		Result:        "updated",
		Reason:        strings.TrimSpace(reason),
		EmittedAt:     time.Now().UTC(),
	})
}

func (s *TransactionService) logger(ctx context.Context) *paymentslogger.Logger {
	if s == nil || s.deps == nil {
		return nil
	}
	entry := s.deps.RuntimeLogger(ctx, "agent_payments", nil)
	if entry == nil {
		return nil
	}
	return paymentslogger.NewLogger(entry)
}

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value("request_id"); v != nil {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}
