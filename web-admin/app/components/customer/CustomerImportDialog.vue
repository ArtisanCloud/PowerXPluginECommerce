<template>
  <UModal
    v-model:open="resolveOpen"
    :title="title"
    :description="subtitle"
    :ui="modalUi"
    :close="false"
    :dismissible="false"
  >

    <template #body>
      <div class="space-y-5 p-4 sm:p-5">
        <UAlert color="gray" variant="soft">
          <template #title>
            {{ t("customer.directory.import.steps.title") }}
          </template>
          <template #description>
            <ol class="list-decimal pl-5 space-y-1 text-xs text-gray-400">
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
            :loading="downloadingTemplate"
            :disabled="downloadingTemplate"
            @click="downloadTemplate"
          >
            {{ t("customer.directory.import.downloadTemplate") }}
          </UButton>
          <span class="text-xs text-gray-400">
            {{ t("customer.directory.import.templateHint") }}
          </span>
        </div>

        <UFormField :label="t('customer.directory.import.uploadLabel')">
          <input
            id="customer-import-file"
            type="file"
            class="block w-full rounded border border-gray-200/40 dark:border-white/10 bg-transparent px-3 py-2 text-sm text-gray-100 placeholder:text-gray-500"
            accept=".csv,.xlsx,.xls"
            @change="handleFileChange"
          />
          <p class="mt-1 text-xs text-gray-400">
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
      <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end">
        <UButton color="neutral" variant="subtle" @click="closeModal">
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
import { useI18n } from "#imports";
import { useApiClient } from "~/composables/api";
import { useToastAlert } from "~/composables/useToastAlert";
import { useCustomerStore } from "~/stores/customer";

const props = defineProps<{
  modelValue?: boolean;
  open?: boolean;
  context?: "directory" | "members";
}>();

const emit = defineEmits<{
  "update:modelValue": [boolean];
  "update:open": [boolean];
  submitted: [];
}>();

const { t } = useI18n();
const store = useCustomerStore();
const { client } = useApiClient();
const toast = useToastAlert();
const selectedFile = ref<File | null>(null);
const selectedFileName = ref("");
const downloadingTemplate = ref(false);
const modalUi = {
  content: "max-w-3xl w-[min(95vw,40rem)]",
  header: "px-5 pt-5 pb-4 border-b border-gray-200/40 dark:border-white/10",
  body: "p-0",
  footer: "px-5 py-4 border-t border-gray-200/40 dark:border-white/10",
};

const resolveOpen = computed({
  get: () => props.open ?? props.modelValue ?? false,
  set: (value: boolean) => {
    emit("update:open", value);
    emit("update:modelValue", value);
  },
});

const importing = computed(() => store.importState.submitting);
const importError = computed(() => store.importState.error);
const lastTaskId = computed(() => store.importState.lastTaskId);

watch(resolveOpen, (value) => {
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

const subtitle = computed(() => t("customer.directory.import.subtitle"));

const downloadTemplate = async () => {
  if (downloadingTemplate.value) return;
  downloadingTemplate.value = true;
  try {
    const response = await client.raw("/admin/customers/import/template", {
      method: "GET",
      responseType: "blob" as const,
    });
    const blob = response?._data as Blob;
    if (!blob) {
      throw new Error("template download failed");
    }
    const disposition = response.headers?.get?.("content-disposition");
    let fallback = "customer_import_template.csv";
    if (disposition) {
      const match = disposition.match(/filename=\"?([^\";]+)\"?/i);
      if (match?.[1]) {
        fallback = decodeURIComponent(match[1]);
      }
    }
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = fallback;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch (error: any) {
    console.error("[CustomerImportDialog] download template failed", error);
    toast.add({
      title: t("customer.directory.import.failed"),
      description: error?.message ?? t("customer.directory.errors.generic"),
      color: "error",
    });
  } finally {
    downloadingTemplate.value = false;
  }
};

const closeModal = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
  resolveOpen.value = false;
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
