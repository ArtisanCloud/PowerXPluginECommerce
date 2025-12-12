<template>
	<div class="space-y-6 px-6 py-8">
		<div class="flex flex-wrap items-center justify-between gap-4">
			<NuxtLink to="/product/spus" class="text-sm text-primary">← 返回列表</NuxtLink>
			<div class="flex flex-wrap gap-2">
				<UButton v-if="spu?.status === 'draft'" color="primary" :loading="submitLoading" @click="submitDraft">
					提交审核
				</UButton>
				<UButton v-if="spu?.status === 'reviewing'" color="emerald" :loading="publishLoading" @click="publishDraft">
					发布
				</UButton>
			</div>
		</div>

		<div v-if="loading">
			<USkeleton class="h-24" />
		</div>
		<div v-else-if="spu" class="space-y-6">
			<div>
				<h1 class="text-2xl font-semibold">{{ spu.name }}</h1>
				<p class="text-sm text-gray-500">当前状态：<UBadge>{{ spu.status }}</UBadge></p>
			</div>
			<UCard>
				<dl class="grid gap-4 md:grid-cols-2">
					<div>
						<dt class="text-sm text-gray-500">编码</dt>
						<dd class="text-base font-medium">{{ spu.code }}</dd>
					</div>
					<div>
						<dt class="text-sm text-gray-500">类目</dt>
						<dd class="text-base font-medium">{{ spu.categoryPath }}</dd>
					</div>
					<div>
						<dt class="text-sm text-gray-500">负责人</dt>
						<dd class="text-base font-medium">{{ spu.responsibleUser || '—' }}</dd>
					</div>
					<div>
						<dt class="text-sm text-gray-500">默认语言</dt>
						<dd class="text-base font-medium">{{ spu.defaultLocale }}</dd>
					</div>
				</dl>
			</UCard>
			<SpuSkuLinker :spu-id="spu.id" />
			<VersionDiff
				:versions="versions"
				:selected-id="selectedVersionId || undefined"
				:detail="versionDetail"
				:loading="versionDetailLoading"
				@select="handleSelectVersion"
				@rollback="handleRollbackVersion"
			/>
			<div class="grid gap-4 lg:grid-cols-2">
				<UCard>
					<template #header>
						<div class="flex items-center justify-between">
							<div>
								<h3 class="text-lg font-semibold">审批面板</h3>
								<p class="text-sm text-gray-500">查看审批链并提交意见。</p>
							</div>
							<UBadge v-if="versionDetail" variant="soft">版本 #{{ versionDetail.versionNumber }}</UBadge>
						</div>
					</template>
					<div v-if="!versionDetail">
						<p class="text-sm text-gray-500">请选择版本以查看审批信息。</p>
					</div>
					<div v-else class="space-y-4">
						<ul class="space-y-3">
							<li
								v-for="approval in versionDetail.approvals"
								:key="approval.id"
								class="rounded border border-gray-200 px-4 py-3"
							>
								<div class="flex items-center justify-between">
									<p class="font-medium">{{ (approval.role || '').toUpperCase() || 'UNKNOWN' }}</p>
									<UBadge :color="approval.status === 'approved' ? 'emerald' : approval.status === 'rejected' ? 'red' : 'gray'">
										{{ approval.status }}
									</UBadge>
								</div>
								<p class="text-sm text-gray-500">
									SLA：{{ formatDate(approval.slaDueAt) }} · 处理：{{ approval.actedBy || '未处理' }}
								</p>
								<p v-if="approval.comment" class="mt-1 text-sm text-gray-600">备注：{{ approval.comment }}</p>
							</li>
						</ul>
						<UFormField label="审批意见">
							<template #default="{ id }">
								<UTextarea :id="id" v-model="approvalComment" rows="3" placeholder="填写审批意见（可选）" />
							</template>
						</UFormField>
						<div class="flex gap-2">
							<UButton color="primary" :loading="approvalSubmitting" @click="submitApproval('approve')"> 通过 </UButton>
							<UButton color="red" variant="soft" :loading="approvalSubmitting" @click="submitApproval('reject')">
								驳回
							</UButton>
						</div>
					</div>
				</UCard>
				<SpuAuditTimeline :events="auditEvents" />
			</div>
		</div>
		<div v-else>
			<p class="text-gray-500">未找到 SPU</p>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useToast } from '#imports'
import { useRoute } from 'vue-router'
import type { SpuDetail } from '~/composables/api/useSpu'
import SpuSkuLinker from '~/app/pages/product/spus/components/SpuSkuLinker.vue'
import VersionDiff from '~/app/pages/product/spus/components/VersionDiff.vue'
import SpuAuditTimeline from '~/components/product/SpuAuditTimeline.vue'
import { useSpuStore } from '~/stores/product/spu'

const route = useRoute()
const toast = useToast()
const store = useSpuStore()
const spu = ref<SpuDetail | null>(null)
const loading = computed(() => store.detailLoading)
const submitLoading = ref(false)
const publishLoading = ref(false)
const versions = computed(() => store.versions)
const versionDetail = computed(() => store.versionDetail)
const versionDetailLoading = computed(() => store.versionLoading)
const selectedVersionId = ref<string | null>(null)
const approvalComment = ref('')
const approvalSubmitting = ref(false)
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

const load = async () => {
	const currentId = route.params.id as string
	const detail = await store.fetchDetail(currentId)
	spu.value = detail ?? null
	selectedVersionId.value = null
	await store.fetchVersions(currentId)
	if (!selectedVersionId.value && versions.value.length) {
		selectedVersionId.value = versions.value[0].id
		await store.fetchVersionDetail(currentId, selectedVersionId.value)
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
		toast.add({ title: '提交失败', color: 'red' })
	} finally {
		submitLoading.value = false
	}
}

const publishDraft = async () => {
	if (!spu.value?.currentVersionId) {
		toast.add({ title: '无可发布版本', color: 'red' })
		return
	}
	try {
		publishLoading.value = true
		await store.publish(spu.value.id, { versionId: spu.value.currentVersionId, channels: ['official'] })
		toast.add({ title: '发布请求已提交' })
		await load()
	} catch (error) {
		console.error(error)
		toast.add({ title: '发布失败', color: 'red' })
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
		toast.add({ title: '加载版本失败', color: 'red' })
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
		toast.add({ title: '回滚失败', color: 'red' })
	}
}

const submitApproval = async (action: 'approve' | 'reject') => {
	if (!spu.value || !selectedVersionId.value) return
	try {
		approvalSubmitting.value = true
		const payload = approvalComment.value ? { comment: approvalComment.value } : undefined
		if (action === 'approve') {
			await store.approveVersion(spu.value.id, selectedVersionId.value, payload)
			toast.add({ title: '已通过当前版本' })
		} else {
			await store.rejectVersion(spu.value.id, selectedVersionId.value, payload)
			toast.add({ title: '已驳回当前版本', color: 'orange' })
		}
		approvalComment.value = ''
	} catch (error) {
		console.error(error)
		toast.add({ title: '操作失败', color: 'red' })
	} finally {
		approvalSubmitting.value = false
	}
}

const formatDate = (value?: string) => {
	if (!value) return '—'
	return new Date(value).toLocaleString()
}
</script>
