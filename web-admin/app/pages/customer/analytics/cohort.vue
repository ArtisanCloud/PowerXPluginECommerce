<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.cohortAnalysis") }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          按用户注册时间分组，分析不同队列的行为表现
        </p>
      </div>
      <div class="flex space-x-2">
        <USelect
          v-model="selectedMetric"
          :options="metricOptions"
          placeholder="选择指标"
        />
        <UButton
          color="primary"
          variant="soft"
          icon="i-heroicons-arrow-down-tray"
        >
          导出数据
        </UButton>
      </div>
    </div>

    <!-- 队列概览 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">活跃队列</p>
            <p class="text-2xl font-bold text-primary-600">12</p>
            <p class="text-xs text-success-500">+2 本月新增</p>
          </div>
          <UIcon
            name="i-heroicons-user-group"
            class="w-8 h-8 text-primary-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">平均队列规模</p>
            <p class="text-2xl font-bold text-success-600">1,250</p>
            <p class="text-xs text-success-500">+8.5% 较上月</p>
          </div>
          <UIcon name="i-heroicons-users" class="w-8 h-8 text-success-500" />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">最佳表现队列</p>
            <p class="text-2xl font-bold text-warning-600">2024-01</p>
            <p class="text-xs text-success-500">65.2% 留存率</p>
          </div>
          <UIcon name="i-heroicons-trophy" class="w-8 h-8 text-warning-500" />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">队列生命周期</p>
            <p class="text-2xl font-bold text-error-600">180</p>
            <p class="text-xs text-gray-500">天（平均）</p>
          </div>
          <UIcon name="i-heroicons-clock" class="w-8 h-8 text-error-500" />
        </div>
      </UCard>
    </div>

    <!-- 队列表现对比 -->
    <UCard class="mb-6">
      <template #header>
        <h3 class="text-lg font-semibold">队列表现对比</h3>
      </template>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b">
              <th class="text-left p-3">队列</th>
              <th class="text-center p-3">用户数</th>
              <th class="text-center p-3">7日留存</th>
              <th class="text-center p-3">30日留存</th>
              <th class="text-center p-3">平均LTV</th>
              <th class="text-center p-3">转化率</th>
              <th class="text-center p-3">表现评级</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="cohort in cohortData"
              :key="cohort.period"
              class="border-b hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              <td class="p-3 font-medium">{{ cohort.period }}</td>
              <td class="p-3 text-center">
                {{ cohort.users?.toLocaleString() || '0' }}
              </td>
              <td class="p-3 text-center">
                <span :class="getRetentionColor(cohort.retention7d)">
                  {{ cohort.retention7d }}%
                </span>
              </td>
              <td class="p-3 text-center">
                <span :class="getRetentionColor(cohort.retention30d)">
                  {{ cohort.retention30d }}%
                </span>
              </td>
              <td class="p-3 text-center font-medium">¥{{ cohort.ltv }}</td>
              <td class="p-3 text-center">{{ cohort.conversionRate }}%</td>
              <td class="p-3 text-center">
                <UBadge
                  :color="
                    cohort.grade === 'A'
                      ? 'success'
                      : cohort.grade === 'B'
                        ? 'warning'
                        : 'error'
                  "
                  variant="soft"
                >
                  {{ cohort.grade }}
                </UBadge>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </UCard>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 队列生命周期 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">队列生命周期分析</h3>
        </template>
        <div
          class="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded"
        >
          <p class="text-gray-500">队列生命周期图表区域</p>
        </div>
      </UCard>

      <!-- 队列价值分布 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">队列价值分布</h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="segment in valueSegments"
            :key="segment.name"
            class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded"
          >
            <div class="flex items-center space-x-3">
              <div
                class="w-4 h-4 rounded"
                :style="{ backgroundColor: segment.color }"
              ></div>
              <span class="font-medium">{{ segment.name }}</span>
            </div>
            <div class="text-right">
              <p class="font-semibold">{{ segment.percentage }}%</p>
              <p class="text-xs text-gray-500">{{ segment.count }} 个队列</p>
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
const selectedMetric = ref("留存率");

// 指标选项
const metricOptions = [
  { label: "留存率", value: "留存率" },
  { label: "转化率", value: "转化率" },
  { label: "生命周期价值", value: "生命周期价值" },
  { label: "活跃度", value: "活跃度" },
];

// 队列数据
const cohortData = ref([
  {
    period: "2024-01",
    users: 1520,
    retention7d: 65.2,
    retention30d: 42.8,
    ltv: 1280,
    conversionRate: 18.5,
    grade: "A",
  },
  {
    period: "2024-02",
    users: 1380,
    retention7d: 58.7,
    retention30d: 38.2,
    ltv: 1150,
    conversionRate: 16.2,
    grade: "B",
  },
  {
    period: "2024-03",
    users: 1650,
    retention7d: 62.1,
    retention30d: 40.5,
    ltv: 1320,
    conversionRate: 17.8,
    grade: "A",
  },
  {
    period: "2024-04",
    users: 1420,
    retention7d: 55.3,
    retention30d: 35.7,
    ltv: 980,
    conversionRate: 14.9,
    grade: "B",
  },
  {
    period: "2024-05",
    users: 1180,
    retention7d: 48.9,
    retention30d: 28.4,
    ltv: 850,
    conversionRate: 12.3,
    grade: "C",
  },
]);

// 价值分段数据
const valueSegments = ref([
  { name: "高价值队列", percentage: 25, count: 3, color: "#10B981" },
  { name: "中等价值队列", percentage: 50, count: 6, color: "#F59E0B" },
  { name: "低价值队列", percentage: 25, count: 3, color: "#EF4444" },
]);

// 获取留存率颜色
const getRetentionColor = (rate: number) => {
  if (rate >= 50) return "text-success-600 font-semibold";
  if (rate >= 30) return "text-warning-600 font-semibold";
  return "text-error-600 font-semibold";
};
</script>
