<template>
  <UModal v-model:open="internalOpen" :ui="drawerUi" :prevent-close="loading">
    <template #header>
      <div>
        <p class="text-xs uppercase tracking-[0.2em] text-primary-500">{{ t('product.sku.approval.title') }}</p>
        <h3 class="text-2xl font-semibold text-gray-900 dark:text-white">
          {{ headerTitle }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('product.sku.approval.subtitle') }}
        </p>
      </div>
    </template>
    <template #body>
      <div v-if="!task" class="text-sm text-gray-500 dark:text-gray-400">
        {{ t('product.sku.approval.empty') }}
      </div>
      <div v-else class="space-y-6">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <p class="text-xs font-medium uppercase text-gray-400">{{ t('product.sku.approval.status') }}</p>
            <UBadge variant="soft" :color="statusColor">{{ statusLabel }}</UBadge>
          </div>
          <div>
            <p class="text-xs font-medium uppercase text-gray-400">{{ t('product.sku.approval.requestedBy') }}</p>
            <p class="text-sm text-gray-900 dark:text-gray-100">{{ task.submittedBy || t('product.sku.approval.unknown') }}</p>
          </div>
          <div>
            <p class="text-xs font-medium uppercase text-gray-400">{{ t('product.sku.approval.affected') }}</p>
            <p class="text-sm text-gray-900 dark:text-gray-100">{{ affectedLabel }}</p>
          </div>
          <div>
            <p class="text-xs font-medium uppercase text-gray-400">{{ t('product.sku.approval.threshold') }}</p>
            <p class="text-sm text-gray-900 dark:text-gray-100">{{ thresholdLabel }}</p>
          </div>
          <div>
            <p class="text-xs font-medium uppercase text-gray-400">{{ t('product.sku.approval.reason') }}</p>
            <p class="text-sm text-gray-900 dark:text-gray-100">{{ reasonLabel }}</p>
          </div>
        </div>
        <div class="rounded-lg border border-gray-200/70 p-4 dark:border-gray-700">
          <p class="text-xs font-medium uppercase text-gray-400">
            {{ t('product.sku.approval.operation') }}
          </p>
          <p class="mt-1 text-base text-gray-900 dark:text-gray-100">{{ operationLabel }}</p>
          <p v-if="task.result?.summary" class="mt-2 text-xs text-gray-500 dark:text-gray-400">
            {{ task.result.summary }}
          </p>
        </div>
        <div class="rounded-lg border border-gray-200/70 p-4 dark:border-gray-700">
          <p class="text-xs font-medium uppercase text-gray-400">
            {{ t('product.sku.approval.scope') }}
          </p>
          <pre class="mt-2 max-h-48 overflow-auto rounded bg-gray-100 p-3 text-xs text-gray-700 dark:bg-gray-800 dark:text-gray-200">{{ scopePreview }}</pre>
        </div>
        <UFormField :label="t('product.sku.approval.noteLabel')" :help="rejectHelp">
          <UTextarea v-model="note" :rows="3" :placeholder="t('product.sku.approval.notePlaceholder')" />
        </UFormField>
        <p class="text-xs text-amber-600 dark:text-amber-400">
          {{ t('product.sku.approval.pendingTip') }}
        </p>
      </div>
    </template>
    <template #footer>
      <div class="flex flex-wrap gap-2">
        <UButton variant="ghost" :disabled="loading" @click="close">
          {{ t('product.sku.approval.actions.cancel') }}
        </UButton>
        <UButton
          color="error"
          variant="soft"
          :disabled="!note.trim()"
          :loading="loading && lastDecision === 'reject'"
          @click="emitReject"
        >
          {{ t('product.sku.approval.actions.reject') }}
        </UButton>
        <UButton
          color="success"
          :loading="loading && lastDecision === 'approve'"
          @click="emitApprove"
        >
          {{ t('product.sku.approval.actions.approve') }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '#imports'
import type { SkuBulkTask } from '~/types/product/sku'

const props = defineProps<{
  modelValue: boolean
  task: SkuBulkTask | null
  loading?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [boolean]; approve: [string]; reject: [string] }>()

const { t } = useI18n()
const note = ref('')
const lastDecision = ref<'approve' | 'reject' | null>(null)

const internalOpen = computed({
  get: () => props.modelValue,
  set: (val: boolean) => emit('update:modelValue', val),
})

watch(
  () => props.task?.taskId,
  () => {
    note.value = ''
    lastDecision.value = null
  },
)

const drawerUi = {
  content: 'w-full sm:max-w-2xl',
  body: 'space-y-4',
}

const headerTitle = computed(() => (props.task ? `#${props.task.taskId}` : '--'))

const statusLabel = computed(() => {
  if (!props.task) return '--'
  if (props.task.approvalState === 'approved') return t('product.sku.tasks.approval.approved')
  if (props.task.approvalState === 'rejected') return t('product.sku.tasks.approval.rejected')
  return t('product.sku.tasks.approval.pending')
})

const statusColor = computed(() => {
  if (!props.task) return 'gray'
  if (props.task.approvalState === 'approved') return 'success'
  if (props.task.approvalState === 'rejected') return 'error'
  return 'warning'
})

const affectedLabel = computed(() => (props.task?.affectedCount ?? 0).toString())

const thresholdLabel = computed(() => {
  if (props.task?.approvalThreshold != null && props.task.approvalThreshold > 0) {
    return String(props.task.approvalThreshold)
  }
  const summary = props.task?.result as Record<string, any> | undefined
  if (summary?.threshold != null) {
    return String(summary.threshold)
  }
  return t('product.sku.approval.notProvided')
})

const reasonLabel = computed(() =>
  props.task?.approvalReason ? props.task.approvalReason : t('product.sku.approval.notProvided'),
)

const operationLabel = computed(() => {
  if (!props.task) return '--'
  switch (props.task.taskType) {
    case 'price_adjustment':
      return t('product.sku.tasks.types.price')
    case 'inventory_adjustment':
      return t('product.sku.tasks.types.inventory')
    case 'import':
      return t('product.sku.tasks.types.import')
    case 'export':
      return t('product.sku.tasks.types.export')
    default:
      return props.task.taskType
  }
})

const scopePreview = computed(() => {
  if (!props.task) return '{}'
  const scope = props.task.scope ?? props.task.result?.scope ?? props.task.result?.filters ?? {}
  try {
    return JSON.stringify(scope, null, 2)
  } catch {
    return '{}'
  }
})

const rejectHelp = computed(() => (!note.value.trim() ? t('product.sku.approval.rejectHint') : ''))

const emitApprove = () => {
  if (!props.task) return
  lastDecision.value = 'approve'
  emit('approve', note.value.trim())
}

const emitReject = () => {
  if (!props.task || !note.value.trim()) return
  lastDecision.value = 'reject'
  emit('reject', note.value.trim())
}

const close = () => {
  if (props.loading) return
  internalOpen.value = false
}
</script>
