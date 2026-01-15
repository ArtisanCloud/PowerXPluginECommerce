<template>
  <view class="min-h-screen w-full bg-background-light font-display text-gray-900" style="padding-bottom: 96px;">
    <view class="fixed top-0 left-0 right-0 z-50 bg-white">
      <view class="flex items-center justify-between px-4" :style="`height:${headerHeight}px; padding-top:${topInset}px;`">
        <view class="flex items-center justify-center h-10 w-10 -ml-2" hover-class="opacity-80" @tap="goBack">
          <text class="text-2xl">‹</text>
        </view>
        <text class="text-base font-extrabold">我的订单</text>
        <view class="flex items-center justify-center h-10 w-10 -mr-2" hover-class="opacity-80" @tap="onSearch">
          <text class="text-xl">⌕</text>
        </view>
      </view>
      <scroll-view scroll-x class="w-full border-b" style="border-color: rgba(0,0,0,0.06); white-space: nowrap;">
        <view class="flex px-2">
          <view
            v-for="t in tabs"
            :key="t.key"
            class="flex-none px-4 py-3"
            :style="t.key === activeTab ? 'color:#4F6F52;border-bottom:2px solid #4F6F52;font-weight:900;' : 'color:rgba(115,115,115,1);'"
            @tap="setTab(t.key)"
          >
            <text class="text-sm font-extrabold">{{ t.label }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <view :style="`padding-top:${headerHeight + 44}px;`" class="px-4 space-y-4">
      <view v-if="loading && orders.length === 0" class="py-10 text-center">
        <text class="text-sm text-muted">加载中...</text>
      </view>
      <view v-else-if="errorMsg && orders.length === 0" class="py-10 text-center">
        <text class="text-sm" style="color:#ef4444;">{{ errorMsg }}</text>
      </view>
      <view v-else-if="filteredOrders.length === 0" class="py-10 text-center">
        <text class="text-sm text-muted">暂无订单</text>
      </view>

      <view
        v-for="o in filteredOrders"
        :key="o.orderId"
        class="rounded-2xl bg-white p-4 shadow-sm"
        hover-class="opacity-95"
        @tap="openDetail(o.orderId)"
      >
        <view class="flex items-center justify-between mb-4">
          <view class="flex items-center" style="gap: 6px;">
            <text class="text-base">🏬</text>
            <text class="text-sm font-extrabold">{{ storeName }}</text>
            <text class="text-muted">›</text>
          </view>
          <text class="text-sm font-semibold" :style="statusColorStyle(o.status)">{{ statusText(o.status) }}</text>
        </view>

        <view v-if="o.firstItem" class="flex gap-3 mb-4">
          <image class="h-20 w-20 shrink-0 rounded-lg bg-gray-100" mode="aspectFill" :src="o.firstItem.thumb" @error="onItemThumbError(o.orderId)" />
          <view class="flex-1 min-w-0">
            <text class="text-sm font-semibold" :number-of-lines="2" style="line-height: 1.35;">{{ o.firstItem.title }}</text>
            <text class="mt-1 text-xs text-muted" :number-of-lines="1">规格：{{ o.firstItem.skuLabel || "—" }}</text>
            <view class="mt-2 flex items-end justify-between">
              <text class="text-sm font-extrabold">{{ formatMoneyFromMinor(o.amounts.currency, o.firstItem.unitPriceMinor) }}</text>
              <text class="text-xs text-muted">x{{ o.firstItem.qty }}</text>
            </view>
          </view>
        </view>
        <view v-else class="py-4">
          <text class="text-sm text-muted">订单商品加载中...</text>
        </view>

        <view class="flex items-center justify-end gap-2 pt-3 border-t" style="border-color: rgba(0,0,0,0.04);">
          <text class="text-xs text-muted mr-2">实付:</text>
          <text class="text-sm font-extrabold text-gray-900 mr-2">{{ formatMoneyFromMinor(o.amounts.currency, o.amounts.total) }}</text>
          <view v-for="b in actionButtons(o.status)" :key="b.key" class="px-4 py-2 rounded-full" :style="b.primary ? 'background:#4F6F52;color:#fff;' : 'border:1px solid rgba(0,0,0,0.12);color:#111827;'" hover-class="opacity-90" @tap.stop="onAction(b.key, o)">
            <text class="text-xs font-extrabold">{{ b.label }}</text>
          </view>
        </view>
      </view>

      <view class="py-6 flex items-center justify-center" style="color:#9ca3af;font-size:12px;">
        <text v-if="loading">加载中...</text>
        <text v-else-if="noMore">没有更多了</text>
        <text v-else hover-class="opacity-80" @tap="loadMore">加载更多</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onLoad, onShow } from "@dcloudio/uni-app";
import { miniAppBatchSkus, type MiniAppSkuBatchItem } from "@/services/miniapp-sku";
import { miniAppGetMyOrder, miniAppListMyOrders, type OrderDetail, type OrderSummary } from "@/services/miniapp-order";
import { miniAppGetProduct } from "@/services/miniapp-product";
import { isLikelyPlaceholderUrl, pickPlaceholderImage } from "@/utils/product-images";

type TabKey = "all" | "pending_payment" | "paid" | "shipped" | "completed" | "refund";

type OrderCardVM = OrderSummary & {
  firstItem?: {
    skuId: string;
    qty: number;
    unitPriceMinor: number;
    title: string;
    skuLabel?: string;
    thumb: string;
  };
};

const storeName = "官方自营旗舰店";

const topInset = ref(44);
const headerHeight = computed(() => topInset.value + 48);

const tabs = [
  { key: "all" as const, label: "全部" },
  { key: "pending_payment" as const, label: "待付款" },
  { key: "paid" as const, label: "待发货" },
  { key: "shipped" as const, label: "待收货" },
  { key: "completed" as const, label: "已完成" },
  { key: "refund" as const, label: "退款/售后" },
];

const activeTab = ref<TabKey>("all");

const loading = ref(false);
const errorMsg = ref("");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

const orders = ref<OrderCardVM[]>([]);
const detailsByOrderId = new Map<string, OrderDetail>();
const skuById = ref<Map<string, MiniAppSkuBatchItem>>(new Map());
const spuCoverById = ref<Map<string, string>>(new Map());
const brokenThumbOrderIds = ref<Set<string>>(new Set());

function ensureTopInset() {
  try {
    const wxAny = (globalThis as any).wx;
    if (wxAny && typeof wxAny.getWindowInfo === "function") {
      const info = wxAny.getWindowInfo();
      const statusBar = Number(info?.statusBarHeight ?? 0);
      topInset.value = Math.max(44, statusBar + 12);
      return;
    }
    const sys = uni.getSystemInfoSync();
    const statusBar = Number((sys as any).statusBarHeight ?? 0);
    topInset.value = Math.max(44, statusBar + 12);
  } catch {
    topInset.value = 44;
  }
}

function goBack() {
  uni.navigateBack();
}

function onSearch() {
  uni.showToast({ title: "搜索待接入", icon: "none" });
}

function openDetail(orderId: string) {
  const id = String(orderId || "").trim();
  if (!id) return;
  uni.navigateTo({ url: `/pages/order/detail?id=${encodeURIComponent(id)}` });
}

function statusText(status: string) {
  const s = String(status || "").trim();
  if (s === "pending_payment") return "待付款";
  if (s === "paid") return "待发货";
  if (s === "shipped") return "待收货";
  if (s === "completed") return "交易成功";
  if (s === "cancelled" || s === "canceled") return "已取消";
  if (s.includes("refund")) return "退款/售后";
  return s || "—";
}

function statusColorStyle(status: string) {
  const s = String(status || "").trim();
  if (s === "pending_payment") return "color:#f97316;";
  if (s === "paid" || s === "shipped") return "color:#4F6F52;";
  if (s === "completed") return "color:#6b7280;";
  if (s === "cancelled" || s === "canceled") return "color:#9ca3af;";
  if (s.includes("refund")) return "color:#ef4444;";
  return "color:#6b7280;";
}

function formatMoneyFromMinor(cur: string, minor: number) {
  const c = String(cur || "CNY").trim().toUpperCase() || "CNY";
  const symbol = c === "CNY" || c === "RMB" ? "¥" : c === "USD" ? "$" : c === "EUR" ? "€" : `${c} `;
  const n = Number(minor);
  if (!Number.isFinite(n)) return `${symbol}--`;
  const major = n / 100;
  return `${symbol}${major.toFixed(2)}`;
}

function matchTab(status: string, tab: TabKey) {
  if (tab === "all") return true;
  const s = String(status || "").trim();
  if (tab === "refund") return s.includes("refund");
  return s === tab;
}

const filteredOrders = computed(() => (orders.value || []).filter((o) => matchTab(o.status, activeTab.value)));
const noMore = computed(() => orders.value.length >= total.value && total.value > 0);

function setTab(key: TabKey) {
  activeTab.value = key;
}

function onItemThumbError(orderId: string) {
  const id = String(orderId || "").trim();
  if (!id) return;
  brokenThumbOrderIds.value.add(id);
  orders.value = orders.value.map((o) => {
    if (o.orderId !== id) return o;
    if (!o.firstItem) return o;
    return { ...o, firstItem: { ...o.firstItem, thumb: "/static/icons/image-placeholder.png" } };
  });
}

function actionButtons(status: string) {
  const s = String(status || "").trim();
  if (s === "pending_payment") return [{ key: "cancel", label: "取消订单" }, { key: "pay", label: "去支付", primary: true }];
  if (s === "shipped") return [{ key: "logistics", label: "查看物流" }, { key: "confirm", label: "确认收货", primary: true }];
  if (s === "completed") return [{ key: "after", label: "申请售后" }, { key: "review", label: "评价", primary: true }];
  if (s === "paid") return [{ key: "remind", label: "提醒发货" }];
  return [];
}

function onAction(action: string, o: OrderCardVM) {
  if (action === "pay") return uni.showToast({ title: "支付待接入", icon: "none" });
  if (action === "cancel") return uni.showToast({ title: "取消订单待接入", icon: "none" });
  if (action === "logistics") return uni.showToast({ title: "物流待接入", icon: "none" });
  if (action === "confirm") return uni.showToast({ title: "确认收货待接入", icon: "none" });
  if (action === "after") return uni.showToast({ title: "售后待接入", icon: "none" });
  if (action === "review") return uni.showToast({ title: "评价待接入", icon: "none" });
  if (action === "remind") return uni.showToast({ title: "已提醒（示意）", icon: "none" });
  uni.showToast({ title: "功能待接入", icon: "none" });
}

async function mapWithConcurrency<T, R>(items: T[], limit: number, fn: (it: T) => Promise<R>): Promise<R[]> {
  const out: R[] = [];
  const queue = items.slice();
  const workers = Array.from({ length: Math.max(1, limit) }).map(async () => {
    while (queue.length) {
      const it = queue.shift()!;
      try {
        out.push(await fn(it));
      } catch {
        // ignore
      }
    }
  });
  await Promise.all(workers);
  return out;
}

function buildCardFromDetail(summary: OrderSummary, detail?: OrderDetail) {
  const id = String(summary.orderId || "").trim();
  const d = detail || detailsByOrderId.get(id);
  if (!d || !Array.isArray(d.items) || !d.items.length) return { ...(summary as any) } as OrderCardVM;

  const first = d.items[0];
  const sku = skuById.value.get(String(first.skuId || "").trim());
  const title = String(sku?.spuName || "").trim() || `SKU ${String(first.skuId || "").slice(0, 8)}`;
  const skuLabel = String(sku?.code || "").trim() || undefined;
  const spuId = String((sku as any)?.spuId || "").trim();
  const coverUrl = spuId ? String(spuCoverById.value.get(spuId) || "").trim() : "";
  const thumbKey = String(spuId || summary.orderId || summary.orderNo || first.skuId || "").trim();
  const thumb =
    (brokenThumbOrderIds.value.has(id) ? "" : "") ||
    (coverUrl && !isLikelyPlaceholderUrl(coverUrl) ? coverUrl : "") ||
    pickPlaceholderImage(thumbKey);
  return {
    ...(summary as any),
    firstItem: {
      skuId: String(first.skuId || ""),
      qty: Number(first.qty || 0) || 0,
      unitPriceMinor: Number(first.unitPrice || 0) || 0,
      title,
      skuLabel,
      thumb,
    },
  } as OrderCardVM;
}

async function enrichOrdersWithDetail(summaries: OrderSummary[]) {
  const ids = summaries.map((s) => String(s.orderId || "").trim()).filter(Boolean);
  const details = await mapWithConcurrency(ids, 6, async (id) => {
    const d = await miniAppGetMyOrder(id);
    detailsByOrderId.set(id, d);
    return d;
  });

  const skuIds = Array.from(
    new Set(
      details
        .flatMap((d) => (d?.items || []).map((it) => String(it?.skuId || "").trim()))
        .filter(Boolean),
    ),
  );
  if (skuIds.length) {
    try {
      const resp = await miniAppBatchSkus(skuIds);
      const next = new Map<string, MiniAppSkuBatchItem>();
      (resp?.items || []).forEach((it) => next.set(String(it.id || "").trim(), it));
      skuById.value = next;
    } catch {
      // ignore
    }
  }

  // 为了与商城/购物车保持一致：按 spuId 拉取商品 coverUrl
  const spuIds = Array.from(
    new Set(
      Array.from(skuById.value.values())
        .map((x) => String((x as any)?.spuId || "").trim())
        .filter(Boolean),
    ),
  );
  if (spuIds.length) {
    const next = new Map<string, string>(spuCoverById.value);
    await mapWithConcurrency(spuIds.filter((id) => !next.has(id)), 6, async (spuId) => {
      try {
        const p = await miniAppGetProduct(spuId);
        const coverUrl = String((p as any)?.coverUrl || "").trim();
        if (coverUrl) next.set(spuId, coverUrl);
      } catch {
        // ignore
      }
      return null as any;
    });
    spuCoverById.value = next;
  }

  const nextCards = summaries.map((s) => buildCardFromDetail(s));
  orders.value = orders.value.concat(nextCards);
}

async function loadFirstPage() {
  if (loading.value) return;
  loading.value = true;
  errorMsg.value = "";
  page.value = 1;
  total.value = 0;
  orders.value = [];
  detailsByOrderId.clear();
  skuById.value = new Map();
  spuCoverById.value = new Map();
  brokenThumbOrderIds.value = new Set();
  try {
    const resp = await miniAppListMyOrders({ page: 1, pageSize: pageSize.value });
    total.value = Number(resp?.total || 0) || 0;
    page.value = Number(resp?.page || 1) || 1;
    const items = Array.isArray(resp?.items) ? resp.items : [];
    await enrichOrdersWithDetail(items);
  } catch (e: any) {
    errorMsg.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadMore() {
  if (loading.value) return;
  if (noMore.value) return;
  loading.value = true;
  errorMsg.value = "";
  try {
    const nextPage = page.value + 1;
    const resp = await miniAppListMyOrders({ page: nextPage, pageSize: pageSize.value });
    total.value = Number(resp?.total || total.value) || total.value;
    page.value = Number(resp?.page || nextPage) || nextPage;
    const items = Array.isArray(resp?.items) ? resp.items : [];
    await enrichOrdersWithDetail(items);
  } catch (e: any) {
    uni.showToast({ title: e?.message || "加载失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}

onLoad((opts) => {
  ensureTopInset();
  const t = String((opts as any)?.tab || "").trim() as TabKey;
  if (tabs.some((x) => x.key === t)) activeTab.value = t;
});

onMounted(() => ensureTopInset());

onShow(() => {
  ensureTopInset();
  void loadFirstPage();
});
</script>
