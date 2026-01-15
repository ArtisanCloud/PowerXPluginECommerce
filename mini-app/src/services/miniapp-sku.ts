import { miniAppRequest } from "./miniapp-client";

export type MiniAppSkuBatchItem = {
  id: string;
  spuId: string;
  code?: string;
  spuName?: string;
  imageUrl?: string;
  price?: number | null;
  currency?: string;
  status?: string;
};

export type MiniAppSkuBatchResponse = {
  items: MiniAppSkuBatchItem[];
};

export async function miniAppBatchSkus(skuIds: string[]) {
  const ids = Array.from(new Set((skuIds || []).map((x) => String(x || "").trim()).filter(Boolean)));
  if (!ids.length) return { items: [] } as MiniAppSkuBatchResponse;
  return await miniAppRequest<MiniAppSkuBatchResponse>({
    method: "POST",
    path: "/skus/batch",
    data: { skuIds: ids },
  });
}

