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

    <UModal v-model="showOwnerModal">
      <UCard :ui="{ body: 'space-y-4' }">
        <template #header>
          <div class="text-lg font-semibold">{{ t("customer.directory.bulk.assignOwner") }}</div>
        </template>
        <UFormField :label="t('customer.directory.bulk.ownerInput')">
          <UInput v-model="ownerForm.owner" placeholder="ops-001" />
        </UFormField>
        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="showOwnerModal = false">
              {{ t("customer.directory.actions.cancel") }}
            </UButton>
            <UButton color="primary" @click="handleAssignOwner" :loading="submitting">
              {{ t("customer.directory.actions.confirm") }}
            </UButton>
          </div>
        </template>
      </UCard>
    </UModal>

    <UModal v-model="showTagsModal">
      <UCard :ui="{ body: 'space-y-4' }">
        <template #header>
          <div class="text-lg font-semibold">
            {{
              tagMode === "add"
                ? t("customer.directory.bulk.addTags")
                : t("customer.directory.bulk.removeTags")
            }}
          </div>
        </template>
        <UFormField :label="t('customer.directory.bulk.tagsInput')">
          <UInput v-model="tagsInput" @keyup.enter.prevent="appendTag" />
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
        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="showTagsModal = false">
              {{ t("customer.directory.actions.cancel") }}
            </UButton>
            <UButton color="primary" @click="handleTags" :loading="submitting">
              {{ t("customer.directory.actions.confirm") }}
            </UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </UCard>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useI18n, useToast } from "#imports";
import { useCustomerStore } from "~/stores/customer";
import { useCustomerBulkActions } from "~/composables/useCustomerBulkActions";

const store = useCustomerStore();
const bulk = useCustomerBulkActions();
const { t } = useI18n();
const toast = useToast();

const showOwnerModal = ref(false);
const showTagsModal = ref(false);
const tagMode = ref<"add" | "remove">("add");
const submitting = ref(false);

const ownerForm = reactive({ owner: "" });
const tagForm = reactive({ tags: [] as string[] });
const tagsInput = ref("");

const selectionCount = computed(() => store.selection.length);
const hasSelection = computed(() => selectionCount.value > 0);

const openOwnerModal = () => {
  ownerForm.owner = "";
  showOwnerModal.value = true;
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

const confirmDisable = async () => {
  if (!hasSelection.value) return;
  submitting.value = true;
  try {
    await bulk.submitBulkAction({
      action: "bulk-disable",
      ids: [...store.selection],
      audit: {
        action: "customer.bulk.disable",
        resource: `customers:${store.selection.length}`,
      },
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
    showOwnerModal.value = false;
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
    showTagsModal.value = false;
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
