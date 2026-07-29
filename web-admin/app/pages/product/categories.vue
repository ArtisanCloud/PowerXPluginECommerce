<template>
	<div class="px-6 py-8 text-white">
		<div class="mx-auto flex max-w-6xl flex-col gap-6">
			<header class="flex flex-wrap items-center justify-between gap-3">
				<div class="flex-1">
					<p class="text-xs uppercase tracking-[0.2em] text-primary-200">商品中心</p>
					<h1 class="mt-1 text-3xl font-semibold text-white">类目管理</h1>
					<p class="text-sm text-white/60">维护类目树（创建/编辑/启停/排序/迁移），并用于前台可展示类目树。</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<UButton icon="i-heroicons-arrow-path" variant="soft" :loading="loading" @click="refreshTree">
						刷新
					</UButton>
					<UButton icon="i-heroicons-plus" color="primary" @click="openCreateRoot">新建根类目</UButton>
				</div>
			</header>

			<div class="grid grid-cols-12 gap-6">
				<UCard class="col-span-12 lg:col-span-5">
					<div class="space-y-4">
						<div class="flex flex-wrap gap-3">
							<UInput
								v-model="keyword"
								icon="i-heroicons-magnifying-glass-20-solid"
								placeholder="搜索（code / displayName / aliasSlug）"
								class="flex-1"
								clearable
							/>
							<USelect v-model="statusFilter" :items="statusItems" class="w-40" />
						</div>

						<div class="max-h-[70vh] overflow-auto rounded-lg border border-white/10 bg-white/5 p-3">
							<div v-if="loading" class="py-8 text-center text-sm text-white/60">加载中…</div>
							<div v-else-if="filteredTree.length === 0" class="py-8 text-center text-sm text-white/60">暂无类目</div>
							<ul v-else class="space-y-1">
								<CategoryTreeItem
									v-for="node in filteredTree"
									:key="node.id"
									:node="node"
									:selected-id="selected?.id"
									@select="selectCategory"
									@create-child="openCreateChild"
								/>
							</ul>
						</div>
					</div>
				</UCard>

				<UCard class="col-span-12 lg:col-span-7">
					<div v-if="!selected" class="py-10 text-center text-sm text-white/60">
						从左侧类目树选择一个类目开始编辑。
					</div>

					<div v-else class="space-y-6">
						<UTabs v-model="activeTab" :items="tabs" class="w-full" />

						<div v-if="activeTab === 'detail'" class="space-y-6">
							<div class="flex flex-wrap items-start justify-between gap-3">
							<div>
								<div class="flex items-center gap-2">
									<h2 class="text-xl font-semibold">{{ selected.displayName }}</h2>
									<UBadge :color="selected.status === 'enabled' ? 'emerald' : 'gray'" variant="soft">
										{{ selected.status === 'enabled' ? '启用' : '停用' }}
									</UBadge>
								</div>
								<p class="mt-1 text-xs text-white/50">ID: {{ selected.id }} · Path: {{ selected.path }}</p>
							</div>
							<div class="flex flex-wrap gap-2">
								<UButton icon="i-heroicons-plus" variant="soft" @click="openCreateChild(selected)">新建子类目</UButton>
								<UButton
									icon="i-heroicons-shopping-bag"
									variant="soft"
									:to="{ path: '/product/spus', query: { categoryPathPrefix: selected.path, categoryName: selected.displayName } }"
								>
									查看商品
								</UButton>
								<UButton icon="i-heroicons-arrows-right-left" variant="soft" @click="openMove(selected)">迁移</UButton>
								<UButton icon="i-heroicons-trash" color="error" variant="soft" @click="openDeleteConfirm">删除</UButton>
							</div>
						</div>

						<div class="grid grid-cols-12 gap-4">
							<UFormField label="类目编码" class="col-span-12 md:col-span-6">
								<UInput v-model="editForm.code" placeholder="例如: electronics" />
							</UFormField>
							<UFormField label="显示名称" class="col-span-12 md:col-span-6">
								<UInput v-model="editForm.displayName" placeholder="例如: 电子产品" />
							</UFormField>
							<UFormField label="别名/Slug" class="col-span-12 md:col-span-6">
								<UInput v-model="editForm.aliasSlug" placeholder="例如: dianzi" />
							</UFormField>
							<UFormField label="同级排序" class="col-span-12 md:col-span-6">
								<UInput v-model.number="editForm.sortOrder" type="number" />
							</UFormField>
							<UFormField label="展示图 URL" class="col-span-12">
								<UInput v-model="editForm.imageUrl" placeholder="https://..." />
							</UFormField>
							<UFormField label="状态" class="col-span-12 md:col-span-6">
								<div class="flex items-center justify-between rounded-lg border border-white/10 bg-white/5 px-3 py-2">
									<span class="text-sm text-white/70">{{ editForm.status === 'enabled' ? '启用' : '停用' }}</span>
									<USwitch :model-value="editForm.status === 'enabled'" @update:model-value="toggleStatus" />
								</div>
							</UFormField>
							<UFormField label="推荐位" class="col-span-12 md:col-span-6">
								<div class="flex items-center justify-between rounded-lg border border-white/10 bg-white/5 px-3 py-2">
									<span class="text-sm text-white/70">isFeatured</span>
									<USwitch v-model="editForm.isFeatured" />
								</div>
							</UFormField>
						</div>

						<div class="flex flex-wrap justify-end gap-3">
							<UButton variant="ghost" :disabled="saving" @click="resetEdit">重置</UButton>
							<UButton color="primary" :loading="saving" @click="saveEdit">保存</UButton>
						</div>
						</div>

						<div v-else-if="activeTab === 'saleSpecs'" class="space-y-4">
							<div class="flex flex-wrap items-center justify-between gap-3">
								<div>
									<h2 class="text-lg font-semibold">{{ t('product.categories.saleSpecs.title') }}</h2>
									<p class="text-sm text-white/60">{{ t('product.categories.saleSpecs.description') }}</p>
								</div>
								<div class="flex flex-wrap gap-2">
									<UButton variant="soft" icon="i-heroicons-arrow-path" :loading="saleSpecsLoading" @click="loadSaleSpecs">
										{{ t('product.categories.saleSpecs.refresh') }}
									</UButton>
									<UButton color="primary" icon="i-heroicons-plus" @click="addSaleSpecGroup">
										{{ t('product.categories.saleSpecs.addGroup') }}
									</UButton>
								</div>
							</div>

							<div v-if="saleSpecsLoading" class="py-8 text-center text-sm text-white/60">
								{{ t('product.categories.saleSpecs.loading') }}
							</div>
							<div v-else class="space-y-4">
								<UCard v-for="(group, groupIndex) in saleSpecGroups" :key="group.__key">
									<template #header>
										<div class="flex flex-wrap items-center justify-between gap-3">
											<div class="flex flex-wrap items-center gap-2">
												<UBadge variant="soft" color="primary">{{ t('product.categories.saleSpecs.groupBadge', { index: groupIndex + 1 }) }}</UBadge>
												<UBadge v-if="group.required" variant="soft" color="emerald">{{ t('product.categories.saleSpecs.required') }}</UBadge>
												<UBadge v-if="group.allowCustom" variant="soft" color="neutral">{{ t('product.categories.saleSpecs.allowCustom') }}</UBadge>
											</div>
											<UButton color="error" variant="soft" size="xs" @click="removeSaleSpecGroup(groupIndex)">
												{{ t('product.categories.saleSpecs.removeGroup') }}
											</UButton>
										</div>
									</template>

									<div class="grid grid-cols-12 gap-4">
										<UFormField :label="t('product.categories.saleSpecs.code')" class="col-span-12 md:col-span-4">
											<UInput v-model="group.code" placeholder="color" />
										</UFormField>
										<UFormField :label="t('product.categories.saleSpecs.name')" class="col-span-12 md:col-span-4">
											<UInput v-model="group.name" :placeholder="t('product.categories.saleSpecs.namePlaceholder')" />
										</UFormField>
										<UFormField :label="t('product.categories.saleSpecs.sortOrder')" class="col-span-12 md:col-span-4">
											<UInput v-model.number="group.sortOrder" type="number" />
										</UFormField>
										<UFormField :label="t('product.categories.saleSpecs.required')" class="col-span-12 md:col-span-4">
											<USwitch v-model="group.required" />
										</UFormField>
										<UFormField :label="t('product.categories.saleSpecs.allowCustom')" class="col-span-12 md:col-span-4">
											<USwitch v-model="group.allowCustom" />
										</UFormField>
										<UFormField :label="t('product.categories.saleSpecs.status')" class="col-span-12 md:col-span-4">
											<USelect v-model="group.status" :items="saleSpecStatusItems" class="w-full" />
										</UFormField>
									</div>

									<div class="mt-4 space-y-3">
										<div class="flex flex-wrap items-center justify-between gap-2">
											<h3 class="text-sm font-semibold">{{ t('product.categories.saleSpecs.options') }}</h3>
											<UButton variant="soft" size="xs" icon="i-heroicons-plus" @click="addSaleSpecOption(groupIndex)">
												{{ t('product.categories.saleSpecs.addOption') }}
											</UButton>
										</div>
										<div v-if="group.options.length === 0" class="rounded border border-dashed border-white/10 p-3 text-sm text-white/60">
											{{ t('product.categories.saleSpecs.emptyOptions') }}
										</div>
										<div v-else class="space-y-3">
											<div v-for="(option, optionIndex) in group.options" :key="option.__key" class="grid grid-cols-12 gap-3 rounded border border-white/10 p-3">
												<UFormField :label="t('product.categories.saleSpecs.optionCode')" class="col-span-12 md:col-span-3">
													<UInput v-model="option.code" placeholder="red" />
												</UFormField>
												<UFormField :label="t('product.categories.saleSpecs.optionName')" class="col-span-12 md:col-span-3">
													<UInput v-model="option.name" :placeholder="t('product.categories.saleSpecs.optionNamePlaceholder')" />
												</UFormField>
												<UFormField :label="t('product.categories.saleSpecs.sortOrder')" class="col-span-12 md:col-span-2">
													<UInput v-model.number="option.sortOrder" type="number" />
												</UFormField>
												<UFormField :label="t('product.categories.saleSpecs.status')" class="col-span-12 md:col-span-2">
													<USelect v-model="option.status" :items="saleSpecStatusItems" class="w-full" />
												</UFormField>
												<div class="col-span-12 flex items-end md:col-span-2">
													<UButton color="error" variant="soft" size="xs" @click="removeSaleSpecOption(groupIndex, optionIndex)">
														{{ t('product.categories.saleSpecs.removeOption') }}
													</UButton>
												</div>
											</div>
										</div>
									</div>
								</UCard>
								<div v-if="saleSpecGroups.length === 0" class="rounded border border-dashed border-white/10 p-4 text-sm text-white/60">
									{{ t('product.categories.saleSpecs.empty') }}
								</div>
								<div class="flex justify-end">
									<UButton color="primary" :loading="saleSpecsSaving" @click="saveSaleSpecs">
										{{ t('product.categories.saleSpecs.save') }}
									</UButton>
								</div>
							</div>
						</div>

						<div v-else-if="activeTab === 'mappings'" class="space-y-4">
							<div class="flex flex-wrap items-center justify-between gap-3">
								<div>
									<h2 class="text-lg font-semibold">渠道类目映射</h2>
									<p class="text-sm text-white/60">为当前类目维护各渠道的 platformCategoryId。</p>
								</div>
								<div class="flex flex-wrap gap-2">
									<UButton variant="soft" icon="i-heroicons-arrow-path" :loading="mappingLoading" @click="loadMappings">刷新</UButton>
									<UButton variant="soft" icon="i-heroicons-arrow-up-tray" @click="openImport">导入 CSV</UButton>
									<UButton variant="soft" icon="i-heroicons-arrow-down-tray" :loading="exporting" @click="exportCsv">导出 CSV</UButton>
									<UButton color="primary" icon="i-heroicons-plus" @click="openMappingModal()">新增映射</UButton>
								</div>
							</div>

							<UTable :data="mappings" :columns="mappingColumns" :loading="mappingLoading" empty-text="暂无映射，可点击“新增映射”。">
								<template #actions-cell="{ row }">
									<div class="flex flex-wrap gap-2">
										<UButton size="xs" variant="soft" @click="openMappingModal(row.original)">编辑</UButton>
										<UButton size="xs" color="error" variant="soft" :loading="deleting === row.original.channel" @click="deleteMapping(row.original.channel)">
											删除
										</UButton>
									</div>
								</template>
							</UTable>
						</div>

						<div v-else-if="activeTab === 'audit'" class="space-y-4">
							<div class="flex flex-wrap items-center justify-between gap-3">
								<div>
									<h2 class="text-lg font-semibold">审计</h2>
									<p class="text-sm text-white/60">展示与当前类目相关的操作记录（映射/导入导出等）。</p>
								</div>
								<UButton variant="soft" icon="i-heroicons-arrow-path" :loading="auditLoading" @click="loadAudit">刷新</UButton>
							</div>
							<div class="space-y-2">
								<div
									v-for="item in auditItems"
									:key="item.id"
									class="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm"
								>
									<div class="flex flex-wrap items-center justify-between gap-2">
										<div class="font-medium text-white">{{ item.action }}</div>
										<div class="text-xs text-white/50">{{ item.occurredAt }}</div>
									</div>
									<div class="mt-1 text-xs text-white/50">{{ item.summary || item.permissionCode }}</div>
								</div>
								<div v-if="!auditLoading && auditItems.length === 0" class="text-sm text-white/60">暂无记录</div>
							</div>
						</div>
					</div>
				</UCard>
			</div>

			<UModal
				v-model:open="createDialogOpen"
				:prevent-close="savingCreate"
				:close="!savingCreate"
				title="新建类目"
				description="创建根类目或子类目。"
				:ui="{ content: 'max-w-2xl w-[90vw]' }"
			>
				<template #body>
					<div class="space-y-4">
						<UFormField label="父类目">
							<USelectMenu
								v-model="createForm.parentId"
								:items="parentOptions"
								value-key="value"
								label-key="label"
								:portal="false"
								placeholder="根类目（无父级）"
								class="w-full"
							/>
						</UFormField>
						<div class="grid grid-cols-12 gap-4">
							<UFormField label="类目编码" class="col-span-12 md:col-span-6">
								<UInput v-model="createForm.code" placeholder="例如: electronics" />
							</UFormField>
							<UFormField label="显示名称" class="col-span-12 md:col-span-6">
								<UInput v-model="createForm.displayName" placeholder="例如: 电子产品" />
							</UFormField>
							<UFormField label="别名/Slug" class="col-span-12 md:col-span-6">
								<UInput v-model="createForm.aliasSlug" placeholder="例如: dianzi" />
							</UFormField>
							<UFormField label="同级排序" class="col-span-12 md:col-span-3">
								<UInput v-model.number="createForm.sortOrder" type="number" />
							</UFormField>
							<UFormField label="状态" class="col-span-12 md:col-span-3">
								<USelect v-model="createForm.status" :items="statusCreateItems" class="w-full" />
							</UFormField>
							<UFormField label="推荐位" class="col-span-12 md:col-span-6">
								<div class="flex items-center justify-between rounded-lg border border-white/10 bg-white/5 px-3 py-2">
									<span class="text-sm text-white/70">isFeatured</span>
									<USwitch v-model="createForm.isFeatured" />
								</div>
							</UFormField>
							<UFormField label="展示图 URL" class="col-span-12">
								<UInput v-model="createForm.imageUrl" placeholder="https://..." />
							</UFormField>
						</div>
					</div>
				</template>
				<template #footer>
					<div class="flex justify-end gap-3">
						<UButton color="neutral" variant="subtle" type="button" :disabled="savingCreate" @click="closeCreateModal">
							取消
						</UButton>
						<UButton color="primary" type="button" :loading="savingCreate" @click="submitCreate">创建</UButton>
					</div>
				</template>
			</UModal>

			<UModal
				v-model:open="moveDialogOpen"
				:prevent-close="savingMove"
				:close="!savingMove"
				title="迁移类目"
				description="将当前类目迁移到新的父类目下。"
				:ui="{ content: 'max-w-2xl w-[90vw]' }"
			>
				<template #body>
					<div class="space-y-4">
						<p class="text-sm text-white/60">
							将 <span class="font-semibold text-white">{{ moveTarget?.displayName }}</span> 迁移到新的父类目下。
						</p>
						<UFormField label="新父类目">
							<USelectMenu
								v-model="moveForm.parentId"
								:items="moveParentOptions"
								value-key="value"
								label-key="label"
								:portal="false"
								placeholder="根类目（无父级）"
								class="w-full"
							/>
						</UFormField>
						<UFormField label="同级排序">
							<UInput v-model.number="moveForm.sortOrder" type="number" />
						</UFormField>
					</div>
				</template>
				<template #footer>
					<div class="flex justify-end gap-3">
						<UButton color="neutral" variant="subtle" type="button" :disabled="savingMove" @click="closeMoveModal">取消</UButton>
						<UButton color="primary" type="button" :loading="savingMove" @click="submitMove">确认迁移</UButton>
					</div>
				</template>
			</UModal>

			<UModal
				v-model:open="mappingModalOpen"
				:prevent-close="savingMapping"
				:close="!savingMapping"
				:title="mappingEditing ? '编辑映射' : '新增映射'"
				description="为当前类目维护 channel 与 platformCategoryId。"
				:ui="{ content: 'max-w-2xl w-[90vw]' }"
			>
				<template #body>
					<div class="space-y-4">
						<UFormField label="channel">
							<UInput v-model="mappingForm.channel" placeholder="例如: taobao" :disabled="Boolean(mappingEditing)" />
						</UFormField>
						<UFormField label="platformCategoryId">
							<UInput v-model="mappingForm.platformCategoryId" placeholder="渠道类目ID" />
						</UFormField>
						<div class="grid grid-cols-12 gap-4">
							<UFormField label="strategy" class="col-span-12 md:col-span-6">
								<USelect v-model="mappingForm.strategy" :items="strategyItems" class="w-full" />
							</UFormField>
							<UFormField label="syncStatus" class="col-span-12 md:col-span-6">
								<USelect v-model="mappingForm.syncStatus" :items="syncStatusItems" class="w-full" />
							</UFormField>
						</div>
						<UFormField label="metadata (JSON)">
							<UTextarea v-model="mappingForm.metadataText" :rows="5" placeholder='{"key":"value"}' />
						</UFormField>
					</div>
				</template>
				<template #footer>
					<div class="flex justify-end gap-3">
						<UButton color="neutral" variant="subtle" type="button" :disabled="savingMapping" @click="closeMappingModal">取消</UButton>
						<UButton color="primary" type="button" :loading="savingMapping" @click="saveMapping">保存</UButton>
					</div>
				</template>
			</UModal>

			<ConfirmDialog
				v-model="deleteConfirmOpen"
				:loading="deletingCategory"
				confirm-color="error"
				title="删除类目"
				:description="deleteConfirmDescription"
				confirm-text="确认删除"
				cancel-text="取消"
				@confirm="confirmDelete"
				@cancel="closeDeleteConfirm"
			/>

			<input ref="importFileInput" type="file" accept=".csv,text/csv" class="hidden" @change="handleImportFile" />
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref, watch } from 'vue'
import ConfirmDialog from '~/components/ConfirmDialog.vue'
import { useToastAlert } from '~/composables/useToastAlert'
import { useCategoryApi, type CategoryNode, type CategoryCreatePayload, type CategorySaleSpecGroup } from '~/composables/api/useCategory'
import { useCategoryMappingApi, type CategoryMapping, type AuditRecord } from '~/composables/api/useCategoryMapping'

const { t } = useI18n()
const toast = useToastAlert()
const api = useCategoryApi()
const mappingApi = useCategoryMappingApi()

const loading = ref(false)
const saving = ref(false)
const savingCreate = ref(false)
const savingMove = ref(false)
const deletingCategory = ref(false)

const tree = ref<CategoryNode[]>([])
const selected = ref<CategoryNode | null>(null)
const activeTab = ref<'detail' | 'saleSpecs' | 'mappings' | 'audit'>('detail')
const tabs = computed(() => [
	{ label: '详情', value: 'detail' },
	{ label: t('product.categories.saleSpecs.tab'), value: 'saleSpecs' },
	{ label: '渠道映射', value: 'mappings' },
	{ label: '审计', value: 'audit' },
])

const keyword = ref('')
const statusFilter = ref<'all' | 'enabled' | 'disabled'>('all')
const statusItems = [
	{ label: '全部', value: 'all' },
	{ label: '启用', value: 'enabled' },
	{ label: '停用', value: 'disabled' },
]
const statusCreateItems = [
	{ label: '启用', value: 'enabled' },
	{ label: '停用', value: 'disabled' },
]

const editForm = ref({
	code: '',
	displayName: '',
	aliasSlug: '',
	sortOrder: 0,
	status: 'enabled' as 'enabled' | 'disabled',
	imageUrl: '',
	isFeatured: false,
})

const mappingLoading = ref(false)
const exporting = ref(false)
const savingMapping = ref(false)
const deleting = ref('')
const mappings = ref<CategoryMapping[]>([])
const mappingColumns = computed(() => [
	{ accessorKey: 'channel', header: 'channel' },
	{ accessorKey: 'platformCategoryId', header: 'platformCategoryId' },
	{ accessorKey: 'strategy', header: 'strategy' },
	{ accessorKey: 'syncStatus', header: 'syncStatus' },
	{ id: 'actions', header: '操作' },
])
const mappingModalOpen = ref(false)
const mappingEditing = ref<CategoryMapping | null>(null)
const mappingForm = ref({
	channel: '',
	platformCategoryId: '',
	strategy: 'manual',
	syncStatus: 'pending',
	metadataText: '{}',
})
const strategyItems = [
	{ label: 'manual', value: 'manual' },
	{ label: 'auto', value: 'auto' },
]
const syncStatusItems = [
	{ label: 'pending', value: 'pending' },
	{ label: 'synced', value: 'synced' },
	{ label: 'failed', value: 'failed' },
	{ label: 'disabled', value: 'disabled' },
]

type EditableSaleSpecOption = {
	__key: string
	id?: string
	groupId?: string
	code: string
	name: string
	sortOrder: number
	meta?: Record<string, any> | null
	status: string
}

type EditableSaleSpecGroup = {
	__key: string
	id?: string
	categoryId?: string
	code: string
	name: string
	sortOrder: number
	required: boolean
	allowCustom: boolean
	status: string
	options: EditableSaleSpecOption[]
}

const saleSpecsLoading = ref(false)
const saleSpecsSaving = ref(false)
const saleSpecGroups = ref<EditableSaleSpecGroup[]>([])
const saleSpecStatusItems = computed(() => [
	{ label: t('product.categories.saleSpecs.statusActive'), value: 'active' },
	{ label: t('product.categories.saleSpecs.statusDisabled'), value: 'disabled' },
])

const auditLoading = ref(false)
const auditItems = ref<AuditRecord[]>([])

const createDialogOpen = ref(false)
const createForm = ref<CategoryCreatePayload>({
	parentId: null,
	code: '',
	displayName: '',
	aliasSlug: '',
	sortOrder: 0,
	status: 'enabled',
	imageUrl: '',
	isFeatured: false,
})

const moveDialogOpen = ref(false)
const moveTarget = ref<CategoryNode | null>(null)
const moveForm = ref<{ parentId: string | null; sortOrder: number }>({ parentId: null, sortOrder: 0 })
const deleteConfirmOpen = ref(false)

const deleteConfirmDescription = computed(() => {
	if (!selected.value) return '确认删除该类目？'
	return '仅允许删除：无子类目，且未被商品引用的类目。'
})

watch(createDialogOpen, (open) => {
	if (!open) document.activeElement?.blur()
})
watch(moveDialogOpen, (open) => {
	if (!open) document.activeElement?.blur()
})
watch(mappingModalOpen, (open) => {
	if (!open) document.activeElement?.blur()
})
watch(deleteConfirmOpen, (open) => {
	if (!open) document.activeElement?.blur()
})

const refreshTree = async () => {
	loading.value = true
	try {
		const resp = await api.tree()
		tree.value = resp?.items ?? []
		if (selected.value) {
			selected.value = findNode(tree.value, selected.value.id)
		}
	} catch (error: any) {
		toast.add({ title: '加载失败', description: error?.message || '无法获取类目树', color: 'error' })
	} finally {
		loading.value = false
	}
}

const selectCategory = (node: CategoryNode) => {
	selected.value = node
	resetEdit()
	if (activeTab.value === 'saleSpecs') {
		loadSaleSpecs()
	}
	if (activeTab.value === 'mappings') {
		loadMappings()
	}
	if (activeTab.value === 'audit') {
		loadAudit()
	}
}

const openDeleteConfirm = () => {
	if (!selected.value) return
	document.activeElement?.blur()
	deleteConfirmOpen.value = true
}

const closeDeleteConfirm = () => {
	document.activeElement?.blur()
	deleteConfirmOpen.value = false
}

const confirmDelete = async () => {
	if (!selected.value) return
	deletingCategory.value = true
	try {
		await api.remove(selected.value.id)
		toast.add({ title: '删除成功', description: selected.value.displayName, color: 'success' })
		closeDeleteConfirm()
		selected.value = null
		activeTab.value = 'detail'
		await refreshTree()
	} catch (error: any) {
		toast.add({ title: '删除失败', description: error?.message || '请确认该类目无子类目且未被商品引用', color: 'error' })
	} finally {
		deletingCategory.value = false
	}
}

const resetEdit = () => {
	if (!selected.value) return
	editForm.value = {
		code: selected.value.code,
		displayName: selected.value.displayName,
		aliasSlug: selected.value.aliasSlug,
		sortOrder: selected.value.sortOrder ?? 0,
		status: (selected.value.status as any) === 'disabled' ? 'disabled' : 'enabled',
		imageUrl: selected.value.imageUrl ?? '',
		isFeatured: Boolean(selected.value.isFeatured),
	}
}

watch(selected, () => resetEdit(), { immediate: false })

const saveEdit = async () => {
	if (!selected.value) return
	saving.value = true
	try {
		const payload = {
			code: editForm.value.code,
			displayName: editForm.value.displayName,
			aliasSlug: editForm.value.aliasSlug,
			sortOrder: editForm.value.sortOrder,
			imageUrl: editForm.value.imageUrl,
			isFeatured: editForm.value.isFeatured,
		}
		const updated = await api.update(selected.value.id, payload)
		toast.add({ title: '保存成功', description: updated.displayName, color: 'success' })
		await refreshTree()
		selected.value = findNode(tree.value, updated.id)
	} catch (error: any) {
		toast.add({ title: '保存失败', description: error?.message || '保存失败', color: 'error' })
	} finally {
		saving.value = false
	}
}

watch(activeTab, (tab) => {
	if (!selected.value) return
	if (tab === 'saleSpecs') {
		loadSaleSpecs()
	}
	if (tab === 'mappings') {
		loadMappings()
	}
	if (tab === 'audit') {
		loadAudit()
	}
})

const saleSpecKey = () => `${Date.now()}-${Math.random().toString(16).slice(2)}`

const toEditableSaleSpecs = (groups: CategorySaleSpecGroup[]): EditableSaleSpecGroup[] =>
	(groups || []).map((group) => ({
		__key: saleSpecKey(),
		id: group.id,
		categoryId: group.categoryId,
		code: group.code,
		name: group.name,
		sortOrder: Number(group.sortOrder || 0),
		required: Boolean(group.required),
		allowCustom: Boolean(group.allowCustom),
		status: group.status || 'active',
		options: (group.options || []).map((option) => ({
			__key: saleSpecKey(),
			id: option.id,
			groupId: option.groupId,
			code: option.code,
			name: option.name,
			sortOrder: Number(option.sortOrder || 0),
			meta: option.meta || null,
			status: option.status || 'active',
		})),
	}))

const loadSaleSpecs = async () => {
	if (!selected.value) return
	saleSpecsLoading.value = true
	try {
		const resp = await api.saleSpecs(selected.value.id)
		saleSpecGroups.value = toEditableSaleSpecs(resp.groups || [])
	} catch (error: any) {
		toast.add({
			title: t('product.categories.saleSpecs.loadFailed'),
			description: error?.message || t('product.categories.saleSpecs.retry'),
			color: 'error',
		})
	} finally {
		saleSpecsLoading.value = false
	}
}

const addSaleSpecGroup = () => {
	saleSpecGroups.value.push({
		__key: saleSpecKey(),
		code: '',
		name: '',
		sortOrder: saleSpecGroups.value.length,
		required: true,
		allowCustom: false,
		status: 'active',
		options: [],
	})
}

const removeSaleSpecGroup = (index: number) => {
	saleSpecGroups.value.splice(index, 1)
}

const addSaleSpecOption = (groupIndex: number) => {
	const group = saleSpecGroups.value[groupIndex]
	if (!group) return
	group.options.push({
		__key: saleSpecKey(),
		code: '',
		name: '',
		sortOrder: group.options.length,
		meta: null,
		status: 'active',
	})
}

const removeSaleSpecOption = (groupIndex: number, optionIndex: number) => {
	const group = saleSpecGroups.value[groupIndex]
	if (!group) return
	group.options.splice(optionIndex, 1)
}

const validateSaleSpecs = () => {
	for (const group of saleSpecGroups.value) {
		if (!group.code.trim() || !group.name.trim()) {
			return t('product.categories.saleSpecs.groupRequired')
		}
		const optionCodes = new Set<string>()
		for (const option of group.options) {
			if (!option.code.trim() || !option.name.trim()) {
				return t('product.categories.saleSpecs.optionRequired')
			}
			const code = option.code.trim().toLowerCase()
			if (optionCodes.has(code)) {
				return t('product.categories.saleSpecs.duplicateOption')
			}
			optionCodes.add(code)
		}
	}
	return ''
}

const saveSaleSpecs = async () => {
	if (!selected.value) return
	const error = validateSaleSpecs()
	if (error) {
		toast.add({ title: error, color: 'error' })
		return
	}
	saleSpecsSaving.value = true
	try {
		const payload = {
			groups: saleSpecGroups.value.map((group) => ({
				id: group.id,
				code: group.code.trim(),
				name: group.name.trim(),
				sortOrder: Number(group.sortOrder || 0),
				required: group.required,
				allowCustom: group.allowCustom,
				status: group.status || 'active',
				options: group.options.map((option) => ({
					id: option.id,
					code: option.code.trim(),
					name: option.name.trim(),
					sortOrder: Number(option.sortOrder || 0),
					meta: option.meta || null,
					status: option.status || 'active',
				})),
			})),
		}
		const resp = await api.replaceSaleSpecs(selected.value.id, payload)
		saleSpecGroups.value = toEditableSaleSpecs(resp.groups || [])
		toast.add({ title: t('product.categories.saleSpecs.saveSuccess'), color: 'success' })
	} catch (error: any) {
		toast.add({
			title: t('product.categories.saleSpecs.saveFailed'),
			description: error?.message || t('product.categories.saleSpecs.retry'),
			color: 'error',
		})
	} finally {
		saleSpecsSaving.value = false
	}
}

const loadMappings = async () => {
	if (!selected.value) return
	mappingLoading.value = true
	try {
		const resp = await mappingApi.list(selected.value.id)
		mappings.value = resp.items ?? []
	} catch (error: any) {
		toast.add({ title: '加载映射失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		mappingLoading.value = false
	}
}

const openMappingModal = (row?: CategoryMapping) => {
	if (!selected.value) return
	mappingEditing.value = row ?? null
	mappingForm.value = {
		channel: row?.channel ?? '',
		platformCategoryId: row?.platformCategoryId ?? '',
		strategy: row?.strategy ?? 'manual',
		syncStatus: row?.syncStatus ?? 'pending',
		metadataText: JSON.stringify(row?.metadata ?? {}, null, 2),
	}
	mappingModalOpen.value = true
}

const closeMappingModal = () => {
	document.activeElement?.blur()
	mappingModalOpen.value = false
}

const saveMapping = async () => {
	if (!selected.value) return
	const categoryId = selected.value.id
	const channel = mappingForm.value.channel.trim()
	const platformCategoryId = mappingForm.value.platformCategoryId.trim()
	if (!channel || !platformCategoryId) {
		toast.add({ title: '请填写 channel 与 platformCategoryId', color: 'error' })
		return
	}
	savingMapping.value = true
	try {
		let metadata: any = {}
		if (mappingForm.value.metadataText.trim()) {
			metadata = JSON.parse(mappingForm.value.metadataText)
		}
		await mappingApi.upsert(categoryId, {
			channel,
			platformCategoryId,
			strategy: mappingForm.value.strategy,
			syncStatus: mappingForm.value.syncStatus,
			metadata,
		})
		toast.add({ title: '保存成功', description: channel, color: 'success' })
		closeMappingModal()
		await loadMappings()
	} catch (error: any) {
		toast.add({ title: '保存失败', description: error?.message || '请检查字段', color: 'error' })
	} finally {
		savingMapping.value = false
	}
}

const deleteMapping = async (channel: string) => {
	if (!selected.value) return
	deleting.value = channel
	try {
		await mappingApi.remove(selected.value.id, channel)
		toast.add({ title: '已删除', description: channel, color: 'success' })
		await loadMappings()
	} catch (error: any) {
		toast.add({ title: '删除失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		deleting.value = ''
	}
}

const importFileInput = ref<HTMLInputElement | null>(null)
const openImport = () => {
	if (!selected.value) return
	importFileInput.value?.click()
}

const handleImportFile = async (e: Event) => {
	if (!selected.value) return
	const input = e.target as HTMLInputElement
	const file = input.files?.[0]
	if (!file) return
	try {
		const result = await mappingApi.importMappings(file, selected.value.id)
		if (result.failed > 0) {
			toast.add({ title: '导入存在失败行', description: `failed=${result.failed}`, color: 'error' })
		} else {
			toast.add({ title: '导入成功', description: `succeeded=${result.succeeded}`, color: 'success' })
		}
		await loadMappings()
	} catch (error: any) {
		toast.add({ title: '导入失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		input.value = ''
	}
}

const exportCsv = async () => {
	if (!selected.value) return
	exporting.value = true
	try {
		const response = await mappingApi.exportMappings(selected.value.id)
		const blob = response?._data as Blob
		if (!blob) throw new Error('export failed')
		const disposition = response.headers?.get?.('content-disposition')
		let fallback = `category_mappings_${selected.value.id}.csv`
		if (disposition) {
			const match = disposition.match(/filename=\"?([^\";]+)\"?/i)
			if (match?.[1]) fallback = decodeURIComponent(match[1])
		}
		const url = URL.createObjectURL(blob)
		const link = document.createElement('a')
		link.href = url
		link.download = fallback
		document.body.appendChild(link)
		link.click()
		document.body.removeChild(link)
		URL.revokeObjectURL(url)
	} catch (error: any) {
		toast.add({ title: '导出失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		exporting.value = false
	}
}

const loadAudit = async () => {
	if (!selected.value) return
	auditLoading.value = true
	try {
		const resp = await mappingApi.audit(selected.value.id, { limit: 50 })
		auditItems.value = resp.items ?? []
	} catch (error: any) {
		toast.add({ title: '加载审计失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		auditLoading.value = false
	}
}

const toggleStatus = async (enabled: boolean) => {
	if (!selected.value) return
	const next = enabled ? 'enabled' : 'disabled'
	editForm.value.status = next
	try {
		await api.setStatus(selected.value.id, { status: next })
		await refreshTree()
		selected.value = findNode(tree.value, selected.value.id)
		toast.add({ title: '状态已更新', description: next === 'enabled' ? '已启用' : '已停用', color: 'success' })
	} catch (error: any) {
		toast.add({ title: '更新失败', description: error?.message || '状态更新失败', color: 'error' })
		resetEdit()
	}
}

const openCreateRoot = () => {
	createForm.value = { parentId: null, code: '', displayName: '', aliasSlug: '', sortOrder: 0, status: 'enabled', imageUrl: '', isFeatured: false }
	createDialogOpen.value = true
}

const closeCreateModal = () => {
	document.activeElement?.blur()
	createDialogOpen.value = false
}

const openCreateChild = (parent: CategoryNode) => {
	createForm.value = {
		parentId: parent.id,
		code: '',
		displayName: '',
		aliasSlug: '',
		sortOrder: 0,
		status: 'enabled',
		imageUrl: '',
		isFeatured: false,
	}
	createDialogOpen.value = true
}

const submitCreate = async () => {
	savingCreate.value = true
	try {
		const payload = { ...createForm.value }
		const created = await api.create(payload as any)
		closeCreateModal()
		toast.add({ title: '创建成功', description: created.displayName, color: 'success' })
		await refreshTree()
		selected.value = findNode(tree.value, created.id)
	} catch (error: any) {
		toast.add({ title: '创建失败', description: error?.message || '创建失败', color: 'error' })
	} finally {
		savingCreate.value = false
	}
}

const openMove = (node: CategoryNode) => {
	moveTarget.value = node
	moveForm.value = { parentId: node.parentId ?? null, sortOrder: node.sortOrder ?? 0 }
	moveDialogOpen.value = true
}

const closeMoveModal = () => {
	document.activeElement?.blur()
	moveDialogOpen.value = false
}

const submitMove = async () => {
	if (!moveTarget.value) return
	savingMove.value = true
	try {
		const payload = { parentId: moveForm.value.parentId, sortOrder: moveForm.value.sortOrder }
		const moved = await api.move(moveTarget.value.id, payload)
		closeMoveModal()
		toast.add({ title: '迁移成功', description: moved.displayName, color: 'success' })
		await refreshTree()
		selected.value = findNode(tree.value, moved.id)
	} catch (error: any) {
		toast.add({ title: '迁移失败', description: error?.message || '迁移失败', color: 'error' })
	} finally {
		savingMove.value = false
	}
}

const parentOptions = computed(() => [{ label: '根类目（无父级）', value: null }, ...flattenTree(tree.value)])
const moveParentOptions = computed(() => {
	const items = flattenTree(tree.value)
	if (!moveTarget.value) {
		return [{ label: '根类目（无父级）', value: null }, ...items]
	}
	const excluded = new Set<string>()
	collectSubtreeIds(moveTarget.value, excluded)
	return [{ label: '根类目（无父级）', value: null }, ...items.filter((opt) => !opt.value || !excluded.has(opt.value))]
})

const filteredTree = computed(() => filterTree(tree.value, keyword.value, statusFilter.value))

onMounted(() => {
	refreshTree()
})

function findNode(nodes: CategoryNode[], id: string): CategoryNode | null {
	for (const node of nodes) {
		if (node.id === id) return node
		const child = node.children ? findNode(node.children, id) : null
		if (child) return child
	}
	return null
}

function flattenTree(nodes: CategoryNode[], level = 0): Array<{ label: string; value: string | null }> {
	const out: Array<{ label: string; value: string | null }> = []
	for (const node of nodes) {
		const prefix = level > 0 ? `${'—'.repeat(Math.min(level, 6))} ` : ''
		out.push({ label: `${prefix}${node.displayName}`, value: node.id })
		if (node.children?.length) {
			out.push(...flattenTree(node.children, level + 1))
		}
	}
	return out
}

function filterTree(nodes: CategoryNode[], keyword: string, status: 'all' | 'enabled' | 'disabled'): CategoryNode[] {
	const kw = keyword.trim().toLowerCase()
	const matchStatus = (n: CategoryNode) => status === 'all' || String(n.status).toLowerCase() === status
	const matchKeyword = (n: CategoryNode) => {
		if (!kw) return true
		return (
			n.code?.toLowerCase().includes(kw) ||
			n.displayName?.toLowerCase().includes(kw) ||
			n.aliasSlug?.toLowerCase().includes(kw)
		)
	}

	const walk = (items: CategoryNode[]): CategoryNode[] => {
		const out: CategoryNode[] = []
		for (const node of items) {
			const children = node.children ? walk(node.children) : []
			const ok = matchStatus(node) && matchKeyword(node)
			if (ok || children.length) {
				out.push({ ...node, children })
			}
		}
		return out
	}
	return walk(nodes)
}

function collectSubtreeIds(node: CategoryNode, out: Set<string>) {
	out.add(node.id)
	for (const child of node.children ?? []) {
		collectSubtreeIds(child, out)
	}
}

const CategoryTreeItem = defineComponent({
	name: 'CategoryTreeItem',
	props: {
		node: { type: Object as any, required: true },
		selectedId: { type: String, default: '' },
	},
	emits: ['select', 'create-child'],
	setup(props, { emit }) {
		const open = ref(true)
		const hasChildren = computed(() => (props.node.children?.length ?? 0) > 0)
		const isSelected = computed(() => props.selectedId === props.node.id)
		return () =>
			h('li', { class: 'space-y-1' }, [
				h(
					'div',
					{
						class: [
							'group flex items-center justify-between rounded-lg border px-2 py-1.5',
							isSelected.value ? 'border-primary-400/60 bg-primary-500/10' : 'border-white/10 bg-white/0 hover:bg-white/5',
						],
					},
					[
						h('button', {
							class: 'flex flex-1 items-center gap-2 text-left',
							onClick: () => emit('select', props.node),
						}, [
							hasChildren.value
								? h('span', { class: 'inline-flex w-5 justify-center text-white/60', onClick: (e: any) => { e.stopPropagation(); open.value = !open.value } }, open.value ? '▾' : '▸')
								: h('span', { class: 'inline-flex w-5 justify-center text-white/20' }, '•'),
							h('div', { class: 'min-w-0' }, [
								h('div', { class: 'truncate text-sm text-white' }, props.node.displayName),
								h('div', { class: 'truncate text-xs text-white/40' }, props.node.code),
							]),
						]),
						h('div', { class: 'flex items-center gap-2 opacity-0 transition group-hover:opacity-100' }, [
							h('button', { class: 'text-xs text-white/60 hover:text-white', onClick: (e: any) => { e.stopPropagation(); emit('create-child', props.node) } }, '+ 子类目'),
						]),
					],
				),
				open.value && hasChildren.value
					? h('ul', { class: 'ml-4 space-y-1 border-l border-white/10 pl-3' }, props.node.children.map((child: any) =>
							h(CategoryTreeItem as any, {
								node: child,
								selectedId: props.selectedId,
								onSelect: (n: any) => emit('select', n),
								onCreateChild: (n: any) => emit('create-child', n),
							}),
					  ))
					: null,
			])
	},
})
</script>
