<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          客户分布分析
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          按会员等级统计客户分布情况
        </p>
      </div>
      <div class="flex gap-3">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-heroicons-arrow-down-tray"
          @click="exportData"
        >
          导出数据
        </UButton>
        <UButton
          color="primary"
          icon="i-heroicons-arrow-path"
          @click="refreshData"
        >
          刷新
        </UButton>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon
              name="i-heroicons-users"
              class="w-6 h-6 text-blue-600 dark:text-blue-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              总客户数
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ totalCustomers?.toLocaleString() || '0' }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon
              name="i-heroicons-star"
              class="w-6 h-6 text-green-600 dark:text-green-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              平均等级
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ averageLevel.toFixed(1) }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon
              name="i-heroicons-arrow-trending-up"
              class="w-6 h-6 text-yellow-600 dark:text-yellow-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              最高等级客户
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ highestTierCustomers?.toLocaleString() || '0' }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon
              name="i-heroicons-arrow-trending-down"
              class="w-6 h-6 text-purple-600 dark:text-purple-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              最低等级客户
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ lowestTierCustomers?.toLocaleString() || '0' }}
            </p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 分布图表 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
      <UCard>
        <template #header>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            客户等级分布（柱状图）
          </h2>
        </template>
        <ClientOnly>
          <VChart
            v-if="barChartData"
            :option="barChartData"
            :style="{ height: '320px' }"
            autoresize
          />
        </ClientOnly>
      </UCard>

      <UCard>
        <template #header>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            客户等级分布（饼图）
          </h2>
        </template>
        <ClientOnly>
          <VChart
            v-if="pieChartData"
            :option="pieChartData"
            :style="{ height: '320px' }"
            autoresize
          />
        </ClientOnly>
      </UCard>
    </div>

    <!-- 等级分布表格 -->
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            等级分布详情
          </h2>
          <div class="flex gap-2">
            <UInput
              v-model="searchQuery"
              placeholder="搜索等级名称..."
              icon="i-heroicons-magnifying-glass"
              size="sm"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredTiers">
        <template #name-cell="{ row }">
          <div class="flex items-center gap-3">
            <div
              class="w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-medium"
              :style="{ backgroundColor: row.color }"
            >
              {{ (row.name || '').charAt(0) }}
            </div>
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ row.name }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ row.description }}
              </div>
            </div>
          </div>
        </template>

        <template #level-cell="{ row }">
          <UBadge variant="soft" size="sm">
            {{ row.level }}
          </UBadge>
        </template>

        <template #customerCount-cell="{ row }">
          <div class="text-center">
            <div class="font-medium text-gray-900 dark:text-white">
              {{ row.customerCount?.toLocaleString() || 0 }}
            </div>
            <div class="text-xs text-gray-500 dark:text-gray-400">
              占比 {{ row.percentage }}%
            </div>
          </div>
        </template>

        <template #growth-cell="{ row }">
          <div class="flex items-center gap-1">
            <UIcon
              :name="
                row.growth >= 0
                  ? 'i-heroicons-arrow-trending-up'
                  : 'i-heroicons-arrow-trending-down'
              "
              :class="[
                'w-4 h-4',
                row.growth >= 0
                  ? 'text-green-600 dark:text-green-400'
                  : 'text-red-600 dark:text-red-400',
              ]"
            />
            <span
              :class="[
                'text-sm font-medium',
                row.growth >= 0
                  ? 'text-green-600 dark:text-green-400'
                  : 'text-red-600 dark:text-red-400',
              ]"
            >
              {{ Math.abs(row.growth) }}%
            </span>
          </div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
// 导入 ECharts 组件
import VChart from "vue-echarts";
import * as echarts from "echarts";

// 设置 ECharts 主题
echarts.registerTheme("custom", {
  color: [
    "#CD7F32",
    "#C0C0C0",
    "#FFD700",
    "#E5E4E2",
    "#B9F2FF"
  ]
});

// 搜索
const searchQuery = ref("");

// 统计数据
const totalCustomers = ref(2745);
const averageLevel = ref(2.8);
const highestTierCustomers = ref(45);
const lowestTierCustomers = ref(1250);

// 等级分布数据
const tiers = ref([
  {
    id: "tier_1",
    name: "青铜会员",
    description: "新用户默认等级",
    color: "#CD7F32",
    level: 1,
    customerCount: 1250,
    percentage: 45.5,
    growth: -2.3,
  },
  {
    id: "tier_2",
    name: "白银会员",
    description: "消费达标的忠实用户",
    color: "#C0C0C0",
    level: 2,
    customerCount: 850,
    percentage: 31.0,
    growth: 5.7,
  },
  {
    id: "tier_3",
    name: "黄金会员",
    description: "高价值客户",
    color: "#FFD700",
    level: 3,
    customerCount: 420,
    percentage: 15.3,
    growth: 12.4,
  },
  {
    id: "tier_4",
    name: "铂金会员",
    description: "VIP客户",
    color: "#E5E4E2",
    level: 4,
    customerCount: 180,
    percentage: 6.6,
    growth: 8.9,
  },
  {
    id: "tier_5",
    name: "钻石会员",
    description: "顶级VIP客户",
    color: "#B9F2FF",
    level: 5,
    customerCount: 45,
    percentage: 1.6,
    growth: 15.2,
  },
]);

// 过滤后的等级列表
const filteredTiers = computed(() => {
  if (!searchQuery.value) return tiers.value;
  return tiers.value.filter((tier) =>
    tier.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});

// 表格列定义
const columns = [
  { accessorKey: "name", header: "等级名称" },
  { accessorKey: "level", header: "等级" },
  { accessorKey: "customerCount", header: "客户数" },
  { accessorKey: "growth", header: "增长率" },
];

// 柱状图数据
const barChartData = computed(() => {
  return {
    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "shadow"
      }
    },
    grid: {
      left: "3%",
      right: "4%",
      bottom: "3%",
      containLabel: true
    },
    xAxis: [
      {
        type: "category",
        data: tiers.value.map((tier) => tier.name),
        axisTick: {
          alignWithLabel: true
        }
      }
    ],
    yAxis: [
      {
        type: "value",
        axisLabel: {
          formatter: function (value: number) {
            return value.toLocaleString();
          }
        }
      }
    ],
    series: [
      {
        name: "客户数",
        type: "bar",
        barWidth: "60%",
        data: tiers.value.map((tier) => tier.customerCount),
        itemStyle: {
          color: function (params: any) {
            return tiers.value[params.dataIndex].color;
          }
        },
        label: {
          show: true,
          position: "top",
          formatter: function (params: any) {
            return params.value.toLocaleString();
          }
        }
      }
    ]
  };
});

// 饼图数据
const pieChartData = computed(() => {
  return {
    tooltip: {
      trigger: "item",
      formatter: function (params: any) {
        const value = params.value || 0;
        const total = tiers.value.reduce((sum, tier) => sum + (tier.customerCount || 0), 0);
        const percentage = ((value / total) * 100).toFixed(1);
        return `${params.name}<br/>${value.toLocaleString()} (${percentage}%)`;
      }
    },
    legend: {
      orient: "vertical",
      left: "left"
    },
    series: [
      {
        name: "客户分布",
        type: "pie",
        radius: ["40%", "70%"],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: "#fff",
          borderWidth: 2
        },
        label: {
          show: false,
          position: "center"
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: "bold",
            formatter: function (params: any) {
              return `${params.name}\n${params.percent.toFixed(1)}%`;
            }
          }
        },
        labelLine: {
          show: false
        },
        data: tiers.value.map((tier) => ({
          value: tier.customerCount,
          name: tier.name,
          itemStyle: {
            color: tier.color
          }
        }))
      }
    ]
  };
});

// 导出数据
const exportData = () => {
  alert("导出数据功能待实现");
};

// 刷新数据
const refreshData = async () => {
  // 模拟API调用
  await new Promise((resolve) => setTimeout(resolve, 1000));
  alert("数据已刷新");
};
</script>
