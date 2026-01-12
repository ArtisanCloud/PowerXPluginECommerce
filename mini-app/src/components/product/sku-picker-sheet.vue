<template>
  <view v-if="modelValue" class="fixed inset-0 z-50">
    <view class="absolute inset-0" style="background: rgba(17, 24, 39, 0.6);" @tap="close"></view>

    <view class="absolute bottom-0 left-0 right-0 z-10 mx-auto flex max-w-md flex-col bg-white rounded-t-2xl shadow-card animate-slide-up" style="height: 85%; max-height: 85%;" @tap.stop>
      <view class="relative flex gap-4 px-5 pt-5 pb-3 shrink-0">
        <view class="shrink-0 rounded-2xl bg-white p-1 shadow-sm" style="width: 96px; height: 96px; margin-top: -26px;">
          <image
            class="h-full w-full rounded-xl"
            mode="aspectFill"
            :src="selectedImageUrl"
          />
        </view>

        <view class="flex-1 pt-1">
          <view class="flex items-baseline gap-1 mb-1">
            <text class="text-primary" style="font-size: 12px; font-weight: 900;">{{ displayCurrencySymbol }}</text>
            <text class="text-primary" style="font-size: 22px; font-weight: 900;">{{ displayPrice }}</text>
          </view>
          <view class="text-gray-500" style="font-size: 12px; line-height: 1.4;">
            已选：<text class="text-gray-900" style="font-weight: 700;">{{ selectedSkuText }}</text>
          </view>
        </view>

        <view class="absolute top-3 right-3 px-2 py-1" hover-class="opacity-80" @tap="close">
          <text class="text-gray-400" style="font-size: 18px; font-weight: 900;">×</text>
        </view>
      </view>

      <scroll-view scroll-y class="flex-1 px-5 pb-4">
        <view v-if="specGroups && specGroups.length" class="mt-1 space-y-5">
          <view v-for="g in orderedSpecGroups" :key="g.code">
            <view class="flex items-center justify-between mb-3">
              <text class="text-gray-900" style="font-size: 12px; font-weight: 900;">{{ g.name }}</text>
              <text v-if="g.required" class="text-gray-400" style="font-size: 11px;">必选</text>
            </view>

            <view class="flex flex-wrap" style="gap: 10px;">
              <view
                v-for="opt in g.options"
                :key="opt.code"
                class="rounded-xl border px-4 py-3"
                :style="specOptionStyle(g.code, opt.code)"
                hover-class="opacity-90"
                @tap="onPickSpec(g.code, opt.code)"
              >
                <text :style="specOptionTextStyle(g.code, opt.code)" style="font-size: 12px; font-weight: 900;">
                  {{ opt.name }}
                </text>
              </view>
            </view>
          </view>
        </view>

        <view v-else class="mt-1">
          <view class="flex items-center justify-between mb-3">
            <text class="text-gray-900" style="font-size: 12px; font-weight: 900;">选择 SKU</text>
            <text class="text-gray-400" style="font-size: 11px;">共 {{ skus.length }} 个</text>
          </view>

          <view class="flex flex-wrap" style="gap: 10px;">
            <view
              v-for="s in skus"
              :key="s.id"
              class="rounded-xl border px-4 py-3"
              :style="skuChipStyle(s.id) + (s.sellable === false ? 'opacity: 0.55; border-color: rgba(229, 231, 235, 1); background: rgba(243, 244, 246, 1);' : '')"
              hover-class="opacity-90"
              @tap="selectSkuLocal(s.id)"
            >
              <text :style="skuChipTextStyle(s.id)" style="font-size: 12px; font-weight: 900;">{{ s.code || s.id }}</text>
              <text v-if="s.price != null" class="ml-1" :style="skuChipSubTextStyle(s.id)" style="font-size: 10px; font-weight: 800;">
                {{ formatMoney(s.currency || 'CNY', s.price) }}
              </text>
            </view>
          </view>
        </view>

        <view class="mt-6 flex items-center justify-between border-t pt-5" style="border-color: rgba(243, 244, 246, 1);">
          <view>
            <text class="text-gray-900" style="font-size: 12px; font-weight: 900;">数量</text>
            <view class="text-gray-500 mt-1" style="font-size: 11px;">单次最多 {{ maxQty }} 件</view>
          </view>

          <view class="flex items-center rounded-lg border bg-gray-50 p-1" style="border-color: rgba(243, 244, 246, 1); gap: 6px;">
            <view
              class="flex items-center justify-center rounded-md bg-white shadow-sm"
              :style="decDisabled ? 'opacity: 0.5;' : ''"
              style="width: 36px; height: 36px;"
              hover-class="opacity-80"
              @tap="dec"
            >
              <text class="text-gray-500" style="font-size: 18px; font-weight: 900;">−</text>
            </view>
            <view class="flex items-center justify-center" style="width: 34px;">
              <text class="text-gray-900" style="font-size: 14px; font-weight: 900;">{{ qty }}</text>
            </view>
            <view
              class="flex items-center justify-center rounded-md bg-white shadow-sm"
              :style="incDisabled ? 'opacity: 0.5;' : ''"
              style="width: 36px; height: 36px;"
              hover-class="opacity-80"
              @tap="inc"
            >
              <text class="text-gray-900" style="font-size: 18px; font-weight: 900;">+</text>
            </view>
          </view>
        </view>

        <view style="height: 90px;"></view>
      </scroll-view>

      <view class="absolute bottom-0 left-0 right-0 border-t bg-white px-4 pt-3 safe-pb" style="border-color: rgba(243, 244, 246, 1);">
        <view class="flex" style="gap: 10px;">
          <view
            class="flex-1 rounded-full border"
            style="height: 44px; border-color: rgba(79, 138, 126, 0.30); background: rgba(79, 138, 126, 0.10);"
            hover-class="opacity-90"
            @tap="confirm('cart')"
          >
            <view class="flex h-full items-center justify-center">
              <text class="text-primary" style="font-size: 12px; font-weight: 900;">加入购物车</text>
            </view>
          </view>
          <view
            class="flex-1 rounded-full bg-primary shadow-cta"
            style="height: 44px;"
            hover-class="opacity-90"
            @tap="confirm('buy')"
          >
            <view class="flex h-full items-center justify-center">
              <text class="text-white" style="font-size: 12px; font-weight: 900;">立即购买</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";

type SkuItem = {
  id: string;
  code?: string;
  price?: number | null;
  currency?: string | null;
  imageUrl?: string | null;
  status?: string | null;
  stockQty?: number | null;
  sellable?: boolean | null;
  sellabilityReasons?: string[] | null;
  availableQty?: number | null;
  specSignature?: string | null;
  spec?: Record<string, string> | null;
};

type SpecOption = {
  id: string;
  code: string;
  name: string;
  sortOrder?: number;
  meta?: any;
  status?: string;
};

type SpecGroup = {
  id: string;
  code: string;
  name: string;
  required?: boolean;
  sortOrder?: number;
  options: SpecOption[];
};

const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    skus: SkuItem[];
    specGroups?: SpecGroup[];
    selectedSkuId?: string;
    coverUrl?: string;
    maxQty?: number;
  }>(),
  {
    specGroups: () => [],
    selectedSkuId: "",
    coverUrl: "",
    maxQty: 5,
  },
);

const emit = defineEmits<{
  (e: "update:modelValue", v: boolean): void;
  (e: "update:selectedSkuId", v: string): void;
  (e: "confirm", payload: { action: "cart" | "buy"; skuId: string; qty: number }): void;
}>();

const localSelectedSkuId = ref(props.selectedSkuId || "");
const qty = ref(1);
const localSelectedSpec = ref<Record<string, string>>({});

watch(
  () => props.selectedSkuId,
  (v) => {
    if (typeof v === "string") localSelectedSkuId.value = v;
  },
);

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return;
    hydrateSelectionFromProps();
    qty.value = 1;
  },
  { immediate: true },
);

const eligibleSkus = computed(() => {
  return (Array.isArray(props.skus) ? props.skus : []).filter((s) => {
    if (typeof s.sellable === "boolean") return s.sellable;
    const stock = s.stockQty;
    if (typeof stock === "number" && stock <= 0) return false;
    return true;
  });
});

const orderedSpecGroups = computed(() => {
  const groups = Array.isArray(props.specGroups) ? props.specGroups.slice() : [];
  groups.sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
  return groups;
});

const selectedSku = computed(() => {
  if (localSelectedSkuId.value) {
    const s = props.skus.find((x) => x.id === localSelectedSkuId.value);
    if (s) return s;
  }
  const m = matchSkuBySpec(localSelectedSpec.value);
  return m || null;
});

const selectedImageUrl = computed(() => {
  return selectedSku.value?.imageUrl || props.coverUrl || "https://picsum.photos/seed/powerx-sku/200/200";
});

const displayCurrencySymbol = computed(() => {
  const c = String(selectedSku.value?.currency || "CNY").toUpperCase();
  return c === "CNY" || c === "RMB" ? "¥" : c === "USD" ? "$" : c === "EUR" ? "€" : `${c} `;
});

const displayPrice = computed(() => {
  const p = selectedSku.value?.price;
  if (p == null) return "--";
  return String(p);
});

const selectedSkuText = computed(() => {
  if (props.specGroups && props.specGroups.length) {
    const pieces: string[] = [];
    orderedSpecGroups.value.forEach((g) => {
      const picked = localSelectedSpec.value[g.code];
      if (!picked) return;
      const opt = (g.options || []).find((o) => o.code === picked);
      pieces.push(opt?.name || picked);
    });
    if (pieces.length) return pieces.join(" / ");
  }
  const s = selectedSku.value;
  return s ? s.code || s.id : "请选择 SKU";
});

const decDisabled = computed(() => qty.value <= 1);
const incDisabled = computed(() => qty.value >= props.maxQty);

function close() {
  emit("update:modelValue", false);
}

function selectSkuLocal(id: string) {
  const next = props.skus.find((x) => x.id === id);
  if (next && typeof next.sellable === "boolean" && !next.sellable) {
    const code = (next.sellabilityReasons || [])[0] || "";
    uni.showToast({ title: sellabilityReasonToText(code), icon: "none" });
    return;
  }
  localSelectedSkuId.value = id;
  emit("update:selectedSkuId", id);
  const s = props.skus.find((x) => x.id === id);
  if (s?.spec) {
    localSelectedSpec.value = { ...s.spec };
  }
}

function dec() {
  if (decDisabled.value) return;
  qty.value = Math.max(1, qty.value - 1);
}

function inc() {
  if (incDisabled.value) return;
  qty.value = Math.min(props.maxQty, qty.value + 1);
}

function confirm(action: "cart" | "buy") {
  const skuId = localSelectedSkuId.value || selectedSku.value?.id || "";
  if (!skuId) {
    uni.showToast({ title: "请选择规格", icon: "none" });
    return;
  }
  if (selectedSku.value && typeof selectedSku.value.sellable === "boolean" && !selectedSku.value.sellable) {
    const code = (selectedSku.value.sellabilityReasons || [])[0] || "";
    uni.showToast({ title: sellabilityReasonToText(code), icon: "none" });
    return;
  }
  if (orderedSpecGroups.value.length) {
    const missing = orderedSpecGroups.value.find((g) => g.required && !localSelectedSpec.value[g.code]);
    if (missing) {
      uni.showToast({ title: `请选择${missing.name}`, icon: "none" });
      return;
    }
  }
  emit("confirm", { action, skuId, qty: qty.value });
  close();
}

function skuChipStyle(id: string) {
  const active = id === localSelectedSkuId.value;
  return active
    ? "border-color: #4F8A7E; background: rgba(79, 138, 126, 0.10);"
    : "border-color: rgba(229, 231, 235, 1); background: #fff;";
}

function skuChipTextStyle(id: string) {
  const active = id === localSelectedSkuId.value;
  return active ? "color: #4F8A7E;" : "color: #111827;";
}

function skuChipSubTextStyle(id: string) {
  const active = id === localSelectedSkuId.value;
  return active ? "color: #4F8A7E;" : "color: rgba(17, 24, 39, 0.65);";
}

function hydrateSelectionFromProps() {
  if (props.selectedSkuId) {
    localSelectedSkuId.value = props.selectedSkuId;
    const s = props.skus.find((x) => x.id === props.selectedSkuId);
    if (s?.spec) localSelectedSpec.value = { ...s.spec };
    return;
  }

  const preferred = eligibleSkus.value[0] || props.skus[0];
  if (!preferred) return;
  localSelectedSkuId.value = preferred.id;
  emit("update:selectedSkuId", preferred.id);
  if (preferred.spec) localSelectedSpec.value = { ...preferred.spec };
}

function matchSkuBySpec(spec: Record<string, string>) {
  const entries = Object.entries(spec || {}).filter(([_, v]) => Boolean(v));
  const list = eligibleSkus.value;
  if (!entries.length) return list[0] || null;
  return (
    list.find((sku) => {
      const skuSpec = sku.spec || {};
      return entries.every(([k, v]) => skuSpec?.[k] === v);
    }) || null
  );
}

function isSpecOptionDisabled(groupCode: string, optionCode: string) {
  const list = eligibleSkus.value;
  const next = { ...(localSelectedSpec.value || {}), [groupCode]: optionCode };
  const entries = Object.entries(next).filter(([_, v]) => Boolean(v));
  if (!entries.length) return false;
  const matched = list.some((sku) => {
    const skuSpec = sku.spec || {};
    return entries.every(([k, v]) => skuSpec?.[k] === v);
  });
  return !matched;
}

function onPickSpec(groupCode: string, optionCode: string) {
  if (isSpecOptionDisabled(groupCode, optionCode)) return;
  localSelectedSpec.value = { ...(localSelectedSpec.value || {}), [groupCode]: optionCode };
  const m = matchSkuBySpec(localSelectedSpec.value);
  if (m?.id) {
    localSelectedSkuId.value = m.id;
    emit("update:selectedSkuId", m.id);
  }
}

function specOptionStyle(groupCode: string, optionCode: string) {
  const disabled = isSpecOptionDisabled(groupCode, optionCode);
  const active = localSelectedSpec.value?.[groupCode] === optionCode;
  if (disabled) return "border-color: rgba(229, 231, 235, 1); background: rgba(243, 244, 246, 1); opacity: 0.55;";
  if (active) return "border-color: #4F8A7E; background: rgba(79, 138, 126, 0.10);";
  return "border-color: rgba(229, 231, 235, 1); background: #fff;";
}

function specOptionTextStyle(groupCode: string, optionCode: string) {
  const disabled = isSpecOptionDisabled(groupCode, optionCode);
  const active = localSelectedSpec.value?.[groupCode] === optionCode;
  if (disabled) return "color: rgba(17, 24, 39, 0.45);";
  return active ? "color: #4F8A7E;" : "color: #111827;";
}

function formatMoney(currency: string, price: number) {
  const c = String(currency || "").toUpperCase();
  const symbol = c === "CNY" || c === "RMB" ? "¥" : c === "USD" ? "$" : c === "EUR" ? "€" : `${c} `;
  return `${symbol}${price}`;
}

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
</script>

<style scoped>
.safe-pb {
  padding-bottom: env(safe-area-inset-bottom);
}

@keyframes slide-up {
  0% {
    transform: translateY(100%);
  }
  100% {
    transform: translateY(0);
  }
}

.animate-slide-up {
  animation: slide-up 0.28s cubic-bezier(0.16, 1, 0.3, 1);
}
</style>
