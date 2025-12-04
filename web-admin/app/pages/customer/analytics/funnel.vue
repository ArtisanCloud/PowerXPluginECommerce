<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.funnelAnalysis") }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          分析用户在关键转化路径上的流失情况
        </p>
      </div>
      <div class="flex space-x-2">
        <USelect
          v-model="selectedPeriod"
          :options="timePeriods"
          placeholder="选择时间段"
        />
        <UButton color="blue" variant="soft" icon="i-heroicons-arrow-down-tray">
          导出数据
        </UButton>
      </div>
    </div>

    <!-- 漏斗概览 -->
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">访问首页</p>
          <p class="text-2xl font-bold text-blue-600">10,000</p>
          <p class="text-xs text-gray-500">100%</p>
        </div>
      </UCard>

      <UCard>
        <div class="text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">浏览商品</p>
          <p class="text-2xl font-bold text-green-600">6,500</p>
          <p class="text-xs text-red-500">-35% 流失</p>
        </div>
      </UCard>

      <UCard>
        <div class="text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">加入购物车</p>
          <p class="text-2xl font-bold text-yellow-600">2,800</p>
          <p class="text-xs text-red-500">-57% 流失</p>
        </div>
      </UCard>

      <UCard>
        <div class="text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">完成支付</p>
          <p class="text-2xl font-bold text-purple-600">1,200</p>
          <p class="text-xs text-red-500">-57% 流失</p>
        </div>
      </UCard>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 漏斗可视化 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">转化漏斗</h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="(step, index) in funnelSteps"
            :key="step.name"
            class="relative"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium">{{ step.name }}</span>
              <div class="text-right">
                <span class="text-lg font-bold">{{
                  step.users?.toLocaleString() || '0'
                }}</span>
                <span class="text-xs text-gray-500 ml-2"
                  >{{ step.percentage }}%</span
                >
              </div>
            </div>
            <div
              class="w-full bg-gray-200 rounded-full h-8 relative overflow-hidden"
            >
              <div
                class="h-8 rounded-full transition-all duration-500"
                :class="step.color"
                :style="{ width: step.percentage + '%' }"
              ></div>
              <div
                class="absolute inset-0 flex items-center justify-center text-white text-sm font-medium"
              >
                {{ step.users?.toLocaleString() || '0' }} 用户
              </div>
            </div>
            <!-- 流失箭头 -->
            <div
              v-if="index < funnelSteps.length - 1"
              class="flex justify-center my-2"
            >
              <UIcon
                name="i-heroicons-arrow-down"
                class="w-5 h-5 text-red-500"
              />
              <span class="text-xs text-red-500 ml-1">
                流失
                {{
                  (funnelSteps[index].users - funnelSteps[index + 1].users)?.toLocaleString() || '0'
                }}
                用户
              </span>
            </div>
          </div>
        </div>
      </UCard>

      <!-- 转化率趋势 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">转化率趋势</h3>
        </template>
        <div
          class="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded"
        >
          <p class="text-gray-500">转化率趋势图表区域</p>
        </div>
      </UCard>
    </div>

    <!-- 详细数据表格 -->
    <UCard class="mt-6">
      <template #header>
        <h3 class="text-lg font-semibold">分步骤详细数据</h3>
      </template>
      <UTable :data="detailData" :columns="columns">
        <template #conversionRate-data="{ row }">
          <div class="flex items-center space-x-2">
            <span>{{ row.conversionRate }}%</span>
            <UBadge
              :color="
                row.conversionRate >= 50
                  ? 'green'
                  : row.conversionRate >= 30
                    ? 'yellow'
                    : 'red'
              "
              variant="soft"
              size="xs"
            >
              {{
                row.conversionRate >= 50
                  ? "优秀"
                  : row.conversionRate >= 30
                    ? "良好"
                    : "需优化"
              }}
            </UBadge>
          </div>
        </template>

        <template #dropRate-data="{ row }">
          <span class="text-red-600">{{ row.dropRate }}%</span>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n();

// 响应式数据
const selectedPeriod = ref("最近7天");

// 时间段选项
const timePeriods = [
  { label: "今日", value: "今日" },
  { label: "最近7天", value: "最近7天" },
  { label: "最近30天", value: "最近30天" },
  { label: "最近90天", value: "最近90天" },
];

// 漏斗步骤数据
const funnelSteps = ref([
  {
    name: "访问首页",
    users: 10000,
    percentage: 100,
    color: "bg-blue-500",
  },
  {
    name: "浏览商品",
    users: 6500,
    percentage: 65,
    color: "bg-green-500",
  },
  {
    name: "加入购物车",
    users: 2800,
    percentage: 28,
    color: "bg-yellow-500",
  },
  {
    name: "完成支付",
    users: 1200,
    percentage: 12,
    color: "bg-purple-500",
  },
]);

// 表格列定义
const columns = [
  { id: "step", key: "step", label: "转化步骤" },
  { id: "users", key: "users", label: "用户数" },
  { id: "conversionRate", key: "conversionRate", label: "转化率" },
  { id: "dropRate", key: "dropRate", label: "流失率" },
  { id: "avgTime", key: "avgTime", label: "平均停留时间" },
];

// 详细数据
const detailData = ref([
  {
    step: "访问首页",
    users: "10,000",
    conversionRate: 65.0,
    dropRate: 35.0,
    avgTime: "2分30秒",
  },
  {
    step: "浏览商品",
    users: "6,500",
    conversionRate: 43.1,
    dropRate: 56.9,
    avgTime: "4分15秒",
  },
  {
    step: "加入购物车",
    users: "2,800",
    conversionRate: 42.9,
    dropRate: 57.1,
    avgTime: "1分45秒",
  },
  {
    step: "完成支付",
    users: "1,200",
    conversionRate: 100,
    dropRate: 0,
    avgTime: "3分20秒",
  },
]);
</script>
