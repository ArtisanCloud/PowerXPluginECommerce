import { useRuntimeConfig } from '#imports'
import { defineStore } from 'pinia'
import type {
	ProductSku,
	SkuBulkTask,
	SkuBulkTaskStatus,
} from '~/types/product/sku'

interface ProductSkuState {
	items: ProductSku[]
	total: number
	loading: boolean
	bulkTasks: Record<string, SkuBulkTask>
	bulkTaskStatus: Record<string, SkuBulkTaskStatus>
	apiBase: string
}

const defaultConfig = () => {
	const config = useRuntimeConfig()
	const { productSkuApiBase } = config.public
	return productSkuApiBase || '/_p/com.powerx.plugin.ecommerce/api/v1/products/skus'
}

export const useProductSkuStore = defineStore('product-sku', {
	state: (): ProductSkuState => ({
		items: [],
		total: 0,
		loading: false,
		bulkTasks: {},
		bulkTaskStatus: {},
		apiBase: defaultConfig(),
	}),
	actions: {
		setApiBase(base: string) {
			if (!base) return
			this.apiBase = base.replace(/\/+$/, '')
		},
		setItems(items: ProductSku[], total?: number) {
			this.items = items
			if (typeof total === 'number') {
				this.total = total
			} else {
				this.total = items.length
			}
		},
		trackBulkTask(task: SkuBulkTask) {
			this.bulkTasks[task.taskId] = task
			if (task.status) {
				this.bulkTaskStatus[task.taskId] = task.status
			}
		},
		updateBulkTaskStatus(taskId: string, status: SkuBulkTaskStatus) {
			if (!taskId) return
			this.bulkTaskStatus[taskId] = status
		},
		reset() {
			this.items = []
			this.total = 0
			this.bulkTasks = {}
			this.bulkTaskStatus = {}
		},
	},
})
