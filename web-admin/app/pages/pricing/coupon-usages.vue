<template>
  <div class="p-6 space-y-6">
    <div>
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">优惠券资产与流水查询</h1>
      <p class="text-sm text-gray-500 dark:text-gray-400">
        支持按券码、订单号、用户ID排查优惠券生命周期。
      </p>
    </div>

    <UCard>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UFormField label="券码">
          <UInput v-model="filters.couponCode" placeholder="PROMO" />
        </UFormField>
        <UFormField label="用户ID">
          <UInput v-model="filters.userId" placeholder="u-1" />
        </UFormField>
        <UFormField label="订单ID">
          <UInput v-model="filters.orderId" placeholder="ord-1" />
        </UFormField>
        <UFormField label="资产状态">
          <USelect v-model="filters.status" :items="statusItems" />
        </UFormField>
      </div>
      <div class="mt-4 flex gap-2">
        <UButton color="primary" icon="i-heroicons-magnifying-glass" :loading="loading" @click="refresh">查询</UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <div class="font-medium">资产列表</div>
      </template>
      <UTable :data="assets" :columns="assetColumns" :loading="loading" />
      <div class="mt-3 text-sm text-gray-500">总数：{{ assetTotal }}</div>
    </UCard>

    <UCard>
      <template #header>
        <div class="font-medium">流水列表</div>
      </template>
      <UTable :data="logs" :columns="logColumns" :loading="loading" />
      <div class="mt-3 text-sm text-gray-500">总数：{{ logTotal }}</div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useCouponsApi, type CouponAsset, type CouponUsageLog } from "~/composables/api/useCoupons";

const api = useCouponsApi();
const toast = useToast();
const loading = ref(false);
const assets = ref<CouponAsset[]>([]);
const logs = ref<CouponUsageLog[]>([]);
const assetTotal = ref(0);
const logTotal = ref(0);

const filters = reactive({
  couponCode: "",
  userId: "",
  orderId: "",
  status: "",
});

const statusItems = [
  { label: "全部", value: "" },
  { label: "available", value: "available" },
  { label: "reserved", value: "reserved" },
  { label: "redeemed", value: "redeemed" },
  { label: "expired", value: "expired" },
  { label: "refunded", value: "refunded" },
];

const assetColumns: TableColumn<CouponAsset>[] = [
  { accessorKey: "coupon_code", header: "券码" },
  { accessorKey: "user_id", header: "用户" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "reserved_order_id", header: "订单ID" },
  { accessorKey: "updated_at", header: "更新时间" },
];

const logColumns: TableColumn<CouponUsageLog>[] = [
  { accessorKey: "coupon_code", header: "券码" },
  { accessorKey: "action", header: "动作" },
  { accessorKey: "action_reason", header: "原因" },
  { accessorKey: "order_id", header: "订单ID" },
  { accessorKey: "created_by", header: "操作人" },
  { accessorKey: "created_at", header: "时间" },
];

const refresh = async () => {
  loading.value = true;
  try {
    const [assetRes, logRes] = await Promise.all([
      api.listAssets({
        couponCode: filters.couponCode || undefined,
        userId: filters.userId || undefined,
        orderId: filters.orderId || undefined,
        status: filters.status || undefined,
      }),
      api.listUsageLogs({
        couponCode: filters.couponCode || undefined,
        userId: filters.userId || undefined,
        orderId: filters.orderId || undefined,
      }),
    ]);
    assets.value = assetRes.items || [];
    logs.value = logRes.items || [];
    assetTotal.value = Number(assetRes.total || 0);
    logTotal.value = Number(logRes.total || 0);
  } catch (error: any) {
    toast.add({ color: "error", title: "查询失败", description: error?.message || "请求失败" });
  } finally {
    loading.value = false;
  }
};

onMounted(refresh);
</script>
