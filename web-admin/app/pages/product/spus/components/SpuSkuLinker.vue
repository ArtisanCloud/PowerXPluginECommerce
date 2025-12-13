<template>
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
		<div v-if="store.skusLoading">
			<USkeleton class="h-24" />
		</div>
		<div v-else class="space-y-4">
			<div v-for="(row, idx) in rows" :key="row.localId" class="space-y-3 rounded border p-4">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<strong>SKU {{ idx + 1 }}</strong>
					<div class="flex gap-2">
						<UButton size="xs" color="gray" variant="ghost" icon="i-heroicons-document-duplicate" @click="cloneRow(idx)">
							复制
						</UButton>
						<UButton size="xs" color="gray" variant="ghost" icon="i-heroicons-trash" @click="removeRow(idx)" :disabled="rows.length === 1">
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

<script setup lang="ts">
import { useToast } from '#imports'
import type { SpuSkuLink } from '~/composables/api/useSpu'
import { useSpuStore } from '~/stores/product/spu'

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

const props = defineProps<{ spuId: string }>()
const store = useSpuStore()
const toast = useToast()
const saving = ref(false)
const rows = reactive<LocalSkuRow[]>([])

const createLocalId = () =>
	typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
		? crypto.randomUUID()
		: Math.random().toString(36).slice(2)

const syncRows = (items = store.skus) => {
	rows.splice(0, rows.length)
	if (!items.length) {
		rows.push(createRow())
		return
	}
	items.forEach((item) => {
		rows.push({
			localId: createLocalId(),
			id: item.id,
			code: item.code,
			name: item.name,
			inventoryRef: item.inventoryRef,
			pricing: { ...item.pricing },
			cloneFrom: item.cloneFrom,
			attributesText: JSON.stringify(item.attributes ?? {}, null, 0),
		})
	})
}

onMounted(async () => {
	await store.fetchSkus(props.spuId)
	syncRows()
})

watch(
	() => store.skus,
	(newItems) => syncRows(newItems),
	{ deep: true }
)

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

const cloneOptions = (idx: number) =>
	rows
		.filter((_, rowIdx) => rowIdx !== idx)
		.map((row) => ({ label: `${row.code} · ${row.name}`, value: row.id ?? row.code }))

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
</script>
