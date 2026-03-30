<template>
  <div>
    <!-- 页面标题和筛选 -->
    <div class="mb-6 space-y-4">
      <div class="flex items-center justify-between">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          {{ $t("orders.title") }}
        </h1>
        <UButton color="primary" @click="createOpen = true">
          {{ $t("orders.create") }}
        </UButton>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <UInput
          v-model="orderNoKeyword"
          :placeholder="$t('orders.filters.orderNoPlaceholder')"
          class="w-56"
        />
        <USelect
          v-model="selectedStatus"
          :items="statusOptions"
          :placeholder="$t('orders.allStatus')"
          class="w-40"
        />
        <USelectMenu
          v-model="filterCustomerId"
          v-model:search-term="filterCustomerSearchTerm"
          :items="filterCustomerOptions"
          value-key="value"
          label-key="label"
          :loading="filterCustomerLoading"
          :portal="false"
          :ui="{ content: 'z-[60] max-h-60 overflow-auto' }"
          searchable
          :ignore-filter="true"
          class="w-56"
          :placeholder="$t('orders.filters.customerPlaceholder')"
        >
          <template #option="{ option }">
            <div class="flex flex-col">
              <span class="text-sm font-medium text-gray-900 dark:text-white">{{ option.label }}</span>
              <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
            </div>
          </template>
          <template #empty>
            <div class="px-3 py-2 text-sm text-gray-500">
              {{ $t("orders.createCustomerEmpty") }}
            </div>
          </template>
        </USelectMenu>
        <UButton variant="ghost" @click="resetFilters">
          {{ $t("common.reset") }}
        </UButton>
      </div>
    </div>

    <!-- 订单表格 -->
    <UCard>
      <UTable
        :data="rows"
        :columns="columns"
        :loading="loading"
        class="w-full"
      >
        <!-- v3: 用 -cell，而不是 -data -->
        <template #amountMinor-cell="{ row }">
          <div class="text-right font-medium">
            {{
              new Intl.NumberFormat("zh-CN", {
                style: "currency",
                currency: row.original.currency || "CNY",
              }).format(Number(row.original.amountMinor || 0) / 100)
            }}
          </div>
        </template>

        <template #customerName-cell="{ row }">
          <NuxtLink
            v-if="row.original.customerId"
            class="text-primary-500 hover:underline"
            :to="customerDetailTo(row.original.customerId)"
          >
            {{ row.original.customerName || row.original.customerId }}
          </NuxtLink>
          <span v-else>-</span>
        </template>

        <template #source-cell="{ row }">
          <UBadge
            color="neutral"
            variant="subtle"
            :title="row.original.channel ? `channel: ${row.original.channel}` : ''"
          >
            {{ sourceText(row.original.createdByType) }}
          </UBadge>
          <span v-if="row.original.channel" class="ml-2 text-xs text-gray-500 dark:text-gray-400">
            {{ row.original.channel }}
          </span>
        </template>

        <template #createdAt-cell="{ getValue }">
          <span>
            {{ new Date(getValue() as string).toLocaleString("zh-CN") }}
          </span>
        </template>

        <template #status-cell="{ getValue }">
          <UBadge :color="getStatusColor(String(getValue()))" variant="subtle">
            {{ statusText(String(getValue())) }}
          </UBadge>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              :disabled="!row.original.orderId"
              @click="goView(row.original.orderId)"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-cog-6-tooth"
              :disabled="!row.original.orderId"
              @click="goEdit(row.original.orderId)"
            >
              {{ $t("common.edit") }}
            </UButton>
            <UButton
              v-if="row.original.status === 'pending_payment'"
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              :disabled="!row.original.orderId || deleting"
              @click="openDelete(row.original.orderId)"
            >
              {{ $t("orders.delete") }}
            </UButton>
          </div>
        </template>
      </UTable>

      <template #footer>
        <div
          class="flex flex-col gap-3 text-sm text-gray-500 dark:text-gray-400 lg:flex-row lg:items-center lg:justify-between"
        >
          <span>{{ $t("common.total", { count: total }) }}</span>
          <div class="flex items-center gap-3">
            <USelect v-model="pageSize" :items="pageSizeItems" class="w-24" />
            <UPagination v-model="page" :total="total" :page-count="pageSize" show-first show-last />
          </div>
        </div>
      </template>
    </UCard>

    <UModal
      v-model:open="createOpen"
      :title="$t('orders.createTitle')"
      :description="$t('orders.createDesc')"
      :prevent-close="creating"
    >
      <template #body>
        <div class="space-y-4 p-4 sm:p-5">
          <UFormField :label="$t('orders.createCustomerId')" required>
            <USelectMenu
              v-model="createForm.customerId"
              v-model:search-term="customerSearchTerm"
              :items="customerOptions"
              value-key="value"
              label-key="label"
              :loading="customerLoading"
              :portal="false"
              :ui="{ content: 'z-[60] max-h-60 overflow-auto' }"
              searchable
              :ignore-filter="true"
              class="w-full"
              :placeholder="$t('orders.createCustomerIdHint')"
            >
              <template #option="{ option }">
                <div class="flex flex-col">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">{{ option.label }}</span>
                  <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
                </div>
              </template>
              <template #empty>
                <div class="px-3 py-2 text-sm text-gray-500">
                  {{ $t("orders.createCustomerEmpty") }}
                </div>
              </template>
            </USelectMenu>
          </UFormField>
          <UFormField :label="$t('orders.createChannel')" required>
            <USelectMenu
              v-model="createForm.channel"
              v-model:search-term="channelSearchTerm"
              :items="channelOptions"
              value-key="value"
              label-key="label"
              :loading="channelLoading"
              :portal="false"
              :ui="{ content: 'z-[60] max-h-60 overflow-auto' }"
              searchable
              :ignore-filter="true"
              class="w-full"
              :placeholder="$t('orders.createChannelHint')"
            >
              <template #option="{ option }">
                <div class="flex flex-col">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">{{ option.label }}</span>
                  <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
                </div>
              </template>
              <template #empty>
                <div class="px-3 py-2 text-sm text-gray-500">
                  {{ $t("orders.createChannelEmpty") }}
                </div>
              </template>
            </USelectMenu>
          </UFormField>
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-gray-900 dark:text-white">
                {{ $t("orders.createAddressBook") }}
                <span class="text-rose-500">*</span>
              </span>
              <div class="flex items-center gap-2">
                <UButton
                  size="xs"
                  variant="outline"
                  icon="i-heroicons-plus"
                  @click="openAddressModal('create')"
                >
                  {{ $t("orders.createAddressAdd") }}
                </UButton>
                <UButton
                  size="xs"
                  variant="ghost"
                  icon="i-heroicons-pencil-square"
                  :disabled="!selectedAddressId"
                  @click="openAddressModal('edit')"
                >
                  {{ $t("orders.createAddressEdit") }}
                </UButton>
              </div>
            </div>
            <USelectMenu
              v-model="selectedAddressId"
              :items="addressOptions"
              value-key="value"
              label-key="label"
              :loading="addressLoading"
              :portal="false"
              :ui="{ content: 'z-[60] max-h-60 overflow-auto' }"
              class="w-full"
              :placeholder="$t('orders.createAddressBookHint')"
              @update:model-value="handleAddressChange"
            >
              <template #option="{ option }">
                <div class="flex flex-col">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">{{ option.label }}</span>
                  <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
                </div>
              </template>
              <template #empty>
                <div class="px-3 py-2 text-sm text-gray-500">
                  {{ $t("orders.createAddressBookEmpty") }}
                </div>
              </template>
            </USelectMenu>
            <div
              v-if="selectedAddress"
              class="rounded-lg border border-gray-200/70 p-3 text-xs text-gray-500 dark:border-gray-700"
            >
              <div class="text-sm font-medium text-gray-900 dark:text-white">
                {{ selectedAddress.shippingAddress?.recipientName || "-" }}
                <span class="ml-2 text-xs text-gray-400">
                  {{ selectedAddress.shippingAddress?.recipientPhone || "" }}
                </span>
              </div>
              <div class="mt-1">
                {{
                  [
                    selectedAddress.shippingAddress?.province,
                    selectedAddress.shippingAddress?.city,
                    selectedAddress.shippingAddress?.district,
                    selectedAddress.shippingAddress?.address1,
                    selectedAddress.shippingAddress?.address2,
                  ]
                    .filter(Boolean)
                    .join(" ")
                }}
              </div>
            </div>
          </div>
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <div class="space-y-1">
                <div class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ $t("orders.createItemsTitle") }}
                </div>
                <div class="text-xs text-gray-500">
                  {{ $t("orders.createItemsGuide") }}
                </div>
              </div>
              <UButton
                size="xs"
                variant="outline"
                icon="i-heroicons-plus"
                @click="addItemRow"
              >
                {{ $t("orders.createItemAdd") }}
              </UButton>
            </div>
            <div
              v-for="(item, index) in createForm.items"
              :key="item.rowId"
              class="rounded-lg border border-gray-200/70 dark:border-gray-700 p-3 space-y-3"
            >
              <div class="flex items-center justify-between">
                <div class="text-xs uppercase tracking-[0.2em] text-gray-500">
                  {{ $t("orders.createItemIndex", { index: index + 1 }) }}
                </div>
                <UButton
                  v-if="createForm.items.length > 1"
                  size="xs"
                  color="neutral"
                  variant="soft"
                  icon="i-heroicons-trash"
                  @click="removeItemRow(item.rowId)"
                >
                  {{ $t("orders.createItemRemove") }}
                </UButton>
              </div>
              <UFormField :label="$t('orders.createSpu')" required>
                <USelectMenu
                  v-model="item.spuId"
                  v-model:search-term="item.spuSearchTerm"
                  :items="item.spuOptions"
                  value-key="value"
                  label-key="label"
                  :loading="item.spuLoading"
                  :portal="false"
                  :ui="{ content: 'z-[60] max-h-60 overflow-auto' }"
                  searchable
                  :ignore-filter="true"
                  class="w-full"
                  :placeholder="$t('orders.createSpuHint')"
                  @update:model-value="() => handleSpuChange(item)"
                  @update:search-term="(value) => handleSpuSearch(item, value)"
                >
                  <template #option="{ option }">
                    <div class="flex flex-col">
                      <span class="text-sm font-medium text-gray-900 dark:text-white">{{ option.label }}</span>
                      <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
                    </div>
                  </template>
                  <template #empty>
                    <div class="px-3 py-2 text-sm text-gray-500">
                      {{ $t("orders.createSpuEmpty") }}
                    </div>
                  </template>
                </USelectMenu>
              </UFormField>

              <div v-if="item.specGroups.length" class="grid gap-3 sm:grid-cols-2">
                <UFormField
                  v-for="group in item.specGroups"
                  :key="group.id"
                  :label="group.name"
                  :required="group.required"
                >
                  <USelectMenu
                    v-model="item.specSelections[group.id]"
                    :items="buildSpecOptions(group)"
                    value-key="value"
                    label-key="label"
                    :portal="false"
                    :ui="{ content: 'z-[60] max-h-48 overflow-auto' }"
                    class="w-full"
                    :placeholder="$t('orders.createSpecHint')"
                    @update:model-value="() => handleSpecChange(item)"
                  />
                </UFormField>
              </div>
              <div v-else class="text-xs text-gray-500">
                {{ $t("orders.createSpecEmpty") }}
              </div>

              <div class="grid gap-3 sm:grid-cols-2">
                <UFormField :label="$t('orders.createSku')" required>
                  <UInput
                    :model-value="item.skuDisplay"
                    :placeholder="item.specGroups.length ? $t('orders.createSkuHint') : $t('orders.createSkuNoSpec')"
                    readonly
                    disabled
                  />
                </UFormField>
                <UFormField :label="$t('orders.createQty')" required>
                  <UInput v-model.number="item.qty" type="number" min="1" :placeholder="$t('orders.createQtyHint')" />
                </UFormField>
              </div>
              <div v-if="item.skuSummary" class="text-xs text-gray-500">{{ item.skuSummary }}</div>
              <div v-if="item.skuId && item.skuPrice !== null" class="text-xs text-gray-500">
                {{ $t("orders.createSkuPrice", { price: formatSkuPrice(item.skuPrice, item.skuCurrency) }) }}
              </div>
              <div v-if="item.skuId && item.skuPrice === null" class="text-xs text-amber-600">
                {{ $t("orders.createSkuPriceEmpty") }}
              </div>
              <div v-if="item.specGroups.length && isSpecSelectionComplete(item) && !item.skuId" class="text-xs text-rose-600">
                {{ $t("orders.createSkuEmpty") }}
              </div>
              <div v-if="item.specGroups.length && !item.skuList.length" class="text-xs text-gray-500">
                {{ $t("orders.createSkuNone") }}
              </div>
              <div v-if="item.skuTotal > item.skuList.length" class="text-xs text-amber-600">
                {{ item.skuScanInProgress ? $t("orders.createSkuScanning") : $t("orders.createSkuTooMany") }}
              </div>
              <div v-if="item.skuTotal > item.skuList.length && !item.skuId" class="flex items-center justify-between text-xs text-rose-600">
                <span>{{ $t("orders.createSkuFallback") }}</span>
                <UButton size="xs" variant="outline" :loading="item.skuScanInProgress" :disabled="item.skuScanInProgress" @click="retrySkuScan(item)">
                  {{ $t("orders.createSkuRetry") }}
                </UButton>
              </div>
            </div>
          </div>
          <UFormField :label="$t('orders.createNote')">
            <UTextarea v-model.trim="createForm.note" :rows="3" :placeholder="$t('orders.createNoteHint')" />
          </UFormField>
          <div class="space-y-3 rounded-lg border border-dashed border-gray-200/80 p-3 text-sm text-gray-600 dark:border-gray-700 dark:text-gray-300">
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">优惠券/礼品卡（占位）</span>
              <USwitch v-model="createBenefitEnabled" />
            </div>
            <div class="text-xs text-gray-500 dark:text-gray-400">
              提交后会在订单创建完成后生成审核记录，不会立即影响订单金额。
            </div>
            <div v-if="createBenefitEnabled" class="space-y-3">
              <UFormField label="类型" required>
                <USelect
                  v-model="createBenefitForm.type"
                  :items="[{ label: '优惠券', value: 'coupon' }, { label: '礼品卡', value: 'giftcard' }]"
                  class="w-full"
                />
              </UFormField>
              <UFormField label="券码/卡号" required>
                <UInput v-model.trim="createBenefitForm.code" placeholder="输入券码或礼品卡号" />
              </UFormField>
              <UFormField v-if="createBenefitForm.type === 'coupon'" label="优惠类型" required>
                <USelect
                  v-model="createBenefitForm.valueType"
                  :items="[{ label: '固定金额', value: 'amount' }, { label: '折扣百分比', value: 'percent' }]"
                  class="w-full"
                />
              </UFormField>
              <UFormField v-else label="抵扣类型">
                <UInput value="余额抵扣" readonly />
              </UFormField>
              <UFormField label="优惠值" required>
                <UInput
                  v-model.trim="createBenefitForm.value"
                  type="number"
                  min="0"
                  step="0.01"
                  :placeholder="createBenefitForm.valueType === 'percent' ? '例如：10（%）' : '例如：100（元）'"
                />
              </UFormField>
              <UFormField label="允许叠加">
                <USwitch v-model="createBenefitForm.stackingAllowed" />
              </UFormField>
              <UFormField label="备注（可选）">
                <UTextarea v-model.trim="createBenefitForm.note" :rows="2" placeholder="补充说明" />
              </UFormField>
            </div>
          </div>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="creating" @click="createOpen = false">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton color="primary" type="button" :loading="creating" @click="submitCreate">
            {{ $t("orders.createSubmit") }}
          </UButton>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="addressModalOpen"
      :title="addressModalMode === 'edit' ? $t('orders.createAddressEditTitle') : $t('orders.createAddressAddTitle')"
      :ui="{ content: 'sm:max-w-xl', footer: 'justify-end gap-2' }"
      :prevent-close="addressModalSaving"
    >
      <template #body>
        <div class="space-y-4 p-4">
          <div class="grid gap-4 sm:grid-cols-2">
            <UFormField :label="$t('orders.createRecipientName')" required>
              <UInput v-model.trim="createForm.recipientName" :placeholder="$t('orders.createRecipientNameHint')" />
            </UFormField>
            <UFormField :label="$t('orders.createRecipientPhone')" required>
              <UInput v-model.trim="createForm.recipientPhone" :placeholder="$t('orders.createRecipientPhoneHint')" />
            </UFormField>
          </div>
          <UFormField :label="$t('orders.createAddress1')" required>
            <UInput v-model.trim="createForm.address1" :placeholder="$t('orders.createAddress1Hint')" />
          </UFormField>
          <div class="grid gap-4 sm:grid-cols-2">
            <UFormField :label="$t('orders.createProvince')">
              <USelectMenu
                v-model="createForm.province"
                :items="provinceItems"
                value-key="value"
                label-key="label"
                :portal="false"
                :ui="{ content: 'z-[60] max-h-60 overflow-auto' }"
                searchable
                editable
                class="w-full"
                :placeholder="$t('orders.createProvinceHint')"
              />
            </UFormField>
            <UFormField :label="$t('orders.createCity')">
              <USelectMenu
                v-model="createForm.city"
                :items="cityItems"
                value-key="value"
                label-key="label"
                :portal="false"
                :ui="{ content: 'z-[60] max-h-60 overflow-auto' }"
                searchable
                editable
                class="w-full"
                :placeholder="$t('orders.createCityHint')"
              />
            </UFormField>
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <UFormField :label="$t('orders.createDistrict')">
              <USelectMenu
                v-model="createForm.district"
                :items="districtItems"
                value-key="value"
                label-key="label"
                :portal="false"
                :ui="{ content: 'z-[60] max-h-60 overflow-auto' }"
                searchable
                editable
                class="w-full"
                :placeholder="$t('orders.createDistrictHint')"
              />
            </UFormField>
            <UFormField :label="$t('orders.createAddress2')">
              <UInput v-model.trim="createForm.address2" :placeholder="$t('orders.createAddress2Hint')" />
            </UFormField>
          </div>
        </div>
      </template>
      <template #footer>
        <UButton variant="ghost" :disabled="addressModalSaving" @click="addressModalOpen = false">
          {{ $t("common.cancel") }}
        </UButton>
        <UButton color="primary" :loading="addressModalSaving" :disabled="addressModalSaving" @click="saveAddressForm">
          {{ $t("common.save") }}
        </UButton>
      </template>
    </UModal>

    <UModal
      v-model:open="deleteOpen"
      :title="$t('orders.deleteConfirmTitle')"
      :description="$t('orders.deleteConfirmDesc')"
      :prevent-close="deleting"
    >
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="deleting" @click="deleteOpen = false">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton color="error" type="button" :loading="deleting" @click="confirmDelete">
            {{ $t("orders.delete") }}
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useOrderApi, type OrderListQuery } from "~/composables/api/useOrder";
import { useSkuApi } from "~/composables/api/useSku";
import { useSpuApi } from "~/composables/api/useSpu";
import { useProductSpecApi, type ProductSpecGroup } from "~/composables/api/useProductSpec";
import type { OrderSummary } from "~/types/order";
import type { ProductSku } from "~/types/product/sku";
import { useCustomerService } from "~/composables/api/services/customerService";
import { useCustomerAddressService, type CustomerAddressDTO } from "~/composables/api/services/customerAddressService";
import { useChannelsApi } from "~/composables/useChannels";
import type { ChannelSummary } from "~/types/channels";

const { t, te } = useI18n();
const toast = useToastAlert();
const api = useOrderApi();
const skuApi = useSkuApi();
const spuApi = useSpuApi();
const specApi = useProductSpecApi();
const customerService = useCustomerService();
const customerAddressService = useCustomerAddressService();
const { listOrderChannels } = useChannelsApi();
const route = useRoute();
const router = useRouter();

const statusItems = [
  { value: "pending_payment", labelKey: "orders.statuses.pending_payment" },
  { value: "paid", labelKey: "orders.statuses.paid" },
  { value: "to_ship", labelKey: "orders.statuses.to_ship" },
  { value: "shipped", labelKey: "orders.statuses.shipped" },
  { value: "cancelled", labelKey: "orders.statuses.cancelled" },
  { value: "draft", labelKey: "orders.statuses.draft" },
] as const;

const normalizeStatus = (value: unknown) => {
  if (typeof value !== "string") return undefined;
  return statusItems.some((item) => item.value === value) ? value : undefined;
};

// 状态筛选：Nuxt UI Select 的 item value 不能是空字符串；用 undefined 表示“未选择/全部”
const selectedStatus = ref<string | undefined>(normalizeStatus(route.query.status));
const orderNoKeyword = ref("");
const filterCustomerId = ref("");
const filterCustomerSearchTerm = ref("");
const filterCustomerOptions = ref<Array<{ value: string; label: string; description?: string }>>([]);
const filterCustomerLoading = ref(false);

const statusOptions = computed(() =>
  statusItems.map((it) => ({
    value: it.value,
    label: t(it.labelKey),
  })),
);

type OrderRow = {
  orderId: string;
  orderNo: string;
  customerId?: string;
  customerName?: string;
  channel?: string;
  createdByType?: string;
  source: string;
  currency: string;
  amountMinor: number;
  status: string;
  createdAt: string;
};

// 表格列定义（v3 TanStack 风格；用 computed 以便 i18n 切换时表头刷新）
const columns = computed<TableColumn<OrderRow>[]>(() => [
  { accessorKey: "orderNo", header: t("orders.orderNo") },
  { accessorKey: "customerName", header: t("orders.customer") },
  { accessorKey: "source", header: t("orders.source") },
  {
    accessorKey: "amountMinor",
    header: t("orders.amount"),
    meta: { class: { td: "text-right" } },
  },
  { accessorKey: "status", header: t("orders.status") },
  { accessorKey: "createdAt", header: t("orders.createdAt") },
  { id: "actions", header: t("orders.actions") },
]);

const loading = ref(false);
const items = ref<OrderSummary[]>([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);

const deleteOpen = ref(false);
const deleting = ref(false);
const deleteTargetId = ref<string>("");
const createOpen = ref(false);
const creating = ref(false);
const customerLoading = ref(false);
const channelLoading = ref(false);
const customerSearchTerm = ref("");
const channelSearchTerm = ref("");
const customerOptions = ref<Array<{ value: string; label: string; description?: string }>>([]);
const channelOptions = ref<Array<{ value: string; label: string; description?: string }>>([]);
const spuCache = ref<SelectOption[]>([]);
const channelCache = ref<Record<string, ChannelSummary>>({});
const addressLoading = ref(false);
const addressOptions = ref<Array<{ value: string; label: string; description?: string }>>([]);
const addressList = ref<CustomerAddressDTO[]>([]);
const selectedAddressId = ref("");
const addressModalOpen = ref(false);
const addressModalSaving = ref(false);
const addressModalMode = ref<"create" | "edit">("create");

const selectedAddress = computed(() =>
  addressList.value.find((item) => item.id === selectedAddressId.value) || null,
);
const createForm = reactive({
  customerId: "",
  channel: "",
  items: [] as OrderItemRow[],
  recipientName: "",
  recipientPhone: "",
  address1: "",
  address2: "",
  province: "",
  city: "",
  district: "",
  note: "",
});
const createBenefitEnabled = ref(false);
const createBenefitForm = reactive({
  type: "coupon",
  code: "",
  valueType: "amount",
  value: "",
  stackingAllowed: false,
  note: "",
});
type SelectOption = { value: string; label: string; description?: string };
type OrderItemRow = {
  rowId: string;
  spuId: string;
  spuName: string;
  spuSearchTerm: string;
  spuOptions: SelectOption[];
  spuLoading: boolean;
  specGroups: ProductSpecGroup[];
  specSelections: Record<string, string>;
  skuId: string;
  skuOptions: SelectOption[];
  skuLoading: boolean;
  skuList: ProductSku[];
  skuSummary: string;
  skuDisplay: string;
  skuPrice?: number | null;
  skuCurrency?: string;
  skuTotal: number;
  skuScanInProgress: boolean;
  qty: number;
};
type RegionOption = {
  label: string;
  value: string;
  cities?: Array<{
    label: string;
    value: string;
    districts?: Array<{ label: string; value: string }>;
  }>;
};

const regionOptions: RegionOption[] = [
  {
    label: "北京",
    value: "北京市",
    cities: [
      {
        label: "北京市",
        value: "北京市",
        districts: [
          { label: "东城区", value: "东城区" },
          { label: "西城区", value: "西城区" },
          { label: "朝阳区", value: "朝阳区" },
          { label: "海淀区", value: "海淀区" },
        ],
      },
    ],
  },
  {
    label: "上海",
    value: "上海市",
    cities: [
      {
        label: "上海市",
        value: "上海市",
        districts: [
          { label: "黄浦区", value: "黄浦区" },
          { label: "徐汇区", value: "徐汇区" },
          { label: "浦东新区", value: "浦东新区" },
          { label: "静安区", value: "静安区" },
        ],
      },
    ],
  },
  {
    label: "广东省",
    value: "广东省",
    cities: [
      {
        label: "广州",
        value: "广州市",
        districts: [
          { label: "天河区", value: "天河区" },
          { label: "越秀区", value: "越秀区" },
          { label: "海珠区", value: "海珠区" },
        ],
      },
      {
        label: "深圳",
        value: "深圳市",
        districts: [
          { label: "南山区", value: "南山区" },
          { label: "福田区", value: "福田区" },
          { label: "宝安区", value: "宝安区" },
        ],
      },
    ],
  },
  {
    label: "浙江省",
    value: "浙江省",
    cities: [
      {
        label: "杭州",
        value: "杭州市",
        districts: [
          { label: "西湖区", value: "西湖区" },
          { label: "拱墅区", value: "拱墅区" },
          { label: "滨江区", value: "滨江区" },
        ],
      },
      {
        label: "宁波",
        value: "宁波市",
        districts: [
          { label: "海曙区", value: "海曙区" },
          { label: "鄞州区", value: "鄞州区" },
          { label: "江北区", value: "江北区" },
        ],
      },
    ],
  },
  {
    label: "江苏省",
    value: "江苏省",
    cities: [
      {
        label: "南京",
        value: "南京市",
        districts: [
          { label: "玄武区", value: "玄武区" },
          { label: "鼓楼区", value: "鼓楼区" },
          { label: "建邺区", value: "建邺区" },
        ],
      },
      {
        label: "苏州",
        value: "苏州市",
        districts: [
          { label: "姑苏区", value: "姑苏区" },
          { label: "工业园区", value: "工业园区" },
          { label: "吴中区", value: "吴中区" },
        ],
      },
    ],
  },
  {
    label: "四川省",
    value: "四川省",
    cities: [
      {
        label: "成都",
        value: "成都市",
        districts: [
          { label: "锦江区", value: "锦江区" },
          { label: "武侯区", value: "武侯区" },
          { label: "高新区", value: "高新区" },
        ],
      },
    ],
  },
];

const provinceItems = computed(() =>
  regionOptions.map((item) => ({ label: item.label, value: item.value })),
);

const cityItems = computed(() => {
  const province = regionOptions.find((item) => item.value === createForm.province);
  return (province?.cities || []).map((city) => ({ label: city.label, value: city.value }));
});

const districtItems = computed(() => {
  const province = regionOptions.find((item) => item.value === createForm.province);
  const city = province?.cities?.find((item) => item.value === createForm.city);
  return (city?.districts || []).map((district) => ({
    label: district.label,
    value: district.value,
  }));
});

const pageSizeItems = [
  { label: "10", value: 10 },
  { label: "20", value: 20 },
  { label: "50", value: 50 },
  { label: "100", value: 100 },
];

const customerNameById = ref<Record<string, string>>({});
const inflightCustomerIds = new Set<string>();

const rows = computed<OrderRow[]>(() =>
  (items.value || []).map((it) => ({
    orderId: it.orderId,
    orderNo: it.orderNo,
    customerId: it.customerId || (it as any).customer_id,
    customerName:
      (it.customerId || (it as any).customer_id)
        ? customerNameById.value[String(it.customerId || (it as any).customer_id)]
        : undefined,
    channel: it.channel || (it as any).channel,
    createdByType: it.createdByType || (it as any).created_by_type,
    source: sourceText(it.createdByType || (it as any).created_by_type),
    currency: it.amounts?.currency || "CNY",
    amountMinor: Number(it.amounts?.total || 0),
    status: it.status,
    createdAt: it.createdAt,
  })),
);

const statusText = (status: string) => {
  const key = `orders.statuses.${status}`;
  return te(key) ? t(key) : status || "-";
};

const sourceText = (createdByType?: string) => {
  const st = String(createdByType || "").trim();
  const key = st ? `orders.sources.${st}` : "";
  if (key && te(key)) return t(key);
  if (!st) return "-";
  return st;
};

// 获取状态颜色（语义色）
const getStatusColor = (status: string) => {
  const colorMap: Record<
    string,
    "warning" | "info" | "primary" | "success" | "neutral"
  > = {
    pending_payment: "warning",
    paid: "success",
    shipped: "primary",
    cancelled: "neutral",
    draft: "info",
  };
  return colorMap[status] || "neutral";
};

const customerDetailTo = (customerId: string) => ({
  path: "/customer",
  query: { customerId },
});

const hydrateCustomerNames = async (next: OrderSummary[]) => {
  const ids = Array.from(
    new Set(
      (next || [])
        .map((it) => String(it.customerId || (it as any).customer_id || "").trim())
        .filter(Boolean),
    ),
  );

  const missing = ids.filter(
    (id) => !customerNameById.value[id] && !inflightCustomerIds.has(id),
  );
  if (!missing.length) return;

  await Promise.allSettled(
    missing.map(async (id) => {
      inflightCustomerIds.add(id);
      try {
        const customer = await customerService.getCustomer(id);
        const name = String(customer?.name || "").trim();
        customerNameById.value[id] = name || id;
      } catch {
        customerNameById.value[id] = id;
      } finally {
        inflightCustomerIds.delete(id);
      }
    }),
  );
};

const goView = (orderId: string) => {
  if (!orderId) return;
  router.push(`/market/orders/${orderId}`);
};

const goEdit = (orderId: string) => {
  if (!orderId) return;
  router.push({ path: `/market/orders/${orderId}`, query: { mode: "edit" } });
};

const openDelete = (orderId: string) => {
  if (!orderId) return;
  deleteTargetId.value = orderId;
  deleteOpen.value = true;
};

const confirmDelete = async () => {
  const orderId = String(deleteTargetId.value || "").trim();
  if (!orderId) return;
  deleting.value = true;
  try {
    await api.cancelOrder(orderId, { reason: t("orders.deleteReason") });
    toast.add({ title: t("orders.deleteSuccess"), color: "success" });
    deleteOpen.value = false;
    deleteTargetId.value = "";
    await refresh();
  } catch (e: any) {
    toast.add({
      title: t("orders.deleteFailedTitle"),
      description: e?.message || t("orders.deleteFailedDesc"),
      color: "error",
    });
  } finally {
    deleting.value = false;
  }
};

const resetCreateForm = () => {
  createForm.customerId = "";
  createForm.channel = "";
  createForm.items = [buildItemRow()];
  createForm.recipientName = "";
  createForm.recipientPhone = "";
  createForm.address1 = "";
  createForm.address2 = "";
  createForm.province = "";
  createForm.city = "";
  createForm.district = "";
  createForm.note = "";
  createBenefitEnabled.value = false;
  createBenefitForm.type = "coupon";
  createBenefitForm.code = "";
  createBenefitForm.valueType = "amount";
  createBenefitForm.value = "";
  createBenefitForm.stackingAllowed = false;
  createBenefitForm.note = "";
};

const buildItemRow = (): OrderItemRow => ({
  rowId: buildRowId(),
  spuId: "",
  spuName: "",
  spuSearchTerm: "",
  spuOptions: [],
  spuLoading: false,
  specGroups: [],
  specSelections: {},
  skuId: "",
  skuOptions: [],
  skuLoading: false,
  skuList: [],
  skuSummary: "",
  skuDisplay: "",
  skuPrice: null,
  skuCurrency: "",
  skuTotal: 0,
  skuScanInProgress: false,
  qty: 1,
});

const addItemRow = () => {
  const row = buildItemRow();
  createForm.items.push(row);
  fetchSpus(row, "");
};

const removeItemRow = (rowId: string) => {
  const idx = createForm.items.findIndex((row) => row.rowId === rowId);
  if (idx >= 0) createForm.items.splice(idx, 1);
};

const buildRowId = () => {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `row-${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

if (!createForm.items.length) {
  createForm.items = [buildItemRow()];
}

const buildCustomerOptions = (items: Array<{ id: string; name?: string; phone?: string; email?: string }>) =>
  (items || []).map((item) => {
    const name = String(item.name || "").trim();
    const contact = String(item.phone || item.email || "").trim();
    const label = name && contact ? `${name} · ${contact}` : name || contact || "客户";
    return {
      value: item.id,
      label,
      description: undefined,
    };
  });

const buildAddressLabel = (addr: CustomerAddressDTO) => {
  const shipping = addr.shippingAddress || ({} as any);
  const name = String(shipping.recipientName || "").trim();
  const phone = String(shipping.recipientPhone || "").trim();
  const line = [shipping.province, shipping.city, shipping.district, shipping.address1, shipping.address2]
    .filter(Boolean)
    .join(" ");
  const title = [name, phone].filter(Boolean).join(" · ") || "收货地址";
  return { title, line };
};

const buildAddressOptions = (items: CustomerAddressDTO[]) =>
  (items || []).map((item) => {
    const labelInfo = buildAddressLabel(item);
    return {
      value: item.id,
      label: item.isDefault ? `${labelInfo.title}（默认）` : labelInfo.title,
      description: labelInfo.line || undefined,
    };
  });

const buildChannelOptions = (items: ChannelSummary[]) => {
  const map: Record<string, ChannelSummary> = {};
  const options = (items || []).map((item) => {
    map[item.id] = item;
    const name = String(item.name || "").trim();
    const platform = String(item.platform || "").trim();
    const storeId = String(item.storeId || "").trim();
    const label = name || platform || item.id;
    const description = [platform, storeId].filter(Boolean).join(" · ") || undefined;
    return {
      value: storeId || item.id,
      label,
      description,
    };
  });
  channelCache.value = map;
  return options;
};

const buildSpuOptions = (items: Array<{ id: string; name?: string; code?: string }>) =>
  (items || []).map((item) => {
    const name = String(item.name || "").trim();
    const code = String(item.code || "").trim();
    const label = name && code ? `${name} · ${code}` : name || code || item.id;
    return {
      value: item.id,
      label,
      description: code && name ? code : undefined,
    };
  });

const getSkuCode = (item: ProductSku | Record<string, any>) =>
  String((item as any)?.skuCode ?? (item as any)?.sku_code ?? "").trim();

const getSpuName = (item: ProductSku | Record<string, any>) =>
  String((item as any)?.spuName ?? (item as any)?.spu_name ?? "").trim();

const getSpecDisplay = (item: ProductSku | Record<string, any>) =>
  String((item as any)?.specDisplay ?? (item as any)?.spec_display ?? "").trim();

const getSkuSalePrice = (item: ProductSku | Record<string, any>) => {
  const raw = (item as any)?.salePrice ?? (item as any)?.sale_price;
  const num = typeof raw === "number" ? raw : Number(raw);
  return Number.isFinite(num) ? num : null;
};

const getSkuCurrency = (item: ProductSku | Record<string, any>) =>
  String((item as any)?.currency ?? (item as any)?.currency_code ?? "").trim();

const getSkuSpecs = (item: ProductSku | Record<string, any>) => {
  const specs = (item as any)?.specs;
  return Array.isArray(specs) ? specs : [];
};

const getSkuSpecGroupId = (spec: Record<string, any>) =>
  String(spec?.specId ?? spec?.spec_id ?? "").trim();

const getSkuSpecValueId = (spec: Record<string, any>) =>
  String(spec?.valueId ?? spec?.value_id ?? "").trim();

const buildSkuDisplay = (item: ProductSku) => {
  const specDisplay = getSpecDisplay(item);
  const spuName = getSpuName(item);
  if (spuName && specDisplay) return `${spuName} · ${specDisplay}`;
  if (specDisplay) return specDisplay;
  const skuCode = getSkuCode(item);
  if (spuName && skuCode) return `${spuName} · ${skuCode}`;
  if (skuCode) return skuCode;
  const shortId = item.id ? item.id.slice(0, 8) : "";
  return shortId ? `SKU-${shortId}` : "SKU";
};

const buildSkuOptions = (items: ProductSku[]) =>
  (items || []).map((item) => {
    const spuName = getSpuName(item);
    const skuLabel = buildSkuDisplay(item);
    const label = spuName ? `${spuName} · ${skuLabel}` : skuLabel;
    const description = getSpecDisplay(item) || undefined;
    return {
      value: item.id,
      label,
      description,
    };
  });

const buildSpecOptions = (group: ProductSpecGroup): SelectOption[] =>
  (group.options || []).map((opt) => ({
    value: opt.id,
    label: opt.name,
    description: opt.code || undefined,
  }));

const extractSkuTotal = (resp: any): number => {
  const raw = resp?.total ?? resp?.pagination?.total ?? resp?.meta?.total;
  const total = Number(raw || 0);
  return Number.isFinite(total) ? total : 0;
};

const fetchCustomers = async (keyword?: string) => {
  customerLoading.value = true;
  try {
    const normalized = String(keyword || "").trim();
    const sort = normalized ? undefined : "-lastOrderAt";
    const resp = await customerService.listCustomers({
      keyword: normalized || undefined,
      page: 1,
      pageSize: 20,
      sort,
    });
    customerOptions.value = buildCustomerOptions(resp?.data || []);
  } catch {
    customerOptions.value = [];
  } finally {
    customerLoading.value = false;
  }
};

const fetchFilterCustomers = async (keyword?: string) => {
  filterCustomerLoading.value = true;
  try {
    const normalized = String(keyword || "").trim();
    const sort = normalized ? undefined : "-lastOrderAt";
    const resp = await customerService.listCustomers({
      keyword: normalized || undefined,
      page: 1,
      pageSize: 20,
      sort,
    });
    filterCustomerOptions.value = buildCustomerOptions(resp?.data || []);
  } catch {
    filterCustomerOptions.value = [];
  } finally {
    filterCustomerLoading.value = false;
  }
};

const fetchAddresses = async (customerId: string) => {
  const id = String(customerId || "").trim();
  if (!id) {
    addressOptions.value = [];
    addressList.value = [];
    selectedAddressId.value = "";
    return;
  }
  addressLoading.value = true;
  try {
    const list = await customerAddressService.listCustomerAddresses(id);
    addressList.value = list || [];
    addressOptions.value = buildAddressOptions(addressList.value);
    const picked = addressList.value.find((item) => item.isDefault) || addressList.value[0] || null;
    if (picked) {
      selectedAddressId.value = picked.id;
    } else {
      selectedAddressId.value = "";
    }
  } catch {
    addressOptions.value = [];
    addressList.value = [];
    selectedAddressId.value = "";
  } finally {
    addressLoading.value = false;
  }
};

const resetAddressForm = () => {
  createForm.recipientName = "";
  createForm.recipientPhone = "";
  createForm.address1 = "";
  createForm.address2 = "";
  createForm.province = "";
  createForm.city = "";
  createForm.district = "";
};

const applyAddressToForm = (addr: CustomerAddressDTO) => {
  const shipping = addr.shippingAddress || ({} as any);
  createForm.recipientName = String(shipping.recipientName || "").trim();
  createForm.recipientPhone = String(shipping.recipientPhone || "").trim();
  createForm.address1 = String(shipping.address1 || "").trim();
  createForm.address2 = String(shipping.address2 || "").trim();
  createForm.province = String(shipping.province || "").trim();
  createForm.city = String(shipping.city || "").trim();
  createForm.district = String(shipping.district || "").trim();
};

const handleAddressChange = () => {
  // Selection only, actual editing handled in modal.
};

const openAddressModal = (mode: "create" | "edit") => {
  if (!createForm.customerId.trim()) {
    toast.add({ title: t("orders.createCustomerIdRequired"), color: "error" });
    return;
  }
  addressModalMode.value = mode;
  if (mode === "edit") {
    const selected = addressList.value.find((item) => item.id === selectedAddressId.value);
    if (selected) {
      applyAddressToForm(selected);
    } else {
      resetAddressForm();
    }
  } else {
    resetAddressForm();
  }
  addressModalOpen.value = true;
};

const validateAddressForm = () => {
  if (!createForm.recipientName.trim()) return t("orders.createRecipientNameRequired");
  if (!createForm.recipientPhone.trim()) return t("orders.createRecipientPhoneRequired");
  if (!createForm.address1.trim()) return t("orders.createAddress1Required");
  return "";
};

const saveAddressForm = async () => {
  if (!createForm.customerId.trim()) return;
  const message = validateAddressForm();
  if (message) {
    toast.add({ title: message, color: "error" });
    return;
  }
  addressModalSaving.value = true;
  try {
    const payload = {
      shippingAddress: {
        ...buildShippingAddress(),
        label: addressModalMode.value === "edit" ? "订单更新" : "订单新增",
      },
    };
    let saved: CustomerAddressDTO | null = null;
    if (addressModalMode.value === "edit" && selectedAddressId.value) {
      const picked = addressList.value.find((item) => item.id === selectedAddressId.value);
      const updated = await customerAddressService.updateCustomerAddress(createForm.customerId.trim(), selectedAddressId.value, {
        ...payload,
        isDefault: picked?.isDefault,
      });
      saved = updated;
      addressList.value = addressList.value.map((item) => (item.id === updated.id ? updated : item));
    } else {
      const isDefault = addressList.value.length === 0;
      const created = await customerAddressService.createCustomerAddress(createForm.customerId.trim(), {
        ...payload,
        isDefault,
      });
      saved = created;
      addressList.value = [created, ...addressList.value];
    }
    addressOptions.value = buildAddressOptions(addressList.value);
    if (saved?.id) {
      selectedAddressId.value = saved.id;
    }
    addressModalOpen.value = false;
  } catch (e: any) {
    toast.add({
      title: t("orders.createAddressSaveFailed"),
      description: e?.message || t("orders.createAddressSaveFailedDesc"),
      color: "error",
    });
  } finally {
    addressModalSaving.value = false;
  }
};

const fetchChannels = async (keyword?: string) => {
  channelLoading.value = true;
  try {
    const normalized = String(keyword || "").trim();
    const resp = await listOrderChannels({
      keyword: normalized || undefined,
      limit: 50,
    });
    channelOptions.value = buildChannelOptions(resp?.items || []);
    if (!createForm.channel && channelOptions.value.length === 1) {
      createForm.channel = channelOptions.value[0].value;
    }
  } catch {
    channelOptions.value = [];
  } finally {
    channelLoading.value = false;
  }
};

const fetchSpus = async (row: OrderItemRow, keyword?: string) => {
  const normalized = String(keyword || "").trim();
  if (!normalized && spuCache.value.length) {
    row.spuOptions = spuCache.value;
    row.spuLoading = false;
    return;
  }
  row.spuLoading = true;
  try {
    const resp = await spuApi.searchSpusForOrder({
      keyword: normalized || undefined,
      pageSize: 20,
    });
    row.spuOptions = buildSpuOptions(resp?.items || []);
    if (!normalized && row.spuOptions.length) {
      spuCache.value = row.spuOptions;
    }
  } catch {
    row.spuOptions = [];
  } finally {
    row.spuLoading = false;
  }
};

const handleSpuSearch = (row: OrderItemRow, value: string) => {
  row.spuSearchTerm = value;
  const timer = spuSearchTimers.get(row.rowId);
  if (timer) clearTimeout(timer);
  spuSearchTimers.set(
    row.rowId,
    setTimeout(() => {
      fetchSpus(row, value);
    }, 300),
  );
};

const handleSpuChange = async (row: OrderItemRow) => {
  const option = row.spuOptions.find((item) => item.value === row.spuId);
  row.spuName = option?.label || "";
  row.specGroups = [];
  row.specSelections = {};
  row.skuList = [];
  row.skuOptions = [];
  row.skuId = "";
  row.skuSummary = "";
  row.skuDisplay = "";
  row.skuPrice = null;
  row.skuCurrency = "";
  row.skuTotal = 0;
  row.skuScanInProgress = false;
  if (!row.spuId) return;
  row.skuLoading = true;
  try {
    const specResp = await specApi.listForOrder(row.spuId);
    row.specGroups = specResp?.groups || [];
    if (!row.specGroups.length) {
      row.skuList = [];
      row.skuOptions = [];
      row.skuTotal = 0;
      return;
    }
    const skuResp = await skuApi.listBySpuForOrder(row.spuId, { page: 1, pageSize: 200 });
    row.skuList = skuResp?.items || [];
    row.skuTotal = extractSkuTotal(skuResp) || row.skuList.length;
    updateSkuOptions(row);
  } catch {
    row.specGroups = [];
    row.skuList = [];
    row.skuOptions = [];
  } finally {
    row.skuLoading = false;
  }
};

const handleSpecChange = (row: OrderItemRow) => {
  updateSkuOptions(row);
};

const handleSkuChange = (row: OrderItemRow) => {
  const sku = row.skuList.find((item) => item.id === row.skuId);
  row.skuSummary = sku ? String(getSpecDisplay(sku) || getSkuCode(sku) || "").trim() : "";
  row.skuDisplay = sku ? buildSkuDisplay(sku) : "";
  row.skuPrice = sku ? getSkuSalePrice(sku) : null;
  row.skuCurrency = sku ? getSkuCurrency(sku) : "";
};

const filterSkusBySpecs = (
  skus: ProductSku[],
  groups: ProductSpecGroup[],
  selections: Record<string, string>,
) =>
  (skus || []).filter((sku) => {
    const specs = getSkuSpecs(sku);
    return groups.every((group) => {
      const selectedValue = selections[group.id];
      if (!selectedValue) return !group.required;
      const match = specs.find((spec: Record<string, any>) => getSkuSpecGroupId(spec) === group.id);
      return getSkuSpecValueId(match || {}) === selectedValue;
    });
  });

const findSkuBySpecs = async (row: OrderItemRow) => {
  if (row.skuScanInProgress || !row.spuId) return;
  row.skuScanInProgress = true;
  const selections = { ...row.specSelections };
  const pageSize = 200;
  try {
    let page = 1;
    const totalPages = row.skuTotal ? Math.ceil(row.skuTotal / pageSize) : 0;
    while (!totalPages || page <= totalPages) {
      const resp = await skuApi.listBySpuForOrder(row.spuId, { page, pageSize });
      const items = resp?.items || [];
      const matched = filterSkusBySpecs(items, row.specGroups, selections);
      if (matched.length) {
        row.skuOptions = buildSkuOptions(matched);
        if (matched.length === 1) {
          row.skuId = matched[0].id;
          row.skuSummary = String(getSpecDisplay(matched[0]) || getSkuCode(matched[0]) || "").trim();
          row.skuDisplay = buildSkuDisplay(matched[0]);
          row.skuPrice = getSkuSalePrice(matched[0]);
          row.skuCurrency = getSkuCurrency(matched[0]);
        }
        return;
      }
      if (items.length < pageSize) break;
      page += 1;
    }
  } catch {
    row.skuOptions = [];
  } finally {
    row.skuScanInProgress = false;
  }
};

const retrySkuScan = (row: OrderItemRow) => {
  if (!row.spuId) return;
  void findSkuBySpecs(row);
};

const updateSkuOptions = (row: OrderItemRow) => {
  if (!row.specGroups.length) {
    row.skuOptions = [];
    row.skuId = "";
    row.skuSummary = "";
    row.skuDisplay = "";
    row.skuPrice = null;
    row.skuCurrency = "";
    return;
  }
  const selected = row.specSelections || {};
  const requiredGroups = row.specGroups.filter((group) => group.required);
  const missingRequired = requiredGroups.some((group) => !selected[group.id]);
  if (missingRequired) {
    row.skuOptions = [];
    row.skuId = "";
    row.skuSummary = "";
    row.skuDisplay = "";
    row.skuPrice = null;
    row.skuCurrency = "";
    return;
  }

  const filtered = filterSkusBySpecs(row.skuList, row.specGroups, selected);
  row.skuOptions = buildSkuOptions(filtered);
  if (filtered.length === 1) {
    row.skuId = filtered[0].id;
    row.skuSummary = String(getSpecDisplay(filtered[0]) || getSkuCode(filtered[0]) || "").trim();
    row.skuDisplay = buildSkuDisplay(filtered[0]);
    row.skuPrice = getSkuSalePrice(filtered[0]);
    row.skuCurrency = getSkuCurrency(filtered[0]);
    return;
  }
  if (row.skuTotal > row.skuList.length) {
    void findSkuBySpecs(row);
    return;
  }
  if (!filtered.some((item) => item.id === row.skuId)) {
    row.skuId = "";
    row.skuSummary = "";
    row.skuDisplay = "";
    row.skuPrice = null;
    row.skuCurrency = "";
  }
};

const isSpecSelectionComplete = (row: OrderItemRow) => {
  if (!row.specGroups.length) return false;
  return row.specGroups.every((group) => !group.required || Boolean(row.specSelections[group.id]));
};

const blurActiveElement = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
};

const buildIdempotencyKey = () => {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `web-admin-${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

const formatSkuPrice = (price?: number | null, currency?: string) => {
  const amount = typeof price === "number" ? price : NaN;
  if (!Number.isFinite(amount)) return "";
  const code = String(currency || "CNY").trim() || "CNY";
  try {
    return new Intl.NumberFormat("zh-CN", { style: "currency", currency: code }).format(amount);
  } catch {
    return `${amount} ${code}`;
  }
};

const validateCreateForm = () => {
  if (!createForm.customerId.trim()) return t("orders.createCustomerIdRequired");
  if (!createForm.channel.trim()) return t("orders.createChannelRequired");
  if (!selectedAddressId.value) return t("orders.createAddressRequired");
  if (!createForm.items.length) return t("orders.createItemsRequired");
  for (const item of createForm.items) {
    if (!item.specGroups.length) return t("orders.createSkuNoSpec");
    if (!item.skuId.trim()) return t("orders.createSkuRequired");
    if (!item.qty || item.qty <= 0) return t("orders.createQtyRequired");
  }
  return "";
};

const buildShippingAddress = () => ({
  recipientName: createForm.recipientName.trim(),
  recipientPhone: createForm.recipientPhone.trim(),
  address1: createForm.address1.trim(),
  address2: createForm.address2.trim() || undefined,
  province: createForm.province.trim() || undefined,
  city: createForm.city.trim() || undefined,
  district: createForm.district.trim() || undefined,
  countryCode: "CN",
});

const submitCreate = async () => {
  const message = validateCreateForm();
  if (message) {
    toast.add({ title: message, color: "error" });
    return;
  }
  creating.value = true;
  try {
    const picked = addressList.value.find((item) => item.id === selectedAddressId.value);
    if (!picked?.shippingAddress) {
      throw new Error(t("orders.createAddressRequired"));
    }
    const shippingAddress = {
      ...picked.shippingAddress,
      countryCode: picked.shippingAddress.countryCode || "CN",
    };
    const created = await api.createOrder(
      {
        customerId: createForm.customerId.trim(),
        channel: createForm.channel.trim(),
        shippingAddressId: picked.id,
        items: createForm.items.map((item) => ({
          skuId: item.skuId.trim(),
          qty: Number(item.qty || 1),
        })),
        shippingAddress,
        note: createForm.note.trim() || undefined,
      },
      buildIdempotencyKey(),
    );
    if (createBenefitEnabled.value) {
      const code = String(createBenefitForm.code || "").trim();
      const value = Number(createBenefitForm.value || 0);
      if (!code) {
        toast.add({ title: "已创建订单，但未提交优惠权益：券码/卡号为空", color: "warning" });
      } else if (!Number.isFinite(value) || value <= 0) {
        toast.add({ title: "已创建订单，但未提交优惠权益：优惠值无效", color: "warning" });
      } else {
        try {
          await api.createBenefitReview(String(created.orderId || ""), {
            benefitType: createBenefitForm.type as "coupon" | "giftcard",
            benefitCode: code,
            valueType: createBenefitForm.valueType as "amount" | "percent" | "balance",
            value,
            currency: created.amounts?.currency || "CNY",
            stackingAllowed: createBenefitForm.stackingAllowed,
            note: createBenefitForm.note?.trim() || undefined,
          });
        } catch (err: any) {
          toast.add({
            title: "订单已创建，但优惠权益提交失败",
            description: err?.message || "请在订单详情里重新提交",
            color: "warning",
          });
        }
      }
    }
    toast.add({ title: t("orders.createSuccess"), color: "success" });
    createOpen.value = false;
    await refresh();
  } catch (e: any) {
    toast.add({
      title: t("orders.createFailedTitle"),
      description: e?.message || t("orders.createFailedDesc"),
      color: "error",
    });
  } finally {
    creating.value = false;
  }
};

const refresh = async () => {
  loading.value = true;
  try {
    const query: OrderListQuery = {
      page: page.value,
      pageSize: pageSize.value,
      status: selectedStatus.value || undefined,
      orderNo: orderNoKeyword.value.trim() || undefined,
      customerId: filterCustomerId.value.trim() || undefined,
    };
    const resp = await api.listOrders(query);
    items.value = resp.items || [];
    total.value = Number(resp.total || 0);
    await hydrateCustomerNames(items.value);
  } catch (e: any) {
    toast.add({
      title: t("orders.loadFailedTitle"),
      description: e?.message || t("orders.loadFailedDesc"),
      color: "error",
    });
  } finally {
    loading.value = false;
  }
};

const resetFilters = () => {
  orderNoKeyword.value = "";
  filterCustomerId.value = "";
  filterCustomerSearchTerm.value = "";
  selectedStatus.value = undefined;
  fetchFilterCustomers();
};

watch([page, pageSize], () => refresh());
watch(
  () => route.query.status,
  (value) => {
    const next = normalizeStatus(value);
    if (next !== selectedStatus.value) {
      selectedStatus.value = next;
    }
  },
);
watch(selectedStatus, (next) => {
  page.value = 1;
  const current = normalizeStatus(route.query.status);
  if (next !== current) {
    const query = { ...route.query } as Record<string, string>;
    if (next) {
      query.status = next;
    } else {
      delete query.status;
    }
    router.replace({ path: route.path, query });
  }
  refresh();
});
watch(filterCustomerId, () => {
  page.value = 1;
  refresh();
});
watch(createOpen, (open) => {
  if (!open) {
    blurActiveElement();
    resetCreateForm();
    customerSearchTerm.value = "";
    channelSearchTerm.value = "";
    addressOptions.value = [];
    addressList.value = [];
    selectedAddressId.value = "";
    addressModalOpen.value = false;
    addressModalMode.value = "create";
    return;
  }
  fetchCustomers(customerSearchTerm.value.trim());
  fetchChannels(channelSearchTerm.value.trim());
  fetchAddresses(createForm.customerId);
  if (!createForm.items.length) {
    createForm.items = [buildItemRow()];
  }
  createForm.items.forEach((row) => {
    fetchSpus(row, row.spuSearchTerm.trim());
  });
});

watch(
  () => createBenefitForm.type,
  (next) => {
    if (next === "giftcard") {
      createBenefitForm.valueType = "balance";
    } else if (createBenefitForm.valueType === "balance") {
      createBenefitForm.valueType = "amount";
    }
  },
);

watch(
  () => createForm.province,
  () => {
    createForm.city = "";
    createForm.district = "";
  },
);

watch(
  () => createForm.city,
  () => {
    createForm.district = "";
  },
);

watch(
  () => createForm.customerId,
  (value) => {
    selectedAddressId.value = "";
    resetAddressForm();
    if (value) {
      fetchAddresses(value);
    }
  },
);

let customerSearchTimer: ReturnType<typeof setTimeout> | null = null;
let channelSearchTimer: ReturnType<typeof setTimeout> | null = null;
let filterCustomerSearchTimer: ReturnType<typeof setTimeout> | null = null;
let orderNoSearchTimer: ReturnType<typeof setTimeout> | null = null;
const spuSearchTimers = new Map<string, ReturnType<typeof setTimeout>>();

watch(customerSearchTerm, (value) => {
  if (customerSearchTimer) clearTimeout(customerSearchTimer);
  customerSearchTimer = setTimeout(() => {
    fetchCustomers(value.trim());
  }, 300);
});

watch(channelSearchTerm, (value) => {
  if (channelSearchTimer) clearTimeout(channelSearchTimer);
  channelSearchTimer = setTimeout(() => {
    fetchChannels(value.trim());
  }, 300);
});

watch(filterCustomerSearchTerm, (value) => {
  if (filterCustomerSearchTimer) clearTimeout(filterCustomerSearchTimer);
  filterCustomerSearchTimer = setTimeout(() => {
    fetchFilterCustomers(value.trim());
  }, 300);
});

watch(orderNoKeyword, (value) => {
  if (orderNoSearchTimer) clearTimeout(orderNoSearchTimer);
  orderNoSearchTimer = setTimeout(() => {
    page.value = 1;
    if (!value.trim()) {
      refresh();
      return;
    }
    refresh();
  }, 300);
});

await refresh();
await fetchFilterCustomers();
</script>
