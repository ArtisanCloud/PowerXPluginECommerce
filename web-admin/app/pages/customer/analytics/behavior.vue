<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.behaviorAnalysis") }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          深入分析用户在平台上的行为模式和偏好
        </p>
      </div>
      <div class="flex space-x-2">
        <USelect
          v-model="selectedTimeRange"
          :options="timeRangeOptions"
          placeholder="选择时间范围"
        />
        <UButton
          color="primary"
          variant="soft"
          icon="i-heroicons-arrow-down-tray"
        >
          导出报告
        </UButton>
      </div>
    </div>

    <!-- 行为概览 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">平均会话时长</p>
            <p class="text-2xl font-bold text-primary-600">8分32秒</p>
            <p class="text-xs text-success-500">+12% 较上周</p>
          </div>
          <UIcon name="i-heroicons-clock" class="w-8 h-8 text-primary-500" />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">页面浏览深度</p>
            <p class="text-2xl font-bold text-success-600">4.2</p>
            <p class="text-xs text-success-500">+0.3 较上周</p>
          </div>
          <UIcon name="i-heroicons-eye" class="w-8 h-8 text-success-500" />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">跳出率</p>
            <p class="text-2xl font-bold text-warning-600">32.5%</p>
            <p class="text-xs text-success-500">-2.1% 较上周</p>
          </div>
          <UIcon
            name="i-heroicons-arrow-right-on-rectangle"
            class="w-8 h-8 text-warning-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">互动率</p>
            <p class="text-2xl font-bold text-error-600">68.7%</p>
            <p class="text-xs text-success-500">+5.2% 较上周</p>
          </div>
          <UIcon
            name="i-heroicons-hand-raised"
            class="w-8 h-8 text-error-500"
          />
        </div>
      </UCard>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
      <!-- 用户行为路径 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">热门行为路径</h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="(path, index) in behaviorPaths"
            :key="index"
            class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="font-medium">路径 {{ index + 1 }}</span>
              <UBadge color="primary" variant="soft">
                {{ path.percentage }}% 用户
              </UBadge>
            </div>
            <div
              class="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400"
            >
              <span
                v-for="(step, stepIndex) in path.steps"
                :key="stepIndex"
                class="flex items-center"
              >
                {{ step }}
                <UIcon
                  v-if="stepIndex < path.steps.length - 1"
                  name="i-heroicons-arrow-right"
                  class="w-4 h-4 mx-2"
                />
              </span>
            </div>
          </div>
        </div>
      </UCard>

      <!-- 设备使用分布 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">设备使用分布</h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="device in deviceData"
            :key="device.name"
            class="flex items-center justify-between"
          >
            <div class="flex items-center space-x-3">
              <UIcon :name="device.icon" class="w-5 h-5 text-gray-500" />
              <span class="font-medium">{{ device.name }}</span>
            </div>
            <div class="flex items-center space-x-3">
              <div class="w-32 bg-gray-200 rounded-full h-2">
                <div
                  class="h-2 rounded-full"
                  :class="device.color"
                  :style="{ width: device.percentage + '%' }"
                ></div>
              </div>
              <span class="text-sm font-semibold w-12 text-right"
                >{{ device.percentage }}%</span
              >
            </div>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 页面热力图 -->
    <UCard class="mb-6">
      <template #header>
        <h3 class="text-lg font-semibold">页面访问热力图</h3>
      </template>
      <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        <div
          v-for="page in pageHeatmap"
          :key="page.name"
          class="p-4 rounded-lg text-center"
          :class="getHeatmapColor(page.visits)"
        >
          <p class="text-sm font-medium text-white">{{ page.name }}</p>
          <p class="text-lg font-bold text-white">
            {{ page.visits?.toLocaleString() || '0' }}
          </p>
          <p class="text-xs text-white opacity-80">访问次数</p>
        </div>
      </div>
    </UCard>

    <!-- 用户行为事件 -->
    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold">用户行为事件统计</h3>
      </template>
      <UTable :data="eventData" :columns="eventColumns">
        <template #trend-data="{ row }">
          <div class="flex items-center space-x-2">
            <UIcon
              :name="
                row.trend > 0
                  ? 'i-heroicons-arrow-trending-up'
                  : 'i-heroicons-arrow-trending-down'
              "
              :class="row.trend > 0 ? 'text-success-500' : 'text-error-500'"
              class="w-4 h-4"
            />
            <span
              :class="row.trend > 0 ? 'text-success-600' : 'text-error-600'"
            >
              {{ Math.abs(row.trend) }}%
            </span>
          </div>
        </template>

        <template #category-data="{ row }">
          <UBadge
            :color="
              row.category === '购买'
                ? 'success'
                : row.category === '浏览'
                  ? 'primary'
                  : 'warning'
            "
            variant="soft"
          >
            {{ row.category }}
          </UBadge>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n();

// 响应式数据
const selectedTimeRange = ref("最近7天");

// 时间范围选项
const timeRangeOptions = [
  { label: "最近24小时", value: "最近24小时" },
  { label: "最近7天", value: "最近7天" },
  { label: "最近30天", value: "最近30天" },
  { label: "最近90天", value: "最近90天" },
];

// 行为路径数据
const behaviorPaths = ref([
  {
    steps: ["首页", "商品列表", "商品详情", "加入购物车", "结算"],
    percentage: 28.5,
  },
  {
    steps: ["首页", "搜索", "商品详情", "收藏"],
    percentage: 22.3,
  },
  {
    steps: ["商品列表", "筛选", "商品详情", "对比", "购买"],
    percentage: 18.7,
  },
  {
    steps: ["首页", "分类", "商品列表", "离开"],
    percentage: 15.2,
  },
]);

// 设备数据
const deviceData = ref([
  {
    name: "移动设备",
    percentage: 68,
    icon: "i-heroicons-device-phone-mobile",
    color: "bg-primary-500",
  },
  {
    name: "桌面设备",
    percentage: 28,
    icon: "i-heroicons-computer-desktop",
    color: "bg-success-500",
  },
  {
    name: "平板设备",
    percentage: 4,
    icon: "i-heroicons-device-tablet",
    color: "bg-warning-500",
  },
]);

// 页面热力图数据
const pageHeatmap = ref([
  { name: "首页", visits: 15420 },
  { name: "商品列表", visits: 12350 },
  { name: "商品详情", visits: 8960 },
  { name: "购物车", visits: 5240 },
  { name: "用户中心", visits: 3180 },
  { name: "订单页", visits: 2890 },
  { name: "搜索页", visits: 4560 },
  { name: "分类页", visits: 6780 },
  { name: "帮助中心", visits: 1240 },
  { name: "关于我们", visits: 890 },
  { name: "联系我们", visits: 650 },
  { name: "优惠活动", visits: 7320 },
]);

// 事件数据
const eventData = ref([
  {
    event: "商品点击",
    count: "45,230",
    category: "浏览",
    avgDuration: "2分15秒",
    trend: 8.5,
  },
  {
    event: "加入购物车",
    count: "12,450",
    category: "购买",
    avgDuration: "30秒",
    trend: 12.3,
  },
  {
    event: "搜索查询",
    count: "28,670",
    category: "浏览",
    avgDuration: "1分45秒",
    trend: -2.1,
  },
  {
    event: "收藏商品",
    count: "8,920",
    category: "互动",
    avgDuration: "15秒",
    trend: 15.7,
  },
  {
    event: "分享商品",
    count: "3,240",
    category: "互动",
    avgDuration: "25秒",
    trend: 22.4,
  },
]);

// 事件表格列
const eventColumns = [
  { id: "event", key: "event", label: "事件名称" },
  { id: "count", key: "count", label: "触发次数" },
  { id: "category", key: "category", label: "事件类别" },
  { id: "avgDuration", key: "avgDuration", label: "平均时长" },
  { id: "trend", key: "trend", label: "趋势" },
];

// 获取热力图颜色
const getHeatmapColor = (visits: number) => {
  if (visits >= 10000) return "bg-error-500";
  if (visits >= 5000) return "bg-warning-500";
  if (visits >= 2000) return "bg-primary-500";
  return "bg-gray-500";
};
</script>
