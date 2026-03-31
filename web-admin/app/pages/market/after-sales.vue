<template>
  <div class="space-y-4 px-4 py-6 sm:px-6">
    <UCard>
      <template #header>
        <div class="space-y-1">
          <h1 class="text-xl font-semibold">{{ t('afterSales.admin.title') }}</h1>
          <p class="text-sm text-gray-500">{{ t('afterSales.admin.subtitle') }}</p>
        </div>
      </template>

      <div class="grid gap-3 sm:grid-cols-4">
        <UCard class="border border-gray-200 dark:border-gray-700">
          <div class="text-xs text-gray-500">{{ t('afterSales.admin.dashboard.pending') }}</div>
          <div class="text-2xl font-semibold">{{ dashboard.pendingCount }}</div>
        </UCard>
        <UCard class="border border-gray-200 dark:border-gray-700">
          <div class="text-xs text-gray-500">{{ t('afterSales.admin.dashboard.processing') }}</div>
          <div class="text-2xl font-semibold">{{ dashboard.processingCount }}</div>
        </UCard>
        <UCard class="border border-gray-200 dark:border-gray-700">
          <div class="text-xs text-gray-500">{{ t('afterSales.admin.dashboard.completed') }}</div>
          <div class="text-2xl font-semibold">{{ dashboard.completedCount }}</div>
        </UCard>
        <UCard class="border border-gray-200 dark:border-gray-700">
          <div class="text-xs text-gray-500">{{ t('afterSales.admin.dashboard.rejected') }}</div>
          <div class="text-2xl font-semibold">{{ dashboard.rejectedCount }}</div>
        </UCard>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center gap-2">
          <UInput v-model.trim="filters.keyword" :placeholder="t('afterSales.admin.filters.keyword')" class="min-w-[220px]" />
          <USelect v-model="filters.status" :items="statusOptions" class="w-[180px]" />
          <USelect v-model="filters.caseType" :items="caseTypeOptions" class="w-[180px]" />
          <UButton color="neutral" variant="soft" :loading="loading" @click="loadCases">{{ t('afterSales.admin.actions.search') }}</UButton>
        </div>
      </template>

      <div v-if="!cases.length" class="text-sm text-gray-500">{{ t('afterSales.admin.list.empty') }}</div>
      <ul v-else class="space-y-2">
        <li v-for="item in cases" :key="item.id" class="rounded-md border border-gray-200 p-3 dark:border-gray-700">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-medium">{{ item.caseNo }}</span>
            <UBadge color="neutral" variant="subtle">{{ item.caseType }}</UBadge>
            <UBadge color="primary" variant="subtle">{{ item.status }}</UBadge>
          </div>
          <div class="mt-1 text-xs text-gray-500">{{ item.orderId }} / {{ item.orderItemId }} / {{ item.customerId }}</div>
          <div class="mt-2 flex flex-wrap gap-2">
            <UButton size="xs" color="neutral" variant="soft" @click="openDetail(item.id)">{{ t('afterSales.admin.actions.detail') }}</UButton>
            <UButton size="xs" @click="actionCase(item.id, 'accept')">{{ t('afterSales.admin.actions.accept') }}</UButton>
            <UButton size="xs" @click="actionCase(item.id, 'review')">{{ t('afterSales.admin.actions.review') }}</UButton>
            <UButton size="xs" color="success" @click="actionCase(item.id, 'approve')">{{ t('afterSales.admin.actions.approve') }}</UButton>
            <UButton size="xs" color="error" @click="openReject(item.id)">{{ t('afterSales.admin.actions.reject') }}</UButton>
            <UButton size="xs" color="neutral" @click="actionCase(item.id, 'complete')">{{ t('afterSales.admin.actions.complete') }}</UButton>
            <UButton size="xs" color="neutral" variant="outline" @click="actionCase(item.id, 'close')">{{ t('afterSales.admin.actions.close') }}</UButton>
          </div>
        </li>
      </ul>
    </UCard>

    <UModal v-model:open="detailOpen" :title="t('afterSales.admin.detail.title')">
      <template #body>
        <div v-if="detail" class="space-y-3">
          <div class="text-sm">
            <div>Case: {{ detail.case.caseNo }}</div>
            <div>Status: {{ detail.case.status }}</div>
            <div>Reason: {{ detail.case.reasonCode || '-' }}</div>
          </div>
          <div>
            <div class="mb-2 text-sm font-medium">{{ t('afterSales.admin.detail.timeline') }}</div>
            <ul class="space-y-2">
              <li v-for="(event, idx) in detail.timeline" :key="`${event.action}-${event.createdAt}-${idx}`" class="rounded border border-gray-200 p-2 text-sm dark:border-gray-700">
                <div>{{ event.action }}: {{ event.fromStatus || '-' }} → {{ event.toStatus }}</div>
                <div class="text-xs text-gray-500">{{ formatDate(event.createdAt) }}</div>
                <div v-if="event.note" class="text-xs">{{ event.note }}</div>
              </li>
            </ul>
          </div>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="rejectOpen" :title="t('afterSales.admin.reject.title')">
      <template #body>
        <div class="space-y-3">
          <UFormField :label="t('afterSales.admin.reject.reasonCode')" required>
            <UInput v-model.trim="rejectForm.reasonCode" :placeholder="t('afterSales.admin.reject.reasonCodePlaceholder')" />
          </UFormField>
          <UFormField :label="t('afterSales.admin.reject.note')">
            <UTextarea v-model.trim="rejectForm.note" :rows="3" />
          </UFormField>
          <div class="flex justify-end">
            <UButton color="error" :loading="acting" @click="submitReject">{{ t('afterSales.admin.actions.reject') }}</UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { useAfterSalesApi, type AfterSaleDetail, type AfterSaleSummary } from '~/composables/api/useAfterSales'

const { t } = useI18n()
const toast = useToastAlert()
const api = useAfterSalesApi()

const loading = ref(false)
const acting = ref(false)
const cases = ref<AfterSaleSummary[]>([])
const detail = ref<AfterSaleDetail | null>(null)
const detailOpen = ref(false)
const rejectOpen = ref(false)
const rejectCaseId = ref('')

const dashboard = reactive({
  pendingCount: 0,
  processingCount: 0,
  completedCount: 0,
  rejectedCount: 0,
})

const filters = reactive({
  status: '',
  caseType: '',
  keyword: '',
})

const rejectForm = reactive({
  reasonCode: '',
  note: '',
})

const statusOptions = computed(() => [
  { label: t('common.all'), value: '' },
  { label: 'pending', value: 'pending' },
  { label: 'accepted', value: 'accepted' },
  { label: 'reviewing', value: 'reviewing' },
  { label: 'approved', value: 'approved' },
  { label: 'rejected', value: 'rejected' },
  { label: 'completed', value: 'completed' },
  { label: 'closed', value: 'closed' },
])

const caseTypeOptions = computed(() => [
  { label: t('common.all'), value: '' },
  { label: 'refund_only', value: 'refund_only' },
  { label: 'return_refund', value: 'return_refund' },
  { label: 'exchange', value: 'exchange' },
])

const formatDate = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

const loadDashboard = async () => {
  try {
    const resp = await api.getDashboard()
    dashboard.pendingCount = Number(resp?.pendingCount || 0)
    dashboard.processingCount = Number(resp?.processingCount || 0)
    dashboard.completedCount = Number(resp?.completedCount || 0)
    dashboard.rejectedCount = Number(resp?.rejectedCount || 0)
  } catch {
    // no-op, avoid blocking list
  }
}

const loadCases = async () => {
  loading.value = true
  try {
    const resp = await api.listAdminCases({ status: filters.status || undefined, caseType: filters.caseType || undefined, keyword: filters.keyword || undefined, page: 1, pageSize: 30 })
    cases.value = resp.items || []
    await loadDashboard()
  } catch (error: any) {
    toast.add({ title: t('afterSales.admin.toast.listFailed'), description: error?.message || t('afterSales.admin.toast.retryLater'), color: 'error' })
  } finally {
    loading.value = false
  }
}

const openDetail = async (id: string) => {
  try {
    detail.value = await api.getAdminCase(id)
    detailOpen.value = true
  } catch (error: any) {
    toast.add({ title: t('afterSales.admin.toast.detailFailed'), description: error?.message || t('afterSales.admin.toast.retryLater'), color: 'error' })
  }
}

const actionCase = async (id: string, action: 'accept' | 'review' | 'approve' | 'complete' | 'close') => {
  acting.value = true
  try {
    if (action === 'accept') await api.acceptCase(id)
    if (action === 'review') await api.reviewCase(id)
    if (action === 'approve') await api.approveCase(id)
    if (action === 'complete') await api.completeCase(id)
    if (action === 'close') await api.closeCase(id)
    toast.add({ title: t('afterSales.admin.toast.actionSuccess'), color: 'success' })
    await loadCases()
  } catch (error: any) {
    toast.add({ title: t('afterSales.admin.toast.actionFailed'), description: error?.message || t('afterSales.admin.toast.retryLater'), color: 'error' })
  } finally {
    acting.value = false
  }
}

const openReject = (id: string) => {
  rejectCaseId.value = id
  rejectForm.reasonCode = ''
  rejectForm.note = ''
  rejectOpen.value = true
}

const submitReject = async () => {
  if (!rejectForm.reasonCode) {
    toast.add({ title: t('afterSales.admin.toast.rejectReasonRequired'), color: 'warning' })
    return
  }
  acting.value = true
  try {
    await api.rejectCase(rejectCaseId.value, { reasonCode: rejectForm.reasonCode, note: rejectForm.note })
    rejectOpen.value = false
    toast.add({ title: t('afterSales.admin.toast.actionSuccess'), color: 'success' })
    await loadCases()
  } catch (error: any) {
    toast.add({ title: t('afterSales.admin.toast.actionFailed'), description: error?.message || t('afterSales.admin.toast.retryLater'), color: 'error' })
  } finally {
    acting.value = false
  }
}

onMounted(() => {
  loadCases()
})
</script>
