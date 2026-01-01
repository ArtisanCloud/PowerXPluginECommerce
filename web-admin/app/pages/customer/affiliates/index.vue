<template>
  <div>
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">推荐/分销</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理推荐关系和奖励发放记录</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-plus" @click="exportData">导出数据</UButton>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-user-group" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">总邀请人数</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ totalInvitations.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-gift" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">总奖励金额</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">¥{{ totalRewards.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon name="i-heroicons-arrow-trending-up" class="w-6 h-6 text-yellow-600 dark:text-yellow-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">本月新增</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ monthlyNew.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon name="i-heroicons-banknotes" class="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">待发放奖励</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">¥{{ pendingRewards.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 标签页 -->
    <div class="mb-6">
      <UTabs v-model="activeTab" :items="tabs" />
    </div>

    <!-- 邀请关系链 -->
    <UCard v-if="activeTab === 'invitations'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">邀请关系链</h2>
          <UInput v-model="invitationSearchQuery" placeholder="搜索邀请人或被邀请人..." icon="i-heroicons-magnifying-glass" size="sm" />
        </div>
      </template>

      <UTable :columns="invitationColumns" :data="invitationTableData">
        <!-- 邀请人 -->
        <template #inviter-cell="{ row, column, getValue }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="getValue(row)?.avatar || avatarPlaceholder"
              :alt="getValue(row)?.name || '邀请人'"
              size="md"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ getValue(row)?.name || '—' }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ getValue(row)?.phone || '—' }}
              </div>
            </div>
          </div>
        </template>

        <!-- 被邀请人 -->
        <template #invitee-cell="{ row, column, getValue }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="getValue(row)?.avatar || avatarPlaceholder"
              :alt="getValue(row)?.name || '被邀请人'"
              size="md"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ getValue(row)?.name || '—' }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ getValue(row)?.phone || '—' }}
              </div>
            </div>
          </div>
        </template>

        <!-- 邀请时间 -->
        <template #invitedAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ formatTime(getValue(row)) }}
          </div>
        </template>

        <!-- 状态 -->
        <template #status-cell="{ row, getValue }">
          <UBadge :color="getValue(row) === 'active' ? 'success' : 'neutral'" variant="soft">
            {{ getValue(row) === 'active' ? '有效' : '无效' }}
          </UBadge>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <UButton
            color="neutral"
            variant="ghost"
            size="sm"
            icon="i-heroicons-eye"
            @click="openInvitation(row)"
          >
            查看
          </UButton>
        </template>
      </UTable>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (invitationCurrentPage - 1) * invitationPageSize + 1 }}
            到 {{ Math.min(invitationCurrentPage * invitationPageSize, totalInvitations) }} 条，共 {{ totalInvitations }} 条
          </div>
          <UPagination v-model="invitationCurrentPage" :page-count="invitationPageCount" :total="totalInvitations" :ui="{ rounded: 'rounded-full' }" />
        </div>
      </template>
    </UCard>

    <!-- 奖励发放记录 -->
    <UCard v-if="activeTab === 'rewards'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">奖励发放记录</h2>
          <UInput v-model="rewardSearchQuery" placeholder="搜索奖励记录..." icon="i-heroicons-magnifying-glass" size="sm" />
        </div>
      </template>

      <UTable :columns="rewardColumns" :data="rewardTableData">
        <!-- 客户 -->
        <template #customer-cell="{ row, getValue }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="getValue(row)?.avatar || avatarPlaceholder"
              :alt="getValue(row)?.name || '客户'"
              size="md"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ getValue(row)?.name || '—' }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ getValue(row)?.phone || '—' }}
              </div>
            </div>
          </div>
        </template>

        <!-- 奖励金额 -->
        <template #amount-cell="{ row, getValue }">
          <div class="font-medium text-gray-900 dark:text-white">
            ¥{{ Number(getValue(row) ?? 0).toLocaleString() }}
          </div>
        </template>

        <!-- 奖励类型 -->
        <template #type-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">{{ getValue(row) || '—' }}</div>
        </template>

        <!-- 关联记录 -->
        <template #relatedRecord-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">{{ getValue(row) || '—' }}</div>
        </template>

        <!-- 创建时间 -->
        <template #createdAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">{{ formatTime(getValue(row)) }}</div>
        </template>

        <!-- 发放时间 -->
        <template #paidAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">{{ getValue(row) ? formatTime(getValue(row)) : '—' }}</div>
        </template>

        <!-- 状态 -->
        <template #status-cell="{ row, getValue }">
          <UBadge
            :color="getValue(row) === 'paid' ? 'success' : getValue(row) === 'pending' ? 'warning' : 'error'"
            variant="soft"
          >
            {{ getValue(row) === 'paid' ? '已发放' : getValue(row) === 'pending' ? '待发放' : '已取消' }}
          </UBadge>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <UButton
            color="neutral"
            variant="ghost"
            size="sm"
            icon="i-heroicons-eye"
            @click="openReward(row)"
          >
            查看
          </UButton>
        </template>
      </UTable>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (rewardCurrentPage - 1) * rewardPageSize + 1 }}
            到 {{ Math.min(rewardCurrentPage * rewardPageSize, totalRewardsRecords) }} 条，共 {{ totalRewardsRecords }} 条
          </div>
          <UPagination v-model="rewardCurrentPage" :page-count="rewardPageCount" :total="totalRewardsRecords" :ui="{ rounded: 'rounded-full' }" />
        </div>
      </template>
    </UCard>

    <!-- 弹窗仅在需要时挂载；传参永远不是 null -->
    <ViewInvitationModal
      v-if="showInvitationModal && currentInvitation"
      v-model:open="showInvitationModal"
      :invitation="currentInvitation"
      @close="closeInvitation"
    />
    <ViewRewardModal
      v-if="showRewardModal && currentReward"
      v-model:open="showRewardModal"
      :reward="currentReward"
      @close="closeReward"
    />
  </div>
</template>

<script setup lang="ts">
import ViewInvitationModal from "~/components/modals/ViewInvitationModal.vue";
import ViewRewardModal from "~/components/modals/ViewRewardModal.vue";

const avatarPlaceholder = "https://api.dicebear.com/7.x/miniavs/svg?seed=placeholder";

const activeTab = ref("invitations");
const invitationSearchQuery = ref("");
const rewardSearchQuery = ref("");

// 分页（示例静态）
const invitationCurrentPage = ref(1);
const invitationPageSize = ref(10);
const totalInvitations = ref(1245);
const invitationPageCount = computed(() => Math.ceil(totalInvitations.value / invitationPageSize.value));

const rewardCurrentPage = ref(1);
const rewardPageSize = ref(10);
const totalRewardsRecords = ref(856);
const rewardPageCount = computed(() => Math.ceil(totalRewardsRecords.value / rewardPageSize.value));

// 统计
const totalRewards = ref(245678);
const monthlyNew = ref(124);
const pendingRewards = ref(15670);

// 弹窗与当前行：默认 undefined（不是 null）
const showInvitationModal = ref(false);
const showRewardModal = ref(false);
const currentInvitation = ref<any>(); // undefined
const currentReward = ref<any>();     // undefined

// Tabs
const tabs = [
  { label: "邀请关系链", value: "invitations" },
  { label: "奖励发放记录", value: "rewards" },
];

// 示例数据
const invitations = ref([
  { id: "inv_1", inviter: { id: "cust_1", name: "张三", phone: "13800138001", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1" }, invitee: { id: "cust_2", name: "李四", phone: "13800138002", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2" }, invitedAt: "2023-09-15 14:30:25", status: "active" },
  { id: "inv_2", inviter: { id: "cust_1", name: "张三", phone: "13800138001", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1" }, invitee: { id: "cust_3", name: "王五", phone: "13800138003", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3" }, invitedAt: "2023-09-12 10:15:42", status: "active" },
  { id: "inv_3", inviter: { id: "cust_2", name: "李四", phone: "13800138002", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2" }, invitee: { id: "cust_4", name: "赵六", phone: "13800138004", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4" }, invitedAt: "2023-09-10 08:45:17", status: "active" },
  { id: "inv_4", inviter: { id: "cust_3", name: "王五", phone: "13800138003", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3" }, invitee: { id: "cust_5", name: "孙七", phone: "13800138005", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=5" }, invitedAt: "2023-09-08 16:22:33", status: "active" },
  { id: "inv_5", inviter: { id: "cust_1", name: "张三", phone: "13800138001", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1" }, invitee: { id: "cust_6", name: "周八", phone: "13800138006", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=6" }, invitedAt: "2023-09-05 11:38:56", status: "inactive" },
]);

const rewards = ref([
  { id: "rew_1", customer: { id: "cust_1", name: "张三", phone: "13800138001", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1" }, amount: 1200, type: "邀请奖励", relatedRecord: "邀请李四注册", createdAt: "2023-09-15 14:30:25", paidAt: "2023-09-16 09:15:30", status: "paid" },
  { id: "rew_2", customer: { id: "cust_2", name: "李四", phone: "13800138002", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2" }, amount: 800, type: "消费返利", relatedRecord: "订单号: ORD20230915001", createdAt: "2023-09-15 16:45:12", paidAt: "2023-09-16 09:15:30", status: "paid" },
  { id: "rew_3", customer: { id: "cust_1", name: "张三", phone: "13800138001", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1" }, amount: 650, type: "邀请奖励", relatedRecord: "邀请王五注册", createdAt: "2023-09-12 10:15:42", paidAt: "2023-09-13 09:22:15", status: "paid" },
  { id: "rew_4", customer: { id: "cust_3", name: "王五", phone: "13800138003", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3" }, amount: 2000, type: "消费返利", relatedRecord: "订单号: ORD20230910002", createdAt: "2023-09-10 18:33:27", paidAt: "2023-09-11 09:10:45", status: "paid" },
  { id: "rew_5", customer: { id: "cust_2", name: "李四", phone: "13800138002", avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2" }, amount: 1200, type: "邀请奖励", relatedRecord: "邀请赵六注册", createdAt: "2023-09-08 16:22:33", paidAt: null, status: "pending" },
]);

// 过滤
const filteredInvitations = computed(() => {
  if (!invitationSearchQuery.value) return invitations.value;
  const q = invitationSearchQuery.value.toLowerCase();
  return invitations.value.filter((inv) =>
    inv.inviter.name.toLowerCase().includes(q) ||
    inv.inviter.phone.includes(invitationSearchQuery.value) ||
    inv.invitee.name.toLowerCase().includes(q) ||
    inv.invitee.phone.includes(invitationSearchQuery.value)
  );
});
const filteredRewards = computed(() => {
  if (!rewardSearchQuery.value) return rewards.value;
  const q = rewardSearchQuery.value.toLowerCase();
  return rewards.value.filter((r) =>
    r.customer.name.toLowerCase().includes(q) ||
    r.customer.phone.includes(rewardSearchQuery.value) ||
    r.type.toLowerCase().includes(q) ||
    r.relatedRecord.toLowerCase().includes(q)
  );
});

// 表格数据
const invitationTableData = computed(() => filteredInvitations.value);
const rewardTableData = computed(() => filteredRewards.value);

// 列
const invitationColumns = [
  { accessorKey: "inviter", header: "邀请人" },
  { accessorKey: "invitee", header: "被邀请人" },
  { accessorKey: "invitedAt", header: "邀请时间" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
];
const rewardColumns = [
  { accessorKey: "customer", header: "客户" },
  { accessorKey: "amount", header: "奖励金额" },
  { accessorKey: "type", header: "奖励类型" },
  { accessorKey: "relatedRecord", header: "关联记录" },
  { accessorKey: "createdAt", header: "创建时间" },
  { accessorKey: "paidAt", header: "发放时间" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
];

// 打开弹窗（兼容 row.original）
const openInvitation = (row: any) => {
  currentInvitation.value = row?.original ?? row;
  showInvitationModal.value = true;
};
const openReward = (row: any) => {
  currentReward.value = row?.original ?? row;
  showRewardModal.value = true;
};

// 关闭弹窗
const closeInvitation = () => {
  showInvitationModal.value = false;
  currentInvitation.value = undefined;
};
const closeReward = () => {
  showRewardModal.value = false;
  currentReward.value = undefined;
};

// 工具
const exportData = () => alert("导出数据功能待实现");
const formatTime = (v?: string | null) => {
  if (!v) return "—";
  const d = new Date(String(v).replace(" ", "T"));
  if (isNaN(d as any)) return String(v);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
};
</script>
