<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">积分规则</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          基于会籍等级配置中的 rules 字段展示积分规则
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" @click="goManageTiers">
        创建规则（等级）
      </UButton>
    </div>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">总规则数</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ totalRules }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">获取规则</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ earnCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500 dark:text-gray-400">消耗规则</p>
        <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ spendCount }}</p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">规则列表</h2>
          <div class="flex gap-2">
            <USelect
              v-model="activeTab"
              :items="tabOptions"
              option-attribute="label"
              value-attribute="value"
              size="sm"
              class="min-w-28"
            />
            <UInput
              v-model="searchQuery"
              placeholder="搜索规则名称/等级"
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
            <div class="text-xs text-gray-500 dark:text-gray-400">等级：{{ row.tierName }}</div>
          </div>
        </template>

        <template #type-cell="{ row }">
          <UBadge :color="row.category === 'earn' ? 'success' : 'warning'" variant="soft">
            {{ row.category === "earn" ? "获取" : "消耗" }}
          </UBadge>
        </template>

        <template #points-cell="{ row }">
          <span class="font-medium" :class="row.points >= 0 ? 'text-green-600' : 'text-orange-600'">
            {{ row.points >= 0 ? `+${row.points}` : row.points }}
          </span>
        </template>

        <template #status-cell="{ row }">
          <UBadge :color="row.status === 'active' ? 'success' : 'neutral'" variant="soft">
            {{ row.status || "draft" }}
          </UBadge>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="sm" color="neutral" variant="ghost" @click="viewRule(row)">
              查看
            </UButton>
            <UButton size="sm" color="primary" variant="ghost" @click="editTier(row)">
              去维护
            </UButton>
          </div>
        </template>
      </UTable>

      <UAlert v-if="!filteredRules.length && !loading" color="gray" class="mt-4">
        暂无积分规则（请先在会籍等级配置 rules）
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
            <div><span class="font-medium">等级：</span>{{ viewingRule.tierName }}</div>
            <div><span class="font-medium">类型：</span>{{ viewingRule.category === "earn" ? "获取" : "消耗" }}</div>
            <div><span class="font-medium">积分：</span>{{ viewingRule.points }}</div>
            <div>
              <span class="font-medium">规则 JSON：</span>
              <pre class="mt-2 overflow-auto rounded bg-gray-100 p-3 text-xs dark:bg-gray-800">{{ prettyRule(viewingRule.raw) }}</pre>
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
import { navigateTo, useToast } from "#imports";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import type { MembershipTier } from "~/types/membership";

type PointRuleRow = {
  id: string;
  tierId: string;
  tierName: string;
  name: string;
  category: "earn" | "spend";
  points: number;
  status: string;
  raw: Record<string, any>;
};

const api = useMembershipAdminApi();
const toast = useToast();

const loading = ref(false);
const tiers = ref<MembershipTier[]>([]);
const activeTab = ref("all");
const searchQuery = ref("");
const viewOpen = ref(false);
const viewingRule = ref<PointRuleRow | null>(null);

const tabOptions = [
  { label: "所有规则", value: "all" },
  { label: "获取规则", value: "earn" },
  { label: "消耗规则", value: "spend" },
];

const columns = computed<TableColumn<PointRuleRow>[]>(() => [
  { accessorKey: "name", header: "规则名称" },
  { accessorKey: "type", header: "规则类型" },
  { accessorKey: "points", header: "积分" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

const toNumber = (value: unknown) => {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
};

const inferCategory = (inputType: string, points: number): "earn" | "spend" => {
  const t = inputType.toLowerCase();
  if (t.includes("spend") || t.includes("deduct") || t.includes("redeem") || t.includes("consume") || t.includes("expiration")) {
    return "spend";
  }
  if (t.includes("earn") || t.includes("reward") || t.includes("purchase") || t.includes("checkin") || t.includes("share") || t.includes("referral")) {
    return "earn";
  }
  return points < 0 ? "spend" : "earn";
};

const collectPointRules = (tier: MembershipTier): PointRuleRow[] => {
  const tierRules = tier.rules && typeof tier.rules === "object" ? (tier.rules as Record<string, any>) : {};

  const rows: PointRuleRow[] = [];
  const pushRule = (raw: Record<string, any>, index: number) => {
    const type = String(raw.type || raw.category || raw.ruleType || "points_rule");
    const points = toNumber(raw.points ?? raw.delta ?? raw.value ?? raw.amount ?? 0);
    const category = inferCategory(type, points);
    const name = String(raw.name || raw.title || `${tier.name} · ${category === "earn" ? "获取" : "消耗"}规则 ${index + 1}`);
    rows.push({
      id: `${tier.id}-${index}`,
      tierId: tier.id,
      tierName: tier.name,
      name,
      category,
      points,
      status: String(raw.status || tier.status || "draft"),
      raw,
    });
  };

  if (Array.isArray(tierRules.pointsRules)) {
    tierRules.pointsRules.forEach((item: any, idx: number) => {
      if (item && typeof item === "object") pushRule(item, idx);
    });
  }

  if (tierRules.points && typeof tierRules.points === "object") {
    const pointsObj = tierRules.points as Record<string, any>;
    if (Array.isArray(pointsObj.earnRules)) {
      pointsObj.earnRules.forEach((item: any, idx: number) => {
        if (item && typeof item === "object") pushRule({ ...item, category: "earn" }, rows.length + idx);
      });
    }
    if (Array.isArray(pointsObj.spendRules)) {
      pointsObj.spendRules.forEach((item: any, idx: number) => {
        if (item && typeof item === "object") pushRule({ ...item, category: "spend" }, rows.length + idx);
      });
    }
  }

  if (!rows.length && Object.keys(tierRules).length > 0) {
    const fallbackPoints = toNumber(tierRules.points ?? tierRules.defaultPoints ?? 0);
    rows.push({
      id: `${tier.id}-fallback`,
      tierId: tier.id,
      tierName: tier.name,
      name: `${tier.name} 默认积分规则`,
      category: fallbackPoints < 0 ? "spend" : "earn",
      points: fallbackPoints,
      status: String(tier.status || "draft"),
      raw: tierRules,
    });
  }

  return rows;
};

const allRules = computed<PointRuleRow[]>(() =>
  tiers.value.flatMap((tier) => collectPointRules(tier)),
);

const filteredRules = computed(() => {
  let list = allRules.value;
  if (activeTab.value !== "all") {
    list = list.filter((item) => item.category === activeTab.value);
  }
  const keyword = searchQuery.value.trim().toLowerCase();
  if (keyword) {
    list = list.filter((item) => {
      return (
        item.name.toLowerCase().includes(keyword) ||
        item.tierName.toLowerCase().includes(keyword)
      );
    });
  }
  return list;
});

const totalRules = computed(() => allRules.value.length);
const earnCount = computed(() => allRules.value.filter((item) => item.category === "earn").length);
const spendCount = computed(() => allRules.value.filter((item) => item.category === "spend").length);

const loadRules = async () => {
  try {
    loading.value = true;
    const resp = await api.listTiers();
    tiers.value = resp?.items || [];
  } catch (error: any) {
    toast.add({ title: "加载积分规则失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    loading.value = false;
  }
};

const goManageTiers = () => navigateTo("/customer/membership/tiers");

const viewRule = (rule: PointRuleRow) => {
  viewingRule.value = rule;
  viewOpen.value = true;
};

const editTier = (rule: PointRuleRow) => {
  navigateTo("/customer/membership/tiers");
};

const prettyRule = (rule: Record<string, any>) => JSON.stringify(rule || {}, null, 2);

onMounted(() => {
  loadRules();
});
</script>
