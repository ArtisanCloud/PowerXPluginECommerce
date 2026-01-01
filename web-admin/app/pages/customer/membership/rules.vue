<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          等级规则
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理会员等级的晋升、降级规则和有效期设置
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" @click="openCreateModal">
        创建规则
      </UButton>
    </div>

    <!-- 规则概览卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon
              name="i-heroicons-arrow-trending-up"
              class="w-6 h-6 text-blue-600 dark:text-blue-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              晋升规则
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ rules.filter((r) => r.type === "promotion").length }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-orange-100 dark:bg-orange-900 rounded-lg">
            <UIcon
              name="i-heroicons-arrow-trending-down"
              class="w-6 h-6 text-orange-600 dark:text-orange-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              降级规则
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ rules.filter((r) => r.type === "demotion").length }}
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
              启用规则
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ rules.filter((r) => r.status === "active").length }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon
              name="i-heroicons-clock"
              class="w-6 h-6 text-purple-600 dark:text-purple-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              有效期规则
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ rules.filter((r) => r.hasExpiry).length }}
            </p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 规则列表 -->
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            规则列表
          </h2>
          <div class="flex gap-2">
            <USelect
              v-model="filterType"
              :options="typeFilterOptions"
              option-attribute="label"
              value-attribute="value"
              size="sm"
            />
            <UInput
              v-model="searchQuery"
              placeholder="搜索规则名称..."
              icon="i-heroicons-magnifying-glass"
              size="sm"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredRules">
        <template #name-cell="{ row }">
          <div class="flex items-center gap-3">
            <div
              class="w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-medium"
              :class="getRuleTypeClass(row.type)"
            >
              <UIcon :name="getRuleTypeIcon(row.type)" class="w-4 h-4" />
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

        <template #type-cell="{ row }">
          <UBadge
            :color="row.type === 'promotion' ? 'success' : 'warning'"
            variant="soft"
          >
            {{ row.type === "promotion" ? "晋升" : "降级" }}
          </UBadge>
        </template>

        <template #conditions-cell="{ row }">
          <div class="space-y-1">
            <div
              v-for="condition in (row.conditions || []).slice(0, 2)"
              :key="condition.type"
              class="text-sm"
            >
              <span class="font-medium"
                >{{ getConditionLabel(condition.type) }}:</span
              >
              <span class="text-gray-600 dark:text-gray-400 ml-1">
                {{ formatConditionValue(condition) }}
              </span>
            </div>
            <div
              v-if="(row.conditions || []).length > 2"
              class="text-xs text-gray-500 dark:text-gray-400"
            >
              +{{ (row.conditions || []).length - 2 }} 个条件
            </div>
          </div>
        </template>

        <template #expiry-cell="{ row }">
          <div v-if="row.hasExpiry" class="text-sm">
            <div class="font-medium text-gray-900 dark:text-white">
              {{ row.expiryDays }} 天
            </div>
            <div class="text-gray-500 dark:text-gray-400">
              {{ row.expiryType === "absolute" ? "绝对期限" : "相对期限" }}
            </div>
          </div>
          <div v-else class="text-gray-500 dark:text-gray-400">永久有效</div>
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
              @click="toggleRuleStatus(row)"
            >
              {{ row.status === "active" ? "停用" : "启用" }}
            </UButton>
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              @click="deleteRule(row)"
            >
              删除
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 查看规则模态框 -->
    <ViewRuleModal
      v-if="viewingRule"
      v-model:open="showViewModal"
      :rule="viewingRule"
      @edit="handleViewToEdit"
    />

    <!-- 创建规则模态框 -->
    <CreateRuleModal
      v-model:open="showCreateModal"
      @created="handleRuleCreated"
    />

    <!-- 编辑规则模态框 -->
    <EditRuleModal
      v-if="editingRule"
      v-model:open="showEditModal"
      :rule="editingRule"
      @updated="handleRuleUpdated"
      @deleted="handleRuleDeleted"
    />
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import ViewRuleModal from "~/components/Modals/ViewRuleModal.vue";
import CreateRuleModal from "~/components/Modals/CreateRuleModal.vue";
import EditRuleModal from "~/components/Modals/EditRuleModal.vue";

// 规则数据类型
type MembershipRule = {
  id: string;
  name: string;
  description: string;
  type: "promotion" | "demotion";
  conditions: Array<{
    type: "spend" | "growth" | "frequency";
    operator: "gte" | "lte" | "eq";
    value: number;
    period?: number; // 时间周期（天）
    description: string;
  }>;
  hasExpiry: boolean;
  expiryType?: "absolute" | "relative";
  expiryDays?: number;
  status: "active" | "inactive";
  priority: number;
  createdAt: string;
  updatedAt: string;
};

// 模态框状态
const showViewModal = ref(false);
const showCreateModal = ref(false);
const showEditModal = ref(false);
const viewingRule = ref<MembershipRule | null>(null);
const editingRule = ref<MembershipRule | null>(null);

// 搜索和过滤
const searchQuery = ref("");
const filterType = ref("all");

// 类型过滤选项
const typeFilterOptions = [
  { label: "全部规则", value: "all" },
  { label: "晋升规则", value: "promotion" },
  { label: "降级规则", value: "demotion" },
];

// 列定义
const columns = computed<TableColumn<MembershipRule>[]>(() => [
  { accessorKey: "name", header: "规则名称" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "conditions", header: "条件" },
  { accessorKey: "expiry", header: "有效期" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

// 规则数据
const rules = ref<MembershipRule[]>([
  {
    id: "rule_1",
    name: "消费晋升规则",
    description: "基于消费金额的等级晋升规则",
    type: "promotion",
    conditions: [
      {
        type: "spend",
        operator: "gte",
        value: 1000,
        period: 365,
        description: "年度消费满1000元",
      },
      {
        type: "frequency",
        operator: "gte",
        value: 5,
        period: 365,
        description: "年度消费次数不少于5次",
      },
    ],
    hasExpiry: false,
    status: "active",
    priority: 1,
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
  {
    id: "rule_2",
    name: "成长值晋升规则",
    description: "基于成长值的等级晋升规则",
    type: "promotion",
    conditions: [
      {
        type: "growth",
        operator: "gte",
        value: 500,
        description: "累计成长值达到500",
      },
    ],
    hasExpiry: true,
    expiryType: "relative",
    expiryDays: 365,
    status: "active",
    priority: 2,
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
  {
    id: "rule_3",
    name: "活跃度降级规则",
    description: "基于活跃度的等级降级规则",
    type: "demotion",
    conditions: [
      {
        type: "frequency",
        operator: "lte",
        value: 0,
        period: 180,
        description: "180天内无消费记录",
      },
    ],
    hasExpiry: false,
    status: "active",
    priority: 3,
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
  {
    id: "rule_4",
    name: "VIP专属晋升规则",
    description: "VIP客户专属的快速晋升规则",
    type: "promotion",
    conditions: [
      {
        type: "spend",
        operator: "gte",
        value: 5000,
        period: 90,
        description: "90天内消费满5000元",
      },
      {
        type: "growth",
        operator: "gte",
        value: 1000,
        description: "累计成长值达到1000",
      },
    ],
    hasExpiry: true,
    expiryType: "absolute",
    expiryDays: 730,
    status: "inactive",
    priority: 4,
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
]);

// 过滤后的规则列表
const filteredRules = computed(() => {
  let filtered = rules.value;

  // 类型过滤
  if (filterType.value !== "all") {
    filtered = filtered.filter((rule) => rule.type === filterType.value);
  }

  // 搜索过滤
  if (searchQuery.value) {
    filtered = filtered.filter(
      (rule) =>
        rule.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        rule.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    );
  }

  return filtered;
});

// 获取规则类型图标
const getRuleTypeIcon = (type: string) => {
  return type === "promotion"
    ? "i-heroicons-arrow-trending-up"
    : "i-heroicons-arrow-trending-down";
};

// 获取规则类型样式
const getRuleTypeClass = (type: string) => {
  return type === "promotion"
    ? "bg-green-500 dark:bg-green-600"
    : "bg-orange-500 dark:bg-orange-600";
};

// 获取条件标签
const getConditionLabel = (type: string) => {
  const labels = {
    spend: "消费金额",
    growth: "成长值",
    frequency: "消费次数",
  };
  return labels[type as keyof typeof labels] || type;
};

// 格式化条件值
const formatConditionValue = (condition: any) => {
  const operators = {
    gte: "≥",
    lte: "≤",
    eq: "=",
  };

  let value = "";
  if (condition.type === "spend") {
    value = `¥${condition.value.toLocaleString()}`;
  } else {
    value = condition.value.toString();
  }

  let result = `${operators[condition.operator as keyof typeof operators]} ${value}`;

  if (condition.period) {
    result += ` (${condition.period}天内)`;
  }

  return result;
};

// 打开查看模态框
const openViewModal = (rule: MembershipRule) => {
  viewingRule.value = rule;
  showViewModal.value = true;
};

// 打开创建模态框
const openCreateModal = () => {
  showCreateModal.value = true;
};

// 打开编辑模态框
const openEditModal = (rule: MembershipRule) => {
  editingRule.value = rule;
  showEditModal.value = true;
};

// 从查看切换到编辑
const handleViewToEdit = (rule: MembershipRule) => {
  showViewModal.value = false;
  viewingRule.value = null;
  editingRule.value = rule;
  showEditModal.value = true;
};

// 处理规则创建
const handleRuleCreated = (newRule: MembershipRule) => {
  rules.value.push(newRule);
  showCreateModal.value = false;
};

// 处理规则更新
const handleRuleUpdated = (updatedRule: MembershipRule) => {
  const index = rules.value.findIndex((r) => r.id === updatedRule.id);
  if (index !== -1) {
    rules.value[index] = updatedRule;
  }
  showEditModal.value = false;
  editingRule.value = null;
};

// 处理规则删除
const handleRuleDeleted = (ruleId: string) => {
  const index = rules.value.findIndex((r) => r.id === ruleId);
  if (index !== -1) {
    rules.value.splice(index, 1);
  }
  showEditModal.value = false;
  editingRule.value = null;
};

// 切换规则状态
const toggleRuleStatus = async (rule: MembershipRule) => {
  const newStatus = rule.status === "active" ? "inactive" : "active";
  const action = newStatus === "active" ? "启用" : "停用";

  if (!confirm(`确定要${action}规则 "${rule.name}" 吗？`)) {
    return;
  }

  // 模拟API调用
  const index = rules.value.findIndex((r) => r.id === rule.id);
  if (index !== -1) {
    rules.value[index].status = newStatus;
    rules.value[index].updatedAt = new Date().toISOString();
  }
};

// 删除规则
const deleteRule = async (rule: MembershipRule) => {
  if (!confirm(`确定要删除规则 "${rule.name}" 吗？此操作不可撤销。`)) {
    return;
  }

  // 模拟API调用
  const index = rules.value.findIndex((r) => r.id === rule.id);
  if (index !== -1) {
    rules.value.splice(index, 1);
  }
};
</script>
