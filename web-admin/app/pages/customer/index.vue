<template>
  <div class="space-y-4">
    <header class="flex flex-wrap gap-3 items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">
          {{ t("customer.directory.title") }}
        </h1>
        <p class="text-sm text-gray-500">
          {{ t("customer.directory.subtitle") }}
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton
          v-if="canManageCustomers"
          icon="i-heroicons-arrow-up-tray"
          color="primary"
          @click="openImportDialog"
        >
          {{ t("customer.directory.actions.import") }}
        </UButton>
        <UButton
          v-if="canExportCustomers"
          icon="i-heroicons-arrow-down-tray"
          variant="soft"
          @click="openExportDialog"
        >
          {{ t("customer.directory.actions.export") }}
        </UButton>
        <UButton icon="i-heroicons-arrow-path" variant="ghost" @click="handleRefresh" :loading="loading">
          {{ t("customer.directory.actions.refresh") }}
        </UButton>
      </div>
    </header>

    <CustomerFilterBar />
    <CustomerBulkActions />

    <UAlert v-if="error" color="red" :title="t('customer.directory.errors.title')" :description="error">
      <template #actions>
        <UButton size="xs" color="red" variant="solid" @click="handleRefresh">
          {{ t("customer.directory.actions.retry") }}
        </UButton>
      </template>
    </UAlert>

    <CustomerTable @view="handleViewCustomer" @pin="handleViewCustomer" />

    <CustomerDetailDrawer
      v-model="detailOpen"
      :customer="activeCustomer"
    />

    <CustomerImportDialog
      v-if="canManageCustomers"
      v-model="importModalOpen"
      context="directory"
      @submitted="handleImportSubmitted"
    />
    <CustomerExportDialog
      v-if="canExportCustomers"
      v-model="exportModalOpen"
      :filters="filters"
      context="directory"
      @submitted="handleExportSubmitted"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useI18n, useToast } from "#imports";
import CustomerFilterBar from "~/components/customer/CustomerFilterBar.vue";
import CustomerTable from "~/components/customer/CustomerTable.vue";
import CustomerBulkActions from "~/components/customer/CustomerBulkActions.vue";
import CustomerDetailDrawer from "~/components/customer/CustomerDetailDrawer.vue";
import CustomerImportDialog from "~/components/customer/CustomerImportDialog.vue";
import CustomerExportDialog from "~/components/customer/CustomerExportDialog.vue";
import { useCustomerStore } from "~/stores/customer";
import { usePermissions } from "~/composables/usePermissions";
import { useCustomerMetrics } from "~/composables/useCustomerMetrics";
import type { Customer } from "~/types/customer";

const store = useCustomerStore();
const { loading, error, filters } = storeToRefs(store);
const { t } = useI18n();
const toast = useToast();
const metrics = useCustomerMetrics();
const { hasPermission } = usePermissions();

const detailOpen = ref(false);
const activeCustomer = ref<Customer | null>(null);
const importModalOpen = ref(false);
const exportModalOpen = ref(false);

const canManageCustomers = computed(() => hasPermission("customer.manage"));
const canExportCustomers = computed(() => hasPermission("customer.export"));
const permissionMetadata = computed(() => ({
  canManage: canManageCustomers.value,
  canExport: canExportCustomers.value,
}));

const handleRefresh = async () => {
  try {
    await store.fetchCustomers();
  } catch (err: any) {
    toast.add({
      title: t("customer.directory.errors.toastTitle"),
      description: err?.message ?? t("customer.directory.errors.generic"),
      color: "error",
    });
  }
};

const handleViewCustomer = (customer: Customer) => {
  activeCustomer.value = customer;
  detailOpen.value = true;
};

watch(
  () => error.value,
  (message) => {
    if (message) {
      toast.add({
        title: t("customer.directory.errors.toastTitle"),
        description: message,
        color: "error",
      });
    }
  }
);

onMounted(async () => {
  store.loadSavedViews();
  await handleRefresh();
});

const openImportDialog = () => {
  metrics.recordEvent({
    name: "customer_import_modal_open",
    metadata: { source: "directory", ...permissionMetadata.value },
  });
  importModalOpen.value = true;
};

const openExportDialog = () => {
  metrics.recordEvent({
    name: "customer_export_modal_open",
    metadata: { source: "directory", ...permissionMetadata.value },
  });
  exportModalOpen.value = true;
};

const handleImportSubmitted = () => {
  metrics.recordEvent({
    name: "customer_import_submitted",
    metadata: { source: "directory" },
  });
};

const handleExportSubmitted = () => {
  metrics.recordEvent({
    name: "customer_export_submitted",
    metadata: { source: "directory" },
  });
};
</script>
