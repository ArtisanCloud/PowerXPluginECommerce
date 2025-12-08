<template>
  <UCard class="mb-4">
    <div class="flex flex-col gap-4">
      <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <UFormField :label="t('customer.membership.filters.tier')">
          <USelectMenu
            v-model="form.tier"
            :options="tierOptions"
            value-attribute="value"
            option-attribute="label"
            clearable
          />
        </UFormField>
        <UFormField :label="t('customer.membership.filters.retentionStatus')">
          <USelectMenu
            v-model="form.retentionStatus"
            :options="retentionOptions"
            value-attribute="value"
            option-attribute="label"
            clearable
          />
        </UFormField>
        <UFormField :label="t('customer.membership.filters.benefitStatus')">
          <USelectMenu
            v-model="form.benefitStatus"
            :options="benefitOptions"
            value-attribute="value"
            option-attribute="label"
            clearable
          />
        </UFormField>
      </div>
      <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <UFormField :label="t('customer.membership.filters.growthRange')">
          <div class="flex items-center gap-2">
            <UInput v-model="form.growthMin" type="number" placeholder="0" />
            <span class="text-sm text-gray-400">~</span>
            <UInput v-model="form.growthMax" type="number" placeholder="200" />
          </div>
        </UFormField>
        <UFormField :label="t('customer.membership.filters.pointsRange')">
          <div class="flex items-center gap-2">
            <UInput v-model="form.pointsMin" type="number" placeholder="0" />
            <span class="text-sm text-gray-400">~</span>
            <UInput v-model="form.pointsMax" type="number" placeholder="2000" />
          </div>
        </UFormField>
        <div class="flex items-end justify-end gap-2">
          <UButton variant="ghost" @click="handleReset">
            {{ t("customer.membership.actions.reset") }}
          </UButton>
          <UButton color="primary" :loading="isSubmitting" @click="handleApply">
            {{ t("customer.membership.actions.apply") }}
          </UButton>
        </div>
      </div>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { reactive, watch, computed, ref } from "vue";
import { storeToRefs } from "pinia";
import { useI18n, useToast } from "#imports";
import { useCustomerStore } from "~/stores/customer";
import type { MembershipFilters } from "~/types/customer";
import { useCustomerMetrics } from "~/composables/useCustomerMetrics";

const store = useCustomerStore();
const { membershipFilters, membershipLoading } = storeToRefs(store);
const { t } = useI18n();
const toast = useToast();
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

const tierOptions = computed(() => [
  { label: t("customer.membership.filters.tierOptions.gold"), value: "gold" },
  { label: t("customer.membership.filters.tierOptions.silver"), value: "silver" },
  { label: t("customer.membership.filters.tierOptions.platinum"), value: "platinum" },
  { label: t("customer.membership.filters.tierOptions.diamond"), value: "diamond" },
]);

const retentionOptions = computed(() => [
  { label: t("customer.membership.segments.safe"), value: "safe" },
  { label: t("customer.membership.segments.warning"), value: "warning" },
  { label: t("customer.membership.segments.downgrade"), value: "downgrade" },
]);

const benefitOptions = computed(() => [
  { label: t("customer.membership.filters.benefitAny"), value: "" },
  { label: t("customer.membership.filters.benefitUnused"), value: "unused" },
  { label: t("customer.membership.filters.benefitUsed"), value: "used" },
]);

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
