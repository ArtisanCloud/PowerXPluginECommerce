import { miniAppRequest } from "./miniapp-client";

export type CreateOrderItem = { skuId: string; qty: number };

export type ShippingAddress = {
  label?: string;
  recipientName: string;
  recipientPhone: string;
  countryCode?: string;
  province?: string;
  city?: string;
  district?: string;
  address1: string;
  address2?: string;
  postalCode?: string;
  metadata?: Record<string, any>;
};

export type CreateOrderRequest = {
  channel: string;
  items: CreateOrderItem[];
  locale?: string;
  shippingAddressId?: string;
  shippingAddress?: ShippingAddress;
};

export type Money = { currency: string; subtotal: number; total: number };

export type OrderSummary = {
  orderId: string;
  orderNo: string;
  status: string;
  amounts: Money;
  createdAt: string;
};

export async function createOrder(req: CreateOrderRequest, idempotencyKey: string) {
  return await miniAppRequest<OrderSummary>({
    method: "POST",
    path: "/orders",
    data: req,
    headers: { "Idempotency-Key": idempotencyKey },
  });
}
