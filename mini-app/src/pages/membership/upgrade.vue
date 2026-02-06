<template>
  <view class="min-h-screen w-full bg-background-light font-display text-text-dark" style="padding-bottom: 120px;">
    <view :style="`padding-top:${topInset}px; background: rgba(246, 248, 246, 0.9);`" class="sticky top-0 z-10 w-full backdrop-blur">
      <view class="flex items-center justify-between px-4 pt-4 pb-2">
        <view class="text-xl font-extrabold tracking-tight">会籍升级</view>
        <view class="h-8 w-8 rounded-full shadow-sm flex items-center justify-center" style="background: rgba(255,255,255,0.8);" hover-class="opacity-80" @tap="back">
          <text class="text-base font-black">✕</text>
        </view>
      </view>
    </view>

    <view class="px-4 pt-2">
      <view class="mb-6 px-1">
        <view class="text-2xl font-extrabold leading-tight">开启更高级别权益</view>
        <view class="mt-2 text-sm text-muted leading-relaxed">升级成为高级会籍，享受更高等级权益与专属服务支持。</view>
      </view>

      <view v-if="loading" class="py-8 text-center text-sm text-muted">加载中...</view>
      <view v-else-if="products.length === 0" class="py-8 text-center text-sm text-muted">
        暂无可用会籍订阅
      </view>
      <view v-else class="space-y-5">
        <view
          v-for="p in products"
          :key="p.id"
          class="rounded-3xl bg-white p-5 shadow-sm"
        >
          <view class="flex items-center gap-4 mb-5">
            <image class="h-16 w-16 rounded-2xl bg-gray-100" mode="aspectFill" :src="p.coverUrl" />
            <view class="flex-1 min-w-0">
              <view class="text-base font-extrabold truncate">{{ p.name }}</view>
              <view class="mt-1 text-xs text-muted">{{ p.subtitle || p.code }}</view>
            </view>
          </view>

          <view class="space-y-3">
            <view
              v-for="plan in p.plans"
              :key="plan.id"
              :class="planCardClass(plan)"
              class="flex items-center justify-between rounded-2xl px-4 py-3"
              hover-class="opacity-90"
              @tap="selectPlan(p, plan)"
            >
              <view class="min-w-0 flex-1">
                <view class="flex items-center gap-2">
                  <view class="text-sm font-semibold">{{ plan.name || plan.planCode }}</view>
                  <view
                    v-if="planHighlightTag(plan)"
                    class="px-2 py-1 rounded font-bold"
                    style="background: rgba(19,236,91,0.18); color: #0d110e; font-size: 10px;"
                  >
                    {{ planHighlightTag(plan) }}
                  </view>
                  <view
                    v-if="plan.id === selectedPlanId"
                    class="px-2 py-1 rounded font-bold"
                    style="background: rgba(19,236,91,0.35); color: #0d110e; font-size: 10px;"
                  >
                    已选
                  </view>
                </view>
                <view class="mt-1 text-xs text-muted">
                  {{ plan.billingLabel }} · {{ formatMoney(plan.currency, plan.price) }}
                </view>
              </view>
              <button
                class="ml-4 rounded-full bg-primary px-5 py-2 text-xs font-extrabold shrink-0"
                style="color: #ffffff; min-width: 96px;"
                @tap.stop="startSubscribe(p, plan)"
              >
                立即开通
              </button>
            </view>
          </view>
        </view>
      </view>

      <view class="mt-8 mb-24">
        <view class="text-center text-base font-extrabold mb-2">会员专属特权</view>
        <view class="text-center text-xs text-primary font-semibold mb-4" hover-class="opacity-80" @tap="toBenefitsDetail">查看权益详情 ›</view>
        <view v-if="selectedBenefits.length === 0" class="text-center text-xs text-muted">
          当前订阅计划暂未配置权益
        </view>
        <view v-else class="grid grid-cols-2 gap-3">
          <view
            v-for="benefit in selectedBenefits"
            :key="benefit.id"
            class="bg-white p-5 rounded-2xl shadow-sm border border-gray-50 flex flex-col items-center text-center"
          >
            <view class="h-12 w-12 rounded-full flex items-center justify-center mb-3" style="background: rgba(19,236,91,0.15);">
              <text class="text-xl text-primary">★</text>
            </view>
            <view class="text-sm font-bold mb-1">{{ benefit.name }}</view>
            <view v-if="benefitSubtitle(benefit)" class="text-xs text-muted leading-relaxed">
              {{ benefitSubtitle(benefit) }}
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { miniAppListProducts, miniAppListSubscriptionPlans } from "@/services/miniapp-product";
import { miniAppListMembershipBenefits } from "@/services/miniapp-membership";

const topInset = ref(40);
const loading = ref(false);

type Plan = {
  id: string;
  skuId?: string;
  benefitIds?: string[];
  planCode?: string;
  name?: string;
  billingCycle?: string;
  billingValue?: number;
  price?: number;
  currency?: string;
  billingLabel: string;
};

type ProductWithPlans = {
  id: string;
  code: string;
  name: string;
  subtitle?: string;
  coverUrl: string;
  plans: Plan[];
};

const products = ref<ProductWithPlans[]>([]);
const selectedProductId = ref("");
const selectedPlanId = ref("");
const benefits = ref<Record<string, { id: string; name: string; items?: any }>>({});

const selectedPlan = computed(() => {
  const product = products.value.find((p) => p.id === selectedProductId.value) || products.value[0];
  if (!product) return undefined;
  return product.plans.find((p) => p.id === selectedPlanId.value) || product.plans[0];
});

const selectedBenefits = computed(() => {
  const ids = selectedPlan.value?.benefitIds || [];
  return ids
    .map((id) => benefits.value[id])
    .filter((b) => b && b.id)
    .slice(0, 4);
});

function ensureTopInset() {
  try {
    const wxAny = (globalThis as any).wx;
    if (wxAny && typeof wxAny.getWindowInfo === "function") {
      const info = wxAny.getWindowInfo();
      const statusBar = Number(info?.statusBarHeight ?? 0);
      topInset.value = Math.max(40, statusBar + 12);
      return;
    }
  } catch {
    // ignore
  }
}

function back() {
  uni.navigateBack({ delta: 1 });
}

function billingLabel(plan: Plan) {
  const cycle = String(plan.billingCycle || "").toLowerCase();
  if (cycle === "monthly") return "月付";
  if (cycle === "quarterly") return "季付";
  if (cycle === "yearly") return "年付";
  if (cycle === "custom" && plan.billingValue) return `自定义(${plan.billingValue}天)`;
  return cycle || "订阅";
}

function formatMoney(currency?: string, amount?: number) {
  const value = typeof amount === "number" ? amount : NaN;
  if (!Number.isFinite(value)) return "--";
  const code = String(currency || "CNY").trim() || "CNY";
  try {
    return new Intl.NumberFormat("zh-CN", { style: "currency", currency: code }).format(value);
  } catch {
    return `${value} ${code}`;
  }
}

function planHighlightTag(plan: Plan) {
  const cycle = String(plan.billingCycle || "").toLowerCase();
  if (cycle === "yearly") return "超值";
  return "";
}

function planCardClass(plan: Plan) {
  if (plan.id === selectedPlanId.value) return "plan-selected";
  return "plan-normal";
}

function selectPlan(product: ProductWithPlans, plan: Plan) {
  selectedProductId.value = product.id;
  selectedPlanId.value = plan.id;
}

function benefitSubtitle(benefit: { items?: any }) {
  const raw = benefit?.items;
  if (Array.isArray(raw) && raw.length > 0) {
    const first = raw[0] || {};
    const name = String(first?.name || first?.title || "").trim();
    const desc = String(first?.desc || first?.description || "").trim();
    if (name && desc) return `${name} · ${desc}`;
    return name || desc;
  }
  if (typeof raw === "string") return raw;
  return "";
}

async function loadProducts() {
  loading.value = true;
  try {
    const resp = await miniAppListProducts({ type: "subscription", hasPlans: true, page: 1, pageSize: 20 });
    const items = Array.isArray(resp?.items) ? resp.items : [];
    const enriched: ProductWithPlans[] = [];
    for (const it of items) {
      const planResp = await miniAppListSubscriptionPlans(it.id);
      const plans = (planResp?.items || []).map((p) => ({
        id: p.id,
        skuId: p.skuId,
        benefitIds: Array.isArray(p.benefitIds) ? p.benefitIds : [],
        planCode: p.planCode,
        name: p.name,
        billingCycle: p.billingCycle,
        billingValue: p.billingValue,
        price: p.price,
        currency: p.currency,
        billingLabel: billingLabel(p as any),
      }));
      enriched.push({
        id: it.id,
        code: it.code,
        name: it.name || it.code,
        subtitle: it.type,
        coverUrl: it.coverUrl || "/static/icons/image-placeholder.png",
        plans,
      });
    }
    products.value = enriched;
    if (enriched.length && enriched[0].plans.length) {
      selectedProductId.value = enriched[0].id;
      selectedPlanId.value = enriched[0].plans[0].id;
    }
  } finally {
    loading.value = false;
  }
}

async function loadBenefits() {
  try {
    const resp = await miniAppListMembershipBenefits();
    const items = Array.isArray(resp?.items) ? resp.items : [];
    const map: Record<string, { id: string; name: string; items?: any }> = {};
    for (const it of items) {
      if (!it?.id) continue;
      map[String(it.id)] = {
        id: String(it.id),
        name: String(it.name || "").trim() || "权益",
        items: it.items,
      };
    }
    benefits.value = map;
  } catch {
    benefits.value = {};
  }
}

function startSubscribe(product: ProductWithPlans, plan: Plan) {
  const skuId = String(plan.skuId || "").trim();
  if (!skuId) {
    uni.showToast({ title: "订阅计划未绑定 SKU", icon: "none" });
    return;
  }
  const channel = String(uni.getStorageSync("miniapp.channel") || "official").trim() || "official";
  const locale = String(uni.getStorageSync("miniapp.locale") || "zh-CN").trim() || "zh-CN";
  const draftKey = `miniapp.checkout.draft.${Date.now()}-${Math.random().toString(16).slice(2)}`;
  const payload = {
    channel,
    locale,
    from: "membership",
    items: [
      {
        skuId,
        spuId: product.id,
        qty: 1,
        title: product.name,
        skuLabel: plan.name || plan.planCode,
        skuCode: plan.planCode,
        imageUrl: product.coverUrl,
        currency: String(plan.currency || "CNY").trim() || "CNY",
        unitPrice: Number(plan.price || 0) || 0,
      },
    ],
  };
  uni.setStorageSync(draftKey, JSON.stringify(payload));
  uni.navigateTo({ url: `/pages/order/confirm?draftKey=${encodeURIComponent(draftKey)}` });
}

function toBenefitsDetail() {
  uni.navigateTo({ url: "/pages/membership/benefits" });
}

onMounted(() => {
  ensureTopInset();
  loadProducts();
  loadBenefits();
});
</script>

<style scoped>
.plan-highlight {
  border: 1px solid rgba(19, 236, 91, 0.3);
  background: rgba(19, 236, 91, 0.06);
}

.plan-normal {
  border: 1px solid #f3f4f6;
  background: #ffffff;
}

.plan-selected {
  border: 1px solid rgba(19, 236, 91, 0.6);
  background: rgba(19, 236, 91, 0.12);
}
</style>
