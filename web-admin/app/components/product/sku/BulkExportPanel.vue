<template>
  <section class="flex flex-col rounded-xl border border-gray-200/70 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-gray-900">
    <header>
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('product.sku.export.title') }}
      </h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('product.sku.export.subtitle') }}
      </p>
    </header>

    <div class="mt-4 grid flex-1 gap-4 md:grid-cols-2">
      <UFormField :label="t('product.sku.export.format')" class="w-full">
        <USelect v-model="format" :options="formatOptions" />
      </UFormField>
      <UFormField :label="t('product.sku.export.limit')" class="w-full">
        <UInput v-model.number="limit" type="number" min="1" :placeholder="t('product.sku.export.limitPlaceholder')" />
      </UFormField>
      <UFormField :label="t('product.sku.export.status')" class="w-full">
        <USelect v-model="status" :options="statusOptions" />
      </UFormField>
      <UFormField :label="t('product.sku.export.keyword')" class="w-full">
        <UInput v-model="keyword" :placeholder="t('product.sku.export.keywordPlaceholder')" />
      </UFormField>
    </div>

    <UAlert
      v-if="downloadInfo"
      color="primary"
      variant="soft"
      class="mt-4"
      :title="t('product.sku.export.readyTitle')"
    >
      <template #description>
        <div class="flex flex-col gap-2 text-sm text-gray-600 dark:text-gray-300 sm:flex-row sm:items-center sm:justify-between">
          <span>
            {{ t('product.sku.export.readyDesc', { file: downloadInfo.fileName || defaultFileName }) }}
          </span>
          <UButton size="sm" color="primary" variant="soft" icon="i-heroicons-arrow-down-tray" @click="openDownload">
            {{ t('product.sku.export.downloadNow') }}
          </UButton>
        </div>
      </template>
    </UAlert>

    <div class="mt-4 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
      <span>{{ t('product.sku.export.hint', { limit: limit || defaultLimit }) }}</span>
      <div class="flex gap-2">
        <UButton variant="ghost" size="sm" @click="resetExport">{{ t('common.reset') }}</UButton>
        <UButton color="primary" size="sm" :loading="exporting" @click="handleExport">
          {{ t('product.sku.export.submit') }}
        </UButton>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n, useToast } from '#imports'
import { useSkuBulkActions } from '~/composables/useSkuBulkActions'
import type { SkuExportPayload } from '~/types/product/sku'

const props = defineProps<{
  keyword?: string | null
  status?: string | null
  spuId?: string | null
}>()

const emit = defineEmits<{
  exported: [taskId?: string]
}>()

const { t } = useI18n()
const toast = useToast()
const bulkActions = useSkuBulkActions()

const defaultFormat = 'csv'
const defaultLimit = 1000
const format = ref<string>(defaultFormat)
const limit = ref<number>(defaultLimit)
const status = ref<string>(props.status ?? '')
const keyword = ref<string>(props.keyword ?? '')
const exporting = ref(false)
const downloadInfo = ref<{ url: string; fileName?: string } | null>(null)

watch(
  () => props.status,
  (val) => {
    if (typeof val === 'string') {
      status.value = val
    } else if (val == null) {
      status.value = ''
    }
  },
)

watch(
  () => props.keyword,
  (val) => {
    if (typeof val === 'string') {
      keyword.value = val
    } else if (val == null) {
      keyword.value = ''
    }
  },
)

const formatOptions = computed(() => [
  { label: 'CSV', value: 'csv' },
])
const statusOptions = computed(() => [
  { label: t('common.all'), value: '' },
  { label: t('status.active'), value: 'active' },
  { label: t('status.inactive'), value: 'inactive' },
])

const defaultFileName = 'sku-export.csv'

const resetExport = () => {
  format.value = defaultFormat
  limit.value = defaultLimit
  downloadInfo.value = null
}

const handleExport = async () => {
  if (exporting.value) return
  exporting.value = true
  try {
    const payload: SkuExportPayload = {
      format: format.value,
      limit: Number.isFinite(Number(limit.value)) ? Number(limit.value) : undefined,
      filters: {
        status: status.value || undefined,
        keyword: keyword.value || undefined,
        spu_id: props.spuId || undefined,
      },
    }
    const task = await bulkActions.submitExport(payload)
    if (task?.taskId) {
      emit('exported', task.taskId)
    }
    const result = (task?.result ?? {}) as Record<string, any>
    const directURL = (result.download_url ?? result.downloadUrl) as string | undefined
    const fileName = (result.file_name ?? result.fileName) as string | undefined
    if (directURL) {
      downloadInfo.value = {
        url: directURL,
        fileName,
      }
      openDownload()
    } else {
      toast.add({
        title: t('product.sku.export.taskCreated'),
        description: task?.taskId ? `ID: ${task.taskId}` : undefined,
        color: 'primary',
      })
    }
  } catch (error: any) {
    toast.add({
      title: error?.message ?? t('product.sku.export.failed'),
      color: 'red',
    })
  } finally {
    exporting.value = false
  }
}

const openDownload = () => {
  if (!downloadInfo.value?.url || typeof window === 'undefined') return
  window.open(downloadInfo.value.url, '_blank', 'noopener')
}
</script>
