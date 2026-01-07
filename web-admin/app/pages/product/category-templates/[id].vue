<template>
	<div class="px-6 py-8 text-white">
		<div class="mx-auto flex max-w-6xl flex-col gap-6">
			<header class="flex flex-wrap items-center justify-between gap-3">
				<div class="flex items-center gap-3">
					<NuxtLink to="/product/category-templates" class="text-sm text-primary">← 返回列表</NuxtLink>
					<div>
						<p class="text-xs uppercase tracking-[0.2em] text-primary-200">商品中心</p>
						<h1 class="mt-1 text-2xl font-semibold text-white">
							模板编辑
							<span v-if="template" class="text-white/70">· {{ template.name }}</span>
						</h1>
					</div>
				</div>
				<div class="flex flex-wrap gap-2">
					<UButton variant="soft" icon="i-heroicons-arrow-path" :loading="loading" @click="refreshAll">刷新</UButton>
					<UButton color="primary" icon="i-heroicons-check" :loading="saving" @click="save">保存</UButton>
					<UButton color="emerald" icon="i-heroicons-rocket-launch" :loading="publishing" @click="publish">发布</UButton>
				</div>
			</header>

			<UCard>
				<div v-if="!template && loading" class="py-10 text-center text-sm text-white/60">加载中…</div>
				<div v-else-if="!template" class="py-10 text-center text-sm text-white/60">模板不存在或无权限访问</div>
				<div v-else class="space-y-6">
					<div class="grid grid-cols-12 gap-4">
						<UFormField label="名称" class="col-span-12 md:col-span-6">
							<UInput v-model="form.name" />
						</UFormField>
						<UFormField label="状态" class="col-span-12 md:col-span-3">
							<USelect v-model="form.status" :items="statusItems" class="w-full" />
						</UFormField>
						<UFormField label="已发布版本" class="col-span-12 md:col-span-3">
							<div class="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white/70">
								v{{ template.currentPublishedVersionNo || 0 }}
							</div>
						</UFormField>
						<UFormField label="备注" class="col-span-12">
							<UTextarea v-model="form.notes" :rows="3" placeholder="用于记录模板用途、变更说明等" />
						</UFormField>
					</div>

					<div class="flex flex-wrap items-center justify-between gap-3">
						<h2 class="text-lg font-semibold">字段</h2>
						<UButton variant="soft" icon="i-heroicons-plus" @click="addField">新增字段</UButton>
					</div>

					<div class="overflow-x-auto rounded-lg border border-white/10 bg-white/5">
						<table class="min-w-full text-sm">
							<thead class="text-left text-white/70">
								<tr class="border-b border-white/10">
									<th class="px-3 py-2">fieldKey</th>
									<th class="px-3 py-2">类型</th>
									<th class="px-3 py-2">必填</th>
									<th class="px-3 py-2">排序</th>
									<th class="px-3 py-2"></th>
								</tr>
							</thead>
							<tbody>
								<tr v-for="(f, idx) in form.fields" :key="idx" class="border-b border-white/10 last:border-b-0">
									<td class="px-3 py-2">
										<UInput v-model="f.fieldKey" placeholder="例如: brand" />
									</td>
									<td class="px-3 py-2">
										<USelect v-model="f.fieldType" :items="fieldTypeItems" class="w-40" />
									</td>
									<td class="px-3 py-2">
										<USwitch v-model="f.required" />
									</td>
									<td class="px-3 py-2">
										<UInput v-model.number="f.sortOrder" type="number" class="w-24" />
									</td>
									<td class="px-3 py-2 text-right">
										<UButton color="error" variant="ghost" icon="i-heroicons-trash" @click="removeField(idx)" />
									</td>
								</tr>
								<tr v-if="form.fields.length === 0">
									<td colspan="5" class="px-3 py-6 text-center text-white/60">暂无字段，点击“新增字段”开始配置。</td>
								</tr>
							</tbody>
						</table>
					</div>

					<UCard class="border border-white/10 bg-white/5">
						<template #header>
							<div class="flex flex-wrap items-center justify-between gap-3">
								<h3 class="text-base font-semibold">预览 / 模拟校验</h3>
								<UButton variant="soft" icon="i-heroicons-play" :loading="simulating" @click="simulate">
									校验样例 attributes
								</UButton>
							</div>
						</template>
						<div class="space-y-3">
							<p class="text-sm text-white/60">用于快速验证必填/类型/枚举规则（MVP）。</p>
							<UTextarea v-model="attributesJson" :rows="6" placeholder='{"brand":"ACME"}' />
						</div>
					</UCard>

					<UCard class="border border-white/10 bg-white/5">
						<template #header>
							<div class="flex flex-wrap items-center justify-between gap-3">
								<h3 class="text-base font-semibold">影响范围</h3>
								<div class="flex gap-2">
									<UButton variant="soft" icon="i-heroicons-arrow-path" :loading="impactLoading" @click="loadImpact">刷新</UButton>
									<UButton variant="soft" icon="i-heroicons-bolt" :loading="recheckLoading" @click="triggerRecheck">触发批量重检</UButton>
								</div>
							</div>
						</template>
						<div class="grid grid-cols-12 gap-4 text-sm">
							<div class="col-span-12 md:col-span-6">
								<div class="text-white/60">绑定类目数</div>
								<div class="mt-1 text-xl font-semibold">{{ impact?.boundCategories ?? 0 }}</div>
							</div>
							<div class="col-span-12 md:col-span-6">
								<div class="text-white/60">受影响商品数（估算）</div>
								<div class="mt-1 text-xl font-semibold">{{ impact?.affectedProducts ?? 0 }}</div>
							</div>
						</div>
					</UCard>

					<UCard class="border border-white/10 bg-white/5">
						<template #header>
							<div class="flex flex-wrap items-center justify-between gap-3">
								<h3 class="text-base font-semibold">版本</h3>
								<div class="flex items-center gap-2">
									<USelectMenu
										v-model="rollbackTarget"
										:options="versionOptions"
										value-attribute="value"
										option-attribute="label"
										placeholder="选择回滚目标版本"
										class="min-w-64"
									/>
									<UButton variant="soft" :disabled="!rollbackTarget" :loading="rollingBack" @click="rollback">
										回滚并发布
									</UButton>
								</div>
							</div>
						</template>
						<ul class="space-y-2 text-sm text-white/70">
							<li v-for="v in versions" :key="v.id" class="flex items-center justify-between rounded-lg border border-white/10 bg-white/5 px-3 py-2">
								<div>
									<div class="font-medium text-white">v{{ v.versionNumber }}</div>
									<div class="text-xs text-white/50">{{ v.publishedAt || v.createdAt }}</div>
								</div>
								<div class="text-xs text-white/50">
									<span v-if="v.rollbackFrom">rollbackFrom: {{ v.rollbackFrom }}</span>
									<span v-else>publishedBy: {{ v.publishedBy || 'system' }}</span>
								</div>
							</li>
							<li v-if="versions.length === 0" class="text-white/60">暂无版本，点击“发布”生成首个版本。</li>
						</ul>
					</UCard>
				</div>
			</UCard>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useToastAlert } from '~/composables/useToastAlert'
import { useCategoryTemplateApi, type CategoryTemplate, type CategoryTemplateField, type TemplateVersionSummary, type TemplateImpactSummary } from '~/composables/api/useCategoryTemplate'

const route = useRoute()
const api = useCategoryTemplateApi()
const toast = useToastAlert()

const id = computed(() => String(route.params.id || '').trim())

const loading = ref(false)
const saving = ref(false)
const publishing = ref(false)
const rollingBack = ref(false)
const simulating = ref(false)
const impactLoading = ref(false)
const recheckLoading = ref(false)

const template = ref<CategoryTemplate | null>(null)
const versions = ref<TemplateVersionSummary[]>([])
const rollbackTarget = ref<string | null>(null)
const impact = ref<TemplateImpactSummary | null>(null)

const statusItems = [
	{ label: '启用', value: 'enabled' },
	{ label: '停用', value: 'disabled' },
]
const fieldTypeItems = ['string', 'number', 'boolean', 'enum', 'json', 'date'].map((v) => ({ label: v, value: v }))

const form = reactive<{ name: string; status: string; notes: string; fields: CategoryTemplateField[] }>({
	name: '',
	status: 'enabled',
	notes: '',
	fields: [],
})

const attributesJson = ref('{\n  \n}')

const versionOptions = computed(() =>
	versions.value.map((v) => ({
		label: `v${v.versionNumber} · ${v.publishedAt || v.createdAt || ''}`.trim(),
		value: v.id,
	}))
)

const loadTemplate = async () => {
	if (!id.value) return
	loading.value = true
	try {
		const detail = await api.get(id.value)
		template.value = detail
		form.name = detail.name || ''
		form.status = detail.status || 'enabled'
		form.notes = detail.notes || ''
		form.fields = (detail.fields || []).map((f) => ({
			fieldKey: f.fieldKey,
			fieldType: f.fieldType || 'string',
			required: Boolean(f.required),
			sortOrder: Number.isFinite(f.sortOrder) ? Number(f.sortOrder) : 0,
			validationRules: f.validationRules,
			defaultValue: f.defaultValue,
			inheritable: f.inheritable ?? true,
		}))
	} catch (error: any) {
		toast.add({ title: '加载失败', description: error?.message || '无法获取模板详情', color: 'error' })
		template.value = null
	} finally {
		loading.value = false
	}
}

const loadVersions = async () => {
	if (!id.value) return
	try {
		const resp = await api.versions(id.value, { limit: 50 })
		versions.value = resp.items ?? []
	} catch {
		versions.value = []
	}
}

const loadImpact = async () => {
	if (!id.value) return
	impactLoading.value = true
	try {
		impact.value = await api.impact(id.value)
	} catch (error: any) {
		toast.add({ title: '获取影响范围失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		impactLoading.value = false
	}
}

const refreshAll = async () => {
	await Promise.all([loadTemplate(), loadVersions(), loadImpact()])
}

const save = async () => {
	if (!id.value) return
	saving.value = true
	try {
		const updated = await api.update(id.value, {
			name: form.name,
			status: form.status as any,
			notes: form.notes,
			fields: form.fields.map((f) => ({
				fieldKey: String(f.fieldKey || '').trim(),
				fieldType: String(f.fieldType || 'string').trim(),
				required: Boolean(f.required),
				sortOrder: Number.isFinite(f.sortOrder) ? Number(f.sortOrder) : 0,
				inheritable: f.inheritable ?? true,
			})),
		})
		template.value = updated
		toast.add({ title: '保存成功', description: updated.name, color: 'success' })
		await refreshAll()
	} catch (error: any) {
		toast.add({ title: '保存失败', description: error?.message || '请检查字段配置', color: 'error' })
	} finally {
		saving.value = false
	}
}

const publish = async () => {
	if (!id.value) return
	publishing.value = true
	try {
		await api.publish(id.value, {})
		toast.add({ title: '已发布', description: '新版本已生成', color: 'success' })
		await refreshAll()
	} catch (error: any) {
		toast.add({ title: '发布失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		publishing.value = false
	}
}

const rollback = async () => {
	if (!id.value || !rollbackTarget.value) return
	rollingBack.value = true
	try {
		await api.rollback(id.value, { targetVersionId: rollbackTarget.value })
		toast.add({ title: '回滚已发布', description: '已生成新版本并切换生效', color: 'success' })
		await refreshAll()
	} catch (error: any) {
		toast.add({ title: '回滚失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		rollingBack.value = false
	}
}

const simulate = async () => {
	if (!id.value) return
	simulating.value = true
	try {
		const attrs = JSON.parse(attributesJson.value || '{}')
		await api.simulate(id.value, { templateId: id.value, attributes: attrs })
		toast.add({ title: '校验通过', description: 'attributes 符合模板规则', color: 'success' })
	} catch (error: any) {
		toast.add({ title: '校验失败', description: error?.message || '请检查必填与类型', color: 'error' })
	} finally {
		simulating.value = false
	}
}

const triggerRecheck = async () => {
	if (!id.value) return
	recheckLoading.value = true
	try {
		const result = await api.triggerRecheck(id.value, { reason: 'manual recheck', dryRun: true })
		toast.add({ title: '已触发批量重检（dryRun）', description: `jobRunId: ${result.jobRunId}`, color: 'success' })
	} catch (error: any) {
		toast.add({ title: '触发失败', description: error?.message || '请稍后重试', color: 'error' })
	} finally {
		recheckLoading.value = false
	}
}

const addField = () => {
	form.fields.push({ fieldKey: '', fieldType: 'string', required: false, sortOrder: form.fields.length * 10 })
}

const removeField = (idx: number) => {
	form.fields.splice(idx, 1)
}

onMounted(() => refreshAll())
</script>

