<template>
  <UCard class="mb-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="text-sm text-gray-600">
        {{ t("customer.directory.bulk.selected", { count: selectionCount }) }}
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton
          icon="i-heroicons-user-plus"
          :disabled="!hasSelection"
          @click="openOwnerModal"
        >
          {{ t("customer.directory.bulk.assignOwner") }}
        </UButton>
        <UButton
          icon="i-heroicons-tag"
          variant="soft"
          :disabled="!hasSelection"
          @click="openTagsModal('add')"
        >
          {{ t("customer.directory.bulk.addTags") }}
        </UButton>
        <UButton
          icon="i-heroicons-tag-minus"
          variant="soft"
          :disabled="!hasSelection"
          @click="openTagsModal('remove')"
        >
          {{ t("customer.directory.bulk.removeTags") }}
        </UButton>
        <UButton
          icon="i-heroicons-no-symbol"
          color="error"
          variant="ghost"
          :disabled="!hasSelection"
          @click="confirmDisable"
        >
          {{ t("customer.directory.bulk.disable") }}
        </UButton>
      </div>
    </div>

    <UModal
      v-model="showOwnerModal"
      :ui="modalUi"
      :close="false"
      :dismissible="false"
    >
      <template #header>
        <div class="text-lg font-semibold">{{ t("customer.directory.bulk.assignOwner") }}</div>
      </template>
      <template #body>
        <UFormField :label="t('customer.directory.bulk.ownerInput')" :ui="inlineFieldUi">
          <UInput v-model="ownerForm.owner" class="w-full" placeholder="ops-001" />
        </UFormField>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton variant="ghost" @click="closeOwnerModal">
            {{ t("customer.directory.actions.cancel") }}
          </UButton>
          <UButton color="primary" @click="handleAssignOwner" :loading="submitting">
            {{ t("customer.directory.actions.confirm") }}
          </UButton>
        </div>
      </template>
    </UModal>

    <UModal
      v-model="showTagsModal"
      :ui="modalUi"
      :close="false"
      :dismissible="false"
    >
      <template #header>
        <div class="text-lg font-semibold">
          {{
            tagMode === "add"
              ? t("customer.directory.bulk.addTags")
              : t("customer.directory.bulk.removeTags")
          }}
        </div>
      </template>
      <template #body>
        <div class="space-y-3">
          <UFormField :label="t('customer.directory.bulk.tagsInput')" :ui="inlineFieldUi">
            <UInput v-model="tagsInput" class="w-full" @keyup.enter.prevent="appendTag" />
          </UFormField>
          <div class="flex flex-wrap gap-2">
            <UBadge
              v-for="tag in tagForm.tags"
              :key="tag"
              variant="subtle"
              class="cursor-pointer"
              @click="removeTag(tag)"
            >
              {{ tag }}
            </UBadge>
          </div>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton variant="ghost" @click="closeTagsModal">
            {{ t("customer.directory.actions.cancel") }}
          </UButton>
          <UButton color="primary" @click="handleTags" :loading="submitting">
            {{ t("customer.directory.actions.confirm") }}
          </UButton>
        </div>
      </template>
    </UModal>
  </UCard>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useI18n } from "#imports";
import { useCustomerStore } from "~/stores/customer";
import { useCustomerBulkActions } from "~/composables/useCustomerBulkActions";
import { useToastAlert } from "~/composables/useToastAlert";

const store = useCustomerStore();
const bulk = useCustomerBulkActions();
const { t } = useI18n();
const toast = useToastAlert();

const showOwnerModal = ref(false);
const showTagsModal = ref(false);
const tagMode = ref<"add" | "remove">("add");
const submitting = ref(false);

const ownerForm = reactive({ owner: "" });
const tagForm = reactive({ tags: [] as string[] });
const tagsInput = ref("");
const modalUi = {
  content: "max-w-lg w-full",
  body: "p-4 sm:p-5",
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

const selectionCount = computed(() => store.selection.length);
const hasSelection = computed(() => selectionCount.value > 0);

const openOwnerModal = () => {
  ownerForm.owner = "";
  showOwnerModal.value = true;
};

const blurActiveElement = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
};

const closeOwnerModal = () => {
  blurActiveElement();
  showOwnerModal.value = false;
};

const appendTag = () => {
  const value = tagsInput.value?.trim();
  if (!value) return;
  if (!tagForm.tags.includes(value)) {
    tagForm.tags.push(value);
  }
  tagsInput.value = "";
};

const removeTag = (tag: string) => {
  tagForm.tags = tagForm.tags.filter((item) => item !== tag);
};

const openTagsModal = (mode: "add" | "remove") => {
  tagMode.value = mode;
  tagForm.tags = [];
  tagsInput.value = "";
  showTagsModal.value = true;
};

const closeTagsModal = () => {
  blurActiveElement();
  showTagsModal.value = false;
};

const confirmDisable = async () => {
  if (!hasSelection.value) return;
  submitting.value = true;
  try {
    await bulk.submitBulkAction({
      action: "bulk-disable",
      ids: [...store.selection],
    });
    store.clearSelection();
    await store.fetchCustomers();
  } catch (error: any) {
    toast.add({
      title: t("customer.directory.bulk.failed"),
      description: error?.message ?? t("customer.directory.bulk.unknownError"),
      color: "error",
    });
  } finally {
    submitting.value = false;
  }
};

const handleAssignOwner = async () => {
  if (!ownerForm.owner?.trim()) {
    toast.add({
      title: t("customer.directory.bulk.ownerRequired"),
      color: "warning",
    });
    return;
  }
  submitting.value = true;
  try {
    await bulk.submitBulkAction({
      action: "assign-owner",
      ids: [...store.selection],
      payload: { owner: ownerForm.owner.trim() },
    });
    closeOwnerModal();
    await store.fetchCustomers();
  } catch (error: any) {
    toast.add({
      title: t("customer.directory.bulk.failed"),
      description: error?.message ?? t("customer.directory.bulk.unknownError"),
      color: "error",
    });
  } finally {
    submitting.value = false;
  }
};

const handleTags = async () => {
  if (!tagForm.tags.length) {
    toast.add({
      title: t("customer.directory.bulk.tagsRequired"),
      color: "warning",
    });
    return;
  }
  submitting.value = true;
  try {
    await bulk.submitBulkAction({
      action: tagMode.value === "add" ? "add-tags" : "remove-tags",
      ids: [...store.selection],
      payload: { tags: [...tagForm.tags] },
    });
    closeTagsModal();
    await store.fetchCustomers();
  } catch (error: any) {
    toast.add({
      title: t("customer.directory.bulk.failed"),
      description: error?.message ?? t("customer.directory.bulk.unknownError"),
      color: "error",
    });
  } finally {
    submitting.value = false;
  }
};
</script>
