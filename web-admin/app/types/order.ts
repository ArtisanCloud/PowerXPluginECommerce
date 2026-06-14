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
  promotion?: PromotionSummary;
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
  priceSource?: string;
};

export type OrderEvent = {
  eventType: string;
  operatorType: string;
  operator?: string;
  createdAt: string;
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
