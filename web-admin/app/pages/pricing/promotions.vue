<template>
  <div class="space-y-4 p-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t("pricing.promotions.title") }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t("pricing.promotions.subtitle") }}</p>
      </div>
      <UButton icon="i-lucide-plus" color="primary" @click="openCreate">{{ t("pricing.promotions.actions.create") }}</UButton>
    </div>

    <div class="grid gap-3 md:grid-cols-5">
      <UInput v-model="filters.keyword" :placeholder="t('pricing.promotions.placeholders.keyword')" />
      <USelect v-model="filters.promotion_type" :items="typeOptions" :placeholder="t('pricing.promotions.fields.type')" />
      <USelect v-model="filters.status" :items="statusOptions" :placeholder="t('pricing.promotions.fields.status')" />
      <UInput v-model="filters.channel" :placeholder="t('pricing.promotions.fields.channel')" />
      <div class="flex gap-2">
        <UButton variant="soft" icon="i-lucide-search" @click="load">{{ t("pricing.promotions.actions.search") }}</UButton>
        <UButton variant="ghost" icon="i-lucide-rotate-ccw" @click="reset">{{ t("pricing.promotions.actions.reset") }}</UButton>
      </div>
    </div>

    <UTable :data="rows" :columns="columns" :loading="loading" class="w-full">
      <template #status-cell="{ row }">
        <UBadge :color="statusColor(row.original.status)" variant="soft">{{ row.original.status }}</UBadge>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex flex-wrap items-center gap-2">
          <UButton size="xs" variant="soft" icon="i-lucide-pencil" @click="openEdit(row.original)">
            {{ t("pricing.promotions.actions.editShort") }}
          </UButton>
          <UButton
            v-if="row.original.status === 'active'"
            size="xs"
            color="warning"
            variant="soft"
            icon="i-lucide-circle-pause"
            @click="pause(row.original.id)"
          >
            {{ t("pricing.promotions.actions.pause") }}
          </UButton>
          <UButton
            v-else
            size="xs"
            color="success"
            variant="soft"
            icon="i-lucide-circle-check"
            @click="activate(row.original.id)"
          >
            {{ t("pricing.promotions.actions.activate") }}
          </UButton>
          <UButton size="xs" variant="soft" icon="i-lucide-copy-plus" @click="clone(row.original.id)">
            {{ t("pricing.promotions.actions.clone") }}
          </UButton>
        </div>
      </template>
    </UTable>

    <UModal
      v-model:open="formOpen"
      :title="editingId ? t('pricing.promotions.actions.edit') : t('pricing.promotions.actions.create')"
      :description="t('pricing.promotions.dialogDescription')"
      :ui="{ content: 'max-w-5xl w-[88vw]', body: 'p-6', footer: 'justify-end' }"
    >
      <template #body>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          <UAlert
            v-if="formError"
            color="error"
            variant="soft"
            icon="i-lucide-circle-alert"
            :description="formError"
            class="lg:col-span-4"
          />
          <UFormField :label="t('pricing.promotions.fields.code')" required class="lg:col-span-2">
            <div class="grid grid-cols-[minmax(0,1fr)_auto] gap-2">
              <UInput v-model="form.code" :disabled="!!editingId" class="w-full" />
              <UButton
                v-if="!editingId"
                type="button"
                variant="soft"
                icon="i-lucide-refresh-cw"
                @click="generateCode"
              >
                {{ t("pricing.promotions.actions.generate") }}
              </UButton>
            </div>
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.name')" required class="lg:col-span-2">
            <UInput v-model="form.name" class="w-full" />
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.type')" required>
            <USelect v-model="form.promotion_type" :items="typeOptions.filter((i) => i.value)" :disabled="!!editingId" class="w-full" />
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.validFrom')" required>
            <UInput v-model="form.valid_from" type="datetime-local" class="w-full" />
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.validTo')" required>
            <UInput v-model="form.valid_to" type="datetime-local" class="w-full" />
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.priority')">
            <UInput v-model.number="form.priority" type="number" min="0" class="w-full" />
          </UFormField>

          <UFormField :label="t('pricing.promotions.fields.minOrderAmount')">
            <UInput v-model.number="form.min_order_amount_minor" type="number" min="0" class="w-full" />
          </UFormField>
          <UFormField v-if="form.promotion_type === 'amount_off'" :label="t('pricing.promotions.fields.discountAmount')" required>
            <UInput v-model.number="form.discount_amount_minor" type="number" min="1" class="w-full" />
          </UFormField>
          <UFormField v-else :label="t('pricing.promotions.fields.discountBps')" required>
            <UInput v-model.number="form.discount_percent_bps" type="number" min="1" max="10000" class="w-full" />
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.maxDiscount')">
            <UInput v-model.number="form.max_discount_minor" type="number" min="0" class="w-full" />
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.scope')">
            <USelect v-model="form.scope_type" :items="scopeOptions" class="w-full" />
          </UFormField>

          <UFormField :label="t('pricing.promotions.fields.channel')" class="lg:col-span-2">
            <UPopover :ui="multiSelectPopoverUi">
              <UButton type="button" color="neutral" variant="outline" trailing-icon="i-lucide-chevron-down" class="h-10 w-full justify-between overflow-hidden px-3">
                <span class="truncate text-left font-normal">{{ selectedSummary(form.channels, channelOptions, t("pricing.promotions.placeholders.channels")) }}</span>
              </UButton>
              <template #content>
                <div class="space-y-2 bg-white p-2 dark:bg-gray-950">
                  <UInput v-model="channelSearch" size="sm" :placeholder="t('pricing.promotions.placeholders.search')" />
                  <div class="max-h-56 overflow-auto rounded-md">
                    <button
                      v-for="item in filteredChannelOptions"
                      :key="item.value"
                      type="button"
                      class="flex w-full items-center gap-2 rounded-md px-2 py-2 text-left text-sm text-gray-800 hover:bg-gray-100 dark:text-gray-100 dark:hover:bg-gray-800"
                      @click="toggleSelected(form.channels, item.value)"
                    >
                      <span class="flex h-4 w-4 shrink-0 items-center justify-center rounded border border-gray-300 bg-white dark:border-gray-600 dark:bg-gray-900">
                        <UIcon v-if="form.channels.includes(item.value)" name="i-lucide-check" class="h-3 w-3 text-primary" />
                      </span>
                      <span class="truncate">{{ item.label }}</span>
                    </button>
                  </div>
                </div>
              </template>
            </UPopover>
          </UFormField>
          <UFormField v-if="form.scope_type === 'sku'" :label="t('pricing.promotions.fields.skuIds')" required class="lg:col-span-2">
            <UInput v-model="form.sku_ids" class="w-full" />
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.exclusionGroup')" :class="form.scope_type === 'sku' ? '' : 'lg:col-span-2'">
            <UPopover :ui="multiSelectPopoverUi">
              <UButton type="button" color="neutral" variant="outline" trailing-icon="i-lucide-chevron-down" class="h-10 w-full justify-between overflow-hidden px-3">
                <span class="truncate text-left font-normal">{{ selectedSummary(form.exclusion_groups, exclusionGroupOptions, t("pricing.promotions.placeholders.exclusionGroups")) }}</span>
              </UButton>
              <template #content>
                <div class="space-y-2 bg-white p-2 dark:bg-gray-950">
                  <UInput v-model="exclusionSearch" size="sm" :placeholder="t('pricing.promotions.placeholders.search')" />
                  <div class="max-h-56 overflow-auto rounded-md">
                    <button
                      v-for="item in filteredExclusionGroupOptions"
                      :key="item.value"
                      type="button"
                      class="flex w-full items-center gap-2 rounded-md px-2 py-2 text-left text-sm text-gray-800 hover:bg-gray-100 dark:text-gray-100 dark:hover:bg-gray-800"
                      @click="toggleSelected(form.exclusion_groups, item.value)"
                    >
                      <span class="flex h-4 w-4 shrink-0 items-center justify-center rounded border border-gray-300 bg-white dark:border-gray-600 dark:bg-gray-900">
                        <UIcon v-if="form.exclusion_groups.includes(item.value)" name="i-lucide-check" class="h-3 w-3 text-primary" />
                      </span>
                      <span class="truncate">{{ item.label }}</span>
                    </button>
                  </div>
                </div>
              </template>
            </UPopover>
          </UFormField>

          <UFormField :label="t('pricing.promotions.fields.stackable')">
            <USwitch v-model="form.stackable" />
          </UFormField>
          <UFormField :label="t('pricing.promotions.fields.stackableWithCoupon')">
            <USwitch v-model="form.stackable_with_coupon" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton variant="ghost" @click="formOpen = false">{{ t("pricing.promotions.actions.cancel") }}</UButton>
          <UButton variant="soft" @click="save('draft')">{{ t("pricing.promotions.actions.saveDraft") }}</UButton>
          <UButton color="primary" @click="save('activate')">{{ t("pricing.promotions.actions.saveAndActivate") }}</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { usePromotionsApi, type PromotionCampaign, type PromotionPayload } from "~/composables/api/usePromotions";
import { useChannelsApi } from "~/composables/useChannels";

type SelectOption = {
  label: string;
  value: string;
};

const api = usePromotionsApi();
const channelsApi = useChannelsApi();
const { t } = useI18n();
const loading = ref(false);
const rows = ref<PromotionCampaign[]>([]);
const formOpen = ref(false);
const editingId = ref("");
const formError = ref("");
const channelSearch = ref("");
const exclusionSearch = ref("");
const dynamicChannelOptions = ref<SelectOption[]>([]);
const multiSelectPopoverUi = {
  content: "z-[80] w-[var(--reka-popper-anchor-width)] min-w-80 overflow-hidden rounded-lg border border-gray-200 bg-white p-0 text-gray-900 shadow-xl ring-1 ring-black/5 dark:border-gray-700 dark:bg-gray-950 dark:text-gray-100 dark:ring-white/10",
};

const allFilterValue = "__all";
const filters = reactive({ keyword: "", promotion_type: allFilterValue, status: allFilterValue, channel: "", page: 1, page_size: 20 });
const typeOptions = computed<SelectOption[]>(() => [
  { label: t("pricing.promotions.options.allTypes"), value: allFilterValue },
  { label: t("pricing.promotions.options.amountOff"), value: "amount_off" },
  { label: t("pricing.promotions.options.percentOff"), value: "percent_off" },
]);
const statusOptions = computed<SelectOption[]>(() => [
  { label: t("pricing.promotions.options.allStatuses"), value: allFilterValue },
  { label: t("pricing.promotions.options.draft"), value: "draft" },
  { label: t("pricing.promotions.options.active"), value: "active" },
  { label: t("pricing.promotions.options.paused"), value: "paused" },
  { label: t("pricing.promotions.options.expired"), value: "expired" },
]);
const scopeOptions = computed<SelectOption[]>(() => [
  { label: t("pricing.promotions.options.allScope"), value: "all" },
  { label: t("pricing.promotions.options.skuScope"), value: "sku" },
]);
const fallbackChannelOptions = computed<SelectOption[]>(() => [
  { label: t("pricing.promotions.channels.official"), value: "official" },
  { label: t("pricing.promotions.channels.miniapp"), value: "miniapp" },
  { label: t("pricing.promotions.channels.h5"), value: "h5" },
]);
const channelOptions = computed<SelectOption[]>(() => dynamicChannelOptions.value.length ? dynamicChannelOptions.value : fallbackChannelOptions.value);
const exclusionGroupOptions = computed<SelectOption[]>(() => [
  { label: t("pricing.promotions.exclusionGroups.orderDiscount"), value: "order_discount" },
  { label: t("pricing.promotions.exclusionGroups.amountOffCampaign"), value: "amount_off_campaign" },
  { label: t("pricing.promotions.exclusionGroups.percentOffCampaign"), value: "percent_off_campaign" },
  { label: t("pricing.promotions.exclusionGroups.channelCampaign"), value: "channel_campaign" },
]);
const filteredChannelOptions = computed(() => filterOptions(channelOptions.value, channelSearch.value));
const filteredExclusionGroupOptions = computed(() => filterOptions(exclusionGroupOptions.value, exclusionSearch.value));

const form = reactive({
  code: "",
  name: "",
  promotion_type: "amount_off",
  valid_from: "",
  valid_to: "",
  min_order_amount_minor: 0,
  discount_amount_minor: 0,
  discount_percent_bps: 8500,
  max_discount_minor: 0,
  scope_type: "all",
  sku_ids: "",
  channels: [] as string[],
  priority: 100,
  stackable: true,
  stackable_with_coupon: true,
  exclusion_groups: [] as string[],
});

const pad2 = (value: number) => String(value).padStart(2, "0");
const toDatetimeLocal = (date: Date) => {
  const year = date.getFullYear();
  const month = pad2(date.getMonth() + 1);
  const day = pad2(date.getDate());
  const hour = pad2(date.getHours());
  const minute = pad2(date.getMinutes());
  return `${year}-${month}-${day}T${hour}:${minute}`;
};

const buildPromotionCode = () => {
  const now = new Date();
  const date = `${now.getFullYear()}${pad2(now.getMonth() + 1)}${pad2(now.getDate())}`;
  const time = `${pad2(now.getHours())}${pad2(now.getMinutes())}${pad2(now.getSeconds())}`;
  return `PROMO-${date}-${time}`;
};

const generateCode = () => {
  form.code = buildPromotionCode();
};

const columns: TableColumn<PromotionCampaign>[] = [
  { accessorKey: "code", header: t("pricing.promotions.fields.code") },
  { accessorKey: "name", header: t("pricing.promotions.fields.name") },
  { accessorKey: "promotion_type", header: t("pricing.promotions.fields.type") },
  { accessorKey: "status", header: t("pricing.promotions.fields.status") },
  { accessorKey: "valid_from", header: t("pricing.promotions.fields.validFrom") },
  { accessorKey: "valid_to", header: t("pricing.promotions.fields.validTo") },
  { id: "actions", header: t("pricing.promotions.fields.actions") },
];

const load = async () => {
  loading.value = true;
  try {
    const resp = await api.list({
      ...filters,
      promotion_type: filters.promotion_type === allFilterValue ? "" : filters.promotion_type as any,
      status: filters.status === allFilterValue ? "" : filters.status as any,
    });
    rows.value = resp?.items || [];
  } finally {
    loading.value = false;
  }
};

const reset = () => {
  Object.assign(filters, { keyword: "", promotion_type: allFilterValue, status: allFilterValue, channel: "", page: 1, page_size: 20 });
  load();
};

const loadChannels = async () => {
  try {
    const resp = await channelsApi.listChannels({ page: 1, pageSize: 200 });
    const byValue = new Map<string, SelectOption>();
    for (const item of resp?.items || []) {
      const value = String(item.platform || item.id || "").trim();
      if (!value || byValue.has(value)) continue;
      const label = String(item.name || item.platform || value).trim();
      if (!label) continue;
      byValue.set(value, { label, value });
    }
    dynamicChannelOptions.value = Array.from(byValue.values());
  } catch {
    dynamicChannelOptions.value = [];
  }
};

const openCreate = () => {
  editingId.value = "";
  formError.value = "";
  const now = new Date();
  const validTo = new Date(now);
  validTo.setDate(validTo.getDate() + 7);
  Object.assign(form, { code: buildPromotionCode(), name: "", promotion_type: "amount_off", valid_from: toDatetimeLocal(now), valid_to: toDatetimeLocal(validTo), min_order_amount_minor: 0, discount_amount_minor: 0, discount_percent_bps: 8500, max_discount_minor: 0, scope_type: "all", sku_ids: "", channels: [], priority: 100, stackable: true, stackable_with_coupon: true, exclusion_groups: [] });
  formOpen.value = true;
};

const openEdit = (row: PromotionCampaign) => {
  editingId.value = row.id;
  formError.value = "";
  const condition: any = row.condition_rule || {};
  const scope: any = row.scope_rule || {};
  const action: any = row.action_rule || {};
  const stacking: any = row.stacking_rule || {};
  Object.assign(form, {
    code: row.code,
    name: row.name,
    promotion_type: row.promotion_type,
    valid_from: toLocal(row.valid_from),
    valid_to: toLocal(row.valid_to),
    min_order_amount_minor: condition.min_order_amount_minor || 0,
    discount_amount_minor: action.discount_amount_minor || 0,
    discount_percent_bps: action.discount_percent_bps || 8500,
    max_discount_minor: action.max_discount_minor || 0,
    scope_type: scope.scope_type || "all",
    sku_ids: (scope.sku_ids || []).join(","),
    channels: scope.channels || [],
    priority: stacking.priority ?? 100,
    stackable: stacking.stackable ?? true,
    stackable_with_coupon: stacking.stackable_with_coupon ?? true,
    exclusion_groups: stacking.exclusion_groups || (stacking.exclusion_group ? [stacking.exclusion_group] : []),
  });
  formOpen.value = true;
};

const parseDatetime = (value: string) => {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
};

const payload = (saveAction: "draft" | "activate"): PromotionPayload | null => {
  formError.value = "";
  const validFrom = parseDatetime(form.valid_from);
  const validTo = parseDatetime(form.valid_to);
  if (!validFrom || !validTo) {
    formError.value = t("pricing.promotions.validation.validTimeRequired");
    return null;
  }
  if (validFrom > validTo) {
    formError.value = t("pricing.promotions.validation.validTimeOrder");
    return null;
  }
  return {
    code: form.code,
    name: form.name,
    promotion_type: form.promotion_type as any,
    valid_from: validFrom.toISOString(),
    valid_to: validTo.toISOString(),
    save_action: saveAction,
    condition_rule: { min_order_amount_minor: Number(form.min_order_amount_minor || 0) },
    scope_rule: { scope_type: form.scope_type, sku_ids: csv(form.sku_ids), channels: form.channels },
    action_rule: form.promotion_type === "amount_off"
      ? { discount_amount_minor: Number(form.discount_amount_minor || 0) }
      : { discount_percent_bps: Number(form.discount_percent_bps || 0), max_discount_minor: Number(form.max_discount_minor || 0) },
    stacking_rule: {
      priority: Number(form.priority || 100),
      stackable: form.stackable,
      stackable_with_coupon: form.stackable_with_coupon,
      exclusion_group: form.exclusion_groups[0] || "",
      exclusion_groups: form.exclusion_groups,
    },
  };
};

const save = async (saveAction: "draft" | "activate") => {
  const nextPayload = payload(saveAction);
  if (!nextPayload) return;
  if (editingId.value) await api.update(editingId.value, nextPayload);
  else await api.create(nextPayload);
  formOpen.value = false;
  await load();
};

const activate = async (id: string) => { await api.activate(id); await load(); };
const pause = async (id: string) => { await api.pause(id); await load(); };
const clone = async (id: string) => { await api.clone(id); await load(); };

const csv = (value: string) => value.split(",").map((v) => v.trim()).filter(Boolean);
const toLocal = (value: string) => value ? value.slice(0, 16) : "";
const statusColor = (status: string) => status === "active" ? "success" : status === "paused" ? "warning" : status === "expired" ? "neutral" : "info";
const filterOptions = (options: SelectOption[], keyword: string) => {
  const q = keyword.trim().toLowerCase();
  if (!q) return options;
  return options.filter((item) => item.label.toLowerCase().includes(q) || item.value.toLowerCase().includes(q));
};
const toggleSelected = (values: string[], value: string) => {
  const index = values.indexOf(value);
  if (index >= 0) values.splice(index, 1);
  else values.push(value);
};
const selectedSummary = (values: string[], options: SelectOption[], placeholder: string) => {
  if (!values.length) return placeholder;
  const byValue = new Map(options.map((item) => [item.value, item.label]));
  return values.map((value) => byValue.get(value) || value).join(", ");
};

onMounted(() => {
  load();
  loadChannels();
});
</script>
