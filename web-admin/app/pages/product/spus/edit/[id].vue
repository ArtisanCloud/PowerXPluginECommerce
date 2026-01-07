<template>
	<div class="space-y-6 px-6 py-8">
		<div class="flex flex-wrap items-center justify-between gap-4">
			<NuxtLink to="/product/spus" class="text-sm text-primary">← 返回列表</NuxtLink>
			<div class="flex flex-wrap gap-2">
				<UButton variant="outline" color="primary" @click="openChannelModal" :disabled="!spu?.id">
					渠道可见性
				</UButton>
				<UButton
					variant="outline"
					color="primary"
					@click="openSubscriptionModal"
					:disabled="!spu?.id || (spu?.type || '').toLowerCase() !== 'subscription'"
				>
					订阅计划
				</UButton>
				<UButton variant="outline" color="primary" @click="openSkuModal" :disabled="!spu?.id">
					关联 SKU
				</UButton>
				<UButton variant="soft" color="primary" @click="openVersionModal">版本时间线</UButton>
				<UButton variant="soft" color="secondary" @click="openApprovalModal">审批记录</UButton>
				<UButton v-if="spu?.status === 'draft'" variant="soft" color="primary" @click="openEditModal">
					编辑基础信息
				</UButton>
				<UButton
					v-if="spu && spu.status !== 'draft'"
					variant="soft"
					color="primary"
					:loading="reviseLoading"
					@click="openReviseModal"
				>
					创建草稿版本
				</UButton>
				<UButton v-if="spu?.status === 'draft'" color="primary" :loading="submitLoading" @click="submitDraft">
					提交审核
				</UButton>
				<UButton v-if="spu?.status === 'reviewing'" color="emerald" :loading="publishLoading" @click="publishDraft">
					发布
				</UButton>
				<UButton v-if="spu?.status === 'published'" color="warning" variant="soft" :loading="withdrawLoading" @click="openWithdrawModal">
					下架
				</UButton>
				<UButton
					v-if="canDelete"
					color="error"
					variant="soft"
					:loading="deleteLoading"
					@click="openDeleteModal"
				>
					删除
				</UButton>
			</div>
		</div>

		<div v-if="loading">
			<USkeleton class="h-24" />
		</div>
		<div v-else-if="spu" class="space-y-6">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ spu.name }}</h1>
				<p class="text-sm text-gray-500 dark:text-gray-400">当前状态：<UBadge>{{ spu.status }}</UBadge></p>
			</div>
			<UCard>
				<dl class="grid gap-4 md:grid-cols-2">
					<div>
						<dt class="text-sm text-gray-500 dark:text-gray-400">编码</dt>
						<dd class="text-base font-medium text-gray-900 dark:text-white">{{ spu.code }}</dd>
					</div>
					<div>
						<dt class="text-sm text-gray-500 dark:text-gray-400">类目</dt>
						<dd class="text-base font-medium text-gray-900 dark:text-white">
							<NuxtLink to="/product/categories" class="text-primary hover:underline">
								{{ categoryPathDisplay || spu.categoryPath || '—' }}
							</NuxtLink>
						</dd>
					</div>
					<div>
						<dt class="text-sm text-gray-500 dark:text-gray-400">类型</dt>
						<dd class="text-base font-medium text-gray-900 dark:text-white">{{ typeLabel(spu.type) }}</dd>
					</div>
					<div>
						<dt class="text-sm text-gray-500 dark:text-gray-400">负责人</dt>
						<dd class="text-base font-medium text-gray-900 dark:text-white">{{ spu.responsibleUser || '—' }}</dd>
					</div>
					<div>
						<dt class="text-sm text-gray-500 dark:text-gray-400">默认语言</dt>
						<dd class="text-base font-medium text-gray-900 dark:text-white">{{ spu.defaultLocale }}</dd>
					</div>
					<div class="md:col-span-2">
						<dt class="text-sm text-gray-500 dark:text-gray-400">标签</dt>
						<dd class="mt-1 flex flex-wrap gap-2">
							<UBadge v-for="tag in (spu.tags || [])" :key="tag" variant="soft" color="neutral">{{ tag }}</UBadge>
							<span v-if="!(spu.tags || []).length" class="text-base font-medium text-gray-900 dark:text-white">—</span>
						</dd>
					</div>
				</dl>
			</UCard>
			<UCard>
				<template #header>
					<div class="flex flex-wrap items-center justify-between gap-4">
						<div>
							<h3 class="text-lg font-semibold text-gray-900 dark:text-white">关联 SKU 摘要</h3>
							<p class="text-sm text-gray-500 dark:text-gray-400">最近关联/生成的 SKU 概览，保存后自动刷新。</p>
						</div>
						<div class="flex items-center gap-3">
							<UBadge variant="soft" color="primary">共 {{ totalSkuCount }} 条</UBadge>
							<UButton size="sm" variant="outline" color="primary" @click="openSkuModal">管理关联</UButton>
						</div>
					</div>
				</template>
				<div v-if="skuSummaryLoading">
					<USkeleton class="h-28" />
				</div>
				<div v-else-if="skuSummaryPreview.length" class="space-y-3">
					<div class="-mx-2 overflow-x-auto sm:mx-0">
						<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm">
							<thead class="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-500">
								<tr>
									<th scope="col" class="px-3 py-2 text-left font-medium">SKU 编码</th>
									<th scope="col" class="px-3 py-2 text-left font-medium">名称</th>
									<th scope="col" class="px-3 py-2 text-left font-medium">价格</th>
									<th scope="col" class="px-3 py-2 text-left font-medium">库存引用</th>
									<th scope="col" class="px-3 py-2 text-left font-medium">属性</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-900/30">
								<tr v-for="sku in skuSummaryPreview" :key="sku.id || sku.code">
									<td class="px-3 py-2 font-medium text-gray-900 dark:text-white">
										<NuxtLink
											v-if="sku.id"
											:to="`/product/skus/${sku.id}`"
											class="text-primary hover:underline"
										>
											{{ sku.code || sku.id }}
										</NuxtLink>
										<span v-else>{{ sku.code || '—' }}</span>
									</td>
									<td class="px-3 py-2 text-gray-600 dark:text-gray-300">
										<div class="line-clamp-2">
											{{ sku.name || '—' }}
										</div>
									</td>
									<td class="px-3 py-2 font-mono text-sm text-gray-900 dark:text-gray-100">
										{{ formatSkuPrice(sku.pricing) }}
									</td>
									<td class="px-3 py-2 text-gray-600 dark:text-gray-300">
										{{ sku.inventoryRef || '—' }}
									</td>
									<td class="px-3 py-2">
										<div v-if="extractAttributePairs(sku.attributes).length" class="flex flex-wrap gap-1">
											<UBadge
												v-for="(pair, idx) in extractAttributePairs(sku.attributes)"
												:key="`${sku.code}-attr-${idx}`"
												variant="soft"
												color="neutral"
											>
												{{ pair[0] }}：{{ pair[1] }}
											</UBadge>
										</div>
										<span v-else class="text-xs text-gray-400">—</span>
									</td>
								</tr>
							</tbody>
						</table>
					</div>
					<p class="text-xs text-gray-500 dark:text-gray-400">
						展示最近 {{ skuSummaryPreview.length }} 条，点击“管理关联”可查看全部。
					</p>
				</div>
				<div v-else class="rounded border border-dashed border-gray-200 dark:border-gray-700 p-4 text-sm text-gray-500 dark:text-gray-400">
					暂无关联 SKU，点击“关联 SKU”可批量生成或手动新增。
				</div>
			</UCard>
			<!-- 业务配置通过按钮触发弹框 -->
		</div>
		<div v-else>
			<p class="text-gray-500 dark:text-gray-400">未找到 SPU</p>
		</div>
	</div>
	<UModal
		v-model:open="versionModalOpen"
		title="版本时间线与字段差异"
		description="查看版本历史、审批链及字段差异"
		:close="{ onClick: closeVersionModal }"
		:ui="modalFullUi"
	>
		<template #body>
			<VersionDiff
				:versions="versions"
				:selected-id="selectedVersionId || undefined"
				:detail="versionDetail"
				:loading="versionDetailLoading"
				@select="handleSelectVersion"
				@rollback="handleRollbackVersion"
			/>
		</template>
	</UModal>
	<UModal
		v-model:open="approvalModalOpen"
		title="审批面板与审计时间轴"
		description="查看审批链并提交意见"
		:close="{ onClick: closeApprovalModal }"
		:ui="modalFullUi"
	>
		<template #body>
			<div class="grid gap-4 lg:grid-cols-2">
				<UCard>
					<template #header>
						<div class="flex items-center justify-between">
							<div>
								<h3 class="text-lg font-semibold text-gray-900 dark:text-white">审批面板</h3>
								<p class="text-sm text-gray-500 dark:text-gray-400">查看审批链并提交意见。</p>
							</div>
							<UBadge v-if="versionDetail" variant="soft">版本 #{{ versionDetail.versionNumber }}</UBadge>
						</div>
					</template>
					<div v-if="!versionDetail">
						<p class="text-sm text-gray-500 dark:text-gray-400">请选择版本以查看审批信息。</p>
					</div>
					<div v-else class="space-y-4">
						<ul class="space-y-3">
							<li
								v-for="approval in versionDetail.approvals"
								:key="approval.id"
								class="rounded border border-gray-200 dark:border-gray-700 px-4 py-3 space-y-1.5"
							>
								<div class="flex items-center justify-between">
									<UBadge color="neutral" variant="soft" class="uppercase">{{ (approval.role || 'unknown').toUpperCase() }}</UBadge>
									<UBadge :color="approvalStatusColor(approval.status)" variant="soft">
										{{ approvalStatusLabel(approval.status) }}
									</UBadge>
								</div>
								<p class="text-sm text-gray-500 dark:text-gray-400">
									SLA：{{ formatDate(approval.slaDueAt) }} · 处理：{{ approval.actedBy || '未处理' }}
								</p>
								<p v-if="approval.comment" class="mt-1 text-sm text-gray-600 dark:text-gray-300">备注：{{ approval.comment }}</p>
							</li>
						</ul>
						<UFormField label="审批意见">
							<template #default="{ id }">
								<UTextarea :id="id" v-model="approvalComment" :rows="3" placeholder="填写审批意见（可选）" />
							</template>
						</UFormField>
						<div class="flex gap-2">
							<UButton color="success" :loading="approvalSubmitting" @click="submitApproval('approve')"> 通过 </UButton>
							<UButton color="error" variant="soft" :loading="approvalSubmitting" @click="submitApproval('reject')">
								驳回
							</UButton>
						</div>
					</div>
				</UCard>
				<SpuAuditTimeline :events="auditEvents" />
			</div>
		</template>
	</UModal>
	<UModal
		v-model:open="editModalOpen"
		title="编辑 SPU（草稿）"
		description="仅草稿可编辑，保存后会更新当前草稿版本。"
		:close="{ onClick: closeEditModal }"
		:prevent-close="editSaving"
		:ui="panelModalUi"
	>
		<template #body>
			<UTabs v-model="editActiveTab" :items="editTabs" class="w-full" />
			<div v-if="editActiveTab === 'basic'" class="mt-4">
				<SpuWizardStepBasic v-model="editBasicModel" />
			</div>
			<div v-else class="mt-4">
				<SpuWizardStepLocale v-model="editLocaleModel" />
			</div>
		</template>
		<template #footer>
			<UButton color="neutral" variant="subtle" :disabled="editSaving" @click="closeEditModal">取消</UButton>
			<UButton color="primary" :loading="editSaving" @click="saveEditModal">保存</UButton>
		</template>
	</UModal>
	<UModal
		v-model:open="reviseModalOpen"
		title="创建草稿版本"
		description="将已发布 SPU 切换为草稿以便编辑（已发布版本仍保留在版本历史中）。"
		:close="{ onClick: closeReviseModal }"
		:prevent-close="reviseLoading"
		:ui="modalUi"
	>
		<template #body>
			<UFormField label="原因" help="必填，用于审计记录">
				<UTextarea v-model="reviseReason" :rows="3" placeholder="例如：更新类目与模板字段" />
			</UFormField>
		</template>
		<template #footer>
			<UButton color="neutral" variant="subtle" :disabled="reviseLoading" @click="closeReviseModal">取消</UButton>
			<UButton color="primary" :loading="reviseLoading" @click="confirmRevise">确认创建</UButton>
		</template>
	</UModal>
	<UModal
		v-model:open="channelModalOpen"
		title="渠道可见性配置"
		description="配置各渠道上架计划与展示内容"
		:close="{ onClick: closeChannelModal }"
		:ui="panelModalUi"
	>
		<template #body>
			<ChannelVisibilityForm v-if="spu" :spu-id="spu.id" />
		</template>
	</UModal>
	<UModal
		v-model:open="subscriptionModalOpen"
		title="订阅计划管理"
		description="针对订阅型商品配置计费与生效范围"
		:close="{ onClick: closeSubscriptionModal }"
		:ui="panelModalUi"
	>
		<template #body>
			<SubscriptionPlanPanel v-if="spu" :spu-id="spu.id" :spu-type="spu.type" />
		</template>
	</UModal>
	<UModal
		v-model:open="skuModalOpen"
		title="关联 SKU"
		description="批量创建/复制 SKU 并同步价格库存信息"
		:close="{ onClick: closeSkuModal }"
		:ui="panelModalUi"
	>
		<template #body>
			<SpuSkuLinker v-if="spu" :spu-id="spu.id" :specs="spuSpecs" @saved="handleSkuLinksUpdated" />
		</template>
	</UModal>
	<UModal
		v-model:open="withdrawModalOpen"
		title="下架渠道"
		description="选择需要下架的渠道并填写原因"
		:close="{ onClick: closeWithdrawModal }"
		:prevent-close="withdrawLoading"
		:ui="modalUi"
	>
		<template #body>
			<UFormField label="渠道" help="为空时默认全部渠道">
				<USelectMenu
					v-model="withdrawForm.channels"
					:items="withdrawChannelItems"
					multiple
					placeholder="全部渠道"
				/>
			</UFormField>
			<UFormField label="下架时间" help="不填写则立即下架">
				<UInput v-model="withdrawForm.withdrawAt" type="datetime-local" />
			</UFormField>
			<UFormField label="下架原因">
				<UTextarea v-model="withdrawForm.reason" :rows="3" placeholder="必填，下架原因" />
			</UFormField>
		</template>
		<template #footer>
			<UButton color="neutral" variant="subtle" @click="closeWithdrawModal">取消</UButton>
			<UButton color="warning" :loading="withdrawLoading" @click="confirmWithdraw">确认下架</UButton>
		</template>
	</UModal>
	<UModal
		v-model:open="deleteModalOpen"
		title="删除 SPU"
		description="仅草稿或已下架的 SPU 可删除，操作不可恢复"
		:close="{ onClick: closeDeleteModal }"
		:prevent-close="deleteLoading"
		:ui="modalUi"
	>
		<template #body>
			<UFormField label="删除原因">
				<UTextarea v-model="deleteReason" :rows="4" placeholder="必填，说明删除原因" />
			</UFormField>
		</template>
		<template #footer>
			<UButton color="neutral" variant="subtle" @click="closeDeleteModal">取消</UButton>
			<UButton color="error" :loading="deleteLoading" @click="confirmDelete">确认删除</UButton>
		</template>
	</UModal>
</template>

<script setup lang="ts">
import { useToast } from '#imports'
import { useRoute, useRouter } from 'vue-router'
import type { SpuDetail, SpuSkuLink } from '~/composables/api/useSpu'
import { useSpuApi } from '~/composables/api/useSpu'
import { useSkuApi } from '~/composables/api/useSku'
import { useCategoryApi, type CategoryNode } from '~/composables/api/useCategory'
import SpuWizardStepBasic from '~/components/product/SpuWizardStepBasic.vue'
import SpuWizardStepLocale from '~/components/product/SpuWizardStepLocale.vue'
import SpuSkuLinker from '~/pages/product/spus/components/SpuSkuLinker.vue'
import VersionDiff from '~/pages/product/spus/components/VersionDiff.vue'
import SpuAuditTimeline from '~/components/product/SpuAuditTimeline.vue'
import ChannelVisibilityForm from '~/components/product/ChannelVisibilityForm.vue'
import SubscriptionPlanPanel from '~/components/product/SubscriptionPlanPanel.vue'
import { useSpuStore } from '~/stores/product/spu'

interface SkuSummaryRow {
	id?: string
	code: string
	status?: string
	name?: string
	inventoryRef?: string
	pricing?: { price: number; currency: string }
	attributes?: Record<string, any>
}

const route = useRoute()
const router = useRouter()
const toast = useToast()
const api = useSpuApi()
const skuApi = useSkuApi()
const categoryApi = useCategoryApi()
const store = useSpuStore()
const spu = ref<SpuDetail | null>(null)
const categoryTreeLoaded = ref(false)
const categoryMap = ref<Record<string, CategoryNode>>({})
const loading = computed(() => store.detailLoading)
const skuSummaryLimit = 5
const skuSummaryItems = ref<SkuSummaryRow[]>([])
const skuSummaryPreview = computed(() => skuSummaryItems.value.slice(0, skuSummaryLimit))
const skuSummaryTotal = ref(0)
const totalSkuCount = computed(() => skuSummaryTotal.value)
const skuSummaryLoading = ref(false)
const submitLoading = ref(false)
const publishLoading = ref(false)
const withdrawLoading = ref(false)
const deleteLoading = ref(false)
const versions = computed(() => store.versions)
const versionDetail = computed(() => store.versionDetail)
const versionDetailLoading = computed(() => store.versionLoading)
const selectedVersionId = ref<string | null>(null)
const approvalComment = ref('')
const approvalSubmitting = ref(false)
const reviseModalOpen = ref(false)
const reviseLoading = ref(false)
const reviseReason = ref('')
const editModalOpen = ref(false)
const editSaving = ref(false)
const editActiveTab = ref<'basic' | 'locale'>('basic')
const editTabs = [
	{ label: '基础信息', value: 'basic' },
	{ label: '多语言内容', value: 'locale' },
]
const editBasicModel = ref<Record<string, any>>({})
const editLocaleModel = ref<{ defaultLocale: string; locales: Array<Record<string, any>> }>({
	defaultLocale: 'zh-CN',
	locales: [{ locale: 'zh-CN', title: '', description: '' }],
})
const modalUi = {
	content: 'max-w-lg w-full',
	body: 'space-y-4 p-4 sm:p-5',
	header: 'px-4 sm:px-5 pt-4 sm:pt-5 pb-0',
	footer: 'px-4 sm:px-5 pb-4 sm:pb-5 pt-0 flex justify-end gap-2',
}
const modalFullUi = {
	content: 'w-screen h-screen max-w-none rounded-none sm:w-screen sm:h-screen sm:rounded-none',
	body: 'p-4 sm:p-6 h-full overflow-y-auto',
	header: 'px-4 sm:px-6 pt-4 sm:pt-6 pb-0',
	footer: 'px-4 sm:px-6 pb-4 sm:pb-6 pt-0',
}
const panelModalUi = {
	content: 'max-w-5xl w-full',
	body: 'p-4 sm:p-6 space-y-4',
	header: 'px-4 sm:px-6 pt-4 sm:pt-6 pb-0',
	footer: 'px-4 sm:px-6 pb-4 sm:pb-6 pt-0',
}
const withdrawModalOpen = ref(false)
const withdrawChannelItems = ref<{ label: string; value: string }[]>([])
const withdrawForm = reactive({
	channels: [] as string[],
	withdrawAt: '',
	reason: '',
})
const versionModalOpen = ref(false)
const approvalModalOpen = ref(false)
const channelModalOpen = ref(false)
const subscriptionModalOpen = ref(false)
const skuModalOpen = ref(false)
const deleteModalOpen = ref(false)
const deleteReason = ref('')
const canDelete = computed(() => {
	const status = spu.value?.status ?? ''
	return status === 'draft' || status === 'offboarded'
})
const auditEvents = computed(() => {
	const events =
		versions.value?.map((item) => ({
			id: `version-${item.id}`,
			title: `版本 #${item.versionNumber} · ${item.status}`,
			description: `提交人：${item.submittedBy || '—'}`,
			timestamp: item.submittedAt || item.createdAt,
		})) ?? []
	if (versionDetail.value) {
		versionDetail.value.approvals.forEach((approval) => {
			events.push({
				id: `approval-${approval.id}`,
				title: `${(approval.role || '').toUpperCase()} · ${approval.status}`,
				description: approval.comment || '',
				timestamp: approval.actedAt || approval.slaDueAt,
			})
		})
	}
	return events
})

const spuSpecs = computed(() => {
	const detail = spu.value as any
	const specs = detail?.specs || detail?.attributes?.specs
	if (!Array.isArray(specs)) {
		return []
	}
	return specs
		.map((spec: any, index: number) => ({
			id: spec.id || spec.specId || spec.code || `spec-${index}`,
			name: spec.name || spec.label || spec.specName || `规格 ${index + 1}`,
			values: (spec.values || spec.options || []).map((val: any, idx: number) => ({
				id: val.id || val.valueId || val.code || `value-${idx}`,
				name: val.name || val.label || val.valueName || val.id,
				code: val.code || val.valueCode,
			})),
		}))
		.filter((spec: any) => spec.values.length)
})

const parseCategoryPathIDs = (path?: string) =>
	String(path || '')
		.split('/')
		.map((p) => p.trim())
		.filter(Boolean)

const categoryPathDisplay = computed(() => {
	const current = spu.value
	if (!current?.categoryPath) return ''
	const ids = parseCategoryPathIDs(current.categoryPath)
	if (!ids.length || !Object.keys(categoryMap.value).length) return ''
	const parts = ids
		.map((id) => categoryMap.value[id])
		.filter(Boolean)
		.map((node) => (node.aliasSlug || node.code || node.displayName || node.id).trim())
		.filter(Boolean)
	return parts.length ? parts.join(' / ') : ''
})

const typeLabel = (value?: string) => {
	const key = String(value || '').toLowerCase()
	const map: Record<string, string> = { one_time: '一次性', subscription: '订阅', bundle: '组合' }
	return map[key] || value || '—'
}

type BadgeColor = 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'error' | 'neutral'

const approvalStatusColor = (status?: string): BadgeColor => {
	switch ((status ?? '').toLowerCase()) {
		case 'approved':
			return 'success'
		case 'rejected':
			return 'error'
		case 'pending':
			return 'warning'
		default:
			return 'neutral'
	}
}

const approvalStatusLabel = (status?: string) => {
	switch ((status ?? '').toLowerCase()) {
		case 'approved':
			return '已通过'
		case 'rejected':
			return '已驳回'
		case 'pending':
			return '待审批'
		case 'escalated':
			return '已升级'
		default:
			return status || '未知'
	}
}

const load = async () => {
	const currentId = route.params.id as string
	const detail = await store.fetchDetail(currentId)
	spu.value = detail ?? null
	if (detail?.categoryId) {
		await ensureCategoryMapLoaded()
	}
	selectedVersionId.value = null
	await store.fetchVersions(currentId)
	if (!selectedVersionId.value && versions.value.length) {
		selectedVersionId.value = versions.value[0].id
		await store.fetchVersionDetail(currentId, selectedVersionId.value)
	}
	if (detail?.id) {
		await loadSkuSummary(detail.id)
	}
}

const ensureCategoryMapLoaded = async () => {
	if (categoryTreeLoaded.value) return
	try {
		const resp = await categoryApi.tree()
		const items = resp.items ?? []
		const out: Record<string, CategoryNode> = {}
		const walk = (nodes: CategoryNode[]) => {
			for (const node of nodes) {
				out[node.id] = node
				if (node.children?.length) walk(node.children)
			}
		}
		walk(items)
		categoryMap.value = out
		categoryTreeLoaded.value = true
	} catch (error) {
		console.warn('加载类目树失败', error)
	}
}

const refreshSpuDetail = async () => {
	if (!spu.value) return
	const detail = await store.fetchDetail(spu.value.id)
	if (detail) {
		spu.value = detail
		if (detail.categoryId) {
			await ensureCategoryMapLoaded()
		}
	}
}

async function loadSkuSummary(targetId?: string) {
	const id = targetId || spu.value?.id
	if (!id) return
	skuSummaryLoading.value = true
	const linkedPromise = store.fetchSkus(id).catch((error) => {
		console.warn('加载 SKU payload 失败', error)
		return []
	})
	try {
		const response = await skuApi.listBySpu(id, { pageSize: 50 })
		const linked = await linkedPromise
		const rows = buildSkuSummaryRows(response?.items ?? [], linked ?? [])
		skuSummaryItems.value = rows
		const totalFromServer = response?.pagination?.total
		skuSummaryTotal.value = typeof totalFromServer === 'number' ? totalFromServer : rows.length
	} catch (error) {
		console.error(error)
		toast.add({ title: '加载 SKU 摘要失败', color: 'error' })
	} finally {
		skuSummaryLoading.value = false
	}
}

const submitDraft = async () => {
	if (!spu.value) return
	try {
		submitLoading.value = true
		await store.submit(spu.value.id)
		toast.add({ title: '已提交审批', description: '等待审核人处理' })
		await load()
	} catch (error) {
		console.error(error)
		toast.add({ title: '提交失败', color: 'error' })
	} finally {
		submitLoading.value = false
	}
}

const blurActiveElement = () => {
	if (typeof document === 'undefined') return
	const active = document.activeElement as HTMLElement | null
	active?.blur()
}

const openVersionModal = () => {
	versionModalOpen.value = true
}

const openApprovalModal = () => {
	approvalModalOpen.value = true
}

const openChannelModal = () => {
	if (!spu.value) return
	channelModalOpen.value = true
}

const openSubscriptionModal = () => {
	if (!spu.value) return
	subscriptionModalOpen.value = true
}

const openSkuModal = () => {
	if (!spu.value) return
	skuModalOpen.value = true
}

const openEditModal = () => {
	const current = spu.value
	if (!current) return
	if ((current.status || '').toLowerCase() !== 'draft') {
		toast.add({ title: '仅草稿可编辑', color: 'warning' })
		return
	}
	const payloadInput = (versionDetail.value as any)?.payload?.input ?? {}
	editBasicModel.value = {
		code: payloadInput.code ?? current.code,
		name: payloadInput.name ?? current.name,
		type: payloadInput.type ?? current.type,
		categoryId: payloadInput.categoryId ?? current.categoryId,
		categoryPath: payloadInput.categoryPath ?? current.categoryPath,
		brandId: payloadInput.brandId ?? current.brandId,
		responsibleUser: payloadInput.responsibleUser ?? current.responsibleUser,
		tags: payloadInput.tags ?? current.tags ?? [],
		attributes: payloadInput.attributes ?? {},
	}
	editLocaleModel.value = {
		defaultLocale: payloadInput.defaultLocale ?? current.defaultLocale ?? 'zh-CN',
		locales: payloadInput.locales ?? current.locales ?? [{ locale: 'zh-CN', title: current.name, description: '' }],
	}
	editActiveTab.value = 'basic'
	editModalOpen.value = true
}

const closeEditModal = () => {
	blurActiveElement()
	editModalOpen.value = false
}

const saveEditModal = async () => {
	if (!spu.value) return
	try {
		editSaving.value = true
		const payload = {
			...editBasicModel.value,
			...editLocaleModel.value,
			tags: Array.isArray((editBasicModel.value as any).tags) ? (editBasicModel.value as any).tags : [],
		}
		await store.update(spu.value.id, payload)
		toast.add({ title: '已保存草稿', color: 'success' })
		closeEditModal()
		await load()
	} catch (error) {
		console.error(error)
		toast.add({ title: '保存失败', color: 'error' })
	} finally {
		editSaving.value = false
	}
}

const openReviseModal = () => {
	reviseReason.value = ''
	reviseModalOpen.value = true
}

const closeReviseModal = () => {
	blurActiveElement()
	reviseModalOpen.value = false
}

const confirmRevise = async () => {
	if (!spu.value) return
	if (!reviseReason.value.trim()) {
		toast.add({ title: '请填写原因', color: 'warning' })
		return
	}
	try {
		reviseLoading.value = true
		await store.revise(spu.value.id, { reason: reviseReason.value.trim() })
		toast.add({ title: '已创建草稿版本', description: '现在可以编辑基础信息并重新提交审批', color: 'success' })
		closeReviseModal()
		await load()
		openEditModal()
	} catch (error) {
		console.error(error)
		toast.add({ title: '创建草稿失败', color: 'error' })
	} finally {
		reviseLoading.value = false
	}
}

const closeWithdrawModal = () => {
	blurActiveElement()
	withdrawModalOpen.value = false
}

const closeDeleteModal = () => {
	blurActiveElement()
	deleteModalOpen.value = false
}

const closeVersionModal = () => {
	blurActiveElement()
	versionModalOpen.value = false
}

const closeApprovalModal = () => {
	blurActiveElement()
	approvalModalOpen.value = false
}

const closeChannelModal = () => {
	blurActiveElement()
	channelModalOpen.value = false
}

const closeSubscriptionModal = () => {
	blurActiveElement()
	subscriptionModalOpen.value = false
}

const closeSkuModal = () => {
	blurActiveElement()
	skuModalOpen.value = false
}

const buildSkuSummaryRows = (officialItems: any[], linkedItems: SpuSkuLink[]): SkuSummaryRow[] => {
	const linkedMap = new Map<string, SpuSkuLink>()
	for (const item of linkedItems ?? []) {
		const key = item?.code?.toLowerCase()
		if (key) {
			linkedMap.set(key, item)
		}
	}
	return (officialItems ?? [])
		.map((item) => normalizeSkuSummaryRow(item, linkedMap))
		.filter((row): row is SkuSummaryRow => Boolean(row))
}

const normalizeSkuSummaryRow = (item: any, linkedMap: Map<string, SpuSkuLink>): SkuSummaryRow | null => {
	const code = (item?.skuCode ?? item?.sku_code ?? '').trim()
	if (!code) {
		return null
	}
	const fallback = linkedMap.get(code.toLowerCase())
	return {
		id: item?.id,
		code,
		status: item?.status,
		name: fallback?.name ?? code,
		inventoryRef: fallback?.inventoryRef,
		pricing: fallback?.pricing,
		attributes: fallback?.attributes,
	}
}

const openWithdrawModal = async () => {
	if (!spu.value) return
	try {
		withdrawLoading.value = true
		const { items } = await api.listSpuChannels(spu.value.id)
		withdrawChannelItems.value =
			items?.map((item) => ({ label: item.channel, value: item.channel })) ?? []
		withdrawForm.channels = []
		withdrawForm.withdrawAt = ''
		withdrawForm.reason = ''
		withdrawModalOpen.value = true
	} catch (error) {
		console.error(error)
		toast.add({ title: '加载渠道失败', color: 'error' })
	} finally {
		withdrawLoading.value = false
	}
}

const openDeleteModal = () => {
	deleteReason.value = ''
	deleteModalOpen.value = true
}

const confirmDelete = async () => {
	if (!spu.value) return
	if (!deleteReason.value.trim()) {
		toast.add({ title: '请填写删除原因', color: 'warning' })
		return
	}
	try {
		deleteLoading.value = true
		await store.delete(spu.value.id, { reason: deleteReason.value.trim() })
		toast.add({ title: '已删除 SPU', color: 'success' })
		closeDeleteModal()
		await router.push('/product/spus')
	} catch (error) {
		console.error(error)
		toast.add({ title: '删除失败', color: 'error' })
	} finally {
		deleteLoading.value = false
	}
}

const confirmWithdraw = async () => {
	if (!spu.value) return
	try {
		withdrawLoading.value = true
		const payload = {
			channels: withdrawForm.channels.length ? withdrawForm.channels : undefined,
			withdrawAt: withdrawForm.withdrawAt ? new Date(withdrawForm.withdrawAt).toISOString() : undefined,
			reason: withdrawForm.reason || undefined,
		}
		await store.withdraw(spu.value.id, payload)
		toast.add({ title: '下架任务已提交', color: 'success' })
		closeWithdrawModal()
		await load()
	} catch (error) {
		console.error(error)
		toast.add({ title: '下架失败', color: 'error' })
	} finally {
		withdrawLoading.value = false
	}
}

const publishDraft = async () => {
	if (!spu.value?.currentVersionId) {
		toast.add({ title: '无可发布版本', color: 'error' })
		return
	}
	try {
		publishLoading.value = true
		await store.publish(spu.value.id, { versionId: spu.value.currentVersionId, channels: ['official'] })
		toast.add({ title: '发布请求已提交' })
		await load()
	} catch (error) {
		console.error(error)
		toast.add({ title: '发布失败', color: 'error' })
	} finally {
		publishLoading.value = false
	}
}

onMounted(load)
watch(
	() => route.params.id,
	() => {
		load()
	}
)

const handleSelectVersion = async (versionId: string) => {
	if (!versionId || versionId === selectedVersionId.value) return
	try {
		selectedVersionId.value = versionId
		await store.fetchVersionDetail(route.params.id as string, versionId)
	} catch (error) {
		console.error(error)
		toast.add({ title: '加载版本失败', color: 'error' })
	}
}

const handleRollbackVersion = async (versionId: string) => {
	if (!versionId) return
	try {
		await store.rollbackVersion(route.params.id as string, versionId, { targetVersionId: versionId, reason: 'Rollback from UI' })
		selectedVersionId.value = store.versionDetail?.id ?? versionId
		toast.add({ title: '已创建回滚草稿' })
	} catch (error) {
		console.error(error)
		toast.add({ title: '回滚失败', color: 'error' })
	}
}

const submitApproval = async (action: 'approve' | 'reject') => {
	if (!spu.value || !selectedVersionId.value) return
	try {
		approvalSubmitting.value = true
		const payload = approvalComment.value ? { comment: approvalComment.value } : undefined
		if (action === 'approve') {
			await store.approveVersion(spu.value.id, selectedVersionId.value, payload)
			toast.add({ title: '已通过当前版本', color: 'success' })
		} else {
			await store.rejectVersion(spu.value.id, selectedVersionId.value, payload)
			toast.add({ title: '已驳回当前版本', color: 'warning' })
		}
		approvalComment.value = ''
		await refreshSpuDetail()
	} catch (error) {
		console.error(error)
		toast.add({ title: '操作失败', color: 'error' })
	} finally {
		approvalSubmitting.value = false
	}
}

const formatDate = (value?: string) => {
	if (!value) return '—'
	return new Date(value).toLocaleString()
}

const formatSkuPrice = (pricing?: { price: number; currency: string }) => {
	if (!pricing || typeof pricing.price === 'undefined' || pricing.price === null) {
		return '—'
	}
	const rawAmount = Number(pricing.price)
	if (!Number.isFinite(rawAmount)) {
		return '—'
	}
	const currency = (pricing.currency || 'CNY').toUpperCase()
	try {
		return new Intl.NumberFormat('zh-CN', {
			style: 'currency',
			currency,
		}).format(rawAmount)
	} catch {
		return `${currency} ${rawAmount.toFixed(2)}`
	}
}

const extractAttributePairs = (attrs?: Record<string, any>) => {
	if (!attrs) return []
	return Object.entries(attrs)
		.map(([key, value]) => [key, normalizeAttributeValue(value)] as [string, string])
		.slice(0, 3)
}

const normalizeAttributeValue = (value: any) => {
	if (value === null || typeof value === 'undefined') {
		return ''
	}
	if (typeof value === 'object') {
		try {
			return JSON.stringify(value)
		} catch {
			return '[object]'
		}
	}
	return String(value)
}

const handleSkuLinksUpdated = async () => {
	if (!spu.value?.id) return
	await loadSkuSummary(spu.value.id)
}
</script>
