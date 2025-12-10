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

    <CustomerTable
      :can-manage="canManageCustomers"
      @view="handleViewCustomer"
      @pin="handleViewCustomer"
      @edit="handleEditCustomer"
      @delete="handleDeleteCustomer"
    />

    <CustomerDetailDrawer
      v-model="detailOpen"
      :customer="activeCustomer"
      :loading="detailLoading"
      :can-manage="canManageCustomers"
      @edit="handleEditCustomer"
      @delete="handleDeleteCustomer"
    />

    <EditCustomerModal
      v-if="canManageCustomers && editingCustomer"
      v-model:open="editModalOpen"
      :customer="editingCustomer"
      @updated="handleCustomerUpdated"
      @close="handleEditModalClose"
    />
    <CustomerDeleteConfirm
      v-if="canManageCustomers"
      v-model:open="deleteConfirmOpen"
      :customer="deleteTarget"
      @deleted="handleCustomerDeleted"
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
import EditCustomerModal from "~/components/Modals/EditCustomerModal.vue";
import CustomerDeleteConfirm from "~/components/customer/CustomerDeleteConfirm.vue";
import type { Customer } from "~/types/customer";

const store = useCustomerStore();
const { loading, error, filters } = storeToRefs(store);
const { t } = useI18n();
const toast = useToast();
const metrics = useCustomerMetrics();
const { hasPermission } = usePermissions();

const detailOpen = ref(false);
const detailLoading = ref(false);
const viewingCustomerId = ref<string | null>(null);
const activeCustomer = ref<Customer | null>(null);
const importModalOpen = ref(false);
const exportModalOpen = ref(false);
const editModalOpen = ref(false);
const deleteConfirmOpen = ref(false);
const editingCustomer = ref<Customer | null>(null);
const deleteTarget = ref<Customer | null>(null);

const canManageCustomers = computed(() => hasPermission("customer.manage"));
const canExportCustomers = computed(() => hasPermission("customer.export"));
const permissionMetadata = computed(() => ({
  canManage: canManageCustomers.value,
  canExport: canExportCustomers.value,
}));

const blurActiveElement = () => {
  if (typeof document === "undefined") return;
  const active = document.activeElement as HTMLElement | null;
  active?.blur?.();
};

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

const handleViewCustomer = async (customer: Customer) => {
  if (!customer?.id) return;
  blurActiveElement();
  activeCustomer.value = customer;
  detailOpen.value = true;
  detailLoading.value = true;
  viewingCustomerId.value = customer.id;
  try {
    const full = await store.fetchCustomerById(customer.id);
    if (viewingCustomerId.value === customer.id && full) {
      activeCustomer.value = full as Customer;
    }
  } catch (err: any) {
    toast.add({
      title: t("customer.directory.errors.toastTitle"),
      description: err?.message ?? t("customer.directory.errors.generic"),
      color: "error",
    });
  } finally {
    if (viewingCustomerId.value === customer.id) {
      detailLoading.value = false;
    }
  }
};

const handleEditCustomer = (customer: Customer) => {
  if (!customer) return;
  editingCustomer.value = customer;
  editModalOpen.value = true;
};

type EditClosePayload = { action?: string; customer?: Customer } | boolean | undefined;

const handleEditModalClose = (payload?: EditClosePayload) => {
  editModalOpen.value = false;
  if (payload && typeof payload === "object" && "customer" in payload && payload.customer) {
    activeCustomer.value = payload.customer;
  }
  editingCustomer.value = null;
};

const handleCustomerUpdated = (customer: Customer) => {
  if (customer) {
    activeCustomer.value = customer;
    editingCustomer.value = customer;
  }
};

const handleDeleteCustomer = (customer: Customer) => {
  if (!customer) return;
  deleteTarget.value = customer;
  deleteConfirmOpen.value = true;
};

const handleCustomerDeleted = (id: string) => {
  if (!id) return;
  if (activeCustomer.value?.id === id) {
    activeCustomer.value = null;
    detailOpen.value = false;
  }
  deleteConfirmOpen.value = false;
  deleteTarget.value = null;
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

watch(detailOpen, (open) => {
  if (!open) {
    detailLoading.value = false;
    viewingCustomerId.value = null;
    blurActiveElement();
  }
});

watch(editModalOpen, (open) => {
  if (!open) {
    editingCustomer.value = null;
  }
});

watch(deleteConfirmOpen, (open) => {
  if (!open) {
    deleteTarget.value = null;
  }
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
