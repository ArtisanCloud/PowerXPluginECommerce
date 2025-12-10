<template>
  <USlideover
    v-model="open"
    :title="t('customer.membership.reminder.title')"
    :ui="drawerUi"
    :close="false"
    :dismissible="false"
  >
    <template #body>
      <div class="space-y-6">
        <div class="text-sm text-gray-600">
          {{ t("customer.membership.reminder.selection", { count: selectedCount }) }}
        </div>

        <UFormField :label="t('customer.membership.reminder.channel')" :ui="inlineFieldUi">
          <USelect
            v-model="channel"
            class="w-full"
            :items="channelOptions"
          />
        </UFormField>

        <UFormField :label="t('customer.membership.reminder.template')" :ui="inlineFieldUi">
          <UInput
            v-model="templateId"
            class="w-full"
            :placeholder="t('customer.membership.reminder.templatePlaceholder')"
          />
        </UFormField>

        <UFormField :label="t('customer.membership.reminder.note')" :ui="stackedFieldUi">
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
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton variant="ghost" @click="closeDrawer">
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
import { useI18n } from "#imports";
import { useCustomerStore } from "~/stores/customer";
import { useToastAlert } from "~/composables/useToastAlert";

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
const toast = useToastAlert();
const drawerUi = {
  content: "max-w-xl w-full",
  body: "p-4 sm:p-6",
  header: "p-4 sm:px-5",
  footer: "p-4 sm:px-5",
};
const inlineFieldUi = {
  root: "flex items-center gap-3 w-full",
  wrapper: "w-28 sm:w-32 shrink-0",
  labelWrapper: "flex items-center gap-2",
  label: "text-sm font-medium text-gray-600 whitespace-nowrap",
  container: "mt-0 flex-1",
};
const stackedFieldUi = {
  ...inlineFieldUi,
  container: "mt-0 flex-1 flex flex-col gap-2",
};

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
    closeDrawer();
  } catch (error: any) {
    toast.add({
      title: t("customer.membership.reminder.failed"),
      description: error?.message ?? reminderState.value.error ?? "",
      color: "error",
    });
  }
};

const closeDrawer = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
  emit("update:modelValue", false);
};
</script>
