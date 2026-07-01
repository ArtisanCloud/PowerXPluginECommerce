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

// 后端金额为“分”（minor），例如 129900 表示 ¥1299.00
export type MoneyMinor = { currency: string; subtotal: number; total: number };

export type OrderSummary = {
  orderId: string;
  orderNo: string;
  status: string;
  amounts: MoneyMinor;
  promotion?: PromotionSummary;
  createdAt: string;
  shippingAddressSnapshot?: ShippingAddress;
};

export type OrderListResponse = {
  items: OrderSummary[];
  page: number;
  pageSize: number;
  total: number;
};

export type OrderItem = {
  skuId: string;
  qty: number;
  unitPrice: number;
  lineAmount: number;
  priceSource?: string;
};

export type OrderEvent = {
  eventType: string;
  operatorType: string;
  operator?: string;
  createdAt: string;
};

export type OrderDetail = {
  summary: OrderSummary;
  items: OrderItem[];
  events: OrderEvent[];
};

export type PromotionQuoteItem = {
  line_id?: string;
  sku_id: string;
  qty: number;
  unit_price_minor: number;
};

export type PromotionLineAllocation = {
  line_id?: string;
  sku_id: string;
  base_amount_minor: number;
  promotion_discount_minor: number;
  after_promotion_amount_minor: number;
};

export type PromotionApplied = {
  promotion_id: string;
  code: string;
  name: string;
  promotion_type: string;
  discount_minor: number;
  priority: number;
  exclusion_group?: string;
  stackable_with_coupon: boolean;
};

export type PromotionRejected = {
  promotion_id?: string;
  code?: string;
  reason: string;
};

export type PromotionSummary = {
  currency: string;
  base_total_minor: number;
  promotion_discount_minor: number;
  after_promotion_total_minor: number;
  coupon_stacking_allowed: boolean;
  line_allocations: PromotionLineAllocation[];
  applied_promotions: PromotionApplied[];
  rejected_promotions: PromotionRejected[];
  priced_at: string;
};

export type PromotionQuoteRequest = {
  channel: string;
  currency: string;
  items: PromotionQuoteItem[];
  submitted_at?: string;
};

function buildQuery(params: Record<string, any>) {
  const pairs: string[] = [];
  Object.keys(params || {}).forEach((k) => {
    const v = (params as any)[k];
    if (v === undefined || v === null || v === "") return;
    pairs.push(`${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`);
  });
  return pairs.length ? `?${pairs.join("&")}` : "";
}

export async function createOrder(req: CreateOrderRequest, idempotencyKey: string) {
  return await miniAppRequest<OrderSummary>({
    method: "POST",
    path: "/orders",
    data: req,
    headers: { "Idempotency-Key": idempotencyKey },
  });
}

export async function quotePromotions(req: PromotionQuoteRequest) {
  return await miniAppRequest<PromotionSummary>({
    method: "POST",
    path: "/promotions/quote",
    data: req,
  });
}

export async function miniAppListMyOrders(params: { page?: number; pageSize?: number } = {}) {
  const query = buildQuery({ page: params.page, pageSize: params.pageSize });
  return await miniAppRequest<OrderListResponse>({
    method: "GET",
    path: `/orders${query}`,
  });
}

export async function miniAppGetMyOrder(id: string) {
  return await miniAppRequest<OrderDetail>({
    method: "GET",
    path: `/orders/${encodeURIComponent(String(id))}`,
  });
}
