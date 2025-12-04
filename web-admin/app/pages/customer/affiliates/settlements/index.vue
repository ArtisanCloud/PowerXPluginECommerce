<template>
  <div class="p-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">佣金结算</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理分销员的佣金结算情况</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-arrow-down-tray" @click="exportData">导出数据</UButton>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon name="i-heroicons-clock" class="w-6 h-6 text-yellow-600 dark:text-yellow-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">待结算</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ pendingCount.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-check-circle" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">已结算</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ settledCount.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-banknotes" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">可提现</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">¥{{ withdrawableAmount.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 标签页 -->
    <div class="mb-6">
      <UTabs v-model="activeTab" :items="tabs" />
    </div>

    <!-- 待结算 -->
    <UCard v-if="activeTab === 'pending'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">待结算</h2>
          <div class="flex gap-3">
            <UInput v-model="pendingSearchQuery" placeholder="搜索分销员..." icon="i-heroicons-magnifying-glass" size="sm" />
            <UButton color="neutral" @click="resetPendingFilters">重置</UButton>
          </div>
        </div>
      </template>

      <UTable :columns="pendingColumns" :data="pendingTableData">
        <!-- 分销员 -->
        <template #distributor-cell="{ row }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="getRowData(row).avatar || avatarPlaceholder"
              :alt="getRowData(row).distributor || '分销员'"
              size="md"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ getRowData(row).distributor || '—' }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                ID: {{ getRowData(row).distributorId || '—' }}
              </div>
            </div>
          </div>
        </template>

        <!-- 订单金额 -->
        <template #orderAmount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 佣金金额 -->
        <template #commissionAmount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 佣金比例 -->
        <template #commissionRate-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ getValue(row) }}%
          </div>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewPendingCommission(getRowData(row))"
            >
              查看
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-check"
              @click="settleCommission(getRowData(row))"
            >
              结算
            </UButton>
          </div>
        </template>
      </UTable>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (pendingCurrentPage - 1) * pendingPageSize + 1 }}
            到 {{ Math.min(pendingCurrentPage * pendingPageSize, filteredPendingItems.length) }} 条，共 {{ filteredPendingItems.length }} 条
          </div>
          <div class="flex items-center gap-4">
            <UPagination
              v-model="pendingCurrentPage"
              :page-count="pendingPageCount"
              :total="filteredPendingItems.length"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500 dark:text-gray-400">每页</span>
              <USelect
                v-model="pendingPageSize"
                :options="[5, 10, 20, 50]"
                size="sm"
                class="w-20"
              />
              <span class="text-sm text-gray-500 dark:text-gray-400">条</span>
            </div>
          </div>
        </div>
      </template>
    </UCard>

    <!-- 已结算 -->
    <UCard v-if="activeTab === 'settled'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">已结算</h2>
          <div class="flex gap-3">
            <UInput v-model="settledSearchQuery" placeholder="搜索分销员..." icon="i-heroicons-magnifying-glass" size="sm" />
            <UButton color="neutral" @click="resetSettledFilters">重置</UButton>
          </div>
        </div>
      </template>

      <UTable :columns="settledColumns" :data="settledTableData">
        <!-- 分销员 -->
        <template #distributor-cell="{ row }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="getRowData(row).avatar || avatarPlaceholder"
              :alt="getRowData(row).distributor || '分销员'"
              size="md"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ getRowData(row).distributor || '—' }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                ID: {{ getRowData(row).distributorId || '—' }}
              </div>
            </div>
          </div>
        </template>

        <!-- 订单金额 -->
        <template #orderAmount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 佣金金额 -->
        <template #commissionAmount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 佣金比例 -->
        <template #commissionRate-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ getValue(row) }}%
          </div>
        </template>

        <!-- 结算时间 -->
        <template #settledAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ formatDate(getValue(row)) }}
          </div>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <UButton
            color="neutral"
            variant="ghost"
            size="sm"
            icon="i-heroicons-eye"
            @click="viewSettledCommission(getRowData(row))"
          >
            查看
          </UButton>
        </template>
      </UTable>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (settledCurrentPage - 1) * settledPageSize + 1 }}
            到 {{ Math.min(settledCurrentPage * settledPageSize, filteredSettledItems.length) }} 条，共 {{ filteredSettledItems.length }} 条
          </div>
          <div class="flex items-center gap-4">
            <UPagination
              v-model="settledCurrentPage"
              :page-count="settledPageCount"
              :total="filteredSettledItems.length"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500 dark:text-gray-400">每页</span>
              <USelect
                v-model="settledPageSize"
                :options="[5, 10, 20, 50]"
                size="sm"
                class="w-20"
              />
              <span class="text-sm text-gray-500 dark:text-gray-400">条</span>
            </div>
          </div>
        </div>
      </template>
    </UCard>

    <!-- 可提现 -->
    <UCard v-if="activeTab === 'withdrawable'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">可提现</h2>
          <div class="flex gap-3">
            <UInput v-model="withdrawableSearchQuery" placeholder="搜索分销员..." icon="i-heroicons-magnifying-glass" size="sm" />
            <UButton color="neutral" @click="resetWithdrawableFilters">重置</UButton>
          </div>
        </div>
      </template>

      <UTable :columns="withdrawableColumns" :data="withdrawableTableData">
        <!-- 分销员 -->
        <template #distributor-cell="{ row }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="getRowData(row).avatar || avatarPlaceholder"
              :alt="getRowData(row).distributor || '分销员'"
              size="md"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ getRowData(row).distributor || '—' }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                ID: {{ getRowData(row).distributorId || '—' }}
              </div>
            </div>
          </div>
        </template>

        <!-- 可提现金额 -->
        <template #withdrawableAmount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 累计佣金 -->
        <template #totalCommission-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 已提现金额 -->
        <template #withdrawnAmount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewWithdrawableItem(getRowData(row))"
            >
              查看
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-arrow-up-tray"
              @click="requestWithdrawal(getRowData(row))"
            >
              提现申请
            </UButton>
          </div>
        </template>
      </UTable>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (withdrawableCurrentPage - 1) * withdrawablePageSize + 1 }}
            到 {{ Math.min(withdrawableCurrentPage * withdrawablePageSize, filteredWithdrawableItems.length) }} 条，共 {{ filteredWithdrawableItems.length }} 条
          </div>
          <div class="flex items-center gap-4">
            <UPagination
              v-model="withdrawableCurrentPage"
              :page-count="withdrawablePageCount"
              :total="filteredWithdrawableItems.length"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500 dark:text-gray-400">每页</span>
              <USelect
                v-model="withdrawablePageSize"
                :options="[5, 10, 20, 50]"
                size="sm"
                class="w-20"
              />
              <span class="text-sm text-gray-500 dark:text-gray-400">条</span>
            </div>
          </div>
        </div>
      </template>
    </UCard>

    <!-- 佣金详情模态框 -->
    <ViewCommissionModal
      v-if="showCommissionModal && currentCommission"
      v-model:open="showCommissionModal"
      :commission="currentCommission"
      @close="closeCommissionModal"
      @settle="handleSettleCommission"
    />

    <!-- 可提现详情模态框 -->
    <ViewCommissionModal
      v-if="showWithdrawableModal && currentWithdrawable"
      v-model:open="showWithdrawableModal"
      :commission="currentWithdrawable"
      @close="closeWithdrawableModal"
    />

    <!-- 提现申请模态框 -->
    <UModal
      v-model:open="showWithdrawalRequestModal"
      title="提现申请"
      description="申请提现可提现金额"
      :close="{ onClick: () => closeWithdrawalRequestModal() }"
      :ui="{
        content: 'w-full sm:max-w-md',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-4 p-4 sm:p-6">
            <div class="grid grid-cols-1 gap-4">
              <div>
                <span class="text-gray-500 dark:text-gray-400">分销员：</span>
                <span class="text-gray-900 dark:text-white font-medium">{{ withdrawalRequest?.distributor || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">可提现金额：</span>
                <span class="text-gray-900 dark:text-white font-medium">¥{{ (withdrawalRequest?.withdrawableAmount || 0).toLocaleString() }}</span>
              </div>
              <div>
                <UInput
                  v-model="withdrawalAmount"
                  label="提现金额"
                  placeholder="请输入提现金额"
                  type="number"
                  :min="0"
                  :max="withdrawalRequest?.withdrawableAmount || 0"
                />
              </div>
              <div>
                <UTextarea
                  v-model="withdrawalRemark"
                  label="备注"
                  placeholder="请输入提现备注（可选）"
                />
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeWithdrawalRequestModal">取消</UButton>
          <UButton color="primary" @click="submitWithdrawalRequest">提交申请</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import ViewCommissionModal from "~/components/Modals/ViewCommissionModal.vue";

const avatarPlaceholder = "https://api.dicebear.com/7.x/miniavs/svg?seed=placeholder";

// 标签页
const activeTab = ref("pending");

// 搜索查询
const pendingSearchQuery = ref("");
const settledSearchQuery = ref("");
const withdrawableSearchQuery = ref("");

// 分页
const pendingCurrentPage = ref(1);
const pendingPageSize = ref(10);

const settledCurrentPage = ref(1);
const settledPageSize = ref(10);

const withdrawableCurrentPage = ref(1);
const withdrawablePageSize = ref(10);

// 模态框
const showCommissionModal = ref(false);
const showWithdrawableModal = ref(false);
const showWithdrawalRequestModal = ref(false);
const currentCommission = ref<any>(null);
const currentWithdrawable = ref<any>(null);
const withdrawalRequest = ref<any>(null);

// 提现申请表单
const withdrawalAmount = ref<number | null>(null);
const withdrawalRemark = ref("");

// Tabs
const tabs = [
  { label: "待结算", value: "pending" },
  { label: "已结算", value: "settled" },
  { label: "可提现", value: "withdrawable" },
];

// 统计数据
const pendingCount = ref(12);
const settledCount = ref(86);
const withdrawableAmount = ref(15670);

// 模拟数据
const pendingItems = ref([
  {
    id: "1",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1",
    distributor: "张三",
    distributorId: "DIS001",
    orderId: "ORD001",
    orderAmount: 299.00,
    commissionRate: 10,
    commissionAmount: 29.90,
    orderDate: "2023-09-15 14:30:25"
  },
  {
    id: "2",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2",
    distributor: "李四",
    distributorId: "DIS002",
    orderId: "ORD002",
    orderAmount: 599.00,
    commissionRate: 15,
    commissionAmount: 89.85,
    orderDate: "2023-09-12 10:15:42"
  },
  {
    id: "3",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3",
    distributor: "王五",
    distributorId: "DIS003",
    orderId: "ORD003",
    orderAmount: 199.00,
    commissionRate: 10,
    commissionAmount: 19.90,
    orderDate: "2023-09-10 08:45:17"
  },
  {
    id: "4",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4",
    distributor: "赵六",
    distributorId: "DIS004",
    orderId: "ORD004",
    orderAmount: 399.00,
    commissionRate: 15,
    commissionAmount: 59.85,
    orderDate: "2023-09-08 16:22:33"
  },
  {
    id: "5",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=5",
    distributor: "孙七",
    distributorId: "DIS005",
    orderId: "ORD005",
    orderAmount: 499.00,
    commissionRate: 12,
    commissionAmount: 59.88,
    orderDate: "2023-09-05 11:38:56"
  }
]);

const settledItems = ref([
  {
    id: "6",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1",
    distributor: "张三",
    distributorId: "DIS001",
    orderId: "ORD006",
    orderAmount: 299.00,
    commissionRate: 10,
    commissionAmount: 29.90,
    settledAt: "2023-09-15 14:30:25"
  },
  {
    id: "7",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2",
    distributor: "李四",
    distributorId: "DIS002",
    orderId: "ORD007",
    orderAmount: 599.00,
    commissionRate: 15,
    commissionAmount: 89.85,
    settledAt: "2023-09-12 10:15:42"
  },
  {
    id: "8",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3",
    distributor: "王五",
    distributorId: "DIS003",
    orderId: "ORD008",
    orderAmount: 199.00,
    commissionRate: 10,
    commissionAmount: 19.90,
    settledAt: "2023-09-10 08:45:17"
  },
  {
    id: "9",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4",
    distributor: "赵六",
    distributorId: "DIS004",
    orderId: "ORD009",
    orderAmount: 399.00,
    commissionRate: 15,
    commissionAmount: 59.85,
    settledAt: "2023-09-08 16:22:33"
  },
  {
    id: "10",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=5",
    distributor: "孙七",
    distributorId: "DIS005",
    orderId: "ORD010",
    orderAmount: 499.00,
    commissionRate: 12,
    commissionAmount: 59.88,
    settledAt: "2023-09-05 11:38:56"
  }
]);

const withdrawableItems = ref([
  {
    id: "11",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1",
    distributor: "张三",
    distributorId: "DIS001",
    withdrawableAmount: 299.50,
    totalCommission: 1299.50,
    withdrawnAmount: 1000.00
  },
  {
    id: "12",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2",
    distributor: "李四",
    distributorId: "DIS002",
    withdrawableAmount: 599.25,
    totalCommission: 2199.25,
    withdrawnAmount: 1600.00
  },
  {
    id: "13",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3",
    distributor: "王五",
    distributorId: "DIS003",
    withdrawableAmount: 199.00,
    totalCommission: 899.00,
    withdrawnAmount: 700.00
  },
  {
    id: "14",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4",
    distributor: "赵六",
    distributorId: "DIS004",
    withdrawableAmount: 399.50,
    totalCommission: 1599.50,
    withdrawnAmount: 1200.00
  },
  {
    id: "15",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=5",
    distributor: "孙七",
    distributorId: "DIS005",
    withdrawableAmount: 499.75,
    totalCommission: 1999.75,
    withdrawnAmount: 1500.00
  }
]);

// 过滤
const filteredPendingItems = computed(() => {
  if (!pendingSearchQuery.value) return pendingItems.value;
  const q = pendingSearchQuery.value.toLowerCase();
  return pendingItems.value.filter(item =>
    item.distributor.toLowerCase().includes(q) ||
    item.distributorId.toLowerCase().includes(q)
  );
});

const filteredSettledItems = computed(() => {
  if (!settledSearchQuery.value) return settledItems.value;
  const q = settledSearchQuery.value.toLowerCase();
  return settledItems.value.filter(item =>
    item.distributor.toLowerCase().includes(q) ||
    item.distributorId.toLowerCase().includes(q)
  );
});

const filteredWithdrawableItems = computed(() => {
  if (!withdrawableSearchQuery.value) return withdrawableItems.value;
  const q = withdrawableSearchQuery.value.toLowerCase();
  return withdrawableItems.value.filter(item =>
    item.distributor.toLowerCase().includes(q) ||
    item.distributorId.toLowerCase().includes(q)
  );
});

// 分页数据
const pendingTableData = computed(() => {
  const start = (pendingCurrentPage.value - 1) * pendingPageSize.value;
  const end = start + pendingPageSize.value;
  return filteredPendingItems.value.slice(start, end);
});

const settledTableData = computed(() => {
  const start = (settledCurrentPage.value - 1) * settledPageSize.value;
  const end = start + settledPageSize.value;
  return filteredSettledItems.value.slice(start, end);
});

const withdrawableTableData = computed(() => {
  const start = (withdrawableCurrentPage.value - 1) * withdrawablePageSize.value;
  const end = start + withdrawablePageSize.value;
  return filteredWithdrawableItems.value.slice(start, end);
});

// 页数
const pendingPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredPendingItems.value.length / pendingPageSize.value))
);

const settledPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredSettledItems.value.length / settledPageSize.value))
);

const withdrawablePageCount = computed(() =>
  Math.max(1, Math.ceil(filteredWithdrawableItems.value.length / withdrawablePageSize.value))
);

// 表格列定义
const pendingColumns = [
  { accessorKey: "distributor", header: "分销员" },
  { accessorKey: "orderId", header: "订单编号" },
  { accessorKey: "orderAmount", header: "订单金额" },
  { accessorKey: "commissionRate", header: "佣金比例" },
  { accessorKey: "commissionAmount", header: "佣金金额" },
  { id: "actions", header: "操作" },
];

const settledColumns = [
  { accessorKey: "distributor", header: "分销员" },
  { accessorKey: "orderId", header: "订单编号" },
  { accessorKey: "orderAmount", header: "订单金额" },
  { accessorKey: "commissionRate", header: "佣金比例" },
  { accessorKey: "commissionAmount", header: "佣金金额" },
  { accessorKey: "settledAt", header: "结算时间" },
  { id: "actions", header: "操作" },
];

const withdrawableColumns = [
  { accessorKey: "distributor", header: "分销员" },
  { accessorKey: "withdrawableAmount", header: "可提现金额" },
  { accessorKey: "totalCommission", header: "累计佣金" },
  { accessorKey: "withdrawnAmount", header: "已提现金额" },
  { id: "actions", header: "操作" },
];

// 工具函数
const getRowData = (row: any) => {
  return row?.original ?? row;
};

const formatDate = (dateString: string) => {
  if (!dateString) return "—";
  const isoLike = dateString.replace(" ", "T");
  const d = new Date(isoLike);
  return isNaN(d.getTime()) ? "—" : d.toLocaleDateString("zh-CN");
};

// 重置筛选
const resetPendingFilters = () => {
  pendingSearchQuery.value = "";
  pendingCurrentPage.value = 1;
};

const resetSettledFilters = () => {
  settledSearchQuery.value = "";
  settledCurrentPage.value = 1;
};

const resetWithdrawableFilters = () => {
  withdrawableSearchQuery.value = "";
  withdrawableCurrentPage.value = 1;
};

// 模态框操作
const viewPendingCommission = (commission: any) => {
  currentCommission.value = commission;
  showCommissionModal.value = true;
};

const viewSettledCommission = (commission: any) => {
  currentCommission.value = commission;
  showCommissionModal.value = true;
};

const viewWithdrawableItem = (item: any) => {
  currentWithdrawable.value = {
    ...item,
    status: "withdrawable"
  };
  showWithdrawableModal.value = true;
};

const closeCommissionModal = () => {
  showCommissionModal.value = false;
  currentCommission.value = null;
};

const closeWithdrawableModal = () => {
  showWithdrawableModal.value = false;
  currentWithdrawable.value = null;
};

// 结算佣金
const settleCommission = (commission: any) => {
  // 在实际应用中，这里应该调用API进行结算
  alert(`已结算分销员 ${commission.distributor} 的佣金 ¥${commission.commissionAmount}`);
  
  // 更新统计数据
  pendingCount.value--;
  settledCount.value++;
  
  // 从待结算列表中移除
  const index = pendingItems.value.findIndex(item => item.id === commission.id);
  if (index !== -1) {
    pendingItems.value.splice(index, 1);
  }
  
  // 添加到已结算列表
  settledItems.value.push({
    ...commission,
    settledAt: new Date().toISOString().slice(0, 19).replace("T", " ")
  });
};

// 处理结算佣金
const handleSettleCommission = (commission: any) => {
  settleCommission(commission);
  closeCommissionModal();
};

// 提现申请
const requestWithdrawal = (item: any) => {
  withdrawalRequest.value = item;
  withdrawalAmount.value = item.withdrawableAmount;
  withdrawalRemark.value = "";
  showWithdrawalRequestModal.value = true;
};

const closeWithdrawalRequestModal = () => {
  showWithdrawalRequestModal.value = false;
  withdrawalRequest.value = null;
  withdrawalAmount.value = null;
  withdrawalRemark.value = "";
};

const submitWithdrawalRequest = () => {
  if (!withdrawalAmount.value || withdrawalAmount.value <= 0) {
    alert("请输入有效的提现金额");
    return;
  }
  
  if (withdrawalAmount.value > (withdrawalRequest.value?.withdrawableAmount || 0)) {
    alert("提现金额不能超过可提现金额");
    return;
  }
  
  // 在实际应用中，这里应该调用API提交提现申请
  alert(`已为分销员 ${withdrawalRequest.value?.distributor} 提交提现申请 ¥${withdrawalAmount.value}`);
  
  closeWithdrawalRequestModal();
};

// 导出数据
const exportData = () => {
  alert("导出数据功能待实现");
};
</script>