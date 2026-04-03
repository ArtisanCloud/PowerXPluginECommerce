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

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	integrationmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/integration"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	customerrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/customer"
	idrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/integration"
	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	skurepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_sku"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
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
	ErrCustomerNotFound        = errors.New("customer not found")
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderNotCancellable     = errors.New("order cannot be cancelled")
	ErrOrderNotEditable        = errors.New("order cannot be edited")
	ErrAdminRequired           = errors.New("admin id is required")
	ErrChannelRequired         = errors.New("channel is required")
	ErrItemsRequired           = errors.New("items are required")
	ErrInvalidQty              = errors.New("qty must be positive")
	ErrDuplicateSKU            = errors.New("duplicate sku in request")
	ErrShippingAddressRequired = errors.New("shipping address is required")
	ErrShippingAddressNotFound = errors.New("shipping address not found")
	ErrInvalidShippingAddress  = errors.New("invalid shipping address")
	ErrSellabilityFailed       = errors.New("sku not sellable")
	ErrOutOfStock              = errors.New("out of stock")
)

type Service struct {
	deps               *app.Deps
	idempotencyTTL     time.Duration
	CustomerRepo       *customerrepo.Repository
	OrderRepo          *orderrepo.OrderRepository
	ItemRepo           *orderrepo.OrderItemRepository
	EventRepo          *orderrepo.OrderEventRepository
	InventoryRepo      *skurepo.InventoryRepository
	IdempotencyRepo    *idrepo.IdempotencyRepository
	SellabilitySvc     *sellabilitysvc.Service
	AddressRepo        *customerrepo.CustomerAddressRepository
	CouponSnapshotRepo *couponrepo.OrderSnapshotRepository
	CouponQuoteSvc     *couponsvc.QuoteService
	CouponReserveSvc   *couponsvc.ReservationService
	CouponReleaseSvc   *couponsvc.ReleaseService
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
		deps:               deps,
		idempotencyTTL:     ttl,
		CustomerRepo:       customerrepo.NewRepository(deps.DB),
		OrderRepo:          orderrepo.NewOrderRepository(deps.DB),
		ItemRepo:           orderrepo.NewOrderItemRepository(deps.DB),
		EventRepo:          orderrepo.NewOrderEventRepository(deps.DB),
		InventoryRepo:      skurepo.NewInventoryRepository(deps.DB),
		IdempotencyRepo:    repository,
		SellabilitySvc:     sellabilitysvc.NewService(deps.DB),
		AddressRepo:        customerrepo.NewCustomerAddressRepository(deps.DB),
		CouponSnapshotRepo: couponrepo.NewOrderSnapshotRepository(deps.DB),
		CouponQuoteSvc:     couponsvc.NewQuoteService(deps),
		CouponReserveSvc:   couponsvc.NewReservationService(deps),
		CouponReleaseSvc:   couponsvc.NewReleaseService(deps),
	}
}

func (s *Service) Ready() bool {
	return s != nil &&
		s.deps != nil &&
		s.deps.DB != nil &&
		s.CustomerRepo != nil &&
		s.OrderRepo != nil &&
		s.InventoryRepo != nil &&
		s.IdempotencyRepo != nil &&
		s.SellabilitySvc != nil &&
		s.AddressRepo != nil &&
		s.CouponSnapshotRepo != nil &&
		s.CouponQuoteSvc != nil &&
		s.CouponReserveSvc != nil &&
		s.CouponReleaseSvc != nil
}

func (s *Service) CreateOrder(ctx context.Context, tenantUUID, adminID, idempotencyKey string, req CreateOrderRequest) (*OrderSummaryDTO, error) {
	if !s.Ready() {
		return nil, ErrOrderServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	adminID = strings.TrimSpace(adminID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)

	req.CustomerID = strings.TrimSpace(req.CustomerID)
	req.Channel = strings.TrimSpace(req.Channel)
	req.Note = strings.TrimSpace(req.Note)

	if adminID == "" {
		return nil, ErrAdminRequired
	}
	if req.CustomerID == "" {
		return nil, ErrCustomerRequired
	}
	if idempotencyKey == "" {
		return nil, ErrIdempotencyKeyRequired
	}
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

	customer, err := s.ensureCustomer(ctx, tenantUUID, req.CustomerID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}

	payloadHash, err := hashCreatePayload(req)
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

	created, err := s.createOrderWithIdempotencyTx(ctx, tenantUUID, adminID, idempotencyKey, skuIDs, req)
	if err != nil {
		_ = s.IdempotencyRepo.Delete(ctx, idempotencyKey)
		return nil, err
	}
	return created, nil
}

func (s *Service) ensureCustomer(ctx context.Context, tenantUUID, customerID string) (*customermodel.Customer, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, ErrCustomerRequired
	}
	ctx = authx.ContextWithTenantUUID(ctx, tenantUUID)
	return s.CustomerRepo.FindByCustomerID(ctx, customerID)
}

func (s *Service) createOrderWithIdempotencyTx(
	ctx context.Context,
	tenantUUID, adminID, idempotencyKey string,
	skuIDs []string,
	req CreateOrderRequest,
) (*OrderSummaryDTO, error) {
	shippingAddrID := strings.TrimSpace(req.ShippingAddressID)
	var shippingSnap *ShippingAddress
	switch {
	case shippingAddrID != "":
		addr, err := s.AddressRepo.GetByID(ctx, tenantUUID, req.CustomerID, shippingAddrID)
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

	sellability, err := s.SellabilitySvc.EvaluateSKUs(ctx, tenantUUID, skuIDs, req.Channel, "")
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
	var quoteResult *couponsvc.QuoteResult
	if len(req.CouponIDs) > 0 {
		quoteItems := make([]couponsvc.QuoteItemInput, 0, len(req.Items))
		for _, it := range req.Items {
			skuID := strings.TrimSpace(it.SKUID)
			quoteItems = append(quoteItems, couponsvc.QuoteItemInput{
				LineID:         skuID,
				SKUID:          skuID,
				Qty:            it.Qty,
				UnitPriceMinor: unitPriceMinor[skuID],
			})
		}
		quoted, err := s.CouponQuoteSvc.Quote(ctx, couponsvc.QuoteInput{
			TenantUUID: tenantUUID,
			UserID:     req.CustomerID,
			Channel:    req.Channel,
			CouponIDs:  req.CouponIDs,
			Items:      quoteItems,
			Currency:   currency,
		})
		if err != nil {
			return nil, err
		}
		quoteResult = quoted
		total = quoted.PayableTotalMinor
	}

	now := time.Now().UTC()
	orderID := uuid.NewString()
	orderNo := generateOrderNo(now)

	priceSnapshot := map[string]any{
		"currency": currency,
		"subtotal": subtotal,
		"total":    total,
		"items":    req.Items,
		"note":     req.Note,
		"pricedAt": now.Format(time.RFC3339Nano),
	}
	if quoteResult != nil {
		priceSnapshot["coupon"] = quoteResult
	}
	priceSnapJSON, _ := json.Marshal(priceSnapshot)
	sellSnapJSON, _ := json.Marshal(sellability)

	var summary *OrderSummaryDTO
	err = s.OrderRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
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

		var shippingAddrPtr *string
		if shippingAddrID != "" {
			shippingAddrPtr = &shippingAddrID
		}
		order := &ordermodel.Order{
			ID:                  orderID,
			TenantUUID:          tenantUUID,
			OrderNo:             orderNo,
			CustomerID:          req.CustomerID,
			Channel:             req.Channel,
			Status:              "pending_payment",
			Currency:            currency,
			SubtotalAmount:      subtotal,
			TotalAmount:         total,
			ShippingAddressID:   shippingAddrPtr,
			ShippingAddressSnap: datatypes.JSON(shippingSnapJSON),
			PriceSnapshot:       datatypes.JSON(priceSnapJSON),
			SellabilitySnap:     datatypes.JSON(sellSnapJSON),
			CreatedByType:       "admin",
			CreatedBy:           adminID,
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

		var reservedCouponIDs []string
		if quoteResult != nil {
			appliedAssetIDs := make([]string, 0, len(quoteResult.AppliedCoupons))
			for _, applied := range quoteResult.AppliedCoupons {
				appliedAssetIDs = append(appliedAssetIDs, applied.AssetID)
			}
			reserved, err := s.CouponReserveSvc.ReserveWithTx(ctx, tx, couponsvc.ReservationInput{
				TenantUUID: tenantUUID,
				OrderID:    orderID,
				AssetIDs:   appliedAssetIDs,
				RequestID:  requestIDFromContext(ctx),
				CreatedBy:  adminID,
			})
			if err != nil {
				return err
			}
			if reserved != nil {
				reservedCouponIDs = append(reservedCouponIDs, reserved.ReservedAssetIDs...)
			}
			lineAllocRaw, _ := json.Marshal(quoteResult.LineAllocations)
			appliedRaw, _ := json.Marshal(quoteResult.AppliedCoupons)
			rejectedRaw, _ := json.Marshal(quoteResult.RejectedCoupons)
			if err := s.CouponSnapshotRepo.UpsertWithTx(ctx, tx, &couponmodel.OrderCouponSnapshot{
				ID:                 uuid.NewString(),
				TenantUUID:         tenantUUID,
				OrderID:            orderID,
				Currency:           quoteResult.Currency,
				BaseTotalMinor:     quoteResult.BaseTotalMinor,
				DiscountTotalMinor: quoteResult.DiscountTotalMinor,
				PayableTotalMinor:  quoteResult.PayableTotalMinor,
				LineAllocations:    datatypes.JSON(lineAllocRaw),
				AppliedCoupons:     datatypes.JSON(appliedRaw),
				RejectedCoupons:    datatypes.JSON(rejectedRaw),
				PricedAt:           quoteResult.PricedAt,
			}); err != nil {
				return err
			}
		}

		eventPayload, _ := json.Marshal(map[string]any{
			"requestId":         requestIDFromContext(ctx),
			"idempotencyKey":    idempotencyKey,
			"channel":           req.Channel,
			"customerId":        req.CustomerID,
			"shippingAddressId": shippingAddrID,
			"note":              req.Note,
			"couponIds":         req.CouponIDs,
		})
		event := &ordermodel.OrderEvent{
			ID:           uuid.NewString(),
			TenantUUID:   tenantUUID,
			OrderID:      orderID,
			EventType:    "order.created",
			OperatorType: "admin",
			Operator:     adminID,
			Payload:      datatypes.JSON(eventPayload),
		}
		if err := s.EventRepo.CreateWithTx(ctx, tx, event); err != nil {
			return err
		}
		if quoteResult != nil {
			couponEventPayload, _ := json.Marshal(map[string]any{
				"requestId":        requestIDFromContext(ctx),
				"couponQuote":      quoteResult,
				"reservedAssetIDs": reservedCouponIDs,
			})
			couponEvent := &ordermodel.OrderEvent{
				ID:           uuid.NewString(),
				TenantUUID:   tenantUUID,
				OrderID:      orderID,
				EventType:    "order.coupon.applied",
				OperatorType: "admin",
				Operator:     adminID,
				Payload:      datatypes.JSON(couponEventPayload),
			}
			if err := s.EventRepo.CreateWithTx(ctx, tx, couponEvent); err != nil {
				return err
			}
		}

		summary = &OrderSummaryDTO{
			OrderID:                 orderID,
			OrderNo:                 orderNo,
			CustomerID:              order.CustomerID,
			Channel:                 order.Channel,
			CreatedByType:           order.CreatedByType,
			Status:                  order.Status,
			Amounts:                 MoneyDTO{Currency: currency, Subtotal: subtotal, Total: total},
			Coupon:                  toCouponSummaryDTO(quoteResult),
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

func requestIDFromContext(ctx context.Context) string {
	if v, ok := authx.RequestIDFromContext(ctx); ok {
		return v
	}
	return ""
}

func hashCreatePayload(req CreateOrderRequest) (string, error) {
	payload := map[string]any{
		"customerId":        strings.TrimSpace(req.CustomerID),
		"channel":           strings.TrimSpace(req.Channel),
		"couponIds":         req.CouponIDs,
		"shippingAddressId": strings.TrimSpace(req.ShippingAddressID),
		"shippingAddress":   req.ShippingAddress,
		"items":             req.Items,
		"note":              strings.TrimSpace(req.Note),
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
