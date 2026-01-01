<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          等级管理
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理会员等级、门槛规则和权益配置
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" @click="openCreateModal">
        创建等级
      </UButton>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon
              name="i-heroicons-star"
              class="w-6 h-6 text-blue-600 dark:text-blue-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              总等级数
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ tiers.length }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon
              name="i-heroicons-check-circle"
              class="w-6 h-6 text-green-600 dark:text-green-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              启用等级
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ tiers.filter((t) => t.status === "active").length }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon
              name="i-heroicons-pause-circle"
              class="w-6 h-6 text-yellow-600 dark:text-yellow-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              停用等级
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ tiers.filter((t) => t.status === "inactive").length }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon
              name="i-heroicons-users"
              class="w-6 h-6 text-purple-600 dark:text-purple-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              会员总数
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ tiers.reduce((sum, t) => sum + (t.memberCount || 0), 0) }}
            </p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 等级列表 -->
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            等级列表
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

        <template #threshold-cell="{ row }">
          <div class="text-sm">
            <div class="font-medium text-gray-900 dark:text-white">
              消费满 ¥{{ row.spendThreshold?.toLocaleString() || 0 }}
            </div>
            <div class="text-gray-500 dark:text-gray-400">
              或积分达 {{ row.pointsThreshold?.toLocaleString() || 0 }}
            </div>
          </div>
        </template>

        <template #benefits-cell="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge
              v-for="benefit in (row.benefits || []).slice(0, 2)"
              :key="benefit.type"
              variant="soft"
              size="sm"
            >
              {{ getBenefitLabel(benefit.type) }}
            </UBadge>
            <UBadge
              v-if="(row.benefits || []).length > 2"
              variant="soft"
              size="sm"
            >
              +{{ (row.benefits || []).length - 2 }} 项权益
            </UBadge>
          </div>
        </template>

        <template #memberCount-cell="{ row }">
          <div class="text-center">
            <div class="font-medium text-gray-900 dark:text-white">
              {{ row.memberCount?.toLocaleString() || 0 }}
            </div>
            <div class="text-xs text-gray-500 dark:text-gray-400">人</div>
          </div>
        </template>

        <template #status-cell="{ row }">
          <UBadge
            :color="row.status === 'active' ? 'success' : 'neutral'"
            variant="soft"
          >
            {{ row.status === "active" ? "启用" : "停用" }}
          </UBadge>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="openViewModal(row)"
            >
              查看
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
              @click="openEditModal(row)"
            >
              编辑
            </UButton>
            <UButton
              :color="row.status === 'active' ? 'warning' : 'success'"
              variant="ghost"
              size="sm"
              :icon="
                row.status === 'active'
                  ? 'i-heroicons-pause'
                  : 'i-heroicons-play'
              "
              @click="toggleTierStatus(row)"
            >
              {{ row.status === "active" ? "停用" : "启用" }}
            </UButton>
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              @click="deleteTier(row)"
            >
              删除
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 查看等级模态框 -->
    <ViewTierModal
      v-if="viewingTier"
      v-model:open="showViewModal"
      :tier="viewingTier"
      @edit="handleViewToEdit"
    />

    <!-- 创建等级模态框 -->
    <CreateTierModal
      v-model:open="showCreateModal"
      @created="handleTierCreated"
    />

    <!-- 编辑等级模态框 -->
    <EditTierModal
      v-if="editingTier"
      v-model:open="showEditModal"
      :tier="editingTier"
      @updated="handleTierUpdated"
      @deleted="handleTierDeleted"
    />
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import ViewTierModal from "~/components/modals/ViewTierModal.vue";
import CreateTierModal from "~/components/modals/CreateTierModal.vue";
import EditTierModal from "~/components/modals/EditTierModal.vue";

// 等级数据类型
type MembershipTier = {
  id: string;
  name: string;
  description: string;
  color: string;
  level: number;
  spendThreshold?: number;
  pointsThreshold?: number;
  benefits: Array<{
    type: string;
    value: number | string;
    description: string;
  }>;
  rules: Array<{
    type: string;
    condition: string;
    value: number | string;
    description: string;
  }>;
  memberCount?: number;
  status: "active" | "inactive";
  createdAt: string;
  updatedAt: string;
};

// 模态框状态
const showViewModal = ref(false);
const showCreateModal = ref(false);
const showEditModal = ref(false);
const viewingTier = ref<MembershipTier | null>(null);
const editingTier = ref<MembershipTier | null>(null);

// 搜索
const searchQuery = ref("");

// 列定义
const columns = computed<TableColumn<MembershipTier>[]>(() => [
  { accessorKey: "name", header: "等级名称" },
  { accessorKey: "level", header: "等级" },
  { accessorKey: "threshold", header: "升级门槛" },
  { accessorKey: "benefits", header: "权益" },
  { accessorKey: "memberCount", header: "会员数" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

// 等级数据
const tiers = ref<MembershipTier[]>([
  {
    id: "tier_1",
    name: "青铜会员",
    description: "新用户默认等级",
    color: "#CD7F32",
    level: 1,
    spendThreshold: 0,
    pointsThreshold: 0,
    benefits: [
      { type: "discount", value: 5, description: "全场商品95折" },
      { type: "points", value: 1, description: "消费1元得1积分" },
    ],
    rules: [
      { type: "spend_limit", condition: "<", value: 1000, description: "消费金额小于1000元" },
    ],
    memberCount: 1250,
    status: "active",
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
  {
    id: "tier_2",
    name: "白银会员",
    description: "消费达标的忠实用户",
    color: "#C0C0C0",
    level: 2,
    spendThreshold: 1000,
    pointsThreshold: 500,
    benefits: [
      { type: "discount", value: 8, description: "全场商品92折" },
      { type: "points", value: 1.2, description: "消费1元得1.2积分" },
      { type: "shipping", value: 0, description: "免运费" },
    ],
    rules: [
      { type: "spend_limit", condition: ">=", value: 1000, description: "消费金额大于等于1000元" },
      { type: "points_limit", condition: ">=", value: 500, description: "积分大于等于500分" },
    ],
    memberCount: 850,
    status: "active",
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
  {
    id: "tier_3",
    name: "黄金会员",
    description: "高价值客户",
    color: "#FFD700",
    level: 3,
    spendThreshold: 5000,
    pointsThreshold: 2000,
    benefits: [
      { type: "discount", value: 12, description: "全场商品88折" },
      { type: "points", value: 1.5, description: "消费1元得1.5积分" },
      { type: "shipping", value: 0, description: "免运费" },
      { type: "priority", value: 1, description: "优先客服" },
    ],
    rules: [
      { type: "spend_limit", condition: ">=", value: 5000, description: "消费金额大于等于5000元" },
      { type: "points_limit", condition: ">=", value: 2000, description: "积分大于等于2000分" },
      { type: "purchase_count", condition: ">=", value: 10, description: "购买次数大于等于10次" },
    ],
    memberCount: 420,
    status: "active",
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
  {
    id: "tier_4",
    name: "铂金会员",
    description: "VIP客户",
    color: "#E5E4E2",
    level: 4,
    spendThreshold: 15000,
    pointsThreshold: 5000,
    benefits: [
      { type: "discount", value: 15, description: "全场商品85折" },
      { type: "points", value: 2, description: "消费1元得2积分" },
      { type: "shipping", value: 0, description: "免运费" },
      { type: "priority", value: 1, description: "优先客服" },
      { type: "exclusive", value: 1, description: "专属活动" },
    ],
    rules: [
      { type: "spend_limit", condition: ">=", value: 15000, description: "消费金额大于等于15000元" },
      { type: "points_limit", condition: ">=", value: 5000, description: "积分大于等于5000分" },
      { type: "purchase_count", condition: ">=", value: 20, description: "购买次数大于等于20次" },
    ],
    memberCount: 180,
    status: "active",
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
  {
    id: "tier_5",
    name: "钻石会员",
    description: "顶级VIP客户",
    color: "#B9F2FF",
    level: 5,
    spendThreshold: 50000,
    pointsThreshold: 15000,
    benefits: [
      { type: "discount", value: 20, description: "全场商品8折" },
      { type: "points", value: 3, description: "消费1元得3积分" },
      { type: "shipping", value: 0, description: "免运费" },
      { type: "priority", value: 1, description: "优先客服" },
      { type: "exclusive", value: 1, description: "专属活动" },
      { type: "personal", value: 1, description: "专属客服" },
    ],
    rules: [
      { type: "spend_limit", condition: ">=", value: 50000, description: "消费金额大于等于50000元" },
      { type: "points_limit", condition: ">=", value: 15000, description: "积分大于等于15000分" },
      { type: "purchase_count", condition: ">=", value: 50, description: "购买次数大于等于50次" },
    ],
    memberCount: 45,
    status: "active",
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
]);

// 过滤后的等级列表
const filteredTiers = computed(() => {
  if (!searchQuery.value) return tiers.value;
  return tiers.value.filter(
    (tier) =>
      tier.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      tier.description.toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});

// 权益类型标签
const getBenefitLabel = (type: string) => {
  const labels = {
    discount: "折扣",
    points: "积分",
    shipping: "包邮",
    priority: "优先",
    exclusive: "专属",
    personal: "专服",
  };
  return labels[type as keyof typeof labels] || type;
};

// 打开查看模态框
const openViewModal = (tier: MembershipTier) => {
  viewingTier.value = tier;
  showViewModal.value = true;
};

// 打开创建模态框
const openCreateModal = () => {
  showCreateModal.value = true;
};

// 打开编辑模态框
const openEditModal = (tier: MembershipTier) => {
  editingTier.value = tier;
  showEditModal.value = true;
};

// 从查看切换到编辑
const handleViewToEdit = (tier: MembershipTier) => {
  showViewModal.value = false;
  viewingTier.value = null;
  editingTier.value = tier;
  showEditModal.value = true;
};

// 处理等级创建
const handleTierCreated = (newTier: MembershipTier) => {
  tiers.value.push(newTier);
  showCreateModal.value = false;
};

// 处理等级更新
const handleTierUpdated = (updatedTier: MembershipTier) => {
  const index = tiers.value.findIndex((t) => t.id === updatedTier.id);
  if (index !== -1) {
    tiers.value[index] = updatedTier;
  }
  showEditModal.value = false;
  editingTier.value = null;
};

// 处理等级删除
const handleTierDeleted = (tierId: string) => {
  const index = tiers.value.findIndex((t) => t.id === tierId);
  if (index !== -1) {
    tiers.value.splice(index, 1);
  }
  showEditModal.value = false;
  editingTier.value = null;
};

// 切换等级状态
const toggleTierStatus = async (tier: MembershipTier) => {
  const newStatus = tier.status === "active" ? "inactive" : "active";
  const action = newStatus === "active" ? "启用" : "停用";

  if (!confirm(`确定要${action}等级 "${tier.name}" 吗？`)) {
    return;
  }

  // 模拟API调用
  const index = tiers.value.findIndex((t) => t.id === tier.id);
  if (index !== -1) {
    tiers.value[index].status = newStatus;
    tiers.value[index].updatedAt = new Date().toISOString();
  }
};

// 删除等级
const deleteTier = async (tier: MembershipTier) => {
  if (!confirm(`确定要删除等级 "${tier.name}" 吗？此操作不可撤销。`)) {
    return;
  }

  // 模拟API调用
  const index = tiers.value.findIndex((t) => t.id === tier.id);
  if (index !== -1) {
    tiers.value.splice(index, 1);
  }
};
</script>
