<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">会籍总览</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          查看会籍等级、权益配置与客户会籍入口
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton to="/customer/membership/benefits" variant="soft" color="primary">
          进入会籍权益
        </UButton>
        <UButton to="/customer/membership/tiers" variant="soft" color="primary">
          进入会籍等级
        </UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <UCard>
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-50 text-primary-600 dark:bg-primary-900/40">
            <UIcon name="i-heroicons-star" class="h-5 w-5" />
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">会籍等级</p>
            <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ tiers.length }}</p>
          </div>
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-50 text-amber-600 dark:bg-amber-900/40">
            <UIcon name="i-heroicons-gift" class="h-5 w-5" />
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">会籍权益</p>
            <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ benefits.length }}</p>
          </div>
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600 dark:bg-emerald-900/40">
            <UIcon name="i-heroicons-users" class="h-5 w-5" />
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">客户会籍入口</p>
            <p class="text-sm text-gray-600 dark:text-gray-300">客户详情 → 权益与代币</p>
          </div>
        </div>
      </UCard>
    </div>

    <div class="grid gap-4 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">等级概览</h2>
            <UButton to="/customer/membership/tiers" size="xs" variant="ghost">查看全部</UButton>
          </div>
        </template>
        <UTable :columns="tierColumns" :data="tiers.slice(0, 5)" :loading="loading">
          <template #status-cell="{ row }">
            <UBadge :color="row.status === 'active' ? 'success' : 'neutral'" variant="soft">
              {{ row.status || '-' }}
            </UBadge>
          </template>
        </UTable>
      </UCard>
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">权益概览</h2>
            <UButton to="/customer/membership/benefits" size="xs" variant="ghost">查看全部</UButton>
          </div>
        </template>
        <UTable :columns="benefitColumns" :data="benefits.slice(0, 5)" :loading="loading">
          <template #status-cell="{ row }">
            <UBadge :color="row.status === 'active' ? 'success' : 'neutral'" variant="soft">
              {{ row.status || '-' }}
            </UBadge>
          </template>
        </UTable>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import type { TableColumn } from "@nuxt/ui";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import type { MembershipBenefit, MembershipTier } from "~/types/membership";

const api = useMembershipAdminApi();
const loading = ref(false);
const tiers = ref<MembershipTier[]>([]);
const benefits = ref<MembershipBenefit[]>([]);

const tierColumns: TableColumn<MembershipTier>[] = [
  { accessorKey: "name", header: "等级名称" },
  { accessorKey: "code", header: "编码" },
  { accessorKey: "status", header: "状态" },
];

const benefitColumns: TableColumn<MembershipBenefit>[] = [
  { accessorKey: "name", header: "权益名称" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "status", header: "状态" },
];

const loadData = async () => {
  try {
    loading.value = true;
    const [tierResp, benefitResp] = await Promise.all([api.listTiers(), api.listBenefits()]);
    tiers.value = tierResp?.items ?? [];
    benefits.value = benefitResp?.items ?? [];
  } finally {
    loading.value = false;
  }
};

onMounted(loadData);
</script>
