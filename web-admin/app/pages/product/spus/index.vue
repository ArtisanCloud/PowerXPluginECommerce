<template>
	<div class="px-6 py-8 text-white">
		<div class="mx-auto flex max-w-6xl flex-col gap-6">
			<header class="flex flex-wrap items-center gap-4">
				<div class="flex-1">
					<p class="text-xs uppercase tracking-[0.2em] text-primary-200">商品中心</p>
					<h1 class="mt-1 text-3xl font-semibold text-white">SPU 列表</h1>
					<p class="text-sm text-white/60">创建、查看并管理所有商品。</p>
				</div>
				<UButton color="primary" size="md" to="/product/spus/create">新建 SPU</UButton>
			</header>

			<UCard class="border border-white/10 bg-white/5 shadow-lg ring-1 ring-white/5 backdrop-blur">
				<div class="grid gap-4 md:grid-cols-[minmax(0,2fr)_minmax(0,1fr)_auto]">
					<UInput
						v-model="filters.keyword"
						placeholder="搜索编码/名称"
						class="w-full"
						icon="i-heroicons-magnifying-glass-20-solid"
						@keyup.enter="fetch"
					/>
					<USelect v-model="filters.status" :options="statusOptions" placeholder="状态" class="w-full" />
					<div class="flex items-center justify-end">
						<UButton color="gray" variant="soft" @click="fetch">过滤</UButton>
					</div>
				</div>
			</UCard>

			<UCard class="border border-white/10 bg-white/5 shadow-lg ring-1 ring-white/5 backdrop-blur">
				<UTable :rows="store.items" :columns="columns" :loading="store.loading">
					<template #actions-data="{ row }">
						<div class="flex flex-wrap gap-2">
							<UButton color="gray" variant="soft" size="xs" :to="`/product/spus/${row.id}`">查看</UButton>
							<UButton
								v-if="row.status === 'draft'"
								size="xs"
								color="primary"
								:loading="submitLoading === row.id"
								@click="handleSubmit(row)"
							>
								提交审核
							</UButton>
							<UButton
								v-if="row.status === 'reviewing'"
								size="xs"
								color="emerald"
								:loading="publishLoading === row.id"
								@click="handlePublish(row)"
							>
								发布
							</UButton>
						</div>
					</template>
				</UTable>
			</UCard>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useToast } from '#imports'
import type { SpuSummary } from '~/composables/api/useSpu'
import { useSpuStore } from '~/stores/product/spu'

const toast = useToast()
const store = useSpuStore()
const filters = reactive({ keyword: '', status: '' })
const statusOptions = [
	{ label: '全部', value: '' },
	{ label: '草稿', value: 'draft' },
	{ label: '审核中', value: 'reviewing' },
	{ label: '已发布', value: 'published' },
]

const columns = [
	{ id: 'code', key: 'code', label: '编码' },
	{ id: 'name', key: 'name', label: '名称' },
	{ id: 'type', key: 'type', label: '类型' },
	{ id: 'status', key: 'status', label: '状态' },
	{ id: 'updatedAt', key: 'updatedAt', label: '更新时间' },
	{ id: 'actions', key: 'actions', label: '操作' },
]

const submitLoading = ref<string | null>(null)
const publishLoading = ref<string | null>(null)

const handleSubmit = async (row: SpuSummary) => {
	try {
		submitLoading.value = row.id
		await store.submit(row.id)
		toast.add({ title: '提交成功', description: `${row.name} 已进入审批流程` })
		fetch()
	} catch (error) {
		console.error(error)
		toast.add({ title: '提交失败', description: '稍后再试', color: 'red' })
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
		toast.add({ title: '发布请求已提交', description: `${row.name} 将同步到渠道` })
		fetch()
	} catch (error) {
		console.error(error)
		toast.add({ title: '发布失败', description: '请确认草稿已提交并稍后重试', color: 'red' })
	} finally {
		publishLoading.value = null
	}
}

const fetch = () => store.fetchList(filters)

onMounted(() => fetch())
</script>
