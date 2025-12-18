<template>
	<SpuLocaleForm ref="formRef" v-model="localValue" @validation="(state) => (isValid = state)" />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useVModel } from '@vueuse/core'
import SpuLocaleForm from '~/components/product/SpuLocaleForm.vue'

const props = defineProps<{
	modelValue: { defaultLocale: string; locales: Array<Record<string, any>> }
}>()
const emit = defineEmits<{
	'update:modelValue': [{ defaultLocale: string; locales: Array<Record<string, any>> }]
}>()

const localValue = useVModel(props, 'modelValue', emit, {
	passive: true,
	deep: true,
	defaultValue: {
		defaultLocale: 'zh-CN',
		locales: [{ locale: 'zh-CN', title: '', description: '' }],
	},
})

const formRef = ref<InstanceType<typeof SpuLocaleForm> | null>(null)
let isValid = true

const validate = () => formRef.value?.validate() ?? isValid

defineExpose({ validate })
</script>
