<template>
	<UCard>
		<template #header>
			<div class="flex items-center justify-between">
				<div>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">渠道可见性</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">配置各渠道上架时间与展示信息。</p>
				</div>
				<UButton size="sm" variant="ghost" icon="i-heroicons-arrow-path" @click="loadChannels" :loading="loading">
					刷新
				</UButton>
			</div>
		</template>
		<div class="space-y-6">
			<div class="grid gap-4 md:grid-cols-2">
				<UFormField label="渠道" :ui="inlineFieldUi">
					<USelect v-model="form.channel" :items="channelOptions" class="w-full" />
				</UFormField>
				<UFormField label="上架状态" :ui="inlineFieldUi">
					<USelect v-model="form.availability" :items="availabilityOptions" class="w-full" />
				</UFormField>
				<UFormField label="上架时间" :ui="inlineFieldUi">
					<UInput v-model="form.publishAt" type="datetime-local" class="w-full" />
				</UFormField>
				<UFormField label="下架时间" :ui="inlineFieldUi">
					<UInput v-model="form.withdrawAt" type="datetime-local" class="w-full" />
				</UFormField>
			</div>
			<UFormField label="渠道标题" :ui="inlineFieldUi">
				<UInput v-model="form.title" placeholder="可选，渠道展示标题" class="w-full" />
			</UFormField>
				<UFormField label="备注" :ui="inlineFieldUi">
					<UTextarea v-model="form.description" :rows="3" placeholder="渠道文案/差异化说明" />
				</UFormField>
			<div class="flex flex-wrap justify-end gap-2">
				<UButton variant="ghost" @click="resetForm">重置</UButton>
				<UButton color="primary" :loading="saving" @click="handleSave">保存渠道</UButton>
			</div>
			<div class="space-y-3">
				<div class="flex items-center justify-between">
					<h4 class="text-base font-semibold text-gray-900 dark:text-white">已配置渠道</h4>
					<p class="text-sm text-gray-500 dark:text-gray-400">共 {{ channels.length }} 个</p>
				</div>
				<UTable :data="pagedChannels" :columns="columns" :loading="loading" class="w-full">
					<template #availability-cell="{ row }">
						<UBadge :color="badgeColor(row.original.availability)">
							{{ availabilityLabel(row.original.availability) }}
						</UBadge>
					</template>
					<template #schedule-cell="{ row }">
					<div class="text-sm text-gray-600 dark:text-gray-300">
							<div>上架：{{ formatTime(row.original.publishAt) || '—' }}</div>
							<div>下架：{{ formatTime(row.original.withdrawAt) || '—' }}</div>
						</div>
					</template>
					<template #actions-cell="{ row }">
						<div class="flex gap-2">
							<UButton size="xs" variant="soft" @click="handleEdit(row.original)">编辑</UButton>
							<UButton
								size="xs"
								color="red"
								variant="soft"
								@click="handleRemove(row.original)"
								:loading="removing === row.original.channel"
							>
								删除
							</UButton>
						</div>
					</template>
				</UTable>
				<div v-if="channels.length" class="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm">
					<span class="text-gray-500 dark:text-gray-400">
						共 {{ channels.length }} 条 · 第 {{ channelPagination.page }} / {{ channelPageCount }} 页
					</span>
					<div class="flex items-center gap-3">
						<USelect v-model="channelPagination.pageSize" :items="channelPageSizeOptions" class="w-28" />
						<UPagination v-model="channelPagination.page" :page-count="channelPageCount" />
					</div>
				</div>
				<UAlert v-if="!channels.length && !loading" color="gray">
					暂无渠道配置，可通过上方表单新增。
				</UAlert>
			</div>
		</div>
	</UCard>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useToast } from '#imports'
import type { TableColumn } from '@nuxt/ui'
import type { SpuChannelEntry, SpuChannelPayload } from '~/composables/api/useSpu'
import { useSpuApi } from '~/composables/api/useSpu'
import { channelSelectOptions } from '~/data/channelCatalog'

const props = defineProps<{ spuId?: string }>()

const toast = useToast()
const api = useSpuApi()
const channels = ref<SpuChannelEntry[]>([])
const loading = ref(false)
const saving = ref(false)
const removing = ref<string | null>(null)
const channelPagination = reactive({ page: 1, pageSize: 5 })
const channelPageSizeOptions = [
	{ label: '5 / 页', value: 5 },
	{ label: '10 / 页', value: 10 },
	{ label: '20 / 页', value: 20 },
]
const inlineFieldUi = {
	root: 'flex flex-col gap-2',
	label: 'text-sm font-medium text-gray-500 dark:text-gray-400',
	container: 'mt-0 w-full',
}
const channelOptions = channelSelectOptions
const availabilityOptions = [
	{ label: '未上架', value: 'unlisted' },
	{ label: '已排期', value: 'scheduled' },
	{ label: '已发布', value: 'published' },
	{ label: '渠道下架', value: 'withheld' },
]
const columns = computed<TableColumn<SpuChannelEntry>[]>(() => [
	{ accessorKey: 'channel', header: '渠道' },
	{ accessorKey: 'availability', header: '状态' },
	{ id: 'schedule', header: '上下线' },
	{ id: 'actions', header: '操作' },
])
const pagedChannels = computed(() => {
	const start = (channelPagination.page - 1) * channelPagination.pageSize
	return channels.value.slice(start, start + channelPagination.pageSize)
})
const channelPageCount = computed(() => {
	const total = Math.ceil(channels.value.length / channelPagination.pageSize)
	return total > 0 ? total : 1
})
const form = reactive<{
	channel: string
	availability: string
	publishAt: string
	withdrawAt: string
	title: string
	description: string
}>({
	channel: 'official',
	availability: 'unlisted',
	publishAt: '',
	withdrawAt: '',
	title: '',
	description: '',
})
const editingChannel = ref<string | null>(null)

const badgeColor = (value: string) => {
	switch (value) {
		case 'published':
			return 'emerald'
		case 'scheduled':
			return 'primary'
		case 'withheld':
			return 'orange'
		default:
			return 'gray'
	}
}
const availabilityLabel = (value: string) => {
	const option = availabilityOptions.find((opt) => opt.value === value)
	return option?.label || value
}
const formatTime = (value?: string) => {
	if (!value) return ''
	return new Date(value).toLocaleString()
}
const resetForm = () => {
	form.channel = 'official'
	form.availability = 'unlisted'
	form.publishAt = ''
	form.withdrawAt = ''
	form.title = ''
	form.description = ''
	editingChannel.value = null
}
const toISO = (value: string) => {
	if (!value) return undefined
	const date = new Date(value)
	return Number.isNaN(date.getTime()) ? undefined : date.toISOString()
}
const handleEdit = (entry: SpuChannelEntry) => {
	editingChannel.value = entry.channel
	form.channel = entry.channel
	form.availability = entry.availability
	form.publishAt = entry.publishAt ? entry.publishAt.substring(0, 16) : ''
	form.withdrawAt = entry.withdrawAt ? entry.withdrawAt.substring(0, 16) : ''
	form.title = entry.contentOverride?.title || ''
	form.description = entry.contentOverride?.description || ''
}
const handleSave = async () => {
	if (!props.spuId) return
	try {
		saving.value = true
		const payload: SpuChannelPayload = {
			channel: form.channel,
			availability: form.availability,
			publishAt: toISO(form.publishAt) || null,
			withdrawAt: toISO(form.withdrawAt) || null,
			title: form.title || undefined,
			description: form.description || undefined,
		}
		await api.upsertSpuChannel(props.spuId, payload)
		toast.add({ title: '渠道已更新', description: `${payload.channel} 已保存` })
		await loadChannels()
		resetForm()
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '保存失败', color: 'red' })
	} finally {
		saving.value = false
	}
}
const handleRemove = async (entry: SpuChannelEntry) => {
	if (!props.spuId) return
	try {
		removing.value = entry.channel
		await api.deleteSpuChannel(props.spuId, entry.channel)
		toast.add({ title: '渠道已删除', description: entry.channel })
		await loadChannels()
		if (editingChannel.value === entry.channel) {
			resetForm()
		}
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '删除失败', color: 'red' })
	} finally {
		removing.value = null
	}
}
const loadChannels = async () => {
	if (!props.spuId) return
	try {
		loading.value = true
		const { items } = await api.listSpuChannels(props.spuId)
		channels.value = items ?? []
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '加载渠道失败', color: 'red' })
	} finally {
		loading.value = false
	}
}

watch(
	() => props.spuId,
	(id) => {
		if (id) {
			loadChannels()
			resetForm()
		}
	},
	{ immediate: true }
)

watch(
	() => channels.value.length,
	() => {
		channelPagination.page = 1
	},
	{ flush: 'post' }
)

watch(
	() => channelPagination.pageSize,
	() => {
		channelPagination.page = 1
	}
)

watch(channelPageCount, (count) => {
	if (channelPagination.page > count) {
		channelPagination.page = count
	}
})
</script>
