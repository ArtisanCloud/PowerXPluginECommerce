<template>
  <UDrawer
    v-model:open="open"
    direction="right"
    :title="drawerTitle"
    :description="drawerDescription"
    portal="body"
    :ui="{
      width: 'max-w-3xl w-full',
      overlay: 'bg-black/30 dark:bg-black/50',
      container: 'flex h-full flex-col',
      body: 'flex-1 overflow-hidden p-0',
      header: 'sr-only'
    }"
  >
    <template #body>
      <div class="flex h-full flex-col">
        <div class="flex items-center justify-between border-b border-gray-200 px-6 py-4 dark:border-gray-800">
          <p class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('customer.directory.drawer.title') }}
          </p>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" size="sm" @click="open = false" />
        </div>
        <div class="flex-1 overflow-y-auto">
          <div v-if="props.loading" class="space-y-4 p-6">
            <USkeleton class="h-20 rounded-2xl" />
            <USkeleton class="h-32 rounded-2xl" />
            <USkeleton class="h-48 rounded-2xl" />
          </div>
          <div v-else-if="customer" class="space-y-6 p-6">
        <section
          class="rounded-2xl border border-gray-100 bg-white/90 p-5 shadow-sm ring-1 ring-gray-50 dark:border-gray-800 dark:bg-gray-900/80 dark:ring-gray-900/40"
        >
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="space-y-3">
              <div class="flex flex-wrap gap-2 text-xs text-gray-500 dark:text-gray-400">
                <UBadge v-if="typeLabel" size="xs" variant="soft" color="neutral">
                  {{ typeLabel }}
                </UBadge>
                <UBadge v-if="customer.membershipTier" size="xs" variant="soft" color="primary">
                  {{ customer.membershipTier }}
                </UBadge>
                <UBadge v-if="sourceLabel" size="xs" variant="soft">
                  {{ sourceLabel }}
                </UBadge>
              </div>
              <div>
                <p class="text-2xl font-semibold text-gray-900 dark:text-white">
                  {{ customer.name }}
                </p>
                <p class="mt-1 flex items-center gap-1 text-xs text-gray-400">
                  <UIcon name="i-heroicons-identification" class="h-4 w-4" />
                  {{ customer.id }}
                </p>
              </div>
              <div class="flex flex-wrap gap-2">
                <UBadge
                  v-for="chip in statusChips"
                  :key="chip.key"
                  :color="chip.color"
                  variant="soft"
                >
                  {{ chip.label }}
                </UBadge>
              </div>
            </div>
            <div class="flex flex-wrap gap-2">
              <UButton
                v-if="allowWrite"
                color="primary"
                variant="soft"
                size="sm"
                icon="i-heroicons-pencil-square"
                @click="handleEdit"
              >
                {{ t('customer.directory.actions.editRecord') }}
              </UButton>
              <UButton
                v-if="allowWrite"
                color="red"
                variant="ghost"
                size="sm"
                icon="i-heroicons-trash"
                @click="handleDelete"
              >
                {{ t('customer.directory.actions.deleteRecord') }}
              </UButton>
            </div>
          </div>

          <div class="mt-5 grid gap-4 text-sm sm:grid-cols-3">
            <div
              class="flex gap-3 rounded-xl border border-gray-100 bg-gray-50/80 px-3 py-2 text-gray-600 dark:border-gray-800 dark:bg-gray-800/40 dark:text-gray-300"
            >
              <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-white text-primary-500 shadow-sm dark:bg-gray-900">
                <UIcon name="i-heroicons-envelope" class="h-4 w-4" />
              </div>
              <div class="min-w-0">
                <p class="text-xs uppercase tracking-wide text-gray-400 dark:text-gray-500">
                  {{ t('customer.directory.table.contact') }}
                </p>
                <p class="truncate font-medium text-gray-900 dark:text-white">
                  {{ contactEmail || t('customer.directory.drawer.unknown') }}
                </p>
                <p v-if="contactPhone" class="text-xs text-gray-500 dark:text-gray-400">
                  {{ contactPhone }}
                </p>
              </div>
            </div>
            <div
              class="flex gap-3 rounded-xl border border-gray-100 bg-gray-50/80 px-3 py-2 text-gray-600 dark:border-gray-800 dark:bg-gray-800/40 dark:text-gray-300"
            >
              <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-white text-blue-500 shadow-sm dark:bg-gray-900">
                <UIcon name="i-heroicons-user" class="h-4 w-4" />
              </div>
              <div class="min-w-0">
                <p class="text-xs uppercase tracking-wide text-gray-400 dark:text-gray-500">
                  {{ t('customer.directory.drawer.accountManager') }}
                </p>
                <p class="truncate font-medium text-gray-900 dark:text-white">
                  {{ customer.accountManager || t('customer.directory.drawer.unassigned') }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('customer.directory.drawer.createdAt') }} · {{ formatDate(customer.createdAt) }}
                </p>
              </div>
            </div>
            <div
              class="flex gap-3 rounded-xl border border-gray-100 bg-gray-50/80 px-3 py-2 text-gray-600 dark:border-gray-800 dark:bg-gray-800/40 dark:text-gray-300"
            >
              <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-white text-amber-500 shadow-sm dark:bg-gray-900">
                <UIcon name="i-heroicons-tag" class="h-4 w-4" />
              </div>
              <div class="min-w-0">
                <p class="text-xs uppercase tracking-wide text-gray-400 dark:text-gray-500">
                  {{ t('customer.directory.table.tags') }}
                </p>
                <div v-if="tagPreview.length" class="mt-1 flex flex-wrap gap-1">
                  <UBadge v-for="tag in tagPreview" :key="tag" size="xs" variant="subtle">
                    {{ tag }}
                  </UBadge>
                  <UBadge v-if="extraTagCount > 0" size="xs" variant="subtle" color="neutral">
                    +{{ extraTagCount }}
                  </UBadge>
                </div>
                <p v-else class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('customer.directory.drawer.unknown') }}
                </p>
              </div>
            </div>
          </div>
        </section>

        <section class="grid gap-4 md:grid-cols-3">
          <div
            v-for="stat in highlightStats"
            :key="stat.key"
            class="rounded-2xl border border-gray-100 bg-white/80 p-4 shadow-sm dark:border-gray-800 dark:bg-gray-900/60"
          >
            <div
              class="flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400"
            >
              <UIcon :name="stat.icon" class="h-4 w-4 text-gray-400 dark:text-gray-500" />
              {{ stat.label }}
            </div>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ stat.value }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ stat.description }}
            </p>
          </div>
        </section>

        <section class="rounded-2xl border border-gray-100 bg-white/80 p-4 shadow-sm dark:border-gray-800 dark:bg-gray-900/70">
          <UTabs v-model="activeTab" :items="tabs">
            <template #content="{ item }">
              <div v-if="item.value === 'overview'" class="space-y-4">
                <div class="grid gap-4 md:grid-cols-2">
                  <div
                    v-for="row in overviewRows"
                    :key="row.key"
                    class="rounded-xl border border-gray-100 px-4 py-3 dark:border-gray-800"
                  >
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
                      {{ row.label }}
                    </p>
                    <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                      {{ row.value }}
                    </p>
                  </div>
                </div>
              </div>
              <div v-else-if="item.value === 'orders'" class="space-y-4 rounded-xl border border-gray-100 p-4 dark:border-gray-800">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t('customer.directory.drawer.lastOrder') }}
                    </p>
                    <p class="text-lg font-semibold text-gray-900 dark:text-white">
                      {{ formatDate(customer.lastOrderAt) }}
                    </p>
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t('customer.directory.drawer.lastOrderAmount', { amount: formattedAmount }) }}
                    </p>
                  </div>
                  <UBadge v-if="ordersTotal" size="xs" variant="soft" color="primary">
                    {{ t('customer.directory.status.total', { total: ordersTotal }) }}
                  </UBadge>
                </div>

                <div v-if="ordersLoading" class="space-y-2">
                  <USkeleton class="h-10 rounded-xl" />
                  <USkeleton class="h-10 rounded-xl" />
                </div>
                <div v-else-if="orders.length" class="space-y-2">
                  <div
                    v-for="order in orders"
                    :key="order.orderId"
                    class="flex items-center justify-between gap-3 rounded-xl border border-gray-100 px-3 py-2 text-sm text-gray-700 dark:border-gray-800 dark:text-gray-200"
                  >
                    <div class="min-w-0">
                      <p class="truncate font-medium text-gray-900 dark:text-white">
                        {{ order.orderNo }}
                      </p>
                      <p class="text-xs text-gray-500 dark:text-gray-400">
                        {{ formatDate(order.createdAt) }}
                      </p>
                    </div>
                    <div class="flex items-center gap-3">
                      <UBadge :color="statusColor(order.status)" variant="soft">
                        {{ statusLabel(order.status) }}
                      </UBadge>
                      <span class="text-sm font-semibold text-gray-900 dark:text-white">
                        ¥{{ formatOrderAmount(order.amounts?.total) }}
                      </span>
                      <UButton size="xs" variant="ghost" @click="openOrderDetail(order.orderId)">
                        查看
                      </UButton>
                    </div>
                  </div>
                </div>
                <p v-else class="text-sm text-gray-500 dark:text-gray-400">
                  {{ t('customer.directory.drawer.lastOrder') }}：{{ t('customer.directory.drawer.unknown') }}
                </p>
              </div>
              <div v-else-if="item.value === 'addresses'" class="space-y-3">
                <CustomerAddressBookPanel :customer-id="customer.id" :can-manage="allowWrite" />
              </div>
              <div
                v-else-if="item.value === 'afterSales'"
                class="rounded-xl border border-dashed border-gray-200 p-4 text-sm text-gray-500 dark:border-gray-800 dark:text-gray-400"
              >
                {{ t('customer.directory.drawer.afterSalesPlaceholder') }}
              </div>
              <div v-else-if="item.value === 'points'" class="grid gap-4 md:grid-cols-2">
                <div class="rounded-xl border border-gray-100 p-4 dark:border-gray-800">
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ t('customer.directory.drawer.tabs.points') }}
                  </p>
                  <p class="text-2xl font-semibold text-gray-900 dark:text-white">
                    {{ formatNumber(customer.points) }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    {{ t('customer.directory.drawer.growthValue', { value: formatNumber(customer.growthValue) }) }}
                  </p>
                </div>
                <div class="rounded-xl border border-gray-100 p-4 dark:border-gray-800">
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ t('customer.directory.drawer.retentionStatus') }}
                  </p>
                  <div class="mt-2 flex items-center gap-2">
                    <UBadge :color="retentionChip.color" variant="soft">
                      {{ retentionChip.label }}
                    </UBadge>
                    <span class="text-sm text-gray-500 dark:text-gray-400">
                      {{ retentionChip.hint || customer.membershipTier || t('customer.directory.drawer.unknown') }}
                    </span>
                  </div>
                </div>
              </div>
              <div v-else-if="item.value === 'entitlements'" class="space-y-4">
                <UAlert
                  v-if="membershipError"
                  color="red"
                  variant="soft"
                  :title="t('customer.directory.drawer.entitlements.loadError')"
                  :description="membershipError"
                />
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t('customer.directory.drawer.entitlements.tokens') }} / {{ t('customer.directory.drawer.entitlements.entitlements') }}
                  </p>
                  <div class="flex flex-wrap items-center gap-2" v-if="allowWrite">
                    <UButton size="xs" variant="soft" color="primary" @click="grantModalOpen = true">
                      手动发放权益
                    </UButton>
                    <UButton size="xs" variant="soft" color="primary" @click="tokenModalOpen = true">
                      调整代币余额
                    </UButton>
                  </div>
                </div>
                <div class="grid gap-4 md:grid-cols-2">
                  <div class="rounded-xl border border-gray-100 p-4 dark:border-gray-800">
                    <div class="flex items-center justify-between">
                      <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
                        {{ t('customer.directory.drawer.entitlements.tokens') }}
                      </p>
                      <UBadge v-if="tokenBalances.length" size="xs" variant="soft" color="primary">
                        {{ t('customer.directory.drawer.entitlements.tokenTypes', { count: tokenBalances.length }) }}
                      </UBadge>
                    </div>
                    <div v-if="membershipLoading" class="mt-3 space-y-2">
                      <USkeleton class="h-6 w-24 rounded" />
                      <USkeleton class="h-6 w-32 rounded" />
                    </div>
                    <div v-else-if="tokenBalances.length" class="mt-3 space-y-2">
                      <div
                        v-for="token in tokenBalances"
                        :key="token.tokenCode"
                        class="flex items-center justify-between rounded-lg border border-gray-100 px-3 py-2 text-sm text-gray-700 dark:border-gray-800 dark:text-gray-200"
                      >
                        <span class="font-medium">{{ token.tokenCode }}</span>
                        <span class="text-gray-500 dark:text-gray-400">
                          {{ formatNumber(token.balance) }}
                        </span>
                      </div>
                    </div>
                    <p v-else class="mt-3 text-sm text-gray-500 dark:text-gray-400">
                      {{ t('customer.directory.drawer.entitlements.emptyTokens') }}
                    </p>
                  </div>
                  <div class="rounded-xl border border-gray-100 p-4 dark:border-gray-800">
                    <div class="flex items-center justify-between">
                      <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
                        {{ t('customer.directory.drawer.entitlements.entitlements') }}
                      </p>
                      <UBadge v-if="entitlements.length" size="xs" variant="soft" color="primary">
                        {{ entitlements.length }}
                      </UBadge>
                    </div>
                    <div v-if="membershipLoading" class="mt-3 space-y-2">
                      <USkeleton class="h-6 w-36 rounded" />
                      <USkeleton class="h-6 w-28 rounded" />
                    </div>
                    <div v-else-if="entitlements.length" class="mt-3 space-y-3">
                      <div
                        v-for="entitlement in entitlements"
                        :key="entitlement.id"
                        class="rounded-lg border border-gray-100 px-3 py-2 text-sm text-gray-700 dark:border-gray-800 dark:text-gray-200"
                      >
                        <div class="flex items-center justify-between gap-2">
                          <span class="font-medium">{{ entitlement.serviceCode }}</span>
                          <div class="flex items-center gap-2">
                            <span class="text-gray-500 dark:text-gray-400">
                              {{ formatQuantity(entitlement.quantity) }}
                            </span>
                            <UButton
                              v-if="allowWrite"
                              size="xs"
                              color="red"
                              variant="ghost"
                              @click="revokeEntitlement(entitlement.id)"
                            >
                              回收
                            </UButton>
                          </div>
                        </div>
                        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                          {{ formatValidity(entitlement.validFrom, entitlement.validTo) }}
                        </p>
                      </div>
                    </div>
                    <p v-else class="mt-3 text-sm text-gray-500 dark:text-gray-400">
                      {{ t('customer.directory.drawer.entitlements.emptyEntitlements') }}
                    </p>
                  </div>
                </div>
              </div>
              <div
                v-else-if="item.value === 'notes'"
                class="rounded-xl border border-gray-100 bg-gray-50/60 p-4 text-sm text-gray-600 dark:border-gray-800 dark:bg-gray-800/40 dark:text-gray-300"
              >
                {{ notesContent }}
              </div>
              <div
                v-else-if="item.value === 'audit'"
                class="rounded-xl border border-dashed border-gray-200 p-4 text-sm text-gray-500 dark:border-gray-800 dark:text-gray-400"
              >
                {{ t('customer.directory.drawer.auditPlaceholder', { id: customer.id }) }}
              </div>
            </template>
          </UTabs>
        </section>
          </div>
          <div v-else class="py-16 text-center text-sm text-gray-500">
            {{ t('customer.directory.drawer.empty') }}
          </div>
        </div>
      </div>
    </template>
  </UDrawer>

  <UModal
    v-model:open="grantModalOpen"
    title="手动发放权益"
    description="为客户手动发放权益。"
    :prevent-close="true"
    :ui="{ content: 'max-w-3xl w-full', body: 'p-4 sm:p-5' }"
  >
    <template #body>
      <div class="space-y-4">
        <UFormField label="服务编码">
          <UInput v-model="grantForm.serviceCode" placeholder="service_code" />
        </UFormField>
        <UFormField label="数量">
          <UInput v-model.number="grantForm.quantity" type="number" min="-1" />
        </UFormField>
        <UFormField label="有效期（到期日）">
          <UInput v-model="grantForm.validTo" type="date" />
        </UFormField>
        <UFormField label="叠加策略">
          <USelect v-model="grantForm.stackPolicy" :items="[{ label: '叠加', value: 'stack' }, { label: '替换', value: 'replace' }]" />
        </UFormField>
        <UFormField label="原因">
          <UInput v-model="grantForm.reason" placeholder="可选" />
        </UFormField>
      </div>
    </template>
    <template #footer>
      <div class="flex justify-end gap-2">
        <UButton variant="ghost" @click="closeGrantModal">取消</UButton>
        <UButton color="primary" @click="submitGrant">确认发放</UButton>
      </div>
    </template>
  </UModal>

  <UModal
    v-model:open="tokenModalOpen"
    title="调整代币余额"
    description="为客户账户调整代币余额。"
    :prevent-close="true"
    :ui="{ content: 'max-w-3xl w-full', body: 'p-4 sm:p-5' }"
  >
    <template #body>
      <div class="space-y-4">
        <UFormField label="代币编码">
          <UInput v-model="tokenForm.tokenCode" placeholder="token_code" />
        </UFormField>
        <UFormField label="调整数量（正数增加/负数减少）">
          <UInput v-model.number="tokenForm.delta" type="number" />
        </UFormField>
        <UFormField label="原因">
          <UInput v-model="tokenForm.reason" placeholder="可选" />
        </UFormField>
      </div>
    </template>
    <template #footer>
      <div class="flex justify-end gap-2">
        <UButton variant="ghost" @click="closeTokenModal">取消</UButton>
        <UButton color="primary" @click="submitTokenAdjust">确认调整</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n, useToast, useRouter } from "#imports";
import CustomerAddressBookPanel from "~/components/customer/CustomerAddressBookPanel.vue";
import { useCustomerApi } from "~/composables/api/useCustomer";
import { useOrderApi } from "~/composables/api/useOrder";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import { useCustomerStore } from "~/stores/customer";
import type { Customer, CustomerEntitlement, CustomerTokenBalance } from "~/types/customer";
import type { OrderSummary } from "~/types/order";

const props = defineProps<{
  customer: Customer | null
  modelValue: boolean
  canManage?: boolean
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [boolean]
  edit: [Customer]
  delete: [Customer]
}>()

const { t, locale } = useI18n()
const toast = useToast()
const store = useCustomerStore()
const customerApi = useCustomerApi()
const orderApi = useOrderApi()
const membershipApi = useMembershipAdminApi()
const router = useRouter()

const open = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const drawerTitle = computed(() => props.customer?.name || t('customer.directory.drawer.title'))
const drawerDescription = computed(() => props.customer?.id || t('customer.directory.subtitle'))

const activeTab = ref("overview");
watch(
  () => props.customer?.id,
  () => {
    activeTab.value = "overview";
  },
);
const allowWrite = computed(() => Boolean(props.canManage))
const entitlements = ref<CustomerEntitlement[]>([])
const tokenBalances = ref<CustomerTokenBalance[]>([])
const membershipLoading = ref(false)
const membershipError = ref('')
const ordersLoading = ref(false)
const orders = ref<OrderSummary[]>([])
const ordersTotal = ref(0)
const grantModalOpen = ref(false)
const tokenModalOpen = ref(false)
const grantForm = reactive({
  serviceCode: '',
  quantity: 1,
  validTo: '',
  stackPolicy: 'stack',
  reason: '',
})
const tokenForm = reactive({
  tokenCode: '',
  delta: 0,
  reason: '',
})

const closeGrantModal = () => {
  document.activeElement?.blur()
  grantModalOpen.value = false
}

const closeTokenModal = () => {
  document.activeElement?.blur()
  tokenModalOpen.value = false
}

const isMasked = (field: string) => (props.customer ? store.isFieldMasked(props.customer, field) : false)

const getMaskedValue = (field: 'email' | 'phone') => {
  const current = props.customer
  if (!current) return ''
  const raw = current[field]
  if (!raw) return ''
  return isMasked(field) ? t('customer.directory.table.masked') : raw
}

const contactEmail = computed(() => getMaskedValue('email'))
const contactPhone = computed(() => getMaskedValue('phone'))

const typeLabel = computed(() => {
  const type = props.customer?.type
  if (!type) return ''
  const map: Record<string, string> = {
    individual: t('customer.directory.filters.typeOptions.individual'),
    enterprise: t('customer.directory.filters.typeOptions.enterprise'),
  }
  return map[type] ?? type
})

const sourceLabel = computed(() => {
  const source = props.customer?.source
  if (!source) return ''
  const map: Record<string, string> = {
    website: t('customer.directory.filters.sourceOptions.website'),
    offline: t('customer.directory.filters.sourceOptions.offline'),
    referral: t('customer.directory.filters.sourceOptions.referral'),
    miniapp: t('customer.directory.filters.sourceOptions.miniapp'),
  }
  return map[source] ?? source
})

const riskLabel = computed(() => {
  const risk = props.customer?.riskLevel
  if (!risk) return ''
  const map: Record<string, string> = {
    low: t('customer.directory.filters.riskOptions.low'),
    medium: t('customer.directory.filters.riskOptions.medium'),
    high: t('customer.directory.filters.riskOptions.high'),
  }
  return map[risk] ?? risk
})

const tagPreview = computed(() => (props.customer?.tags ?? []).slice(0, 3))
const extraTagCount = computed(() => {
  const total = props.customer?.tags?.length ?? 0
  const shown = tagPreview.value.length
  return total > shown ? total - shown : 0
})

const retentionChip = computed(() => {
  const status = props.customer?.membershipSnapshot?.retentionStatus
  if (!status) {
    return {
      label: t('customer.directory.drawer.unknown'),
      color: 'neutral',
      hint: '',
    }
  }
  const labelMap: Record<string, string> = {
    safe: t('customer.membership.segments.safe'),
    warning: t('customer.membership.segments.warning'),
    downgrade: t('customer.membership.segments.downgrade'),
  }
  const colorMap: Record<string, 'success' | 'warning' | 'error' | 'neutral'> = {
    safe: 'success',
    warning: 'warning',
    downgrade: 'error',
  }
  return {
    label: labelMap[status] ?? status,
    color: colorMap[status] ?? 'neutral',
    hint: props.customer?.membershipSnapshot?.tier || props.customer?.membershipTier || '',
  }
})

const formatDate = (value?: string | null) => {
  if (!value) return t('customer.directory.drawer.unknown')
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  const localeCode = locale.value === 'en' ? 'en-US' : 'zh-CN'
  return new Intl.DateTimeFormat(localeCode, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(parsed)
}

const formatNumber = (value?: number | null) => {
  if (value === undefined || value === null) return '0'
  const localeCode = locale.value === 'en' ? 'en-US' : 'zh-CN'
  return new Intl.NumberFormat(localeCode, { maximumFractionDigits: 0 }).format(value)
}

const formatCurrency = (value?: number | null) => {
  if (value === undefined || value === null) return '0'
  const localeCode = locale.value === 'en' ? 'en-US' : 'zh-CN'
  return new Intl.NumberFormat(localeCode, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(value)
}

const formattedAmount = computed(() => formatCurrency(props.customer?.lastOrderAmount))
const formatOrderAmount = (value?: number | null) => formatCurrency(value ? value / 100 : 0)
const formatQuantity = (value?: number | null) => {
  if (value === undefined || value === null) {
    return t('customer.directory.drawer.unknown')
  }
  if (value < 0) {
    return t('customer.directory.drawer.entitlements.unlimited')
  }
  return formatNumber(value)
}

const formatValidity = (from?: string, to?: string) => {
  const fromLabel = from ? formatDate(from) : t('customer.directory.drawer.unknown')
  const toLabel = to ? formatDate(to) : t('customer.directory.drawer.entitlements.noExpiry')
  return t('customer.directory.drawer.entitlements.validityRange', { from: fromLabel, to: toLabel })
}

const statusLabel = (status?: string) => {
  switch (status) {
    case 'active':
      return t('customer.directory.table.statusActive')
    case 'blocked':
      return t('customer.directory.table.statusBlocked')
    default:
      return t('customer.directory.table.statusInactive')
  }
}

const statusColor = (status?: string): 'success' | 'error' | 'neutral' => {
  switch (status) {
    case 'active':
      return 'success'
    case 'blocked':
      return 'error'
    default:
      return 'neutral'
  }
}

const statusChips = computed(() => {
  const chips: Array<{ key: string; label: string; color: 'success' | 'warning' | 'error' | 'neutral' }> = []
  const current = props.customer
  if (!current) return chips
  if (current.status) {
    chips.push({
      key: 'status',
      label: statusLabel(current.status),
      color: statusColor(current.status),
    })
  }
  if (current.riskLevel) {
    chips.push({
      key: 'risk',
      label: `${t('customer.directory.filters.risk')} · ${riskLabel.value || current.riskLevel}`,
      color: 'warning',
    })
  }
  if (current.membershipSnapshot?.retentionStatus) {
    chips.push({
      key: 'retention',
      label: `${t('customer.directory.drawer.retentionStatus')} · ${retentionChip.value.label}`,
      color: retentionChip.value.color,
    })
  }
  return chips
})

const regionDisplay = computed(() => {
  const region = props.customer?.region || props.customer?.country
  return region || t('customer.directory.drawer.unknown')
})

const overviewRows = computed(() => {
  const current = props.customer
  if (!current) return []
  return [
    {
      key: 'source',
      label: t('customer.directory.filters.source'),
      value: sourceLabel.value || current.source || t('customer.directory.drawer.unknown'),
    },
    {
      key: 'type',
      label: t('customer.directory.filters.type'),
      value: typeLabel.value || t('customer.directory.drawer.unknown'),
    },
    {
      key: 'region',
      label: t('customer.directory.filters.region'),
      value: regionDisplay.value,
    },
    {
      key: 'risk',
      label: t('customer.directory.filters.risk'),
      value: riskLabel.value || current.riskLevel || t('customer.directory.drawer.unknown'),
    },
    {
      key: 'created',
      label: t('customer.directory.drawer.createdAt'),
      value: formatDate(current.createdAt),
    },
    {
      key: 'lastOrder',
      label: t('customer.directory.drawer.lastOrder'),
      value: formatDate(current.lastOrderAt),
    },
  ]
})

const highlightStats = computed(() => {
  const current = props.customer
  if (!current) return []
  return [
    {
      key: 'orders',
      label: t('customer.directory.drawer.tabs.orders'),
      icon: 'i-heroicons-shopping-bag',
      value: current.lastOrderAmount ? `¥${formatCurrency(current.lastOrderAmount)}` : '—',
      description: `${t('customer.directory.drawer.lastOrder')} · ${formatDate(current.lastOrderAt)}`,
    },
    {
      key: 'points',
      label: t('customer.directory.drawer.tabs.points'),
      icon: 'i-heroicons-sparkles',
      value: formatNumber(current.points),
      description: t('customer.directory.drawer.growthValue', {
        value: formatNumber(current.growthValue),
      }),
    },
    {
      key: 'retention',
      label: t('customer.directory.drawer.retentionStatus'),
      icon: 'i-heroicons-shield-check',
      value: retentionChip.value.label,
      description: `${t('customer.directory.table.tier')} · ${
        current.membershipTier || t('customer.directory.drawer.unknown')
      }`,
    },
  ]
})

const notesContent = computed(() => {
  const note = props.customer?.notes ?? props.customer?.metadata?.notes
  if (!note) {
    return t('customer.directory.drawer.notesPlaceholder')
  }
  return typeof note === 'string' ? note : JSON.stringify(note, null, 2)
})

const handleEdit = () => {
  if (props.customer) {
    emit('edit', props.customer)
  }
}

const handleDelete = () => {
  if (props.customer) {
    emit('delete', props.customer)
  }
}

const openOrderDetail = (orderId: string) => {
  if (!orderId) return
  router.push(`/market/orders/${orderId}`)
}

const resetGrantForm = () => {
  grantForm.serviceCode = ''
  grantForm.quantity = 1
  grantForm.validTo = ''
  grantForm.stackPolicy = 'stack'
  grantForm.reason = ''
}

const resetTokenForm = () => {
  tokenForm.tokenCode = ''
  tokenForm.delta = 0
  tokenForm.reason = ''
}

const submitGrant = async () => {
  if (!props.customer?.id) return
  if (!grantForm.serviceCode.trim()) {
    toast.add({ title: '请填写服务编码', color: 'red' })
    return
  }
  if (!grantForm.quantity || grantForm.quantity === 0) {
    toast.add({ title: '数量必须不为 0', color: 'red' })
    return
  }
  try {
    await membershipApi.grantEntitlement({
      customerId: props.customer.id,
      serviceCode: grantForm.serviceCode.trim(),
      quantity: Number(grantForm.quantity),
      validTo: grantForm.validTo ? new Date(grantForm.validTo).toISOString() : undefined,
      stackPolicy: grantForm.stackPolicy,
      reason: grantForm.reason.trim() || undefined,
    })
    toast.add({ title: '权益已发放', color: 'green' })
    closeGrantModal()
    resetGrantForm()
    loadMembershipAssets()
  } catch (error: any) {
    toast.add({ title: error?.message || '发放失败', color: 'red' })
  }
}

const submitTokenAdjust = async () => {
  if (!props.customer?.id) return
  if (!tokenForm.tokenCode.trim()) {
    toast.add({ title: '请填写代币编码', color: 'red' })
    return
  }
  if (!tokenForm.delta || tokenForm.delta === 0) {
    toast.add({ title: '调整数量必须不为 0', color: 'red' })
    return
  }
  try {
    await membershipApi.adjustToken({
      customerId: props.customer.id,
      tokenCode: tokenForm.tokenCode.trim(),
      delta: Number(tokenForm.delta),
      reason: tokenForm.reason.trim() || undefined,
    })
    toast.add({ title: '代币余额已调整', color: 'green' })
    closeTokenModal()
    resetTokenForm()
    loadMembershipAssets()
  } catch (error: any) {
    toast.add({ title: error?.message || '调整失败', color: 'red' })
  }
}

const revokeEntitlement = async (entitlementId: string) => {
  if (!props.customer?.id || !entitlementId) return
  if (!confirm('确定要回收该权益吗？')) return
  try {
    await membershipApi.revokeEntitlement({ entitlementId })
    toast.add({ title: '权益已回收', color: 'green' })
    loadMembershipAssets()
  } catch (error: any) {
    toast.add({ title: error?.message || '回收失败', color: 'red' })
  }
}

const loadMembershipAssets = async () => {
  if (!props.customer?.id) {
    entitlements.value = []
    tokenBalances.value = []
    return
  }
  try {
    membershipLoading.value = true
    membershipError.value = ''
    const [entitlementResp, tokenResp] = await Promise.all([
      customerApi.getCustomerEntitlements(props.customer.id),
      customerApi.getCustomerTokenBalances(props.customer.id),
    ])
    entitlements.value = entitlementResp?.items ?? []
    tokenBalances.value = tokenResp?.items ?? []
  } catch (error: any) {
    console.error(error)
    entitlements.value = []
    tokenBalances.value = []
    membershipError.value = error?.message || t('customer.directory.drawer.entitlements.loadError')
  } finally {
    membershipLoading.value = false
  }
}

const loadCustomerOrders = async () => {
  if (!props.customer?.id) {
    orders.value = []
    ordersTotal.value = 0
    return
  }
  try {
    ordersLoading.value = true
    const resp = await orderApi.listOrders({
      customerId: props.customer.id,
      page: 1,
      pageSize: 5,
    })
    orders.value = resp?.items ?? []
    ordersTotal.value = resp?.total ?? 0
  } catch (error) {
    orders.value = []
    ordersTotal.value = 0
    console.error(error)
  } finally {
    ordersLoading.value = false
  }
}

const tabs = computed(() => [
  { label: t("customer.directory.drawer.tabs.overview"), value: "overview" },
  { label: t("customer.directory.drawer.tabs.orders"), value: "orders" },
  { label: locale.value === "en" ? "Addresses" : "收货地址", value: "addresses" },
  { label: t("customer.directory.drawer.tabs.afterSales"), value: "afterSales" },
  { label: t("customer.directory.drawer.tabs.points"), value: "points" },
  { label: t("customer.directory.drawer.tabs.entitlements"), value: "entitlements" },
  { label: t("customer.directory.drawer.tabs.notes"), value: "notes" },
  { label: t("customer.directory.drawer.tabs.audit"), value: "audit" },
]);

watch(
  () => props.customer?.id,
  () => {
    activeTab.value = 'overview'
  },
)

watch(
  () => [open.value, props.customer?.id],
  ([isOpen, id]) => {
    if (!isOpen || !id) {
      entitlements.value = []
      tokenBalances.value = []
      membershipError.value = ''
      orders.value = []
      ordersTotal.value = 0
      return
    }
    loadMembershipAssets()
    loadCustomerOrders()
  },
  { immediate: true },
)
</script>
