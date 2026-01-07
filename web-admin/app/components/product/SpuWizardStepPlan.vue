<template>
	<div class="space-y-4">
		<UAlert v-if="!isSubscription" color="gray" variant="soft">
			<template #title>订阅计划可选</template>
			当前 SPU 类型非订阅，可跳过本步骤，若后续切换为订阅请返回填写计划。
		</UAlert>
		<div v-for="(plan, idx) in localValue.plans" :key="idx" class="space-y-4 rounded border p-4">
			<div class="flex items-center justify-between">
				<strong>订阅计划 {{ idx + 1 }}</strong>
				<UButton
					color="gray"
					variant="ghost"
					size="xs"
					icon="i-heroicons-trash"
					@click="remove(idx)"
					:disabled="!isSubscription || localValue.plans.length === 1"
				>
					移除
				</UButton>
			</div>
			<div class="grid gap-4 md:grid-cols-2">
				<UFormField label="计划编码">
					<template #default="{ id }">
						<UInput :id="id" v-model.trim="plan.planCode" placeholder="monthly" :disabled="!isSubscription" />
					</template>
				</UFormField>
				<UFormField label="计划名称">
					<template #default="{ id }">
						<UInput :id="id" v-model.trim="plan.name" placeholder="月付订阅" :disabled="!isSubscription" />
					</template>
				</UFormField>
				<UFormField label="计费周期">
					<template #default="{ id }">
						<USelect :id="id" v-model="plan.billingCycle" :items="billingOptions" :disabled="!isSubscription" />
					</template>
				</UFormField>
				<UFormField label="价格 (含税)">
					<template #default="{ id }">
						<UInput :id="id" v-model.number="plan.price" type="number" min="0" step="0.01" :disabled="!isSubscription" />
					</template>
				</UFormField>
				<UFormField label="币种">
					<template #default="{ id }">
						<UInput :id="id" v-model.trim="plan.currency" placeholder="CNY" maxlength="3" :disabled="!isSubscription" />
					</template>
				</UFormField>
				<UFormField label="试用天数">
					<template #default="{ id }">
						<UInput :id="id" v-model.number="plan.trialDays" type="number" min="0" :disabled="!isSubscription" />
					</template>
				</UFormField>
			</div>
			<div class="grid gap-4 md:grid-cols-2">
				<UFormField label="生效范围">
					<template #default="{ id }">
						<USelect :id="id" v-model="plan.effectScope" :items="effectScopeOptions" :disabled="!isSubscription" />
					</template>
				</UFormField>
				<UFormField label="取消策略">
					<template #default="{ id }">
						<USelect :id="id" v-model="plan.cancelPolicy" :items="cancelPolicyOptions" :disabled="!isSubscription" />
					</template>
				</UFormField>
			</div>
			<UCheckbox v-model="plan.autoRenew" label="支持自动续费" :disabled="!isSubscription" />
		</div>
		<UButton color="primary" variant="ghost" icon="i-heroicons-plus-circle" @click="append" :disabled="!isSubscription">
			添加订阅计划
		</UButton>
	</div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, toRaw, watch } from 'vue'

type Plan = {
	planCode: string
	name: string
	billingCycle: string
	price: number
	currency: string
	trialDays: number
	autoRenew: boolean
	effectScope: string
	cancelPolicy: string
}

const props = defineProps<{
	modelValue: { plans: Plan[] }
	spuType?: string
}>()
const emit = defineEmits<{
	'update:modelValue': [{ plans: Plan[] }]
}>()

const billingOptions = [
	{ label: '月付', value: 'monthly' },
	{ label: '季付', value: 'quarterly' },
	{ label: '年付', value: 'yearly' },
	{ label: '自定义', value: 'custom' },
]
const effectScopeOptions = [
	{ label: '仅新订阅', value: 'new_only' },
	{ label: '新+存量', value: 'new_and_existing' },
]
const cancelPolicyOptions = [
	{ label: '随时取消', value: 'anytime' },
	{ label: '需提前通知', value: 'notice_required' },
	{ label: '锁定期', value: 'locked_in' },
]

const localValue = reactive<{ plans: Plan[] }>({ plans: [] })
const isSubscription = computed(() => props.spuType === 'subscription')
const syncingFromProps = ref(false)
const lastEmittedKey = ref('')

const toPlainPlan = (input: Partial<Plan> | undefined | null): Plan => {
	const raw = (input ? toRaw(input as any) : {}) as Partial<Plan>
	return {
		...createPlan(),
		...raw,
		planCode: String(raw.planCode ?? '').trim(),
		name: String(raw.name ?? '').trim(),
		billingCycle: String(raw.billingCycle ?? createPlan().billingCycle),
		currency: String(raw.currency ?? createPlan().currency).toUpperCase(),
		price: typeof raw.price === 'number' ? raw.price : Number(raw.price ?? 0),
		trialDays: typeof raw.trialDays === 'number' ? raw.trialDays : Number(raw.trialDays ?? 0),
		autoRenew: typeof raw.autoRenew === 'boolean' ? raw.autoRenew : createPlan().autoRenew,
		effectScope: String(raw.effectScope ?? createPlan().effectScope),
		cancelPolicy: String(raw.cancelPolicy ?? createPlan().cancelPolicy),
	}
}

const clonePlans = (plans: Plan[] | undefined | null) => (plans ?? []).map((plan) => toPlainPlan(plan))

const plansKey = (plans: Plan[] | undefined | null) => {
	const items = (plans ?? []).map((plan) => toPlainPlan(plan))
	return JSON.stringify(items)
}

const ensurePlans = () => {
	if (!localValue.plans.length) {
		localValue.plans.push(createPlan())
	}
}

watch(
	() => props.modelValue?.plans,
	(plans) => {
		const nextKey = plansKey(plans ?? [])
		const currentKey = plansKey(localValue.plans)
		if (nextKey === currentKey) return
		syncingFromProps.value = true
		localValue.plans = clonePlans(plans)
		if (isSubscription.value) {
			ensurePlans()
		}
		lastEmittedKey.value = plansKey(localValue.plans)
		syncingFromProps.value = false
	},
	{ immediate: true }
)

watch(
	() => isSubscription.value,
	(active) => {
		if (active) {
			syncingFromProps.value = true
			ensurePlans()
			lastEmittedKey.value = plansKey(localValue.plans)
			syncingFromProps.value = false
		}
	}
)

watch(
	localValue,
	() => {
		if (syncingFromProps.value) return
		const key = plansKey(localValue.plans)
		if (key === lastEmittedKey.value) return
		lastEmittedKey.value = key
		emit('update:modelValue', { plans: localValue.plans.map((plan) => toPlainPlan(plan)) })
	},
	{ deep: true }
)

const append = () => {
	localValue.plans.push(createPlan())
}

const remove = (index: number) => {
	if (localValue.plans.length === 1) return
	localValue.plans.splice(index, 1)
}

function createPlan(): Plan {
	return {
		planCode: '',
		name: '',
		billingCycle: 'monthly',
		price: 0,
		currency: 'CNY',
		trialDays: 0,
		autoRenew: true,
		effectScope: 'new_only',
		cancelPolicy: 'anytime',
	}
}
</script>
