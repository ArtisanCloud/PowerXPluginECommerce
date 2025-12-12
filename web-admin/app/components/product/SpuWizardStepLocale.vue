<template>
	<SpuLocaleForm ref="formRef" v-model="localValue" @validation="(state) => (isValid = state)" />
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import SpuLocaleForm from '~/components/product/SpuLocaleForm.vue'

const props = defineProps<{
	modelValue: { defaultLocale: string; locales: Array<Record<string, any>> }
}>()
const emit = defineEmits<{
	'update:modelValue': [{ defaultLocale: string; locales: Array<Record<string, any>> }]
}>()

const cloneLocales = (entries: Array<Record<string, any>> = []) =>
	entries.map((entry) => ({ ...entry }))

const localValue = reactive({
	defaultLocale: 'zh-CN',
	locales: [{ locale: 'zh-CN', title: '', description: '' }],
})

const formRef = ref<InstanceType<typeof SpuLocaleForm> | null>(null)
let isValid = true

watch(
	() => props.modelValue,
	(val) => {
		if (!val) return
		localValue.defaultLocale = val.defaultLocale || 'zh-CN'
		localValue.locales = val.locales?.length ? cloneLocales(val.locales) : [{ locale: 'zh-CN', title: '', description: '' }]
	},
	{ immediate: true }
)

watch(
	localValue,
	() =>
		emit('update:modelValue', {
			defaultLocale: localValue.defaultLocale,
			locales: cloneLocales(localValue.locales),
		}),
	{ deep: true }
)

const validate = () => formRef.value?.validate() ?? isValid

defineExpose({ validate })
</script>
