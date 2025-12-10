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
          {{ t("customer.directory.import.subtitle") }}
        </p>
      </div>
    </template>

    <template #body>
      <div class="space-y-4">
        <UAlert color="gray" variant="soft">
          <template #title>
            {{ t("customer.directory.import.steps.title") }}
          </template>
          <template #description>
            <ol class="list-decimal pl-5 space-y-1 text-xs text-gray-600">
              <li>{{ t("customer.directory.import.steps.download") }}</li>
              <li>{{ t("customer.directory.import.steps.fill") }}</li>
              <li>{{ t("customer.directory.import.steps.upload") }}</li>
            </ol>
          </template>
        </UAlert>

        <div class="flex flex-wrap items-center gap-3">
          <UButton
            icon="i-heroicons-arrow-down-tray"
            variant="ghost"
            @click="downloadTemplate"
          >
            {{ t("customer.directory.import.downloadTemplate") }}
          </UButton>
          <span class="text-xs text-gray-500">
            {{ t("customer.directory.import.templateHint") }}
          </span>
        </div>

        <UFormField :label="t('customer.directory.import.uploadLabel')">
          <input
            id="customer-import-file"
            type="file"
            class="block w-full rounded border border-gray-200 bg-white px-3 py-2 text-sm"
            accept=".csv,.xlsx,.xls"
            @change="handleFileChange"
          />
          <p class="mt-1 text-xs text-gray-500">
            {{ selectedFileName || t("customer.directory.import.emptyUpload") }}
          </p>
        </UFormField>

        <UAlert
          v-if="importError"
          color="red"
          variant="soft"
          :title="t('customer.directory.import.failed')"
          :description="importError"
        />

        <UAlert
          v-else-if="lastTaskId"
          color="primary"
          variant="soft"
          :title="t('customer.directory.import.taskCreated', { id: lastTaskId })"
          :description="t('customer.directory.import.taskDesc')"
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
          :loading="importing"
          @click="handleSubmit"
        >
          {{ t("customer.directory.import.submit") }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n, useRuntimeConfig } from "#imports";
import { useCustomerStore } from "~/stores/customer";

const props = defineProps<{
  modelValue: boolean;
  context?: "directory" | "members";
}>();

const emit = defineEmits<{
  "update:modelValue": [boolean];
  submitted: [];
}>();

const config = useRuntimeConfig();
const { t } = useI18n();
const store = useCustomerStore();
const selectedFile = ref<File | null>(null);
const selectedFileName = ref("");
const modalUi = {
  content: "max-w-3xl w-full",
  body: "p-4 sm:p-5",
  header: "p-4 sm:px-5",
  footer: "p-4 sm:px-5",
};

const open = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit("update:modelValue", value),
});

const importing = computed(() => store.importState.submitting);
const importError = computed(() => store.importState.error);
const lastTaskId = computed(() => store.importState.lastTaskId);

watch(open, (value) => {
  if (!value) {
    selectedFile.value = null;
    selectedFileName.value = "";
    store.resetImportState();
  }
});

const title = computed(() =>
  props.context === "members"
    ? t("customer.directory.import.membersTitle")
    : t("customer.directory.import.title")
);

const downloadTemplate = () => {
  const baseUrl = config.public?.apiBaseUrl || "";
  const url = `${baseUrl}/customers/import/template`;
  window.open(url, "_blank");
};

const closeModal = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
  emit("update:modelValue", false);
};

const handleFileChange = (event: Event) => {
  const input = event.target as HTMLInputElement;
  const [file] = input.files || [];
  selectedFile.value = file || null;
  selectedFileName.value = file?.name || "";
};

const handleSubmit = async () => {
  if (!selectedFile.value) {
    store.importState.error = t("customer.directory.import.noFile");
    return;
  }
  try {
    await store.submitImportTask({
      file: selectedFile.value,
      context: props.context || "directory",
    });
    emit("submitted");
    closeModal();
  } catch (error) {
    console.error("[CustomerImportDialog] import failed", error);
  }
};
</script>
