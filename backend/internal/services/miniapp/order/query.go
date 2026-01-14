package order

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	"gorm.io/gorm"
)

func (s *Service) ListOrders(ctx context.Context, tenantUUID, customerID string, page, pageSize int) (*OrderListResponse, error) {
	if !s.Ready() {
		return nil, ErrOrderServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}

	result, err := s.OrderRepo.List(ctx, tenantUUID, orderrepo.OrderListFilter{CustomerID: customerID}, page, pageSize)
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

func (s *Service) GetOrderDetail(ctx context.Context, tenantUUID, customerID, orderID string) (*OrderDetailDTO, error) {
	if !s.Ready() {
		return nil, ErrOrderServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	orderID = strings.TrimSpace(orderID)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
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
	if strings.TrimSpace(ord.CustomerID) != customerID {
		return nil, ErrOrderNotFound
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
			SKUID:       it.SKUID,
			Qty:         it.Qty,
			UnitPrice:   it.UnitPrice,
			LineAmount:  it.LineAmount,
			PriceSource: it.PriceSource,
		})
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
			Status:                  ord.Status,
			Amounts:                 MoneyDTO{Currency: ord.Currency, Subtotal: ord.SubtotalAmount, Total: ord.TotalAmount},
			ShippingAddressSnapshot: shippingSnap,
			CreatedAt:               ord.CreatedAt,
		},
		Items:  items,
		Events: evs,
	}, nil
}
