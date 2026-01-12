<template>
  <div class="space-y-6 px-6 py-8">
    <NuxtLink :to="backTo" class="inline-flex items-center text-sm text-primary-500">
      <UIcon name="i-heroicons-arrow-left" class="mr-1 h-4 w-4" />
      {{ backText }}
    </NuxtLink>
    <div>
      <p class="text-xs uppercase tracking-[0.3em] text-primary-500">{{ $t('product.sku.detailPage.kicker') }}</p>
      <h1 class="text-3xl font-semibold text-gray-900 dark:text-white">
        {{ displayTitle }}
      </h1>
      <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.detailPage.subtitle') }}</p>
      <p v-if="displaySubTitle" class="mt-1 text-xs text-gray-500 dark:text-gray-400">
        {{ displaySubTitle }}
      </p>
    </div>

    <UTabs v-model="activeTab" :items="tabs" class="w-full">
      <template #channels>
        <ChannelMappingTab :sku-id="skuId" />
      </template>
      <template #inventory>
        <InventorySnapshotCard :sku-id="skuId" />
      </template>
      <template #barcode>
        <BarcodePanel :sku-id="skuId" />
      </template>
      <template #serials>
        <SerialBatchSection :sku-id="skuId" />
      </template>
    </UTabs>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useI18n } from '#imports'
import ChannelMappingTab from '~/components/product/sku/ChannelMappingTab.vue'
import InventorySnapshotCard from '~/components/product/sku/InventorySnapshotCard.vue'
import BarcodePanel from '~/components/product/sku/BarcodePanel.vue'
import SerialBatchSection from '~/components/product/sku/SerialBatchSection.vue'
import { useSkuApi } from '~/composables/api/useSku'

const route = useRoute()
const { t } = useI18n()
const skuApi = useSkuApi()

const skuId = computed(() => String(route.params.id ?? ''))
type DetailTab = 'channels' | 'inventory' | 'barcode' | 'serials'

const parseDetailTab = (value: unknown): DetailTab | null => {
  const tab = String(value ?? '').trim()
  if (tab === 'inventory' || tab === 'channels' || tab === 'barcode' || tab === 'serials') return tab
  return null
}

const activeTab = ref<DetailTab>(parseDetailTab(route.query.tab) ?? 'channels')
const tabs = computed(() => [
  { label: t('product.sku.detailPage.channelTab'), slot: 'channels', value: 'channels' },
  { label: t('product.sku.detailPage.inventoryTab'), slot: 'inventory', value: 'inventory' },
  { label: t('product.sku.detailPage.barcodeTab'), slot: 'barcode', value: 'barcode' },
  { label: t('product.sku.detailPage.serialTab'), slot: 'serials', value: 'serials' },
])

const skuInfo = ref<any>(null)
const displayTitle = computed(() => {
  const code = String(skuInfo.value?.sku_code ?? skuInfo.value?.skuCode ?? '').trim()
  if (code) return `SKU ${code}`
  return `SKU ${skuId.value}`
})

const displaySubTitle = computed(() => {
  const spuName = String(skuInfo.value?.spu_name ?? skuInfo.value?.spuName ?? '').trim()
  const spec = String(skuInfo.value?.spec_display ?? skuInfo.value?.specDisplay ?? '').trim()
  const parts = [spuName, spec].filter(Boolean)
  return parts.join(' · ')
})

const backTo = computed(() => {
  const from = String(route.query.from ?? '').trim()
  if (from === 'inventory') return '/inventory/stock'
  return '/product/skus'
})
const backText = computed(() => {
  const from = String(route.query.from ?? '').trim()
  if (from === 'inventory') return '返回库存列表'
  return t('product.sku.detailPage.back')
})

const syncTabFromQuery = () => {
  const tab = parseDetailTab(route.query.tab)
  if (tab) activeTab.value = tab
}

onMounted(async () => {
  try {
    skuInfo.value = await skuApi.get(skuId.value, { locale: 'zh-CN' })
  } catch (e) {
    console.warn('[SKU Detail] failed to load sku info', e)
    skuInfo.value = null
  }
})

watch(
  () => route.query.tab,
  () => syncTabFromQuery(),
)
</script>
