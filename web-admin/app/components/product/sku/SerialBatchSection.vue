<template>
  <section class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-primary-500">{{ $t('product.sku.serialPanel.kicker') }}</p>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.serialPanel.title') }}</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.serialPanel.subtitle') }}</p>
      </div>
      <div class="flex gap-2">
        <USelect
          v-model="filters.status"
          :items="statusItems"
          class="min-w-[160px]"
          :placeholder="$t('product.sku.serialPanel.filters.status')"
        />
        <UButton color="primary" icon="i-heroicons-plus" @click="openForm">
          {{ $t('common.add') }}
        </UButton>
        <UButton icon="i-heroicons-arrow-path" variant="ghost" :loading="loading" @click="fetchRecords">
          {{ $t('common.refresh') }}
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h4 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ $t('product.sku.serialPanel.table.title', { count: records.length }) }}
          </h4>
          <span class="text-xs text-gray-500">{{ $t('product.sku.serialPanel.table.hint') }}</span>
        </div>
      </template>
      <div v-if="loading">
        <USkeleton class="h-20 w-full" />
      </div>
      <div v-else-if="records.length" class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th class="px-4 py-2 text-left">{{ $t('product.sku.serialPanel.table.serial') }}</th>
              <th class="px-4 py-2 text-left">{{ $t('product.sku.serialPanel.table.batch') }}</th>
              <th class="px-4 py-2 text-left">{{ $t('product.sku.serialPanel.table.status') }}</th>
              <th class="px-4 py-2 text-left">{{ $t('product.sku.serialPanel.table.expires') }}</th>
              <th class="px-4 py-2 text-left">{{ $t('product.sku.serialPanel.table.created') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="record in records" :key="record.id">
              <td class="px-4 py-2 font-mono text-xs text-gray-900 dark:text-white">{{ record.serialNo }}</td>
              <td class="px-4 py-2 text-sm text-gray-500">{{ record.batchNo || '—' }}</td>
              <td class="px-4 py-2">
                <UBadge size="xs" :color="statusColor(record.status)">
                  {{ statusLabel(record.status) }}
                </UBadge>
              </td>
              <td class="px-4 py-2 text-sm text-gray-500">{{ formatDate(record.expiresAt) }}</td>
              <td class="px-4 py-2 text-sm text-gray-500">{{ formatDate(record.createdAt) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    <div v-else class="text-sm text-gray-500 dark:text-gray-400">
      {{ $t('product.sku.serialPanel.empty') }}
    </div>
  </UCard>

    <UModal
      v-model:open="formOpen"
      :title="$t('product.sku.serialPanel.form.title')"
      :description="$t('product.sku.serialPanel.subtitle')"
      :prevent-close="saving"
      :ui="{ content: 'max-w-3xl w-full max-h-[calc(100dvh-2rem)] overflow-hidden' }"
    >
      <template #body>
        <UForm id="sku-serial-form" :state="form" class="grid gap-4 md:grid-cols-2 p-1" @submit.prevent="handleSubmit">
          <UFormField :label="$t('product.sku.serialPanel.form.serial')" required>
            <UInput v-model="form.serialNo" />
          </UFormField>
          <UFormField :label="$t('product.sku.serialPanel.form.batch')">
            <UInput v-model="form.batchNo" />
          </UFormField>
          <UFormField :label="$t('product.sku.serialPanel.form.status')">
            <USelect v-model="form.status" :items="statusItems" class="w-full" />
          </UFormField>
          <UFormField :label="$t('product.sku.serialPanel.form.expires')">
            <UInput v-model="form.expiresAt" type="date" />
          </UFormField>
        </UForm>
      </template>

      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end sm:gap-3">
          <UButton color="neutral" variant="outline" :disabled="saving" @click="closeForm">
            {{ $t('common.cancel') }}
          </UButton>
          <UButton variant="ghost" :disabled="saving" @click="resetForm">
            {{ $t('common.reset') }}
          </UButton>
          <UButton type="submit" form="sku-serial-form" color="primary" :loading="saving">
            {{ $t('product.sku.serialPanel.form.submit') }}
          </UButton>
        </div>
      </template>
    </UModal>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n, useToast } from '#imports'
import { useProductSkuStore } from '~/stores/productSku'
import type { SkuSerialRecord } from '~/types/product/sku'

const props = defineProps<{ skuId: string }>()

const store = useProductSkuStore()
const { t } = useI18n()
const toast = useToast()

const loading = ref(false)
const saving = ref(false)
const formOpen = ref(false)
const filters = reactive<{ status: string | null }>({
  status: null,
})
const form = reactive({
  serialNo: '',
  batchNo: '',
  status: 'available',
  expiresAt: '',
})

const statusItems = computed(() => [
  { label: t('product.sku.serialPanel.status.available'), value: 'available' },
  { label: t('product.sku.serialPanel.status.allocated'), value: 'allocated' },
  { label: t('product.sku.serialPanel.status.consumed'), value: 'consumed' },
])

const records = computed<SkuSerialRecord[]>(() => store.serialRecords[props.skuId] ?? [])

const fetchRecords = async () => {
  if (!props.skuId) return
  loading.value = true
  try {
    await store.fetchSerialRecords(props.skuId, {
      status: filters.status || undefined,
      limit: 200,
    })
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (!props.skuId || !form.serialNo) {
    toast.add({ title: t('product.sku.serialPanel.errors.serialRequired'), color: 'amber' })
    return
  }
  saving.value = true
  try {
    await store.createSerialRecord(props.skuId, {
      serialNo: form.serialNo,
      batchNo: form.batchNo || undefined,
      status: form.status as SkuSerialRecord['status'],
      expiresAt: form.expiresAt ? new Date(form.expiresAt).toISOString() : undefined,
    })
    toast.add({ title: t('product.sku.serialPanel.toastSaved') })
    resetForm()
    closeForm()
  } catch (error: any) {
    toast.add({ title: t('common.error'), description: error?.message, color: 'red' })
  } finally {
    saving.value = false
  }
}

const resetForm = () => {
  form.serialNo = ''
  form.batchNo = ''
  form.status = 'available'
  form.expiresAt = ''
}

const openForm = () => {
  resetForm()
  formOpen.value = true
}

const closeForm = () => {
  if (typeof window !== 'undefined') {
    ;(document.activeElement as HTMLElement | null)?.blur?.()
  }
  formOpen.value = false
}

const formatDate = (value?: string) => {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}

const statusLabel = (status?: string) => {
  switch (status) {
    case 'allocated':
      return t('product.sku.serialPanel.status.allocated')
    case 'consumed':
      return t('product.sku.serialPanel.status.consumed')
    default:
      return t('product.sku.serialPanel.status.available')
  }
}

const statusColor = (status?: string) => {
  switch (status) {
    case 'allocated':
      return 'warning'
    case 'consumed':
      return 'gray'
    default:
      return 'success'
  }
}

watch(
  () => [props.skuId, filters.status],
  () => {
    fetchRecords()
  },
  { immediate: true },
)
</script>
