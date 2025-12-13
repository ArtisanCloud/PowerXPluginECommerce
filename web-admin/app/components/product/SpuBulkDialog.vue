<template>
	<UModal v-model="internalOpen" :ui="modalUi" :close="false" :dismissible="false">
		<template #header>
			<div class="text-lg font-semibold">
				{{ mode === 'import' ? '批量导入 SPU' : '批量导出 SPU' }}
			</div>
		</template>
		<template #body>
			<div class="space-y-4">
				<div v-if="mode === 'import'" class="space-y-3">
					<UFormField label="导入模板" :ui="inlineFieldUi">
						<USelect v-model="templateId" class="w-full" :items="templateOptions" />
					</UFormField>
					<UFormField label="导入文件" :ui="inlineFieldUi">
						<UInput type="file" accept=".csv,.xlsx" @change="handleFileChange" />
					</UFormField>
					<p class="text-sm text-gray-500">支持 CSV/Excel，50 行以上建议分批上传。</p>
				</div>
				<div v-else class="space-y-4">
					<UFormField label="导出字段" :ui="inlineFieldUi">
						<USelectMenu
							v-model="selectedFields"
							:options="fieldOptions"
							multiple
							class="w-full"
							placeholder="选择字段"
						/>
					</UFormField>
					<UFormField label="状态筛选" :ui="inlineFieldUi">
						<USelect v-model="exportStatus" :items="statusOptions" class="w-full" />
					</UFormField>
					<UFormField label="关键字" :ui="inlineFieldUi">
						<UInput v-model="exportKeyword" placeholder="编码或名称关键字" class="w-full" />
					</UFormField>
				</div>
			</div>
		</template>
		<template #footer>
			<div class="flex w-full justify-end gap-2">
				<UButton variant="ghost" @click="close">取消</UButton>
				<UButton color="primary" :loading="submitting" @click="handleSubmit">
					{{ mode === 'import' ? '开始导入' : '开始导出' }}
				</UButton>
			</div>
		</template>
	</UModal>
</template>

<script setup lang="ts">
import { useToast } from '#imports'
import { useSpuApi } from '~/composables/api/useSpu'

interface Props {
	mode: 'import' | 'export'
	modelValue: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
	(event: 'update:modelValue', value: boolean): void
	(event: 'submitted', payload: { taskId: string; mode: 'import' | 'export' }): void
}>()

const toast = useToast()
const api = useSpuApi()
const internalOpen = computed({
	get: () => props.modelValue,
	set: (val) => emit('update:modelValue', val),
})
const submitting = ref(false)
const templateId = ref('default')
const fileRef = ref<File | null>(null)
const selectedFields = ref<string[]>(['code', 'name', 'status'])
const exportStatus = ref<string | null>(null)
const exportKeyword = ref('')

const templateOptions = [
	{ label: '默认模板', value: 'default' },
	{ label: '订阅模板', value: 'subscription' },
]
const fieldOptions = [
	{ label: '编码', value: 'code' },
	{ label: '名称', value: 'name' },
	{ label: '类型', value: 'type' },
	{ label: '状态', value: 'status' },
	{ label: '更新时间', value: 'updatedAt' },
]
const statusOptions = [
	{ label: '全部', value: null },
	{ label: '草稿', value: 'draft' },
	{ label: '审核中', value: 'reviewing' },
	{ label: '已发布', value: 'published' },
]
const inlineFieldUi = {
	root: 'flex flex-col space-y-2',
	label: 'text-sm font-medium text-gray-500',
	container: 'mt-0 w-full',
}
const modalUi = {
	content: 'max-w-2xl w-full',
	body: 'p-4 sm:p-5 space-y-4',
	header: 'p-4 sm:px-5',
	footer: 'p-4 sm:px-5',
}

const handleFileChange = (event: Event) => {
	const target = event.target as HTMLInputElement
	const files = target.files
	fileRef.value = files && files.length ? files[0] : null
}

const close = () => {
	internalOpen.value = false
	if (props.mode === 'import') {
		fileRef.value = null
	}
}

const handleSubmit = async () => {
	try {
		submitting.value = true
		let taskId = ''
		if (props.mode === 'import') {
			if (!fileRef.value) {
				toast.add({ title: '请选择导入文件', color: 'red' })
				return
			}
			const formData = new FormData()
			formData.append('templateId', templateId.value)
			formData.append('file', fileRef.value)
			const result = await api.importSpus(formData)
			taskId = result?.taskId || ''
		} else {
			const payload = {
				fields: selectedFields.value.length ? selectedFields.value : ['code', 'name', 'status'],
				filters: {
					status: exportStatus.value ?? undefined,
					keyword: exportKeyword.value || undefined,
				},
			}
			const result = await api.exportSpus(payload)
			taskId = result?.taskId || ''
		}
		if (taskId) {
			toast.add({ title: '任务已创建', description: `ID：${taskId}` })
			emit('submitted', { taskId, mode: props.mode })
			close()
		} else {
			toast.add({ title: '任务创建失败', color: 'red' })
		}
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '操作失败', color: 'red' })
	} finally {
		submitting.value = false
	}
}
</script>
