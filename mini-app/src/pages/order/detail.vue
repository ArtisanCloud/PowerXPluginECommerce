<template>
  <view class="min-h-screen w-full bg-background-light font-display text-gray-900" style="padding-bottom: 120px;">
    <view
      class="sticky top-0 z-50 bg-white px-4 py-3"
      :style="`padding-top:${topInset}px; backdrop-filter: blur(12px); background: rgba(255,255,255,0.9);`"
    >
      <view class="flex items-center justify-between">
        <view class="flex items-center justify-center h-10 w-10 -ml-2" hover-class="opacity-80" @tap="goBack">
          <text class="text-2xl">‹</text>
        </view>
        <text class="text-base font-extrabold">订单详情</text>
        <view class="flex items-center justify-center h-10 w-10 -mr-2" hover-class="opacity-80" @tap="onShare">
          <text class="text-xl">↗</text>
        </view>
      </view>
    </view>

    <view class="px-6 py-8 text-white" style="background: linear-gradient(135deg, #4F6F52 0%, #739072 100%);">
      <view class="flex items-center justify-between">
        <view>
          <text class="text-2xl font-extrabold">{{ statusText }}</text>
          <view class="mt-2 text-sm" style="opacity: 0.85;">{{ statusSubText }}</view>
        </view>
        <text class="text-5xl" style="opacity: 0.35;">🚚</text>
      </view>
    </view>

    <view class="px-4 -mt-4 relative z-10 space-y-3">
      <view class="rounded-xl bg-white p-4 shadow-sm flex items-start" style="gap: 10px;">
        <text class="text-primary text-lg">📦</text>
        <view class="flex-1">
          <text class="text-sm font-semibold" style="line-height: 1.5;">{{ latestEventText }}</text>
          <view class="mt-1 text-xs text-muted">{{ latestEventTime }}</view>
        </view>
        <text class="text-muted">›</text>
      </view>

      <view class="rounded-xl bg-white p-4 shadow-sm flex items-start" style="gap: 10px;">
        <text class="text-primary text-lg">📍</text>
        <view class="flex-1">
          <view class="flex items-center" style="gap: 8px;">
            <text class="text-sm font-extrabold">{{ addrName }}</text>
            <text class="text-sm text-muted">{{ addrPhone }}</text>
          </view>
          <text class="mt-1 text-xs text-muted" style="line-height: 1.5;">{{ addrFull }}</text>
        </view>
      </view>

      <view class="rounded-xl bg-white p-4 shadow-sm">
        <view class="flex items-center mb-4 border-b pb-2" style="gap: 8px; border-color: rgba(0,0,0,0.04);">
          <text class="text-lg">🏬</text>
          <text class="text-sm font-extrabold">{{ storeName }}</text>
        </view>

        <view v-if="loading" class="py-6 text-center">
          <text class="text-sm text-muted">加载中...</text>
        </view>
        <view v-else-if="errorMsg" class="py-6 text-center">
          <text class="text-sm" style="color:#ef4444;">{{ errorMsg }}</text>
        </view>

        <view v-else class="space-y-4">
          <view v-for="it in displayItems" :key="it.skuId" class="flex" style="gap: 12px;">
            <image class="h-20 w-20 rounded-lg bg-gray-100 shrink-0" mode="aspectFill" :src="it.thumb" @error="onItemThumbError(it.skuId)" />
            <view class="flex-1 flex flex-col justify-between min-w-0">
              <view>
                <text class="text-sm font-semibold" :number-of-lines="2">{{ it.title }}</text>
                <text class="mt-1 text-xs text-muted">规格：{{ it.skuLabel || "—" }}</text>
              </view>
              <view class="flex items-end justify-between">
                <text class="text-sm font-extrabold">{{ formatMoneyFromMinor(currency, it.unitPriceMinor) }}</text>
                <text class="text-xs text-muted">x{{ it.qty }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <view class="rounded-xl bg-white p-4 shadow-sm space-y-3">
        <view class="flex items-center justify-between text-xs">
          <text class="text-muted">订单编号</text>
          <view class="flex items-center" style="gap: 6px;" hover-class="opacity-80" @tap="copy(orderNo)">
            <text class="text-xs">{{ orderNo }}</text>
            <text class="text-primary text-xs">复制</text>
          </view>
        </view>
        <view class="flex items-center justify-between text-xs">
          <text class="text-muted">创建时间</text>
          <text class="text-xs">{{ createdAtText }}</text>
        </view>
        <view class="flex items-center justify-between text-xs">
          <text class="text-muted">支付方式</text>
          <text class="text-xs">微信支付</text>
        </view>
        <view class="flex items-center justify-between text-xs">
          <text class="text-muted">订单备注</text>
          <text class="text-xs text-muted">无</text>
        </view>
      </view>

      <view class="rounded-xl bg-white p-4 shadow-sm space-y-3">
        <view class="flex items-center justify-between text-sm">
          <text class="text-muted">商品总额</text>
          <text class="text-sm font-semibold">{{ subtotalText }}</text>
        </view>
        <view class="flex items-center justify-between text-sm">
          <text class="text-muted">运费</text>
          <text class="text-sm font-semibold">{{ shippingText }}</text>
        </view>
        <view class="flex items-center justify-between text-sm">
          <text class="text-muted">自动促销</text>
          <text class="text-sm font-semibold" style="color:#ef4444;">-{{ discountText }}</text>
        </view>
        <view v-if="appliedPromotionNames" class="flex items-center justify-between text-xs" style="gap: 12px;">
          <text class="text-muted shrink-0">已享活动</text>
          <text class="text-primary text-right" :number-of-lines="2">{{ appliedPromotionNames }}</text>
        </view>
        <view class="pt-3 border-t flex justify-end items-center" style="gap: 6px; border-color: rgba(0,0,0,0.04);">
          <text class="text-sm font-semibold">实付金额：</text>
          <text class="text-xl font-extrabold text-primary">{{ totalText }}</text>
        </view>
      </view>
    </view>

    <view class="fixed bottom-0 left-0 right-0 z-50 bg-white border-t" style="border-color: rgba(0,0,0,0.06);">
      <view class="px-4 pb-2" :style="`padding-bottom: calc(env(safe-area-inset-bottom) + 8px);`">
        <view class="flex justify-end gap-3 py-3 border-b" style="border-color: rgba(0,0,0,0.04);">
          <template v-if="canPay">
            <view class="px-5 py-2 rounded-full border" style="border-color: rgba(0,0,0,0.12);" hover-class="opacity-90" @tap="onAction('contact')">
              <text class="text-sm font-semibold">联系客服</text>
            </view>
            <view class="px-5 py-2 rounded-full" style="background:#4F6F52;" hover-class="opacity-90" @tap="onPay">
              <text class="text-sm font-semibold text-white">{{ isPaying ? "支付中..." : "去支付" }}</text>
            </view>
          </template>
          <template v-else>
            <view class="px-5 py-2 rounded-full border" style="border-color: rgba(0,0,0,0.12);" hover-class="opacity-90" @tap="onAction('contact')">
              <text class="text-sm font-semibold">联系客服</text>
            </view>
            <view class="px-5 py-2 rounded-full border" style="border-color: rgba(0,0,0,0.12);" hover-class="opacity-90" @tap="onAction('logistics')">
              <text class="text-sm font-semibold">查看物流</text>
            </view>
            <view class="px-5 py-2 rounded-full" style="background:#4F6F52;" hover-class="opacity-90" @tap="onAction('confirm')">
              <text class="text-sm font-semibold text-white">确认收货</text>
            </view>
          </template>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import { buildIdempotencyKey, createPaymentTransaction, getWechatProviderIdNumber, requestMiniAppPayment } from "@/services/miniapp-payment";
import { miniAppBatchSkus, type MiniAppSkuBatchItem } from "@/services/miniapp-sku";
import { miniAppGetMyOrder, type OrderDetail } from "@/services/miniapp-order";
import { maskPhone } from "@/services/miniapp-address";
import { miniAppGetProduct } from "@/services/miniapp-product";
import { isLikelyPlaceholderUrl, pickPlaceholderImage } from "@/utils/product-images";

const storeName = "官方自营旗舰店";

const topInset = ref(44);
const loading = ref(false);
const errorMsg = ref("");
const isPaying = ref(false);

const orderId = ref("");
const detail = ref<OrderDetail | null>(null);
const skuMap = ref<Map<string, MiniAppSkuBatchItem>>(new Map());
const spuCoverById = ref<Map<string, string>>(new Map());
const brokenSkuIds = ref<Set<string>>(new Set());

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

function onShare() {
  uni.showToast({ title: "分享待接入", icon: "none" });
}

function onAction(key: string) {
  if (key === "contact") return uni.showToast({ title: "客服待接入", icon: "none" });
  if (key === "logistics") return uni.showToast({ title: "物流待接入", icon: "none" });
  if (key === "confirm") return uni.showToast({ title: "确认收货待接入", icon: "none" });
  uni.showToast({ title: "功能待接入", icon: "none" });
}

function copy(text: string) {
  const s = String(text || "").trim();
  if (!s) return;
  uni.setClipboardData({ data: s, success: () => uni.showToast({ title: "已复制", icon: "none" }) });
}

function formatMoneyFromMinor(cur: string, minor: number) {
  const c = String(cur || "CNY").trim().toUpperCase() || "CNY";
  const symbol = c === "CNY" || c === "RMB" ? "¥" : c === "USD" ? "$" : c === "EUR" ? "€" : `${c} `;
  const n = Number(minor);
  if (!Number.isFinite(n)) return `${symbol}--`;
  const major = n / 100;
  return `${symbol}${major.toFixed(2)}`;
}

function formatDateTime(iso: string) {
  const s = String(iso || "").trim();
  if (!s) return "—";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

const currency = computed(() => detail.value?.summary?.amounts?.currency || "CNY");
const orderNo = computed(() => detail.value?.summary?.orderNo || "—");
const createdAtText = computed(() => formatDateTime(detail.value?.summary?.createdAt || ""));

const subtotalText = computed(() => formatMoneyFromMinor(currency.value, Number(detail.value?.summary?.amounts?.subtotal || 0)));
const totalText = computed(() => formatMoneyFromMinor(currency.value, Number(detail.value?.summary?.amounts?.total || 0)));
const shippingText = computed(() => formatMoneyFromMinor(currency.value, 0));
const promotionDiscountMinor = computed(() => Number(detail.value?.summary?.promotion?.promotion_discount_minor || 0));
const discountText = computed(() => formatMoneyFromMinor(currency.value, promotionDiscountMinor.value));
const appliedPromotionNames = computed(() => {
  const list = detail.value?.summary?.promotion?.applied_promotions || [];
  return list.map((item: any) => item.name || item.code).filter(Boolean).join("、");
});

const statusText = computed(() => {
  const s = String(detail.value?.summary?.status || "").trim();
  if (s === "pending_payment") return "待付款";
  if (s === "paying") return "支付处理中";
  if (s === "paid") return "待发货";
  if (s === "shipped") return "待收货";
  if (s === "completed") return "交易成功";
  if (s === "cancelled" || s === "canceled") return "已取消";
  if (s.includes("refund")) return "退款/售后";
  return s || "—";
});

const statusSubText = computed(() => {
  const s = String(detail.value?.summary?.status || "").trim();
  if (s === "pending_payment") return "订单已创建，请尽快完成支付";
  if (s === "paying") return "支付处理中，请稍等片刻";
  if (s === "paid") return "已支付，等待卖家发货";
  if (s === "shipped") return "包裹正在派送中，请耐心等待";
  if (s === "completed") return "订单已完成，感谢你的支持";
  if (s === "cancelled" || s === "canceled") return "订单已取消";
  return "订单状态更新中";
});

const canPay = computed(() => String(detail.value?.summary?.status || "").trim() === "pending_payment");

const latestEvent = computed(() => {
  const evs = Array.isArray(detail.value?.events) ? detail.value!.events : [];
  if (!evs.length) return null;
  // 后端通常是按创建时间升序返回；取最后一条当“最新”
  return evs[evs.length - 1] || null;
});

const latestEventText = computed(() => {
  const ev = latestEvent.value;
  if (!ev) return "订单已创建";
  const t = String(ev.eventType || "").trim();
  if (t === "order.created") return "订单已创建";
  return t || "订单状态更新";
});
const latestEventTime = computed(() => formatDateTime(latestEvent.value?.createdAt || ""));

const addr = computed(() => detail.value?.summary?.shippingAddressSnapshot || null);
const addrName = computed(() => String(addr.value?.recipientName || "").trim() || "—");
const addrPhone = computed(() => maskPhone(String(addr.value?.recipientPhone || "")));
const addrFull = computed(() => {
  const a = addr.value;
  if (!a) return "—";
  return [a.province, a.city, a.district, a.address1, a.address2].filter(Boolean).join("");
});

const displayItems = computed(() => {
  const items = Array.isArray(detail.value?.items) ? detail.value!.items : [];
  return items.map((it) => {
    const skuId = String((it as any)?.skuId || "").trim();
    const sku = skuMap.value.get(skuId);
    const spuId = String((sku as any)?.spuId || "").trim();
    const coverUrl = spuId ? String(spuCoverById.value.get(spuId) || "").trim() : "";
    const title = String(sku?.spuName || "").trim() || `SKU ${skuId.slice(0, 8)}`;
    const skuLabel = String(sku?.code || "").trim() || undefined;
    const thumbKey = String(spuId || orderNo.value || skuId);
    const thumb =
      (brokenSkuIds.value.has(skuId) ? "" : "") ||
      (coverUrl && !isLikelyPlaceholderUrl(coverUrl) ? coverUrl : "") ||
      pickPlaceholderImage(thumbKey);
    return {
      skuId,
      qty: Number((it as any)?.qty || 0) || 0,
      unitPriceMinor: Number((it as any)?.unitPrice || 0) || 0,
      title,
      skuLabel,
      thumb,
    };
  });
});

function onItemThumbError(skuId: string) {
  const id = String(skuId || "").trim();
  if (!id) return;
  brokenSkuIds.value.add(id);
}

async function load() {
  if (!orderId.value) return;
  if (loading.value) return;
  loading.value = true;
  errorMsg.value = "";
  try {
    const d = await miniAppGetMyOrder(orderId.value);
    detail.value = d;
    const skuIds = Array.from(new Set((d?.items || []).map((x: any) => String(x?.skuId || "").trim()).filter(Boolean)));
    if (skuIds.length) {
      const resp = await miniAppBatchSkus(skuIds);
      const map = new Map<string, MiniAppSkuBatchItem>();
      (resp?.items || []).forEach((x) => map.set(String(x.id || "").trim(), x));
      skuMap.value = map;

      const spuIds = Array.from(
        new Set(
          (resp?.items || [])
            .map((x: any) => String(x?.spuId || "").trim())
            .filter(Boolean),
        ),
      );
      if (spuIds.length) {
        const next = new Map<string, string>();
        await Promise.all(
          spuIds.slice(0, 20).map(async (spuId) => {
            try {
              const p = await miniAppGetProduct(spuId);
              const coverUrl = String((p as any)?.coverUrl || "").trim();
              if (coverUrl) next.set(spuId, coverUrl);
            } catch {
              // ignore
            }
          }),
        );
        spuCoverById.value = next;
      } else {
        spuCoverById.value = new Map();
      }
    } else {
      skuMap.value = new Map();
      spuCoverById.value = new Map();
    }
  } catch (e: any) {
    errorMsg.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

async function onPay() {
  if (isPaying.value) return;
  const d = detail.value?.summary;
  if (!d) return;
  const orderIdText = String(d.orderId || "").trim();
  const orderNoText = String(d.orderNo || "").trim();
  if (!orderIdText || !orderNoText) return;
  const providerId = getWechatProviderIdNumber();
  if (!providerId) {
    uni.showToast({ title: "支付渠道未配置", icon: "none" });
    return;
  }
  isPaying.value = true;
  try {
    const resp = await createPaymentTransaction({
      orderId: orderIdText,
      payMethod: "wechat_jsapi",
      providerId,
      client: "miniapp",
      idempotencyKey: buildIdempotencyKey("pay"),
    });
    let status = "paying";
    if (resp?.wechat?.appId) {
      const outcome = await requestMiniAppPayment(resp.wechat);
      if (outcome.status === "cancel") status = "canceled";
      if (outcome.status === "fail") status = "failed";
      if (outcome.message) uni.showToast({ title: outcome.message, icon: "none" });
    } else {
      status = "pending_payment";
      uni.showToast({ title: "支付渠道未配置", icon: "none" });
    }
    const query = `orderId=${encodeURIComponent(orderIdText)}&orderNo=${encodeURIComponent(orderNoText)}&status=${encodeURIComponent(status)}&currency=${encodeURIComponent(
      String(d.amounts?.currency || "CNY"),
    )}&total=${encodeURIComponent(String(d.amounts?.total || 0))}&transactionId=${encodeURIComponent(
      String(resp?.transactionId || ""),
    )}`;
    uni.navigateTo({ url: `/pages/order/success?${query}` });
  } catch (e: any) {
    uni.showToast({ title: e?.message || "支付发起失败", icon: "none" });
  } finally {
    isPaying.value = false;
  }
}

onLoad((opts) => {
  ensureTopInset();
  orderId.value = String((opts as any)?.id || "").trim();
  void load();
});

onMounted(() => ensureTopInset());
</script>
