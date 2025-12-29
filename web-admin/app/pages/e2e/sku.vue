<template>
	<div class="space-y-10 p-6">
		<div>
			<p class="text-xs uppercase tracking-[0.3em] text-primary-500">E2E Harness</p>
			<h1 class="text-2xl font-semibold text-gray-900 dark:text-white">Product SKU Playground</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400">
				该页面仅在开发环境加载，用于 Playwright 覆盖 SKU 三大用户故事。
			</p>
		</div>

		<section id="us1-generator" class="space-y-4">
			<div>
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white">US1 · SKU 生成器</h2>
				<p class="text-sm text-gray-500 dark:text-gray-400">模拟规格生成与批量写入。</p>
			</div>
			<UCard>
				<div class="flex flex-wrap items-center justify-between gap-4">
					<div class="space-y-1">
						<p class="text-sm text-gray-500">SPU: E2E 商品</p>
						<p class="text-sm text-gray-500">规格：颜色 × 容量</p>
					</div>
					<div class="flex gap-2">
						<UButton @click="generatePreview" data-testid="btn-generate">生成预览</UButton>
						<UButton color="primary" :disabled="!state.previewReady" @click="writeSku">写入 SKU</UButton>
					</div>
				</div>
				<div class="mt-4 space-y-2">
					<p v-if="state.previewReady" class="text-emerald-600">可写入 · 共 {{ candidateCount }} 条</p>
					<p v-if="state.generatorMessage" class="text-primary-600">{{ state.generatorMessage }}</p>
					<UBadge v-if="state.previewReady" color="success" data-testid="generator-ready">候选可写入</UBadge>
				</div>
			</UCard>
		</section>

		<section id="us2-bulk" class="space-y-4">
			<div class="space-y-1">
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white">US2 · 批量调整</h2>
				<p class="text-sm text-gray-500 dark:text-gray-400">模拟批量任务创建、审批信息与任务列表。</p>
			</div>
			<UAlert v-if="state.bulkTaskMessage" color="primary" variant="soft" data-testid="bulk-banner">
				<div class="font-medium">{{ state.bulkTaskMessage }}</div>
				<div class="text-sm text-gray-500">{{ state.bulkTaskId }}</div>
			</UAlert>
			<UCard data-testid="bulk-dialog" class="max-w-2xl space-y-4">
				<template #header>
					<div class="flex items-center justify-between">
						<div class="text-lg font-semibold">批量调整价格 / 库存</div>
						<UBadge color="primary" variant="soft">Harness</UBadge>
					</div>
				</template>
				<form class="space-y-4" @submit.prevent="submitBulkTask" data-testid="bulk-form">
					<UFormGroup label="调整幅度（%）" required>
						<UInput
							v-model.number="bulkForm.adjustment"
							type="number"
							data-testid="bulk-adjustment-input"
							aria-label="调整幅度（%）"
						/>
					</UFormGroup>
					<UFormGroup label="备注">
						<UTextarea
							v-model="bulkForm.note"
							rows="3"
							placeholder="将写入任务审计日志"
							data-testid="bulk-note-input"
							aria-label="备注"
						/>
					</UFormGroup>
					<div class="flex justify-end gap-2 pt-2">
						<UButton variant="ghost" @click="bulkForm.note = ''">清空备注</UButton>
						<UButton type="submit" color="primary" data-testid="submit-bulk-btn">提交批量任务</UButton>
					</div>
				</form>
				<div
					v-if="state.bulkTaskMessage"
					class="rounded-md bg-primary-50/80 p-3 text-sm text-primary-700"
					data-testid="bulk-result"
					aria-live="polite"
				>
					<div class="font-semibold" data-testid="bulk-result-message">{{ state.bulkTaskMessage }}</div>
					<div class="font-mono text-xs text-primary-600" data-testid="bulk-result-id">{{ state.bulkTaskId }}</div>
				</div>
			</UCard>
	</section>

		<section id="us3-channels" class="space-y-6">
			<div>
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white">US3 · 渠道映射/条码/序列号</h2>
				<p class="text-sm text-gray-500 dark:text-gray-400">提供简单表单以模拟真实 UI 行为。</p>
			</div>

			<UCard>
				<template #header>
					<div class="flex items-center justify-between">
						<h3 class="text-lg font-semibold text-gray-900 dark:text-white">渠道映射</h3>
						<UButton size="xs" variant="soft" icon="i-heroicons-pencil" @click="state.channelEditing = true">
							编辑
						</UButton>
					</div>
				</template>
				<div class="space-y-4">
					<UFormGroup label="渠道 SKU ID">
						<UInput
							v-model="channelForm.channelSkuId"
							placeholder="SKU-E2E"
							data-testid="channel-sku-input"
							aria-label="渠道 SKU ID"
						/>
					</UFormGroup>
					<UFormGroup label="状态">
						<USelect
							v-model="channelForm.status"
							:options="channelStatusOptions"
							data-testid="channel-status-select"
							aria-label="渠道状态"
						/>
					</UFormGroup>
					<div class="flex gap-2">
						<UButton color="primary" @click="saveChannel">保存映射</UButton>
						<UButton variant="soft" @click="publishChannel">推送发布</UButton>
					</div>
					<div class="space-y-1 text-sm text-gray-600 dark:text-gray-400">
						<p v-if="state.channelMessage">{{ state.channelMessage }}</p>
						<p v-if="state.publishMessage">{{ state.publishMessage }}</p>
					</div>
				</div>
			</UCard>

			<UCard id="us3-barcode">
				<template #header>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">条码管理</h3>
				</template>
				<div class="space-y-4">
					<UFormGroup label="条码前缀">
						<UInput
							v-model="barcodePrefix"
							placeholder="QA"
							data-testid="barcode-prefix-input"
							aria-label="条码前缀"
						/>
					</UFormGroup>
					<UButton color="primary" @click="generateBarcodes">生成条码</UButton>
					<div v-if="barcodeResults.length" class="space-y-1">
						<p class="text-sm text-gray-500">最新生成：</p>
						<p v-for="code in barcodeResults" :key="code" class="font-mono text-primary-600">{{ code }}</p>
					</div>
				</div>
			</UCard>

			<UCard id="us3-serials">
				<template #header>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">序列号记录</h3>
				</template>
				<div class="space-y-4">
					<UFormGroup label="序列号">
						<UInput
							v-model="serialInput"
							placeholder="SER-001"
							data-testid="serial-input"
							aria-label="序列号"
						/>
					</UFormGroup>
					<UButton color="primary" @click="saveSerial">保存记录</UButton>
					<p v-if="state.serialMessage" class="text-emerald-600">{{ state.serialMessage }}</p>
					<div class="space-y-1">
						<p v-for="serial in serials" :key="serial" class="font-mono text-sm text-gray-700 dark:text-gray-200">
							{{ serial }}
						</p>
					</div>
				</div>
			</UCard>
		</section>
	</div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { createError } from 'h3'
import { useToast } from '#imports'

if (process.env.NODE_ENV !== 'development') {
	throw createError({ statusCode: 404, statusMessage: 'Not Found' })
}

const toast = useToast()
const state = reactive({
	previewReady: false,
	generatorMessage: '',
	channelMessage: '',
	publishMessage: '',
	channelEditing: false,
	bulkTaskMessage: '',
	bulkTaskId: '',
	serialMessage: '',
})

const candidateCount = 4
const bulkForm = reactive({
	adjustment: 5,
	note: '',
})

const channelForm = reactive({
	channelSkuId: 'SKU-E2E',
	status: 'pending',
})

const channelStatusOptions = [
	{ label: '待发布', value: 'pending' },
	{ label: '已发布', value: 'published' },
	{ label: '失败', value: 'failed' },
]

const barcodePrefix = ref('QA')
const barcodeResults = ref<string[]>([])
const serialInput = ref('')
const serials = ref<string[]>([])

const generatePreview = () => {
	state.previewReady = true
	state.generatorMessage = `共 ${candidateCount} 条候选可写入`
}

const writeSku = () => {
	if (!state.previewReady) {
		toast.add({ title: '请先生成预览', color: 'warning' })
		return
	}
	state.generatorMessage = '已写入 1 条 SKU'
	toast.add({ title: '已写入 1 条 SKU', color: 'primary' })
}

const submitBulkTask = () => {
	state.bulkTaskMessage = '批量任务已创建'
	state.bulkTaskId = '#task-e2e'
	toast.add({ title: '批量任务已创建', color: 'primary' })
}

const saveChannel = () => {
	state.channelMessage = '渠道映射已保存'
	toast.add({ title: '渠道映射已保存', color: 'success' })
}

const publishChannel = () => {
	state.publishMessage = '发布已触发'
	toast.add({ title: '发布已触发', color: 'primary' })
}

const generateBarcodes = () => {
	const prefix = barcodePrefix.value.trim() || 'QA'
	barcodeResults.value = [`BARCODE-${prefix}-1`, `BARCODE-${prefix}-2`]
	toast.add({ title: '条码已生成', color: 'primary' })
}

const saveSerial = () => {
	const serial = (serialInput.value || 'SER-001').toUpperCase()
	serials.value.unshift(serial)
	serialInput.value = ''
	state.serialMessage = '序列号已保存'
	toast.add({ title: '序列号已保存', color: 'primary' })
}
</script>
