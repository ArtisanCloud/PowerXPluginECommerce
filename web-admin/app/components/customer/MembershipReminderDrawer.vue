<template>
  <USlideover v-model="open" :title="t('customer.membership.reminder.title')">
    <div class="space-y-6">
      <div class="text-sm text-gray-600">
        {{ t("customer.membership.reminder.selection", { count: selectedCount }) }}
      </div>

      <UFormField :label="t('customer.membership.reminder.channel')">
        <USelectMenu
          v-model="channel"
          :options="channelOptions"
          value-attribute="value"
          option-attribute="label"
        />
      </UFormField>

      <UFormField :label="t('customer.membership.reminder.template')">
        <UInput
          v-model="templateId"
          :placeholder="t('customer.membership.reminder.templatePlaceholder')"
        />
      </UFormField>

      <UFormField :label="t('customer.membership.reminder.note')">
        <UTextarea
          v-model="note"
          :placeholder="t('customer.membership.reminder.notePlaceholder')"
        />
      </UFormField>

      <div class="rounded-lg border border-gray-100 bg-gray-50 p-4">
        <p class="text-sm font-medium text-gray-700">
          {{ t("customer.membership.reminder.metricsTitle") }}
        </p>
        <div class="mt-3 grid grid-cols-2 gap-4 text-sm text-gray-600">
          <div>
            <p class="text-xs text-gray-500">
              {{ t("customer.membership.reminder.successRate") }}
            </p>
            <p class="text-xl font-semibold text-emerald-600">
              {{ successRate }}%
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">
              {{ t("customer.membership.reminder.failureRate") }}
            </p>
            <p class="text-xl font-semibold text-rose-500">
              {{ failureRate }}%
            </p>
          </div>
        </div>
        <p v-if="reminderState.error" class="mt-3 text-sm text-red-500">
          {{ reminderState.error }}
        </p>
        <p v-if="reminderState.lastTaskId" class="mt-1 text-xs text-gray-400">
          {{ t("customer.membership.reminder.lastTask", { id: reminderState.lastTaskId }) }}
        </p>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-2">
        <UButton variant="ghost" @click="emit('update:modelValue', false)">
          {{ t("customer.membership.actions.cancel") }}
        </UButton>
        <UButton
          color="primary"
          :loading="reminderState.submitting"
          :disabled="selectedCount === 0"
          @click="handleSubmit"
        >
          {{ t("customer.membership.reminder.submit") }}
        </UButton>
      </div>
    </template>
  </USlideover>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useI18n, useToast } from "#imports";
import { useCustomerStore } from "~/stores/customer";

const props = defineProps<{
  modelValue: boolean;
  selectedIds: string[];
}>();

const emit = defineEmits<{
  "update:modelValue": [boolean];
  completed: [];
}>();

const store = useCustomerStore();
const { reminderState } = storeToRefs(store);
const { t } = useI18n();
const toast = useToast();

const channel = ref(reminderState.value.channel);
const templateId = ref(reminderState.value.templateId);
const note = ref("");

const open = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit("update:modelValue", value),
});

const selectedCount = computed(() => props.selectedIds.length);

const successRate = computed(() => {
  if (!reminderState.value.total) return 0;
  return Math.round(
    (reminderState.value.success / reminderState.value.total) * 100
  );
});

const failureRate = computed(() => {
  if (!reminderState.value.total) return 0;
  return Math.max(0, 100 - successRate.value);
});

const channelOptions = computed(() => [
  { label: t("customer.membership.reminder.channelSms"), value: "sms" },
  { label: t("customer.membership.reminder.channelEmail"), value: "email" },
  { label: t("customer.membership.reminder.channelInapp"), value: "inapp" },
]);

watch(
  () => reminderState.value.channel,
  (next) => {
    if (!open.value) return;
    channel.value = next;
  }
);

watch(
  () => reminderState.value.templateId,
  (next) => {
    if (!open.value) return;
    templateId.value = next;
  }
);

const handleSubmit = async () => {
  if (!selectedCount.value) {
    toast.add({
      title: t("customer.membership.reminder.emptySelection"),
      color: "warning",
    });
    return;
  }
  if (!templateId.value?.trim()) {
    toast.add({
      title: t("customer.membership.reminder.templateRequired"),
      color: "warning",
    });
    return;
  }

  try {
    await store.triggerMembershipReminder({
      ids: [...props.selectedIds],
      channel: channel.value,
      templateId: templateId.value.trim(),
      metadata: note.value ? { note: note.value } : undefined,
    });
    note.value = "";
    emit("completed");
    emit("update:modelValue", false);
  } catch (error: any) {
    toast.add({
      title: t("customer.membership.reminder.failed"),
      description: error?.message ?? reminderState.value.error ?? "",
      color: "error",
    });
  }
};
</script>
