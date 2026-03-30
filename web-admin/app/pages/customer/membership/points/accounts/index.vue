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

      <UTable :columns="columns" :data="filteredAccounts" :loading="loading || adjusting">
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
              @click="adjustPoints(row)"
            >
              赠送
            </UButton>
            <UButton
              color="warning"
              variant="ghost"
              size="sm"
              icon="i-heroicons-minus-circle"
              @click="adjustPoints(row)"
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
            显示第 {{ totalAccounts === 0 ? 0 : (currentPage - 1) * pageSize + 1 }} 到 {{ Math.min(currentPage * pageSize, totalAccounts) }} 条，共 {{ totalAccounts }} 条
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
import { computed, onMounted, ref, watch } from "vue";
import { navigateTo, useToast } from "#imports";
import BatchAdjustPointsModal from "~/components/modals/BatchAdjustPointsModal.vue";
import { useCustomerApi } from "~/composables/api/useCustomer";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import type { MembershipInsight } from "~/types/customer";

type PointsAccount = {
  id: string;
  customerId: string;
  customerName: string;
  customerPhone: string;
  avatar: string;
  levelName: string;
  levelColor: string;
  balance: number;
};

const customerApi = useCustomerApi();
const membershipApi = useMembershipAdminApi();
const toast = useToast();

const showBatchAdjustModal = ref(false);
const showAdjustModal = ref(false);
const loading = ref(false);
const adjusting = ref(false);

const currentAccount = ref<PointsAccount | null>(null);
const searchQuery = ref("");

const currentPage = ref(1);
const pageSize = ref(10);

const todayEarned = ref(0);
const todaySpent = ref(0);

const accounts = ref<PointsAccount[]>([]);
const customers = ref<Array<{ id: string; name: string; phone: string; avatar: string; points: number; selected: boolean }>>([]);

const columns = [
  { accessorKey: "customer", header: "客户" },
  { accessorKey: "level", header: "会员等级" },
  { accessorKey: "balance", header: "积分余额" },
  { id: "actions", header: "操作" },
];

const filteredAccountsRaw = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  if (!keyword) return accounts.value;
  return accounts.value.filter(
    (account) =>
      account.customerName.toLowerCase().includes(keyword) ||
      account.customerPhone.includes(keyword),
  );
});

const filteredAccounts = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return filteredAccountsRaw.value.slice(start, end);
});

const totalAccounts = computed(() => filteredAccountsRaw.value.length);
const pageCount = computed(() => Math.max(1, Math.ceil(filteredAccountsRaw.value.length / pageSize.value)));
const totalBalance = computed(() =>
  accounts.value.reduce((sum, account) => sum + (Number(account.balance) || 0), 0),
);

const levelColors = ["#CD7F32", "#C0C0C0", "#FFD700", "#E5E4E2", "#B9F2FF", "#111827"];
const colorOfLevel = (name: string) => {
  if (!name) return "#9CA3AF";
  let hash = 0;
  for (let i = 0; i < name.length; i += 1) hash = (hash << 5) - hash + name.charCodeAt(i);
  return levelColors[Math.abs(hash) % levelColors.length];
};

const mapInsightToAccount = (insight: MembershipInsight): PointsAccount => {
  const customerId = insight.customer?.id || insight.snapshot?.customerId || "";
  const customerName = insight.customer?.name || customerId || "未知客户";
  const customerPhone = insight.customer?.phone || "-";
  const levelName =
    insight.snapshot?.tier || insight.customer?.membershipTierLabel || insight.customer?.membershipTier || "未分层";
  return {
    id: customerId,
    customerId,
    customerName,
    customerPhone,
    avatar: `https://api.dicebear.com/7.x/miniavs/svg?seed=${encodeURIComponent(customerId || customerName)}`,
    levelName,
    levelColor: colorOfLevel(levelName),
    balance: Number(insight.snapshot?.points ?? insight.customer?.points ?? 0) || 0,
  };
};

const loadPageData = async () => {
  try {
    loading.value = true;
    const membersResp = await customerApi.listMembers({ page: 1, pageSize: 1000 });

    const rows = (membersResp?.data || []) as MembershipInsight[];
    accounts.value = rows.map(mapInsightToAccount);

    customers.value = accounts.value.map((account) => ({
      id: account.customerId,
      name: account.customerName,
      phone: account.customerPhone,
      avatar: account.avatar,
      points: account.balance,
      selected: false,
    }));
  } catch (error: any) {
    toast.add({ title: "加载积分账户失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    loading.value = false;
  }
};

const openBatchAdjustModal = () => {
  showBatchAdjustModal.value = true;
};

const handleBatchAdjustSubmit = async (data: { type: "add" | "deduct"; points: number; reason: string; customerIds: string[] }) => {
  if (!data.customerIds?.length) return;
  try {
    adjusting.value = true;
    const delta = data.type === "add" ? Math.abs(data.points) : -Math.abs(data.points);
    let successCount = 0;

    for (const customerId of data.customerIds) {
      try {
        await membershipApi.adjustToken({
          customerId,
          tokenCode: "points",
          delta,
          reason: data.reason,
        });
        successCount += 1;
      } catch {
        // noop
      }
    }

    await loadPageData();
    toast.add({
      title: "批量调整完成",
      description: `成功 ${successCount}/${data.customerIds.length} 个客户`,
      color: successCount > 0 ? "success" : "error",
    });
  } catch (error: any) {
    toast.add({ title: "批量调整积分失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    adjusting.value = false;
  }
};

const handleBatchAdjustClose = () => {
  showBatchAdjustModal.value = false;
};

const adjustPoints = (account: PointsAccount) => {
  currentAccount.value = account;
  showAdjustModal.value = true;
};

const handleAdjustSubmit = async (data: { type: "add" | "deduct"; points: number; reason: string; accountId: string }) => {
  try {
    adjusting.value = true;
    const delta = data.type === "add" ? Math.abs(data.points) : -Math.abs(data.points);

    await membershipApi.adjustToken({
      customerId: data.accountId,
      tokenCode: "points",
      delta,
      reason: data.reason,
    });

    await loadPageData();
    toast.add({ title: `${data.type === "add" ? "赠送" : "扣减"}积分成功`, color: "success" });
  } catch (error: any) {
    toast.add({ title: "调整积分失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    adjusting.value = false;
  }
};

const handleAdjustClose = () => {
  showAdjustModal.value = false;
  currentAccount.value = null;
};

const viewAccountDetails = (account: PointsAccount) => {
  navigateTo(`/customer/membership/points/accounts/${account.id}`);
};

watch(searchQuery, () => {
  currentPage.value = 1;
});

watch(pageCount, (value) => {
  if (currentPage.value > value) {
    currentPage.value = value;
  }
});

onMounted(() => {
  loadPageData();
});
</script>
