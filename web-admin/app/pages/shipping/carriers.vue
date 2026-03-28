<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">物流承运商</h1>
        <p class="text-gray-500 dark:text-gray-400">
          维护合作承运商信息、服务范围及履约表现。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton variant="ghost" color="neutral" icon="i-heroicons-arrow-down-tray">
          导出
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">新增承运商</UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <UCard v-for="card in summaryCards" :key="card.title">
        <div class="space-y-1">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
          <div class="text-3xl font-semibold text-gray-900 dark:text-white">
            {{ card.value }}
          </div>
          <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
            {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% 较上周
          </p>
        </div>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between gap-2">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">智能分单容量计划</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">维护承运商配额，供自动分单使用。</p>
          </div>
          <UButton size="sm" color="primary" :loading="planSubmitting" @click="createCapacityPlan">
            新增容量计划
          </UButton>
        </div>
      </template>
      <div class="grid gap-2 md:grid-cols-5">
        <UInput v-model="planForm.name" placeholder="计划名称" />
        <UInput v-model="planForm.carrierId" placeholder="承运商ID" />
        <UInput v-model="planForm.warehouseId" placeholder="仓库ID（可选）" />
        <UInput v-model="planForm.destinationZone" placeholder="目的区域（可选）" />
        <UInput v-model.number="planForm.dailyCapacity" type="number" placeholder="日容量" />
      </div>
      <ul class="mt-3 space-y-2 text-xs">
        <li v-for="item in capacityPlans" :key="item.id" class="rounded border border-gray-200 p-2 dark:border-gray-800">
          {{ item.name }} · {{ item.carrierId }} · 容量 {{ item.usedCapacity }}/{{ item.dailyCapacity }}
        </li>
        <li v-if="!capacityPlans.length" class="text-gray-500">暂无计划</li>
      </ul>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">承运商列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">支持按区域与状态筛选。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-56"
              placeholder="搜索承运商/联系人"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="statusFilter"
              class="w-40"
              :options="statusOptions"
              placeholder="服务状态"
            />
            <USelect
              v-model="coverageFilter"
              class="w-40"
              :options="coverageOptions"
              placeholder="覆盖区域"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredCarriers">
        <template #status-cell="{ getValue }">
          <UBadge :color="getStatusMeta(getValue()).color" variant="subtle">
            {{ getStatusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #rating-cell="{ getValue }">
          <div class="flex items-center gap-1 text-amber-500">
            <UIcon
              v-for="index in 5"
              :key="index"
              :name="index <= Math.round(getValue()) ? 'i-heroicons-star-solid' : 'i-heroicons-star'"
              class="h-4 w-4"
            />
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ getValue().toFixed(1) }}</span>
          </div>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost" color="neutral" @click="openRoutingPreview(row.original)">
              联合路由仿真
            </UButton>
            <UButton size="xs" variant="ghost" @click="switchProvider(row.original)">
              切换 Provider
            </UButton>
            <UButton size="xs" variant="ghost" color="primary" @click="testProvider(row.original)">
              连通性测试
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UModal v-model:open="routingPreviewOpen" title="联合路由仿真">
      <template #body>
        <div class="space-y-3">
          <div class="grid gap-2 sm:grid-cols-2">
            <UInput v-model="routingForm.warehouseId" placeholder="仓库ID（可选）" />
            <UInput v-model="routingForm.destinationZone" placeholder="目的区域（如 CN-EAST）" />
            <UInput v-model="routingForm.serviceCode" placeholder="服务编码（默认 std）" />
            <UInput v-model.number="routingForm.weight" type="number" step="0.1" placeholder="重量(kg)" />
          </div>
          <div class="text-xs text-gray-500">
            当前偏好承运商：{{ routingForm.preferredCarrierName || "未指定" }}
          </div>
          <UButton color="primary" :loading="routingPreviewLoading" @click="previewRouting">
            执行仿真
          </UButton>
          <UCard v-if="routingPreviewResult">
            <p class="text-sm">策略：{{ routingPreviewResult.strategy }}（{{ routingPreviewResult.reason }}）</p>
            <p class="text-sm">结果承运商：{{ routingPreviewResult.carrierName }}</p>
            <p class="text-xs text-gray-500">{{ routingPreviewResult.explain || "-" }}</p>
            <ul class="mt-2 space-y-1 text-xs">
              <li v-for="item in routingPreviewResult.candidates || []" :key="item.carrierId">
                {{ item.carrierName }} · score={{ Number(item.finalScore || 0).toFixed(2) }}
              </li>
            </ul>
          </UCard>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useLogisticsApi } from "~/composables/api";

definePageMeta({
  name: "shipping-carriers",
});

type CarrierStatus = "active" | "monitor" | "suspended";

type Carrier = {
  id: string;
  code: string;
  type: string;
  name: string;
  provider: string;
  coverage: string;
  contact: string;
  phone: string;
  status: CarrierStatus;
  avgTime: string;
  onTimeRate: number;
  rating: number;
  rawConfig: Record<string, any>;
  rawCapabilities: Record<string, any>;
};

const logisticsApi = useLogisticsApi();
const carriers = ref<Carrier[]>([]);
const capacityPlans = ref<any[]>([]);
const planSubmitting = ref(false);
const routingPreviewOpen = ref(false);
const routingPreviewLoading = ref(false);
const routingPreviewResult = ref<any>(null);
const routingForm = reactive({
  warehouseId: "",
  destinationZone: "GLOBAL",
  serviceCode: "std",
  weight: 1,
  preferredCarrierId: "",
  preferredCarrierName: "",
});
const planForm = reactive({
  name: "",
  carrierId: "",
  warehouseId: "",
  destinationZone: "",
  dailyCapacity: 50,
});

const keyword = ref("");
const statusFilter = ref<CarrierStatus | "">("");
const coverageFilter = ref("");

const normalizeStatus = (status: string): CarrierStatus => {
  const v = String(status || "").toLowerCase();
  if (v === "active") return "active";
  if (v === "disabled" || v === "inactive") return "suspended";
  return "monitor";
};

const loadCarriers = async () => {
  const rows = await logisticsApi.listCarriers();
  carriers.value = rows.map((row) => {
    const cfg = (row.config || {}) as Record<string, any>;
    const cap = (row.capabilities || {}) as Record<string, any>;
    const onTimeRate = Number(cfg.onTimeRate ?? cfg.on_time_rate ?? 0);
    const rating = Number(cfg.rating ?? 0);
    const avgHours = Number(cfg.avgHours ?? cfg.avg_hours ?? 0);
    return {
      id: row.id,
      code: row.code,
      type: row.type,
      name: row.name || row.code,
      provider: String(cfg.provider || row.type || "self"),
      coverage: String(cap.coverage || row.type || "未配置"),
      contact: row.contactName || "-",
      phone: row.contactPhone || "-",
      status: normalizeStatus(row.status),
      avgTime: avgHours > 0 ? `${(avgHours / 24).toFixed(1)} 天` : "-",
      onTimeRate,
      rating,
      rawConfig: cfg,
      rawCapabilities: cap,
    };
  });
  capacityPlans.value = await logisticsApi.listAllocationPlans({ limit: 50 });
};

onMounted(() => {
  loadCarriers();
});

const createCapacityPlan = async () => {
  planSubmitting.value = true;
  try {
    await logisticsApi.upsertAllocationPlan({
      name: planForm.name || `计划-${Date.now()}`,
      carrier_id: planForm.carrierId,
      warehouse_id: planForm.warehouseId || undefined,
      destination_zone: planForm.destinationZone || undefined,
      daily_capacity: Number(planForm.dailyCapacity || 0),
      status: "active",
    });
    await loadCarriers();
  } finally {
    planSubmitting.value = false;
  }
};

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "合作中", value: "active" },
  { label: "观察中", value: "monitor" },
  { label: "暂停", value: "suspended" },
];

const coverageOptions = computed(() =>
  [{ label: "全部区域", value: "" }].concat(
    Array.from(new Set(carriers.value.map((carrier) => carrier.coverage))).map((area) => ({
      label: area,
      value: area,
    })),
  ),
);

const columns = computed<TableColumn<Carrier>[]>(() => [
  { accessorKey: "name", header: "承运商" },
  { accessorKey: "provider", header: "Provider" },
  { accessorKey: "coverage", header: "覆盖范围" },
  { accessorKey: "avgTime", header: "平均时效" },
  {
    accessorKey: "onTimeRate",
    header: "准时率",
    cell: ({ getValue }) => `${Number(getValue() || 0).toFixed(1)}%`,
  },
  { accessorKey: "rating", header: "评分" },
  { accessorKey: "status", header: "状态" },
  {
    accessorKey: "contact",
    header: "联系人",
    cell: ({ row }) => `${row.original.contact} / ${row.original.phone}`,
  },
  { id: "actions", header: "操作" },
]);

const providerCycle = ["self", "sf", "jd", "cainiao", "dhl", "other"];

const switchProvider = async (carrier: Carrier) => {
  const idx = providerCycle.indexOf(String(carrier.provider || "self"));
  const nextProvider = providerCycle[(idx + 1) % providerCycle.length];
  await logisticsApi.upsertCarrier({
    id: carrier.id,
    name: carrier.name,
    code: carrier.code,
    type: carrier.type || "self",
    status: carrier.status === "suspended" ? "disabled" : "active",
    contact_name: carrier.contact === "-" ? "" : carrier.contact,
    contact_phone: carrier.phone === "-" ? "" : carrier.phone,
    capabilities: carrier.rawCapabilities,
    config: {
      ...carrier.rawConfig,
      provider: nextProvider,
    },
  });
  await loadCarriers();
};

const testProvider = async (carrier: Carrier) => {
  await logisticsApi.testCarrier(carrier.id);
};

const openRoutingPreview = (carrier: Carrier) => {
  routingForm.preferredCarrierId = carrier.id;
  routingForm.preferredCarrierName = carrier.name;
  routingPreviewResult.value = null;
  routingPreviewOpen.value = true;
};

const previewRouting = async () => {
  routingPreviewLoading.value = true;
  try {
    routingPreviewResult.value = await logisticsApi.simulateRoutingOptimizer({
      request_key: `carrier-sim#${Date.now()}`,
      warehouse_id: routingForm.warehouseId || undefined,
      destination_zone: routingForm.destinationZone || undefined,
      weight: Number(routingForm.weight || 0),
      preferred_carrier_id: routingForm.preferredCarrierId || undefined,
    });
  } finally {
    routingPreviewLoading.value = false;
  }
};

const filteredCarriers = computed(() =>
  carriers.value.filter((carrier) => {
    const matchesKeyword =
      !keyword.value ||
      carrier.name.includes(keyword.value) ||
      carrier.contact.includes(keyword.value);
    const matchesStatus = !statusFilter.value || carrier.status === statusFilter.value;
    const matchesCoverage = !coverageFilter.value || carrier.coverage === coverageFilter.value;
    return matchesKeyword && matchesStatus && matchesCoverage;
  }),
);

const getStatusMeta = (status: CarrierStatus | "") => {
  switch (status) {
    case "active":
      return { label: "合作中", color: "success" as const };
    case "monitor":
      return { label: "观察中", color: "warning" as const };
    case "suspended":
      return { label: "暂停", color: "neutral" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const summaryCards = computed(() => {
  const total = carriers.value.length || 1;
  return [
    {
      title: "合作承运商",
      value: carriers.value.filter((item) => item.status === "active").length,
      trend: 0,
    },
    {
      title: "平均准时率",
      value:
        (
          carriers.value.reduce((sum, carrier) => sum + carrier.onTimeRate, 0) /
          total
        ).toFixed(1) + "%",
      trend: 0,
    },
    {
      title: "平均评分",
      value: (
        carriers.value.reduce((sum, carrier) => sum + carrier.rating, 0) / total
      ).toFixed(1),
      trend: 0,
    },
  ];
});
</script>
