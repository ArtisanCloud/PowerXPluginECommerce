<template>
  <section class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-primary-500">{{ $t('product.sku.barcodePanel.kicker') }}</p>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.barcodePanel.title') }}</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.barcodePanel.subtitle') }}</p>
      </div>
      <UButton variant="ghost" icon="i-heroicons-arrow-path" :loading="loading" @click="resetResults">
        {{ $t('common.reset') }}
      </UButton>
    </div>

    <UTabs v-model="mode" :items="tabItems" class="w-full">
      <template #auto>
        <UForm class="grid gap-4 md:grid-cols-3" @submit.prevent="handleGenerate">
          <UFormGroup :label="$t('product.sku.barcodePanel.fields.prefix')">
            <UInput v-model="autoForm.prefix" maxlength="8" />
          </UFormGroup>
          <UFormGroup :label="$t('product.sku.barcodePanel.fields.count')">
            <UInput v-model.number="autoForm.count" type="number" min="1" max="100" />
          </UFormGroup>
          <UFormGroup :label="$t('product.sku.barcodePanel.fields.length')">
            <UInput v-model.number="autoForm.length" type="number" min="6" max="18" />
          </UFormGroup>
          <div class="md:col-span-3 flex justify-end">
            <UButton type="submit" color="primary" :loading="loading">
              {{ $t('product.sku.barcodePanel.actions.generate') }}
            </UButton>
          </div>
        </UForm>
      </template>
      <template #manual>
        <div class="space-y-3">
          <UTextarea
            v-model="manualCodes"
            :rows="4"
            :placeholder="$t('product.sku.barcodePanel.fields.codesPlaceholder')"
          />
          <div class="flex items-center justify-between text-xs text-gray-500">
            <span>{{ $t('product.sku.barcodePanel.fields.codesHint', { count: manualCodeCount }) }}</span>
            <UButton size="xs" variant="ghost" :loading="loading" @click="handleGenerate">
              {{ $t('product.sku.barcodePanel.actions.validate') }}
            </UButton>
          </div>
        </div>
      </template>
    </UTabs>

    <UCard v-if="results.length">
      <template #header>
        <div class="flex items-center justify-between">
          <h4 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ $t('product.sku.barcodePanel.resultsTitle', { count: results.length }) }}
          </h4>
          <UButton
            size="xs"
            variant="soft"
            icon="i-heroicons-arrow-down-tray"
            :disabled="!labels.length"
            @click="downloadAllLabels"
          >
            {{ $t('product.sku.barcodePanel.actions.downloadAll') }}
          </UButton>
        </div>
      </template>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th class="px-4 py-2 text-left">{{ $t('product.sku.barcodePanel.table.barcode') }}</th>
              <th class="px-4 py-2 text-left">{{ $t('product.sku.barcodePanel.table.status') }}</th>
              <th class="px-4 py-2 text-left">{{ $t('product.sku.barcodePanel.table.conflict') }}</th>
              <th class="px-4 py-2 text-right">{{ $t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="row in results" :key="row.barcode">
              <td class="px-4 py-2 font-mono text-sm text-gray-900 dark:text-white">{{ row.barcode }}</td>
              <td class="px-4 py-2">
                <UBadge :color="row.unique ? 'success' : 'error'" size="xs">
                  {{ row.unique ? $t('product.sku.barcodePanel.status.unique') : $t('product.sku.barcodePanel.status.duplicate') }}
                </UBadge>
              </td>
              <td class="px-4 py-2 text-sm text-gray-500">
                <span v-if="row.conflictSkuId">{{ row.conflictSkuId }}</span>
                <span v-else>—</span>
              </td>
              <td class="px-4 py-2 text-right">
                <div class="flex justify-end gap-2">
                  <UButton size="xs" variant="ghost" icon="i-heroicons-document-duplicate" @click="copyBarcode(row.barcode)">
                    {{ $t('product.sku.barcodePanel.actions.copy') }}
                  </UButton>
                  <UButton
                    v-if="hasLabel(row.barcode)"
                    size="xs"
                    variant="soft"
                    icon="i-heroicons-printer"
                    @click="downloadSingle(row.barcode)"
                  >
                    {{ $t('product.sku.barcodePanel.actions.download') }}
                  </UButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </UCard>
    <div v-else-if="!loading" class="text-sm text-gray-500 dark:text-gray-400">
      {{ $t('product.sku.barcodePanel.empty') }}
    </div>

    <div v-if="labels.length" class="grid gap-4 md:grid-cols-3">
      <UCard
        v-for="label in labels"
        :key="label.barcode"
        class="flex flex-col items-center justify-center text-center"
      >
        <div class="w-full" v-html="label.svg" />
        <p class="mt-2 font-mono text-xs text-gray-500">{{ label.barcode }}</p>
      </UCard>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n, useToast } from '#imports'
import { useProductSkuStore } from '~/stores/productSku'
import type { SkuBarcodeLabel } from '~/types/product/sku'

const props = defineProps<{ skuId: string }>()

const store = useProductSkuStore()
const { t } = useI18n()
const toast = useToast()

const loading = ref(false)
const mode = ref<'auto' | 'manual'>('auto')
const autoForm = reactive({
  prefix: '',
  count: 5,
  length: 12,
})
const manualCodes = ref('')

const tabItems = computed(() => [
  { label: t('product.sku.barcodePanel.autoTab'), slot: 'auto', value: 'auto' },
  { label: t('product.sku.barcodePanel.manualTab'), slot: 'manual', value: 'manual' },
])

const resultState = computed(() => (props.skuId ? store.barcodeResults[props.skuId] : null))
const results = computed(() => resultState.value?.items ?? [])
const labels = computed(() => resultState.value?.labels ?? [])
const manualCodeCount = computed(() => splitManualCodes().length)

const handleGenerate = async () => {
  if (!props.skuId) return
  const payload =
    mode.value === 'auto'
      ? {
          mode: 'auto' as const,
          prefix: autoForm.prefix,
          count: autoForm.count,
          length: autoForm.length,
        }
      : {
          mode: 'manual' as const,
          codes: splitManualCodes(),
        }
  if (mode.value === 'manual' && (!payload.codes || !payload.codes.length)) {
    toast.add({ title: t('product.sku.barcodePanel.errors.noCodes'), color: 'amber' })
    return
  }
  loading.value = true
  try {
    await store.generateBarcodes(props.skuId, payload)
    toast.add({ title: t('product.sku.barcodePanel.toastSuccess') })
  } catch (error: any) {
    toast.add({ title: t('common.error'), description: error?.message, color: 'red' })
  } finally {
    loading.value = false
  }
}

const splitManualCodes = () =>
  manualCodes.value
    .split(/[\n,;\s]+/)
    .map((code) => code.trim())
    .filter(Boolean)

const copyBarcode = async (code: string) => {
  try {
    await navigator?.clipboard?.writeText(code)
    toast.add({ title: t('product.sku.barcodePanel.toastCopied') })
  } catch {
    toast.add({ title: t('common.error'), description: t('product.sku.barcodePanel.errors.copy'), color: 'red' })
  }
}

const hasLabel = (barcode: string) => labels.value.some((label) => label.barcode === barcode)

const downloadSingle = (barcode: string) => {
  const target = labels.value.find((label) => label.barcode === barcode)
  if (target) {
    downloadLabel(target)
  }
}

const downloadAllLabels = () => {
  labels.value.forEach((label) => downloadLabel(label))
}

const downloadLabel = (label: SkuBarcodeLabel) => {
  if (typeof window === 'undefined' || !label?.svg) return
  const blob = new Blob([label.svg], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${label.barcode}.svg`
  link.click()
  URL.revokeObjectURL(url)
}

const resetResults = () => {
  if (!props.skuId) return
  store.barcodeResults[props.skuId] = null
  manualCodes.value = ''
}

defineExpose({ handleGenerate })
</script>
