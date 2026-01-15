package order

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	integrationmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/integration"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	customerrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/customer"
	idrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/integration"
	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	skurepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_sku"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	sellabilitysvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/sellability"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	defaultWarehouseID = "default"
	idempotencyScope   = "order"
	idempotencyOp      = "create"
)

var (
	ErrOrderServiceUnavailable = errors.New("order service unavailable")
	ErrIdempotencyKeyRequired  = errors.New("idempotency key is required")
	ErrIdempotencyConflict     = errors.New("idempotency key conflict")
	ErrIdempotencyInProgress   = errors.New("idempotency request in progress")
	ErrCustomerRequired        = errors.New("customer id is required")
	ErrChannelRequired         = errors.New("channel is required")
	ErrItemsRequired           = errors.New("items are required")
	ErrInvalidQty              = errors.New("qty must be positive")
	ErrDuplicateSKU            = errors.New("duplicate sku in request")
	ErrShippingAddressRequired = errors.New("shipping address is required")
	ErrShippingAddressNotFound = errors.New("shipping address not found")
	ErrInvalidShippingAddress  = errors.New("invalid shipping address")
	ErrOrderNotFound           = errors.New("order not found")
	ErrSellabilityFailed       = errors.New("sku not sellable")
	ErrOutOfStock              = errors.New("out of stock")
)

type Service struct {
	deps            *app.Deps
	idempotencyTTL  time.Duration
	OrderRepo       *orderrepo.OrderRepository
	ItemRepo        *orderrepo.OrderItemRepository
	EventRepo       *orderrepo.OrderEventRepository
	InventoryRepo   *skurepo.InventoryRepository
	IdempotencyRepo *idrepo.IdempotencyRepository
	SellabilitySvc  *sellabilitysvc.Service
	AddressRepo     *customerrepo.CustomerAddressRepository
}

func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{deps: deps}
	}
	ttl := 24 * time.Hour
	if deps.Config != nil {
		ttl = deps.Config.IntegrationIdempotencyTTL()
	}
	fallback := idrepo.NewPostgresIdempotencyProvider(deps.DB, ttl)
	repository := idrepo.NewIdempotencyRepository(deps.DB, nil, fallback, nil)
	return &Service{
		deps:            deps,
		idempotencyTTL:  ttl,
		OrderRepo:       orderrepo.NewOrderRepository(deps.DB),
		ItemRepo:        orderrepo.NewOrderItemRepository(deps.DB),
		EventRepo:       orderrepo.NewOrderEventRepository(deps.DB),
		InventoryRepo:   skurepo.NewInventoryRepository(deps.DB),
		IdempotencyRepo: repository,
		SellabilitySvc:  sellabilitysvc.NewService(deps.DB),
		AddressRepo:     customerrepo.NewCustomerAddressRepository(deps.DB),
	}
}

func (s *Service) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.OrderRepo != nil && s.InventoryRepo != nil && s.IdempotencyRepo != nil && s.SellabilitySvc != nil && s.AddressRepo != nil
}

func (s *Service) CreateOrder(ctx context.Context, tenantUUID, customerID, idempotencyKey string, req CreateOrderRequest) (*OrderSummaryDTO, error) {
	if !s.Ready() {
		return nil, ErrOrderServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	if idempotencyKey == "" {
		return nil, ErrIdempotencyKeyRequired
	}
	req.Channel = strings.TrimSpace(req.Channel)
	if req.Channel == "" {
		return nil, ErrChannelRequired
	}
	if len(req.Items) == 0 {
		return nil, ErrItemsRequired
	}

	skuIDs := make([]string, 0, len(req.Items))
	seen := map[string]struct{}{}
	for _, it := range req.Items {
		id := strings.TrimSpace(it.SKUID)
		if id == "" {
			return nil, errors.New("skuId is required")
		}
		if it.Qty <= 0 {
			return nil, ErrInvalidQty
		}
		if _, ok := seen[id]; ok {
			return nil, ErrDuplicateSKU
		}
		seen[id] = struct{}{}
		skuIDs = append(skuIDs, id)
	}

	payloadHash, err := hashCreatePayload(customerID, req)
	if err != nil {
		return nil, err
	}

	claim, err := s.IdempotencyRepo.Claim(ctx, &integrationmodel.IdempotencyRecord{
		Key:         idempotencyKey,
		TenantUuid:  tenantUUID,
		Scope:       idempotencyScope,
		Operation:   idempotencyOp,
		PayloadHash: payloadHash,
	})
	if err != nil {
		return nil, err
	}
	if claim == nil || claim.Record == nil {
		return nil, errors.New("idempotency claim unavailable")
	}

	if claim.Status == idrepo.ClaimStatusExisting {
		if strings.TrimSpace(claim.Record.PayloadHash) != "" && claim.Record.PayloadHash != payloadHash {
			return nil, ErrIdempotencyConflict
		}
		if len(claim.Record.Response) > 0 {
			var cached OrderSummaryDTO
			if err := json.Unmarshal([]byte(claim.Record.Response), &cached); err != nil {
				return nil, err
			}
			return &cached, nil
		}
		return nil, ErrIdempotencyInProgress
	}

	created, err := s.createOrderWithIdempotencyTx(ctx, tenantUUID, customerID, idempotencyKey, skuIDs, req)
	if err != nil {
		_ = s.IdempotencyRepo.Delete(ctx, idempotencyKey)
		return nil, err
	}
	return created, nil
}

func (s *Service) createOrderWithIdempotencyTx(
	ctx context.Context,
	tenantUUID, customerID, idempotencyKey string,
	skuIDs []string,
	req CreateOrderRequest,
) (*OrderSummaryDTO, error) {
	shippingAddrID := strings.TrimSpace(req.ShippingAddressID)
	var shippingSnap *ShippingAddress
	switch {
	case shippingAddrID != "":
		addr, err := s.AddressRepo.GetByID(ctx, tenantUUID, customerID, shippingAddrID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrShippingAddressNotFound
			}
			return nil, err
		}
		shippingSnap = shippingSnapshotFromAddress(addr)
	case req.ShippingAddress != nil:
		if !isValidShippingAddress(req.ShippingAddress) {
			return nil, ErrInvalidShippingAddress
		}
		shippingSnap = req.ShippingAddress
	default:
		return nil, ErrShippingAddressRequired
	}
	shippingSnapJSON, _ := json.Marshal(shippingSnap)

	sellability, err := s.SellabilitySvc.EvaluateSKUs(ctx, tenantUUID, skuIDs, req.Channel, req.Locale)
	if err != nil {
		return nil, err
	}
	itemBySKU := map[string]sellabilitysvc.ItemDTO{}
	for _, it := range sellability.Items {
		itemBySKU[strings.TrimSpace(it.SKUID)] = it
	}

	currency := ""
	unitPriceMinor := map[string]int64{}
	for _, it := range req.Items {
		skuID := strings.TrimSpace(it.SKUID)
		info, ok := itemBySKU[skuID]
		if !ok {
			return nil, fmt.Errorf("sku not found: %s", skuID)
		}
		if !info.Sellable {
			return nil, fmt.Errorf("%w: %s", ErrSellabilityFailed, strings.Join(info.Reasons, ","))
		}
		if info.AvailableQty < int(it.Qty) {
			return nil, ErrOutOfStock
		}
		if info.Price == nil || info.Price.Amount <= 0 || strings.TrimSpace(info.Price.Currency) == "" {
			return nil, ErrSellabilityFailed
		}
		if currency == "" {
			currency = strings.TrimSpace(info.Price.Currency)
		} else if currency != strings.TrimSpace(info.Price.Currency) {
			return nil, errors.New("mixed currency is not supported")
		}
		unitPriceMinor[skuID] = int64(math.Round(info.Price.Amount * 100))
	}

	subtotal := int64(0)
	for _, it := range req.Items {
		price := unitPriceMinor[strings.TrimSpace(it.SKUID)]
		subtotal += price * it.Qty
	}
	total := subtotal

	now := time.Now().UTC()
	orderID := uuid.NewString()
	orderNo := generateOrderNo(now)

	priceSnapshot := map[string]any{
		"currency": currency,
		"subtotal": subtotal,
		"total":    total,
		"items":    req.Items,
		"pricedAt": now.Format(time.RFC3339Nano),
	}
	priceSnapJSON, _ := json.Marshal(priceSnapshot)

	sellSnapJSON, _ := json.Marshal(sellability)

	var summary *OrderSummaryDTO
	err = s.OrderRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		// Re-check and lock inventory within the same transaction (strong consistency).
		for _, it := range req.Items {
			skuID := strings.TrimSpace(it.SKUID)
			row, err := s.InventoryRepo.EnsureWarehouseRowForUpdate(ctx, tx, tenantUUID, skuID, defaultWarehouseID)
			if err != nil {
				return err
			}
			available := row.AvailableQty - row.LockedQty
			if available < it.Qty {
				return ErrOutOfStock
			}
			after := row.LockedQty + it.Qty
			if err := tx.WithContext(ctx).
				Model(&productskumodel.ProductSKUInventory{}).
				Where("id = ?", row.ID).
				Update("locked_qty", after).Error; err != nil {
				return err
			}
		}

		order := &ordermodel.Order{
			ID:                  orderID,
			TenantUUID:          tenantUUID,
			OrderNo:             orderNo,
			CustomerID:          customerID,
			Channel:             req.Channel,
			Status:              "pending_payment",
			Currency:            currency,
			SubtotalAmount:      subtotal,
			TotalAmount:         total,
			ShippingAddressID:   shippingAddrID,
			ShippingAddressSnap: datatypes.JSON(shippingSnapJSON),
			PriceSnapshot:       datatypes.JSON(priceSnapJSON),
			SellabilitySnap:     datatypes.JSON(sellSnapJSON),
			CreatedByType:       "customer",
			CreatedBy:           customerID,
		}
		if err := s.OrderRepo.CreateWithTx(ctx, tx, order); err != nil {
			return err
		}

		items := make([]*ordermodel.OrderItem, 0, len(req.Items))
		for _, it := range req.Items {
			skuID := strings.TrimSpace(it.SKUID)
			price := unitPriceMinor[skuID]
			items = append(items, &ordermodel.OrderItem{
				ID:         uuid.NewString(),
				TenantUUID: tenantUUID,
				OrderID:    orderID,
				SKUID:      skuID,
				Qty:        it.Qty,
				UnitPrice:  price,
				LineAmount: price * it.Qty,
			})
		}
		if err := s.ItemRepo.CreateBatchWithTx(ctx, tx, items); err != nil {
			return err
		}

		eventPayload, _ := json.Marshal(map[string]any{
			"requestId":         requestIDFromContext(ctx),
			"idempotencyKey":    idempotencyKey,
			"channel":           req.Channel,
			"shippingAddressId": shippingAddrID,
		})
		event := &ordermodel.OrderEvent{
			ID:           uuid.NewString(),
			TenantUUID:   tenantUUID,
			OrderID:      orderID,
			EventType:    "order.created",
			OperatorType: "customer",
			Operator:     customerID,
			Payload:      datatypes.JSON(eventPayload),
		}
		if err := s.EventRepo.CreateWithTx(ctx, tx, event); err != nil {
			return err
		}

		summary = &OrderSummaryDTO{
			OrderID:                 orderID,
			OrderNo:                 orderNo,
			Status:                  order.Status,
			Amounts:                 MoneyDTO{Currency: currency, Subtotal: subtotal, Total: total},
			ShippingAddressSnapshot: shippingSnap,
			// CreatedAt is not the DB-created timestamp, but deterministic enough for API response.
			CreatedAt: now,
		}

		respRaw, _ := json.Marshal(summary)
		expiresAt := time.Now().UTC().Add(s.idempotencyTTL)
		updates := map[string]any{
			"response_data": datatypes.JSON(respRaw),
			"expires_at":    expiresAt,
		}
		return tx.WithContext(ctx).
			Model(&integrationmodel.IdempotencyRecord{}).
			Where("key = ?", idempotencyKey).
			Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func hashCreatePayload(customerID string, req CreateOrderRequest) (string, error) {
	payload := map[string]any{
		"customerId":        customerID,
		"channel":           strings.TrimSpace(req.Channel),
		"shippingAddressId": strings.TrimSpace(req.ShippingAddressID),
		"shippingAddress":   req.ShippingAddress,
		"locale":            strings.TrimSpace(req.Locale),
		"items":             req.Items,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func isValidShippingAddress(addr *ShippingAddress) bool {
	if addr == nil {
		return false
	}
	if strings.TrimSpace(addr.RecipientName) == "" {
		return false
	}
	if strings.TrimSpace(addr.RecipientPhone) == "" {
		return false
	}
	if strings.TrimSpace(addr.Address1) == "" {
		return false
	}
	return true
}

func shippingSnapshotFromAddress(addr *customermodel.CustomerAddress) *ShippingAddress {
	if addr == nil {
		return nil
	}
	var metadata map[string]any
	if len(addr.Metadata) > 0 {
		_ = json.Unmarshal([]byte(addr.Metadata), &metadata)
	}
	return &ShippingAddress{
		Label:          strings.TrimSpace(addr.Label),
		RecipientName:  strings.TrimSpace(addr.RecipientName),
		RecipientPhone: strings.TrimSpace(addr.RecipientPhone),
		CountryCode:    strings.TrimSpace(addr.CountryCode),
		Province:       strings.TrimSpace(addr.Province),
		City:           strings.TrimSpace(addr.City),
		District:       strings.TrimSpace(addr.District),
		Address1:       strings.TrimSpace(addr.Address1),
		Address2:       strings.TrimSpace(addr.Address2),
		PostalCode:     strings.TrimSpace(addr.PostalCode),
		Metadata:       metadata,
	}
}

func generateOrderNo(now time.Time) string {
	t := now.UTC().Format("20060102150405")
	suffix := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))[:8]
	return fmt.Sprintf("O%s%s", t, suffix)
}

func requestIDFromContext(ctx context.Context) string {
	if v, ok := authx.RequestIDFromContext(ctx); ok {
		return v
	}
	return ""
}
