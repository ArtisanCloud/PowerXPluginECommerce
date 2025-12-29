<template>
	<div class="grid gap-4 md:grid-cols-2">
		<UFormField label="条码前缀">
			<template #default="{ id }">
				<UInput :id="id" v-model.trim="localValue.barcodePrefix" placeholder="SKU-" />
			</template>
		</UFormField>
		<UFormField label="起订量">
			<template #default="{ id }">
				<UInput :id="id" v-model.number="localValue.minOrderQty" type="number" min="0" />
			</template>
		</UFormField>
		<UFormField label="成本价">
			<template #default="{ id }">
				<UInput :id="id" v-model.number="localValue.costPrice" type="number" min="0" step="0.01" />
			</template>
		</UFormField>
		<UFormField label="重量 (kg)">
			<template #default="{ id }">
				<UInput :id="id" v-model.number="localValue.weight" type="number" min="0" step="0.01" />
			</template>
		</UFormField>
		<UFormField label="尺寸/包装">
			<template #default="{ id }">
				<UInput :id="id" v-model.trim="localValue.dimensions" placeholder="10x20x5cm" />
			</template>
		</UFormField>
	</div>
</template>

<script setup lang="ts">
import type { SkuGeneratorDefaults } from '~/types/product/sku'

const modelValue = defineModel<SkuGeneratorDefaults>({ required: true })
const localValue = reactive<SkuGeneratorDefaults>({ ...(modelValue.value || {}) })

watch(
	() => modelValue.value,
	(next) => {
		Object.assign(localValue, next || {})
	},
	{ deep: true, immediate: true }
)

watch(
	localValue,
	(next) => {
		modelValue.value = { ...next }
	},
	{ deep: true }
)
</script>
