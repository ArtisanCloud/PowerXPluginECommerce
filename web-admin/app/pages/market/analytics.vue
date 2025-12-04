<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        {{ $t("nav.analytics") }}
      </h1>
      <div class="flex space-x-2">
        <USelect
          v-model="selectedPeriod"
          :options="timePeriods"
          option-attribute="label"
          value-attribute="value"
          placeholder="选择时间段"
        />
        <UButton
          color="blue"
          variant="soft"
          icon="i-heroicons-arrow-down-tray"
          @click="exportReport"
        >
          导出报告
        </UButton>
      </div>
    </div>

    <!-- 指标卡 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">总销售额</p>
            <p class="text-2xl font-bold text-green-600">
              {{ formatCNY(kpis.sales) }}
            </p>
            <p
              class="text-xs"
              :class="kpis.salesWow >= 0 ? 'text-green-500' : 'text-red-500'"
            >
              {{ signedPercent(kpis.salesWow) }} 较上期
            </p>
          </div>
          <UIcon
            name="i-heroicons-currency-yen"
            class="w-8 h-8 text-green-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">订单数量</p>
            <p class="text-2xl font-bold text-blue-600">
              {{ kpis.orders?.toLocaleString() || '0' }}
            </p>
            <p
              class="text-xs"
              :class="kpis.ordersWow >= 0 ? 'text-blue-500' : 'text-red-500'"
            >
              {{ signedPercent(kpis.ordersWow) }} 较上期
            </p>
          </div>
          <UIcon
            name="i-heroicons-shopping-cart"
            class="w-8 h-8 text-blue-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">客户数量</p>
            <p class="text-2xl font-bold text-purple-600">
              {{ kpis.customers?.toLocaleString() || '0' }}
            </p>
            <p
              class="text-xs"
              :class="
                kpis.customersWow >= 0 ? 'text-purple-500' : 'text-red-500'
              "
            >
              {{ signedPercent(kpis.customersWow) }} 较上期
            </p>
          </div>
          <UIcon name="i-heroicons-users" class="w-8 h-8 text-purple-500" />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">转化率</p>
            <p class="text-2xl font-bold text-orange-600">
              {{ percent(kpis.cv) }}
            </p>
            <p
              class="text-xs"
              :class="kpis.cvWow >= 0 ? 'text-green-500' : 'text-red-500'"
            >
              {{ signedPercent(kpis.cvWow) }} 较上期
            </p>
          </div>
          <UIcon name="i-heroicons-chart-pie" class="w-8 h-8 text-orange-500" />
        </div>
      </UCard>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
      <!-- 销售趋势图（占位） -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">
            销售趋势（{{ selectedPeriod }}）
          </h3>
        </template>
        <div
          class="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded"
        >
          <p class="text-gray-500">销售趋势图表区域</p>
        </div>
      </UCard>

      <!-- 热销 TOP10 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">
            热销商品 TOP10（{{ selectedPeriod }}）
          </h3>
        </template>
        <div class="space-y-3">
          <div
            v-for="(product, index) in topProductsComputed"
            :key="product.id"
            class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded"
          >
            <div class="flex items-center space-x-3">
              <div
                class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold"
                :class="getRankColor(index + 1)"
              >
                {{ index + 1 }}
              </div>
              <span class="font-medium">{{ product.name }}</span>
            </div>
            <div class="text-right">
              <p class="font-semibold">{{ product.sales?.toLocaleString() || '0' }}</p>
              <p class="text-xs text-gray-500">销量</p>
            </div>
          </div>
        </div>
      </UCard>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 用户来源 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">
            用户来源分析（{{ selectedPeriod }}）
          </h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="source in userSourcesComputed"
            :key="source.name"
            class="flex items-center justify-between"
          >
            <div class="flex items-center space-x-3">
              <div
                class="w-4 h-4 rounded"
                :style="{ backgroundColor: source.color }"
              ></div>
              <span>{{ source.name }}</span>
            </div>
            <div class="text-right">
              <span class="font-semibold">{{ source.percentage }}%</span>
              <p class="text-xs text-gray-500">
                {{ source.users?.toLocaleString() || '0' }} 用户
              </p>
            </div>
          </div>
        </div>
      </UCard>

      <!-- 地区销售 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">
            地区销售分布（{{ selectedPeriod }}）
          </h3>
        </template>
        <div class="space-y-3">
          <div
            v-for="region in regionSalesComputed"
            :key="region.name"
            class="flex items-center justify-between"
          >
            <span>{{ region.name }}</span>
            <div class="flex items-center space-x-2">
              <div class="w-20 bg-gray-200 rounded-full h-2">
                <div
                  class="bg-blue-500 h-2 rounded-full"
                  :style="{ width: region.percentage + '%' }"
                ></div>
              </div>
              <span class="text-sm font-medium">{{
                formatCNY(region.sales)
              }}</span>
            </div>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n();

// 选择的时间段
const selectedPeriod = ref<"今日" | "本周" | "本月" | "本季度" | "本年">(
  "本月"
);

// 时间段选项
const timePeriods = [
  { label: "今日", value: "今日" },
  { label: "本周", value: "本周" },
  { label: "本月", value: "本月" },
  { label: "本季度", value: "本季度" },
  { label: "本年", value: "本年" },
];

// ===== 数据（源） =====
const topProducts = ref([
  { id: 1, name: "iPhone 15 Pro", sales: 1234 },
  { id: 2, name: "MacBook Air M3", sales: 987 },
  { id: 3, name: "AirPods Pro 2", sales: 856 },
  { id: 4, name: "iPad Pro 12.9", sales: 743 },
  { id: 5, name: "Apple Watch Series 9", sales: 652 },
]);

const userSources = ref([
  { name: "直接访问", percentage: 35, users: 1456, color: "#3B82F6" },
  { name: "搜索引擎", percentage: 28, users: 1167, color: "#10B981" },
  { name: "社交媒体", percentage: 20, users: 833, color: "#F59E0B" },
  { name: "邮件营销", percentage: 12, users: 500, color: "#EF4444" },
  { name: "其他", percentage: 5, users: 208, color: "#6B7280" },
]);

// 注意：sales 用 number 存储
const regionSales = ref([
  { name: "北京", sales: 234567, percentage: 85 },
  { name: "上海", sales: 198432, percentage: 72 },
  { name: "广州", sales: 156789, percentage: 58 },
  { name: "深圳", sales: 134567, percentage: 49 },
  { name: "杭州", sales: 98765, percentage: 36 },
]);

// KPI（基准数据）
const baseKpis = {
  sales: 1234567,
  orders: 8456,
  customers: 2345,
  cv: 0.032, // 3.2%
  salesWow: 0.125, // 同比/环比演示
  ordersWow: 0.082,
  customersWow: 0.153,
  cvWow: -0.005,
};

// ===== 基础格式化工具 =====
const formatCNY = (n: number) =>
  new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "CNY",
    maximumFractionDigits: 0,
  }).format(n);

const percent = (v: number, digits = 1) => `${(v * 100).toFixed(digits)}%`;

const signedPercent = (v: number, digits = 1) =>
  `${v >= 0 ? "+" : ""}${(v * 100).toFixed(digits)}%`;

// ===== “时间段”模拟倍率（示意：不同区间放大/缩小数据） =====
const periodScale: Record<string, number> = {
  今日: 0.05,
  本周: 0.25,
  本月: 1,
  本季度: 3,
  本年: 12,
};

// ===== 受控计算数据 =====
const kpis = computed(() => {
  const s = periodScale[selectedPeriod.value] ?? 1;
  return {
    sales: Math.round(baseKpis.sales * s),
    orders: Math.round(baseKpis.orders * s),
    customers: Math.round(baseKpis.customers * s),
    cv: baseKpis.cv, // 比率型可不缩放
    salesWow: baseKpis.salesWow,
    ordersWow: baseKpis.ordersWow,
    customersWow: baseKpis.customersWow,
    cvWow: baseKpis.cvWow,
  };
});

const topProductsComputed = computed(() => {
  const s = periodScale[selectedPeriod.value] ?? 1;
  return topProducts.value.map((p) => ({
    ...p,
    sales: Math.max(0, Math.round(p.sales * s)),
  }));
});

const userSourcesComputed = computed(() => {
  const s = periodScale[selectedPeriod.value] ?? 1;
  // users 按区间缩放，percentage 保持分布
  return userSources.value.map((u) => ({
    ...u,
    users: Math.max(0, Math.round(u.users * s)),
  }));
});

const regionSalesComputed = computed(() => {
  const s = periodScale[selectedPeriod.value] ?? 1;
  return regionSales.value.map((r) => ({
    ...r,
    sales: Math.max(0, Math.round(r.sales * s)),
  }));
});

// 导出（最简单版：导出 TOP10 CSV 示例）
function exportReport() {
  const rows = [
    ["Rank", "Product", "Sales"],
    ...topProductsComputed.value.map((p, i) => [
      String(i + 1),
      p.name,
      String(p.sales),
    ]),
  ];
  const csv = rows
    .map((r) => r.map((x) => `"${String(x).replace(/"/g, '""')}"`).join(","))
    .join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `top-products-${selectedPeriod.value}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

// 排名颜色
const getRankColor = (rank: number) => {
  if (rank === 1) return "bg-yellow-500 text-white";
  if (rank === 2) return "bg-gray-400 text-white";
  if (rank === 3) return "bg-orange-600 text-white";
  return "bg-blue-500 text-white";
};
</script>
