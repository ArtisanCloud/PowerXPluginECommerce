<template>
  <section class="space-y-5 rounded-xl border border-gray-200/70 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-gray-900">
    <header>
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('product.sku.import.title') }}
      </h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('product.sku.import.subtitle') }}
      </p>
    </header>

    <UAlert color="gray" variant="soft">
      <template #title>
        {{ t('product.sku.import.steps.title') }}
      </template>
      <template #description>
        <ol class="list-decimal pl-5 text-xs leading-6 text-gray-500 dark:text-gray-400">
          <li>{{ t('product.sku.import.steps.download') }}</li>
          <li>{{ t('product.sku.import.steps.fill') }}</li>
          <li>{{ t('product.sku.import.steps.upload') }}</li>
        </ol>
      </template>
    </UAlert>

    <div class="grid gap-4 md:grid-cols-[minmax(0,260px)_1fr]">
      <UFormField :label="t('product.sku.import.mode.label')">
        <USelect v-model="mode" :options="modeOptions" />
      </UFormField>
      <div class="flex flex-wrap items-center gap-3">
        <UButton
          icon="i-heroicons-arrow-down-tray"
          color="primary"
          variant="soft"
          :loading="downloadingTemplate"
          :disabled="downloadingTemplate"
          @click="downloadTemplate"
        >
          {{ t('product.sku.import.downloadTemplate') }}
        </UButton>
        <UTooltip v-if="errorReportUrl" :text="t('product.sku.import.errorReport.hint')">
          <UButton
            color="warning"
            variant="soft"
            icon="i-heroicons-document-arrow-down"
            @click="downloadErrorReport"
          >
            {{ t('product.sku.import.downloadError') }}
          </UButton>
        </UTooltip>
      </div>
    </div>

    <div>
      <label
        class="flex w-full cursor-pointer flex-col items-center justify-center rounded-lg border border-dashed border-gray-300 bg-gray-50 px-5 py-8 text-center text-sm text-gray-500 transition hover:border-primary-400 hover:bg-primary-50/40 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
      >
        <input
          ref="fileInputRef"
          type="file"
          class="hidden"
          accept=".csv,.xlsx,.xls"
          @change="handleFileChange"
        />
        <span class="text-base font-medium text-gray-700 dark:text-gray-200">
          {{ selectedFileName || t('product.sku.import.dropzone.title') }}
        </span>
        <span class="mt-1 text-xs text-gray-400 dark:text-gray-500">
          {{ t('product.sku.import.dropzone.hint') }}
        </span>
      </label>
      <div class="mt-2 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
        <span>
          {{ selectedFileName ? t('product.sku.import.fileSelected', { name: selectedFileName }) : t('product.sku.import.emptyUpload') }}
        </span>
        <button
          v-if="selectedFile"
          type="button"
          class="text-primary-600 hover:underline dark:text-primary-400"
          @click="clearFile"
        >
          {{ t('product.sku.import.clear') }}
        </button>
      </div>
    </div>

    <UAlert
      v-if="importError"
      color="red"
      variant="soft"
      :title="t('product.sku.import.failed')"
      :description="importError"
    />
    <UAlert
      v-else-if="lastTaskId"
      color="primary"
      variant="soft"
      :title="t('product.sku.import.taskCreated', { id: lastTaskId })"
      :description="t('product.sku.import.taskDesc')"
    />

    <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
      <UButton variant="ghost" color="neutral" @click="resetState">
        {{ t('common.reset') }}
      </UButton>
      <UButton
        color="primary"
        :loading="submitting"
        :disabled="!selectedFile"
        @click="handleSubmit"
      >
        {{ t('product.sku.import.submit') }}
      </UButton>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n, useToast } from '#imports'
import { useApiClient } from '~/composables/api/_client'
import { useSkuBulkActions } from '~/composables/useSkuBulkActions'
import type { SkuImportMode } from '~/types/product/sku'

const props = defineProps<{
  defaultMode?: SkuImportMode
}>()

const emit = defineEmits<{
  submitted: [taskId?: string]
}>()

const { t } = useI18n()
const toast = useToast()
const { client } = useApiClient()
const bulkActions = useSkuBulkActions()

const mode = ref<SkuImportMode>(props.defaultMode ?? 'upsert')
const selectedFile = ref<File | null>(null)
const selectedFileName = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)
const submitting = ref(false)
const downloadingTemplate = ref(false)
const importError = ref<string | null>(null)
const lastTaskId = ref<string | null>(null)
const errorReportUrl = ref<string | null>(null)

const modeOptions = computed(() => [
  { label: t('product.sku.import.mode.upsert'), value: 'upsert' },
  { label: t('product.sku.import.mode.inventoryOnly'), value: 'inventory-only' },
])

const handleFileChange = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    selectedFile.value = file
    selectedFileName.value = file.name
    importError.value = null
    errorReportUrl.value = null
  }
}

const clearFile = () => {
  selectedFile.value = null
  selectedFileName.value = ''
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

const resetState = () => {
  clearFile()
  importError.value = null
  lastTaskId.value = null
  errorReportUrl.value = null
}

const handleSubmit = async () => {
  if (!selectedFile.value) {
    toast.add({ title: t('product.sku.import.emptyUpload'), color: 'red' })
    return
  }
  submitting.value = true
  importError.value = null
  errorReportUrl.value = null
  try {
    const task = await bulkActions.submitImport({
      file: selectedFile.value,
      mode: mode.value,
    })
    lastTaskId.value = task?.taskId ?? null
    if (task?.errorReportUrl) {
      errorReportUrl.value = task.errorReportUrl
    }
    emit('submitted', lastTaskId.value ?? undefined)
    clearFile()
  } catch (error: any) {
    importError.value = error?.message ?? t('product.sku.import.failed')
    const report =
      error?.data?.error_report_url ??
      error?.data?.errorReportUrl ??
      error?.errorReportUrl
    if (report) {
      errorReportUrl.value = report
    }
  } finally {
    submitting.value = false
  }
}

const downloadTemplate = async () => {
  if (downloadingTemplate.value) return
  downloadingTemplate.value = true
  try {
    const response = await client.raw('/admin/product/skus/import/template', {
      method: 'GET',
      query: { mode: mode.value },
      responseType: 'blob' as const,
    })
    const blob = response?._data as Blob
    if (!blob) {
      throw new Error('Template download failed')
    }
    const disposition = response.headers?.get?.('content-disposition')
    let fallback = 'sku_import_template.xlsx'
    if (disposition) {
      const match = disposition.match(/filename="?([^\";]+)"?/i)
      if (match?.[1]) {
        fallback = decodeURIComponent(match[1])
      }
    }
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = fallback
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch (error: any) {
    console.error('[BulkImportUploader] download template failed', error)
    toast.add({
      title: t('product.sku.import.failed'),
      description: error?.message ?? t('product.sku.import.failed'),
      color: 'red',
    })
  } finally {
    downloadingTemplate.value = false
  }
}

const downloadErrorReport = () => {
  if (!errorReportUrl.value) return
  const url = errorReportUrl.value
  if (typeof window !== 'undefined') {
    window.open(url, '_blank', 'noopener')
  }
}
</script>
