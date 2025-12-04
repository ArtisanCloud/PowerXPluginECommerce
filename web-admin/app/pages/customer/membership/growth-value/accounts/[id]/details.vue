<template>
  <div>
    <!-- 页面标题和返回按钮 -->
    <div class="flex justify-between items-center mb-6">
      <div class="flex items-center gap-4">
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-heroicons-arrow-left"
          @click="goBack"
        >
          返回
        </UButton>
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
            成长值明细
          </h1>
          <p class="text-gray-600 dark:text-gray-400 mt-1">
            查看成长值获取和消耗记录
          </p>
        </div>
      </div>
      <div class="flex gap-3">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-heroicons-arrow-down-tray"
          @click="exportData"
        >
          导出记录
        </UButton>
      </div>
    </div>

    <!-- 子菜单导航 -->
    <div class="mb-6">
      <div class="border-b border-gray-200 dark:border-gray-800">
        <nav class="-mb-px flex space-x-8">
          <UButton
            :to="`/customer/membership/growth-value/accounts/${accountId}`"
            variant="ghost"
            color="neutral"
            class="whitespace-nowrap py-4 px-1 border-b-2 border-transparent font-medium text-sm"
            :class="{
              'border-primary-500 text-primary-600 dark:text-primary-400': $route.path === `/customer/membership/growth-value/accounts/${accountId}`,
              'text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300 dark:hover:border-gray-700': $route.path !== `/customer/membership/growth-value/accounts/${accountId}`
            }"
          >
            账户信息
          </UButton>
          <UButton
            :to="`/customer/membership/growth-value/accounts/${accountId}/details`"
            variant="ghost"
            color="neutral"
            class="whitespace-nowrap py-4 px-1 border-b-2 border-transparent font-medium text-sm"
            :class="{
              'border-primary-500 text-primary-600 dark:text-primary-400': $route.path === `/customer/membership/growth-value/accounts/${accountId}/details`,
              'text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300 dark:hover:border-gray-700': $route.path !== `/customer/membership/growth-value/accounts/${accountId}/details`
            }"
          >
            成长值明细
          </UButton>
        </nav>
      </div>
    </div>

    <!-- 客户信息卡片 -->
    <UCard class="mb-6">
      <div class="flex items-center gap-4">
        <UAvatar
          :src="account.avatar"
          :alt="account.customerName"
          size="lg"
          :ui="{ rounded: 'rounded-full' }"
        />
        <div>
          <div class="text-xl font-bold text-gray-900 dark:text-white">
            {{ account.customerName }}
          </div>
          <div class="text-gray-600 dark:text-gray-400">
            {{ account.customerPhone }}
          </div>
        </div>
        <div class="ml-auto text-right">
          <div class="text-sm text-gray-500 dark:text-gray-400">当前成长值余额</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ account.balance?.toLocaleString() || '0' }}
          </div>
        </div>
      </div>
    </UCard>

    <!-- 筛选和统计 -->
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-6 mb-6">
      <!-- 筛选条件 -->
      <UCard class="lg:col-span-1">
        <div class="space-y-4">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
            筛选条件
          </h3>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              时间范围
            </label>
            <USelect
              v-model="filter.timeRange"
              :options="timeRangeOptions"
              option-attribute="label"
              value-attribute="value"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              类型
            </label>
            <USelect
              v-model="filter.type"
              :options="typeOptions"
              option-attribute="label"
              value-attribute="value"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              关键词
            </label>
            <UInput
              v-model="filter.keyword"
              placeholder="搜索备注..."
              icon="i-heroicons-magnifying-glass"
            />
          </div>

          <div class="flex gap-2 pt-2">
            <UButton
              color="primary"
              @click="applyFilter"
            >
              应用筛选
            </UButton>
            <UButton
              color="neutral"
              variant="outline"
              @click="resetFilter"
            >
              重置
            </UButton>
          </div>
        </div>
      </UCard>

      <!-- 晋升进度和统计信息 -->
      <div class="lg:col-span-3 space-y-4">
        <!-- 晋升进度 -->
        <UCard>
          <template #header>
            <h3 class="text-md font-semibold text-gray-900 dark:text-white">
              等级晋升进度
            </h3>
          </template>

          <div class="space-y-4">
            <div v-for="rule in nextLevelRules" :key="rule.id" class="space-y-2">
              <div class="flex justify-between text-sm">
                <span class="text-gray-600 dark:text-gray-400">
                  晋升到 {{ rule.levelName }}
                </span>
                <span class="font-medium text-gray-900 dark:text-white">
                  {{ account.balance }} / {{ rule.growthValueThreshold }}
                </span>
              </div>
              <UProgress
                :value="Math.min(100, (account.balance / rule.growthValueThreshold) * 100)"
                :color="rule.levelColor"
                size="sm"
              />
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                还需 {{ Math.max(0, rule.growthValueThreshold - account.balance) }} 成长值
              </div>
            </div>

            <div v-if="nextLevelRules.length === 0" class="text-center py-2 text-gray-500 dark:text-gray-400">
              当前已是最高会员等级
            </div>
          </div>
        </UCard>

        <!-- 统计信息 -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <UCard>
            <div class="flex items-center">
              <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
                <UIcon
                  name="i-heroicons-arrow-trending-up"
                  class="w-6 h-6 text-green-600 dark:text-green-400"
                />
              </div>
              <div class="ml-4">
                <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
                  总获得成长值
                </p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ totalEarned?.toLocaleString() || '0' }}
                </p>
              </div>
            </div>
          </UCard>

          <UCard>
            <div class="flex items-center">
              <div class="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
                <UIcon
                  name="i-heroicons-arrow-trending-down"
                  class="w-6 h-6 text-red-600 dark:text-red-400"
                />
              </div>
              <div class="ml-4">
                <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
                  总消耗成长值
                </p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ totalSpent?.toLocaleString() || '0' }}
                </p>
              </div>
            </div>
          </UCard>

          <UCard>
            <div class="flex items-center">
              <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                <UIcon
                  name="i-heroicons-calculator"
                  class="w-6 h-6 text-blue-600 dark:text-blue-400"
                />
              </div>
              <div class="ml-4">
                <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
                  净成长值变化
                </p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ netChange?.toLocaleString() || '0' }}
                </p>
              </div>
            </div>
          </UCard>
        </div>
      </div>
    </div>

    <!-- 成长值记录列表 -->
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            成长值记录
          </h2>
          <div class="text-sm text-gray-500 dark:text-gray-400">
            共 {{ totalRecords }} 条记录
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredRecords">
        <template #type-cell="{ row }">
          <UBadge
            :color="row.type === 'earned' ? 'success' : 'error'"
            variant="soft"
          >
            {{ row.type === "earned" ? "获得" : "消耗" }}
          </UBadge>
        </template>

        <template #growthValue-cell="{ row }">
          <div
            :class="[
              'font-medium',
              row.type === 'earned'
                ? 'text-green-600 dark:text-green-400'
                : 'text-red-600 dark:text-red-400',
            ]"
          >
            {{ row.type === "earned" ? "+" : "-" }}{{ row.growthValue?.toLocaleString() || '0' }}
          </div>
        </template>

        <template #balance-cell="{ row }">
          <div class="font-medium text-gray-900 dark:text-white">
            {{ row.balanceAfter?.toLocaleString() || '0' }}
          </div>
        </template>

        <template #source-cell="{ row }">
          <div>
            <div class="font-medium text-gray-900 dark:text-white">
              {{ row.source }}
            </div>
            <div class="text-sm text-gray-500 dark:text-gray-400">
              {{ row.sourceDetail }}
            </div>
          </div>
        </template>
      </UTable>

      <!-- 分页 -->
      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (currentPage - 1) * pageSize + 1 }} 到 {{ Math.min(currentPage * pageSize, totalRecords) }} 条，共 {{ totalRecords }} 条
          </div>
          <UPagination
            v-model="currentPage"
            :page-count="pageCount"
            :total="totalRecords"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>
  </div>
</template>

<script setup lang="ts">
// 获取路由参数
const route = useRoute();
const accountId = route.params.id;

// 页面导航
const router = useRouter();

// 分页
const currentPage = ref(1);
const pageSize = ref(10);
const totalRecords = ref(125);
const pageCount = computed(() => Math.ceil(totalRecords.value / pageSize.value));

// 筛选条件
const filter = ref({
  timeRange: "30d",
  type: "all",
  keyword: "",
});

// 时间范围选项
const timeRangeOptions = [
  { label: "最近7天", value: "7d" },
  { label: "最近30天", value: "30d" },
  { label: "最近90天", value: "90d" },
  { label: "最近1年", value: "1y" },
  { label: "自定义", value: "custom" },
];

// 类型选项
const typeOptions = [
  { label: "全部", value: "all" },
  { label: "获得", value: "earned" },
  { label: "消耗", value: "spent" },
];

// 账户信息
const account = ref({
  id: accountId,
  customerName: "张三",
  customerPhone: "13800138001",
  avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1",
  balance: 12500,
  levelName: "钻石会员",
  levelColor: "#B9F2FF",
});

// 统计数据
const totalEarned = ref(18250);
const totalSpent = ref(4000);
const netChange = computed(() => totalEarned.value - totalSpent.value);

// 等级晋升规则数据
const allLevelRules = ref([
  {
    id: 1,
    levelName: "青铜会员",
    levelColor: "#CD7F32",
    growthValueThreshold: 0,
  },
  {
    id: 2,
    levelName: "白银会员",
    levelColor: "#C0C0C0",
    growthValueThreshold: 500,
  },
  {
    id: 3,
    levelName: "黄金会员",
    levelColor: "#FFD700",
    growthValueThreshold: 2000,
  },
  {
    id: 4,
    levelName: "铂金会员",
    levelColor: "#E5E4E2",
    growthValueThreshold: 5000,
  },
  {
    id: 5,
    levelName: "钻石会员",
    levelColor: "#B9F2FF",
    growthValueThreshold: 15000,
  },
  {
    id: 6,
    levelName: "黑金会员",
    levelColor: "#000000",
    growthValueThreshold: 50000,
  }
]);

// 下一级晋升规则
const nextLevelRules = computed(() => {
  // 找到当前等级
  const currentLevelIndex = allLevelRules.value.findIndex(
    rule => rule.levelName === account.value.levelName
  );

  // 如果当前不是最高等级，返回下一个等级的规则
  if (currentLevelIndex >= 0 && currentLevelIndex < allLevelRules.value.length - 1) {
    return [allLevelRules.value[currentLevelIndex + 1]];
  }

  // 如果已是最高等级，返回空数组
  return [];
});

// 成长值记录数据
const records = ref([
  {
    id: "rec_1",
    type: "earned",
    growthValue: 1200,
    balanceAfter: 12500,
    source: "购物奖励",
    sourceDetail: "订单号: ORD20230915001",
    createdAt: "2023-09-15 14:30:25",
  },
  {
    id: "rec_2",
    type: "spent",
    growthValue: 800,
    balanceAfter: 11300,
    source: "等级晋升",
    sourceDetail: "从白银会员晋升到黄金会员",
    createdAt: "2023-09-12 10:15:42",
  },
  {
    id: "rec_3",
    type: "earned",
    growthValue: 650,
    balanceAfter: 12100,
    source: "签到奖励",
    sourceDetail: "连续签到第7天",
    createdAt: "2023-09-10 08:45:17",
  },
  {
    id: "rec_4",
    type: "earned",
    growthValue: 2000,
    balanceAfter: 11450,
    source: "活动奖励",
    sourceDetail: "中秋节活动",
    createdAt: "2023-09-08 16:22:33",
  },
  {
    id: "rec_5",
    type: "spent",
    growthValue: 1200,
    balanceAfter: 9450,
    source: "等级晋升",
    sourceDetail: "从青铜会员晋升到白银会员",
    createdAt: "2023-09-05 11:38:56",
  },
  {
    id: "rec_6",
    type: "earned",
    growthValue: 500,
    balanceAfter: 10650,
    source: "推荐奖励",
    sourceDetail: "推荐好友注册",
    createdAt: "2023-09-01 09:12:44",
  },
  {
    id: "rec_7",
    type: "earned",
    growthValue: 800,
    balanceAfter: 10150,
    source: "购物奖励",
    sourceDetail: "订单号: ORD20230828002",
    createdAt: "2023-08-28 19:55:31",
  },
  {
    id: "rec_8",
    type: "spent",
    growthValue: 500,
    balanceAfter: 9350,
    source: "等级晋升",
    sourceDetail: "从青铜会员晋升到白银会员",
    createdAt: "2023-08-25 14:27:18",
  },
  {
    id: "rec_9",
    type: "earned",
    growthValue: 300,
    balanceAfter: 9850,
    source: "评价奖励",
    sourceDetail: "商品评价: PROD20230820001",
    createdAt: "2023-08-20 20:15:33",
  },
  {
    id: "rec_10",
    type: "earned",
    growthValue: 1000,
    balanceAfter: 9550,
    source: "购物奖励",
    sourceDetail: "订单号: ORD20230815003",
    createdAt: "2023-08-15 15:42:17",
  },
  {
    id: "rec_11",
    type: "earned",
    growthValue: 200,
    balanceAfter: 8550,
    source: "签到奖励",
    sourceDetail: "连续签到第3天",
    createdAt: "2023-08-10 09:30:45",
  },
  {
    id: "rec_12",
    type: "earned",
    growthValue: 1500,
    balanceAfter: 8350,
    source: "活动奖励",
    sourceDetail: "七夕节活动",
    createdAt: "2023-08-05 18:22:11",
  },
  {
    id: "rec_13",
    type: "earned",
    growthValue: 400,
    balanceAfter: 6850,
    source: "分享奖励",
    sourceDetail: "分享商品到微博",
    createdAt: "2023-07-28 12:15:33",
  },
  {
    id: "rec_14",
    type: "earned",
    growthValue: 700,
    balanceAfter: 6450,
    source: "购物奖励",
    sourceDetail: "订单号: ORD20230725001",
    createdAt: "2023-07-25 16:45:21",
  },
  {
    id: "rec_15",
    type: "spent",
    growthValue: 1000,
    balanceAfter: 5750,
    source: "等级晋升",
    sourceDetail: "从青铜会员晋升到白银会员",
    createdAt: "2023-07-20 11:22:44",
  },
]);

// 过滤后的记录
const filteredRecords = computed(() => {
  let result = [...records.value];

  // 根据类型筛选
  if (filter.value.type !== "all") {
    result = result.filter((record) => record.type === filter.value.type);
  }

  // 根据关键词筛选
  if (filter.value.keyword) {
    result = result.filter(
      (record) =>
        record.source.toLowerCase().includes(filter.value.keyword.toLowerCase()) ||
        record.sourceDetail.toLowerCase().includes(filter.value.keyword.toLowerCase())
    );
  }

  return result;
});

// 表格列定义
const columns = [
  { accessorKey: "createdAt", header: "时间" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "growthValue", header: "成长值变化" },
  { accessorKey: "balance", header: "变动后余额" },
  { accessorKey: "source", header: "来源/用途" },
];

// 返回上一页
const goBack = () => {
  router.back();
};

// 应用筛选
const applyFilter = () => {
  // 实际应用筛选逻辑
  console.log("应用筛选:", filter.value);
};

// 重置筛选
const resetFilter = () => {
  filter.value = {
    timeRange: "30d",
    type: "all",
    keyword: "",
  };
};

// 导出数据
const exportData = () => {
  alert("导出数据功能待实现");
};
</script>
