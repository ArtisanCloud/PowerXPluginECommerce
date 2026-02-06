<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">会籍权益</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          查看会籍权益配置（仅展示，编辑功能待开放）
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" :disabled="true">
        创建权益（未开放）
      </UButton>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">权益列表</h2>
          <UInput
            v-model="searchQuery"
            placeholder="搜索权益名称"
            icon="i-heroicons-magnifying-glass"
            size="sm"
          />
        </div>
      </template>

      <UTable :columns="columns" :data="filteredBenefits" :loading="loading">
        <template #status-cell="{ row }">
          <UBadge :color="row.status === 'active' ? 'success' : 'neutral'" variant="soft">
            {{ row.status || '-' }}
          </UBadge>
        </template>
        <template #items-cell="{ row }">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ formatItems(row.items) }}
          </span>
        </template>
        <template #createdAt-cell="{ row }">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ formatDate(row.createdAt) }}
          </span>
        </template>
      </UTable>

      <UAlert v-if="!filteredBenefits.length && !loading" color="gray" class="mt-4">
        暂无权益配置
      </UAlert>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import type { MembershipBenefit } from "~/types/membership";
import type { TableColumn } from "@nuxt/ui";

const api = useMembershipAdminApi();
const loading = ref(false);
const searchQuery = ref("");
const benefits = ref<MembershipBenefit[]>([]);

const columns = computed<TableColumn<MembershipBenefit>[]>(() => [
  { accessorKey: "name", header: "权益名称" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "items", header: "内容" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "创建时间" },
]);

const filteredBenefits = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  if (!keyword) return benefits.value;
  return benefits.value.filter((benefit) =>
    String(benefit.name || "").toLowerCase().includes(keyword),
  );
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

const formatItems = (items?: Record<string, any> | null) => {
  if (!items) return "-";
  if (Array.isArray(items)) return `${items.length} 项`;
  return "已配置";
};

const loadBenefits = async () => {
  try {
    loading.value = true;
    const resp = await api.listBenefits();
    benefits.value = resp?.items ?? [];
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadBenefits();
});
</script>
