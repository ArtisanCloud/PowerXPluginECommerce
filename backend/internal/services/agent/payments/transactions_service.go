package payments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	paymentslogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TransactionService struct {
	deps        *app.Deps
	transactions *paymentrepo.PaymentTransactionRepository
	providers   *paymentrepo.PaymentProviderRepository
	orders      *orderrepo.OrderRepository
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
	}
}

func (s *TransactionService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.transactions != nil && s.orders != nil
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tenantUUID string, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	if !s.Ready() {
		return nil, ErrPaymentServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, ErrInvalidArgument
	}
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.OrderNo = strings.TrimSpace(req.OrderNo)
	req.Currency = strings.TrimSpace(req.Currency)
	req.PayMethod = strings.TrimSpace(req.PayMethod)
	req.Client = strings.TrimSpace(req.Client)
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.OrderID == "" || req.OrderNo == "" || req.AmountMinor <= 0 || req.Currency == "" || req.PayMethod == "" || req.Client == "" || req.IdempotencyKey == "" {
		return nil, ErrInvalidArgument
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
	if strings.TrimSpace(orderRow.OrderNo) != "" && strings.TrimSpace(orderRow.OrderNo) != req.OrderNo {
		return nil, ErrInvalidArgument
	}
	if strings.TrimSpace(orderRow.Status) != "pending_payment" {
		return nil, ErrOrderNotPayable
	}

	if existing, err := s.findActiveTransaction(ctx, tx, tenantUUID, req.OrderID); err == nil && existing != nil {
		provider, _ := s.loadProvider(ctx, tx, existing.ProviderID)
		resp := s.buildCreateResponse(existing, provider)
		_ = tx.Commit().Error
		return resp, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	provider, err := s.resolveProvider(ctx, tx, tenantUUID, req.PayMethod)
	if err != nil {
		return nil, err
	}
	providerID := provider.ID
	providerType := strings.TrimSpace(provider.ProviderType)

	metadata := map[string]any{
		"client":          req.Client,
		"openid":          strings.TrimSpace(req.OpenID),
		"idempotency_key": req.IdempotencyKey,
		"provider_type":   providerType,
	}
	metadataRaw, _ := json.Marshal(metadata)
	transaction := &models.PaymentTransaction{
		BaseModel:      models.BaseModel{TenantUuid: tenantUUID},
		TransactionNo:  newTransactionNo(req.OrderNo),
		OrderID:        req.OrderID,
		OrderNo:        req.OrderNo,
		ProviderID:     providerID,
		PayMethod:      req.PayMethod,
		AmountTotal:    req.AmountMinor,
		AmountCurrency: req.Currency,
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

	resp := s.buildCreateResponse(transaction, provider)
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

	var row models.PaymentTransaction
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

	var row models.PaymentTransaction
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

func (s *TransactionService) findActiveTransaction(ctx context.Context, tx *gorm.DB, tenantUUID, orderID string) (*models.PaymentTransaction, error) {
	var row models.PaymentTransaction
	err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ? AND status IN ?", tenantUUID, orderID, []string{"pending_payment", "paying"}).
		Order("created_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *TransactionService) resolveProvider(ctx context.Context, tx *gorm.DB, tenantUUID, payMethod string) (*models.PaymentProvider, error) {
	providerType := mapProviderType(payMethod)
	if providerType == "" {
		providerType = "wechat"
	}
	var row models.PaymentProvider
	err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND provider_type = ? AND status = ?", tenantUUID, providerType, "active").
		Order("id DESC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProviderUnavailable
		}
		return nil, err
	}
	return &row, nil
}

func (s *TransactionService) loadProvider(ctx context.Context, tx *gorm.DB, providerID uint64) (*models.PaymentProvider, error) {
	if providerID == 0 {
		return nil, ErrProviderUnavailable
	}
	var row models.PaymentProvider
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

func (s *TransactionService) buildCreateResponse(txn *models.PaymentTransaction, provider *models.PaymentProvider) *CreateTransactionResponse {
	if txn == nil {
		return nil
	}
	resp := &CreateTransactionResponse{
		TransactionID: strconv.FormatUint(txn.ID, 10),
		Status:        strings.TrimSpace(txn.Status),
	}
	if strings.EqualFold(strings.TrimSpace(txn.PayMethod), "wechat_jsapi") || strings.EqualFold(strings.TrimSpace(txn.PayMethod), "wechat") {
		resp.Wechat = buildWechatParams(provider, txn.TransactionNo)
	}
	return resp
}

func (s *TransactionService) applyStatusUpdate(ctx context.Context, tenantUUID string, row *models.PaymentTransaction, status, reason string) error {
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
			Model(&models.PaymentTransaction{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, row.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		s.emitStatusEvent(ctx, tenantUUID, row, status, reason)
		return nil
	})
}

func buildWechatParams(provider *models.PaymentProvider, transactionNo string) *WechatPayParams {
	if provider == nil || len(provider.Credentials) == 0 {
		return nil
	}
	appID := extractAppID(provider.Credentials)
	if appID == "" {
		return nil
	}
	nonce := randomHex(16)
	stamp := strconv.FormatInt(time.Now().Unix(), 10)
	return &WechatPayParams{
		AppID:     appID,
		TimeStamp: stamp,
		NonceStr:  nonce,
		Package:   fmt.Sprintf("prepay_id=%s", strings.TrimSpace(transactionNo)),
		SignType:  "RSA",
		PaySign:   randomHex(32),
	}
}

func randomHex(size int) string {
	if size <= 0 {
		size = 16
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func newTransactionNo(orderNo string) string {
	trimmed := strings.TrimSpace(orderNo)
	trimmed = strings.ReplaceAll(trimmed, "-", "")
	if len(trimmed) > 20 {
		trimmed = trimmed[len(trimmed)-20:]
	}
	return fmt.Sprintf("TX-%s-%d", trimmed, time.Now().Unix())
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

func extractAppID(raw datatypes.JSON) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	for _, key := range []string{"appId", "appid", "app_id"} {
		if v, ok := payload[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
			if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" {
				return s
			}
		}
	}
	return ""
}

func (s *TransactionService) emitCallbackEvent(ctx context.Context, tenantUUID string, row *models.PaymentTransaction, providerID uint64, result, reason string) {
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

func (s *TransactionService) emitStatusEvent(ctx context.Context, tenantUUID string, row *models.PaymentTransaction, status, reason string) {
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
