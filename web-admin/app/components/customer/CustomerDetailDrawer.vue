<template>
  <USlideover v-model="open" :title="t('customer.directory.drawer.title')" :ui="{ width: 'max-w-3xl' }">
    <template v-if="customer">
      <div class="space-y-6">
        <section class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-xs text-gray-500">{{ t("customer.directory.table.name") }}</p>
            <p class="text-base font-medium text-gray-900">{{ customer.name }}</p>
            <p class="text-xs text-gray-400">{{ customer.id }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t("customer.directory.table.tier") }}</p>
            <p class="text-base font-medium text-gray-900">
              {{ customer.membershipTier || t("customer.directory.drawer.unknown") }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t("customer.directory.table.contact") }}</p>
            <p class="text-sm text-gray-900">
              <span v-if="customer.email && !isMasked('email')">{{ customer.email }}</span>
              <span v-else-if="customer.email">{{ t("customer.directory.table.masked") }}</span>
            </p>
            <p class="text-sm text-gray-900">
              <span v-if="customer.phone && !isMasked('phone')">{{ customer.phone }}</span>
              <span v-else-if="customer.phone">{{ t("customer.directory.table.masked") }}</span>
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t("customer.directory.drawer.accountManager") }}</p>
            <p class="text-base font-medium text-gray-900">
              {{ customer.accountManager || t("customer.directory.drawer.unassigned") }}
            </p>
          </div>
        </section>

        <UTabs v-model="activeTab" :items="tabs">
          <template #item="{ item }">
            <div v-if="item.key === 'overview'" class="space-y-3 text-sm text-gray-600">
              <p>{{ customer.source || t("customer.directory.drawer.unknownSource") }}</p>
              <p>
                {{ t("customer.directory.drawer.createdAt") }}：
                {{
                  customer.createdAt
                    ? new Date(customer.createdAt).toLocaleString()
                    : t("customer.directory.drawer.unknown")
                }}
              </p>
            </div>
            <div v-else-if="item.key === 'orders'" class="space-y-2 text-sm text-gray-600">
              <p>
                {{ t("customer.directory.drawer.lastOrder") }}：
                {{
                  customer.lastOrderAt
                    ? new Date(customer.lastOrderAt).toLocaleString()
                    : t("customer.directory.drawer.unknown")
                }}
              </p>
              <p>{{ t("customer.directory.drawer.lastOrderAmount", { amount: customer.lastOrderAmount || 0 }) }}</p>
            </div>
            <div v-else-if="item.key === 'afterSales'" class="text-sm text-gray-500">
              {{ t("customer.directory.drawer.afterSalesPlaceholder") }}
            </div>
            <div v-else-if="item.key === 'points'" class="space-y-2 text-sm text-gray-600">
              <p>{{ t("customer.directory.drawer.growthValue", { value: customer.growthValue ?? 0 }) }}</p>
              <p>{{ t("customer.directory.drawer.points", { value: customer.points ?? 0 }) }}</p>
              <p>
                {{ t("customer.directory.drawer.retentionStatus") }}：
                {{
                  customer.membershipSnapshot?.retentionStatus ||
                  t("customer.directory.drawer.unknown")
                }}
              </p>
            </div>
            <div v-else-if="item.key === 'notes'" class="text-sm text-gray-600 whitespace-pre-line">
              {{ customer.metadata?.notes || t("customer.directory.drawer.notesPlaceholder") }}
            </div>
            <div v-else-if="item.key === 'audit'" class="text-sm text-gray-500">
              {{ t("customer.directory.drawer.auditPlaceholder", { id: customer.id }) }}
            </div>
          </template>
        </UTabs>
      </div>
    </template>
    <template v-else>
      <div class="text-center text-sm text-gray-500 py-6">
        {{ t("customer.directory.drawer.empty") }}
      </div>
    </template>
  </USlideover>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "#imports";
import type { Customer } from "~/types/customer";
import { useCustomerStore } from "~/stores/customer";

const props = defineProps<{
  customer: Customer | null;
  modelValue: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [boolean];
}>();

const open = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit("update:modelValue", value),
});

const { t } = useI18n();
const store = useCustomerStore();
const activeTab = ref("overview");

const tabs = computed(() => [
  { key: "overview", label: t("customer.directory.drawer.tabs.overview") },
  { key: "orders", label: t("customer.directory.drawer.tabs.orders") },
  { key: "afterSales", label: t("customer.directory.drawer.tabs.afterSales") },
  { key: "points", label: t("customer.directory.drawer.tabs.points") },
  { key: "notes", label: t("customer.directory.drawer.tabs.notes") },
  { key: "audit", label: t("customer.directory.drawer.tabs.audit") },
]);

const isMasked = (field: string) => (props.customer ? store.isFieldMasked(props.customer, field) : false);
</script>
