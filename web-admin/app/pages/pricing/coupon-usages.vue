<template>
  <div class="space-y-6 p-6">
    <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">券资产与流水</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          按模板、用户、券码、订单追踪发放、锁定、核销、释放、退款全过程。
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton color="neutral" variant="soft" icon="i-heroicons-arrow-path" :loading="loading" @click="refresh">
          刷新
        </UButton>
        <UButton color="primary" variant="soft" icon="i-heroicons-ticket" to="/pricing/coupons">
          回到模板
        </UButton>
      </div>
    </div>

    <UCard>
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-6">
        <UFormField label="模板ID" class="xl:col-span-2">
          <UInput v-model="filters.templateId" placeholder="按发券模板追踪" />
        </UFormField>
        <UFormField label="券码">
          <UInput v-model="filters.couponCode" placeholder="PROMO" />
        </UFormField>
        <UFormField label="用户ID">
          <UInput v-model="filters.userId" placeholder="客户 ID" />
        </UFormField>
        <UFormField label="订单ID">
          <UInput v-model="filters.orderId" placeholder="订单 ID" />
        </UFormField>
        <UFormField label="资产状态">
          <USelect v-model="filters.status" :items="statusItems" />
        </UFormField>
      </div>
      <div class="mt-4 flex flex-wrap items-center gap-2">
        <UButton color="primary" icon="i-heroicons-magnifying-glass" :loading="loading" @click="refresh">
          查询
        </UButton>
        <UButton color="neutral" variant="ghost" icon="i-heroicons-x-mark" @click="resetFilters">
          清空
        </UButton>
        <span v-if="filters.templateId" class="text-sm text-gray-500 dark:text-gray-400">
          当前按模板追踪：{{ filters.templateId }}
        </span>
      </div>
    </UCard>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
      <UCard v-for="item in summaryCards" :key="item.label">
        <div class="flex items-center justify-between gap-3">
          <div>
            <div class="text-sm text-gray-500 dark:text-gray-400">{{ item.label }}</div>
            <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ item.value }}</div>
          </div>
          <UIcon :name="item.icon" class="h-8 w-8 text-primary-500" />
        </div>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div class="font-medium">发放结果概览</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">
              {{ issuedUserCount }} 个客户，{{ assetTotal }} 张券；可用 {{ statusCounts.available }}，已锁定 {{ statusCounts.reserved }}，已核销 {{ statusCounts.redeemed }}。
            </div>
          </div>
          <UBadge color="neutral" variant="soft">总数 {{ assetTotal }}</UBadge>
        </div>
      </template>

      <div class="overflow-x-auto">
        <UTable :data="assetRows" :columns="assetColumns" :loading="loading" class="min-w-[1100px]" />
      </div>
    </UCard>

    <div class="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.15fr)_minmax(360px,0.85fr)]">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between gap-3">
            <div>
              <div class="font-medium">流水列表</div>
              <div class="text-sm text-gray-500 dark:text-gray-400">按时间倒序展示生命周期动作。</div>
            </div>
            <UBadge color="neutral" variant="soft">总数 {{ logTotal }}</UBadge>
          </div>
        </template>

        <div class="overflow-x-auto">
          <UTable :data="logRows" :columns="logColumns" :loading="loading" class="min-w-[960px]" />
        </div>
      </UCard>

      <UCard>
        <template #header>
          <div>
            <div class="font-medium">当前券时间线</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">
              {{ selectedAsset ? selectedAsset.coupon_code : "从资产列表选择一张券" }}
            </div>
          </div>
        </template>

        <div v-if="selectedAsset" class="space-y-4">
          <div class="rounded-md border border-gray-200 p-4 dark:border-gray-800">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="truncate font-medium text-gray-900 dark:text-white">{{ selectedAsset.coupon_code }}</div>
                <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  用户 {{ selectedAsset.user_id }} · 模板 {{ selectedAsset.template_id }}
                </div>
              </div>
              <UBadge :color="statusColor(selectedAsset.status)" variant="soft">
                {{ statusLabel(selectedAsset.status) }}
              </UBadge>
            </div>
          </div>

          <div v-if="selectedAssetLogs.length" class="space-y-3">
            <div v-for="log in selectedAssetLogs" :key="log.id" class="flex gap-3">
              <div class="mt-1 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400">
                <UIcon :name="actionIcon(log.action)" class="h-4 w-4" />
              </div>
              <div class="min-w-0 flex-1 border-b border-gray-100 pb-3 dark:border-gray-800">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-medium text-gray-900 dark:text-white">{{ actionLabel(log.action) }}</span>
                  <UBadge color="neutral" variant="soft">{{ log.action_reason || "-" }}</UBadge>
                </div>
                <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ formatDateTime(log.created_at) }}
                  <span v-if="log.order_id"> · 订单 {{ log.order_id }}</span>
                  <span v-if="log.created_by"> · {{ log.created_by }}</span>
                </div>
              </div>
            </div>
          </div>
          <UAlert
            v-else
            color="warning"
            variant="soft"
            icon="i-heroicons-exclamation-triangle"
            title="还没有查到这张券的流水"
            description="新发券会写入 issue 流水；历史数据如果是在修复前发放，可能只有资产记录。"
          />
        </div>
        <UAlert
          v-else
          color="neutral"
          variant="soft"
          icon="i-heroicons-cursor-arrow-rays"
          title="选择一张券查看明细"
          description="资产列表里点“追踪”，右侧会展示这张券完整生命周期。"
        />
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { h, resolveComponent } from "vue";
import { useCouponsApi, type CouponAsset, type CouponUsageLog } from "~/composables/api/useCoupons";

type AssetRow = CouponAsset & {
  status_text: string;
  valid_period: string;
};

type LogRow = CouponUsageLog & {
  action_text: string;
  created_time: string;
};

const UBadge = resolveComponent("UBadge");
const UButton = resolveComponent("UButton");

const api = useCouponsApi();
const route = useRoute();
const toast = useToast();
const loading = ref(false);
const assets = ref<CouponAsset[]>([]);
const logs = ref<CouponUsageLog[]>([]);
const assetTotal = ref(0);
const logTotal = ref(0);
const selectedAssetId = ref("");

const filters = reactive({
  templateId: "",
  couponCode: "",
  userId: "",
  orderId: "",
  status: "all",
});

const statusItems = [
  { label: "全部", value: "all" },
  { label: "可用", value: "available" },
  { label: "已锁定", value: "reserved" },
  { label: "已核销", value: "redeemed" },
  { label: "已过期", value: "expired" },
  { label: "已退款", value: "refunded" },
];

const statusTextMap: Record<string, string> = {
  available: "可用",
  reserved: "已锁定",
  redeemed: "已核销",
  expired: "已过期",
  refunded: "已退款",
};

const actionTextMap: Record<string, string> = {
  issue: "发放",
  reserve: "锁定",
  redeem: "核销",
  release: "释放",
  refund: "退款",
  expire: "过期",
};

const statusLabel = (status?: string) => statusTextMap[String(status || "")] || status || "-";
const actionLabel = (action?: string) => actionTextMap[String(action || "")] || action || "-";

const statusColor = (status?: string) => {
  switch (status) {
    case "available":
      return "success";
    case "reserved":
      return "warning";
    case "redeemed":
      return "primary";
    case "expired":
      return "neutral";
    case "refunded":
      return "info";
    default:
      return "neutral";
  }
};

const actionIcon = (action?: string) => {
  switch (action) {
    case "issue":
      return "i-heroicons-paper-airplane";
    case "reserve":
      return "i-heroicons-lock-closed";
    case "redeem":
      return "i-heroicons-check-circle";
    case "release":
      return "i-heroicons-lock-open";
    case "refund":
      return "i-heroicons-arrow-uturn-left";
    default:
      return "i-heroicons-clock";
  }
};

const formatDateTime = (value?: string | null) => {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("zh-CN", { hour12: false });
};

const formatPeriod = (from?: string | null, to?: string | null) => {
  if (!from && !to) return "-";
  return `${formatDateTime(from)} 至 ${formatDateTime(to)}`;
};

const issuedUserCount = computed(() => new Set(assets.value.map((item) => item.user_id).filter(Boolean)).size);

const statusCounts = computed(() => {
  const counts: Record<string, number> = { available: 0, reserved: 0, redeemed: 0, expired: 0, refunded: 0 };
  for (const asset of assets.value) {
    const key = asset.status || "unknown";
    counts[key] = (counts[key] || 0) + 1;
  }
  return counts;
});

const selectedAsset = computed(() => assets.value.find((item) => item.id === selectedAssetId.value) || assets.value[0]);

const selectedAssetLogs = computed(() => {
  const assetID = selectedAsset.value?.id;
  if (!assetID) return [];
  return logs.value.filter((item) => item.asset_id === assetID);
});

const summaryCards = computed(() => [
  { label: "发放客户", value: issuedUserCount.value, icon: "i-heroicons-users" },
  { label: "券资产", value: assetTotal.value, icon: "i-heroicons-ticket" },
  { label: "可用", value: statusCounts.value.available || 0, icon: "i-heroicons-check-badge" },
  { label: "已核销", value: statusCounts.value.redeemed || 0, icon: "i-heroicons-banknotes" },
]);

const assetRows = computed<AssetRow[]>(() =>
  assets.value.map((item) => ({
    ...item,
    status_text: statusLabel(item.status),
    valid_period: formatPeriod(item.valid_from, item.valid_to),
  })),
);

const logRows = computed<LogRow[]>(() =>
  logs.value.map((item) => ({
    ...item,
    action_text: actionLabel(item.action),
    created_time: formatDateTime(item.created_at),
  })),
);

const assetColumns: TableColumn<AssetRow>[] = [
  {
    accessorKey: "coupon_code",
    header: "券码",
    cell: ({ row }) =>
      h("div", { class: "min-w-[160px]" }, [
        h("div", { class: "font-medium text-gray-900 dark:text-white" }, row.original.coupon_code),
        h("div", { class: "text-xs text-gray-500 dark:text-gray-400" }, row.original.id),
      ]),
  },
  { accessorKey: "template_id", header: "模板ID" },
  { accessorKey: "user_id", header: "客户" },
  {
    accessorKey: "status_text",
    header: "状态",
    cell: ({ row }) =>
      h(UBadge, { color: statusColor(row.original.status), variant: "soft" }, () => row.original.status_text),
  },
  { accessorKey: "reserved_order_id", header: "关联订单" },
  { accessorKey: "valid_period", header: "有效期" },
  {
    accessorKey: "updated_at",
    header: "更新时间",
    cell: ({ row }) => formatDateTime(row.original.updated_at),
  },
  {
    id: "actions",
    header: "操作",
    cell: ({ row }) =>
      h(
        UButton,
        {
          size: "xs",
          color: selectedAssetId.value === row.original.id ? "primary" : "neutral",
          variant: selectedAssetId.value === row.original.id ? "solid" : "soft",
          icon: "i-heroicons-map",
          onClick: () => {
            selectedAssetId.value = row.original.id;
          },
        },
        () => "追踪",
      ),
  },
];

const logColumns: TableColumn<LogRow>[] = [
  { accessorKey: "coupon_code", header: "券码" },
  { accessorKey: "user_id", header: "客户" },
  {
    accessorKey: "action_text",
    header: "动作",
    cell: ({ row }) =>
      h("div", { class: "flex items-center gap-2" }, [
        h(UBadge, { color: "primary", variant: "soft" }, () => row.original.action_text),
        h("span", { class: "text-xs text-gray-500 dark:text-gray-400" }, row.original.action_reason || "-"),
      ]),
  },
  { accessorKey: "order_id", header: "订单ID" },
  { accessorKey: "created_by", header: "操作人" },
  { accessorKey: "created_time", header: "时间" },
];

const refresh = async () => {
  loading.value = true;
  try {
    const params = {
      templateId: filters.templateId || undefined,
      couponCode: filters.couponCode || undefined,
      userId: filters.userId || undefined,
      orderId: filters.orderId || undefined,
      status: filters.status === "all" ? undefined : filters.status,
      page: 1,
      pageSize: 100,
    };
    const [assetRes, logRes] = await Promise.all([
      api.listAssets(params),
      api.listUsageLogs({
        templateId: filters.templateId || undefined,
        couponCode: filters.couponCode || undefined,
        userId: filters.userId || undefined,
        orderId: filters.orderId || undefined,
        page: 1,
        pageSize: 100,
      }),
    ]);
    assets.value = assetRes.items || [];
    logs.value = logRes.items || [];
    assetTotal.value = Number(assetRes.total || 0);
    logTotal.value = Number(logRes.total || 0);
    if (!assets.value.some((item) => item.id === selectedAssetId.value)) {
      selectedAssetId.value = assets.value[0]?.id || "";
    }
  } catch (error: any) {
    toast.add({ color: "error", title: "查询失败", description: error?.message || "请求失败" });
  } finally {
    loading.value = false;
  }
};

const resetFilters = () => {
  filters.templateId = "";
  filters.couponCode = "";
  filters.userId = "";
  filters.orderId = "";
  filters.status = "all";
  selectedAssetId.value = "";
  refresh();
};

const applyRouteQuery = () => {
  filters.templateId = typeof route.query.templateId === "string" ? route.query.templateId : "";
  filters.couponCode = typeof route.query.couponCode === "string" ? route.query.couponCode : "";
  filters.userId = typeof route.query.userId === "string" ? route.query.userId : "";
  filters.orderId = typeof route.query.orderId === "string" ? route.query.orderId : "";
  filters.status = typeof route.query.status === "string" ? route.query.status : "all";
};

onMounted(() => {
  applyRouteQuery();
  refresh();
});
</script>
