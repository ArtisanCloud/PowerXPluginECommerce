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
