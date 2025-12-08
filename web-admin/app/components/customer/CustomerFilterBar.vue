<template>
  <UCard class="mb-4">
    <div class="flex flex-col gap-4">
      <div class="flex flex-wrap gap-4">
        <UFormField :label="t('customer.directory.filters.keyword')" class="flex-1 min-w-[220px]">
          <UInput
            v-model="form.keyword"
            :placeholder="t('customer.directory.filters.keywordPlaceholder')"
            icon="i-heroicons-magnifying-glass"
            clearable
          />
        </UFormField>
        <UFormField :label="t('customer.directory.filters.tier')" class="w-48">
          <USelectMenu
            v-model="form.tier"
            :options="tierOptions"
            value-attribute="value"
            option-attribute="label"
            clearable
          />
        </UFormField>
        <UFormField :label="t('customer.directory.filters.type')" class="w-48">
          <USelectMenu
            v-model="form.type"
            :options="typeOptions"
            value-attribute="value"
            option-attribute="label"
            clearable
          />
        </UFormField>
        <UFormField :label="t('customer.directory.filters.risk')" class="w-48">
          <USelectMenu
            v-model="form.riskLevel"
            :options="riskOptions"
            value-attribute="value"
            option-attribute="label"
            clearable
          />
        </UFormField>
      </div>
      <div class="flex flex-wrap gap-4">
        <UFormField :label="t('customer.directory.filters.source')" class="w-48">
          <USelectMenu
            v-model="form.source"
            :options="sourceOptions"
            value-attribute="value"
            option-attribute="label"
            clearable
          />
        </UFormField>
        <UFormField :label="t('customer.directory.filters.region')" class="w-48">
          <UInput v-model="form.region" clearable />
        </UFormField>
        <UFormField :label="t('customer.directory.filters.tags')" class="flex-1 min-w-[220px]">
          <UInput
            v-model="tagsInput"
            :placeholder="t('customer.directory.filters.tagsPlaceholder')"
            @keyup.enter.prevent="appendTag"
          />
          <div class="flex flex-wrap gap-2 mt-2">
            <UBadge
              v-for="tag in form.tags"
              :key="tag"
              variant="subtle"
              class="cursor-pointer"
              @click="removeTag(tag)"
            >
              {{ tag }}
            </UBadge>
          </div>
        </UFormField>
        <div class="flex-1 flex items-end justify-end gap-2">
          <UButton :label="t('customer.directory.actions.reset')" variant="ghost" @click="handleReset" />
          <UButton :label="t('customer.directory.actions.apply')" color="primary" @click="handleSubmit" />
        </div>
      </div>
      <div class="flex flex-wrap items-center justify-between">
        <div class="flex items-center gap-3">
          <UButton
            icon="i-heroicons-bookmark"
            variant="soft"
            :label="t('customer.directory.savedViews.save')"
            @click="openSaveModal"
          />
          <USelectMenu
            v-model="selectedViewId"
            :options="viewOptions"
            value-attribute="value"
            option-attribute="label"
            placeholder="Select view"
            :popper="{ placement: 'bottom-start' }"
            @update:model-value="handleApplyView"
            class="min-w-[220px]"
          />
          <UButton
            v-if="selectedViewId"
            icon="i-heroicons-trash"
            size="sm"
            variant="ghost"
            :title="t('customer.directory.savedViews.delete')"
            @click="handleDeleteView"
          />
        </div>
        <div class="text-sm text-gray-500">
          {{ t("customer.directory.status.lastSynced") }}：
          <span>{{ lastFetchedAtDisplay || t("customer.directory.status.never") }}</span>
        </div>
      </div>
    </div>

    <UModal v-model="showSaveModal">
      <UCard :ui="{ body: 'space-y-4' }">
        <template #header>
          <div class="text-lg font-semibold">
            {{ t("customer.directory.savedViews.title") }}
          </div>
        </template>
        <UFormField :label="t('customer.directory.savedViews.name')">
          <UInput v-model="newViewName" autofocus />
        </UFormField>
        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="showSaveModal = false">
              {{ t("customer.directory.actions.cancel") }}
            </UButton>
            <UButton color="primary" @click="handleSaveView">
              {{ t("customer.directory.actions.save") }}
            </UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </UCard>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useI18n } from "#imports";
import { storeToRefs } from "pinia";
import { useCustomerStore } from "~/stores/customer";

const store = useCustomerStore();
const { filters, savedViews, lastFetchedAt, activeViewId } = storeToRefs(store);
const { t } = useI18n();

const form = reactive({ ...filters.value });
const tagsInput = ref("");
const showSaveModal = ref(false);
const newViewName = ref("");
const selectedViewId = ref<string | null>(null);

watch(
  () => filters.value,
  (next) => {
    Object.assign(form, next);
  }
);

watch(
  () => activeViewId.value,
  (id) => {
    selectedViewId.value = id;
  },
  { immediate: true }
);

const tierOptions = computed(() => [
  { label: t("customer.directory.filters.tierOptions.gold"), value: "gold" },
  { label: t("customer.directory.filters.tierOptions.silver"), value: "silver" },
  { label: t("customer.directory.filters.tierOptions.platinum"), value: "platinum" },
]);

const typeOptions = computed(() => [
  { label: t("customer.directory.filters.typeOptions.individual"), value: "individual" },
  { label: t("customer.directory.filters.typeOptions.enterprise"), value: "enterprise" },
]);

const riskOptions = computed(() => [
  { label: t("customer.directory.filters.riskOptions.low"), value: "low" },
  { label: t("customer.directory.filters.riskOptions.medium"), value: "medium" },
  { label: t("customer.directory.filters.riskOptions.high"), value: "high" },
]);

const sourceOptions = computed(() => [
  { label: t("customer.directory.filters.sourceOptions.website"), value: "website" },
  { label: t("customer.directory.filters.sourceOptions.offline"), value: "offline" },
  { label: t("customer.directory.filters.sourceOptions.referral"), value: "referral" },
  { label: t("customer.directory.filters.sourceOptions.miniapp"), value: "miniapp" },
]);

const viewOptions = computed(() =>
  savedViews.value.map((view) => ({
    label: view.name,
    value: view.id,
  }))
);

const lastFetchedAtDisplay = computed(() => {
  if (!lastFetchedAt.value) return "";
  return new Date(lastFetchedAt.value).toLocaleString();
});

const appendTag = () => {
  const value = tagsInput.value?.trim();
  if (!value) return;
  if (!Array.isArray(form.tags)) {
    form.tags = [];
  }
  if (!form.tags.includes(value)) {
    form.tags.push(value);
  }
  tagsInput.value = "";
};

const removeTag = (tag: string) => {
  if (!Array.isArray(form.tags)) return;
  form.tags = form.tags.filter((item) => item !== tag);
};

const handleSubmit = async () => {
  store.setFilters({ ...form, page: 1 });
  await store.fetchCustomers();
};

const handleReset = async () => {
  store.resetFilters();
  Object.assign(form, store.filters);
  await store.fetchCustomers();
};

const openSaveModal = () => {
  newViewName.value = "";
  showSaveModal.value = true;
};

const handleSaveView = () => {
  try {
    store.saveCurrentView(newViewName.value);
    showSaveModal.value = false;
  } catch (error) {
    // rely on UI toast from caller
    console.error(error);
  }
};

const handleApplyView = (id: string | null) => {
  if (!id) return;
  store.applySavedView(id);
  Object.assign(form, store.filters);
  store.fetchCustomers();
};

const handleDeleteView = () => {
  if (!selectedViewId.value) return;
  store.deleteSavedView(selectedViewId.value);
  selectedViewId.value = null;
};
</script>
