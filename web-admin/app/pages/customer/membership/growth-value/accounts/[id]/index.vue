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
            成长值账户详情
          </h1>
          <p class="text-gray-600 dark:text-gray-400 mt-1">
            查看和管理客户成长值账户
          </p>
        </div>
      </div>
      <div class="flex gap-3">
        <UButton
          color="primary"
          icon="i-heroicons-plus-circle"
          @click="adjustGrowthValue('add')"
        >
          赠送成长值
        </UButton>
        <UButton
          color="warning"
          icon="i-heroicons-minus-circle"
          @click="adjustGrowthValue('deduct')"
        >
          扣减成长值
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

    <!-- 账户信息卡片 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
      <!-- 客户基本信息 -->
      <UCard class="lg:col-span-2">
        <template #header>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            客户信息
          </h2>
        </template>
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
        </div>
        <div class="mt-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
              客户ID
            </label>
            <div class="mt-1 text-gray-900 dark:text-white">
              {{ account.customerId }}
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
              注册时间
            </label>
            <div class="mt-1 text-gray-900 dark:text-white">
              {{ account.registeredAt }}
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
              当前等级
            </label>
            <div class="mt-1 flex items-center gap-2">
              <div
                class="w-3 h-3 rounded-full"
                :style="{ backgroundColor: account.levelColor }"
              ></div>
              <span class="text-gray-900 dark:text-white">{{ account.levelName }}</span>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
              等级有效期
            </label>
            <div class="mt-1 text-gray-900 dark:text-white">
              {{ account.levelExpiry }}
            </div>
          </div>
        </div>
      </UCard>

      <!-- 账户统计 -->
      <div class="space-y-6">
        <UCard>
          <div class="text-center">
            <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
              当前成长值余额
            </div>
            <div class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">
              {{ account.balance?.toLocaleString() || '0' }}
            </div>
            <div class="mt-4 flex justify-center">
              <UBadge color="primary" variant="soft">
                {{ account.levelName }}
              </UBadge>
            </div>
          </div>
        </UCard>

        <UCard>
          <div class="space-y-4">
            <div class="flex justify-between">
              <div class="text-sm text-gray-500 dark:text-gray-400">总获得成长值</div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ totalEarned?.toLocaleString() || '0' }}
              </div>
            </div>
            <div class="flex justify-between">
              <div class="text-sm text-gray-500 dark:text-gray-400">总消耗成长值</div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ totalSpent?.toLocaleString() || '0' }}
              </div>
            </div>
            <div class="flex justify-between">
              <div class="text-sm text-gray-500 dark:text-gray-400">净成长值变化</div>
              <div
                :class="[
                  'font-medium',
                  netChange >= 0
                    ? 'text-green-600 dark:text-green-400'
                    : 'text-red-600 dark:text-red-400'
                ]"
              >
                {{ netChange >= 0 ? '+' : '' }}{{ netChange?.toLocaleString() || '0' }}
              </div>
            </div>
          </div>
        </UCard>

        <!-- 等级晋升进度 -->
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

            <div v-if="nextLevelRules.length === 0" class="text-center py-4 text-gray-500 dark:text-gray-400">
              当前已是最高会员等级
            </div>
          </div>
        </UCard>
      </div>
    </div>

    <!-- 最近成长值记录 -->
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            最近成长值记录
          </h2>
          <UButton
            :to="`/customer/membership/growth-value/accounts/${accountId}/details`"
            variant="ghost"
            size="sm"
            trailing-icon="i-heroicons-arrow-right"
          >
            查看全部
          </UButton>
        </div>
      </template>

      <div class="overflow-hidden">
        <UTable :columns="recentColumns" :data="recentRecords">
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
        </UTable>
      </div>
    </UCard>

    <!-- 调整成长值模态框 -->
    <UModal
      v-model:open="showAdjustModal"
      :title="adjustType === 'add' ? '赠送成长值' : '扣减成长值'"
      :description="adjustType === 'add' ? '为客户账户增加成长值' : '从客户账户扣减成长值'"
      :ui="{
        content: 'w-full sm:max-w-md',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-6 p-4 sm:p-6">
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  客户
                </label>
                <div class="flex items-center gap-3">
                  <UAvatar
                    :src="account.avatar"
                    :alt="account.customerName"
                    size="sm"
                    :ui="{ rounded: 'rounded-full' }"
                  />
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">
                      {{ account.customerName }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      {{ account.customerPhone }}
                    </div>
                  </div>
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  当前成长值余额
                </label>
                <div class="text-lg font-bold text-gray-900 dark:text-white">
                  {{ account.balance?.toLocaleString() || '0' }}
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  调整数量 <span class="text-red-500">*</span>
                </label>
                <UInput
                  v-model.number="adjustForm.growthValue"
                  type="number"
                  placeholder="请输入调整数量"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  调整原因 <span class="text-red-500">*</span>
                </label>
                <UTextarea
                  v-model="adjustForm.reason"
                  placeholder="请输入调整原因"
                  rows="3"
                />
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeAdjustModal">取消</UButton>
          <UButton
            color="primary"
            :loading="adjustLoading"
            :disabled="!isAdjustFormValid"
            @click="saveAdjustment"
          >
            确认调整
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
// 获取路由参数
const route = useRoute();
const accountId = route.params.id;

// 页面导航
const router = useRouter();

// 模态框状态
const showAdjustModal = ref(false);
const adjustType = ref<"add" | "deduct">("add");
const adjustLoading = ref(false);

// 调整表单
const adjustForm = ref({
  growthValue: 0,
  reason: "",
});

// 表单验证
const isAdjustFormValid = computed(() => {
  return adjustForm.value.growthValue > 0 && adjustForm.value.reason.trim() !== "";
});

// 账户信息
const account = ref({
  id: accountId,
  customerId: "cust_1",
  customerName: "张三",
  customerPhone: "13800138001",
  avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1",
  balance: 12500,
  levelName: "钻石会员",
  levelColor: "#B9F2FF",
  registeredAt: "2022-03-15",
  levelExpiry: "2024-03-15",
});

// 统计数据
const totalEarned = ref(17500);
const totalSpent = ref(2000);
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

// 最近记录数据
const recentRecords = ref([
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
    growthValue: 300,
    balanceAfter: 10650,
    source: "评价奖励",
    sourceDetail: "商品评价: PROD20230901001",
    createdAt: "2023-09-01 19:22:15",
  },
  {
    id: "rec_7",
    type: "earned",
    growthValue: 500,
    balanceAfter: 10350,
    source: "分享奖励",
    sourceDetail: "分享商品到朋友圈",
    createdAt: "2023-08-28 14:15:33",
  },
  {
    id: "rec_8",
    type: "earned",
    growthValue: 1000,
    balanceAfter: 9850,
    source: "购物奖励",
    sourceDetail: "订单号: ORD20230825002",
    createdAt: "2023-08-25 16:45:21",
  },
]);

// 表格列定义
const recentColumns = [
  { accessorKey: "createdAt", header: "时间" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "growthValue", header: "成长值变化" },
  { accessorKey: "balance", header: "变动后余额" },
  { accessorKey: "source", header: "来源/用途" },
];

// 返回上一页
const goBack = () => {
  router.push("/customer/membership/growth-value");
};

// 调整成长值
const adjustGrowthValue = (type: "add" | "deduct") => {
  adjustType.value = type;
  adjustForm.value = {
    growthValue: 0,
    reason: "",
  };
  showAdjustModal.value = true;
};

// 关闭调整模态框
const closeAdjustModal = () => {
  showAdjustModal.value = false;
};

// 保存调整
const saveAdjustment = async () => {
  if (!isAdjustFormValid.value) return;

  adjustLoading.value = true;

  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // 更新账户余额
    if (adjustType.value === "add") {
      account.value.balance += adjustForm.value.growthValue;
      totalEarned.value += adjustForm.value.growthValue;
    } else {
      account.value.balance = Math.max(0, account.value.balance - adjustForm.value.growthValue);
      totalSpent.value += adjustForm.value.growthValue;
    }

    // 添加到最近记录
    recentRecords.value.unshift({
      id: `rec_${Date.now()}`,
      type: adjustType.value === "add" ? "earned" : "spent",
      growthValue: adjustForm.value.growthValue,
      balanceAfter: account.value.balance,
      source: adjustType.value === "add" ? "管理员调整" : "管理员调整",
      sourceDetail: adjustForm.value.reason,
      createdAt: new Date().toLocaleString("zh-CN", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit"
      }).replace(/\//g, "-").replace(",", ""),
    });

    // 限制最近记录数量为5条
    if (recentRecords.value.length > 5) {
      recentRecords.value.pop();
    }

    // 显示成功消息
    alert(`${adjustType.value === "add" ? "赠送" : "扣减"}成长值成功`);

    // 关闭模态框
    closeAdjustModal();
  } catch (error) {
    console.error("调整成长值失败:", error);
    alert("调整成长值失败，请重试");
  } finally {
    adjustLoading.value = false;
  }
};
</script>
