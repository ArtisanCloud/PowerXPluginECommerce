<template>
  <div class="space-y-6 px-6 py-8">
    <NuxtLink to="/product/skus" class="inline-flex items-center text-sm text-primary-500">
      <UIcon name="i-heroicons-arrow-left" class="mr-1 h-4 w-4" />
      {{ $t('product.sku.detailPage.back') }}
    </NuxtLink>
    <div>
      <p class="text-xs uppercase tracking-[0.3em] text-primary-500">{{ $t('product.sku.detailPage.kicker') }}</p>
      <h1 class="text-3xl font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.detailPage.title', { id: skuId }) }}</h1>
      <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.detailPage.subtitle') }}</p>
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
import { computed, ref } from 'vue'
import { useRoute, useI18n } from '#imports'
import ChannelMappingTab from '~/components/product/sku/ChannelMappingTab.vue'
import InventorySnapshotCard from '~/components/product/sku/InventorySnapshotCard.vue'
import BarcodePanel from '~/components/product/sku/BarcodePanel.vue'
import SerialBatchSection from '~/components/product/sku/SerialBatchSection.vue'

const route = useRoute()
const { t } = useI18n()

const skuId = computed(() => String(route.params.id ?? ''))
type DetailTab = 'channels' | 'inventory' | 'barcode' | 'serials'

const activeTab = ref<DetailTab>('channels')
const tabs = computed(() => [
  { label: t('product.sku.detailPage.channelTab'), slot: 'channels' },
  { label: t('product.sku.detailPage.inventoryTab'), slot: 'inventory' },
  { label: t('product.sku.detailPage.barcodeTab'), slot: 'barcode' },
  { label: t('product.sku.detailPage.serialTab'), slot: 'serials' },
])
</script>
