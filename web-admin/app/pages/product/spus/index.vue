<template>
	<div class="px-6 py-8 text-white">
		<div class="mx-auto flex max-w-6xl flex-col gap-6">
			<header class="flex flex-wrap items-center justify-between gap-3">
				<div class="flex-1">
					<p class="text-xs uppercase tracking-[0.2em] text-primary-200">商品中心</p>
					<h1 class="mt-1 text-3xl font-semibold text-white">SPU 列表</h1>
					<p class="text-sm text-white/60">管理与审阅所有商品，支持批量导入与渠道同步。</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<UButton icon="i-heroicons-arrow-up-tray" color="primary" variant="soft" @click="importDialogOpen = true">
						导入
					</UButton>
					<UButton icon="i-heroicons-arrow-down-tray" variant="soft" @click="exportDialogOpen = true">
						导出
					</UButton>
					<UButton color="primary" to="/product/spus/create">新建 SPU</UButton>
				</div>
			</header>

			<UCard>
				<div class="space-y-4">
					<div v-if="categoryBadge" class="flex flex-wrap items-center gap-3">
						<UBadge color="primary" variant="soft">{{ categoryBadge }}</UBadge>
						<UButton size="xs" variant="ghost" @click="clearCategoryFilter">清除类目筛选</UButton>
					</div>
					<div class="grid grid-cols-12 gap-4">
						<UFormField label="关键字" :ui="inlineFieldUi" class="col-span-12 md:col-span-4">
							<UInput
								v-model="filters.keyword"
								icon="i-heroicons-magnifying-glass-20-solid"
								placeholder="编码或名称"
								clearable
								@keyup.enter="applyFilters"
							/>
						</UFormField>
						<UFormField label="状态" :ui="inlineFieldUi" class="col-span-12 sm:col-span-6 md:col-span-3">
							<USelect v-model="filters.status" :items="statusItems" class="w-full" />
						</UFormField>
						<UFormField label="类型" :ui="inlineFieldUi" class="col-span-12 sm:col-span-6 md:col-span-3">
							<USelect v-model="filters.type" :items="typeItems" class="w-full" />
						</UFormField>
						<UFormField label="渠道" :ui="inlineFieldUi" class="col-span-12 sm:col-span-6 md:col-span-2">
							<USelect v-model="filters.channel" :items="channelItems" class="w-full" />
						</UFormField>
					</div>
					<div class="flex flex-wrap justify-end gap-3">
						<UButton variant="ghost" @click="resetFilters">重置</UButton>
						<UButton color="primary" @click="applyFilters">应用过滤</UButton>
					</div>
				</div>
			</UCard>

			<UCard>
				<UTable
					:data="items"
					:columns="columns"
					:loading="loading"
					empty-text="暂无 SPU，可通过右上角按钮创建。"
					class="overflow-x-auto"
				>
					<template #type-cell="{ row }">
						<UBadge variant="soft">{{ typeLabel(row.original.type) }}</UBadge>
					</template>
					<template #status-cell="{ row }">
						<UBadge :color="statusColor(row.original.status)">
							{{ statusLabel(row.original.status) }}
						</UBadge>
					</template>
					<template #updatedAt-cell="{ row }">
						<span class="text-sm text-gray-400">{{ formatTime(row.original.updatedAt) }}</span>
					</template>
					<template #actions-cell="{ row }">
						<div class="flex flex-wrap gap-2">
							<UButton size="xs" variant="soft" :to="`/product/spus/edit/${row.original.id}`">编辑</UButton>
							<UButton
								v-if="row.original.status === 'draft'"
								size="xs"
								color="primary"
								:loading="submitLoading === row.original.id"
								@click="handleSubmit(row.original)"
							>
								提交
							</UButton>
							<UButton
								v-if="row.original.status === 'reviewing'"
								size="xs"
								color="emerald"
								:loading="publishLoading === row.original.id"
								@click="handlePublish(row.original)"
							>
								发布
							</UButton>
							<UButton
								v-if="row.original.status === 'published'"
								size="xs"
								color="warning"
								variant="soft"
								@click="openListWithdraw(row.original)"
							>
								下架
							</UButton>
							<UButton
								v-if="canDeleteStatus(row.original.status)"
								size="xs"
								color="error"
								variant="soft"
								@click="openListDelete(row.original)"
							>
								删除
							</UButton>
						</div>
					</template>
				</UTable>
				<div class="mt-4 flex flex-wrap items-center justify-between gap-4">
					<div class="text-sm text-gray-400">
						共 {{ total }} 条记录，当前第 {{ pagination.page }} / {{ pageCount }} 页
					</div>
					<div class="flex items-center gap-3">
						<USelect v-model="pagination.pageSize" :items="pageSizeItems" class="w-32" />
						<UPagination v-model="pagination.page" :page-count="pageCount" />
					</div>
				</div>
			</UCard>
		</div>

		<SpuBulkDialog v-model="importDialogOpen" mode="import" @submitted="handleBulkSubmitted" />
		<SpuBulkDialog v-model="exportDialogOpen" mode="export" @submitted="handleBulkSubmitted" />

		<UModal
			v-model:open="listWithdraw.open"
			title="下架 SPU"
			:description="listWithdraw.target?.name || '未选择'"
			:close="{ onClick: closeListWithdraw }"
			:prevent-close="listWithdraw.loading"
			:ui="modalUi"
		>
				<template #body>
					<UFormField label="渠道（默认全部）">
						<USelectMenu
							v-model="listWithdraw.form.channels"
							:items="listWithdraw.items"
							:portal="false"
							multiple
							placeholder="全部渠道"
						/>
					</UFormField>
				<UFormField label="下架时间" help="不填写则立即下架">
					<UInput v-model="listWithdraw.form.withdrawAt" type="datetime-local" />
				</UFormField>
				<UFormField label="下架原因">
					<UTextarea v-model="listWithdraw.form.reason" :rows="3" placeholder="必填，下架原因" />
				</UFormField>
			</template>
			<template #footer>
				<UButton color="neutral" variant="subtle" @click="closeListWithdraw">取消</UButton>
				<UButton color="warning" :loading="listWithdraw.loading" @click="confirmListWithdraw">确认下架</UButton>
			</template>
		</UModal>

		<UModal
			v-model:open="listDelete.open"
			title="删除 SPU"
			:description="`${listDelete.target?.name || '未选择'} · 仅草稿/下架可删除`"
			:close="{ onClick: closeListDelete }"
			:prevent-close="listDelete.loading"
			:ui="modalUi"
		>
			<template #body>
				<UFormField label="删除原因">
					<UTextarea v-model="listDelete.reason" :rows="4" placeholder="必填，说明删除原因" />
				</UFormField>
			</template>
			<template #footer>
				<UButton color="neutral" variant="subtle" @click="closeListDelete">取消</UButton>
				<UButton color="error" :loading="listDelete.loading" @click="confirmListDelete">确认删除</UButton>
			</template>
		</UModal>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter, useToast } from '#imports'
import type { TableColumn } from '@nuxt/ui'
import SpuBulkDialog from '~/components/product/SpuBulkDialog.vue'
import type { SpuSummary } from '~/composables/api/useSpu'
import { useSpuApi } from '~/composables/api/useSpu'
import { useSpuStore } from '~/stores/product/spu'
import { channelSelectOptions } from '~/data/channelCatalog'

const toast = useToast()
const route = useRoute()
const router = useRouter()
const store = useSpuStore()
const api = useSpuApi()
const { items, loading, total } = storeToRefs(store)
const filters = reactive<{
	keyword: string
	status: string | null
	type: string | null
	channel: string | null
	categoryId: string
	categoryPathPrefix: string
	categoryName: string
}>({
	keyword: '',
	status: null,
	type: null,
	channel: null,
	categoryId: '',
	categoryPathPrefix: '',
	categoryName: '',
})
const pagination = reactive({ page: 1, pageSize: 10 })
const modalUi = {
	content: 'max-w-lg w-full',
	body: 'space-y-4 p-4 sm:p-5',
	header: 'px-4 sm:px-5 pt-4 sm:pt-5 pb-0',
	footer: 'px-4 sm:px-5 pb-4 sm:pb-5 pt-0 flex justify-end gap-2',
}
const submitLoading = ref<string | null>(null)
const publishLoading = ref<string | null>(null)
const importDialogOpen = ref(false)
const exportDialogOpen = ref(false)
const listWithdraw = reactive({
	open: false,
	loading: false,
	target: null as SpuSummary | null,
	items: [] as { label: string; value: string }[],
	form: {
		channels: [] as string[],
		withdrawAt: '',
		reason: '',
	},
})
const listDelete = reactive({
	open: false,
	loading: false,
	target: null as SpuSummary | null,
	reason: '',
})

const statusItems = [
	{ label: '全部', value: null },
	{ label: '草稿', value: 'draft' },
	{ label: '审核中', value: 'reviewing' },
	{ label: '已发布', value: 'published' },
]
const typeItems = [
	{ label: '全部类型', value: null },
	{ label: '一次性', value: 'one_time' },
	{ label: '订阅', value: 'subscription' },
	{ label: '组合', value: 'bundle' },
]
const channelItems = [
	{ label: '全部渠道', value: null },
	...channelSelectOptions.map((item) => ({ label: item.label, value: item.value })),
]
const pageSizeItems = [
	{ label: '10 / 页', value: 10 },
	{ label: '20 / 页', value: 20 },
	{ label: '50 / 页', value: 50 },
]
const inlineFieldUi = {
	root: 'flex flex-col gap-2',
	label: 'text-sm font-medium text-gray-400',
	container: 'mt-0 w-full',
}
const columns = computed<TableColumn<SpuSummary>[]>(() => [
	{ accessorKey: 'code', header: '编码' },
	{ accessorKey: 'name', header: '名称' },
	{ accessorKey: 'type', header: '类型' },
	{ accessorKey: 'status', header: '状态' },
	{ accessorKey: 'updatedAt', header: '更新时间' },
	{ id: 'actions', header: '操作' },
])
const pageCount = computed(() => Math.max(1, Math.ceil(total.value / pagination.pageSize)))
const categoryBadge = computed(() => {
	if (filters.categoryName) return `类目：${filters.categoryName}`
	if (filters.categoryId) return `类目 ID：${filters.categoryId}`
	if (filters.categoryPathPrefix) return `类目 Path：${filters.categoryPathPrefix}`
	return ''
})

const normalizeQueryString = (value: unknown) => {
	if (Array.isArray(value)) return typeof value[0] === 'string' ? value[0] : ''
	return typeof value === 'string' ? value : ''
}

const typeLabel = (value: string) => {
	const map: Record<string, string> = { one_time: '一次性', subscription: '订阅', bundle: '组合' }
	return map[value] || value || '—'
}
const statusLabel = (value: string) => {
	const map: Record<string, string> = { draft: '草稿', reviewing: '审核中', published: '已发布' }
	return map[value] || value
}
const statusColor = (value: string) => {
	switch (value) {
		case 'published':
			return 'emerald'
		case 'reviewing':
			return 'primary'
		default:
			return 'gray'
	}
}
const formatTime = (value?: string) => (value ? new Date(value).toLocaleString() : '—')
const canDeleteStatus = (status: string) => ['draft', 'offboarded'].includes(status)

const fetchList = () =>
	store.fetchList({
		keyword: filters.keyword || undefined,
		status: filters.status ?? undefined,
		type: filters.type ?? undefined,
		channel: filters.channel ?? undefined,
		categoryId: filters.categoryId || undefined,
		categoryPathPrefix: filters.categoryPathPrefix || undefined,
		page: pagination.page,
		pageSize: pagination.pageSize,
	})

const applyFilters = () => {
	pagination.page = 1
	fetchList()
}
const resetFilters = () => {
	filters.keyword = ''
	filters.status = null
	filters.type = null
	filters.channel = null
	applyFilters()
}

const clearCategoryFilter = async () => {
	filters.categoryId = ''
	filters.categoryPathPrefix = ''
	filters.categoryName = ''
	await router.replace({
		query: {
			...route.query,
			categoryId: undefined,
			categoryPathPrefix: undefined,
			categoryName: undefined,
		},
	})
	applyFilters()
}
watch(
	() => pagination.page,
	() => fetchList()
)
watch(
	() => pagination.pageSize,
	() => {
		pagination.page = 1
		fetchList()
	}
)

const handleSubmit = async (row: SpuSummary) => {
	try {
		submitLoading.value = row.id
		await store.submit(row.id)
		toast.add({ title: '提交成功', description: `${row.name} 已进入审批流程` })
		fetchList()
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '提交失败', color: 'red' })
	} finally {
		submitLoading.value = null
	}
}

const handlePublish = async (row: SpuSummary) => {
	try {
		publishLoading.value = row.id
		const detail = await store.fetchDetail(row.id)
		const versionId = detail?.currentVersionId
		if (!versionId) {
			throw new Error('缺少版本信息')
		}
		await store.publish(row.id, { versionId, channels: ['official'] })
		toast.add({ title: '发布请求已提交', description: `${row.name} 将同步渠道` })
		fetchList()
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '发布失败', color: 'red' })
	} finally {
		publishLoading.value = null
	}
}

const blurActiveElement = () => {
	if (typeof document === 'undefined') return
	const active = document.activeElement as HTMLElement | null
	active?.blur()
}

const closeListWithdraw = () => {
	blurActiveElement()
	listWithdraw.open = false
}

const closeListDelete = () => {
	blurActiveElement()
	listDelete.open = false
}

const openListWithdraw = async (row: SpuSummary) => {
	try {
		listWithdraw.loading = true
		listWithdraw.target = row
		const { items } = await api.listSpuChannels(row.id)
		listWithdraw.items = (items ?? []).map((item) => ({ label: item.channel, value: item.channel }))
		listWithdraw.form.channels = []
		listWithdraw.form.withdrawAt = ''
		listWithdraw.form.reason = ''
		listWithdraw.open = true
	} catch (error) {
		console.error(error)
		toast.add({ title: '加载渠道失败', color: 'red' })
	} finally {
		listWithdraw.loading = false
	}
}

const confirmListWithdraw = async () => {
	if (!listWithdraw.target) return
	if (!listWithdraw.form.reason.trim()) {
		toast.add({ title: '请填写下架原因', color: 'orange' })
		return
	}
	try {
		listWithdraw.loading = true
		await store.withdraw(listWithdraw.target.id, {
			channels: listWithdraw.form.channels.length ? listWithdraw.form.channels : undefined,
			withdrawAt: listWithdraw.form.withdrawAt ? new Date(listWithdraw.form.withdrawAt).toISOString() : undefined,
			reason: listWithdraw.form.reason.trim(),
		})
		toast.add({ title: '已提交下架', description: `${listWithdraw.target.name} 将下架`, color: 'warning' })
		closeListWithdraw()
		fetchList()
	} catch (error) {
		console.error(error)
		toast.add({ title: '下架失败', color: 'red' })
	} finally {
		listWithdraw.loading = false
	}
}

const openListDelete = (row: SpuSummary) => {
	if (!canDeleteStatus(row.status)) {
		toast.add({ title: '当前状态不可删除', color: 'orange' })
		return
	}
	listDelete.target = row
	listDelete.reason = ''
	listDelete.open = true
}

const confirmListDelete = async () => {
	if (!listDelete.target) return
	if (!listDelete.reason.trim()) {
		toast.add({ title: '请填写删除原因', color: 'orange' })
		return
	}
	try {
		listDelete.loading = true
		await store.delete(listDelete.target.id, { reason: listDelete.reason.trim() })
		toast.add({ title: '已删除 SPU', description: `${listDelete.target.name} 已移除`, color: 'success' })
		closeListDelete()
		fetchList()
	} catch (error) {
		console.error(error)
		toast.add({ title: '删除失败', color: 'red' })
	} finally {
		listDelete.loading = false
	}
}

const handleBulkSubmitted = ({ taskId, mode }: { taskId: string; mode: 'import' | 'export' }) => {
	const message = mode === 'import' ? '导入任务已排队' : '导出任务已排队'
	toast.add({ title: message, description: `任务编号：${taskId}` })
}

onMounted(() => {
	filters.categoryId = normalizeQueryString(route.query.categoryId)
	filters.categoryPathPrefix = normalizeQueryString(route.query.categoryPathPrefix)
	filters.categoryName = normalizeQueryString(route.query.categoryName)
	fetchList()
})
</script>
