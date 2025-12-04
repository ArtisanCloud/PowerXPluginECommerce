<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.retentionAnalysis") }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          分析用户在不同时间段的留存情况
        </p>
      </div>
      <div class="flex space-x-2">
        <USelect
          v-model="selectedCohort"
          :options="cohortTypes"
          placeholder="选择队列类型"
        />
        <UButton color="blue" variant="soft" icon="i-heroicons-arrow-down-tray">
          导出数据
        </UButton>
      </div>
    </div>

    <!-- 留存概览 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">次日留存</p>
            <p class="text-2xl font-bold text-green-600">68.5%</p>
            <p class="text-xs text-green-500">+2.3% 较上周</p>
          </div>
          <UIcon
            name="i-heroicons-calendar-days"
            class="w-8 h-8 text-green-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">7日留存</p>
            <p class="text-2xl font-bold text-blue-600">42.3%</p>
            <p class="text-xs text-blue-500">+1.8% 较上周</p>
          </div>
          <UIcon name="i-heroicons-calendar" class="w-8 h-8 text-blue-500" />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">30日留存</p>
            <p class="text-2xl font-bold text-purple-600">28.7%</p>
            <p class="text-xs text-red-500">-0.5% 较上周</p>
          </div>
          <UIcon
            name="i-heroicons-calendar-days"
            class="w-8 h-8 text-purple-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">平均留存</p>
            <p class="text-2xl font-bold text-orange-600">46.5%</p>
            <p class="text-xs text-green-500">+1.2% 较上周</p>
          </div>
          <UIcon name="i-heroicons-chart-bar" class="w-8 h-8 text-orange-500" />
        </div>
      </UCard>
    </div>

    <!-- 留存热力图 -->
    <UCard class="mb-6">
      <template #header>
        <h3 class="text-lg font-semibold">留存热力图</h3>
      </template>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b">
              <th class="text-left p-2">注册日期</th>
              <th class="text-center p-2">用户数</th>
              <th class="text-center p-2">Day 1</th>
              <th class="text-center p-2">Day 7</th>
              <th class="text-center p-2">Day 14</th>
              <th class="text-center p-2">Day 30</th>
              <th class="text-center p-2">Day 60</th>
              <th class="text-center p-2">Day 90</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in retentionData" :key="row.date" class="border-b">
              <td class="p-2 font-medium">{{ row.date }}</td>
              <td class="p-2 text-center">{{ row.users }}</td>
              <td
                v-for="(rate, index) in row.retention"
                :key="index"
                class="p-2 text-center"
                :class="getRetentionColor(rate)"
              >
                {{ rate }}%
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </UCard>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 留存趋势图 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">留存趋势</h3>
        </template>
        <div
          class="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded"
        >
          <p class="text-gray-500">留存趋势图表区域</p>
        </div>
      </UCard>

      <!-- 渠道留存对比 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">渠道留存对比</h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="channel in channelRetention"
            :key="channel.name"
            class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded"
          >
            <div class="flex items-center space-x-3">
              <div
                class="w-4 h-4 rounded"
                :style="{ backgroundColor: channel.color }"
              ></div>
              <span class="font-medium">{{ channel.name }}</span>
            </div>
            <div class="text-right">
              <p class="font-semibold">{{ channel.retention }}%</p>
              <p class="text-xs text-gray-500">7日留存</p>
            </div>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n();

// 响应式数据
const selectedCohort = ref("按注册日期");

// 队列类型选项
const cohortTypes = [
  { label: "按注册日期", value: "按注册日期" },
  { label: "按首次购买", value: "按首次购买" },
  { label: "按渠道来源", value: "按渠道来源" },
];

// 留存数据
const retentionData = ref([
  {
    date: "2024-01-15",
    users: 1250,
    retention: [68.5, 42.3, 35.2, 28.7, 22.1, 18.9],
  },
  {
    date: "2024-01-14",
    users: 1180,
    retention: [71.2, 45.8, 38.1, 31.2, 24.5, 20.3],
  },
  {
    date: "2024-01-13",
    users: 1320,
    retention: [66.8, 40.1, 33.7, 26.9, 21.2, 17.8],
  },
  {
    date: "2024-01-12",
    users: 1090,
    retention: [69.7, 43.5, 36.8, 29.4, 23.1, 19.6],
  },
  {
    date: "2024-01-11",
    users: 1410,
    retention: [72.1, 46.2, 39.5, 32.8, 25.7, 21.4],
  },
]);

// 渠道留存数据
const channelRetention = ref([
  { name: "自然搜索", retention: 45.2, color: "#3B82F6" },
  { name: "社交媒体", retention: 38.7, color: "#10B981" },
  { name: "付费广告", retention: 35.1, color: "#F59E0B" },
  { name: "邮件营销", retention: 52.3, color: "#EF4444" },
  { name: "直接访问", retention: 48.9, color: "#8B5CF6" },
]);

// 获取留存率颜色
const getRetentionColor = (rate: number) => {
  if (rate >= 50) return "bg-green-100 text-green-800";
  if (rate >= 30) return "bg-yellow-100 text-yellow-800";
  if (rate >= 15) return "bg-orange-100 text-orange-800";
  return "bg-red-100 text-red-800";
};
</script>
