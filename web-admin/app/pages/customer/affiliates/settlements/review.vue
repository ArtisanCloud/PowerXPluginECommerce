<template>
  <div class="p-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">提现申请审核</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">审核分销员的提现申请</p>
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
            <p class="text-sm text-gray-600 dark:text-gray-400">待审核</p>
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
            <p class="text-sm text-gray-600 dark:text-gray-400">已通过</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ approvedCount.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
            <UIcon name="i-heroicons-x-circle" class="w-6 h-6 text-red-600 dark:text-red-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">已拒绝</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ rejectedCount.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 标签页 -->
    <div class="mb-6">
      <UTabs v-model="activeTab" :items="tabs" />
    </div>

    <!-- 待审核 -->
    <UCard v-if="activeTab === 'pending'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">待审核</h2>
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

        <!-- 申请金额 -->
        <template #amount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 申请时间 -->
        <template #appliedAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ formatDate(getValue(row)) }}
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
              @click="viewPendingWithdrawal(getRowData(row))"
            >
              查看
            </UButton>
            <UButton
              color="success"
              variant="ghost"
              size="sm"
              icon="i-heroicons-check"
              @click="approveWithdrawal(getRowData(row))"
            >
              通过
            </UButton>
            <UButton
              color="red"
              variant="ghost"
              size="sm"
              icon="i-heroicons-x-mark"
              @click="rejectWithdrawal(getRowData(row))"
            >
              拒绝
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

    <!-- 已通过 -->
    <UCard v-if="activeTab === 'approved'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">已通过</h2>
          <div class="flex gap-3">
            <UInput v-model="approvedSearchQuery" placeholder="搜索分销员..." icon="i-heroicons-magnifying-glass" size="sm" />
            <UButton color="neutral" @click="resetApprovedFilters">重置</UButton>
          </div>
        </div>
      </template>

      <UTable :columns="approvedColumns" :data="approvedTableData">
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

        <!-- 申请金额 -->
        <template #amount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 申请时间 -->
        <template #appliedAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ formatDate(getValue(row)) }}
          </div>
        </template>

        <!-- 审核时间 -->
        <template #reviewedAt-cell="{ row, getValue }">
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
            @click="viewApprovedWithdrawal(getRowData(row))"
          >
            查看
          </UButton>
        </template>
      </UTable>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (approvedCurrentPage - 1) * approvedPageSize + 1 }}
            到 {{ Math.min(approvedCurrentPage * approvedPageSize, filteredApprovedItems.length) }} 条，共 {{ filteredApprovedItems.length }} 条
          </div>
          <div class="flex items-center gap-4">
            <UPagination
              v-model="approvedCurrentPage"
              :page-count="approvedPageCount"
              :total="filteredApprovedItems.length"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500 dark:text-gray-400">每页</span>
              <USelect
                v-model="approvedPageSize"
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

    <!-- 已拒绝 -->
    <UCard v-if="activeTab === 'rejected'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">已拒绝</h2>
          <div class="flex gap-3">
            <UInput v-model="rejectedSearchQuery" placeholder="搜索分销员..." icon="i-heroicons-magnifying-glass" size="sm" />
            <UButton color="neutral" @click="resetRejectedFilters">重置</UButton>
          </div>
        </div>
      </template>

      <UTable :columns="rejectedColumns" :data="rejectedTableData">
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

        <!-- 申请金额 -->
        <template #amount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 申请时间 -->
        <template #appliedAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ formatDate(getValue(row)) }}
          </div>
        </template>

        <!-- 审核时间 -->
        <template #reviewedAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ formatDate(getValue(row)) }}
          </div>
        </template>

        <!-- 拒绝原因 -->
        <template #reason-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white max-w-xs truncate" :title="getValue(row)">
            {{ getValue(row) || '—' }}
          </div>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <UButton
            color="neutral"
            variant="ghost"
            size="sm"
            icon="i-heroicons-eye"
            @click="viewRejectedWithdrawal(getRowData(row))"
          >
            查看
          </UButton>
        </template>
      </UTable>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (rejectedCurrentPage - 1) * rejectedPageSize + 1 }}
            到 {{ Math.min(rejectedCurrentPage * rejectedPageSize, filteredRejectedItems.length) }} 条，共 {{ filteredRejectedItems.length }} 条
          </div>
          <div class="flex items-center gap-4">
            <UPagination
              v-model="rejectedCurrentPage"
              :page-count="rejectedPageCount"
              :total="filteredRejectedItems.length"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500 dark:text-gray-400">每页</span>
              <USelect
                v-model="rejectedPageSize"
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

    <!-- 提现申请详情模态框 -->
    <ViewWithdrawalModal
      v-if="showWithdrawalModal && currentWithdrawal"
      v-model:open="showWithdrawalModal"
      :withdrawal="currentWithdrawal"
      :is-review-mode="isReviewMode"
      @close="closeWithdrawalModal"
      @approve="handleApproveWithdrawal"
      @reject="handleRejectWithdrawal"
    />

    <!-- 拒绝原因模态框 -->
    <UModal
      v-model:open="showRejectModal"
      title="拒绝提现申请"
      description="请填写拒绝分销员提现申请的原因"
      :close="{ onClick: () => closeRejectModal() }"
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
                <span class="text-gray-900 dark:text-white font-medium">{{ rejectWithdrawalItem?.distributor || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">申请金额：</span>
                <span class="text-gray-900 dark:text-white font-medium">¥{{ (rejectWithdrawalItem?.amount || 0).toLocaleString() }}</span>
              </div>
              <div>
                <UTextarea
                  v-model="rejectReason"
                  label="拒绝原因"
                  placeholder="请输入拒绝原因"
                  :rows="3"
                />
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeRejectModal">取消</UButton>
          <UButton color="red" @click="confirmReject">确认拒绝</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import ViewWithdrawalModal from "~/components/modals/ViewWithdrawalModal.vue";

const avatarPlaceholder = "https://api.dicebear.com/7.x/miniavs/svg?seed=placeholder";

// 标签页
const activeTab = ref("pending");

// 搜索查询
const pendingSearchQuery = ref("");
const approvedSearchQuery = ref("");
const rejectedSearchQuery = ref("");

// 分页
const pendingCurrentPage = ref(1);
const pendingPageSize = ref(10);

const approvedCurrentPage = ref(1);
const approvedPageSize = ref(10);

const rejectedCurrentPage = ref(1);
const rejectedPageSize = ref(10);

// 模态框
const showWithdrawalModal = ref(false);
const showRejectModal = ref(false);
const currentWithdrawal = ref<any>(null);
const rejectWithdrawalItem = ref<any>(null);

// 拒绝原因
const rejectReason = ref("");

// Tabs
const tabs = [
  { label: "待审核", value: "pending" },
  { label: "已通过", value: "approved" },
  { label: "已拒绝", value: "rejected" },
];

// 统计数据
const pendingCount = ref(5);
const approvedCount = ref(12);
const rejectedCount = ref(3);

// 是否为审核模式
const isReviewMode = ref(true);

// 模拟数据
const pendingItems = ref([
  {
    id: "1",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1",
    distributor: "张三",
    distributorId: "DIS001",
    amount: 299.50,
    appliedAt: "2023-09-15 14:30:25",
    bankAccount: "6222021234567890123",
    accountName: "张三",
    totalCommission: 1299.50,
    withdrawnAmount: 1000.00,
    withdrawableAmount: 299.50
  },
  {
    id: "2",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2",
    distributor: "李四",
    distributorId: "DIS002",
    amount: 599.25,
    appliedAt: "2023-09-16 09:15:42",
    bankAccount: "6222021234567890124",
    accountName: "李四",
    totalCommission: 2199.25,
    withdrawnAmount: 1600.00,
    withdrawableAmount: 599.25
  },
  {
    id: "3",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3",
    distributor: "王五",
    distributorId: "DIS003",
    amount: 199.00,
    appliedAt: "2023-09-17 11:20:33",
    bankAccount: "6222021234567890125",
    accountName: "王五",
    totalCommission: 899.00,
    withdrawnAmount: 700.00,
    withdrawableAmount: 199.00
  }
]);

const approvedItems = ref([
  {
    id: "4",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4",
    distributor: "赵六",
    distributorId: "DIS004",
    amount: 399.50,
    appliedAt: "2023-09-10 16:45:12",
    reviewedAt: "2023-09-11 09:30:45",
    bankAccount: "6222021234567890126",
    accountName: "赵六",
    totalCommission: 1599.50,
    withdrawnAmount: 1200.00,
    withdrawableAmount: 399.50
  },
  {
    id: "5",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=5",
    distributor: "孙七",
    distributorId: "DIS005",
    amount: 499.75,
    appliedAt: "2023-09-12 14:22:33",
    reviewedAt: "2023-09-13 10:15:22",
    bankAccount: "6222021234567890127",
    accountName: "孙七",
    totalCommission: 1999.75,
    withdrawnAmount: 1500.00,
    withdrawableAmount: 499.75
  }
]);

const rejectedItems = ref([
  {
    id: "6",
    avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=6",
    distributor: "周八",
    distributorId: "DIS006",
    amount: 299.00,
    appliedAt: "2023-09-05 10:30:15",
    reviewedAt: "2023-09-06 14:20:33",
    reason: "账户信息不完整",
    bankAccount: "6222021234567890128",
    accountName: "周八",
    totalCommission: 1299.00,
    withdrawnAmount: 1000.00,
    withdrawableAmount: 299.00
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

const filteredApprovedItems = computed(() => {
  if (!approvedSearchQuery.value) return approvedItems.value;
  const q = approvedSearchQuery.value.toLowerCase();
  return approvedItems.value.filter(item =>
    item.distributor.toLowerCase().includes(q) ||
    item.distributorId.toLowerCase().includes(q)
  );
});

const filteredRejectedItems = computed(() => {
  if (!rejectedSearchQuery.value) return rejectedItems.value;
  const q = rejectedSearchQuery.value.toLowerCase();
  return rejectedItems.value.filter(item =>
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

const approvedTableData = computed(() => {
  const start = (approvedCurrentPage.value - 1) * approvedPageSize.value;
  const end = start + approvedPageSize.value;
  return filteredApprovedItems.value.slice(start, end);
});

const rejectedTableData = computed(() => {
  const start = (rejectedCurrentPage.value - 1) * rejectedPageSize.value;
  const end = start + rejectedPageSize.value;
  return filteredRejectedItems.value.slice(start, end);
});

// 页数
const pendingPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredPendingItems.value.length / pendingPageSize.value))
);

const approvedPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredApprovedItems.value.length / approvedPageSize.value))
);

const rejectedPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredRejectedItems.value.length / rejectedPageSize.value))
);

// 表格列定义
const pendingColumns = [
  { accessorKey: "distributor", header: "分销员" },
  { accessorKey: "amount", header: "申请金额" },
  { accessorKey: "appliedAt", header: "申请时间" },
  { id: "actions", header: "操作" },
];

const approvedColumns = [
  { accessorKey: "distributor", header: "分销员" },
  { accessorKey: "amount", header: "申请金额" },
  { accessorKey: "appliedAt", header: "申请时间" },
  { accessorKey: "reviewedAt", header: "审核时间" },
  { id: "actions", header: "操作" },
];

const rejectedColumns = [
  { accessorKey: "distributor", header: "分销员" },
  { accessorKey: "amount", header: "申请金额" },
  { accessorKey: "appliedAt", header: "申请时间" },
  { accessorKey: "reviewedAt", header: "审核时间" },
  { accessorKey: "reason", header: "拒绝原因" },
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

const resetApprovedFilters = () => {
  approvedSearchQuery.value = "";
  approvedCurrentPage.value = 1;
};

const resetRejectedFilters = () => {
  rejectedSearchQuery.value = "";
  rejectedCurrentPage.value = 1;
};

// 模态框操作
const viewPendingWithdrawal = (withdrawal: any) => {
  currentWithdrawal.value = {
    ...withdrawal,
    status: "pending"
  };
  showWithdrawalModal.value = true;
};

const viewApprovedWithdrawal = (withdrawal: any) => {
  currentWithdrawal.value = {
    ...withdrawal,
    status: "approved"
  };
  showWithdrawalModal.value = true;
};

const viewRejectedWithdrawal = (withdrawal: any) => {
  currentWithdrawal.value = {
    ...withdrawal,
    status: "rejected"
  };
  showWithdrawalModal.value = true;
};

const closeWithdrawalModal = () => {
  showWithdrawalModal.value = false;
  currentWithdrawal.value = null;
};

// 通过提现申请
const approveWithdrawal = (withdrawal: any) => {
  // 在实际应用中，这里应该调用API通过提现申请
  alert(`已通过分销员 ${withdrawal.distributor} 的提现申请 ¥${withdrawal.amount}`);
  
  // 更新统计数据
  pendingCount.value--;
  approvedCount.value++;
  
  // 从待审核列表中移除
  const index = pendingItems.value.findIndex(item => item.id === withdrawal.id);
  if (index !== -1) {
    pendingItems.value.splice(index, 1);
  }
  
  // 添加到已通过列表
  approvedItems.value.push({
    ...withdrawal,
    reviewedAt: new Date().toISOString().slice(0, 19).replace("T", " ")
  });
};

// 处理通过提现申请
const handleApproveWithdrawal = (withdrawal: any) => {
  approveWithdrawal(withdrawal);
  closeWithdrawalModal();
};

// 拒绝提现申请
const rejectWithdrawal = (withdrawal: any) => {
  rejectWithdrawalItem.value = withdrawal;
  rejectReason.value = "";
  showRejectModal.value = true;
};

const closeRejectModal = () => {
  showRejectModal.value = false;
  rejectWithdrawalItem.value = null;
  rejectReason.value = "";
};

const confirmReject = () => {
  if (!rejectReason.value.trim()) {
    alert("请输入拒绝原因");
    return;
  }
  
  if (!rejectWithdrawalItem.value) return;
  
  // 在实际应用中，这里应该调用API拒绝提现申请
  alert(`已拒绝分销员 ${rejectWithdrawalItem.value.distributor} 的提现申请`);
  
  // 更新统计数据
  pendingCount.value--;
  rejectedCount.value++;
  
  // 从待审核列表中移除
  const index = pendingItems.value.findIndex(item => item.id === rejectWithdrawalItem.value.id);
  if (index !== -1) {
    pendingItems.value.splice(index, 1);
  }
  
  // 添加到已拒绝列表
  rejectedItems.value.push({
    ...rejectWithdrawalItem.value,
    reviewedAt: new Date().toISOString().slice(0, 19).replace("T", " "),
    reason: rejectReason.value
  });
  
  closeRejectModal();
};

// 处理拒绝提现申请
const handleRejectWithdrawal = (withdrawal: any) => {
  rejectWithdrawal(withdrawal);
  closeWithdrawalModal();
};

// 导出数据
const exportData = () => {
  alert("导出数据功能待实现");
};
</script>
