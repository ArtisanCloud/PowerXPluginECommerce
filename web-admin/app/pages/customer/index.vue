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
      <div class="flex gap-2">
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
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useI18n, useToast } from "#imports";
import CustomerFilterBar from "~/components/customer/CustomerFilterBar.vue";
import CustomerTable from "~/components/customer/CustomerTable.vue";
import CustomerBulkActions from "~/components/customer/CustomerBulkActions.vue";
import CustomerDetailDrawer from "~/components/customer/CustomerDetailDrawer.vue";
import { useCustomerStore } from "~/stores/customer";
import type { Customer } from "~/types/customer";

const store = useCustomerStore();
const { loading, error } = storeToRefs(store);
const { t } = useI18n();
const toast = useToast();

const detailOpen = ref(false);
const activeCustomer = ref<Customer | null>(null);

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
</script>
