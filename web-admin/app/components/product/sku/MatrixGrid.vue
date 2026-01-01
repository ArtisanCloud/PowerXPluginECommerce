<template>
  <div class="space-y-4">
    <div class="rounded-xl border border-dashed border-gray-300 bg-gray-50 p-6 text-center text-sm text-gray-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300">
      {{ t('product.sku.matrix.placeholder') }}
    </div>
    <BulkTaskStatusList :tasks="tasks" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { useSkuBulkTaskTracker } from '~/composables/useSkuBulkTaskTracker'
import BulkTaskStatusList from './BulkTaskStatusList.vue'

const { t } = useI18n()
const { tasks } = useSkuBulkTaskTracker()
</script>
