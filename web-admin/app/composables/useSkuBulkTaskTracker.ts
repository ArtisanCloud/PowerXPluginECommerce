import { computed, onBeforeUnmount, onMounted } from 'vue'
import { useProductSkuStore } from '~/stores/productSku'
import type { SkuBulkTask } from '~/types/product/sku'
import { useWsBusClient, type WsBusEvent } from '~/composables/useWsBusClient'

const ACTIVE_STATUSES = new Set(['pending', 'approved', 'running'])
const TASK_PROGRESS_TOPICS = [
  'task.progress',
  'powerx.task.progress.v1',
  'worker.task.updated',
  'sku.bulk.progress',
  'product.sku.bulk.progress.v1',
] as const

const wsBindingState: {
  refs: number
  handler: ((event: WsBusEvent) => void) | null
} = {
  refs: 0,
  handler: null,
}

export function useSkuBulkTaskTracker() {
  const store = useProductSkuStore()
  const ws = useWsBusClient()

  const tasks = computed<SkuBulkTask[]>(() => {
    return Object.values(store.bulkTasks || {}).sort((a, b) => {
      const left = a.updatedAt ?? a.createdAt ?? ''
      const right = b.updatedAt ?? b.createdAt ?? ''
      return right.localeCompare(left)
    })
  })

  const activeTasks = computed(() => tasks.value.filter((task) => ACTIVE_STATUSES.has(task.status)))

  const bindStream = () => {
    if (wsBindingState.handler) {
      return
    }
    const handler = (event: WsBusEvent) => {
      store.applyBulkTaskEvent(event)
    }
    TASK_PROGRESS_TOPICS.forEach((topic) => ws.subscribe(topic, handler))
    ws.connect()
    wsBindingState.handler = handler
  }

  const unbindStream = () => {
    if (!wsBindingState.handler || wsBindingState.refs > 0) {
      return
    }
    TASK_PROGRESS_TOPICS.forEach((topic) => ws.unsubscribe(topic, wsBindingState.handler!))
    wsBindingState.handler = null
  }

  const refresh = async () => {
    await store.refreshActiveBulkTasks()
  }

  onMounted(() => {
    wsBindingState.refs += 1
    bindStream()
  })

  onBeforeUnmount(() => {
    wsBindingState.refs = Math.max(0, wsBindingState.refs - 1)
    unbindStream()
  })

  return {
    tasks,
    activeTasks,
    refresh,
  }
}
