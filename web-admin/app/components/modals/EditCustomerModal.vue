<template>
  <UModal
    v-model:open="isOpen"
    :title="t('customer.directory.modals.edit.title')"
    :description="t('customer.directory.modals.edit.description')"
    :ui="{ content: 'max-w-4xl w-[min(95vw,48rem)]', body: 'p-0', footer: 'px-4 py-3' }"
    :close="{ onClick: () => close(false) }"
  >
    <template #body>
      <form id="edit-customer-form" class="p-4 sm:p-5 space-y-6" @submit.prevent="saveCustomer">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <UFormField :label="t('customer.directory.form.name')" required>
            <UInput v-model="form.name" :placeholder="t('customer.directory.form.namePlaceholder')" />
          </UFormField>
          <UFormField :label="t('customer.directory.form.type')" required>
            <USelect
              v-model="form.customerType"
              :items="customerTypeItems"
              class="w-full"
              :placeholder="t('customer.directory.form.typePlaceholder')"
              @update:model-value="handleSelectChange"
            />
          </UFormField>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <UFormField :label="t('customer.directory.form.phone')" required>
            <UInput v-model="form.phone" type="tel" :placeholder="t('customer.directory.form.phonePlaceholder')" />
          </UFormField>
          <UFormField :label="t('customer.directory.form.email')">
            <UInput v-model="form.email" type="email" :placeholder="t('customer.directory.form.emailPlaceholder')" />
          </UFormField>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <UFormField :label="t('customer.directory.form.gender')">
            <USelect
              v-model="form.gender"
              :items="genderItems"
              class="w-full"
              :placeholder="t('customer.directory.form.genderPlaceholder')"
              @update:model-value="handleSelectChange"
            />
          </UFormField>
          <UFormField :label="t('customer.directory.form.birthDate')">
            <UInput v-model="form.birthDate" type="date" :placeholder="t('customer.directory.form.birthPlaceholder')" />
          </UFormField>
        </div>

        <UFormField :label="t('customer.directory.form.address')">
          <UTextarea v-model="form.address" :rows="3" :placeholder="t('customer.directory.form.addressPlaceholder')" />
        </UFormField>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <UFormField :label="t('customer.directory.form.tier')">
            <USelect
              v-model="form.membershipTier"
              :items="membershipTierItems"
              class="w-full"
              :placeholder="t('customer.directory.form.tierPlaceholder')"
              @update:model-value="handleSelectChange"
            />
          </UFormField>
          <UFormField :label="t('customer.directory.form.status')">
            <USelect
              v-model="form.status"
              :items="statusItems"
              class="w-full"
              :placeholder="t('customer.directory.form.statusPlaceholder')"
              @update:model-value="handleSelectChange"
            />
          </UFormField>
        </div>

        <UFormField :label="t('customer.directory.form.tags')">
          <div class="flex flex-wrap gap-2 mb-2">
            <UBadge
              v-for="(tag, i) in form.tags"
              :key="`${tag}-${i}`"
              variant="soft"
              class="cursor-pointer"
              @click="removeTag(i)"
            >
              {{ tag }}
              <UIcon name="i-heroicons-x-mark" class="w-3 h-3 ml-1" />
            </UBadge>
          </div>
          <div class="flex gap-2">
            <UInput
              v-model="newTag"
              size="sm"
              :placeholder="t('customer.directory.form.tagsPlaceholder')"
              @keyup.enter.prevent="addTag"
            />
            <UButton
              size="sm"
              variant="outline"
              :disabled="!newTag.trim()"
              @click="addTag"
            >
              {{ t("customer.directory.form.addTag") }}
            </UButton>
          </div>
        </UFormField>

        <UFormField :label="t('customer.directory.form.notes')">
          <UTextarea v-model="form.notes" :rows="3" :placeholder="t('customer.directory.form.notesPlaceholder')" />
        </UFormField>

        <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 text-sm">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <span class="text-gray-500 dark:text-gray-400">{{ t("customer.directory.form.registeredAt") }}：</span>
              <span class="text-gray-900 dark:text-white ml-2">{{ (customer as any)?.registrationDate || "—" }}</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-gray-500 dark:text-gray-400">{{ t("customer.directory.form.currentStatus") }}：</span>
              <UBadge
                :color="customer.status === 'active' ? 'success' : customer.status === 'blocked' ? 'error' : 'neutral'"
                variant="soft"
              >
                {{ statusLabel(customer.status) }}
              </UBadge>
            </div>
            <div>
              <span class="text-gray-500 dark:text-gray-400">{{ t("customer.directory.form.totalOrders") }}：</span>
              <span class="text-gray-900 dark:text-white ml-2">{{ (customer as any)?.totalOrders ?? 0 }}</span>
            </div>
            <div>
              <span class="text-gray-500 dark:text-gray-400">{{ t("customer.directory.form.totalSpent") }}：</span>
              <span class="text-gray-900 dark:text-white ml-2">
                ¥{{ ((customer as any)?.totalSpent || 0).toLocaleString() }}
              </span>
            </div>
          </div>
        </div>
      </form>
    </template>

    <template #footer>
      <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end">
        <UButton color="neutral" variant="subtle" :disabled="saving" @click="close(false)">
          {{ t("customer.directory.actions.cancel") }}
        </UButton>
        <UButton
          type="submit"
          form="edit-customer-form"
          color="primary"
          :loading="saving"
          :disabled="!isValid || saving"
        >
          {{ t("customer.directory.modals.edit.submit") }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useI18n } from "#imports";
import type { Customer, CustomerStatus } from "~/types/customer";
import { useCustomerStore } from "~/stores/customer";
import { useToastAlert } from "~/composables/useToastAlert";

const props = defineProps<{ open?: boolean; customer: Customer }>();
const emit = defineEmits<{
  "update:open": [boolean];
  updated: [customer: Customer];
  close: [payload: any];
}>();

const { t } = useI18n();
const toast = useToastAlert();
const store = useCustomerStore();
const { mutationState } = storeToRefs(store);

const isOpen = computed({
  get: () => props.open ?? true,
  set: (value: boolean) => emit("update:open", value),
});

const blurActiveElement = () => {
  if (typeof document === "undefined") return;
  const active = document.activeElement as HTMLElement | null;
  active?.blur?.();
};

function close(payload: any) {
  blurActiveElement();
  emit("close", payload);
  emit("update:open", false);
}

const form = ref({
  name: "",
  email: "",
  phone: "",
  customerType: "individual",
  gender: "",
  birthDate: "",
  address: "",
  membershipTier: "",
  status: "active" as CustomerStatus | "",
  tags: [] as string[],
  notes: "",
});

const newTag = ref("");
const saving = computed(() => mutationState.value.updating);

const customerTypeItems = [
  { value: "individual", label: "个人客户" },
  { value: "enterprise", label: "企业客户" },
];

const genderItems = [
  { value: "male", label: "男" },
  { value: "female", label: "女" },
  { value: "other", label: "其他" },
];

const membershipTierItems = [
  { value: "bronze", label: "青铜会员" },
  { value: "silver", label: "白银会员" },
  { value: "gold", label: "黄金会员" },
  { value: "platinum", label: "铂金会员" },
  { value: "diamond", label: "钻石会员" },
];

const statusItems = [
  { value: "active", label: "活跃" },
  { value: "inactive", label: "非活跃" },
  { value: "blocked", label: "受限" },
];

const handleSelectChange = () => blurActiveElement();

const statusLabel = (status?: string) => {
  switch (status) {
    case "active":
      return t("customer.directory.table.statusActive");
    case "blocked":
      return t("customer.directory.table.statusBlocked");
    default:
      return t("customer.directory.table.statusInactive");
  }
};

const isValid = computed(
  () =>
    form.value.name?.trim() &&
    form.value.phone?.trim() &&
    form.value.customerType &&
    form.value.membershipTier &&
    form.value.status,
);

const hydrate = () => {
  const c = props.customer;
  form.value = {
    name: c?.name || "",
    email: c?.email || "",
    phone: c?.phone || "",
    customerType: c?.type || "individual",
    gender: (c as any)?.gender || "",
    birthDate: (c as any)?.birthDate || "",
    address: c?.region || "",
    membershipTier: c?.membershipTier || "",
    status: (c?.status as CustomerStatus) || "active",
    tags: Array.isArray(c?.tags) ? [...(c?.tags as string[])] : [],
    notes: c?.notes || (c?.metadata?.notes as string) || "",
  };
  newTag.value = "";
};

watch(
  () => props.customer,
  () => {
    if (props.customer) {
      hydrate();
    }
  },
  { immediate: true },
);

watch(isOpen, (value) => {
  if (value) {
    hydrate();
  } else {
    blurActiveElement();
  }
});

const addTag = () => {
  const tag = newTag.value.trim();
  if (tag && !form.value.tags.includes(tag)) {
    form.value.tags.push(tag);
    newTag.value = "";
  }
};

const removeTag = (index: number) => form.value.tags.splice(index, 1);

const saveCustomer = async () => {
  if (!props.customer || !isValid.value) {
    toast.add({
      title: t("customer.directory.modals.edit.failed"),
      description: t("customer.directory.modals.edit.missing"),
      color: "red",
    });
    return;
  }
  try {
    const updated = await store.updateCustomer(props.customer.id, {
      name: form.value.name.trim(),
      email: form.value.email.trim() || undefined,
      phone: form.value.phone.trim(),
      type: form.value.customerType as any,
      membershipTier: form.value.membershipTier,
      status: form.value.status as CustomerStatus,
      tags: form.value.tags,
      notes: form.value.notes.trim() || undefined,
      source: props.customer.source,
      region: form.value.address || undefined,
    });
    toast.add({
      title: t("customer.directory.modals.edit.success"),
      color: "green",
    });
    emit("updated", updated);
    close({ action: "update", customer: updated });
  } catch (error: any) {
    toast.add({
      title: t("customer.directory.modals.edit.failed"),
      description: error?.data?.message || error?.message,
      color: "red",
    });
  }
};
</script>
