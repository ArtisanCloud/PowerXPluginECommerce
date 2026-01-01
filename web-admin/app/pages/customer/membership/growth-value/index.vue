<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          成长值账户
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理客户成长值账户和查看成长值明细
        </p>
      </div>
      <div class="flex gap-3">
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="openBatchAdjustModal"
        >
          批量调整成长值
        </UButton>
      </div>
    </div>

    <!-- 等级晋升规则 -->
    <div class="mb-6">
      <UCard>
        <template #header>
          <div class="flex justify-between items-center">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              等级晋升规则
            </h2>
            <UButton
              to="/customer/membership/growth-value/promotion-rules"
              variant="ghost"
              size="sm"
              trailing-icon="i-heroicons-arrow-right"
            >
              管理规则
            </UButton>
          </div>
        </template>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead>
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">等级</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">晋升类型</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">成长值要求</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">状态</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-for="rule in promotionRules" :key="rule.id">
                <td class="px-4 py-3 whitespace-nowrap">
                  <div class="flex items-center gap-2">
                    <div
                      class="w-3 h-3 rounded-full"
                      :style="{ backgroundColor: rule.levelColor }"
                    ></div>
                    <span>{{ rule.levelName }}</span>
                  </div>
                </td>
                <td class="px-4 py-3 whitespace-nowrap">
                  <UBadge
                    :color="rule.type === 'auto' ? 'success' : 'warning'"
                    variant="soft"
                  >
                    {{ rule.type === "auto" ? "自动" : "手动" }}
                  </UBadge>
                </td>
                <td class="px-4 py-3 whitespace-nowrap text-gray-900 dark:text-white">
                  {{ rule.growthValueThreshold?.toLocaleString() || 'N/A' }}
                </td>
                <td class="px-4 py-3 whitespace-nowrap">
                  <UBadge
                    :color="rule.status === 'active' ? 'success' : 'neutral'"
                    variant="soft"
                  >
                    {{ rule.status === "active" ? "启用" : "停用" }}
                  </UBadge>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </UCard>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon
              name="i-heroicons-user-group"
              class="w-6 h-6 text-blue-600 dark:text-blue-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              总账户数
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ totalAccounts?.toLocaleString() || '0' }}
            </p>
          </div>
        </div>
      </UCard>

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
              总成长值余额
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ totalBalance?.toLocaleString() || '0' }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon
              name="i-heroicons-arrow-up-circle"
              class="w-6 h-6 text-yellow-600 dark:text-yellow-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              今日获得
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ todayEarned?.toLocaleString() || '0' }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon
              name="i-heroicons-arrow-down-circle"
              class="w-6 h-6 text-purple-600 dark:text-purple-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              等级晋升
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ recentPromotions }}
            </p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 账户列表 -->
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            成长值账户列表
          </h2>
          <div class="flex gap-2">
            <UInput
              v-model="searchQuery"
              placeholder="搜索客户姓名或手机号..."
              icon="i-heroicons-magnifying-glass"
              size="sm"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredAccounts">
        <template #customer-cell="{ row }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="row.avatar"
              :alt="row.customerName"
              size="md"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ row.customerName }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ row.customerPhone }}
              </div>
            </div>
          </div>
        </template>

        <template #level-cell="{ row }">
          <div class="flex items-center gap-2">
            <div
              class="w-3 h-3 rounded-full"
              :style="{ backgroundColor: row.levelColor }"
            ></div>
            <span>{{ row.levelName }}</span>
          </div>
        </template>

        <template #balance-cell="{ row }">
          <div class="font-medium text-gray-900 dark:text-white">
            {{ row.balance?.toLocaleString() || '0' }}
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewAccountDetails(row)"
            >
              查看
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-plus-circle"
              @click="adjustGrowthValue(row, 'add')"
            >
              赠送
            </UButton>
            <UButton
              color="warning"
              variant="ghost"
              size="sm"
              icon="i-heroicons-minus-circle"
              @click="adjustGrowthValue(row, 'deduct')"
            >
              扣减
            </UButton>
          </div>
        </template>
      </UTable>

      <!-- 分页 -->
      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (currentPage - 1) * pageSize + 1 }} 到 {{ Math.min(currentPage * pageSize, totalAccounts) }} 条，共 {{ totalAccounts }} 条
          </div>
          <UPagination
            v-model="currentPage"
            :page-count="pageCount"
            :total="totalAccounts"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 批量调整成长值模态框 -->
    <BatchAdjustGrowthValueModal
      v-model:open="showBatchAdjustModal"
      mode="batch"
      :customers="customers"
      @submit="handleBatchAdjustSubmit"
      @close="handleBatchAdjustClose"
    />

    <!-- 调整成长值模态框 -->
    <BatchAdjustGrowthValueModal
      v-model:open="showAdjustModal"
      mode="single"
      :account="currentAccount"
      @submit="handleAdjustSubmit"
      @close="handleAdjustClose"
    />
  </div>
</template>

<script setup lang="ts">
// 导入模态框组件
import BatchAdjustGrowthValueModal from "~/components/modals/BatchAdjustGrowthValueModal.vue";

// 模态框状态
const showBatchAdjustModal = ref(false);
const showAdjustModal = ref(false);

// 当前操作的账户
const currentAccount = ref<any>(null);

// 搜索
const searchQuery = ref("");

// 分页
const currentPage = ref(1);
const pageSize = ref(10);
const totalAccounts = ref(1245);
const pageCount = computed(() => Math.ceil(totalAccounts.value / pageSize.value));

// 统计数据
const totalBalance = ref(2456789);
const todayEarned = ref(12450);
const recentPromotions = ref(23);

// 等级晋升规则数据
const promotionRules = ref([
  {
    id: 1,
    levelName: "青铜会员",
    levelColor: "#CD7F32",
    type: "auto",
    growthValueThreshold: 0,
    status: "active"
  },
  {
    id: 2,
    levelName: "白银会员",
    levelColor: "#C0C0C0",
    type: "auto",
    growthValueThreshold: 500,
    status: "active"
  },
  {
    id: 3,
    levelName: "黄金会员",
    levelColor: "#FFD700",
    type: "auto",
    growthValueThreshold: 2000,
    status: "active"
  },
  {
    id: 4,
    levelName: "铂金会员",
    levelColor: "#E5E4E2",
    type: "auto",
    growthValueThreshold: 5000,
    status: "active"
  },
  {
    id: 5,
    levelName: "钻石会员",
    levelColor: "#B9F2FF",
    type: "auto",
    growthValueThreshold: 15000,
    status: "active"
  },
  {
    id: 6,
    levelName: "黑金会员",
    levelColor: "#000000",
    type: "manual",
    growthValueThreshold: 50000,
    status: "active"
  },
]);

// 账户数据
const accounts = ref([
  {
    id: "acc_1",
    customerName: "张三",
    customerPhone: "13800138001",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1",
    levelName: "钻石会员",
    levelColor: "#B9F2FF",
    balance: 12500,
    customerId: "cust_1",
  },
  {
    id: "acc_2",
    customerName: "李四",
    customerPhone: "13800138002",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2",
    levelName: "铂金会员",
    levelColor: "#E5E4E2",
    balance: 8650,
    customerId: "cust_2",
  },
  {
    id: "acc_3",
    customerName: "王五",
    customerPhone: "13800138003",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3",
    levelName: "黄金会员",
    levelColor: "#FFD700",
    balance: 5200,
    customerId: "cust_3",
  },
  {
    id: "acc_4",
    customerName: "赵六",
    customerPhone: "13800138004",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4",
    levelName: "白银会员",
    levelColor: "#C0C0C0",
    balance: 2450,
    customerId: "cust_4",
  },
  {
    id: "acc_5",
    customerName: "孙七",
    customerPhone: "13800138005",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=5",
    levelName: "青铜会员",
    levelColor: "#CD7F32",
    balance: 800,
    customerId: "cust_5",
  },
  {
    id: "acc_6",
    customerName: "周八",
    customerPhone: "13800138006",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=6",
    levelName: "钻石会员",
    levelColor: "#B9F2FF",
    balance: 14200,
    customerId: "cust_6",
  },
  {
    id: "acc_7",
    customerName: "吴九",
    customerPhone: "13800138007",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=7",
    levelName: "铂金会员",
    levelColor: "#E5E4E2",
    balance: 7800,
    customerId: "cust_7",
  },
  {
    id: "acc_8",
    customerName: "郑十",
    customerPhone: "13800138008",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=8",
    levelName: "黄金会员",
    levelColor: "#FFD700",
    balance: 4500,
    customerId: "cust_8",
  },
  {
    id: "acc_9",
    customerName: "王芳",
    customerPhone: "13800138009",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=9",
    levelName: "白银会员",
    levelColor: "#C0C0C0",
    balance: 1800,
    customerId: "cust_9",
  },
  {
    id: "acc_10",
    customerName: "李明",
    customerPhone: "13800138010",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=10",
    levelName: "青铜会员",
    levelColor: "#CD7F32",
    balance: 350,
    customerId: "cust_10",
  },
]);

// 过滤后的账户列表
const filteredAccounts = computed(() => {
  if (!searchQuery.value) return accounts.value;
  return accounts.value.filter(
    (account) =>
      account.customerName.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      account.customerPhone.includes(searchQuery.value)
  );
});

// 客户数据（用于批量调整）
const customers = ref([
  {
    id: "cust_1",
    name: "张三",
    phone: "13800138001",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1",
    growthValue: 12500,
    selected: false,
  },
  {
    id: "cust_2",
    name: "李四",
    phone: "13800138002",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2",
    growthValue: 8650,
    selected: false,
  },
  {
    id: "cust_3",
    name: "王五",
    phone: "13800138003",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3",
    growthValue: 5200,
    selected: false,
  },
  {
    id: "cust_4",
    name: "赵六",
    phone: "13800138004",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4",
    growthValue: 2450,
    selected: false,
  },
  {
    id: "cust_5",
    name: "孙七",
    phone: "13800138005",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=5",
    growthValue: 800,
    selected: false,
  },
  {
    id: "cust_6",
    name: "周八",
    phone: "13800138006",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=6",
    growthValue: 14200,
    selected: false,
  },
  {
    id: "cust_7",
    name: "吴九",
    phone: "13800138007",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=7",
    growthValue: 7800,
    selected: false,
  },
  {
    id: "cust_8",
    name: "郑十",
    phone: "13800138008",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=8",
    growthValue: 4500,
    selected: false,
  },
  {
    id: "cust_9",
    name: "王芳",
    phone: "13800138009",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=9",
    growthValue: 1800,
    selected: false,
  },
  {
    id: "cust_10",
    name: "李明",
    phone: "13800138010",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=10",
    growthValue: 350,
    selected: false,
  },
]);

// 表格列定义
const columns = [
  { accessorKey: "customer", header: "客户" },
  { accessorKey: "level", header: "会员等级" },
  { accessorKey: "balance", header: "成长值余额" },
  { id: "actions", header: "操作" },
];

// 打开批量调整模态框
const openBatchAdjustModal = () => {
  showBatchAdjustModal.value = true;
};

// 处理批量调整提交
const handleBatchAdjustSubmit = async (data) => {
  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // 更新账户成长值
    accounts.value.forEach((account) => {
      const customer = customers.value.find((c) =>
        data.customerIds.includes(c.id) && c.id === account.customerId
      );
      if (customer) {
        if (data.type === "add") {
          account.balance += data.growthValue;
        } else {
          account.balance = Math.max(0, account.balance - data.growthValue);
        }
      }
    });

    // 显示成功消息
    alert(`成功为 ${data.customerIds.length} 个客户${data.type === "add" ? "赠送" : "扣减"}成长值`);
  } catch (error) {
    console.error("批量调整成长值失败:", error);
    alert("批量调整成长值失败，请重试");
  }
};

// 处理批量调整关闭
const handleBatchAdjustClose = () => {
  showBatchAdjustModal.value = false;
};

// 打开调整模态框
const adjustGrowthValue = (account: any, type: string) => {
  currentAccount.value = account;
  showAdjustModal.value = true;
};

// 处理调整提交
const handleAdjustSubmit = async (data) => {
  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // 更新账户成长值
    const account = accounts.value.find((a) => a.id === data.accountId);
    if (account) {
      if (data.type === "add") {
        account.balance += data.growthValue;
      } else {
        account.balance = Math.max(0, account.balance - data.growthValue);
      }
    }

    // 显示成功消息
    alert(`${data.type === "add" ? "赠送" : "扣减"}成长值成功`);
  } catch (error) {
    console.error("调整成长值失败:", error);
    alert("调整成长值失败，请重试");
  }
};

// 处理调整关闭
const handleAdjustClose = () => {
  showAdjustModal.value = false;
  currentAccount.value = null;
};

// 查看账户详情
const viewAccountDetails = (account: any) => {
  // 跳转到账户详情页面
  navigateTo(`/customer/membership/growth-value/accounts/${account.id}`);
};
</script>
