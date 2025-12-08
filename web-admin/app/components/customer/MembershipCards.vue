<template>
  <UCard class="mb-4">
    <template #header>
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h3 class="text-lg font-semibold text-gray-900">
            {{ t("customer.membership.cards.title") }}
          </h3>
          <p class="text-xs text-gray-500">
            {{ t("customer.membership.cards.subtitle") }}
          </p>
        </div>
        <div class="flex items-center gap-3 text-sm text-gray-500">
          <span>
            {{ t("customer.membership.cards.lastUpdated") }}：
            <span class="font-medium text-gray-900">
              {{ lastUpdatedDisplay || t("customer.membership.cards.never") }}
            </span>
          </span>
          <UButton
            size="sm"
            variant="ghost"
            icon="i-heroicons-arrow-path"
            :loading="loading"
            @click="$emit('refresh')"
          >
            {{ t("customer.membership.cards.refresh") }}
          </UButton>
        </div>
      </div>
    </template>
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
      <div
        v-for="card in cards"
        :key="card.key"
        class="rounded-lg border border-gray-100 p-4 shadow-sm bg-white"
      >
        <p class="text-sm text-gray-500">{{ card.label }}</p>
        <p class="mt-2 text-2xl font-semibold text-gray-900">
          <USkeleton v-if="loading" class="h-7 w-20" />
          <span v-else>{{ card.value }}</span>
        </p>
        <p class="text-xs text-gray-400 mt-1">{{ card.description }}</p>
      </div>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "#imports";
import type { MembershipStats } from "~/types/customer";

const props = defineProps<{
  stats: MembershipStats;
  loading?: boolean;
  lastUpdated?: string | null;
}>();

defineEmits<{
  refresh: [];
}>();

const { t } = useI18n();

const cards = computed(() => [
  {
    key: "total",
    label: t("customer.membership.cards.total"),
    value: props.stats?.total?.toLocaleString() ?? "0",
    description: t("customer.membership.cards.totalDesc"),
  },
  {
    key: "active",
    label: t("customer.membership.cards.active"),
    value: props.stats?.active?.toLocaleString() ?? "0",
    description: t("customer.membership.cards.activeDesc"),
  },
  {
    key: "warning",
    label: t("customer.membership.cards.warning"),
    value: props.stats?.warning?.toLocaleString() ?? "0",
    description: t("customer.membership.cards.warningDesc"),
  },
  {
    key: "downgrade",
    label: t("customer.membership.cards.downgrade"),
    value: props.stats?.downgrade?.toLocaleString() ?? "0",
    description: t("customer.membership.cards.downgradeDesc"),
  },
  {
    key: "growth",
    label: t("customer.membership.cards.averageGrowth"),
    value: props.stats?.averageGrowthValue?.toFixed(1) ?? "0.0",
    description: t("customer.membership.cards.averageGrowthDesc"),
  },
]);

const lastUpdatedDisplay = computed(() => {
  if (!props.lastUpdated) return "";
  return new Date(props.lastUpdated).toLocaleString();
});
</script>
