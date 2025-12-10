<template>
  <UModal
    v-model="open"
    :ui="modalUi"
    :close="false"
    :dismissible="false"
  >
    <template #header>
      <div>
        <p class="text-base font-semibold text-gray-900">
          {{ title }}
        </p>
        <p class="text-sm text-gray-500">
          {{ t("customer.directory.export.subtitle") }}
        </p>
      </div>
    </template>

    <template #body>
      <div class="space-y-4">
        <UAlert color="gray" variant="soft">
          <template #title>
            {{ t("customer.directory.export.summaryTitle") }}
          </template>
          <template #description>
            <div class="flex flex-wrap gap-2">
              <UBadge v-for="item in summary" :key="item" size="xs" variant="subtle">
                {{ item }}
              </UBadge>
              <span v-if="summary.length === 0" class="text-xs text-gray-500">
                {{ t("customer.directory.export.noFilters") }}
              </span>
            </div>
          </template>
        </UAlert>

        <UFormField :label="t('customer.directory.export.fieldLabel')">
          <USelect
            v-model="selectedFields"
            class="w-full"
            multiple
            :items="fieldOptions"
          >
            <template #default="{ modelValue }">
              <div
                v-if="Array.isArray(modelValue) && modelValue.length"
                class="flex flex-wrap gap-2"
              >
                <UBadge
                  v-for="field in modelValue"
                  :key="field"
                  variant="subtle"
                  size="xs"
                >
                  {{ formatField(field) }}
                </UBadge>
              </div>
              <span v-else class="text-sm text-gray-400">
                {{ t("customer.directory.export.fieldPlaceholder") }}
              </span>
            </template>
          </USelect>
          <p class="mt-1 text-xs text-gray-500">
            {{ t("customer.directory.export.fieldHint") }}
          </p>
        </UFormField>

        <UAlert
          v-if="exportError"
          color="red"
          variant="soft"
          :title="t('customer.directory.export.failed')"
          :description="exportError"
        />
        <UAlert
          v-else-if="lastTaskId"
          color="primary"
          variant="soft"
          :title="t('customer.directory.export.taskCreated', { id: lastTaskId })"
          :description="t('customer.directory.export.taskDesc')"
        />
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton variant="ghost" @click="closeModal">
          {{ t("customer.directory.actions.cancel") }}
        </UButton>
        <UButton
          color="primary"
          :loading="exporting"
          @click="handleSubmit"
        >
          {{ t("customer.directory.export.submit") }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "#imports";
import { useCustomerStore } from "~/stores/customer";
import type { CustomerListFilters } from "~/types/customer";

const props = defineProps<{
  modelValue: boolean;
  filters?: Partial<CustomerListFilters>;
  context?: "directory" | "members";
}>();

const emit = defineEmits<{
  "update:modelValue": [boolean];
  submitted: [];
}>();

const store = useCustomerStore();
const { t } = useI18n();
const modalUi = {
  content: "max-w-3xl w-full",
  body: "p-4 sm:p-5",
  header: "p-4 sm:px-5",
  footer: "p-4 sm:px-5",
};

const directoryFields = [
  "id",
  "name",
  "email",
  "phone",
  "membershipTier",
  "status",
  "source",
  "createdAt",
];
const membershipFields = [
  "id",
  "name",
  "membershipTier",
  "growthValue",
  "points",
  "retentionStatus",
  "lastBenefitUsedAt",
];

const open = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit("update:modelValue", value),
});

const defaultFields = computed(() =>
  props.context === "members" ? membershipFields : directoryFields
);

const fieldOptions = computed(() =>
  defaultFields.value.map((field) => ({
    value: field,
    label: formatField(field),
  }))
);

const selectedFields = ref<string[]>([...defaultFields.value]);

const closeModal = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
  emit("update:modelValue", false);
};

watch(open, (value) => {
  if (value) {
    selectedFields.value = [...defaultFields.value];
    store.resetExportState();
  }
});

const exporting = computed(() => store.exportState.submitting);
const exportError = computed(() => store.exportState.error);
const lastTaskId = computed(() => store.exportState.lastTaskId);

const title = computed(() =>
  props.context === "members"
    ? t("customer.directory.export.membersTitle")
    : t("customer.directory.export.title")
);

const summary = computed(() => {
  const active: string[] = [];
  const filters = props.filters || {};
  const formatRange = (range?: [number, number]) => {
    if (!range) return "";
    const [min, max] = range;
    if (min && max) return `${min}~${max}`;
    if (min && !max) return `>${min}`;
    if (!min && max) return `<${max}`;
    return "";
  };
  const pushIf = (condition: any, label: string, value?: any) => {
    if (!condition) return;
    active.push(
      value ? `${label}：${value}` : `${label}`
    );
  };
  if (filters.keyword) pushIf(true, t("customer.directory.filters.keyword"), filters.keyword);
  pushIf(filters.tier, t("customer.directory.filters.tier"), filters.tier);
  pushIf(filters.source, t("customer.directory.filters.source"), filters.source);
  pushIf(filters.type, t("customer.directory.filters.type"), filters.type);
  pushIf(filters.region, t("customer.directory.filters.region"), filters.region);
  pushIf(
    filters.tags?.length,
    t("customer.directory.filters.tags"),
    filters.tags?.join(", ")
  );
  pushIf(filters.retentionStatus, t("customer.membership.filters.retentionStatus"), filters.retentionStatus);
  pushIf(filters.benefitStatus, t("customer.membership.filters.benefitStatus"), filters.benefitStatus);
  const growth = formatRange(filters.growthRange);
  pushIf(growth, t("customer.membership.filters.growthRange"), growth);
  const points = formatRange(filters.pointsRange);
  pushIf(points, t("customer.membership.filters.pointsRange"), points);
  return active;
});

function formatField(field: string) {
  const key = `customer.directory.export.fields.${field}`;
  const translated = t(key);
  if (translated === key) {
    return field;
  }
  return translated;
}

const handleSubmit = async () => {
  if (!selectedFields.value.length) {
    store.exportState.error = t("customer.directory.export.noFields");
    return;
  }
  try {
    await store.submitExportTask({
      fields: [...selectedFields.value],
      filters: props.filters as CustomerListFilters,
      context: props.context || "directory",
    });
    emit("submitted");
    closeModal();
  } catch (error) {
    console.error("[CustomerExportDialog] export failed", error);
  }
};
</script>
