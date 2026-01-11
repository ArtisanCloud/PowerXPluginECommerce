<template>
  <section class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-primary-500">{{ $t('product.sku.inventoryPanel.kicker') }}</p>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.inventoryPanel.title') }}</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.inventoryPanel.subtitle') }}</p>
      </div>
      <div class="flex gap-2">
        <UButton icon="i-heroicons-arrow-path" variant="ghost" :loading="loading" @click="fetchSnapshot">
          {{ $t('common.refresh') }}
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <h4 class="text-lg font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.inventoryPanel.adjustTitle') }}</h4>
      </template>
      <div class="grid gap-4 md:grid-cols-12">
        <UFormField :label="$t('product.sku.inventoryPanel.adjustDelta')" class="md:col-span-4">
          <UInput
            v-model="deltaText"
            type="number"
            inputmode="numeric"
            step="1"
            data-testid="inventory-delta"
            :disabled="adjusting || loading"
          />
        </UFormField>
        <div class="flex items-end md:col-span-2">
          <UButton
            color="primary"
            :loading="adjusting"
            :disabled="!canSubmitDelta || loading"
            data-testid="inventory-apply"
            @click="applyDelta"
          >
            {{ $t('product.sku.inventoryPanel.applyDelta') }}
          </UButton>
        </div>
        <div v-if="errorMessage" class="md:col-span-12">
          <UAlert color="error" variant="soft" :title="$t('common.error')" :description="errorMessage" />
        </div>
      </div>
    </UCard>

    <div v-if="loading">
      <USkeleton class="h-24 w-full" />
    </div>
    <div v-else-if="snapshot">
      <div class="grid gap-4 md:grid-cols-3">
        <UCard>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.inventoryPanel.available') }}</p>
          <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ snapshot.summary.availableQty }}</p>
        </UCard>
        <UCard>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.inventoryPanel.locked') }}</p>
          <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ snapshot.summary.lockedQty }}</p>
        </UCard>
        <UCard>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.inventoryPanel.transit') }}</p>
          <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ snapshot.summary.inTransitQty }}</p>
        </UCard>
      </div>
      <UAlert
        v-if="snapshot.isStale"
        color="warning"
        variant="soft"
        class="mt-4"
        :title="$t('product.sku.inventoryPanel.staleTitle')"
        :description="$t('product.sku.inventoryPanel.staleDesc')"
      />
      <UCard class="mt-4">
        <template #header>
          <h4 class="text-lg font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.inventoryPanel.tableTitle') }}</h4>
        </template>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
            <thead class="bg-gray-50 dark:bg-gray-800">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500">{{ $t('product.sku.inventoryPanel.columns.warehouse') }}</th>
                <th class="px-4 py-2 text-right text-xs font-medium uppercase tracking-wide text-gray-500">{{ $t('product.sku.inventoryPanel.columns.available') }}</th>
                <th class="px-4 py-2 text-right text-xs font-medium uppercase tracking-wide text-gray-500">{{ $t('product.sku.inventoryPanel.columns.locked') }}</th>
                <th class="px-4 py-2 text-right text-xs font-medium uppercase tracking-wide text-gray-500">{{ $t('product.sku.inventoryPanel.columns.transit') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500">{{ $t('product.sku.inventoryPanel.columns.syncedAt') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr v-for="warehouse in snapshot.warehouses" :key="warehouse.warehouseId">
                <td class="px-4 py-2 font-medium text-gray-900 dark:text-white">
                  {{ warehouse.warehouseId }}
                  <UBadge
                    v-if="warehouse.alertLevel"
                    size="xs"
                    :color="warehouse.alertLevel === 'alert' ? 'error' : 'warning'"
                    class="ml-2"
                  >
                    {{ warehouse.alertLevel === 'alert' ? $t('product.sku.inventoryPanel.alert') : $t('product.sku.inventoryPanel.warning') }}
                  </UBadge>
                </td>
                <td class="px-4 py-2 text-right">{{ warehouse.availableQty }}</td>
                <td class="px-4 py-2 text-right">{{ warehouse.lockedQty }}</td>
                <td class="px-4 py-2 text-right">{{ warehouse.inTransitQty }}</td>
                <td class="px-4 py-2 text-left text-gray-500">{{ formatDateTime(warehouse.lastSyncedAt) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </UCard>
    </div>
    <div v-else class="text-sm text-gray-500 dark:text-gray-400">
      {{ $t('product.sku.inventoryPanel.empty') }}
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useProductSkuStore } from '~/stores/productSku'
import type { SkuInventorySnapshot } from '~/types/product/sku'

const props = defineProps<{ skuId: string }>()

const store = useProductSkuStore()

const loading = ref(false)
const adjusting = ref(false)
const deltaText = ref('')
const errorMessage = ref('')
const snapshot = computed<SkuInventorySnapshot | null>(() => store.inventorySnapshots[props.skuId] ?? null)

const fetchSnapshot = async () => {
  if (!props.skuId) return
  loading.value = true
  errorMessage.value = ''
  try {
    await store.fetchInventorySnapshot(props.skuId)
  } finally {
    loading.value = false
  }
}

const canSubmitDelta = computed(() => {
  const trimmed = deltaText.value.trim()
  if (!trimmed) return false
  const parsed = Number.parseInt(trimmed, 10)
  return Number.isFinite(parsed) && parsed !== 0
})

const applyDelta = async () => {
  if (!props.skuId) return
  const parsed = Number.parseInt(deltaText.value.trim(), 10)
  if (!Number.isFinite(parsed) || parsed === 0) {
    errorMessage.value = 'delta 必须为非 0 整数'
    return
  }
  adjusting.value = true
  errorMessage.value = ''
  try {
    await store.adjustInventorySnapshot(props.skuId, parsed)
    deltaText.value = ''
  } catch (error: any) {
    errorMessage.value = error?.message ? String(error.message) : String(error)
  } finally {
    adjusting.value = false
  }
}

const formatDateTime = (value?: string) => {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}

watch(
  () => props.skuId,
  () => {
    fetchSnapshot()
  },
  { immediate: true },
)
</script>
