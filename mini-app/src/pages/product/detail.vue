<template>
  <view class="bg-background-light font-display text-gray-900 antialiased">
    <view class="relative mx-auto flex min-h-screen w-full max-w-md flex-col overflow-hidden bg-background-light">
      <!-- 顶部悬浮导航 -->
      <view
        class="fixed left-0 right-0 top-0 z-50 mx-auto flex max-w-md items-center justify-between px-4"
        :style="`padding-top:${topInset}px; padding-bottom: 12px; background: linear-gradient(to bottom, rgba(0,0,0,0.35), rgba(0,0,0,0));`"
      >
        <view
          class="flex h-10 w-10 items-center justify-center rounded-full"
          style="background: rgba(255,255,255,0.82);"
          hover-class="opacity-90"
          @tap="goBack"
        >
          <text style="font-size: 16px; font-weight: 900; color: #1f2937;">←</text>
        </view>
        <view class="flex items-center" style="gap: 12px;">
          <view
            class="flex h-10 w-10 items-center justify-center rounded-full"
            style="background: rgba(255,255,255,0.82);"
            hover-class="opacity-90"
            @tap="onShare"
          >
            <text style="font-size: 12px; font-weight: 900; color: #1f2937;">分享</text>
          </view>
          <view
            class="flex h-10 w-10 items-center justify-center rounded-full"
            style="background: rgba(255,255,255,0.82);"
            hover-class="opacity-90"
            @tap="onMore"
          >
            <text style="font-size: 12px; font-weight: 900; color: #1f2937;">···</text>
          </view>
        </view>
      </view>

      <!-- 页面内容（使用页面原生滚动） -->
      <view class="pb-32">
        <!-- 轮播图 -->
        <view class="relative w-full" style="aspect-ratio: 4 / 5; background: #e5e7eb;">
          <swiper
            class="h-full w-full"
            :circular="gallery.length > 1"
            :autoplay="false"
            :indicator-dots="false"
            :current="galleryIndex"
            @change="onGalleryChange"
          >
            <swiper-item v-for="(img, idx) in gallery" :key="`${img}-${idx}`">
              <image class="h-full w-full" mode="aspectFill" :src="img" />
            </swiper-item>
          </swiper>
          <view
            v-if="gallery.length > 1"
            class="absolute bottom-6 flex items-center rounded-full px-3 py-1"
            style="left: 50%; transform: translateX(-50%); background: rgba(0,0,0,0.18);"
          >
            <view v-for="(_, idx) in gallery" :key="idx" :style="idx === gallery.length - 1 ? '' : 'margin-right: 8px;'">
              <view
                class="rounded-full"
                :style="idx === galleryIndex ? 'width:8px;height:8px;background:#fff;' : 'width:6px;height:6px;background:rgba(255,255,255,0.55);'"
              />
            </view>
          </view>
        </view>

        <!-- 加载/错误状态 -->
        <view v-if="loading" class="-mt-4 bg-white px-5 py-4">
          <text class="text-gray-500" style="font-size: 12px;">加载中...</text>
        </view>
        <view v-else-if="errorMsg" class="-mt-4 bg-white px-5 py-4">
          <text style="font-size: 12px; color: #ef4444;">{{ errorMsg }}</text>
        </view>

        <view v-else>
          <!-- 主信息卡片 -->
          <view class="relative -mt-4 rounded-t-3xl bg-white px-5 pt-6 pb-5 shadow-sm">
          <view class="mb-2 flex items-baseline" style="gap: 8px;">
            <text class="text-primary" style="font-size: 28px; font-weight: 900;">
              {{ displayPrice }}
            </text>
            <text v-if="displayOriginPrice" class="text-gray-400 line-through" style="font-size: 12px;">
              {{ displayOriginPrice }}
            </text>
            <view
              v-if="showOfferTag"
              class="ml-2 rounded-md px-2 py-1"
              style="background: rgba(217,119,6,0.10);"
            >
              <text style="font-size: 10px; font-weight: 900; color: #d97706;">限时优惠</text>
            </view>
          </view>

          <view v-if="commissionText" class="mb-4 w-full rounded-lg border px-3 py-2" style="border-color: rgba(245,158,11,0.25); background: rgba(245,158,11,0.08);">
            <view class="flex items-center">
              <text style="font-size: 12px; font-weight: 800; color: #92400e;">分销专享：</text>
              <text style="font-size: 12px; font-weight: 800; color: #b45309;">{{ commissionText }}</text>
              <view class="ml-auto">
                <text style="font-size: 12px; color: rgba(180,83,9,0.65);">›</text>
              </view>
            </view>
          </view>

          <text class="mb-3 text-gray-900" style="font-size: 18px; font-weight: 900; line-height: 1.25;" :number-of-lines="2">
            {{ productTitle }}
          </text>

          <view class="flex items-center justify-between border-t border-b py-3" style="border-color: rgba(243,244,246,1);">
            <text class="text-gray-500" style="font-size: 12px;">已售 {{ soldText }}</text>
            <view class="h-3 w-px bg-gray-200"></view>
            <text class="text-gray-500" style="font-size: 12px;">库存 {{ stockText }}</text>
            <view class="h-3 w-px bg-gray-200"></view>
            <text class="text-gray-500" style="font-size: 12px;">{{ locationText }}</text>
          </view>

          <view class="mt-3 flex items-center overflow-x-auto no-scrollbar" style="gap: 16px;">
            <view class="flex items-center" style="gap: 6px;">
              <text class="text-primary" style="font-size: 12px; font-weight: 900;">✓</text>
              <text class="text-gray-600" style="font-size: 12px;">官方保障</text>
            </view>
            <view class="flex items-center" style="gap: 6px;">
              <text class="text-primary" style="font-size: 12px; font-weight: 900;">✓</text>
              <text class="text-gray-600" style="font-size: 12px;">免运费</text>
            </view>
            <view class="flex items-center" style="gap: 6px;">
              <text class="text-primary" style="font-size: 12px; font-weight: 900;">✓</text>
              <text class="text-gray-600" style="font-size: 12px;">7天退换</text>
            </view>
          </view>
          </view>

          <view style="height: 8px;"></view>

          <!-- 已选规格 -->
          <view
            class="w-full bg-white px-5 py-4"
            hover-class="bg-gray-50"
            @tap="openSpecPicker"
          >
            <view class="flex items-center justify-between">
              <view class="flex flex-col" style="gap: 4px;">
                <text class="text-gray-900" style="font-size: 14px; font-weight: 900;">已选规格</text>
                <text class="text-gray-500" style="font-size: 12px;">{{ selectedSpecText }}</text>
              </view>
              <view class="flex items-center" style="gap: 12px;">
                <view class="flex" style="margin-right: 4px;">
                  <view class="-mr-2 h-6 w-6 rounded-full" style="border: 2px solid #fff; background: #111827;"></view>
                  <view class="-mr-2 h-6 w-6 rounded-full" style="border: 2px solid #fff; background: #e5e7eb;"></view>
                  <view class="h-6 w-6 rounded-full" style="border: 2px solid #fff; background: #1e3a8a;"></view>
                </view>
                <text class="text-gray-400" style="font-size: 14px;">›</text>
              </view>
            </view>
          </view>

          <view style="height: 8px;"></view>

          <!-- 评价 -->
          <view class="bg-white px-5 py-4">
          <view class="mb-4 flex items-center justify-between">
            <view class="flex items-center" style="gap: 8px;">
              <text class="text-gray-900" style="font-size: 16px; font-weight: 900;">评价</text>
              <text class="text-gray-500" style="font-size: 12px;">({{ reviews.length }})</text>
            </view>
            <view class="flex items-center" style="gap: 4px;" hover-class="opacity-80" @tap="onOpenAllReviews">
              <text class="text-primary" style="font-size: 12px; font-weight: 800;">{{ positiveRateText }}</text>
              <text class="text-primary" style="font-size: 14px;">›</text>
            </view>
          </view>

          <view v-for="(r, idx) in reviews" :key="r.id" :style="idx === reviews.length - 1 ? '' : 'margin-bottom: 16px;'">
            <view class="flex" style="gap: 12px;">
              <image class="h-8 w-8 rounded-full" mode="aspectFill" :src="r.avatarUrl" />
              <view class="flex-1">
                <view class="mb-1 flex items-center justify-between">
                  <text class="text-gray-900" style="font-size: 12px; font-weight: 800;">{{ r.name }}</text>
                  <text class="text-gray-400" style="font-size: 10px;">{{ r.timeText }}</text>
                </view>
                <view class="mb-1 flex">
                  <text v-for="idx in 5" :key="idx" style="font-size: 12px; color: #facc15;">{{ idx <= r.stars ? "★" : "☆" }}</text>
                </view>
                <text class="text-gray-600" style="font-size: 12px; line-height: 1.5;" :number-of-lines="2">
                  {{ r.content }}
                </text>
              </view>
            </view>
            <view v-if="r.photos.length" class="mt-2 flex" style="gap: 8px; padding-left: 44px;">
              <image
                v-for="(p, idx) in r.photos"
                :key="`${r.id}-${idx}`"
                class="h-16 w-16 rounded-lg"
                mode="aspectFill"
                :src="p"
                @tap="previewImages(r.photos, p)"
              />
            </view>
          </view>
          </view>

          <view style="height: 8px;"></view>

          <!-- 店铺 -->
          <view class="bg-white px-5 py-5">
          <view class="flex items-center" style="gap: 12px;">
            <image class="h-12 w-12 rounded-full" mode="aspectFill" :src="storeInfo.logoUrl" />
            <view class="min-w-0 flex-1">
              <text class="truncate text-gray-900" style="font-size: 14px; font-weight: 900;">{{ storeInfo.name }}</text>
              <view class="mt-1 flex items-center" style="gap: 8px;">
                <view class="rounded px-2 py-1" style="background: #386657;">
                  <text style="font-size: 10px; font-weight: 900; color: #fff;">官方</text>
                </view>
                <text class="text-gray-500" style="font-size: 12px;">{{ storeInfo.ratingText }}</text>
              </view>
            </view>
            <view
              class="h-8 rounded-full border px-4"
              style="border-color: #386657;"
              hover-class="opacity-90"
              @tap="onVisitStore"
            >
              <view class="flex h-full items-center justify-center">
                <text style="font-size: 12px; font-weight: 900; color: #386657;">进店</text>
              </view>
            </view>
          </view>

          <view class="mt-5 grid grid-cols-3 gap-4 border-t pt-4" style="border-color: rgba(243,244,246,1);">
            <view class="flex flex-col items-center" style="gap: 2px;">
              <text class="text-gray-900" style="font-size: 14px; font-weight: 900;">{{ storeInfo.items }}</text>
              <text class="text-gray-400" style="font-size: 10px; font-weight: 700;">商品</text>
            </view>
            <view class="flex flex-col items-center" style="gap: 2px;">
              <text class="text-gray-900" style="font-size: 14px; font-weight: 900;">{{ storeInfo.fans }}</text>
              <text class="text-gray-400" style="font-size: 10px; font-weight: 700;">粉丝</text>
            </view>
            <view class="flex flex-col items-center" style="gap: 2px;">
              <text class="text-gray-900" style="font-size: 14px; font-weight: 900;">{{ storeInfo.service }}</text>
              <text class="text-gray-400" style="font-size: 10px; font-weight: 700;">服务</text>
            </view>
          </view>
          </view>

          <view style="height: 8px;"></view>

          <!-- 商品详情 -->
          <view class="bg-white px-5 py-6 pb-8">
          <view class="mb-4 flex items-center" style="gap: 8px;">
            <view class="h-4 w-1 rounded-full bg-primary"></view>
            <text class="text-gray-900" style="font-size: 16px; font-weight: 900;">商品详情</text>
          </view>
          <view class="text-gray-600" style="font-size: 14px; line-height: 1.7;">
            <text v-if="product?.description">{{ product.description }}</text>
            <text v-else>暂无详情描述</text>
          </view>

          <view v-if="detailImages.length" class="mt-4">
            <image
              v-for="(img, idx) in detailImages"
              :key="`${img}-${idx}`"
              class="mb-3 w-full rounded-2xl"
              mode="widthFix"
              :src="img"
              @tap="previewImages(detailImages, img)"
            />
          </view>

          <view class="mt-4 rounded-xl border p-4" style="border-color: rgba(243,244,246,1); background: rgba(249,250,251,1);">
            <text class="mb-3 text-gray-900" style="font-size: 12px; font-weight: 900;">核心亮点</text>
            <view v-for="(f, idx) in keyFeatures" :key="idx" class="flex" style="gap: 8px;" :style="idx === keyFeatures.length - 1 ? '' : 'margin-bottom: 8px;'">
              <text class="text-primary" style="font-size: 14px; font-weight: 900;">✓</text>
              <text class="text-gray-700" style="font-size: 13px; line-height: 1.5;">{{ f }}</text>
            </view>
          </view>
          </view>
        </view>
      </view>

      <!-- 底部操作栏 -->
	      <view class="fixed bottom-0 left-0 right-0 z-40 mx-auto max-w-md border-t bg-white safe-pb" style="border-color: rgba(243,244,246,1);">
	        <view class="flex items-center justify-between px-3 py-2" style="gap: 12px; height: 60px;">
	          <view class="flex items-center px-3" style="gap: 18px;">
	            <view class="flex flex-col items-center" hover-class="opacity-80" @tap="onStore">
	              <image class="mb-1" style="width: 22px; height: 22px;" mode="aspectFit" src="/static/icons/shop.svg" />
	              <text style="font-size: 10px; line-height: 1; font-weight: 800; color: #6b7280;">店铺</text>
	            </view>
	            <view class="flex flex-col items-center" hover-class="opacity-80" @tap="onChat">
	              <image class="mb-1" style="width: 22px; height: 22px;" mode="aspectFit" src="/static/icons/chat.svg" />
	              <text style="font-size: 10px; line-height: 1; font-weight: 800; color: #6b7280;">客服</text>
	            </view>
	            <view class="relative flex flex-col items-center" hover-class="opacity-80" @tap="toCart">
	              <image class="mb-1" style="width: 22px; height: 22px;" mode="aspectFit" src="/static/icons/cart.svg" />
	              <text style="font-size: 10px; line-height: 1; font-weight: 800; color: #6b7280;">购物车</text>
	              <view
	                v-if="cartCount > 0"
	                class="absolute -top-1 -right-2 flex items-center justify-center rounded-full bg-red-500 px-1"
	                style="min-width: 14px; height: 14px;"
	              >
                <text style="font-size: 9px; font-weight: 900; color: #fff;">{{ cartCount }}</text>
              </view>
            </view>
          </view>

          <view class="flex flex-1 items-center" style="gap: 8px; height: 40px;">
            <view
              class="flex-1 rounded-full border"
              :style="!isSubscription && !hasAnySellableSku ? 'border-color: #d1d5db; background: rgba(107,114,128,0.08); opacity: 0.6;' : 'border-color: #386657; background: rgba(56,102,87,0.10);'"
              hover-class="opacity-90"
              @tap="onAddToCart"
            >
              <view class="flex h-10 items-center justify-center">
                <text style="font-size: 12px; font-weight: 900; color: #386657;">加入购物车</text>
              </view>
            </view>
            <view
              class="flex-1 rounded-full bg-primary"
              :style="!isSubscription && !hasAnySellableSku ? 'opacity: 0.6;' : ''"
              hover-class="opacity-90"
              @tap="onBuyNow"
            >
              <view class="flex h-10 items-center justify-center">
                <text style="font-size: 12px; font-weight: 900; color: #fff;">立即购买</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <SkuPickerSheet
        v-if="!isSubscription"
        v-model="skuSheetOpen"
        v-model:selectedSkuId="selectedSkuId"
        :skus="skus"
        :spec-groups="specGroups"
        :cover-url="coverUrl"
        :max-qty="10"
        @confirm="onSkuSheetConfirm"
      />
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import SkuPickerSheet from "@/components/product/sku-picker-sheet.vue";
import {
  miniAppGetProduct,
  miniAppGetProductDetailWithSpec,
  miniAppGetSellability,
  miniAppListSkus,
  miniAppListSubscriptionPlans,
  type MiniAppProductDetail,
  type MiniAppSpecGroup,
  type MiniAppSkuSummary,
  type MiniAppSubscriptionPlan,
} from "@/services/miniapp-product";
import { addCartItem, cartCount as countCart, getLocalCart } from "@/services/cart";
import { isLikelyPlaceholderUrl } from "@/utils/product-images";

const topInset = ref(40);
const spuId = ref("");
const sellabilityChannel = ref(String(uni.getStorageSync("miniapp.channel") || "official").trim() || "official");
const sellabilityLocale = ref(String(uni.getStorageSync("miniapp.locale") || "zh-CN").trim() || "zh-CN");

const loading = ref(false);
const errorMsg = ref("");

const product = ref<MiniAppProductDetail | null>(null);
const skus = ref<MiniAppSkuSummary[]>([]);
const specGroups = ref<MiniAppSpecGroup[]>([]);
const plans = ref<MiniAppSubscriptionPlan[]>([]);

const selectedSkuId = ref("");
const selectedPlanId = ref("");
const skuSheetOpen = ref(false);

const coverUrl = ref("");

const isSubscription = ref(false);
const galleryIndex = ref(0);

const productTitle = computed(() => product.value?.name || product.value?.code || "—");

const displayPrice = computed(() => {
  if (product.value?.priceLabel) return product.value.priceLabel;
  const currency = product.value?.currency || "CNY";
  const p = product.value?.minPrice ?? product.value?.maxPrice;
  if (p == null) return "--";
  return formatMoney(currency, p);
});

const displayOriginPrice = computed(() => {
  const min = product.value?.minPrice;
  const max = product.value?.maxPrice;
  if (min == null || max == null) return "";
  if (max <= min) return "";
  return formatMoney(product.value?.currency || "CNY", max);
});

const showOfferTag = computed(() => Boolean(displayOriginPrice.value));

const commissionText = computed(() => {
  const min = product.value?.minPrice;
  if (min == null) return "";
  const commission = Math.max(1, Math.round(min * 0.1));
  return `${formatMoney(product.value?.currency || "CNY", commission)} / 单`;
});

const soldText = computed(() => "2k+");
const stockText = computed(() => {
  if (isSubscription.value) return "—";
  const s = skus.value.find((x) => x.id === selectedSkuId.value);
  if (typeof s?.availableQty === "number") return String(Math.max(0, s.availableQty));
  if (typeof s?.stockQty === "number") return String(Math.max(0, s.stockQty));
  const total = skus.value.reduce((acc, x) => acc + (typeof x.availableQty === "number" ? Math.max(0, x.availableQty) : 0), 0);
  if (total > 0) return String(total);
  return "—";
});
const locationText = computed(() => "广东");

function sellabilityReasonToText(code: string) {
  switch (String(code || "").toUpperCase()) {
    case "NO_PUBLIC_PRICE":
      return "暂无价格";
    case "OUT_OF_STOCK":
      return "缺货";
    case "CHANNEL_DISABLED":
    case "NOT_IN_AVAILABILITY_WINDOW":
    case "CHANNEL_STATUS_BLOCKED":
      return "暂不可售";
    case "SKU_NOT_ONLINE":
      return "未上架";
    default:
      return "暂不可售";
  }
}

const selectedSku = computed(() => skus.value.find((x) => x.id === selectedSkuId.value) || null);
const selectedSkuSellable = computed(() => Boolean(selectedSku.value && selectedSku.value.sellable === true));
const hasAnySellableSku = computed(() => skus.value.some((s) => s.sellable === true));
const selectedSkuDisableReason = computed(() => {
  const code = (selectedSku.value?.sellabilityReasons || [])[0] || "";
  return sellabilityReasonToText(code);
});

const gallery = computed(() => {
  const urls = new Set<string>();
  if (coverUrl.value) urls.add(coverUrl.value);
  skus.value.forEach((s) => {
    if (s.imageUrl) urls.add(s.imageUrl);
  });
	  const list = Array.from(urls).filter(Boolean);
	  if (!list.length && coverUrl.value) list.push(coverUrl.value);
	  return list.slice(0, 6);
	});

const detailImages = computed(() => gallery.value.slice(0, 2));

const keyFeatures = computed(() => {
  const type = String(product.value?.type || "").toLowerCase();
  return type === "subscription"
    ? ["按周期自动扣费，随时可取消", "支持试用期与自动续费配置", "订单/订阅状态与权益可扩展"]
    : ["支持多 SKU/多价格区间", "图文详情与参数可扩展", "下单/购物车能力待接入"];
});

const selectedSpecText = computed(() => {
  if (isSubscription.value) {
    const p = plans.value.find((x) => x.id === selectedPlanId.value);
    return p ? `${p.name} · ${formatMoney(p.currency, p.price)}` : "请选择订阅计划";
  }
  const s = skus.value.find((x) => x.id === selectedSkuId.value);
  if (!s) return "请选择 SKU";
  const price = s.price != null ? formatMoney(s.currency || "CNY", s.price) : "--";
  return `${s.code} · ${price}`;
});

const positiveRateText = computed(() => "98% 好评");

const cartCount = ref(0);

const storeInfo = ref({
  name: "PowerX 官方店",
  logoUrl: "https://picsum.photos/seed/powerx-store/128/128",
  ratingText: "4.9 评分",
  items: "245",
  fans: "12k",
  service: "100%",
});

type Review = {
  id: string;
  name: string;
  timeText: string;
  stars: number;
  content: string;
  avatarUrl: string;
  photos: string[];
};

const reviews = ref<Review[]>([
  {
    id: "r1",
    name: "Sarah M.",
    timeText: "2天前",
    stars: 5,
    content: "质感很好，做工细腻；实际体验超出预期，性价比很高。",
    avatarUrl: "https://picsum.photos/seed/powerx-user-1/80/80",
    photos: ["https://picsum.photos/seed/powerx-review-1/200/200", "https://picsum.photos/seed/powerx-review-2/200/200"],
  },
]);

function ensureTopInset() {
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
}

function goBack() {
  uni.navigateBack();
}

function selectSku(id: string) {
  selectedSkuId.value = id;
}

function selectPlan(id: string) {
  selectedPlanId.value = id;
}

function formatMoney(currency: string, price: number) {
  const c = String(currency || "").toUpperCase();
  const symbol = c === "CNY" || c === "RMB" ? "¥" : c === "USD" ? "$" : c === "EUR" ? "€" : `${c} `;
  const s = String(price);
  return `${symbol}${s}`;
}

function onPrimaryAction() {
  uni.showToast({ title: "购物车能力待接入", icon: "none" });
}

function onAddToCart() {
  if (isSubscription.value) {
    onPrimaryAction();
    return;
  }
  if (!skus.value.length) {
    uni.showToast({ title: "暂无可选 SKU", icon: "none" });
    return;
  }
  if (!hasAnySellableSku.value) {
    uni.showToast({ title: selectedSkuDisableReason.value || "暂不可售", icon: "none" });
    return;
  }
  skuSheetOpen.value = true;
}

function onBuyNow() {
  if (isSubscription.value) {
    uni.showToast({ title: "下单能力待接入", icon: "none" });
    return;
  }
  if (!skus.value.length) {
    uni.showToast({ title: "暂无可选 SKU", icon: "none" });
    return;
  }
  if (!hasAnySellableSku.value) {
    uni.showToast({ title: selectedSkuDisableReason.value || "暂不可售", icon: "none" });
    return;
  }
  skuSheetOpen.value = true;
}

function toCart() {
  uni.switchTab({ url: "/pages/cart/index" });
}

function onStore() {
  uni.switchTab({ url: "/pages/mall/index" });
}

function onChat() {
  uni.showToast({ title: "客服能力待接入", icon: "none" });
}

function onShare() {
  uni.showToast({ title: "分享能力待接入", icon: "none" });
}

function onMore() {
  uni.showToast({ title: "更多操作待接入", icon: "none" });
}

function onVisitStore() {
  uni.showToast({ title: "店铺页待接入", icon: "none" });
}

function onOpenAllReviews() {
  uni.showToast({ title: "评价列表待接入", icon: "none" });
}

function previewImages(urls: string[], current: string) {
  if (!urls.length) return;
  uni.previewImage({ urls, current });
}

function onGalleryChange(e: any) {
  const current = Number(e?.detail?.current ?? 0);
  galleryIndex.value = Number.isFinite(current) ? current : 0;
}

function onSkuSheetConfirm(payload: { action: "cart" | "buy"; skuId: string; qty: number }) {
  selectSku(payload.skuId);
  if (payload.action === "cart") {
    const sku = skus.value.find((x) => x.id === payload.skuId);
    const rawMax = Number(sku?.availableQty ?? sku?.stockQty ?? 10);
    const stockMax = Number.isFinite(rawMax) && rawMax > 0 ? Math.max(1, Math.floor(rawMax)) : 10;
    const maxQty = Math.min(10, stockMax);
    const currency = String((sku as any)?.currency || product.value?.currency || "CNY").trim() || "CNY";
    const unitPriceRaw = Number((sku as any)?.price ?? 0);
    const unitPrice = Number.isFinite(unitPriceRaw) && unitPriceRaw >= 0 ? unitPriceRaw : undefined;
    const skuImageUrl = String((sku as any)?.imageUrl || "").trim();
    const cover = String(coverUrl.value || "").trim();
    addCartItem(payload.skuId, payload.qty, {
      spuId: String(spuId.value || "").trim() || undefined,
      title: String(product.value?.name || "").trim() || "商品",
      // 与商城保持一致：避免把 picsum 这类“占位域名”写进购物车，统一走占位图池
      imageUrl:
        (cover && !isLikelyPlaceholderUrl(cover) ? cover : "") ||
        (skuImageUrl && !isLikelyPlaceholderUrl(skuImageUrl) ? skuImageUrl : "") ||
        undefined,
      // 购物车里 skuLabel 用于“规格下拉”展示，这里只放 SKU code，价格在卡片主区域展示即可。
      skuLabel: String((sku as any)?.code || "").trim() || undefined,
      skuCode: String((sku as any)?.code || "").trim() || undefined,
      maxQty,
      currency,
      unitPrice,
    });
    cartCount.value = countCart(getLocalCart().items);
    uni.showToast({ title: `已加入购物车 x${payload.qty}`, icon: "none" });
    return;
  }
  uni.showToast({ title: `已选 ${payload.qty} 件，请到购物车结算`, icon: "none" });
  uni.switchTab({ url: "/pages/cart/index" });
}

function openSpecPicker() {
  if (isSubscription.value) {
    if (!plans.value.length) return;
    uni.showActionSheet({
      itemList: plans.value.map((p) => `${p.name} · ${formatMoney(p.currency, p.price)}`),
      success: (res) => {
        const idx = Number(res?.tapIndex ?? -1);
        const p = plans.value[idx];
        if (p) selectPlan(p.id);
      },
    });
    return;
  }

  if (!skus.value.length) return;
  uni.showActionSheet({
    itemList: skus.value.map((s) => `${s.code} · ${s.price != null ? formatMoney(s.currency || "CNY", s.price) : "--"}`),
    success: (res) => {
      const idx = Number(res?.tapIndex ?? -1);
      const s = skus.value[idx];
      if (s) selectSku(s.id);
    },
  });
}

async function loadAll() {
  if (!spuId.value) {
    errorMsg.value = "缺少商品ID";
    return;
  }
  loading.value = true;
  errorMsg.value = "";
  try {
    const p = await miniAppGetProduct(spuId.value);
    product.value = p;
	    coverUrl.value = p?.coverUrl || "/static/icons/image-placeholder.svg";
    isSubscription.value = String(p?.type || "").toLowerCase() === "subscription";

    if (isSubscription.value) {
      const resp = await miniAppListSubscriptionPlans(spuId.value);
      plans.value = Array.isArray(resp?.items) ? resp.items : [];
      if (!selectedPlanId.value && plans.value.length) selectedPlanId.value = plans.value[0].id;
      return;
    }

    try {
      const detail = await miniAppGetProductDetailWithSpec(spuId.value);
      if (detail?.spu) product.value = detail.spu;
	      coverUrl.value = (detail?.spu?.coverUrl || product.value?.coverUrl) ?? coverUrl.value;
      specGroups.value = Array.isArray(detail?.spec?.groups) ? detail.spec!.groups : [];
      skus.value = Array.isArray(detail?.skus) ? detail.skus : [];
    } catch {
      const resp = await miniAppListSkus(spuId.value, 1, 50);
      skus.value = Array.isArray(resp?.items) ? resp.items : [];
      specGroups.value = [];
    }

    if (skus.value.length) {
      try {
        const sellability = await miniAppGetSellability(spuId.value, {
          channel: sellabilityChannel.value,
          locale: sellabilityLocale.value,
        });
        const items = Array.isArray(sellability?.items) ? sellability.items : [];
        const bySku = new Map<string, (typeof items)[number]>();
        items.forEach((it) => {
          if (!it?.skuId) return;
          bySku.set(String(it.skuId), it);
        });
        skus.value = skus.value.map((s) => {
          const it = bySku.get(String(s.id));
          if (!it) {
            return {
              ...s,
              sellable: false,
              sellabilityReasons: ["UNKNOWN"],
              availableQty: 0,
            };
          }
          const price = it.price?.amount;
          const currency = it.price?.currency;
          return {
            ...s,
            sellable: Boolean(it.sellable),
            sellabilityReasons: Array.isArray(it.reasons) ? it.reasons : [],
            availableQty: Number.isFinite(it.availableQty) ? Math.max(0, Number(it.availableQty)) : 0,
            price: typeof price === "number" ? price : s.price,
            currency: typeof currency === "string" && currency ? currency : s.currency,
          };
        });
      } catch {
        skus.value = skus.value.map((s) => ({
          ...s,
          sellable: false,
          sellabilityReasons: ["UNKNOWN"],
          availableQty: typeof s.stockQty === "number" ? Math.max(0, s.stockQty) : 0,
        }));
      }
    }

    if (!selectedSkuId.value && skus.value.length) {
      const firstSellable = skus.value.find((s) => s.sellable === true);
      selectedSkuId.value = (firstSellable || skus.value[0]).id;
    } else if (selectedSkuId.value) {
      const current = skus.value.find((s) => s.id === selectedSkuId.value);
      if (current && current.sellable === false) {
        const firstSellable = skus.value.find((s) => s.sellable === true);
        if (firstSellable) selectedSkuId.value = firstSellable.id;
      }
    }
  } catch (e: any) {
    errorMsg.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

onLoad((query: any) => {
  spuId.value = String(query?.id || "");
});

onMounted(() => {
  ensureTopInset();
  cartCount.value = countCart(getLocalCart().items);
  loadAll();
});
</script>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.safe-pb {
  padding-bottom: env(safe-area-inset-bottom);
}
</style>
