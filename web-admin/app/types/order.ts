export type Money = {
  currency: string;
  subtotal: number;
  total: number;
};

export type OrderSummary = {
  orderId: string;
  orderNo: string;
  customerId?: string;
  channel?: string;
  createdByType?: string;
  status: string;
  amounts: Money;
  createdAt: string;
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

export type CancelOrderRequest = {
  reason?: string;
};
