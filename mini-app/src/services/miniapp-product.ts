import { miniAppRequest } from "./miniapp-client";

export type MiniAppProductSummary = {
  id: string;
  code: string;
  name: string;
  type?: string;
  status?: string;
  updatedAt?: string;
  coverUrl?: string;
  minPrice?: number;
  maxPrice?: number;
  currency?: string;
  priceLabel?: string;
  skuCount?: number;
  sellability?: {
    sellable: boolean;
    reasons: string[];
  };
};

export type MiniAppProductListResponse = {
  items: MiniAppProductSummary[];
  page: number;
  pageSize: number;
  total: number;
};

export type MiniAppProductTagItem = {
  tag: string;
  count: number;
};

export type MiniAppProductTagListResponse = {
  items: MiniAppProductTagItem[];
};

export type MiniAppProductDetail = {
  id: string;
  code: string;
  name: string;
  type?: string;
  status?: string;
  categoryId?: string;
  categoryPath?: string;
  tags?: string[];
  coverUrl?: string;
  subtitle?: string;
  description?: string;
  minPrice?: number;
  maxPrice?: number;
  currency?: string;
  priceLabel?: string;
  skuCount?: number;
};

export type MiniAppSkuSummary = {
  id: string;
  spuId: string;
  code: string;
  status?: string;
  barcode?: string;
  createdAt?: string;
  updatedAt?: string;
  imageUrl?: string;
  price?: number;
  currency?: string;
  specSignature?: string;
  spec?: Record<string, string>;
  stockQty?: number;
  sellable?: boolean;
  sellabilityReasons?: string[];
  availableQty?: number;
};

export type MiniAppSkuListResponse = {
  items: MiniAppSkuSummary[];
  page: number;
  pageSize: number;
  total: number;
};

export type MiniAppSpecOption = {
  id: string;
  code: string;
  name: string;
  sortOrder?: number;
  meta?: any;
  status?: string;
};

export type MiniAppSpecGroup = {
  id: string;
  code: string;
  name: string;
  required?: boolean;
  sortOrder?: number;
  options: MiniAppSpecOption[];
};

export type MiniAppProductDetailWithSpecResponse = {
  spu: MiniAppProductDetail;
  spec?: {
    groups: MiniAppSpecGroup[];
  };
  skus: MiniAppSkuSummary[];
};

export type MiniAppProductListParams = {
  categoryId?: string;
  categoryPathPrefix?: string;
  tags?: string | string[];
  tag?: string;
  keyword?: string;
  q?: string;
  type?: string;
  sort?: "comprehensive" | "sales" | "price" | "updatedAt";
  order?: "asc" | "desc";
  minPrice?: number;
  maxPrice?: number;
  inStock?: boolean;
  hasPlans?: boolean;
  channel?: string;
  locale?: string;
  includeSellability?: 0 | 1;
  page?: number;
  pageSize?: number;
};

function normalizeTags(tags?: string | string[]) {
  if (tags === undefined || tags === null) return undefined;
  if (Array.isArray(tags)) {
    const cleaned = tags
      .map((t) => String(t ?? "").trim())
      .filter(Boolean);
    return cleaned.length ? cleaned.join(",") : undefined;
  }
  const s = String(tags).trim();
  return s ? s : undefined;
}

function buildQuery(params: Record<string, any>) {
  const parts: string[] = [];
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null) return;
    const s = String(v).trim();
    if (!s) return;
    parts.push(`${encodeURIComponent(k)}=${encodeURIComponent(s)}`);
  });
  return parts.length ? `?${parts.join("&")}` : "";
}

export async function miniAppListProducts(params: MiniAppProductListParams = {}) {
  const tags = normalizeTags(params.tags) ?? normalizeTags(params.tag);
  const query = buildQuery({
    categoryId: params.categoryId,
    categoryPathPrefix: params.categoryPathPrefix,
    tags,
    keyword: params.keyword,
    q: params.q,
    type: params.type,
    sort: params.sort,
    order: params.order,
    minPrice: params.minPrice,
    maxPrice: params.maxPrice,
    inStock: params.inStock,
    hasPlans: params.hasPlans,
    channel: params.channel,
    locale: params.locale,
    includeSellability: params.includeSellability,
    page: params.page ?? 1,
    pageSize: params.pageSize ?? 20,
  });
  return await miniAppRequest<MiniAppProductListResponse>({
    method: "GET",
    path: `/products${query}`,
  });
}

export async function miniAppListProductTags(params: { categoryId?: string; categoryPathPrefix?: string; limit?: number } = {}) {
  const query = buildQuery({
    categoryId: params.categoryId,
    categoryPathPrefix: params.categoryPathPrefix,
    limit: params.limit,
  });
  return await miniAppRequest<MiniAppProductTagListResponse>({
    method: "GET",
    path: `/products/tags${query}`,
  });
}

export async function miniAppGetProduct(id: string) {
  return await miniAppRequest<MiniAppProductDetail>({
    method: "GET",
    path: `/products/${encodeURIComponent(String(id))}`,
  });
}

export async function miniAppGetProductDetailWithSpec(id: string) {
  return await miniAppRequest<MiniAppProductDetailWithSpecResponse>({
    method: "GET",
    path: `/products/${encodeURIComponent(String(id))}/detail`,
  });
}

export async function miniAppListSkus(spuId: string, page = 1, pageSize = 50) {
  const query = buildQuery({ page, pageSize });
  return await miniAppRequest<MiniAppSkuListResponse>({
    method: "GET",
    path: `/products/${encodeURIComponent(String(spuId))}/skus${query}`,
  });
}

export type MiniAppSellabilityPrice = {
  amount: number;
  currency: string;
};

export type MiniAppSellabilityItem = {
  skuId: string;
  sellable: boolean;
  reasons: string[];
  price: MiniAppSellabilityPrice | null;
  availableQty: number;
};

export type MiniAppSellabilityResponse = {
  spuId: string;
  channel: string;
  items: MiniAppSellabilityItem[];
};

export async function miniAppGetSellability(spuId: string, params: { channel: string; locale?: string }) {
  const query = buildQuery({
    channel: params.channel,
    locale: params.locale,
  });
  return await miniAppRequest<MiniAppSellabilityResponse>({
    method: "GET",
    path: `/products/${encodeURIComponent(String(spuId))}/sellability${query}`,
  });
}

export type MiniAppSubscriptionPlan = {
  id: string;
  skuId?: string;
  planCode: string;
  name: string;
  billingCycle: string;
  billingValue?: number;
  price: number;
  currency: string;
  trialDays?: number;
  autoRenew?: boolean;
  cancelPolicy?: string;
  status?: string;
};

export type MiniAppSubscriptionPlanListResponse = {
  items: MiniAppSubscriptionPlan[];
};

export async function miniAppListSubscriptionPlans(spuId: string) {
  return await miniAppRequest<MiniAppSubscriptionPlanListResponse>({
    method: "GET",
    path: `/products/${encodeURIComponent(String(spuId))}/plans`,
  });
}
