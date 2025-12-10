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
          icon="i-heroicons-user-plus"
          color="primary"
          @click="openCreateDialog"
        >
          {{ t("customer.directory.actions.create") }}
        </UButton>
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
      :can-manage="canManageCustomers"
      @edit="handleEditCustomer"
      @delete="handleDeleteCustomer"
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
    <CreateCustomerModal
      v-if="canManageCustomers"
      v-model:open="createModalOpen"
      @created="handleCreatedCustomer"
    />
    <EditCustomerModal
      v-if="canManageCustomers && editingCustomer"
      v-model:open="editModalOpen"
      :customer="editingCustomer as Customer"
      @updated="handleViewCustomer"
    />
    <CustomerDeleteConfirm
      v-if="canManageCustomers && deleteTarget"
      v-model:open="deleteModalOpen"
      :customer="deleteTarget"
      @deleted="handleDeleteCompleted"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useI18n } from "#imports";
import CustomerFilterBar from "~/components/customer/CustomerFilterBar.vue";
import CustomerTable from "~/components/customer/CustomerTable.vue";
import CustomerBulkActions from "~/components/customer/CustomerBulkActions.vue";
import CustomerDetailDrawer from "~/components/customer/CustomerDetailDrawer.vue";
import CustomerImportDialog from "~/components/customer/CustomerImportDialog.vue";
import CustomerExportDialog from "~/components/customer/CustomerExportDialog.vue";
import CreateCustomerModal from "~/components/Modals/CreateCustomerModal.vue";
import EditCustomerModal from "~/components/Modals/EditCustomerModal.vue";
import CustomerDeleteConfirm from "~/components/customer/CustomerDeleteConfirm.vue";
import { useCustomerStore } from "~/stores/customer";
import { usePermissions } from "~/composables/usePermissions";
import { useCustomerMetrics } from "~/composables/useCustomerMetrics";
import { useToastAlert } from "~/composables/useToastAlert";
import type { Customer } from "~/types/customer";

const store = useCustomerStore();
const { loading, error, filters } = storeToRefs(store);
const { t } = useI18n();
const toast = useToastAlert();
const metrics = useCustomerMetrics();
const { hasPermission } = usePermissions();

const detailOpen = ref(false);
const activeCustomer = ref<Customer | null>(null);
const importModalOpen = ref(false);
const exportModalOpen = ref(false);
const createModalOpen = ref(false);
const editModalOpen = ref(false);
const deleteModalOpen = ref(false);
const editingCustomer = ref<Customer | null>(null);
const deleteTarget = ref<Customer | null>(null);

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

const openCreateDialog = () => {
	createModalOpen.value = true;
};

const handleViewCustomer = (customer: Customer) => {
  activeCustomer.value = customer;
  detailOpen.value = true;
};

const handleCreatedCustomer = async (customer: Customer) => {
  toast.add({
    title: t("customer.directory.modals.create.success"),
    color: "green",
  });
  await handleRefresh().catch(() => undefined);
  handleViewCustomer(customer);
};

const handleEditCustomer = (customer: Customer) => {
	if (!canManageCustomers.value) return;
	editingCustomer.value = customer;
	editModalOpen.value = true;
};

const handleDeleteCustomer = (customer: Customer) => {
	if (!canManageCustomers.value) return;
	deleteTarget.value = customer;
	deleteModalOpen.value = true;
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

watch(editModalOpen, (open) => {
  if (!open) {
    editingCustomer.value = null;
  }
});

watch(deleteModalOpen, (open) => {
  if (!open) {
    deleteTarget.value = null;
  }
});

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

const handleDeleteCompleted = (id: string) => {
  deleteModalOpen.value = false;
  if (activeCustomer.value?.id === id) {
    activeCustomer.value = null;
    detailOpen.value = false;
  }
};
</script>
