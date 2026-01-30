<template>
	<UCard>
		<template #header>
			<div class="flex items-center justify-between gap-3">
				<div>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">订阅计划</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">适用于订阅型商品，可配置计费周期和生效范围。</p>
				</div>
				<UButton
					size="sm"
					variant="soft"
					icon="i-heroicons-plus"
					:disabled="!isSubscription"
					@click="startCreate"
					:title="isSubscription ? '新增订阅计划' : '仅订阅型 SPU 可配置计划'"
				>
					新计划
				</UButton>
			</div>
		</template>
		<div class="space-y-6">
			<div v-if="isSubscription" class="grid gap-4 md:grid-cols-2">
				<UFormField label="计划编码" :ui="fieldUi">
					<UInput v-model="form.planCode" :disabled="isEditingExisting" placeholder="如 monthly" />
				</UFormField>
		<UFormField label="计划名称" :ui="fieldUi">
			<UInput v-model="form.name" placeholder="如 月度订阅" />
		</UFormField>
		<UFormField label="会籍等级 ID" :ui="fieldUi">
			<UInput v-model="membershipTierId" placeholder="绑定 membership tier ID" />
		</UFormField>
				<UFormField label="SKU ID" :ui="fieldUi">
					<UInput v-model="skuId" placeholder="绑定订阅 SKU（UUID）" />
				</UFormField>
				<UFormField label="权益 IDs" :ui="fieldUi">
					<UInput v-model="benefitIdsInput" placeholder="多个以逗号分隔" />
				</UFormField>
				<UFormField label="计费周期" :ui="fieldUi">
					<USelect v-model="form.billingCycle" :items="billingOptions" class="w-full" />
				</UFormField>
				<UFormField v-if="form.billingCycle === 'custom'" label="自定义周期（天）" :ui="fieldUi">
					<UInput v-model.number="form.billingValue" type="number" min="1" placeholder="30" />
				</UFormField>
				<UFormField label="价格" :ui="fieldUi">
					<UInput v-model.number="form.price" type="number" min="0" step="0.01" class="w-full" />
				</UFormField>
				<UFormField label="币种" :ui="fieldUi">
					<UInput v-model="form.currency" maxlength="3" placeholder="CNY" class="w-full" />
				</UFormField>
				<UFormField label="Token Code" :ui="fieldUi">
					<UInput v-model="tokenCode" placeholder="service_credit" />
				</UFormField>
				<UFormField label="Token 数量" :ui="fieldUi">
					<UInput v-model.number="tokenAmount" type="number" min="0" placeholder="0" />
				</UFormField>
				<UFormField label="Token 有效期(天)" :ui="fieldUi">
					<UInput v-model.number="tokenExpireDays" type="number" min="0" placeholder="0" />
				</UFormField>
				<UFormField label="Token 结转" :ui="fieldUi">
					<USwitch v-model="tokenRollover" />
				</UFormField>
				<UFormField label="试用天数" :ui="fieldUi">
					<UInput v-model.number="form.trialDays" type="number" min="0" placeholder="0" />
				</UFormField>
				<UFormField label="自动续费" :ui="fieldUi">
					<USwitch v-model="form.autoRenew" />
				</UFormField>
				<UFormField label="取消策略" :ui="fieldUi">
					<USelect v-model="form.cancelPolicy" :items="cancelPolicyOptions" class="w-full" />
				</UFormField>
				<UFormField label="作用范围" :ui="fieldUi">
					<USelect v-model="form.effectScope" :items="effectScopeOptions" class="w-full" />
				</UFormField>
			</div>
			<div v-else class="text-sm text-gray-500 dark:text-gray-400">
				当前 SPU 类型非订阅，无需配置订阅计划。
			</div>
			<div class="flex flex-wrap justify-end gap-2" v-if="isSubscription">
				<UButton variant="ghost" @click="resetForm">取消</UButton>
				<UButton color="primary" :loading="saving" @click="handleSave">
					{{ isEditingExisting ? '更新计划' : '新增计划' }}
				</UButton>
			</div>
			<div>
				<h4 class="text-base font-semibold mb-2 text-gray-900 dark:text-white">计划列表</h4>
				<UTable :data="pagedPlans" :columns="columns" :loading="loading">
					<template #skuId-cell="{ row }">
						<span class="text-xs font-mono text-gray-600 dark:text-gray-300">
							{{ formatSkuId(row.original) }}
						</span>
					</template>
					<template #membershipTierId-cell="{ row }">
						<span class="text-xs font-mono text-gray-600 dark:text-gray-300">
							{{ formatMembershipTierId(row.original) }}
						</span>
					</template>
					<template #benefitIds-cell="{ row }">
						<span class="text-xs text-gray-500 dark:text-gray-400">
							{{ formatBenefitIds(row.original) }}
						</span>
					</template>
					<template #billingCycle-cell="{ row }">
						{{ billingLabel(row.original.billingCycle, row.original.billingValue) }}
					</template>
					<template #effectScope-cell="{ row }">
						<UBadge :color="row.original.effectScope === 'new_and_existing' ? 'primary' : 'gray'">
							{{ effectScopeLabel(row.original.effectScope) }}
						</UBadge>
					</template>
					<template #actions-cell="{ row }">
						<div class="flex gap-2">
							<UButton size="xs" variant="soft" @click="handleEdit(row.original)">编辑</UButton>
							<UButton
								size="xs"
								color="red"
								variant="soft"
								@click="handleDelete(row.original)"
								:loading="removing === row.original.id"
							>
								禁用
							</UButton>
						</div>
					</template>
				</UTable>
				<div v-if="plans.length" class="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm">
					<span class="text-gray-500 dark:text-gray-400">
						共 {{ plans.length }} 条 · 第 {{ planPagination.page }} / {{ planPageCount }} 页
					</span>
					<div class="flex items-center gap-3">
						<USelect v-model="planPagination.pageSize" :items="planPageSizeOptions" class="w-28" />
						<UPagination v-model="planPagination.page" :page-count="planPageCount" />
					</div>
				</div>
				<UAlert v-if="!plans.length && !loading" color="gray" class="mt-3">暂无订阅计划。</UAlert>
			</div>
		</div>
	</UCard>
</template>

<script setup lang="ts">
import { useToast } from '#imports'
import type { TableColumn } from '@nuxt/ui'
import type { SpuSubscriptionPlan, SubscriptionPlanPayload } from '~/composables/api/useSpu'
import { useSpuApi } from '~/composables/api/useSpu'

const props = defineProps<{ spuId?: string; spuType?: string }>()

const toast = useToast()
const api = useSpuApi()
const plans = ref<SpuSubscriptionPlan[]>([])
const loading = ref(false)
const saving = ref(false)
const removing = ref<string | null>(null)
const planPagination = reactive({ page: 1, pageSize: 5 })
const planPageSizeOptions = [
	{ label: '5 / 页', value: 5 },
	{ label: '10 / 页', value: 10 },
	{ label: '20 / 页', value: 20 },
]
const form = reactive<SubscriptionPlanPayload>({
	planCode: '',
	name: '',
	billingCycle: 'monthly',
	billingValue: undefined,
	price: 0,
	currency: 'CNY',
	trialDays: 0,
	autoRenew: true,
	cancelPolicy: 'anytime',
	effectScope: 'new_only',
	metadata: {},
})
const skuId = ref('')
const membershipTierId = ref('')
const benefitIdsInput = ref('')
const tokenCode = ref('')
const tokenAmount = ref<number | null>(null)
const tokenExpireDays = ref<number | null>(null)
const tokenRollover = ref(false)
const editingPlanId = ref<string | null>(null)
const isSubscription = computed(() => (props.spuType || '').toLowerCase() === 'subscription')
const isEditingExisting = computed(() => Boolean(editingPlanId.value))

const fieldUi = {
	root: 'flex flex-col gap-2',
	label: 'text-sm font-medium text-gray-500 dark:text-gray-400',
	container: 'mt-0 w-full',
}
const billingOptions = [
	{ label: '月付', value: 'monthly' },
	{ label: '季付', value: 'quarterly' },
	{ label: '年付', value: 'yearly' },
	{ label: '自定义', value: 'custom' },
]
const cancelPolicyOptions = [
	{ label: '随时取消', value: 'anytime' },
	{ label: '需提前通知', value: 'notice_required' },
	{ label: '锁定期', value: 'locked_in' },
]
const effectScopeOptions = [
	{ label: '仅新订阅', value: 'new_only' },
	{ label: '新+存量订阅', value: 'new_and_existing' },
]
const columns = computed<TableColumn<SpuSubscriptionPlan>[]>(() => [
	{ accessorKey: 'planCode', header: '计划编码' },
	{ accessorKey: 'name', header: '名称' },
	{ id: 'membershipTierId', header: '会籍等级' },
	{ id: 'skuId', header: 'SKU ID' },
	{ id: 'benefitIds', header: '权益' },
	{ accessorKey: 'billingCycle', header: '计费周期' },
	{ accessorKey: 'price', header: '价格' },
	{ accessorKey: 'effectScope', header: '作用范围' },
	{ id: 'actions', header: '操作' },
])
const pagedPlans = computed(() => {
	const start = (planPagination.page - 1) * planPagination.pageSize
	return plans.value.slice(start, start + planPagination.pageSize)
})
const planPageCount = computed(() => {
	const total = Math.ceil(plans.value.length / planPagination.pageSize)
	return total > 0 ? total : 1
})
const billingLabel = (cycle: string, value?: number) => {
	const map: Record<string, string> = { monthly: '月付', quarterly: '季付', yearly: '年付', custom: '自定义' }
	if (cycle === 'custom' && value) {
		return `自定义 (${value}天)`
	}
	return map[cycle] || cycle
}
const effectScopeLabel = (value: string) => (value === 'new_and_existing' ? '新+存量' : '仅新订阅')
const resolveSkuId = (plan: SpuSubscriptionPlan) => {
	const direct = String(plan.skuId || '').trim()
	if (direct) return direct
	const meta = (plan.metadata || {}) as Record<string, any>
	return String(meta.skuId || meta.sku_id || meta.skuID || '').trim()
}
const formatSkuId = (plan: SpuSubscriptionPlan) => resolveSkuId(plan) || '-'
const resolveMembershipTierId = (plan: SpuSubscriptionPlan) => {
	const meta = (plan.metadata || {}) as Record<string, any>
	return String(meta.membershipTierId || meta.membership_tier_id || meta.membershipTierID || '').trim()
}
const formatMembershipTierId = (plan: SpuSubscriptionPlan) => resolveMembershipTierId(plan) || '-'
const resolveBenefitIds = (plan: SpuSubscriptionPlan) => {
	const meta = (plan.metadata || {}) as Record<string, any>
	const raw = meta.benefitIds || meta.benefit_ids
	if (Array.isArray(raw)) return raw.map((x: any) => String(x).trim()).filter(Boolean)
	if (typeof raw === 'string') return raw.split(',').map((x) => x.trim()).filter(Boolean)
	return [] as string[]
}
const formatBenefitIds = (plan: SpuSubscriptionPlan) => {
	const ids = resolveBenefitIds(plan)
	return ids.length ? ids.join(', ') : '-'
}

const resetForm = () => {
	Object.assign(form, {
		planCode: '',
		name: '',
		billingCycle: 'monthly',
		billingValue: undefined,
		price: 0,
		currency: 'CNY',
		trialDays: 0,
		autoRenew: true,
		cancelPolicy: 'anytime',
		effectScope: 'new_only',
		metadata: {},
	})
	skuId.value = ''
	membershipTierId.value = ''
	benefitIdsInput.value = ''
	tokenCode.value = ''
	tokenAmount.value = null
	tokenExpireDays.value = null
	tokenRollover.value = false
	editingPlanId.value = null
}

const startCreate = () => {
	if (!isSubscription.value) {
		toast.add({ title: '当前 SPU 非订阅类型，无法创建订阅计划', color: 'orange' })
		return
	}
	resetForm()
}

const handleEdit = (plan: SpuSubscriptionPlan) => {
	editingPlanId.value = plan.id
	skuId.value = resolveSkuId(plan)
	membershipTierId.value = resolveMembershipTierId(plan)
	benefitIdsInput.value = resolveBenefitIds(plan).join(', ')
	const meta = (plan.metadata || {}) as Record<string, any>
	tokenCode.value = String(meta.tokenCode || meta.token_code || '').trim()
	const tokenAmountRaw = meta.tokenAmount ?? meta.token_amount
	const tokenExpireRaw = meta.tokenExpireDays ?? meta.token_expire_days
	tokenAmount.value = Number.isFinite(Number(tokenAmountRaw)) ? Number(tokenAmountRaw) : null
	tokenExpireDays.value = Number.isFinite(Number(tokenExpireRaw)) ? Number(tokenExpireRaw) : null
	tokenRollover.value = Boolean(meta.tokenRollover ?? meta.token_rollover)
	Object.assign(form, {
		planCode: plan.planCode,
		name: plan.name,
		billingCycle: plan.billingCycle,
		billingValue: plan.billingValue,
		price: plan.price,
		currency: plan.currency,
		trialDays: plan.trialDays,
		autoRenew: plan.autoRenew,
		cancelPolicy: plan.cancelPolicy,
		effectScope: plan.effectScope,
		metadata: plan.metadata || {},
	})
}

const handleSave = async () => {
	if (!props.spuId || !isSubscription.value) {
		toast.add({ title: '当前 SPU 非订阅类型', color: 'red' })
		return
	}
	try {
		saving.value = true
		const benefitIds = benefitIdsInput.value
			.split(',')
			.map((v) => v.trim())
			.filter(Boolean)
		const payload: SubscriptionPlanPayload = {
			...form,
			metadata: {
				...(form.metadata || {}),
				skuId: skuId.value || undefined,
				membershipTierId: membershipTierId.value || undefined,
				benefitIds: benefitIds.length ? benefitIds : undefined,
				tokenCode: tokenCode.value || undefined,
				tokenAmount: tokenAmount.value ?? undefined,
				tokenExpireDays: tokenExpireDays.value ?? undefined,
				tokenRollover: tokenRollover.value || undefined,
			},
		}
		if (editingPlanId.value) {
			await api.updateSubscriptionPlan(props.spuId, editingPlanId.value, payload)
			toast.add({ title: '计划已更新' })
		} else {
			if (!form.planCode) {
				toast.add({ title: '请填写计划编码', color: 'red' })
				return
			}
			await api.createSubscriptionPlan(props.spuId, payload)
			toast.add({ title: '计划已创建' })
		}
		await loadPlans()
		resetForm()
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '保存失败', color: 'red' })
	} finally {
		saving.value = false
	}
}

const handleDelete = async (plan: SpuSubscriptionPlan) => {
	if (!props.spuId) return
	try {
		removing.value = plan.id
		await api.deleteSubscriptionPlan(props.spuId, plan.id)
		toast.add({ title: '计划已禁用', description: plan.name })
		await loadPlans()
		if (editingPlanId.value === plan.id) {
			resetForm()
		}
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '操作失败', color: 'red' })
	} finally {
		removing.value = null
	}
}

const loadPlans = async () => {
	if (!props.spuId || !isSubscription.value) {
		plans.value = []
		return
	}
	try {
		loading.value = true
		const { items } = await api.listSubscriptionPlans(props.spuId)
		plans.value = (items ?? []).filter((plan) => plan.status !== 'archived')
	} catch (error: any) {
		console.error(error)
		toast.add({ title: error?.message || '加载订阅计划失败', color: 'red' })
	} finally {
		loading.value = false
	}
}

watch(
	() => [props.spuId, props.spuType],
	() => {
		loadPlans()
		resetForm()
	},
	{ immediate: true }
)

watch(
	() => plans.value.length,
	() => {
		planPagination.page = 1
	},
	{ flush: 'post' }
)

watch(
	() => planPagination.pageSize,
	() => {
		planPagination.page = 1
	}
)

watch(planPageCount, (count) => {
	if (planPagination.page > count) {
		planPagination.page = count
	}
})
</script>
