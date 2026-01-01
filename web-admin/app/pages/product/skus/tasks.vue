<template>
  <div class="px-4 py-8 lg:px-8">
    <div class="mx-auto flex max-w-6xl flex-col gap-6">
      <header class="flex flex-col gap-2">
        <p class="text-sm uppercase tracking-[0.3em] text-primary-500">{{ $t('product.sku.taskPage.kicker') }}</p>
        <h1 class="text-3xl font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.taskPage.title') }}</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.taskPage.subtitle') }}</p>
      </header>

      <UCard>
        <div class="flex flex-col gap-4 md:flex-row md:items-center">
          <div class="flex-1">
            <UFormField :label="$t('product.sku.taskPage.searchLabel')">
              <UInput v-model="lookupId" :placeholder="$t('product.sku.taskPage.searchPlaceholder')" icon="i-heroicons-magnifying-glass" />
            </UFormField>
          </div>
          <div class="flex gap-2">
            <UButton color="primary" :loading="lookupLoading" @click="handleLookup">
              {{ $t('product.sku.taskPage.searchButton') }}
            </UButton>
            <UButton variant="ghost" @click="refreshSelected" :disabled="!selectedTask">
              {{ $t('product.sku.taskPage.refresh') }}
            </UButton>
          </div>
        </div>
      </UCard>

      <div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
        <BulkTaskStatusList
          :tasks="bulkTasks"
          :selected-task-id="selectedTaskId || undefined"
          @select="selectTask"
          @review="openApprovalDrawer"
        />
        <UCard>
          <template #header>
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ selectedTask ? `#${selectedTask.taskId}` : $t('product.sku.taskPage.emptyTitle') }}
              </h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ selectedTask ? $t('product.sku.taskPage.detailSub') : $t('product.sku.taskPage.emptySubtitle') }}
              </p>
            </div>
          </template>
          <div v-if="selectedTask" class="space-y-4">
            <div class="grid grid-cols-1 gap-3 text-sm text-gray-600 dark:text-gray-300">
              <p><span class="text-gray-400">{{ $t('product.sku.taskPage.status') }}：</span>{{ statusLabel }}</p>
              <p><span class="text-gray-400">{{ $t('product.sku.taskPage.type') }}：</span>{{ typeLabel }}</p>
              <p><span class="text-gray-400">{{ $t('product.sku.taskPage.affected') }}：</span>{{ selectedTask.affectedCount ?? 0 }}</p>
              <p><span class="text-gray-400">{{ $t('product.sku.taskPage.submittedBy') }}：</span>{{ selectedTask.submittedBy || 'system' }}</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <UButton size="sm" variant="soft" @click="refreshSelected" :loading="detailRefreshing">
                {{ $t('product.sku.taskPage.refresh') }}
              </UButton>
              <UButton
                v-if="downloadUrl"
                size="sm"
                color="primary"
                variant="soft"
                icon="i-heroicons-arrow-down-tray"
                @click="openDownload"
              >
                {{ $t('product.sku.tasks.download') }}
              </UButton>
              <UButton
                v-if="selectedTask.errorReportUrl"
                size="sm"
                color="warning"
                variant="soft"
                icon="i-heroicons-document-text"
                @click="openLink(selectedTask.errorReportUrl)"
              >
                {{ $t('product.sku.tasks.errorReport') }}
              </UButton>
              <UButton
                v-if="selectedTask.status === 'failed'"
                size="sm"
                color="secondary"
                :loading="retrying"
                @click="handleRetry"
              >
                {{ $t('product.sku.taskPage.retry') }}
              </UButton>
              <UButton
                v-if="selectedTask.approvalRequired && (!selectedTask.approvalState || selectedTask.approvalState === 'pending')"
                size="sm"
                color="amber"
                variant="soft"
                @click="openApprovalDrawer(selectedTask)"
              >
                {{ $t('product.sku.tasks.review') }}
              </UButton>
            </div>
            <div>
              <p class="text-xs font-medium uppercase text-gray-400">{{ $t('product.sku.taskPage.resultTitle') }}</p>
              <pre class="mt-2 max-h-72 overflow-auto rounded bg-gray-100 p-3 text-xs text-gray-800 dark:bg-gray-900 dark:text-gray-100">{{ formattedResult }}</pre>
            </div>
          </div>
          <div v-else class="text-sm text-gray-500 dark:text-gray-400">
            {{ $t('product.sku.taskPage.emptyBody') }}
          </div>
        </UCard>
      </div>
    </div>

    <ApprovalDrawer
      v-model="approvalDrawerOpen"
      :task="approvalTarget"
      :loading="approvalLoading"
      @approve="(note) => decideApproval('approve', note)"
      @reject="(note) => decideApproval('reject', note)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useToast, useI18n } from '#imports'
import BulkTaskStatusList from '~/components/product/sku/BulkTaskStatusList.vue'
import ApprovalDrawer from '~/components/product/sku/ApprovalDrawer.vue'
import { useSkuBulkTaskTracker } from '~/composables/useSkuBulkTaskTracker'
import { useSkuBulkActions } from '~/composables/useSkuBulkActions'
import { useProductSkuStore } from '~/stores/productSku'
import type { SkuBulkTask } from '~/types/product/sku'

const { t } = useI18n()
const { tasks: bulkTasks } = useSkuBulkTaskTracker()
const bulkActions = useSkuBulkActions()
const store = useProductSkuStore()
const toast = useToast()

const lookupId = ref('')
const lookupLoading = ref(false)
const selectedTaskId = ref<string>('')
const detailRefreshing = ref(false)
const retrying = ref(false)
const approvalDrawerOpen = ref(false)
const approvalTarget = ref<SkuBulkTask | null>(null)
const approvalLoading = ref(false)

const selectedTask = computed(() => {
  const current = bulkTasks.value.find((task) => task.taskId === selectedTaskId.value)
  return current ?? bulkTasks.value[0] ?? null
})

watch(
  () => bulkTasks.value.length,
  () => {
    if (!bulkTasks.value.length) {
      selectedTaskId.value = ''
    } else if (!selectedTaskId.value) {
      selectedTaskId.value = bulkTasks.value[0].taskId
    }
  },
  { immediate: true },
)

const statusLabel = computed(() => {
  if (!selectedTask.value) return '--'
  const statusKey = selectedTask.value.status || 'pending'
  return t(`product.sku.tasks.status.${statusKey}` as const)
})

const typeLabel = computed(() => {
  if (!selectedTask.value) return '--'
  switch (selectedTask.value.taskType) {
    case 'price_adjustment':
      return t('product.sku.tasks.types.price')
    case 'inventory_adjustment':
      return t('product.sku.tasks.types.inventory')
    case 'import':
      return t('product.sku.tasks.types.import')
    case 'export':
      return t('product.sku.tasks.types.export')
    default:
      return selectedTask.value.taskType
  }
})

const downloadUrl = computed(() => {
  const result = selectedTask.value?.result as Record<string, any> | undefined
  return (result?.download_url ?? result?.downloadUrl) as string | undefined
})

const formattedResult = computed(() => {
  if (!selectedTask.value) return '{}'
  const payload = {
    scope: selectedTask.value.scope ?? {},
    operation: selectedTask.value.operation ?? {},
    result: selectedTask.value.result ?? {},
  }
  try {
    return JSON.stringify(payload, null, 2)
  } catch {
    return '{}'
  }
})

const handleLookup = async () => {
  const id = lookupId.value.trim()
  if (!id) return
  lookupLoading.value = true
  try {
    await store.refreshBulkTask(id)
    selectedTaskId.value = id
    toast.add({ title: t('product.sku.taskPage.lookupSuccess'), description: `#${id}`, color: 'primary' })
  } catch (error: any) {
    toast.add({
      title: t('product.sku.taskPage.lookupFailed'),
      description: error?.message ?? '',
      color: 'red',
    })
  } finally {
    lookupLoading.value = false
  }
}

const selectTask = (task: SkuBulkTask) => {
  selectedTaskId.value = task.taskId
}

const refreshSelected = async () => {
  if (!selectedTask.value) return
  detailRefreshing.value = true
  try {
    await store.refreshBulkTask(selectedTask.value.taskId)
  } finally {
    detailRefreshing.value = false
  }
}

const openDownload = () => {
  if (downloadUrl.value) {
    openLink(downloadUrl.value)
  }
}

const openLink = (url?: string) => {
  if (!url || typeof window === 'undefined') return
  window.open(url, '_blank', 'noopener')
}

const handleRetry = async () => {
  if (!selectedTask.value) return
  retrying.value = true
  try {
    await bulkActions.retryTask(selectedTask.value.taskId)
  } finally {
    retrying.value = false
  }
}

const openApprovalDrawer = (task: SkuBulkTask) => {
  approvalTarget.value = task
  approvalDrawerOpen.value = true
}

const decideApproval = async (decision: 'approve' | 'reject', note: string) => {
  if (!approvalTarget.value) return
  approvalLoading.value = true
  try {
    await bulkActions.decideApproval({ taskId: approvalTarget.value.taskId, decision, note })
    approvalDrawerOpen.value = false
  } catch (error: any) {
    toast.add({
      title: t('product.sku.approval.toastFailed'),
      description: error?.message ?? '',
      color: 'red',
    })
  } finally {
    approvalLoading.value = false
  }
}
</script>
