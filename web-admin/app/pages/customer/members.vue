<template>
  <div class="space-y-4">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">
          {{ t("customer.membership.title") }}
        </h1>
        <p class="text-sm text-gray-500">
          {{ t("customer.membership.subtitle") }}
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton
          icon="i-heroicons-arrow-down-tray"
          variant="ghost"
          :loading="exporting"
          @click="handleExport"
        >
          {{ t("customer.membership.actions.export") }}
        </UButton>
        <UButton
          icon="i-heroicons-arrow-path"
          variant="ghost"
          :loading="membershipLoading"
          @click="handleRefresh"
        >
          {{ t("customer.membership.actions.refresh") }}
        </UButton>
      </div>
    </header>

    <MembershipFilterBar />

    <MembershipCards
      :stats="membershipStats"
      :loading="membershipLoading"
      :last-updated="membershipLastFetchedAt"
      @refresh="handleRefresh"
    />

    <UAlert
      v-if="membershipError"
      color="red"
      :title="t('customer.membership.messages.errorTitle')"
      :description="membershipError"
    >
      <template #actions>
        <UButton
          size="xs"
          color="red"
          variant="solid"
          :loading="membershipLoading"
          @click="handleRefresh"
        >
          {{ t("customer.directory.actions.retry") }}
        </UButton>
      </template>
    </UAlert>

    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap items-center gap-2">
        <UBadge
          v-for="segment in segments"
          :key="segment.key"
          :color="segment.color"
          :variant="selectedSegment === segment.key ? 'solid' : 'subtle'"
          class="cursor-pointer"
          @click="handleSegmentClick(segment.key)"
        >
          {{ segment.label }} · {{ segment.count }}
        </UBadge>
        <UButton
          v-if="selectedSegment"
          size="xs"
          variant="ghost"
          icon="i-heroicons-x-mark"
          @click="handleClearSegment"
        >
          {{ t("customer.membership.actions.clearSegment") }}
        </UButton>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton
          icon="i-heroicons-exclamation-circle"
          variant="soft"
          @click="handleSegmentClick('downgrade')"
        >
          {{ t("customer.membership.actions.quickDowngrade") }}
        </UButton>
        <UButton
          color="primary"
          icon="i-heroicons-megaphone"
          :disabled="!hasSelection"
          @click="reminderDrawerOpen = true"
        >
          {{ t("customer.membership.actions.remind") }}
        </UButton>
      </div>
    </div>

    <UCard>
      <div class="flex items-center justify-between mb-3 text-sm text-gray-500">
        <span>
          {{ t("customer.directory.status.total", { total: membershipStats.total }) }}
        </span>
        <USelectMenu
          v-model="pageSize"
          :options="pageSizeOptions"
          value-attribute="value"
          option-attribute="label"
          class="w-32"
        />
      </div>
      <UTable :rows="rows" :columns="columns" :loading="membershipLoading">
        <template #select-cell="{ row }">
          <UCheckbox :model-value="selectionSet.has(row.id)" @change="() => toggleRow(row.id)" />
        </template>
        <template #name-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-medium text-gray-900">{{ row.customer.name }}</span>
            <span class="text-xs text-gray-500">{{ row.customer.id }}</span>
          </div>
        </template>
        <template #tier-cell="{ row }">
          <UBadge variant="subtle">
            {{ row.snapshot.tier || "—" }}
          </UBadge>
        </template>
        <template #growthValue-cell="{ row }">
          {{ row.snapshot.growthValue ?? "—" }}
        </template>
        <template #points-cell="{ row }">
          {{ row.snapshot.points ?? "—" }}
        </template>
        <template #retentionStatus-cell="{ row }">
          <UBadge :color="retentionColor(row.snapshot.retentionStatus)" variant="subtle">
            {{ retentionLabel(row.snapshot.retentionStatus) }}
          </UBadge>
        </template>
        <template #lastBenefitUsedAt-cell="{ row }">
          {{ formatDate(row.snapshot.lastBenefitUsedAt) }}
        </template>
        <template #actions-cell="{ row }">
          <UButton
            size="xs"
            variant="ghost"
            icon="i-heroicons-bell-alert"
            @click="openReminderFor(row.id)"
          >
            {{ t("customer.membership.table.remind") }}
          </UButton>
        </template>
        <template #empty>
          <div class="py-6 text-center text-sm text-gray-500">
            {{ t("customer.membership.table.empty") }}
          </div>
        </template>
      </UTable>

      <div class="mt-4 flex items-center justify-between text-sm text-gray-500">
        <span>
          {{ t("customer.directory.status.selected", { count: selectionSet.size }) }}
        </span>
        <UPagination v-model="page" :total="membershipStats.total" :page-count="pageSize" />
      </div>
    </UCard>

    <MembershipReminderDrawer
      v-model="reminderDrawerOpen"
      :selected-ids="selection"
      @completed="handleReminderComplete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useI18n, useToast } from "#imports";
import type { TableColumn } from "@nuxt/ui";
import MembershipFilterBar from "~/components/customer/MembershipFilterBar.vue";
import MembershipCards from "~/components/customer/MembershipCards.vue";
import MembershipReminderDrawer from "~/components/customer/MembershipReminderDrawer.vue";
import { useCustomerStore } from "~/stores/customer";
import { useCustomerMetrics } from "~/composables/useCustomerMetrics";
import { useCustomerBulkActions } from "~/composables/useCustomerBulkActions";
import type { MembershipInsight, CustomerListFilters } from "~/types/customer";

type MembershipRow = {
  id: string;
  customer: MembershipInsight["customer"];
  snapshot: MembershipInsight["snapshot"];
};

const store = useCustomerStore();
const {
  membershipInsights,
  membershipStats,
  membershipSegments,
  membershipLoading,
  membershipError,
  membershipLastFetchedAt,
  membershipFilters,
  selection,
} = storeToRefs(store);

const { t } = useI18n();
const toast = useToast();
const metrics = useCustomerMetrics();
const bulkActions = useCustomerBulkActions();

const reminderDrawerOpen = ref(false);
const exporting = ref(false);
const page = ref(membershipFilters.value.page || 1);
const pageSize = ref(membershipFilters.value.pageSize || 20);
const selectedSegment = ref<string | null>(membershipFilters.value.retentionStatus || null);

const rows = computed<MembershipRow[]>(() =>
  membershipInsights.value.map((entry) => ({
    id: entry.customer.id,
    customer: entry.customer,
    snapshot: entry.snapshot,
  }))
);

const selectionSet = computed(() => new Set(selection.value));
const hasSelection = computed(() => selection.value.length > 0);

const segments = computed(() => [
  {
    key: "safe",
    label: t("customer.membership.segments.safe"),
    count: membershipSegments.value.safe,
    color: "success",
  },
  {
    key: "warning",
    label: t("customer.membership.segments.warning"),
    count: membershipSegments.value.warning,
    color: "warning",
  },
  {
    key: "downgrade",
    label: t("customer.membership.segments.downgrade"),
    count: membershipSegments.value.downgrade,
    color: "error",
  },
]);

const columns = computed<TableColumn<MembershipRow>[]>(() => [
  { accessorKey: "select", header: "", size: 48, sortable: false },
  { accessorKey: "name", header: t("customer.membership.table.name") },
  { accessorKey: "tier", header: t("customer.membership.table.tier") },
  { accessorKey: "growthValue", header: t("customer.membership.table.growth") },
  { accessorKey: "points", header: t("customer.membership.table.points") },
  { accessorKey: "retentionStatus", header: t("customer.membership.table.retention") },
  { accessorKey: "lastBenefitUsedAt", header: t("customer.membership.table.lastBenefit") },
  { accessorKey: "actions", header: t("customer.membership.table.actions") },
]);

const pageSizeOptions = [
  { label: "20", value: 20 },
  { label: "50", value: 50 },
  { label: "100", value: 100 },
];

const handleRefresh = async () => {
  const stopTimer = metrics.startLatencyTimer("membership_manual_refresh");
  try {
    await store.fetchMemberships();
  } finally {
    stopTimer();
  }
};

const toggleRow = (id: string) => {
  store.toggleSelection(id);
};

const formatDate = (value?: string | null) => {
  if (!value) return "—";
  return new Date(value).toLocaleString();
};

const retentionColor = (status?: string) => {
  switch (status) {
    case "warning":
      return "warning";
    case "downgrade":
      return "error";
    default:
      return "success";
  }
};

const retentionLabel = (status?: string) => {
  if (!status) return t("customer.membership.segments.safe");
  switch (status) {
    case "warning":
      return t("customer.membership.segments.warning");
    case "downgrade":
      return t("customer.membership.segments.downgrade");
    default:
      return t("customer.membership.segments.safe");
  }
};

const handleSegmentClick = async (segment: string) => {
  if (selectedSegment.value === segment) {
    await handleClearSegment();
    return;
  }
  selectedSegment.value = segment;
  metrics.recordEvent({
    name: "membership_segment_apply",
    metadata: { segment },
  });
  await store.fetchMemberships({ retentionStatus: segment, page: 1 });
};

const handleClearSegment = async () => {
  selectedSegment.value = null;
  metrics.recordEvent({
    name: "membership_segment_clear",
  });
  await store.fetchMemberships({ retentionStatus: undefined, page: 1 });
};

const openReminderFor = (id: string) => {
  store.replaceSelection([id]);
  reminderDrawerOpen.value = true;
};

const handleReminderComplete = () => {
  store.clearSelection();
};

const handleExport = async () => {
  exporting.value = true;
  const filters: CustomerListFilters = {
    tier: membershipFilters.value.tier,
    retentionStatus: membershipFilters.value.retentionStatus,
    growthRange: membershipFilters.value.growthRange,
    pointsRange: membershipFilters.value.pointsRange,
    benefitStatus: membershipFilters.value.benefitStatus,
  };
  metrics.recordEvent({
    name: "membership_export",
    metadata: { count: membershipStats.value.total },
  });
  try {
    await bulkActions.submitExport({
      filters,
      audit: {
        action: "customer.membership.export",
        resource: `customers:members:${membershipStats.value.total}`,
      },
    });
  } catch (error: any) {
    toast.add({
      title: t("customer.membership.messages.exportFailed"),
      description: error?.message,
      color: "error",
    });
  } finally {
    exporting.value = false;
  }
};

watch(
  () => membershipError.value,
  (message) => {
    if (message) {
      toast.add({
        title: t("customer.membership.messages.errorTitle"),
        description: message,
        color: "error",
      });
    }
  }
);

watch(
  () => membershipFilters.value.page,
  (next) => {
    if (typeof next === "number" && next !== page.value) {
      page.value = next;
    }
  }
);

watch(
  () => membershipFilters.value.pageSize,
  (next) => {
    if (typeof next === "number" && next !== pageSize.value) {
      pageSize.value = next;
    }
  }
);

watch(
  () => membershipFilters.value.retentionStatus,
  (next) => {
    selectedSegment.value = next || null;
  }
);

watch(page, async (value, oldValue) => {
  if (value === oldValue) return;
  await store.fetchMemberships({ page: value });
});

watch(pageSize, async (value, oldValue) => {
  if (value === oldValue) return;
  await store.fetchMemberships({ pageSize: value, page: 1 });
});

onMounted(async () => {
  if (!membershipInsights.value.length) {
    await store.fetchMemberships();
  }
});
</script>
