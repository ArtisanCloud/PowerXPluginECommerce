<template>
  <div class="space-y-6 px-6 py-8">
    <header>
      <p class="text-xs uppercase tracking-[0.3em] text-primary-500">{{ $t('product.sku.inventoryPage.kicker') }}</p>
      <h1 class="text-3xl font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.inventoryPage.title') }}</h1>
      <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.inventoryPage.subtitle') }}</p>
    </header>

    <UCard>
      <UForm @submit.prevent="handleLookup" class="grid gap-4 md:grid-cols-[1fr_auto] items-end">
        <UFormGroup :label="$t('product.sku.inventoryPage.skuLabel')" class="md:col-span-1">
          <UInput v-model="inputId" :placeholder="$t('product.sku.inventoryPage.skuPlaceholder')" />
        </UFormGroup>
        <div class="flex gap-2">
          <UButton type="submit" color="primary">{{ $t('product.sku.inventoryPage.lookup') }}</UButton>
          <UButton variant="ghost" @click="clearSelection">{{ $t('common.reset') }}</UButton>
        </div>
      </UForm>
    </UCard>

    <InventorySnapshotCard v-if="currentSkuId" :key="currentSkuId" :sku-id="currentSkuId" />
    <div v-else class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.inventoryPage.empty') }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import InventorySnapshotCard from '~/components/product/sku/InventorySnapshotCard.vue'

const inputId = ref('')
const currentSkuId = ref('')

const handleLookup = () => {
  currentSkuId.value = inputId.value.trim()
}

const clearSelection = () => {
  inputId.value = ''
  currentSkuId.value = ''
}
</script>
