<template>
	<UModal v-model="internalOpen" :ui="modalUi" :dismissible="false">
		<template #header>
			<div class="text-lg font-semibold text-gray-900 dark:text-white">批量调整价格 / 库存</div>
		</template>
		<template #body>
			<div class="space-y-4">
				<div class="rounded border border-dashed border-gray-200 bg-gray-50 p-3 text-sm text-gray-600 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300">
					已选择 <strong>{{ selection.length }}</strong> 个 SKU，变更将在后台异步执行。
				</div>
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
					<UFormField label="调整类型" required>
						<USelect v-model="operationType" :options="operationOptions" />
					</UFormField>
					<UFormField :label="valueLabel" required>
						<UInput v-model.number="operationValue" type="number" :placeholder="valuePlaceholder" />
					</UFormField>
					<UFormField label="审批原因（可选）">
						<UInput v-model="approvalReason" placeholder="如超 5% 需说明" />
					</UFormField>
					<UFormField label="备注">
						<UTextarea v-model="notes" :rows="2" placeholder="将写入任务审计日志" />
					</UFormField>
				</div>
				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<h4 class="text-sm font-semibold text-gray-900 dark:text-white">预览</h4>
						<span class="text-xs text-gray-500">最多展示前 {{ previewRows.length }} 条</span>
					</div>
					<div v-if="!selection.length" class="rounded border border-dashed p-4 text-center text-sm text-gray-500">请选择至少一个 SKU 后再操作。</div>
					<div v-else class="max-h-64 overflow-y-auto rounded border border-gray-200 dark:border-gray-800">
						<table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-gray-700">
							<thead class="bg-gray-50 dark:bg-gray-800/50">
								<tr>
									<th class="px-3 py-2 text-left font-medium text-gray-500 dark:text-gray-300">SKU</th>
									<th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-300">当前值</th>
									<th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-300">调整后</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
								<tr v-for="row in previewRows" :key="row.id">
									<td class="px-3 py-2 font-medium text-gray-900 dark:text-white">{{ row.code }}</td>
									<td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ row.currentLabel }}</td>
									<td class="px-3 py-2 text-right text-primary-600 dark:text-primary-400">{{ row.newLabel }}</td>
								</tr>
							</tbody>
						</table>
					</div>
					<p v-if="needsApproval" class="text-xs text-amber-600 dark:text-amber-400">超出台阶或大额变更将自动进入审批，审批通过后才会执行。</p>
				</div>
			</div>
		</template>
		<template #footer>
			<div class="flex w-full justify-end gap-2">
				<UButton variant="ghost" @click="close">取消</UButton>
				<UButton color="primary" :disabled="!selection.length" :loading="submitting" @click="handleSubmit">
					提交批量任务
				</UButton>
			</div>
		</template>
	</UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { ProductSku, BulkOperationType, SkuBulkTaskRequest } from '~/types/product/sku'
import { useSkuBulkActions } from '~/composables/useSkuBulkActions'
import { useToast } from '#imports'

interface Props {
	modelValue: boolean
	selection: ProductSku[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
	(event: 'update:modelValue', value: boolean): void
	(event: 'submitted'): void
}>()

const internalOpen = computed({
	get: () => props.modelValue,
	set: (val: boolean) => emit('update:modelValue', val),
})

const operationType = ref<BulkOperationType>('price_percent')
const operationValue = ref<number>(5)
const approvalReason = ref('')
const notes = ref('')
const submitting = ref(false)

const bulkActions = useSkuBulkActions()
const toast = useToast()

const modalUi = {
	content: 'sm:max-w-3xl',
}

const operationOptions = [
	{ label: '价格 + 固定金额', value: 'price_fixed' },
	{ label: '价格 + 百分比', value: 'price_percent' },
	{ label: '库存 += 数量', value: 'inventory_fixed' },
	{ label: '库存直接替换', value: 'inventory_replace' },
]

const valueLabel = computed(() => {
	if (operationType.value === 'price_percent') {
		return '调整幅度（%）'
	}
	if (operationType.value === 'price_fixed') {
		return '调整金额（元）'
	}
	return '数量'
})

const valuePlaceholder = computed(() => {
	switch (operationType.value) {
		case 'price_percent':
			return '例如 5 表示 +5%'
		case 'price_fixed':
			return '例如 10 表示 +10 元'
		case 'inventory_fixed':
			return '例如 50 表示 +50 件'
		default:
			return '例如 100 表示覆盖为 100'
	}
})

const previewRows = computed(() => {
	return props.selection.slice(0, 8).map((sku) => {
		const base = resolveCurrentValue(sku)
		const next = computeNewValue(base)
		return {
			id: sku.id,
			code: sku.skuCode ?? 'N/A',
			currentLabel: formatValue(base),
			newLabel: formatValue(next),
		}
	})
})

const needsApproval = computed(() => {
	return (
		operationType.value === 'price_percent' && Math.abs(operationValue.value) >= 5
	)
})

const resolveCurrentValue = (sku: ProductSku): number => {
	if (operationType.value.startsWith('price')) {
		const tierPrice =
			sku.priceRefs?.[0]?.tiers?.[0]?.price ??
			(sku.priceRefs?.[0] as any)?.price ??
			0
		return Number(tierPrice) || 0
	}
	const firstInventory = sku.inventory?.[0]
	return Number(firstInventory?.availableQty ?? 0)
}

const computeNewValue = (base: number): number => {
	const value = Number(operationValue.value || 0)
	switch (operationType.value) {
		case 'price_percent':
			return Number((base * (1 + value / 100)).toFixed(2))
		case 'price_fixed':
			return Number((base + value).toFixed(2))
		case 'inventory_fixed':
			return Math.max(0, Math.round(base + value))
		case 'inventory_replace':
			return Math.max(0, Math.round(value))
		default:
			return base
	}
}

const formatValue = (val: number): string => {
	if (operationType.value.startsWith('price')) {
		return `¥${val.toFixed(2)}`
	}
	return `${val} 件`
}

const close = () => {
	internalOpen.value = false
}

const handleSubmit = async () => {
	if (!props.selection.length) {
		toast.add({ title: '请选择需要调整的 SKU', color: 'red' })
		return
	}
	if (Number.isNaN(operationValue.value)) {
		toast.add({ title: '请输入有效的调整值', color: 'red' })
		return
	}
	submitting.value = true
	try {
		const normalizedValue =
			operationType.value === 'price_percent'
				? Number(operationValue.value) / 100
				: Number(operationValue.value)
		const payload: SkuBulkTaskRequest = {
			scope: { skuIds: props.selection.map((item) => item.id) },
			operation: {
				type: operationType.value,
				value: normalizedValue,
			},
			approvalContext: approvalReason.value
				? { reason: approvalReason.value }
				: undefined,
		}
		await bulkActions.submitAdjustment(payload)
		toast.add({ title: '批量任务已提交', color: 'primary' })
		emit('submitted')
		close()
	} catch (error: any) {
		toast.add({
			title: '提交失败',
			description: error?.message ?? '请稍后再试',
			color: 'red',
		})
	} finally {
		submitting.value = false
	}
}

watch(
	() => props.modelValue,
	(opened) => {
		if (opened) {
			submitting.value = false
		}
	},
)
</script>
