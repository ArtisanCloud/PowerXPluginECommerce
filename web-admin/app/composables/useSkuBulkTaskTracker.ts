import { computed, onBeforeUnmount, watch } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { useProductSkuStore } from '~/stores/productSku'
import type { SkuBulkTask } from '~/types/product/sku'

const ACTIVE_STATUSES = new Set(['pending', 'approved', 'running'])

export function useSkuBulkTaskTracker() {
	const store = useProductSkuStore()
	const tasks = computed<SkuBulkTask[]>(() => {
		return Object.values(store.bulkTasks || {}).sort((a, b) => {
			const left = a.updatedAt ?? a.createdAt ?? ''
			const right = b.updatedAt ?? b.createdAt ?? ''
			return right.localeCompare(left)
		})
	})
	const activeTasks = computed(() => tasks.value.filter((task) => ACTIVE_STATUSES.has(task.status)))

	const refresh = async () => {
		await store.refreshActiveBulkTasks()
	}

	const { pause, resume } = useIntervalFn(refresh, 8000, { immediate: false })

	watch(
		() => activeTasks.value.length,
		(count) => {
			if (count > 0) {
				resume()
			} else {
				pause()
			}
		},
		{ immediate: true },
	)

	onBeforeUnmount(() => pause())

	return {
		tasks,
		activeTasks,
		refresh,
	}
}
