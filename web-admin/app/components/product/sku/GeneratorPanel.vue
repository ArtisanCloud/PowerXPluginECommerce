<template>
	<UCard class="space-y-6">
		<template #header>
			<div class="flex flex-wrap items-center justify-between gap-4">
				<div>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">SKU 生成器</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">
						选择规格组合并设置默认字段，一键生成 SKU。
					</p>
				</div>
				<div class="flex gap-2">
					<UButton
						variant="ghost"
						color="gray"
						:disabled="!canGenerate"
						:loading="generatorLoading"
						@click="handleGenerate"
					>
						生成预览
					</UButton>
					<UButton
						color="primary"
						:disabled="!hasSelectableCandidates"
						:loading="createLoading"
						@click="commitSkus"
					>
						写入 SKU
					</UButton>
				</div>
			</div>
		</template>

		<div class="grid gap-4 lg:grid-cols-3">
			<div class="space-y-4 lg:col-span-1">
				<h4 class="text-base font-medium text-gray-900 dark:text-white">默认字段</h4>
				<DefaultValueForm v-model="defaults" />
			</div>
			<div class="space-y-4 lg:col-span-2">
				<h4 class="text-base font-medium text-gray-900 dark:text-white">规格选择</h4>
				<div v-if="!specSelections.length" class="rounded border border-dashed p-4 text-sm text-gray-500">
					未检测到规格配置，请在 SPU 规格设置完成后再试。
				</div>
				<div v-else class="space-y-4">
					<div v-for="spec in specSelections" :key="spec.id" class="space-y-2">
						<div class="flex items-center justify-between">
							<label class="text-sm font-medium">{{ spec.name }}</label>
							<span class="text-xs text-gray-500">{{ selections[spec.id]?.length || 0 }} / {{ spec.values.length }}</span>
						</div>
						<USelectMenu
							v-model="selections[spec.id]"
							:items="spec.values.map((val) => ({ label: val.name, value: val.id }))"
							multiple
							placeholder="选择需要生成的规格值"
						/>
					</div>
				</div>
			</div>
		</div>

		<div v-if="generatorLoading" class="flex justify-center py-6">
			<UIcon name="i-heroicons-arrow-path" class="h-8 w-8 animate-spin text-primary-500" />
		</div>
		<div v-else>
			<div v-if="!candidates.length" class="rounded border border-dashed p-4 text-sm text-gray-500">
				点击“生成预览”以查看候选结果。
			</div>
			<div v-else class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800 text-sm">
					<thead class="bg-gray-50 dark:bg-gray-800/50">
						<tr>
							<th class="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">选择</th>
							<th class="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">规格组合</th>
							<th class="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">预览编码</th>
							<th class="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">冲突</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200 dark:divide-gray-800">
						<tr v-for="(candidate, idx) in candidates" :key="idx" class="hover:bg-gray-50 dark:hover:bg-gray-800/30">
							<td class="px-4 py-2">
								<UCheckbox v-model="candidate.selected" :disabled="candidate.exists" />
							</td>
							<td class="px-4 py-2">
								<div class="space-y-1">
									<p v-if="!candidate.specs.length" class="text-gray-500 dark:text-gray-400">默认 SKU</p>
									<p v-else>
										<span
											v-for="spec in candidate.specs"
											:key="spec.specId"
											class="mr-2 inline-flex items-center rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-700 dark:bg-gray-800 dark:text-gray-200"
										>
											{{ spec.specName || spec.specId }}：{{ spec.valueName || spec.valueId }}
										</span>
									</p>
								</div>
							</td>
							<td class="px-4 py-2 font-mono text-sm">{{ candidate.previewSkuCode }}</td>
							<td class="px-4 py-2">
								<div v-if="candidate.exists" class="flex items-center gap-1 text-amber-600 dark:text-amber-400">
									<UIcon name="i-heroicons-exclamation-triangle" />
									已有 SKU
								</div>
								<div v-else-if="candidate.conflictReasons?.length" class="text-amber-600 dark:text-amber-400">
									{{ candidate.conflictReasons.join(', ') }}
								</div>
								<div v-else class="text-emerald-600 dark:text-emerald-400">可写入</div>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	</UCard>
</template>

<script setup lang="ts">
import { useToast } from '#imports'
import DefaultValueForm from './DefaultValueForm.vue'
import { useProductSkuStore } from '~/stores/productSku'
import type { SkuGeneratorDefaults, SkuGeneratorRequest, SkuSpecSelection } from '~/types/product/sku'

interface SpecDefinition {
	id: string
	name: string
	values: Array<{ id: string; name: string; code?: string }>
}

const props = defineProps<{
	spuId: string
	specs?: SpecDefinition[]
}>()
const emit = defineEmits<{ generated: [void] }>()

const store = useProductSkuStore()
const toast = useToast()
const defaults = ref<SkuGeneratorDefaults>({ ...store.generatorDefaults })
const selections = reactive<Record<string, string[]>>({})

const specSelections = computed(() => props.specs ?? [])
const candidates = computed(() => store.generatorCandidates)
const generatorLoading = computed(() => store.generatorLoading)
const createLoading = computed(() => store.upsertLoading)
const canGenerate = computed(() => Boolean(props.spuId) && specSelections.value.length > 0)
const hasSelectableCandidates = computed(() =>
	candidates.value.some((candidate) => candidate.selected && !candidate.exists),
)

watch(
	() => props.specs,
	(next) => {
		if (!next?.length) return
		next.forEach((spec) => {
			if (!Array.isArray(selections[spec.id]) || selections[spec.id].length === 0) {
				selections[spec.id] = spec.values.map((val) => val.id)
			}
		})
	},
	{ immediate: true, deep: true },
)

watch(
	() => defaults.value,
	(val) => {
		store.setGeneratorDefaults(val)
	},
	{ deep: true },
)

const buildRequest = (): SkuGeneratorRequest => ({
	specSelections: specSelections.value.map((spec) => ({
		specId: spec.id,
		specName: spec.name,
		valueIds: selections[spec.id]?.length ? selections[spec.id] : spec.values.map((val) => val.id),
		values: spec.values.map((val) => ({
			valueId: val.id,
			valueName: val.name,
			valueCode: val.code,
		})),
	})),
	defaults: defaults.value,
})

const handleGenerate = async () => {
	if (!canGenerate.value) return
	try {
		await store.fetchGeneratorCandidates(props.spuId, buildRequest())
		toast.add({ title: '已生成预览组合', color: 'primary' })
	} catch (error) {
		console.error(error)
		toast.add({ title: '生成失败', description: (error as any)?.message || '请稍后再试', color: 'red' })
	}
}

const commitSkus = async () => {
	const selected = candidates.value.filter((candidate) => candidate.selected && !candidate.exists)
	if (!selected.length) {
		toast.add({ title: '请选择需要写入的组合', color: 'amber' })
		return
	}
	try {
		const payload = {
			skus: selected.map((candidate) => ({
				spuId: props.spuId,
				skuCode: candidate.previewSkuCode,
				specs: candidate.specs,
				barcode: buildBarcode(candidate.previewSkuCode),
				status: 'draft',
				minOrderQty: defaults.value.minOrderQty ?? 0,
				defaultValues: candidate.defaultValues || defaults.value,
			})),
		}
		await store.createSkus(payload)
		toast.add({ title: `已写入 ${selected.length} 条 SKU`, color: 'emerald' })
		emit('generated')
	} catch (error) {
		console.error(error)
		toast.add({ title: '写入失败', description: (error as any)?.message || '请稍后重试', color: 'red' })
	}
}

const buildBarcode = (previewCode: string) => {
	const prefix = defaults.value.barcodePrefix?.trim() || ''
	return `${prefix}${previewCode}`.replace(/--+/g, '-')
}
</script>
