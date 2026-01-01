<template>
<section class="rounded-xl border border-gray-200/70 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-gray-900">
    <header class="flex items-center justify-between">
      <div>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('product.sku.tasks.title') }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('product.sku.tasks.subtitle') }}
        </p>
      </div>
      <UBadge v-if="tasks.length" color="primary" variant="soft">
        {{ t('product.sku.tasks.count', { count: tasks.length }) }}
      </UBadge>
    </header>

    <div v-if="!tasks.length" class="mt-4 rounded-lg border border-dashed border-gray-200 bg-gray-50 p-4 text-sm text-gray-500 dark:border-gray-800 dark:bg-gray-800/40 dark:text-gray-300">
      {{ t('product.sku.tasks.empty') }}
    </div>

    <ul v-else class="mt-4 space-y-3">
      <li
        v-for="task in tasks"
        :key="task.taskId"
        class="rounded-lg border px-4 py-3 shadow-sm transition hover:border-primary-200 dark:border-gray-800 dark:hover:border-primary-400/60"
        :class="{
          'border-primary-300 ring-2 ring-primary-100 dark:ring-primary-500/50': task.taskId === selectedTaskId,
        }"
        @click="handleSelect(task)"
      >
        <div class="flex flex-wrap items-center gap-2">
          <UBadge size="xs" variant="subtle" :color="typeColor(task.taskType)" class="uppercase">
            {{ typeLabel(task.taskType) }}
          </UBadge>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">#{{ task.taskId }}</span>
          <UBadge size="xs" variant="soft" :color="statusColor(task.status)">
            {{ statusLabel(task) }}
          </UBadge>
        </div>
        <div class="mt-2 flex flex-wrap items-center gap-x-6 gap-y-1 text-sm text-gray-600 dark:text-gray-300">
          <span v-if="task.stats">
            {{ t('product.sku.tasks.stats', { ok: task.stats?.succeeded ?? 0, fail: task.stats?.failed ?? 0 }) }}
          </span>
          <span v-if="task.approvalRequired" class="text-amber-600 dark:text-amber-400">
            {{ approvalLabel(task) }}
          </span>
          <span v-if="task.submittedBy">{{ t('product.sku.tasks.submittedBy', { user: task.submittedBy }) }}</span>
        </div>
        <div class="mt-3 flex flex-wrap gap-2">
          <UButton
            v-if="downloadUrl(task)"
            size="xs"
            color="primary"
            variant="soft"
            icon="i-heroicons-arrow-down-tray"
            @click.stop="openLink(downloadUrl(task))"
          >
            {{ t('product.sku.tasks.download') }}
          </UButton>
          <UButton
            v-if="task.errorReportUrl"
            size="xs"
            color="warning"
            variant="soft"
            icon="i-heroicons-document-text"
            @click.stop="openLink(task.errorReportUrl)"
          >
            {{ t('product.sku.tasks.errorReport') }}
          </UButton>
          <UButton
            v-if="task.approvalRequired && (!task.approvalState || task.approvalState === 'pending')"
            size="xs"
            color="amber"
            variant="soft"
            icon="i-heroicons-shield-exclamation"
            @click.stop="emit('review', task)"
          >
            {{ t('product.sku.tasks.review') }}
          </UButton>
        </div>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import type { SkuBulkTask } from '~/types/product/sku'

const props = defineProps<{
  tasks: SkuBulkTask[]
  selectedTaskId?: string
}>()

const emit = defineEmits<{
  review: [SkuBulkTask]
  select: [SkuBulkTask]
}>()

const { t } = useI18n()

const typeLabel = (type?: string) => {
  switch (type) {
    case 'size_adjustment':
    case 'price_adjustment':
      return t('product.sku.tasks.types.price')
    case 'inventory_adjustment':
      return t('product.sku.tasks.types.inventory')
    case 'import':
      return t('product.sku.tasks.types.import')
    case 'export':
      return t('product.sku.tasks.types.export')
    default:
      return t('product.sku.tasks.types.generic')
  }
}

const typeColor = (type?: string) => {
  if (type === 'import') return 'sky'
  if (type === 'export') return 'indigo'
  if (type?.startsWith('price')) return 'primary'
  if (type?.startsWith('inventory')) return 'emerald'
  return 'gray'
}

const statusColor = (status?: string) => {
  switch (status) {
    case 'succeeded':
      return 'success'
    case 'failed':
      return 'error'
    case 'running':
    case 'approved':
      return 'primary'
    case 'cancelled':
      return 'gray'
    default:
      return 'warning'
  }
}

const STATUS_KEY: Record<string, string> = {
  pending: 'product.sku.tasks.status.pending',
  approved: 'product.sku.tasks.status.approved',
  running: 'product.sku.tasks.status.running',
  succeeded: 'product.sku.tasks.status.succeeded',
  failed: 'product.sku.tasks.status.failed',
  cancelled: 'product.sku.tasks.status.cancelled',
}

const statusLabel = (task: SkuBulkTask) => {
  const status = task.status ?? 'pending'
  if (status === 'succeeded' && downloadUrl(task)) {
    return t('product.sku.tasks.status.ready')
  }
  return t(STATUS_KEY[status] ?? STATUS_KEY.pending)
}

const approvalLabel = (task: SkuBulkTask) => {
  if (!task.approvalRequired) return ''
  if (task.approvalState === 'approved') {
    return t('product.sku.tasks.approval.approved')
  }
  if (task.approvalState === 'rejected') {
    return t('product.sku.tasks.approval.rejected')
  }
  return t('product.sku.tasks.approval.pending')
}

const downloadUrl = (task: SkuBulkTask) => {
  const result = task.result as Record<string, any> | undefined
  return (result?.download_url ?? result?.downloadUrl) as string | undefined
}

const openLink = (url?: string) => {
  if (!url || typeof window === 'undefined') return
  window.open(url, '_blank', 'noopener')
}

const handleSelect = (task: SkuBulkTask) => {
  emit('select', task)
}
</script>
