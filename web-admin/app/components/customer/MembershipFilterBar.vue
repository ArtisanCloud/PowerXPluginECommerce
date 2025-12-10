<template>
  <UCard class="mb-4">
    <div class="flex flex-col gap-6">
      <div class="grid grid-cols-12 gap-4">
        <UFormField
          :label="t('customer.membership.filters.tier')"
          class="col-span-12 sm:col-span-6 xl:col-span-3"
          :ui="inlineFieldUi"
        >
          <USelect v-model="tierModel" class="w-full" :items="tierOptions" />
        </UFormField>
        <UFormField
          :label="t('customer.membership.filters.retentionStatus')"
          class="col-span-12 sm:col-span-6 xl:col-span-3"
          :ui="inlineFieldUi"
        >
          <USelect v-model="retentionModel" class="w-full" :items="retentionOptions" />
        </UFormField>
        <UFormField
          :label="t('customer.membership.filters.benefitStatus')"
          class="col-span-12 sm:col-span-6 xl:col-span-3"
          :ui="inlineFieldUi"
        >
          <USelect v-model="benefitModel" class="w-full" :items="benefitOptions" />
        </UFormField>
        <div class="col-span-12 sm:col-span-6 xl:col-span-3 flex items-center justify-end gap-2">
          <UButton variant="ghost" @click="handleReset">
            {{ t("customer.membership.actions.reset") }}
          </UButton>
          <UButton color="primary" :loading="isSubmitting" @click="handleApply">
            {{ t("customer.membership.actions.apply") }}
          </UButton>
        </div>
      </div>
      <div class="grid grid-cols-12 gap-4">
        <UFormField
          :label="t('customer.membership.filters.growthRange')"
          class="col-span-12 lg:col-span-6"
          :ui="rangeFieldUi"
        >
          <div class="flex items-center gap-2 w-full">
            <UInput v-model="form.growthMin" class="w-full" type="number" placeholder="0" />
            <span class="text-sm text-gray-400">~</span>
            <UInput v-model="form.growthMax" class="w-full" type="number" placeholder="200" />
          </div>
        </UFormField>
        <UFormField
          :label="t('customer.membership.filters.pointsRange')"
          class="col-span-12 lg:col-span-6"
          :ui="rangeFieldUi"
        >
          <div class="flex items-center gap-2 w-full">
            <UInput v-model="form.pointsMin" class="w-full" type="number" placeholder="0" />
            <span class="text-sm text-gray-400">~</span>
            <UInput v-model="form.pointsMax" class="w-full" type="number" placeholder="2000" />
          </div>
        </UFormField>
      </div>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { reactive, watch, computed, ref } from "vue";
import { storeToRefs } from "pinia";
import { useI18n } from "#imports";
import { useCustomerStore } from "~/stores/customer";
import type { MembershipFilters } from "~/types/customer";
import { useCustomerMetrics } from "~/composables/useCustomerMetrics";
import { useToastAlert } from "~/composables/useToastAlert";

const store = useCustomerStore();
const { membershipFilters, membershipLoading } = storeToRefs(store);
const { t } = useI18n();
const toast = useToastAlert();
const metrics = useCustomerMetrics();

const form = reactive({
  tier: membershipFilters.value.tier || "",
  retentionStatus: membershipFilters.value.retentionStatus || "",
  benefitStatus: membershipFilters.value.benefitStatus || "",
  growthMin: membershipFilters.value.growthRange?.[0] ?? "",
  growthMax: membershipFilters.value.growthRange?.[1] ?? "",
  pointsMin: membershipFilters.value.pointsRange?.[0] ?? "",
  pointsMax: membershipFilters.value.pointsRange?.[1] ?? "",
});
const ANY_OPTION_VALUE = "__all__";

const inlineFieldUi = {
  root: "flex items-center gap-3 w-full",
  wrapper: "w-28 sm:w-32 shrink-0",
  labelWrapper: "flex items-center gap-2",
  label: "text-sm font-medium text-gray-600 whitespace-nowrap",
  container: "mt-0 flex-1",
};
const rangeFieldUi = {
  ...inlineFieldUi,
  container: "mt-0 flex-1",
};

const tierOptions = computed(() => [
  { label: t("customer.membership.filters.any"), value: ANY_OPTION_VALUE },
  { label: t("customer.membership.filters.tierOptions.gold"), value: "gold" },
  { label: t("customer.membership.filters.tierOptions.silver"), value: "silver" },
  { label: t("customer.membership.filters.tierOptions.platinum"), value: "platinum" },
  { label: t("customer.membership.filters.tierOptions.diamond"), value: "diamond" },
]);

const retentionOptions = computed(() => [
  { label: t("customer.membership.filters.any"), value: ANY_OPTION_VALUE },
  { label: t("customer.membership.segments.safe"), value: "safe" },
  { label: t("customer.membership.segments.warning"), value: "warning" },
  { label: t("customer.membership.segments.downgrade"), value: "downgrade" },
]);

const benefitOptions = computed(() => [
  { label: t("customer.membership.filters.benefitAny"), value: ANY_OPTION_VALUE },
  { label: t("customer.membership.filters.benefitUnused"), value: "unused" },
  { label: t("customer.membership.filters.benefitUsed"), value: "used" },
]);

const createSelectModel = <K extends keyof typeof form>(key: K) =>
  computed<string>({
    get: () => (form[key] ? String(form[key]) : ANY_OPTION_VALUE),
    set: (value) => {
      form[key] = (value === ANY_OPTION_VALUE ? "" : value) as (typeof form)[K];
    },
  });

const tierModel = createSelectModel("tier");
const retentionModel = createSelectModel("retentionStatus");
const benefitModel = createSelectModel("benefitStatus");

const isSubmitting = ref(false);

const syncForm = (next: MembershipFilters) => {
  form.tier = next.tier || "";
  form.retentionStatus = next.retentionStatus || "";
  form.benefitStatus = next.benefitStatus || "";
  form.growthMin = next.growthRange?.[0] ?? "";
  form.growthMax = next.growthRange?.[1] ?? "";
  form.pointsMin = next.pointsRange?.[0] ?? "";
  form.pointsMax = next.pointsRange?.[1] ?? "";
};

watch(
  () => membershipFilters.value,
  (value) => syncForm(value),
  { deep: true }
);

const toFilters = (): MembershipFilters => {
  const filters: MembershipFilters = {
    page: 1,
    pageSize: membershipFilters.value.pageSize,
  };
  if (form.tier) filters.tier = form.tier;
  if (form.retentionStatus) filters.retentionStatus = form.retentionStatus;
  if (form.benefitStatus) filters.benefitStatus = form.benefitStatus;
  if (form.growthMin !== "" || form.growthMax !== "") {
    const min = form.growthMin === "" ? undefined : Number(form.growthMin);
    const max = form.growthMax === "" ? undefined : Number(form.growthMax);
    filters.growthRange = [min ?? 0, max ?? min ?? 0] as [number, number];
  }
  if (form.pointsMin !== "" || form.pointsMax !== "") {
    const min = form.pointsMin === "" ? undefined : Number(form.pointsMin);
    const max = form.pointsMax === "" ? undefined : Number(form.pointsMax);
    filters.pointsRange = [min ?? 0, max ?? min ?? 0] as [number, number];
  }
  return filters;
};

const handleApply = async () => {
  const filters = toFilters();
  if (
    filters.growthRange &&
    filters.growthRange[1] &&
    filters.growthRange[0] > filters.growthRange[1]
  ) {
    toast.add({
      title: t("customer.membership.messages.invalidGrowthRange"),
      color: "warning",
    });
    return;
  }
  isSubmitting.value = true;
  const stopTimer = metrics.startLatencyTimer("membership_filters_apply");
  metrics.recordEvent({
    name: "membership_filters_submit",
    metadata: { ...filters },
  });
  try {
    await store.fetchMemberships(filters);
  } finally {
    stopTimer();
    isSubmitting.value = false;
  }
};

const handleReset = async () => {
  store.resetMembershipFilters();
  syncForm(store.membershipFilters);
  metrics.recordEvent({
    name: "membership_filters_reset",
  });
  await store.fetchMemberships();
};

watch(
  () => membershipLoading.value,
  (loading) => {
    if (!loading) {
      isSubmitting.value = false;
    }
  }
);
</script>
