<template>
  <view class="min-h-screen w-full bg-background-light font-display text-text-dark" style="padding-bottom: 160px;">
    <view
      class="sticky top-0 z-50 border-b bg-surface-95 px-4 py-3 shadow-sm"
      style="border-color: rgba(0,0,0,0.04); backdrop-filter: blur(12px);"
      :style="`padding-top:${topInset}px;`"
    >
      <view class="flex items-center justify-between">
        <view class="text-base font-extrabold text-gray-900">购物车 ({{ totalCount }})</view>
        <view class="flex items-center gap-3">
          <view v-if="items.length" class="text-sm font-semibold text-muted" hover-class="opacity-70" @tap="toggleManage">
            {{ manageMode ? "完成" : "管理" }}
          </view>
          <view v-if="items.length" class="text-sm font-semibold text-red-500" hover-class="opacity-70" @tap="clearCart()">清空</view>
        </view>
      </view>
    </view>

    <view v-if="!items.length" class="px-4 py-10">
      <view class="rounded-2xl bg-white p-6 text-center shadow-sm">
        <view class="text-base font-extrabold">购物车空空如也</view>
        <view class="mt-2 text-sm text-muted">去商城挑选一些商品吧</view>
        <view class="mt-4">
          <button class="rounded-full bg-primary px-6 py-2 text-sm font-extrabold text-white" @tap="toMall">去逛逛</button>
        </view>
      </view>
    </view>

    <view v-else class="px-4 pt-4">
      <view class="mb-3 flex items-center gap-2 px-1">
        <view class="rounded-lg bg-white p-2 shadow-sm" style="border: 1px solid rgba(0,0,0,0.04);">
          <text class="text-primary font-black" style="font-size: 14px;">店</text>
        </view>
        <text class="text-sm font-extrabold">官方自营店</text>
        <text class="text-muted">›</text>
      </view>

      <view class="overflow-hidden rounded-2xl bg-white shadow-sm" style="border: 1px solid rgba(0,0,0,0.04);">
        <view
          v-for="(it, idx) in items"
          :key="it.skuId"
          class="flex gap-3 px-4 py-4"
          :style="idx === items.length - 1 ? '' : 'border-bottom: 1px solid rgba(0,0,0,0.04);'"
        >
          <view class="flex items-center justify-center pt-1">
            <view
              class="h-5 w-5 rounded-full border flex items-center justify-center"
              :style="checkboxStyle(it)"
              hover-class="opacity-80"
              @tap="toggleItemSelected(it.skuId)"
            >
              <view v-if="isItemSelectable(it) && isItemSelected(it)" class="h-3 w-3 rounded-full bg-primary"></view>
            </view>
          </view>

          <image class="h-20 w-20 shrink-0 rounded-xl bg-gray-100" mode="aspectFill" :src="cartThumb(it)" @error="onCartThumbError(it.skuId)" />

          <view class="min-w-0 flex-1">
            <view class="flex items-start justify-between gap-2">
              <view class="min-w-0 flex-1">
                <view class="flex items-center gap-2">
                  <text
                    class="text-sm font-extrabold"
                    :style="isItemSelectable(it) ? 'color:#111827;' : 'color: rgba(17, 24, 39, 0.45);'"
                    :number-of-lines="2"
                  >
                    {{ it.title || "商品" }}
                  </text>
                  <view v-if="itemBadges.get(it.skuId)" class="rounded-md bg-gray-100 px-2" style="padding-top: 2px; padding-bottom: 2px;">
                    <text class="text-10 font-extrabold text-gray-500">{{ itemBadges.get(it.skuId) }}</text>
                  </view>
                </view>
                <view
                  v-if="it.skuLabel"
                  class="mt-2 inline-flex items-center rounded-lg bg-gray-50 px-3 py-1 text-xs font-semibold text-muted"
                  style="border: 1px solid rgba(0,0,0,0.04);"
                  hover-class="opacity-80"
                  @tap="onOpenSkuPicker(it)"
                >
                  <text :number-of-lines="1">{{ it.skuLabel }}</text>
                  <text class="ml-1 text-muted">▾</text>
                </view>
              </view>

              <view class="-mr-1 -mt-1 flex items-center">
                <view class="p-1" hover-class="opacity-70" @tap="remove(it.skuId)">
                  <image class="h-5 w-5" mode="aspectFit" src="/static/icons/trash-red.svg" />
                </view>
              </view>
            </view>

            <view class="mt-3 flex items-end justify-between">
              <view class="flex items-baseline gap-1">
                <text class="text-xs font-extrabold" :style="isItemSelectable(it) ? 'color:#386657;' : 'color: rgba(17, 24, 39, 0.35);'">¥</text>
                <text class="text-lg font-extrabold" :style="isItemSelectable(it) ? 'color:#386657;' : 'color: rgba(17, 24, 39, 0.35);'">
                  {{ formatUnitPrice(it) }}
                </text>
              </view>

              <view
                class="flex h-8 items-center rounded-lg bg-white shadow-sm"
                :style="isItemSelectable(it) ? 'border: 1px solid rgba(0,0,0,0.10);' : 'border: 1px solid rgba(0,0,0,0.06); background: rgba(249, 250, 251, 1); opacity: 0.7;'"
              >
                <view
                  class="w-8 h-full flex items-center justify-center"
                  :style="isItemSelectable(it) ? '' : 'opacity: 0.5;'"
                  hover-class="opacity-70"
                  @tap="decrease(it.skuId)"
                >
                  <text class="text-muted font-black">−</text>
                </view>
                <view class="w-8 h-full flex items-center justify-center">
                  <text class="text-sm font-extrabold text-gray-700">{{ it.qty }}</text>
                </view>
                <view
                  class="w-8 h-full flex items-center justify-center"
                  :style="isItemSelectable(it) ? '' : 'opacity: 0.5;'"
                  hover-class="opacity-70"
                  @tap="increase(it.skuId)"
                >
                  <text class="text-muted font-black">+</text>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>
    </view>

    <view class="cart-bottom-bar fixed left-0 right-0 z-20 bg-white px-4 py-3">
      <view class="flex items-center justify-between" style="gap: 12px;">
        <view class="flex items-center gap-2" hover-class="opacity-80" @tap="toggleSelectAll">
          <view class="h-5 w-5 rounded-full border flex items-center justify-center" :style="selectAllStyle">
            <view v-if="allSelected" class="h-3 w-3 rounded-full bg-primary"></view>
          </view>
          <text class="text-xs font-semibold text-muted">全选</text>
        </view>

        <view class="ml-auto flex items-center" style="gap: 12px;">
          <view class="flex flex-col items-end">
            <view class="flex items-baseline gap-1">
              <text class="text-xs text-muted font-semibold">合计:</text>
              <text class="text-base font-extrabold text-gray-900">{{ totalPriceText }}</text>
            </view>
            <text v-if="hasSelectedPrice" class="text-10 text-primary font-semibold" style="opacity: 0.8;">已省 ¥0.00</text>
          </view>
          <view
            class="checkout-btn flex items-center justify-center rounded-full bg-primary shadow-sm"
            :style="checkoutDisabled ? 'opacity: 0.5;' : ''"
            hover-class="opacity-90"
            @tap="onCheckoutTap"
          >
            <text class="text-sm font-extrabold text-white">去结算 ({{ selectedCount }})</text>
          </view>
        </view>
      </view>
    </view>

    <SkuPickerSheet
      v-model="skuSheetOpen"
      mode="cart"
      :show-price="false"
      :show-qty="false"
      v-model:selectedSkuId="skuPickerSelectedSkuId"
      :skus="skuPickerSkus"
      :spec-groups="skuPickerSpecGroups"
      :cover-url="skuPickerCoverUrl"
      :max-qty="10"
      @pick="onSkuPicked"
    />
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { syncTabBarSelected } from "@/utils/tabbar";
import { cartCount, clearLocalCart, getLocalCart, removeCartItem, setCartItemQty, setLocalCart, type LocalCartItem, type LocalCart } from "@/services/cart";
import { isLoggedIn } from "@/services/session";
import { miniAppBatchSkus } from "@/services/miniapp-sku";
import SkuPickerSheet from "@/components/product/sku-picker-sheet.vue";
import { miniAppGetProduct, miniAppGetProductDetailWithSpec, miniAppGetSellability, type MiniAppSpecGroup, type MiniAppSkuSummary } from "@/services/miniapp-product";
import { isLikelyPlaceholderUrl, pickPlaceholderImage } from "@/utils/product-images";

type SellabilityMeta = {
  sellable: boolean;
  reasons: string[];
  availableQty: number;
  price?: { amount: number; currency: string } | null;
};

const topInset = ref<number>(40);
const items = ref<LocalCartItem[]>([]);
const fallbackThumb = "/static/icons/image-placeholder.png";
const enriching = ref(false);
const manageMode = ref(false);
const sellabilityBySkuId = ref<Map<string, SellabilityMeta>>(new Map());
const brokenThumbSkuIds = ref<Set<string>>(new Set());

const skuSheetOpen = ref(false);
const skuPickerSpuId = ref("");
const skuPickerOldSkuId = ref("");
const skuPickerSelectedSkuId = ref("");
const skuPickerSkus = ref<MiniAppSkuSummary[]>([]);
const skuPickerSpecGroups = ref<MiniAppSpecGroup[]>([]);
const skuPickerCoverUrl = ref("");
const spuDetailCache = new Map<string, { skus: MiniAppSkuSummary[]; specGroups: MiniAppSpecGroup[]; coverUrl: string }>();
const spuCoverCache = new Map<string, string>();

const totalCount = computed(() => cartCount(items.value));

function formatMoney(currency: string, amount: number) {
  const c = String(currency || "CNY").trim().toUpperCase() || "CNY";
  const symbol = c === "CNY" || c === "RMB" ? "¥" : c === "USD" ? "$" : c === "EUR" ? "€" : `${c} `;
  const n = Number(amount);
  if (!Number.isFinite(n)) return `${symbol}--`;
  const fixed = Math.round(n * 100) / 100;
  const text = Number.isInteger(fixed) ? String(fixed) : fixed.toFixed(2);
  return `${symbol}${text}`;
}

function getMeta(it: LocalCartItem): SellabilityMeta | null {
  const m = sellabilityBySkuId.value.get(String(it.skuId || "").trim());
  return m || null;
}

function isItemSelectable(it: LocalCartItem) {
  const m = getMeta(it);
  if (!m) return true;
  return Boolean(m.sellable) && Number(m.availableQty || 0) > 0;
}

function isItemSelected(it: LocalCartItem) {
  return Boolean(it.selected !== false);
}

const selectableItems = computed(() => (items.value || []).filter((it) => isItemSelectable(it)));
const selectedItems = computed(() => selectableItems.value.filter((it) => isItemSelected(it)));
const selectedCount = computed(() => cartCount(selectedItems.value));

const hasSelectedPrice = computed(() => {
  if (!selectedItems.value.length) return false;
  return selectedItems.value.every((it) => {
    const price = Number(it.unitPrice);
    const currency = String(it.currency || "").trim();
    return Number.isFinite(price) && price >= 0 && Boolean(currency);
  });
});

const totalPriceText = computed(() => {
  if (!selectedItems.value.length) return formatMoney("CNY", 0);
  if (!hasSelectedPrice.value) return "—";
  const currencies = new Set<string>();
  let currency = "CNY";
  const sum = selectedItems.value.reduce((acc, it) => {
    const price = Number(it.unitPrice);
    const c = String(it.currency || "").trim();
    if (c) currencies.add(c);
    if (currencies.size > 1) return acc;
    if (c) currency = c;
    if (!Number.isFinite(price) || price < 0) return acc;
    return acc + price * Number(it.qty || 0);
  }, 0);
  if (currencies.size > 1) return "—";
  if (!Number.isFinite(sum) || sum <= 0) return formatMoney(currency, 0);
  return formatMoney(currency, sum);
});

const allSelected = computed(() => {
  if (!selectableItems.value.length) return false;
  return selectableItems.value.every((it) => isItemSelected(it));
});

const checkoutDisabled = computed(() => selectedItems.value.length === 0 || !hasSelectedPrice.value);

function checkboxStyle(it: LocalCartItem) {
  if (!isItemSelectable(it)) return "border-color: rgba(229, 231, 235, 1); background: rgba(243, 244, 246, 1);";
  if (isItemSelected(it)) return "border-color: rgba(56,102,87,0.65); background: rgba(56,102,87,0.10);";
  return "border-color: rgba(209, 213, 219, 1); background: #fff;";
}

const selectAllStyle = computed(() => {
  if (!selectableItems.value.length) return "border-color: rgba(229, 231, 235, 1); background: rgba(243, 244, 246, 1);";
  return allSelected.value
    ? "border-color: rgba(56,102,87,0.65); background: rgba(56,102,87,0.10);"
    : "border-color: rgba(209, 213, 219, 1); background: #fff;";
});

function sellabilityReasonToBadge(reasons: string[]) {
  const code = String((reasons || [])[0] || "").toUpperCase();
  switch (code) {
    case "OUT_OF_STOCK":
      return "补货中";
    case "NO_PUBLIC_PRICE":
      return "暂无价格";
    case "SKU_NOT_ONLINE":
      return "已下架";
    case "CHANNEL_DISABLED":
    case "NOT_IN_AVAILABILITY_WINDOW":
    case "CHANNEL_STATUS_BLOCKED":
      return "暂不可售";
    default:
      return "不可售";
  }
}

const itemBadges = computed(() => {
  const out = new Map<string, string>();
  sellabilityBySkuId.value.forEach((m, skuId) => {
    if (m.sellable && Number(m.availableQty || 0) > 0) return;
    out.set(skuId, sellabilityReasonToBadge(m.reasons || []));
  });
  return out;
});

const AUTH_REDIRECT_KEY = "miniapp.auth.redirect";
function goLogin(redirectUrl: string) {
  uni.setStorageSync(AUTH_REDIRECT_KEY, redirectUrl);
  uni.navigateTo({ url: "/pages/auth/index?tab=login" });
}

async function ensureLoggedIn(): Promise<boolean> {
  if (isLoggedIn()) return true;
  const ok = await new Promise<boolean>((resolve) => {
    uni.showModal({
      title: "需要登录",
      content: "登录后才能继续结算。",
      confirmText: "去登录",
      cancelText: "再看看",
      success: (res) => resolve(Boolean(res?.confirm)),
      fail: () => resolve(false),
    });
  });
  if (ok) goLogin("/pages/cart/index");
  return false;
}

function isUnauthorizedError(err: any) {
  const msg = String(err?.message || err || "").toLowerCase();
  return msg.includes("unauthorized") || msg.includes("401") || msg.includes("token") || msg.includes("not logged") || msg.includes("请先登录");
}

function loadLocal() {
  const cart: LocalCart = getLocalCart();
  items.value = cart.items || [];
}

function cartThumb(it: LocalCartItem) {
  if (brokenThumbSkuIds.value.has(String(it?.skuId || "").trim())) return fallbackThumb;
  const url = String(it?.imageUrl || "").trim();
  // 过滤已知“占位域名”（如 picsum），统一回退到与商城一致的占位图池
  if (url && !isLikelyPlaceholderUrl(url)) return url;
  const key = String(it?.spuId || it?.skuId || "").trim();
  if (key) return pickPlaceholderImage(key);
  return fallbackThumb;
}

function formatUnitPrice(it: LocalCartItem) {
  const price = Number(it.unitPrice);
  if (!Number.isFinite(price) || price < 0) return "--";
  const fixed = Math.round(price * 100) / 100;
  return Number.isInteger(fixed) ? String(fixed) : fixed.toFixed(2);
}

function toggleManage() {
  manageMode.value = !manageMode.value;
}

function toggleItemSelected(skuId: string) {
  const id = String(skuId || "").trim();
  if (!id) return;
  const next = (items.value || []).map((it) => {
    if (it.skuId !== id) return it;
    if (!isItemSelectable(it)) return it;
    return { ...it, selected: !(it.selected !== false) };
  });
  items.value = next;
  const cart = getLocalCart();
  setLocalCart(next, cart.updatedAt);
}

function toggleSelectAll() {
  const shouldSelect = !allSelected.value;
  const next = (items.value || []).map((it) => {
    if (!isItemSelectable(it)) return it;
    return { ...it, selected: shouldSelect };
  });
  items.value = next;
  const cart = getLocalCart();
  setLocalCart(next, cart.updatedAt);
}

function ensureDeselectedInvalid() {
  const map = sellabilityBySkuId.value;
  if (!map.size) return;
  const next = (items.value || []).map((it) => {
    const m = map.get(String(it.skuId || "").trim());
    if (!m) return it;
    if (!m.sellable || Number(m.availableQty || 0) <= 0) {
      if (it.selected === false) return it;
      return { ...it, selected: false };
    }
    return it;
  });
  if (next.every((x, i) => x.selected === items.value[i]?.selected)) return;
  items.value = next;
  const cart = getLocalCart();
  setLocalCart(next, cart.updatedAt);
}

function onOpenSkuPicker(_it: LocalCartItem) {
  void openSkuPicker(_it);
}

async function openSkuPicker(it: LocalCartItem) {
  const spuId = String(it?.spuId || "").trim();
  if (!spuId) {
    uni.showToast({ title: "商品信息加载中，请稍后再试", icon: "none" });
    return;
  }
  skuPickerSpuId.value = spuId;
  skuPickerOldSkuId.value = String(it?.skuId || "").trim();
  skuPickerSelectedSkuId.value = String(it?.skuId || "").trim();

  const cached = spuDetailCache.get(spuId);
  if (cached) {
    skuPickerSkus.value = cached.skus;
    skuPickerSpecGroups.value = cached.specGroups;
    skuPickerCoverUrl.value = cached.coverUrl;
    skuSheetOpen.value = true;
    return;
  }

  try {
    const detail = await miniAppGetProductDetailWithSpec(spuId);
    const skus = Array.isArray(detail?.skus) ? detail.skus : [];
    const groups = Array.isArray(detail?.spec?.groups) ? detail.spec!.groups : [];
    const coverUrl = String(detail?.spu?.coverUrl || "").trim() || "/static/icons/image-placeholder.svg";
    spuDetailCache.set(spuId, { skus, specGroups: groups, coverUrl });
    skuPickerSkus.value = skus;
    skuPickerSpecGroups.value = groups;
    skuPickerCoverUrl.value = coverUrl;
    skuSheetOpen.value = true;
  } catch (e: any) {
    uni.showToast({ title: e?.message || "加载规格失败", icon: "none" });
  }
}

function buildSkuLabel(sku: MiniAppSkuSummary | undefined | null) {
  if (!sku) return "";
  return String((sku as any)?.code || "").trim() || String(sku.id || "").trim();
}

function onSkuPicked(payload: { skuId: string }) {
  const newSkuId = String(payload?.skuId || "").trim();
  const oldSkuId = String(skuPickerOldSkuId.value || "").trim();
  if (!newSkuId || !oldSkuId) return;
  if (newSkuId === oldSkuId) return;

  const current = getLocalCart();
  const oldItem = (current.items || []).find((x) => x.skuId === oldSkuId);
  if (!oldItem) return;

  const spuId = String(skuPickerSpuId.value || oldItem.spuId || "").trim() || undefined;
  const cached = spuDetailCache.get(String(spuId || "").trim());
  const pickedSku = cached?.skus?.find((s) => String(s.id || "").trim() === newSkuId) || skuPickerSkus.value.find((s) => String(s.id || "").trim() === newSkuId);

  const nextItems: LocalCartItem[] = [];
  for (const it of current.items || []) {
    if (it.skuId === oldSkuId) continue;
    nextItems.push({ ...it });
  }

  const exist = nextItems.find((x) => x.skuId === newSkuId);
  if (exist) {
    exist.qty = Number(exist.qty || 0) + Number(oldItem.qty || 0);
    exist.selected = (oldItem.selected !== false) || (exist.selected !== false);
    exist.spuId = exist.spuId || spuId;
    exist.title = exist.title || oldItem.title;
    exist.imageUrl = exist.imageUrl || oldItem.imageUrl;
    exist.skuCode = exist.skuCode || String((pickedSku as any)?.code || "").trim() || oldItem.skuCode;
    exist.skuLabel = exist.skuLabel || buildSkuLabel(pickedSku) || oldItem.skuLabel;
    exist.currency = exist.currency || String(pickedSku?.currency || "").trim() || oldItem.currency;
    exist.unitPrice = Number.isFinite(Number(exist.unitPrice)) ? exist.unitPrice : (pickedSku?.price ?? oldItem.unitPrice);
  } else {
    nextItems.push({
      skuId: newSkuId,
      spuId,
      qty: Number(oldItem.qty || 1),
      selected: oldItem.selected !== false,
      title: oldItem.title,
      imageUrl: oldItem.imageUrl,
      skuCode: String((pickedSku as any)?.code || "").trim() || oldItem.skuCode,
      skuLabel: buildSkuLabel(pickedSku) || oldItem.skuLabel,
      currency: String(pickedSku?.currency || "").trim() || oldItem.currency,
      unitPrice: pickedSku?.price ?? oldItem.unitPrice,
      maxQty: oldItem.maxQty,
    });
  }

  setLocalCart(nextItems, current.updatedAt);
  loadLocal();
  void enrichItemsFromServer().then(() => ensureDeselectedInvalid());
}

function onCartThumbError(skuId: string) {
  const id = String(skuId || "").trim();
  if (!id) return;
  brokenThumbSkuIds.value.add(id);
}

async function enrichItemsFromServer() {
  if (enriching.value) return;
  const current = items.value || [];
  if (!current.length) {
    sellabilityBySkuId.value = new Map();
    return;
  }

  enriching.value = true;
  try {
    const resp = await miniAppBatchSkus(current.map((x) => x.skuId));
    const map = new Map((resp?.items || []).map((x) => [String(x.id || "").trim(), x]));
    // 购物车里如果 sku batch 没返回 imageUrl，则用商品 coverUrl（与商城列表一致）
    const spuIdsForCover = Array.from(
      new Set(
        current
          .map((it) => String(it?.spuId || "").trim())
          .concat((resp?.items || []).map((x) => String((x as any)?.spuId || "").trim()))
          .filter(Boolean),
      ),
    );
    await Promise.all(
      spuIdsForCover.map(async (spuId) => {
        if (spuCoverCache.has(spuId)) return;
        try {
          const p = await miniAppGetProduct(spuId);
          const coverUrl = String((p as any)?.coverUrl || "").trim();
          if (coverUrl) spuCoverCache.set(spuId, coverUrl);
        } catch {
          // ignore
        }
      }),
    );

    const next = current.map((it) => {
      const s = map.get(String(it.skuId || "").trim());
      if (!s) return it;
      const unitPrice = Number.isFinite(Number(it.unitPrice)) && Number(it.unitPrice) >= 0 ? it.unitPrice : (s.price ?? it.unitPrice);
      const currency = String(it.currency || "").trim() || String(s.currency || "").trim() || undefined;
      const spuId = String(it.spuId || "").trim() || String(s.spuId || "").trim() || undefined;
      const localImage = String(it.imageUrl || "").trim();
      const remoteImage = String(s.imageUrl || "").trim();
      const coverImage = spuId ? String(spuCoverCache.get(spuId) || "").trim() : "";
      // 与商城列表一致：优先用商品 coverUrl（而不是 sku 的 imageUrl），避免出现“购物车与商城图不一致”
      const imageUrl =
        (coverImage && !isLikelyPlaceholderUrl(coverImage) ? coverImage : "") ||
        (remoteImage && !isLikelyPlaceholderUrl(remoteImage) ? remoteImage : "") ||
        (localImage && !isLikelyPlaceholderUrl(localImage) ? localImage : "") ||
        undefined;
      const title = String(it.title || "").trim() || String(s.spuName || "").trim() || undefined;
      const skuCode = String(it.skuCode || "").trim() || String(s.code || "").trim() || undefined;
      const nextItem = { ...it, spuId, unitPrice, currency, imageUrl, title, skuCode };
      if (imageUrl && imageUrl !== localImage) brokenThumbSkuIds.value.delete(String(it?.skuId || "").trim());
      return nextItem;
    });
    items.value = next;
    const cart = getLocalCart();
    setLocalCart(next, cart.updatedAt);

    const channel = String(uni.getStorageSync("miniapp.channel") || "official").trim() || "official";
    const locale = String(uni.getStorageSync("miniapp.locale") || "zh-CN").trim() || "zh-CN";
    const spuIds = Array.from(new Set(next.map((x) => String(x.spuId || "").trim()).filter(Boolean)));

    const nextMap = new Map<string, SellabilityMeta>();
    await Promise.all(
      spuIds.map(async (spuId) => {
        try {
          const r = await miniAppGetSellability(spuId, { channel, locale });
          (r?.items || []).forEach((it: any) => {
            const skuId = String(it?.skuId || "").trim();
            if (!skuId) return;
            nextMap.set(skuId, {
              sellable: Boolean(it?.sellable),
              reasons: Array.isArray(it?.reasons) ? it.reasons : [],
              availableQty: Number.isFinite(Number(it?.availableQty)) ? Math.max(0, Number(it.availableQty)) : 0,
              price: it?.price ?? null,
            });
          });
        } catch {
        }
      }),
    );
    sellabilityBySkuId.value = nextMap;

    const patched = (items.value || []).map((it) => {
      const m = nextMap.get(String(it.skuId || "").trim());
      if (!m) return it;
      if (!m.sellable || m.availableQty <= 0) return it.selected === false ? it : { ...it, selected: false };
      const maxByStock = Number(m.availableQty);
      if (Number.isFinite(maxByStock) && maxByStock > 0 && it.qty > maxByStock) {
        return { ...it, qty: maxByStock };
      }
      return it;
    });
    const changed = patched.some((x, i) => x.qty !== (items.value[i]?.qty ?? x.qty) || x.selected !== (items.value[i]?.selected ?? x.selected));
    if (changed) {
      items.value = patched;
      const cart2 = getLocalCart();
      setLocalCart(patched, cart2.updatedAt);
      const clampedAny = patched.some((x, i) => x.qty !== (next[i]?.qty ?? x.qty));
      if (clampedAny) uni.showToast({ title: "部分商品库存不足，已调整数量", icon: "none" });
    }
  } catch {
  } finally {
    enriching.value = false;
  }
}

onMounted(() => {
  try {
    const wxAny = (globalThis as any).wx;
    if (wxAny && typeof wxAny.getWindowInfo === "function") {
      const info = wxAny.getWindowInfo();
      const statusBar = Number(info?.statusBarHeight ?? 0);
      topInset.value = Math.max(40, statusBar + 12);
      return;
    }
    const sys = uni.getSystemInfoSync();
    const statusBar = Number((sys as any).statusBarHeight ?? 0);
    topInset.value = Math.max(40, statusBar + 12);
  } catch {
    topInset.value = 40;
  }
});

onShow(() => {
  syncTabBarSelected("pages/cart/index");
  loadLocal();
  void enrichItemsFromServer().then(() => ensureDeselectedInvalid());
});

function toMall() {
  uni.switchTab({ url: "/pages/mall/index" });
}

function increase(skuId: string) {
  const it = items.value.find((x) => x.skuId === skuId);
  if (it && !isItemSelectable(it)) return;
  const max = Number(it?.maxQty);
  const maxQty = Number.isFinite(max) && max > 0 ? Math.floor(max) : 10;
  const next = (it?.qty || 0) + 1;
  const m = it ? getMeta(it) : null;
  const stockMax = m && Number.isFinite(Number(m.availableQty)) && Number(m.availableQty) > 0 ? Number(m.availableQty) : undefined;
  const effectiveMax = stockMax ? Math.min(maxQty, stockMax) : maxQty;
  if (next > effectiveMax) {
    uni.showToast({ title: `最多 ${effectiveMax} 件`, icon: "none" });
    return;
  }
  setCartItemQty(skuId, next);
  loadLocal();
}

function decrease(skuId: string) {
  const it = items.value.find((x) => x.skuId === skuId);
  if (it && !isItemSelectable(it)) return;
  const next = (it?.qty || 0) - 1;
  if (next <= 0) {
    remove(skuId);
    return;
  }
  setCartItemQty(skuId, next);
  loadLocal();
}

function remove(skuId: string) {
  removeCartItem(skuId);
  loadLocal();
  void enrichItemsFromServer();
}

async function clearCart(confirm = true) {
  if (confirm) {
    const ok = await new Promise<boolean>((resolve) => {
      uni.showModal({
        title: "清空购物车",
        content: "确认移除购物车内的所有商品？",
        confirmText: "清空",
        confirmColor: "#ef4444",
        success: (res) => resolve(Boolean(res?.confirm)),
        fail: () => resolve(false),
      });
    });
    if (!ok) return;
  }
  clearLocalCart();
  loadLocal();
  uni.showToast({ title: "已清空", icon: "none" });
}

async function checkout() {
  if (!selectedItems.value.length) {
    uni.showToast({ title: "请选择要结算的商品", icon: "none" });
    return;
  }
  if (!(await ensureLoggedIn())) return;
  try {
    const channel = String(uni.getStorageSync("miniapp.channel") || "official").trim() || "official";
    const locale = String(uni.getStorageSync("miniapp.locale") || "zh-CN").trim() || "zh-CN";
    const draftKey = `miniapp.checkout.draft.${Date.now()}-${Math.random().toString(16).slice(2)}`;
    const payload = {
      channel,
      locale,
      from: "cart",
      items: selectedItems.value.map((x) => ({
        skuId: x.skuId,
        spuId: x.spuId,
        qty: x.qty,
        title: x.title,
        skuLabel: x.skuLabel,
        imageUrl: x.imageUrl,
        currency: x.currency,
        unitPrice: x.unitPrice,
      })),
    };
    uni.setStorageSync(draftKey, JSON.stringify(payload));
    uni.navigateTo({ url: `/pages/order/confirm?draftKey=${encodeURIComponent(draftKey)}` });
  } catch (e: any) {
    if (isUnauthorizedError(e)) {
      await ensureLoggedIn();
    } else {
      uni.showToast({ title: e?.message || "下单失败，请稍后重试", icon: "none" });
    }
  }
}

function onCheckoutTap() {
  if (checkoutDisabled.value) return;
  void checkout();
}
</script>

<style scoped>
.cart-bottom-bar {
  box-shadow: 0 -10px 30px rgba(0, 0, 0, 0.06);
  bottom: calc(env(safe-area-inset-bottom) + 64px);
}

.checkout-btn {
  height: 40px;
  padding-left: 18px;
  padding-right: 18px;
  min-width: 124px;
}
</style>
