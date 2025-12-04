<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.userSegmentation") }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          基于用户行为和属性进行智能分群，实现精准营销
        </p>
      </div>
      <div class="flex space-x-2">
        <UButton color="primary" variant="outline" icon="i-heroicons-plus">
          创建分群
        </UButton>
        <UButton
          color="primary"
          variant="soft"
          icon="i-heroicons-arrow-down-tray"
        >
          导出分群
        </UButton>
      </div>
    </div>

    <!-- 分群概览 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">活跃分群</p>
            <p class="text-2xl font-bold text-primary-600">8</p>
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
            <p class="text-sm text-gray-500 dark:text-gray-400">覆盖用户</p>
            <p class="text-2xl font-bold text-success-600">24,580</p>
            <p class="text-xs text-success-500">+12.3% 较上月</p>
          </div>
          <UIcon name="i-heroicons-users" class="w-8 h-8 text-success-500" />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">平均转化率</p>
            <p class="text-2xl font-bold text-warning-600">18.7%</p>
            <p class="text-xs text-success-500">+3.2% 较上月</p>
          </div>
          <UIcon
            name="i-heroicons-chart-bar"
            class="w-8 h-8 text-warning-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">营销活动</p>
            <p class="text-2xl font-bold text-error-600">15</p>
            <p class="text-xs text-success-500">进行中</p>
          </div>
          <UIcon name="i-heroicons-megaphone" class="w-8 h-8 text-error-500" />
        </div>
      </UCard>
    </div>

    <!-- 用户分群列表 -->
    <UCard class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold">用户分群列表</h3>
          <div class="flex space-x-2">
            <UInput
              v-model="searchQuery"
              placeholder="搜索分群..."
              icon="i-heroicons-magnifying-glass"
              size="sm"
            />
            <USelect
              v-model="selectedStatus"
              :options="statusOptions"
              placeholder="状态筛选"
              size="sm"
            />
          </div>
        </div>
      </template>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="segment in filteredSegments"
          :key="segment.id"
          class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg hover:shadow-md transition-shadow"
        >
          <div class="flex justify-between items-start mb-3">
            <div>
              <h4 class="font-semibold text-gray-900 dark:text-white">
                {{ segment.name }}
              </h4>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ segment.description }}
              </p>
            </div>
            <UBadge
              :color="
                segment.status === '活跃'
                  ? 'success'
                  : segment.status === '暂停'
                    ? 'warning'
                    : 'neutral'
              "
              variant="soft"
              size="xs"
            >
              {{ segment.status }}
            </UBadge>
          </div>

          <div class="space-y-2 mb-4">
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">用户数量</span>
              <span class="font-medium">{{
                segment.userCount?.toLocaleString() || '0'
              }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">转化率</span>
              <span class="font-medium text-success-600"
                >{{ segment.conversionRate }}%</span
              >
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">平均客单价</span>
              <span class="font-medium">¥{{ segment.avgOrderValue }}</span>
            </div>
          </div>

          <div class="flex space-x-2">
            <UButton size="xs" variant="outline" class="flex-1">
              查看详情
            </UButton>
            <UButton size="xs" color="primary" class="flex-1">
              创建活动
            </UButton>
          </div>
        </div>
      </div>
    </UCard>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 分群表现对比 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">分群表现对比</h3>
        </template>
        <div
          class="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded"
        >
          <p class="text-gray-500">分群表现对比图表区域</p>
        </div>
      </UCard>

      <!-- 分群规则分布 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">分群规则分布</h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="rule in segmentRules"
            :key="rule.type"
            class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded"
          >
            <div class="flex items-center space-x-3">
              <div
                class="w-4 h-4 rounded"
                :style="{ backgroundColor: rule.color }"
              ></div>
              <span class="font-medium">{{ rule.type }}</span>
            </div>
            <div class="text-right">
              <p class="font-semibold">{{ rule.count }}</p>
              <p class="text-xs text-gray-500">个分群</p>
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
const searchQuery = ref("");
const selectedStatus = ref("");

// 状态选项
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "活跃", value: "活跃" },
  { label: "暂停", value: "暂停" },
  { label: "已完成", value: "已完成" },
];

// 用户分群数据
const segments = ref([
  {
    id: 1,
    name: "高价值客户",
    description: "近90天消费金额>5000元的用户",
    userCount: 2580,
    conversionRate: 45.2,
    avgOrderValue: 1280,
    status: "活跃",
  },
  {
    id: 2,
    name: "新用户",
    description: "注册时间<30天的用户",
    userCount: 4200,
    conversionRate: 12.8,
    avgOrderValue: 320,
    status: "活跃",
  },
  {
    id: 3,
    name: "流失预警",
    description: "60天未活跃但历史有购买的用户",
    userCount: 1850,
    conversionRate: 8.5,
    avgOrderValue: 450,
    status: "活跃",
  },
  {
    id: 4,
    name: "移动端用户",
    description: "主要通过移动设备访问的用户",
    userCount: 8900,
    conversionRate: 18.7,
    avgOrderValue: 580,
    status: "活跃",
  },
  {
    id: 5,
    name: "复购客户",
    description: "购买次数>=3次的用户",
    userCount: 3200,
    conversionRate: 38.9,
    avgOrderValue: 950,
    status: "活跃",
  },
  {
    id: 6,
    name: "价格敏感型",
    description: "主要购买促销商品的用户",
    userCount: 5600,
    conversionRate: 22.3,
    avgOrderValue: 280,
    status: "暂停",
  },
]);

// 分群规则分布
const segmentRules = ref([
  { type: "行为规则", count: 12, color: "#3B82F6" },
  { type: "属性规则", count: 8, color: "#10B981" },
  { type: "交易规则", count: 15, color: "#F59E0B" },
  { type: "时间规则", count: 6, color: "#EF4444" },
  { type: "地理规则", count: 4, color: "#8B5CF6" },
]);

// 计算属性：过滤后的分群
const filteredSegments = computed(() => {
  return segments.value.filter((segment) => {
    const matchesSearch =
      segment.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      segment.description
        .toLowerCase()
        .includes(searchQuery.value.toLowerCase());
    const matchesStatus =
      !selectedStatus.value || segment.status === selectedStatus.value;
    return matchesSearch && matchesStatus;
  });
});
</script>
