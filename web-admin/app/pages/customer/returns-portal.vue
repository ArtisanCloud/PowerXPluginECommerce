<template>
  <div class="space-y-4 px-4 py-6 sm:px-6">
    <UCard>
      <template #header>
        <div class="space-y-1">
          <h1 class="text-xl font-semibold">{{ t('afterSales.customer.title') }}</h1>
          <p class="text-sm text-gray-500">{{ t('afterSales.customer.subtitle') }}</p>
        </div>
      </template>

      <div class="grid gap-3 sm:grid-cols-2">
        <UFormField :label="t('afterSales.customer.form.orderId')" required>
          <UInput v-model.trim="form.orderId" :placeholder="t('afterSales.customer.form.orderIdPlaceholder')" />
        </UFormField>
        <UFormField :label="t('afterSales.customer.form.orderItemId')" required>
          <UInput v-model.trim="form.orderItemId" :placeholder="t('afterSales.customer.form.orderItemIdPlaceholder')" />
        </UFormField>
        <UFormField :label="t('afterSales.customer.form.caseType')" required>
          <USelect v-model="form.caseType" :items="caseTypeOptions" class="w-full" />
        </UFormField>
        <UFormField :label="t('afterSales.customer.form.reasonCode')" required>
          <UInput v-model.trim="form.reasonCode" :placeholder="t('afterSales.customer.form.reasonCodePlaceholder')" />
        </UFormField>
      </div>

      <UFormField :label="t('afterSales.customer.form.reasonDetail')" class="mt-3">
        <UTextarea v-model.trim="form.reasonDetail" :rows="3" :placeholder="t('afterSales.customer.form.reasonDetailPlaceholder')" />
      </UFormField>

      <div class="mt-4 flex gap-2">
        <UButton :loading="submitting" @click="submitCase">{{ t('afterSales.customer.actions.submit') }}</UButton>
        <UButton color="neutral" variant="soft" :loading="loading" @click="loadCases">{{ t('afterSales.customer.actions.refresh') }}</UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-base font-semibold">{{ t('afterSales.customer.list.title') }}</h2>
          <span class="text-xs text-gray-500">{{ t('afterSales.customer.list.total', { count: cases.length }) }}</span>
        </div>
      </template>

      <div v-if="!cases.length" class="text-sm text-gray-500">{{ t('afterSales.customer.list.empty') }}</div>
      <ul v-else class="space-y-2">
        <li v-for="item in cases" :key="item.id" class="rounded-md border border-gray-200 p-3 dark:border-gray-700">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-medium">{{ item.caseNo || item.id }}</span>
            <UBadge color="neutral" variant="subtle">{{ item.caseType }}</UBadge>
            <UBadge color="primary" variant="subtle">{{ item.status }}</UBadge>
          </div>
          <div class="mt-1 text-xs text-gray-500">
            {{ t('afterSales.customer.list.orderItem', { orderId: item.orderId, orderItemId: item.orderItemId }) }}
          </div>
          <div class="mt-2">
            <UButton size="xs" color="neutral" variant="soft" :loading="detailLoading && detailCaseId === item.id" @click="openDetail(item.id)">
              {{ t('afterSales.customer.actions.viewDetail') }}
            </UButton>
          </div>
        </li>
      </ul>
    </UCard>

    <UModal v-model:open="detailOpen" :title="t('afterSales.customer.detail.title')">
      <template #body>
        <div v-if="selectedCase" class="space-y-4">
          <div class="rounded-md border border-gray-200 p-3 text-sm dark:border-gray-700">
            <div><span class="text-gray-500">CaseNo:</span> {{ selectedCase.case.caseNo || selectedCase.case.id }}</div>
            <div><span class="text-gray-500">Status:</span> {{ selectedCase.case.status }}</div>
          </div>

          <div>
            <div class="mb-2 text-sm font-medium">{{ t('afterSales.customer.detail.timeline') }}</div>
            <ul v-if="selectedCase.timeline?.length" class="space-y-2">
              <li v-for="(event, idx) in selectedCase.timeline" :key="`${event.action}-${event.createdAt}-${idx}`" class="rounded-md border border-gray-200 p-2 text-sm dark:border-gray-700">
                <div class="font-medium">{{ event.action }} → {{ event.toStatus || '-' }}</div>
                <div class="text-xs text-gray-500">{{ formatDate(event.createdAt) }}</div>
                <div v-if="event.note" class="mt-1 text-xs">{{ event.note }}</div>
              </li>
            </ul>
            <div v-else class="text-sm text-gray-500">{{ t('afterSales.customer.detail.emptyTimeline') }}</div>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { useAfterSalesApi, type AfterSaleCaseType, type AfterSaleDetail, type AfterSaleSummary } from '~/composables/api/useAfterSales'

const { t } = useI18n()
const toast = useToastAlert()
const api = useAfterSalesApi()

const loading = ref(false)
const submitting = ref(false)
const detailLoading = ref(false)
const detailOpen = ref(false)
const detailCaseId = ref('')

const cases = ref<AfterSaleSummary[]>([])
const selectedCase = ref<AfterSaleDetail | null>(null)

const form = reactive({
  orderId: '',
  orderItemId: '',
  caseType: 'refund_only' as AfterSaleCaseType,
  reasonCode: '',
  reasonDetail: '',
})

const caseTypeOptions = computed(() => [
  { label: t('afterSales.customer.caseType.refundOnly'), value: 'refund_only' },
  { label: t('afterSales.customer.caseType.returnRefund'), value: 'return_refund' },
  { label: t('afterSales.customer.caseType.exchange'), value: 'exchange' },
])

const formatDate = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

const loadCases = async () => {
  loading.value = true
  try {
    const resp = await api.listMiniAppCases({ page: 1, pageSize: 20 })
    cases.value = resp.items || []
  } catch (error: any) {
    toast.add({ title: t('afterSales.customer.toast.listFailed'), description: error?.message || t('afterSales.customer.toast.retryLater'), color: 'error' })
  } finally {
    loading.value = false
  }
}

const submitCase = async () => {
  if (!form.orderId || !form.orderItemId || !form.reasonCode) {
    toast.add({ title: t('afterSales.customer.toast.required'), color: 'warning' })
    return
  }

  submitting.value = true
  try {
    await api.createMiniAppCase({
      orderId: form.orderId,
      orderItemId: form.orderItemId,
      caseType: form.caseType,
      reasonCode: form.reasonCode,
      reasonDetail: form.reasonDetail || undefined,
    })
    toast.add({ title: t('afterSales.customer.toast.submitSuccess'), color: 'success' })
    form.reasonCode = ''
    form.reasonDetail = ''
    await loadCases()
  } catch (error: any) {
    toast.add({ title: t('afterSales.customer.toast.submitFailed'), description: error?.message || t('afterSales.customer.toast.retryLater'), color: 'error' })
  } finally {
    submitting.value = false
  }
}

const openDetail = async (id: string) => {
  detailLoading.value = true
  detailCaseId.value = id
  try {
    selectedCase.value = await api.getMiniAppCase(id)
    detailOpen.value = true
  } catch (error: any) {
    toast.add({ title: t('afterSales.customer.toast.detailFailed'), description: error?.message || t('afterSales.customer.toast.retryLater'), color: 'error' })
  } finally {
    detailLoading.value = false
    detailCaseId.value = ''
  }
}

onMounted(() => {
  loadCases()
})
</script>
