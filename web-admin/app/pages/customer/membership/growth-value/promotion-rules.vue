<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">等级晋升逻辑</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          基于会籍等级 rules 字段展示成长值晋升规则
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" @click="goCreateTier">
        创建规则（等级）
      </UButton>
    </div>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-4">
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">总规则数</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ totalRules }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">自动晋升</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ autoCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">手动晋升</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ manualCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">启用规则</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ activeCount }}</p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">晋升规则列表</h2>
          <div class="flex gap-2">
            <USelect
              v-model="modeFilter"
              :items="modeFilterOptions"
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

        <template #mode-cell="{ row }">
          <UBadge :color="row.mode === 'auto' ? 'success' : 'warning'" variant="soft">
            {{ row.mode === "auto" ? "自动" : "手动" }}
          </UBadge>
        </template>

        <template #threshold-cell="{ row }">
          <div class="text-sm text-gray-700 dark:text-gray-300">
            {{ row.threshold > 0 ? `${row.threshold} 成长值` : "未配置" }}
          </div>
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
        暂无晋升规则（请先在会籍等级里配置 rules）
      </UAlert>
    </UCard>

    <UModal v-model:open="viewOpen">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">晋升规则详情</h3>
          </template>
          <div v-if="viewingRule" class="space-y-3 text-sm">
            <div><span class="font-medium">规则名称：</span>{{ viewingRule.name }}</div>
            <div><span class="font-medium">等级编码：</span>{{ viewingRule.code }}</div>
            <div><span class="font-medium">晋升模式：</span>{{ viewingRule.mode === "auto" ? "自动" : "手动" }}</div>
            <div><span class="font-medium">成长值门槛：</span>{{ viewingRule.threshold || 0 }}</div>
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
import { navigateTo, useToast } from "#imports";

type RuleMode = "auto" | "manual";

type PromotionRuleRow = {
  id: string;
  name: string;
  code: string;
  mode: RuleMode;
  threshold: number;
  status: string;
  rawRules: Record<string, any>;
};

const api = useMembershipAdminApi();
const toast = useToast();

const loading = ref(false);
const rowActionId = ref("");
const rowActionType = ref<"" | "toggle" | "delete">("");
const searchQuery = ref("");
const modeFilter = ref<"all" | RuleMode>("all");
const tiers = ref<MembershipTier[]>([]);
const viewOpen = ref(false);
const viewingRule = ref<PromotionRuleRow | null>(null);

const modeFilterOptions = [
  { label: "全部", value: "all" },
  { label: "自动", value: "auto" },
  { label: "手动", value: "manual" },
];

const columns = computed<TableColumn<PromotionRuleRow>[]>(() => [
  { accessorKey: "name", header: "规则名称" },
  { accessorKey: "mode", header: "晋升模式" },
  { accessorKey: "threshold", header: "成长值门槛" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

const parseMode = (rules: Record<string, any>): RuleMode => {
  if (rules.mode === "manual" || rules.approvalProcess || rules.approvers) {
    return "manual";
  }
  return "auto";
};

const parseThreshold = (rules: Record<string, any>): number => {
  const upgrade = rules.upgrade && typeof rules.upgrade === "object" ? rules.upgrade : null;
  const value = Number(
    upgrade?.minGrowthValue ??
    rules.growthValueThreshold ??
    rules.threshold ??
    0,
  );
  return Number.isFinite(value) && value > 0 ? value : 0;
};

const mapTierToRule = (tier: MembershipTier): PromotionRuleRow => {
  const rawRules = tier.rules && typeof tier.rules === "object" ? tier.rules : {};
  return {
    id: tier.id,
    name: tier.name || "未命名规则",
    code: tier.code || "-",
    mode: parseMode(rawRules),
    threshold: parseThreshold(rawRules),
    status: tier.status || "draft",
    rawRules,
  };
};

const rules = computed(() => tiers.value.map(mapTierToRule));

const filteredRules = computed(() => {
  let list = rules.value;

  if (modeFilter.value !== "all") {
    list = list.filter((rule) => rule.mode === modeFilter.value);
  }

  const keyword = searchQuery.value.trim().toLowerCase();
  if (keyword) {
    list = list.filter((rule) => (
      rule.name.toLowerCase().includes(keyword)
      || rule.code.toLowerCase().includes(keyword)
    ));
  }

  return list;
});

const totalRules = computed(() => rules.value.length);
const autoCount = computed(() => rules.value.filter((rule) => rule.mode === "auto").length);
const manualCount = computed(() => rules.value.filter((rule) => rule.mode === "manual").length);
const activeCount = computed(() => rules.value.filter((rule) => rule.status === "active").length);

const loadRules = async () => {
  loading.value = true;
  try {
    const resp = await api.listTiers();
    tiers.value = resp?.items ?? [];
  } catch (error: any) {
    toast.add({
      title: "加载晋升规则失败",
      description: error?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    loading.value = false;
  }
};

const prettyRules = (value: Record<string, any>) => JSON.stringify(value || {}, null, 2);

const viewRule = (rule: PromotionRuleRow) => {
  viewingRule.value = rule;
  viewOpen.value = true;
};

const goCreateTier = () => {
  navigateTo("/customer/membership/tiers");
};

const toggleRuleStatus = async (rule: PromotionRuleRow) => {
  const nextStatus = rule.status === "active" ? "inactive" : "active";
  rowActionId.value = rule.id;
  rowActionType.value = "toggle";
  try {
    await api.updateTierStatus(rule.id, { status: nextStatus });
    toast.add({
      title: `规则已${nextStatus === "active" ? "启用" : "停用"}`,
      color: "success",
    });
    await loadRules();
  } catch (error: any) {
    toast.add({
      title: "更新状态失败",
      description: error?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    rowActionId.value = "";
    rowActionType.value = "";
  }
};

const deleteRule = async (rule: PromotionRuleRow) => {
  if (!window.confirm(`确定删除规则 \"${rule.name}\" 吗？`)) return;
  rowActionId.value = rule.id;
  rowActionType.value = "delete";
  try {
    await api.deleteTier(rule.id);
    toast.add({ title: "规则已删除", color: "success" });
    await loadRules();
  } catch (error: any) {
    toast.add({
      title: "删除失败",
      description: error?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    rowActionId.value = "";
    rowActionType.value = "";
  }
};

onMounted(() => {
  loadRules();
});
</script>
