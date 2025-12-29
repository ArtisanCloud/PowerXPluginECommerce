<template>
		<div class="space-y-6">
			<UTabs v-model="activePanel" :items="panelTabs" class="w-full">
				<template #manual>
					<UCard class="space-y-4">
				<template #header>
					<div class="flex flex-wrap items-center justify-between gap-4">
						<div>
							<h3 class="text-lg font-semibold text-gray-900 dark:text-white">关联 SKU</h3>
							<p class="text-sm text-gray-500 dark:text-gray-400">批量创建/复制 SKU 并同步价格、库存引用。</p>
					</div>
					<div class="flex gap-2">
						<UButton icon="i-heroicons-plus" @click="addRow">新增 SKU</UButton>
						<UButton color="primary" :loading="saving" @click="handleSave">保存关联</UButton>
					</div>
				</div>
			</template>
			<div v-if="skuListLoading">
				<USkeleton class="h-24" />
			</div>
			<div v-else class="space-y-4">
				<div v-for="(row, idx) in rows" :key="row.localId" class="space-y-3 rounded border p-4">
					<div class="flex flex-wrap items-center justify-between gap-2">
						<strong>SKU {{ idx + 1 }}</strong>
						<div class="flex gap-2">
							<UButton size="xs" color="primary" variant="solid" icon="i-heroicons-document-duplicate" @click="cloneRow(idx)">
								复制
							</UButton>
							<UButton
								size="xs"
								color="error"
								variant="solid"
								icon="i-heroicons-trash"
								@click="removeRow(idx)"
								:disabled="rows.length === 1"
							>
								移除
							</UButton>
						</div>
					</div>
					<div class="grid gap-4 md:grid-cols-3">
						<UFormField label="SKU 编码" required>
							<template #default="{ id }">
								<UInput :id="id" v-model.trim="row.code" :data-testid="`sku-code-${idx}`" placeholder="SKU-0001" />
							</template>
						</UFormField>
						<UFormField label="SKU 名称" required>
							<template #default="{ id }">
								<UInput :id="id" v-model.trim="row.name" :data-testid="`sku-name-${idx}`" placeholder="标准版" />
							</template>
						</UFormField>
						<UFormField label="库存引用">
							<template #default="{ id }">
								<UInput :id="id" v-model.trim="row.inventoryRef" placeholder="inventory-id" />
							</template>
						</UFormField>
						<UFormField label="价格 (含税)" required>
							<template #default="{ id }">
								<UInput :id="id" v-model.number="row.pricing.price" type="number" min="0" step="0.01" />
							</template>
						</UFormField>
						<UFormField label="币种">
							<template #default="{ id }">
								<UInput :id="id" v-model.trim="row.pricing.currency" maxlength="3" placeholder="CNY" />
							</template>
						</UFormField>
						<UFormField label="克隆来源">
							<template #default="{ id }">
							<USelect
								:id="id"
								v-model="row.cloneFrom"
								:items="cloneOptions(idx)"
								:disabled="!hasCloneChoice(idx)"
								placeholder="选择已有 SKU"
							/>
						</template>
					</UFormField>
				</div>
				<UFormField label="扩展属性 JSON">
					<template #default="{ id }">
						<UTextarea :id="id" v-model="row.attributesText" :rows="2" placeholder='{"color":"red"}' />
					</template>
				</UFormField>
				</div>
			</div>
					</UCard>
				</template>
				<template #generator>
					<GeneratorPanel :spu-id="props.spuId" :specs="props.specs" @generated="handleGeneratorResult" />
				</template>
			</UTabs>
		</div>
	</template>

<script setup lang="ts">
import { useToast } from '#imports'
import type { SpuSkuLink } from '~/composables/api/useSpu'
import { useSkuApi } from '~/composables/api/useSku'
import GeneratorPanel from '~/components/product/sku/GeneratorPanel.vue'
import { useSpuStore } from '~/stores/product/spu'
import type { ProductSku } from '~/types/product/sku'

interface LocalSkuRow {
	localId: string
	id?: string
	code: string
	name: string
	inventoryRef?: string
	pricing: { price: number; currency: string }
	cloneFrom?: string
	attributesText: string
}

interface GeneratorSpecDefinition {
	id: string
	name: string
	values: Array<{ id: string; name: string; code?: string }>
}

const props = defineProps<{ spuId: string; specs?: GeneratorSpecDefinition[] }>()
const emit = defineEmits<{ saved: [] }>()
const store = useSpuStore()
const skuApi = useSkuApi()
const toast = useToast()
const saving = ref(false)
const rows = reactive<LocalSkuRow[]>([])
const skuListLoading = ref(false)
const activePanel = ref<'generator' | 'manual'>('manual')
const panelTabs = [
	{ label: '手动关联', slot: 'manual', value: 'manual' },
	{ label: '批量生成', slot: 'generator', value: 'generator' },
]

const createLocalId = () =>
	typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
		? crypto.randomUUID()
		: Math.random().toString(36).slice(2)

const syncRows = (official: ProductSku[], payload: SpuSkuLink[]) => {
	rows.splice(0, rows.length)
	const payloadMap = new Map<string, SpuSkuLink>()
	;(payload ?? []).forEach((item) => {
		const key = item.code?.toLowerCase()
		if (key) {
			payloadMap.set(key, item)
		}
	})

	if (official.length) {
		official.forEach((sku) => {
			const key = sku.skuCode?.toLowerCase()
			const linked = key ? payloadMap.get(key) : undefined
			rows.push(composeRowFromSources(sku, linked))
		})
	}

	if (!rows.length) {
		rows.push(createRow())
	}
}

const composeRowFromSources = (sku: ProductSku, payload?: SpuSkuLink): LocalSkuRow => ({
	localId: createLocalId(),
	id: payload?.id ?? sku.id,
	code: sku.skuCode || payload?.code || '',
	name: payload?.name || sku.skuCode || '',
	inventoryRef: payload?.inventoryRef,
	pricing: normalizePricing(payload?.pricing),
	cloneFrom: payload?.cloneFrom,
	attributesText: stringifyAttributes(payload?.attributes),
})

const composeRowFromPayload = (payload: SpuSkuLink): LocalSkuRow => ({
	localId: createLocalId(),
	id: payload.id,
	code: payload.code,
	name: payload.name,
	inventoryRef: payload.inventoryRef,
	pricing: normalizePricing(payload.pricing),
	cloneFrom: payload.cloneFrom,
	attributesText: stringifyAttributes(payload.attributes),
})

const normalizePricing = (pricing?: { price: number; currency: string }) => {
	const price = typeof pricing?.price === 'number' ? pricing.price : 0
	const currency = pricing?.currency || 'CNY'
	return { price, currency }
}

const stringifyAttributes = (attributes?: Record<string, any>) => {
	if (!attributes) {
		return ''
	}
	try {
		return JSON.stringify(attributes, null, 2)
	} catch {
		return ''
	}
}

const refreshSkus = async (shouldNotify = false) => {
	skuListLoading.value = true
	try {
		const [officialResp, payload] = await Promise.all([
			skuApi.listBySpu(props.spuId, { pageSize: 200 }).catch(() => ({ items: [] })),
			store.fetchSkus(props.spuId).catch(() => []),
		])
		const officialItems = (officialResp?.items ?? []).map(normalizeOfficialSku)
		const payloadItems = payload ?? []
		syncRows(officialItems, payloadItems)
		if (shouldNotify) {
			emit('saved')
		}
	} catch (error) {
		console.error(error)
		toast.add({ title: '加载 SKU 失败', color: 'error' })
	} finally {
		skuListLoading.value = false
	}
}

const handleGeneratorResult = async () => {
	await refreshSkus(true)
}

onMounted(refreshSkus)

const addRow = () => {
	rows.push(createRow())
}

const cloneRow = (idx: number) => {
	const target = rows[idx]
	if (!target) return
	rows.push({
		localId: createLocalId(),
		code: `${target.code}-copy`,
		name: `${target.name} Copy`,
		inventoryRef: target.inventoryRef,
		pricing: { ...target.pricing },
		cloneFrom: target.id ?? target.code,
		attributesText: target.attributesText,
	})
}

const removeRow = (idx: number) => {
	if (rows.length === 1) return
	rows.splice(idx, 1)
}

const cloneOptions = (idx: number) => {
	const options = rows
		.filter((_, rowIdx) => rowIdx !== idx && (rows[rowIdx].id || rows[rowIdx].code))
		.map((row) => ({
			label: `${row.code || '未命名'} · ${row.name || '未命名'}`,
			value: row.id ?? row.code!,
		}))
	return options.length ? options : [{ label: '暂无可克隆 SKU', value: '__placeholder__', disabled: true }]
}

const hasCloneChoice = (idx: number) => cloneOptions(idx).some((option) => !option.disabled)

const handleSave = async () => {
	try {
		saving.value = true
		const payload = rows.map<SpuSkuLink>((row) => ({
			id: row.id,
			code: row.code,
			name: row.name,
			inventoryRef: row.inventoryRef,
			pricing: row.pricing,
			cloneFrom: row.cloneFrom,
			attributes: parseAttributes(row.attributesText),
		}))
		await store.saveSkus(props.spuId, payload)
		toast.add({ title: 'SKU 关联已保存' })
		await refreshSkus(true)
	} catch (error) {
		console.error(error)
		toast.add({ title: '保存失败', description: '请检查字段是否完整', color: 'red' })
	} finally {
		saving.value = false
	}
}

const parseAttributes = (text: string) => {
	if (!text) return undefined
	try {
		return JSON.parse(text)
	} catch {
		return undefined
	}
}

const createRow = (): LocalSkuRow => ({
	localId: createLocalId(),
	code: '',
	name: '',
	pricing: { price: 0, currency: 'CNY' },
	attributesText: '',
})

const normalizeOfficialSku = (item: any): ProductSku => ({
	id: item?.id,
	spuId: item?.spuId ?? item?.spu_id,
	skuCode: item?.skuCode ?? item?.sku_code ?? '',
	status: item?.status,
	barcode: item?.barcode,
	createdAt: item?.createdAt ?? item?.created_at,
	updatedAt: item?.updatedAt ?? item?.updated_at,
	specs: item?.specs ?? [],
	minOrderQty: item?.minOrderQty ?? item?.min_order_qty,
})
</script>
