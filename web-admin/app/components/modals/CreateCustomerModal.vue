<template>
  <UModal
    v-model:open="isOpen"
    :title="t('customer.directory.modals.create.title')"
    :description="t('customer.directory.modals.create.description')"
    :ui="{ content: 'max-w-4xl w-[min(95vw,48rem)]', body: 'p-0', footer: 'px-4 py-3' }"
  >
    <template #body>
      <form id="create-customer-form" class="p-4 sm:p-5 space-y-6" @submit.prevent="submit">
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
          <UFormField :label="t('customer.directory.form.source')">
            <USelect
              v-model="form.source"
              :items="sourceItems"
              class="w-full"
              :placeholder="t('customer.directory.form.sourcePlaceholder')"
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
      </form>
    </template>

    <template #footer>
      <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end">
        <UButton color="neutral" variant="subtle" :disabled="creating" @click="handleCancel">
          {{ t("customer.directory.actions.cancel") }}
        </UButton>
        <UButton
          type="submit"
          form="create-customer-form"
          color="primary"
          :loading="creating"
          :disabled="creating || !valid"
        >
          {{ t("customer.directory.actions.create") }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useI18n } from "#imports";
import type { Customer } from "~/types/customer";
import { useCustomerStore } from "~/stores/customer";
import { useToastAlert } from "~/composables/useToastAlert";

const props = defineProps<{ open?: boolean }>();
const emit = defineEmits<{
  "update:open": [boolean];
  created: [customer: Customer];
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
  membershipTier: "gold",
  source: "website",
  tags: [] as string[],
  notes: "",
});

const newTag = ref("");
const creating = computed(() => mutationState.value.creating);

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

const sourceItems = [
  { value: "website", label: "官网注册" },
  { value: "miniapp", label: "小程序" },
  { value: "offline", label: "线下门店" },
  { value: "referral", label: "朋友推荐" },
  { value: "advertisement", label: "广告投放" },
  { value: "other", label: "其他" },
];

const handleSelectChange = () => blurActiveElement();

const valid = computed(
  () =>
    form.value.name.trim() &&
    form.value.phone.trim() &&
    form.value.customerType &&
    form.value.membershipTier &&
    form.value.source,
);

const addTag = () => {
  const tag = newTag.value.trim();
  if (tag && !form.value.tags.includes(tag)) {
    form.value.tags.push(tag);
    newTag.value = "";
  }
};

const removeTag = (index: number) => form.value.tags.splice(index, 1);

function handleCancel() {
  close(false);
}

const reset = () => {
  form.value = {
    name: "",
    email: "",
    phone: "",
    customerType: "individual",
    gender: "",
    birthDate: "",
    address: "",
    membershipTier: "gold",
    source: "website",
    tags: [],
    notes: "",
  };
  newTag.value = "";
};

watch(isOpen, (value) => {
  if (value) {
    reset();
  } else {
    blurActiveElement();
  }
});

async function submit() {
  if (!valid.value) {
    toast.add({
      title: t("customer.directory.modals.create.failed"),
      description: t("customer.directory.modals.create.missing"),
      color: "red",
    });
    return;
  }
  try {
    const customer = await store.createCustomer({
      name: form.value.name.trim(),
      type: form.value.customerType as any,
      phone: form.value.phone.trim(),
      email: form.value.email.trim() || undefined,
      membershipTier: form.value.membershipTier,
      source: form.value.source,
      region: form.value.address || undefined,
      country: undefined,
      accountManager: undefined,
      tags: form.value.tags,
      notes: form.value.notes.trim() || undefined,
    });
    emit("created", customer);
    close({ action: "create", customer });
    reset();
  } catch (error: any) {
    toast.add({
      title: t("customer.directory.modals.create.failed"),
      description: error?.data?.message || error?.message,
      color: "red",
    });
  }
}
</script>
