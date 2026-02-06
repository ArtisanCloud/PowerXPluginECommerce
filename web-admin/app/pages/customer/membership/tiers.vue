<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">会籍等级</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          查看会籍等级配置（仅展示，编辑功能待开放）
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" :disabled="true">
        创建等级（未开放）
      </UButton>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">等级列表</h2>
          <UInput
            v-model="searchQuery"
            placeholder="搜索等级名称/编码"
            icon="i-heroicons-magnifying-glass"
            size="sm"
          />
        </div>
      </template>

      <UTable :columns="columns" :data="filteredTiers" :loading="loading">
        <template #status-cell="{ row }">
          <UBadge :color="row.status === 'active' ? 'success' : 'neutral'" variant="soft">
            {{ row.status || '-' }}
          </UBadge>
        </template>
        <template #createdAt-cell="{ row }">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ formatDate(row.createdAt) }}
          </span>
        </template>
      </UTable>

      <UAlert v-if="!filteredTiers.length && !loading" color="gray" class="mt-4">
        暂无会籍等级配置
      </UAlert>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import type { MembershipTier } from "~/types/membership";
import type { TableColumn } from "@nuxt/ui";

const api = useMembershipAdminApi();
const loading = ref(false);
const searchQuery = ref("");
const tiers = ref<MembershipTier[]>([]);

const columns = computed<TableColumn<MembershipTier>[]>(() => [
  { accessorKey: "name", header: "等级名称" },
  { accessorKey: "code", header: "等级编码" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "创建时间" },
]);

const filteredTiers = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  if (!keyword) return tiers.value;
  return tiers.value.filter((tier) => {
    return (
      String(tier.name || "").toLowerCase().includes(keyword) ||
      String(tier.code || "").toLowerCase().includes(keyword)
    );
  });
});

const formatDate = (value?: string) => {
  if (!value) return "-";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "short",
    day: "numeric",
  }).format(parsed);
};

const loadTiers = async () => {
  try {
    loading.value = true;
    const resp = await api.listTiers();
    tiers.value = resp?.items ?? [];
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadTiers();
});
</script>
