<template>
  <view class="min-h-screen w-full bg-background-light font-display text-gray-900" style="padding-bottom: 110px;">
    <!-- 顶部悬浮导航 -->
    <view
      class="sticky top-0 z-50 bg-surface-95 px-4 pb-3 shadow-sm"
      :style="`padding-top:${topInset}px;`"
    >
      <view class="flex items-center">
        <view
          class="-ml-2 flex h-10 w-10 items-center justify-center rounded-full"
          hover-class="opacity-80"
          @tap="goBack"
        >
          <text class="text-lg font-black">←</text>
        </view>
        <view class="flex-1 pr-8 text-center text-base font-extrabold">确认订单</view>
      </view>
    </view>

    <view class="px-4 pt-4 space-y-4">
	      <!-- 地址 -->
	      <view v-if="!isVirtualOrder" class="relative overflow-hidden rounded-xl bg-white shadow-sm" hover-class="opacity-95" @tap="onSelectAddress">
	        <view class="p-4 flex items-center gap-3">
	          <view class="h-10 w-10 shrink-0 rounded-full bg-primary-10 flex items-center justify-center">
	            <text class="text-primary font-black">址</text>
	          </view>
	          <view class="flex-1 min-w-0">
	            <view class="flex items-end gap-2 mb-1">
	              <text class="text-base font-extrabold">{{ addressTitle }}</text>
	              <text class="text-xs text-muted font-semibold">{{ addressPhoneMasked }}</text>
	            </view>
	            <text class="text-sm text-muted" :number-of-lines="2">{{ addressFull }}</text>
	          </view>
	          <text class="text-muted">›</text>
	        </view>
        <view
          class="absolute bottom-0 left-0 right-0"
          style="height: 4px; background: repeating-linear-gradient(-45deg,#ef4444,#ef4444 12px,#ffffff 12px,#ffffff 24px,#3b82f6 24px,#3b82f6 36px,#ffffff 36px,#ffffff 48px); opacity:.6;"
        />
      </view>

      <!-- 商品清单 -->
      <view class="overflow-hidden rounded-xl bg-white shadow-sm">
        <view class="px-4 py-3 border-b" style="border-color: rgba(0,0,0,0.04);">
          <view class="flex items-center gap-2">
            <text class="text-xs text-muted">店铺</text>
            <text class="text-sm font-extrabold">官方自营店</text>
          </view>
        </view>

	        <view class="p-4 space-y-4">
	          <view v-for="it in items" :key="it.skuId" class="flex gap-3">
	            <image class="h-20 w-20 shrink-0 rounded-lg bg-gray-100" mode="aspectFill" :src="it.thumb" @error="onThumbError(it.skuId)" />
	            <view class="flex-1 flex flex-col justify-between min-w-0">
              <view>
                <text class="text-sm font-bold" :number-of-lines="2">{{ it.title }}</text>
                <view v-if="it.skuLabel" class="mt-1 flex flex-wrap gap-2">
                  <view class="rounded bg-gray-50 px-2 py-1">
                    <text class="text-xs text-muted">{{ it.skuLabel }}</text>
                  </view>
                </view>
              </view>
              <view class="mt-2 flex items-center justify-between">
                <text class="text-sm font-extrabold text-gray-900">{{ it.priceText }}</text>
                <text class="text-xs text-muted font-semibold">x{{ it.qty }}</text>
              </view>
            </view>
          </view>
        </view>

        <view class="px-4 pb-2">
          <view v-if="!isVirtualOrder" class="flex items-center justify-between py-3 border-t" style="border-color: rgba(0,0,0,0.04);">
            <text class="text-sm text-muted">配送方式</text>
            <view class="flex items-center gap-1" hover-class="opacity-80" @tap="noop">
              <text class="text-sm font-semibold">标准快递</text>
              <view class="ml-1 rounded bg-primary-10 px-2 py-1">
                <text class="text-xs font-extrabold text-primary">包邮</text>
              </view>
              <text class="text-muted">›</text>
            </view>
          </view>
          <view v-if="!isVirtualOrder" class="flex items-start gap-4 py-3 border-t" style="border-color: rgba(0,0,0,0.04);">
            <text class="text-sm text-muted shrink-0 mt-1">留言</text>
            <textarea
              v-model="buyerNote"
              class="flex-1 bg-transparent text-sm text-right"
              placeholder="给商家留言（可选）"
              :maxlength="200"
              auto-height
            />
          </view>
          <view class="flex items-center justify-end gap-2 py-3 border-t" style="border-color: rgba(0,0,0,0.04);">
            <text class="text-xs text-muted">{{ items.length }} 件</text>
            <text class="text-sm font-semibold">小计：<text class="text-base font-extrabold">{{ totalText }}</text></text>
          </view>
        </view>
      </view>

      <!-- 金额汇总 -->
      <view class="rounded-xl bg-white p-4 shadow-sm space-y-3">
        <view class="flex justify-between text-sm">
          <text class="text-muted">商品金额</text>
          <text class="font-semibold">{{ totalText }}</text>
        </view>
        <view v-if="!isVirtualOrder" class="flex justify-between text-sm">
          <text class="text-muted">运费</text>
          <text class="font-semibold">{{ shippingText }}</text>
        </view>
        <view class="flex justify-between text-sm">
          <text class="text-muted">优惠</text>
          <text class="text-primary font-semibold">{{ discountText }}</text>
        </view>
        <view v-if="promotionLoading" class="flex justify-between text-xs">
          <text class="text-muted">自动促销</text>
          <text class="text-muted">计算中...</text>
        </view>
        <view v-else-if="appliedPromotionNames" class="flex justify-between text-xs" style="gap: 12px;">
          <text class="text-muted shrink-0">已享活动</text>
          <text class="text-primary font-semibold text-right" :number-of-lines="2">{{ appliedPromotionNames }}</text>
        </view>
        <view v-else-if="promotionError" class="flex justify-between text-xs" style="gap: 12px;">
          <text class="text-muted shrink-0">自动促销</text>
          <text class="text-muted text-right" :number-of-lines="2">{{ promotionError }}</text>
        </view>
        <view class="pt-3 border-t flex justify-end" style="border-color: rgba(0,0,0,0.04);">
          <text class="text-sm font-extrabold">合计：{{ payableText }}</text>
        </view>
      </view>
    </view>

    <!-- 底部提交 -->
    <view class="fixed bottom-0 left-0 right-0 z-40 bg-white border-t" style="border-color: rgba(0,0,0,0.06);">
      <view class="px-4 py-2" :style="`padding-bottom: calc(env(safe-area-inset-bottom) + 8px);`">
        <view class="flex items-center justify-between">
          <view class="flex flex-col">
            <view class="flex items-baseline gap-1">
              <text class="text-xs text-muted font-semibold">合计：</text>
              <text class="text-xl font-extrabold text-primary">{{ payableText }}</text>
            </view>
            <text v-if="savedText" class="text-10 text-primary font-semibold" style="opacity: 0.8;">{{ savedText }}</text>
          </view>
	          <view
	            class="submit-btn flex items-center justify-center rounded-full bg-primary shadow-sm"
	            :style="submitting || !items.length ? 'opacity: 0.6;' : ''"
	            hover-class="opacity-90"
	            @tap="onSubmitTap"
	          >
	            <text class="text-sm font-extrabold text-white">{{ submitting ? "提交中..." : "提交订单" }}</text>
	          </view>
	        </view>
	      </view>
	    </view>
	  </view>
	</template>

<script setup lang="ts">
		import { computed, onMounted, ref } from "vue";
		import { onLoad, onShow } from "@dcloudio/uni-app";
		import { createOrder, quotePromotions, type PromotionSummary } from "@/services/miniapp-order";
		import { isLoggedIn } from "@/services/session";
		import { clearLocalCart, getLocalCart } from "@/services/cart";
		import { formatFullAddress, getSelectedAddressId, maskPhone, miniAppListMyAddresses, setSelectedAddressId, type MiniAppCustomerAddress } from "@/services/miniapp-address";
		import { isLikelyPlaceholderUrl, pickPlaceholderImage } from "@/utils/product-images";

type DraftItem = {
  skuId: string;
  spuId?: string;
  qty: number;
  title: string;
  skuLabel?: string;
  imageUrl?: string;
  currency?: string;
  unitPrice?: number;
};

type DraftPayload = {
  channel: string;
  locale: string;
  items: DraftItem[];
  from?: string;
};

	const topInset = ref(44);
		const submitting = ref(false);
		const buyerNote = ref("");

	const selectedAddress = ref<MiniAppCustomerAddress | null>(null);
	const addressTitle = computed(() => selectedAddress.value?.shippingAddress?.recipientName || "收货人（待选择）");
		const addressPhoneMasked = computed(() =>
		  selectedAddress.value ? maskPhone(selectedAddress.value.shippingAddress?.recipientPhone) : "—",
		);
		const addressFull = computed(() =>
		  selectedAddress.value ? formatFullAddress(selectedAddress.value.shippingAddress) : "请选择收货地址",
		);

	const draft = ref<DraftPayload | null>(null);
	const brokenThumbSkuIds = ref<Set<string>>(new Set());
const promotionQuote = ref<PromotionSummary | null>(null);
const promotionLoading = ref(false);
const promotionError = ref("");

function formatMoney(currency: string, amountMajor: number) {
  const c = String(currency || "CNY").trim().toUpperCase() || "CNY";
  const symbol = c === "CNY" || c === "RMB" ? "¥" : c === "USD" ? "$" : c === "EUR" ? "€" : `${c} `;
  const n = Number(amountMajor);
  if (!Number.isFinite(n)) return `${symbol}--`;
  const fixed = Math.round(n * 100) / 100;
  const text = Number.isInteger(fixed) ? String(fixed) : fixed.toFixed(2);
  return `${symbol}${text}`;
}

	function cartThumb(it: DraftItem) {
	  if (brokenThumbSkuIds.value.has(String(it.skuId || "").trim())) return "/static/logo.png";
	  const url = String(it.imageUrl || "").trim();
	  if (url && !isLikelyPlaceholderUrl(url)) return url;
	  const key = String(it.spuId || it.skuId || "").trim();
	  if (key) return pickPlaceholderImage(key);
	  return "/static/logo.png";
	}

const items = computed(() => {
  const list = draft.value?.items || [];
  return list.map((it) => {
    const currency = String(it.currency || "CNY").trim() || "CNY";
    const unit = Number(it.unitPrice ?? 0);
    const priceText = Number.isFinite(unit) && unit >= 0 ? formatMoney(currency, unit) : "—";
    return {
      skuId: it.skuId,
      title: it.title || "商品",
      skuLabel: it.skuLabel,
      qty: Number(it.qty || 0),
      thumb: cartThumb(it),
      priceText,
      currency,
      unitPrice: unit,
    };
  });
});

const currency = computed(() => {
  const c = new Set(items.value.map((x) => x.currency).filter(Boolean));
  return c.size === 1 ? Array.from(c)[0] : "CNY";
});

const totalAmount = computed(() => {
  const list = items.value;
  if (!list.length) return 0;
  if (new Set(list.map((x) => x.currency)).size > 1) return 0;
  return list.reduce((sum, it) => {
    if (!Number.isFinite(it.unitPrice) || it.unitPrice < 0) return sum;
    return sum + it.unitPrice * Number(it.qty || 0);
  }, 0);
});

const totalText = computed(() => formatMoney(currency.value, totalAmount.value));
const shippingText = computed(() => formatMoney(currency.value, 0));
const totalAmountMinor = computed(() => Math.max(0, Math.round(totalAmount.value * 100)));
const promotionDiscountMinor = computed(() => Number(promotionQuote.value?.promotion_discount_minor || 0));
const payableAmountMinor = computed(() => {
  const quoted = Number(promotionQuote.value?.after_promotion_total_minor || 0);
  return quoted > 0 || promotionDiscountMinor.value > 0 ? quoted : totalAmountMinor.value;
});
const discountText = computed(() => `-${formatMoneyFromMinor(currency.value, promotionDiscountMinor.value)}`);
const payableText = computed(() => formatMoneyFromMinor(currency.value, payableAmountMinor.value));
const savedText = computed(() => promotionDiscountMinor.value > 0 ? `已优惠 ${formatMoneyFromMinor(currency.value, promotionDiscountMinor.value)}` : "");
const appliedPromotionNames = computed(() => {
  const list = promotionQuote.value?.applied_promotions || [];
  return list.map((item) => item.name || item.code).filter(Boolean).join("、");
});
const isVirtualOrder = computed(() => String(draft.value?.from || "").trim() === "membership");

function formatMoneyFromMinor(currency: string, amountMinor: number) {
  return formatMoney(currency, Math.max(0, Number(amountMinor || 0)) / 100);
}

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

	function noop() {
	  uni.showToast({ title: "功能待接入", icon: "none" });
	}

		async function loadSelectedAddress() {
		  if (!isLoggedIn()) {
		    selectedAddress.value = null;
		    return;
		  }
		  try {
			    const list = await miniAppListMyAddresses();
			    const items = Array.isArray(list) ? list : [];
		    if (!items.length) {
		      selectedAddress.value = null;
		      return;
		    }
		    const selectedId = getSelectedAddressId();
		    const picked =
		      (selectedId ? items.find((x) => x.id === selectedId) : null) ||
		      items.find((x) => x.isDefault) ||
		      items[0] ||
		      null;
		    selectedAddress.value = picked;
		    if (picked?.id) setSelectedAddressId(picked.id);
		  } catch {
		    selectedAddress.value = null;
		  }
		}

		function onSelectAddress() {
		  uni.navigateTo({ url: "/pages/address/index?from=order" });
		}

	function onThumbError(skuId: string) {
	  const id = String(skuId || "").trim();
	  if (!id) return;
	  brokenThumbSkuIds.value.add(id);
	}

	function onSubmitTap() {
	  if (submitting.value || !items.value.length) return;
	  if (!isVirtualOrder.value && !selectedAddress.value) {
	    uni.showToast({ title: "请选择收货地址", icon: "none" });
	    return;
	  }
	  void submitOrder();
	}

function loadDraft(options?: any) {
  const key = String(options?.draftKey || "").trim();
  if (key) {
    try {
      const raw = String(uni.getStorageSync(key) || "").trim();
      if (raw) {
        draft.value = JSON.parse(raw) as DraftPayload;
        try {
          uni.removeStorageSync(key);
        } catch {}
        return;
      }
    } catch {}
  }
  // fallback: use local cart snapshot
  const cart = getLocalCart();
  const channel = String(uni.getStorageSync("miniapp.channel") || "official").trim() || "official";
  const locale = String(uni.getStorageSync("miniapp.locale") || "zh-CN").trim() || "zh-CN";
  draft.value = {
    channel,
    locale,
    items: (cart.items || []).map((x) => ({
      skuId: x.skuId,
      spuId: x.spuId,
      qty: x.qty,
      title: x.title || "商品",
      skuLabel: x.skuLabel,
      imageUrl: x.imageUrl,
      currency: x.currency,
      unitPrice: x.unitPrice,
    })),
    from: "cart",
  };
}

async function refreshPromotionQuote() {
  promotionQuote.value = null;
  promotionError.value = "";
  const payload = draft.value;
  if (!payload || !payload.items?.length || isVirtualOrder.value) return;
  if (!isLoggedIn()) return;
  const quoteItems = items.value
    .map((item) => ({
      line_id: item.skuId,
      sku_id: item.skuId,
      qty: Number(item.qty || 0),
      unit_price_minor: Math.round(Number(item.unitPrice || 0) * 100),
    }))
    .filter((item) => item.sku_id && item.qty > 0 && item.unit_price_minor >= 0);
  if (!quoteItems.length) return;
  promotionLoading.value = true;
  try {
    promotionQuote.value = await quotePromotions({
      channel: payload.channel,
      currency: currency.value,
      items: quoteItems,
    });
  } catch (error: any) {
    promotionError.value = error?.message || "暂未匹配可用活动";
  } finally {
    promotionLoading.value = false;
  }
}

		async function submitOrder() {
		  if (submitting.value) return;
		  if (!items.value.length) return;
		  if (!isVirtualOrder.value && !selectedAddress.value) {
		    uni.showToast({ title: "请选择收货地址", icon: "none" });
		    return;
		  }
	  if (!isLoggedIn()) {
	    uni.showToast({ title: "请先登录", icon: "none" });
	    uni.navigateTo({ url: "/pages/auth/index?tab=login" });
	    return;
  }
  submitting.value = true;
  try {
    const payload = draft.value;
    if (!payload) throw new Error("缺少订单信息");
    const idem = `order-${Date.now()}-${Math.random().toString(16).slice(2)}`;
	    const orderPayload: any = {
	      channel: payload.channel,
	      locale: payload.locale,
	      items: payload.items.map((x) => ({ skuId: x.skuId, qty: x.qty })),
	    };
	    if (!isVirtualOrder.value && selectedAddress.value?.id) {
	      orderPayload.shippingAddressId = selectedAddress.value.id;
	    }
	    const order = await createOrder(orderPayload, idem);
    if (payload.from === "cart") {
      clearLocalCart();
    }
    const q = `orderNo=${encodeURIComponent(order.orderNo)}&orderId=${encodeURIComponent(order.orderId)}&currency=${encodeURIComponent(
      order.amounts.currency,
    )}&total=${encodeURIComponent(String(order.amounts.total))}&subtotal=${encodeURIComponent(String(order.amounts.subtotal))}`;
    uni.redirectTo({ url: `/pages/order/success?${q}` });
  } catch (e: any) {
    uni.showToast({ title: e?.message || "下单失败，请稍后重试", icon: "none" });
  } finally {
    submitting.value = false;
  }
}

		onLoad((options) => {
		  ensureTopInset();
		  loadDraft(options);
		  void refreshPromotionQuote();
		  if (!isVirtualOrder.value) {
		    void loadSelectedAddress();
		  }
		});

	onMounted(() => {
	  ensureTopInset();
	  void refreshPromotionQuote();
	});

		onShow(() => {
		  if (!isVirtualOrder.value) {
		    void loadSelectedAddress();
		  }
		});
		</script>

	<style scoped>
	.submit-btn {
	  height: 40px;
	  padding-left: 18px;
	  padding-right: 18px;
	  min-width: 124px;
	}
	</style>
