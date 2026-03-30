<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">等级规则</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          基于会籍等级配置中的 rules 字段展示与管理规则
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" @click="goCreateTier">
        创建规则（等级）
      </UButton>
    </div>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-4">
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">晋升规则</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ promotionCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">降级规则</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ demotionCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">启用规则</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ activeCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">有效期规则</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ expiryCount }}</p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">规则列表</h2>
          <div class="flex gap-2">
            <USelect
              v-model="filterType"
              :items="typeFilterOptions"
              option-attribute="label"
              value-attribute="value"
              size="sm"
              class="min-w-32"
            />
            <UInput
              v-model="searchQuery"
              placeholder="搜索规则名称/等级编码"
              icon="i-heroicons-magnifying-glass"
              size="sm"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredRules" :loading="loading">
        <template #name-cell="{ row }">
          <div>
            <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400">等级编码：{{ row.code }}</div>
          </div>
        </template>

        <template #type-cell="{ row }">
          <UBadge :color="row.type === 'promotion' ? 'success' : 'warning'" variant="soft">
            {{ row.type === "promotion" ? "晋升" : "降级" }}
          </UBadge>
        </template>

        <template #conditions-cell="{ row }">
          <div class="text-sm text-gray-600 dark:text-gray-400">{{ summarizeConditions(row.conditions) }}</div>
        </template>

        <template #expiry-cell="{ row }">
          <div v-if="row.hasExpiry" class="text-sm text-gray-700 dark:text-gray-300">
            {{ row.expiryDays }} 天（{{ row.expiryType === "absolute" ? "绝对" : "相对" }}）
          </div>
          <div v-else class="text-sm text-gray-500 dark:text-gray-400">永久有效</div>
        </template>

        <template #status-cell="{ row }">
          <UBadge :color="row.status === 'active' ? 'success' : 'neutral'" variant="soft">
            {{ row.status || "-" }}
          </UBadge>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton color="neutral" variant="ghost" size="sm" @click="viewRule(row)">
              查看
            </UButton>
            <UButton
              :color="row.status === 'active' ? 'warning' : 'success'"
              variant="ghost"
              size="sm"
              :loading="rowActionId === row.id && rowActionType === 'toggle'"
              @click="toggleRuleStatus(row)"
            >
              {{ row.status === "active" ? "停用" : "启用" }}
            </UButton>
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              :loading="rowActionId === row.id && rowActionType === 'delete'"
              @click="deleteRule(row)"
            >
              删除
            </UButton>
          </div>
        </template>
      </UTable>

      <UAlert v-if="!filteredRules.length && !loading" color="gray" class="mt-4">
        暂无规则（请先在会籍等级里创建）
      </UAlert>
    </UCard>

    <UModal v-model:open="viewOpen">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">规则详情</h3>
          </template>
          <div v-if="viewingRule" class="space-y-3 text-sm">
            <div><span class="font-medium">规则名称：</span>{{ viewingRule.name }}</div>
            <div><span class="font-medium">等级编码：</span>{{ viewingRule.code }}</div>
            <div>
              <span class="font-medium">规则 JSON：</span>
              <pre class="mt-2 overflow-auto rounded bg-gray-100 p-3 text-xs dark:bg-gray-800">{{ prettyRules(viewingRule.rawRules) }}</pre>
            </div>
          </div>
        </UCard>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import type { TableColumn } from "@nuxt/ui";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import type { MembershipTier } from "~/types/membership";
import { useToast, navigateTo } from "#imports";

type RuleCondition = {
  type?: string;
  operator?: string;
  value?: number;
  period?: number;
};

type RuleRow = {
  id: string;
  name: string;
  code: string;
  type: "promotion" | "demotion";
  conditions: RuleCondition[];
  hasExpiry: boolean;
  expiryType?: "absolute" | "relative";
  expiryDays?: number;
  status: string;
  rawRules: Record<string, any>;
};

const api = useMembershipAdminApi();
const toast = useToast();

const loading = ref(false);
const rowActionId = ref("");
const rowActionType = ref<"" | "toggle" | "delete">("");
const searchQuery = ref("");
const filterType = ref("all");
const tiers = ref<MembershipTier[]>([]);
const viewOpen = ref(false);
const viewingRule = ref<RuleRow | null>(null);

const typeFilterOptions = [
  { label: "全部规则", value: "all" },
  { label: "晋升规则", value: "promotion" },
  { label: "降级规则", value: "demotion" },
];

const columns = computed<TableColumn<RuleRow>[]>(() => [
  { accessorKey: "name", header: "规则名称" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "conditions", header: "条件" },
  { accessorKey: "expiry", header: "有效期" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

const mapTierToRule = (tier: MembershipTier): RuleRow => {
  const rawRules = tier.rules && typeof tier.rules === "object" ? tier.rules : {};
  const type = rawRules.type === "demotion" ? "demotion" : "promotion";
  const conditions = Array.isArray(rawRules.conditions) ? rawRules.conditions : [];
  const expiryDays = Number(rawRules.expiryDays || 0) || undefined;
  const hasExpiry = Boolean(rawRules.hasExpiry || expiryDays);
  const expiryType = rawRules.expiryType === "absolute" ? "absolute" : "relative";

  return {
    id: tier.id,
    name: tier.name || "未命名规则",
    code: tier.code || "-",
    type,
    conditions,
    hasExpiry,
    expiryType,
    expiryDays,
    status: tier.status || "draft",
    rawRules,
  };
};

const rules = computed(() => tiers.value.map(mapTierToRule));

const filteredRules = computed(() => {
  let list = rules.value;

  if (filterType.value !== "all") {
    list = list.filter((rule) => rule.type === filterType.value);
  }

  const keyword = searchQuery.value.trim().toLowerCase();
  if (keyword) {
    list = list.filter((rule) => {
      return (
        rule.name.toLowerCase().includes(keyword) ||
        rule.code.toLowerCase().includes(keyword)
      );
    });
  }

  return list;
});

const promotionCount = computed(() => rules.value.filter((r) => r.type === "promotion").length);
const demotionCount = computed(() => rules.value.filter((r) => r.type === "demotion").length);
const activeCount = computed(() => rules.value.filter((r) => r.status === "active").length);
const expiryCount = computed(() => rules.value.filter((r) => r.hasExpiry).length);

const loadRules = async () => {
  try {
    loading.value = true;
    const resp = await api.listTiers();
    tiers.value = resp?.items ?? [];
  } catch (error: any) {
    toast.add({ title: "加载规则失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    loading.value = false;
  }
};

const summarizeConditions = (conditions: RuleCondition[]) => {
  if (!Array.isArray(conditions) || conditions.length === 0) return "未配置";
  return `${conditions.length} 条条件`;
};

const prettyRules = (value: Record<string, any>) => JSON.stringify(value || {}, null, 2);

const viewRule = (rule: RuleRow) => {
  viewingRule.value = rule;
  viewOpen.value = true;
};

const goCreateTier = () => {
  navigateTo("/customer/membership/tiers");
};

const toggleRuleStatus = async (rule: RuleRow) => {
  const nextStatus = rule.status === "active" ? "inactive" : "active";
  rowActionId.value = rule.id;
  rowActionType.value = "toggle";
  try {
    await api.updateTierStatus(rule.id, { status: nextStatus });
    toast.add({ title: `规则已${nextStatus === "active" ? "启用" : "停用"}`, color: "success" });
    await loadRules();
  } catch (error: any) {
    toast.add({ title: "更新状态失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    rowActionId.value = "";
    rowActionType.value = "";
  }
};

const deleteRule = async (rule: RuleRow) => {
  if (!window.confirm(`确定删除规则 \"${rule.name}\" 吗？`)) return;
  rowActionId.value = rule.id;
  rowActionType.value = "delete";
  try {
    await api.deleteTier(rule.id);
    toast.add({ title: "规则已删除", color: "success" });
    await loadRules();
  } catch (error: any) {
    toast.add({ title: "删除失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    rowActionId.value = "";
    rowActionType.value = "";
  }
};

onMounted(() => {
  loadRules();
});
</script>
