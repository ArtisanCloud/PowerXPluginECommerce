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
  shippingAddressSnapshot?: ShippingAddress;
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

export type OrderBenefitReview = {
  id: number;
  orderId: string;
  orderNo: string;
  benefitType: string;
  benefitCode: string;
  valueType: string;
  value: number;
  amountMinor: number;
  currency: string;
  stackingAllowed: boolean;
  status: string;
  submittedBy: string;
  submittedAt: string;
  reviewedBy?: string;
  reviewedAt?: string | null;
  reviewReason?: string;
  note?: string;
  createdAt: string;
  updatedAt: string;
};

export type OrderDetail = {
  summary: OrderSummary;
  items: OrderItem[];
  events: OrderEvent[];
};

export type CancelOrderRequest = {
  reason?: string;
};

export type CreateOrderItemInput = {
  skuId: string;
  qty: number;
};

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
};

export type CreateOrderRequest = {
  customerId: string;
  channel: string;
  shippingAddressId?: string;
  shippingAddress?: ShippingAddress;
  items: CreateOrderItemInput[];
  note?: string;
};
