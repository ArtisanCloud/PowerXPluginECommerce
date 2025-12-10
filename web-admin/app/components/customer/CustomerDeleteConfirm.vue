<template>
  <UModal
    v-model:open="isOpen"
    :title="t('customer.directory.modals.delete.title')"
    :description="t('customer.directory.modals.delete.desc')"
    :ui="{ content: 'sm:max-w-lg', footer: 'justify-end gap-3' }"
  >
    <template #body>
      <UCard class="border border-gray-200 dark:border-gray-800">
        <div class="space-y-4 p-4">
          <p class="text-sm text-gray-600">
            {{ t('customer.directory.modals.delete.target', { name: customer?.name || '—' }) }}
          </p>
          <UFormField :label="t('customer.directory.modals.delete.reason')" required>
            <UTextarea
              v-model="reason"
              :rows="4"
              :placeholder="t('customer.directory.modals.delete.placeholder')"
            />
          </UFormField>
          <p class="text-xs text-gray-500">
            {{ t('customer.directory.modals.delete.hint') }}
          </p>
        </div>
      </UCard>
    </template>
    <template #footer>
      <UButton variant="ghost" @click="handleCancel">
        {{ t('customer.directory.actions.cancel') }}
      </UButton>
      <UButton
        color="red"
        :loading="deleting"
        :disabled="!reason.trim() || deleting"
        @click="handleConfirm"
      >
        {{ t('customer.directory.modals.delete.confirm') }}
      </UButton>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "#imports";
import type { Customer } from "~/types/customer";
import { useCustomerStore } from "~/stores/customer";
import { useToastAlert } from "~/composables/useToastAlert";

const props = defineProps<{ open?: boolean; customer: Customer | null }>();
const emit = defineEmits<{
  "update:open": [boolean];
  deleted: [id: string];
}>();

const { t } = useI18n();
const toast = useToastAlert();
const store = useCustomerStore();

const isOpen = computed({
  get: () => props.open ?? false,
  set: (value: boolean) => emit("update:open", value),
});

const blurActiveElement = () => {
  if (typeof document === "undefined") return;
  const active = document.activeElement as HTMLElement | null;
  active?.blur?.();
};

const reason = ref("");
const deleting = computed(() => store.mutationState.deleting);

watch(
  () => props.customer,
  () => {
    reason.value = "";
  }
);

watch(isOpen, (value) => {
  if (!value) {
    blurActiveElement();
  }
});

const customer = computed(() => props.customer);

const handleCancel = () => {
  blurActiveElement();
  emit("update:open", false);
};

const handleConfirm = async () => {
  if (!customer.value?.id) {
    return;
  }
  try {
    await store.deleteCustomer({
      id: customer.value.id,
      reason: reason.value.trim(),
      customer: customer.value,
    });
    toast.add({
      title: t("customer.directory.modals.delete.success"),
      color: "green",
    });
    emit("deleted", customer.value.id);
    blurActiveElement();
    emit("update:open", false);
  } catch (error: any) {
    toast.add({
      title: t("customer.directory.modals.delete.failed"),
      description: error?.data?.message || error?.message,
      color: "red",
    });
  }
};
</script>
