<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          积分账户
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理客户积分账户、查看积分明细和批量调整积分
        </p>
      </div>
      <div class="flex gap-3">
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="openBatchAdjustModal"
        >
          批量调整积分
        </UButton>
      </div>
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
              总积分余额
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
              今日消耗
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ todaySpent?.toLocaleString() || '0' }}
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
            积分账户列表
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
              @click="adjustPoints(row, 'add')"
            >
              赠送
            </UButton>
            <UButton
              color="warning"
              variant="ghost"
              size="sm"
              icon="i-heroicons-minus-circle"
              @click="adjustPoints(row, 'deduct')"
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

    <!-- 批量调整积分模态框 -->
    <BatchAdjustPointsModal
      v-model:open="showBatchAdjustModal"
      mode="batch"
      :customers="customers"
      @submit="handleBatchAdjustSubmit"
      @close="handleBatchAdjustClose"
    />

    <!-- 调整积分模态框 -->
    <BatchAdjustPointsModal
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
import BatchAdjustPointsModal from "~/components/Modals/BatchAdjustPointsModal.vue";

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
const todaySpent = ref(8765);

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
    points: 12500,
    selected: false,
  },
  {
    id: "cust_2",
    name: "李四",
    phone: "13800138002",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2",
    points: 8650,
    selected: false,
  },
  {
    id: "cust_3",
    name: "王五",
    phone: "13800138003",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3",
    points: 5200,
    selected: false,
  },
  {
    id: "cust_4",
    name: "赵六",
    phone: "13800138004",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4",
    points: 2450,
    selected: false,
  },
  {
    id: "cust_5",
    name: "孙七",
    phone: "13800138005",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=5",
    points: 800,
    selected: false,
  },
]);

// 表格列定义
const columns = [
  { accessorKey: "customer", header: "客户" },
  { accessorKey: "level", header: "会员等级" },
  { accessorKey: "balance", header: "积分余额" },
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

    // 更新账户积分
    accounts.value.forEach((account) => {
      const customer = customers.value.find((c) =>
        data.customerIds.includes(c.id) && c.id === account.customerId
      );
      if (customer) {
        if (data.type === "add") {
          account.balance += data.points;
        } else {
          account.balance = Math.max(0, account.balance - data.points);
        }
      }
    });

    // 显示成功消息
    alert(`成功为 ${data.customerIds.length} 个客户${data.type === "add" ? "赠送" : "扣减"}积分`);
  } catch (error) {
    console.error("批量调整积分失败:", error);
    alert("批量调整积分失败，请重试");
  }
};

// 处理批量调整关闭
const handleBatchAdjustClose = () => {
  showBatchAdjustModal.value = false;
};

// 打开调整模态框
const adjustPoints = (account: any, type: string) => {
  currentAccount.value = account;
  showAdjustModal.value = true;
};

// 处理调整提交
const handleAdjustSubmit = async (data) => {
  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // 更新账户积分
    const account = accounts.value.find((a) => a.id === data.accountId);
    if (account) {
      if (data.type === "add") {
        account.balance += data.points;
      } else {
        account.balance = Math.max(0, account.balance - data.points);
      }
    }

    // 显示成功消息
    alert(`${data.type === "add" ? "赠送" : "扣减"}积分成功`);
  } catch (error) {
    console.error("调整积分失败:", error);
    alert("调整积分失败，请重试");
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
  navigateTo(`/customer/membership/points/accounts/${account.id}`);
};
</script>
