<template>
	<div class="px-6 py-8 text-white">
		<div class="mx-auto flex max-w-6xl flex-col gap-6">
			<header class="flex flex-wrap items-center justify-between gap-3">
				<div class="flex-1">
					<p class="text-xs uppercase tracking-[0.2em] text-primary-200">商品中心</p>
					<h1 class="mt-1 text-3xl font-semibold text-white">类目模板</h1>
					<p class="text-sm text-white/60">配置模板字段并发布版本，用于 SPU 录入校验。</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<UButton icon="i-heroicons-arrow-path" variant="soft" :loading="loading" @click="fetchList">刷新</UButton>
					<UButton icon="i-heroicons-plus" color="primary" :loading="creating" @click="createTemplate">新建模板</UButton>
				</div>
			</header>

			<UCard>
				<div class="grid grid-cols-12 gap-4">
					<UFormField label="关键字" class="col-span-12 md:col-span-6">
						<UInput
							v-model="filters.keyword"
							icon="i-heroicons-magnifying-glass-20-solid"
							placeholder="按名称搜索"
							clearable
							@keyup.enter="fetchList"
						/>
					</UFormField>
					<UFormField label="状态" class="col-span-12 md:col-span-3">
						<USelect v-model="filters.status" :items="statusItems" class="w-full" />
					</UFormField>
					<div class="col-span-12 md:col-span-3 flex items-end justify-end gap-2">
						<UButton variant="ghost" @click="resetFilters">重置</UButton>
						<UButton color="primary" @click="fetchList">查询</UButton>
					</div>
				</div>
			</UCard>

			<UCard>
				<UTable :data="items" :columns="columns" :loading="loading" empty-text="暂无模板，可点击右上角新建。">
					<template #status-cell="{ row }">
						<UBadge :color="row.original.status === 'enabled' ? 'emerald' : 'gray'" variant="soft">
							{{ row.original.status }}
						</UBadge>
					</template>
					<template #published-cell="{ row }">
						<span class="text-sm text-white/70">v{{ row.original.currentPublishedVersionNo || 0 }}</span>
					</template>
					<template #actions-cell="{ row }">
						<div class="flex flex-wrap gap-2">
							<UButton size="xs" variant="soft" :to="`/product/category-templates/${row.original.id}`">编辑</UButton>
						</div>
					</template>
				</UTable>
				<div class="mt-4 flex flex-wrap items-center justify-between gap-4">
					<div class="text-sm text-white/60">共 {{ total }} 条记录</div>
					<div class="flex items-center gap-3">
						<USelect v-model="filters.pageSize" :items="pageSizeItems" class="w-28" />
						<UPagination v-model="filters.page" :page-count="pageCount" />
					</div>
				</div>
			</UCard>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useToastAlert } from '~/composables/useToastAlert'
import { useCategoryTemplateApi, type CategoryTemplate } from '~/composables/api/useCategoryTemplate'

const api = useCategoryTemplateApi()
const toast = useToastAlert()

const loading = ref(false)
const creating = ref(false)
const items = ref<CategoryTemplate[]>([])
const total = ref(0)

const STATUS_ALL = '__all__'

const statusItems = [
	{ label: '全部', value: STATUS_ALL },
	{ label: '启用', value: 'enabled' },
	{ label: '停用', value: 'disabled' },
]
const pageSizeItems = [10, 20, 50, 100].map((n) => ({ label: String(n), value: n }))

const filters = reactive({
	keyword: '',
	status: STATUS_ALL,
	page: 1,
	pageSize: 20,
})

const pageCount = computed(() => {
	const size = filters.pageSize || 20
	return Math.max(1, Math.ceil((total.value || 0) / size))
})

watch(
	() => [filters.page, filters.pageSize],
	() => fetchList(),
)

const columns = [
	{ id: 'name', accessorKey: 'name', header: '名称' },
	{ id: 'status', accessorKey: 'status', header: '状态' },
	{ id: 'published', header: '已发布版本' },
	{ id: 'updatedAt', accessorKey: 'updatedAt', header: '更新时间' },
	{ id: 'actions', header: '操作' },
]

const fetchList = async () => {
	loading.value = true
	try {
		const resp = await api.list({
			keyword: filters.keyword || undefined,
			status: filters.status === STATUS_ALL ? undefined : filters.status || undefined,
			page: filters.page,
			pageSize: filters.pageSize,
		})
		items.value = resp.items ?? []
		total.value = resp.meta?.total ?? items.value.length
	} catch (error: any) {
		toast.add({ title: '加载失败', description: error?.message || '无法获取模板列表', color: 'error' })
	} finally {
		loading.value = false
	}
}

const resetFilters = () => {
	filters.keyword = ''
	filters.status = STATUS_ALL
	filters.page = 1
	filters.pageSize = 20
	fetchList()
}

const createTemplate = async () => {
	creating.value = true
	try {
		const created = await api.create({
			name: `新模板 ${new Date().toISOString().slice(0, 10)}`,
			status: 'enabled',
			fields: [],
		})
		toast.add({ title: '已创建模板', description: created.name, color: 'success' })
		await navigateTo(`/product/category-templates/${created.id}`)
	} catch (error: any) {
		toast.add({ title: '创建失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		creating.value = false
	}
}

onMounted(() => fetchList())
</script>
