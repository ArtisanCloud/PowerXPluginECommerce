package order

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	"gorm.io/gorm"
)

func (s *Service) ListOrders(ctx context.Context, tenantUUID string, filter orderrepo.OrderListFilter, page, pageSize int) (*OrderListResponse, error) {
	if !s.Ready() {
		return nil, ErrOrderServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)

	result, err := s.OrderRepo.List(ctx, tenantUUID, filter, page, pageSize)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &OrderListResponse{Items: []OrderSummaryDTO{}, Page: page, PageSize: pageSize, Total: 0}, nil
	}

	items := make([]OrderSummaryDTO, 0, len(result.Items))
	for _, it := range result.Items {
		var shippingSnap *ShippingAddress
		if len(it.ShippingAddressSnap) > 0 {
			var snap ShippingAddress
			if err := json.Unmarshal([]byte(it.ShippingAddressSnap), &snap); err == nil {
				shippingSnap = &snap
			}
		}
		items = append(items, OrderSummaryDTO{
			OrderID:                 it.ID,
			OrderNo:                 it.OrderNo,
			CustomerID:              it.CustomerID,
			Channel:                 it.Channel,
			CreatedByType:           it.CreatedByType,
			Status:                  it.Status,
			Amounts:                 MoneyDTO{Currency: it.Currency, Subtotal: it.SubtotalAmount, Total: it.TotalAmount},
			ShippingAddressSnapshot: shippingSnap,
			CreatedAt:               it.CreatedAt,
		})
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	return &OrderListResponse{Items: items, Page: page, PageSize: pageSize, Total: result.Total}, nil
}

func (s *Service) GetOrderDetail(ctx context.Context, tenantUUID, orderID string) (*OrderDetailDTO, error) {
	if !s.Ready() {
		return nil, ErrOrderServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, errors.New("order id is required")
	}

	ord, err := s.OrderRepo.GetByID(ctx, tenantUUID, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	orderItems, err := s.ItemRepo.ListByOrderID(ctx, tenantUUID, orderID)
	if err != nil {
		return nil, err
	}
	events, err := s.EventRepo.ListByOrderID(ctx, tenantUUID, orderID, 200)
	if err != nil {
		return nil, err
	}

	items := make([]OrderItemDTO, 0, len(orderItems))
	for _, it := range orderItems {
		items = append(items, OrderItemDTO{
			SKUID:      it.SKUID,
			Qty:        it.Qty,
			UnitPrice:  it.UnitPrice,
			LineAmount: it.LineAmount,
		})
	}
	if len(items) == 0 {
		items = buildFallbackItems(ord.PriceSnapshot, ord.SellabilitySnap)
	}

	evs := make([]OrderEventDTO, 0, len(events))
	for _, ev := range events {
		evs = append(evs, OrderEventDTO{
			EventType:    ev.EventType,
			OperatorType: ev.OperatorType,
			Operator:     ev.Operator,
			CreatedAt:    ev.CreatedAt,
		})
	}

	var shippingSnap *ShippingAddress
	if len(ord.ShippingAddressSnap) > 0 {
		var snap ShippingAddress
		if err := json.Unmarshal([]byte(ord.ShippingAddressSnap), &snap); err == nil {
			shippingSnap = &snap
		}
	}

	return &OrderDetailDTO{
		Summary: OrderSummaryDTO{
			OrderID:                 ord.ID,
			OrderNo:                 ord.OrderNo,
			CustomerID:              ord.CustomerID,
			Channel:                 ord.Channel,
			CreatedByType:           ord.CreatedByType,
			Status:                  ord.Status,
			Amounts:                 MoneyDTO{Currency: ord.Currency, Subtotal: ord.SubtotalAmount, Total: ord.TotalAmount},
			ShippingAddressSnapshot: shippingSnap,
			CreatedAt:               ord.CreatedAt,
		},
		Items:  items,
		Events: evs,
	}, nil
}

type priceSnapshotItem struct {
	SKUID string `json:"skuId"`
	Qty   int64  `json:"qty"`
}

type priceSnapshotPayload struct {
	Items []priceSnapshotItem `json:"items"`
}

type sellabilityItem struct {
	SKUID string `json:"skuId"`
	Price *struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	} `json:"price"`
}

type sellabilitySnapshotPayload struct {
	Items []sellabilityItem `json:"items"`
}

func buildFallbackItems(priceSnapRaw, sellSnapRaw []byte) []OrderItemDTO {
	if len(priceSnapRaw) == 0 {
		return nil
	}
	var priceSnap priceSnapshotPayload
	if err := json.Unmarshal(priceSnapRaw, &priceSnap); err != nil {
		return nil
	}
	if len(priceSnap.Items) == 0 {
		return nil
	}
	priceMap := map[string]int64{}
	if len(sellSnapRaw) > 0 {
		var sellSnap sellabilitySnapshotPayload
		if err := json.Unmarshal(sellSnapRaw, &sellSnap); err == nil {
			for _, it := range sellSnap.Items {
				if it.Price == nil || it.Price.Amount <= 0 {
					continue
				}
				priceMap[strings.TrimSpace(it.SKUID)] = int64(it.Price.Amount * 100)
			}
		}
	}
	items := make([]OrderItemDTO, 0, len(priceSnap.Items))
	for _, it := range priceSnap.Items {
		skuID := strings.TrimSpace(it.SKUID)
		if skuID == "" || it.Qty <= 0 {
			continue
		}
		unit := priceMap[skuID]
		items = append(items, OrderItemDTO{
			SKUID:       skuID,
			Qty:         it.Qty,
			UnitPrice:   unit,
			LineAmount:  unit * it.Qty,
			PriceSource: "snapshot",
		})
	}
	return items
}
