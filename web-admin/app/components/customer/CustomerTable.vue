<template>
  <UCard>
    <div class="flex items-center justify-between mb-3">
      <div class="text-sm text-gray-500">
        {{ t("customer.directory.status.total", { total }) }}
      </div>
      <USelectMenu
        v-model="pageSize"
        :options="pageSizeOptions"
        value-attribute="value"
        option-attribute="label"
        class="w-32"
      />
    </div>
    <UTable
      :rows="rows"
      :columns="columns"
      :loading="loading"
      :page="{ page: filters.page, pageCount: pageCount }"
      @update:page="handlePageChange"
    >
      <template #select-cell="{ row }">
        <UCheckbox
          :model-value="selection.has(row.id)"
          @change="() => store.toggleSelection(row.id)"
        />
      </template>
      <template #name-cell="{ row }">
        <div class="flex flex-col">
          <span class="font-medium text-gray-900">{{ row.name }}</span>
          <span class="text-xs text-gray-500">{{ row.id }}</span>
        </div>
      </template>
      <template #contact-cell="{ row }">
        <div class="text-sm text-gray-700">
          <span v-if="row.email && !isMasked(row, 'email')">{{ row.email }}</span>
          <span v-else-if="row.email">{{ t("customer.directory.table.masked") }}</span>
          <span v-if="row.phone && !isMasked(row, 'phone')" class="block">{{ row.phone }}</span>
          <span v-else-if="row.phone" class="block">{{ t("customer.directory.table.masked") }}</span>
        </div>
      </template>
      <template #tags-cell="{ row }">
        <div class="flex flex-wrap gap-1">
          <UBadge
            v-for="tag in row.tags"
            :key="`${row.id}-${tag}`"
            size="xs"
            variant="subtle"
          >
            {{ tag }}
          </UBadge>
        </div>
      </template>
      <template #status-cell="{ row }">
        <UBadge
          :color="row.status === 'active' ? 'success' : row.status === 'blocked' ? 'error' : 'neutral'"
          variant="subtle"
        >
          {{ statusLabel(row.status) }}
        </UBadge>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex gap-2">
          <UTooltip :text="t('customer.directory.actions.view')" :popper="{ placement: 'top' }">
            <UButton
              icon="i-heroicons-eye"
              variant="ghost"
              size="sm"
              @click="$emit('view', row)"
            />
          </UTooltip>
          <UTooltip :text="t('customer.directory.actions.pin')" :popper="{ placement: 'top' }">
            <UButton
              icon="i-heroicons-clipboard-document-list"
              variant="ghost"
              size="sm"
              @click="$emit('pin', row)"
            />
          </UTooltip>
        </div>
      </template>
    </UTable>
    <div class="flex items-center justify-between mt-4">
      <div class="text-sm text-gray-500">
        {{ t("customer.directory.status.selected", { count: selection.size }) }}
      </div>
      <UPagination
        v-model="page"
        :total="total"
        :page-count="filters.pageSize"
        show-first
        show-last
      />
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useI18n } from "#imports";
import type { TableColumn } from "@nuxt/ui";
import { useCustomerStore } from "~/stores/customer";
import type { Customer } from "~/types/customer";

const emit = defineEmits<{
  view: [Customer];
  pin: [Customer];
}>();

const store = useCustomerStore();
const { list, filters, loading, meta } = storeToRefs(store);
const { t } = useI18n();

const page = ref(filters.value.page || 1);
const pageSize = ref(filters.value.pageSize || 20);

watch(page, (value) => {
  store.setPage(value);
  store.fetchCustomers();
});

watch(pageSize, (value) => {
  store.setPageSize(value);
  store.fetchCustomers();
});

const selection = computed(() => new Set(store.selection));
const total = computed(() => meta.value.total || 0);
const pageCount = computed(() =>
  filters.value.pageSize ? Math.ceil(total.value / filters.value.pageSize) : 1
);

const pageSizeOptions = [
  { label: "20", value: 20 },
  { label: "50", value: 50 },
  { label: "100", value: 100 },
];

const rows = computed(() => list.value || []);

const columns = computed<TableColumn<Customer>[]>(() => [
  { accessorKey: "select", header: "", sortable: false, size: 48 },
  { accessorKey: "name", header: t("customer.directory.table.name") },
  { accessorKey: "contact", header: t("customer.directory.table.contact") },
  { accessorKey: "membershipTier", header: t("customer.directory.table.tier") },
  { accessorKey: "source", header: t("customer.directory.table.source") },
  { accessorKey: "tags", header: t("customer.directory.table.tags") },
  { accessorKey: "lastOrderAt", header: t("customer.directory.table.lastOrder") },
  { accessorKey: "status", header: t("customer.directory.table.status") },
  { id: "actions", header: t("customer.directory.table.actions") },
]);

const isMasked = (customer: Customer, field: string) =>
  store.isFieldMasked(customer, field);

const statusLabel = (status?: string) => {
  switch (status) {
    case "active":
      return t("customer.directory.table.statusActive");
    case "blocked":
      return t("customer.directory.table.statusBlocked");
    default:
      return t("customer.directory.table.statusInactive");
  }
};

const handlePageChange = (value: number) => {
  page.value = value || 1;
};
watch(
  () => filters.value.page,
  (value) => {
    if (value && value !== page.value) {
      page.value = value;
    }
  }
);

watch(
  () => filters.value.pageSize,
  (value) => {
    if (value && value !== pageSize.value) {
      pageSize.value = value;
    }
  }
);
</script>
