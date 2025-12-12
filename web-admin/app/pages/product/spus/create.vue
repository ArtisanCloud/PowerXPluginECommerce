<template>
	<div class="px-6 py-8 text-white">
		<div class="mx-auto flex max-w-6xl flex-col gap-8">
			<header class="flex flex-wrap items-center gap-3">
				<NuxtLink to="/product/spus" class="text-sm text-primary">← 返回列表</NuxtLink>
				<div class="flex-1">
					<p class="text-xs uppercase tracking-[0.2em] text-primary-200">商品中心</p>
					<h1 class="mt-1 text-3xl font-semibold text-white">创建 SPU</h1>
					<p class="text-sm text-white/60">按照步骤填写基础信息、多语言内容与订阅计划。</p>
				</div>
			</header>

			<UCard class="border border-white/10 bg-white/5 shadow-lg ring-1 ring-white/5 backdrop-blur px-6 py-5">
				<UStepper v-model="stepperIndex" :items="stepperItems" color="primary" />
			</UCard>

			<UCard class="border border-white/10 bg-white/5 shadow-lg ring-1 ring-white/5 backdrop-blur px-6 py-8">
				<div class="space-y-8">
					<component
						:is="currentVisibleStep.component"
						v-model="currentVisibleStep.model"
						v-bind="currentVisibleStep.key === 'plans' ? { spuType: currentType } : {}"
					/>
					<div class="flex items-center justify-between">
						<UButton color="gray" variant="soft" :disabled="currentStep === 0" @click="currentStep--">上一步</UButton>
						<UButton
							color="primary"
							:loading="submitting && isLastStep"
							:data-testid="isLastStep ? 'wizard-submit-button' : 'wizard-next-button'"
							@click="handleNext"
						>
							{{ isLastStep ? '提交草稿' : '下一步' }}
						</UButton>
					</div>
				</div>
			</UCard>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useToast } from '#imports'
import { computed, reactive, ref, markRaw, watch } from 'vue'
import SpuWizardStepBasic from '~/components/product/SpuWizardStepBasic.vue'
import SpuWizardStepLocale from '~/components/product/SpuWizardStepLocale.vue'
import SpuWizardStepPlan from '~/components/product/SpuWizardStepPlan.vue'
import { useSpuStore } from '~/stores/product/spu'

const router = useRouter()
const toast = useToast()
const store = useSpuStore()

const allSteps = reactive([
	{ key: 'basic', title: '基础信息', component: markRaw(SpuWizardStepBasic), model: { type: 'one_time' } },
	{
		key: 'locale',
		title: '多语言内容',
		component: markRaw(SpuWizardStepLocale),
		model: { defaultLocale: 'zh-CN', locales: [{ locale: 'zh-CN', title: '', description: '' }] },
	},
	{ key: 'plans', title: '订阅计划', component: markRaw(SpuWizardStepPlan), model: { plans: [] } },
])
const currentStep = ref(0)
const submitting = ref(false)
const currentType = computed(() => allSteps[0].model.type || 'one_time')
const visibleSteps = computed(() =>
	currentType.value === 'subscription' ? allSteps : allSteps.filter((step) => step.key !== 'plans')
)
const isLastStep = computed(() => currentStep.value === visibleSteps.value.length - 1)
const stepDescriptions: Record<string, string> = {
	basic: '填写基础字段与类目信息',
	locale: '完善多语言文案',
	plans: '配置订阅计划/价格',
}
const stepperItems = computed(() =>
	visibleSteps.value.map((step) => ({
		title: step.title,
		description: stepDescriptions[step.key] ?? '',
	}))
)
const currentVisibleStep = computed(() => visibleSteps.value[currentStep.value] ?? visibleSteps.value[0])

const moveToStep = (targetIdx: number) => {
	if (targetIdx === currentStep.value) return
	if (targetIdx < 0 || targetIdx >= visibleSteps.value.length) return
	if (targetIdx > currentStep.value) {
		for (let idx = 0; idx < targetIdx; idx += 1) {
			if (!validateStep(idx)) {
				return
			}
		}
	}
	currentStep.value = targetIdx
}

const stepperIndex = computed({
	get: () => currentStep.value,
	set: (val: number) => moveToStep(val),
})

watch(
	visibleSteps,
	(stepsList) => {
		if (!stepsList.length) return
		if (currentStep.value >= stepsList.length) {
			currentStep.value = stepsList.length - 1
		}
	},
	{ immediate: true }
)

const showValidationError = (message: string) => {
	toast.add({ title: '表单未完成', description: message, color: 'red' })
}

const validateBasic = () => {
	const { code, name, categoryId, categoryPath } = allSteps[0].model
	const missing: string[] = []
	if (!code) missing.push('SPU 编码')
	if (!name) missing.push('名称')
	if (!categoryId) missing.push('类目 ID')
	if (!categoryPath) missing.push('类目路径')
	if (missing.length) {
		showValidationError(`请填写：${missing.join('、')}`)
		return false
	}
	return true
}

const validateLocale = () => {
	const { defaultLocale, locales = [] } = allSteps[1].model
	if (!defaultLocale) {
		showValidationError('请选择默认语言')
		return false
	}
	const defaultEntry = locales.find((entry: any) => entry.locale === defaultLocale)
	if (!defaultEntry || !defaultEntry.title) {
		showValidationError('默认语言需填写标题')
		return false
	}
	return true
}

const validatePlans = () => {
	if (currentType.value !== 'subscription') {
		return true
	}
	const plans = allSteps[2].model.plans || []
	if (!plans.length) {
		showValidationError('订阅 SPU 至少配置一个订阅计划')
		return false
	}
	for (let i = 0; i < plans.length; i += 1) {
		const plan = plans[i]
		if (!plan.planCode || !plan.name || !plan.billingCycle) {
			showValidationError(`订阅计划 ${i + 1} 信息不完整`)
			return false
		}
		if (!plan.price || plan.price <= 0) {
			showValidationError(`订阅计划 ${i + 1} 价格需大于 0`)
			return false
		}
	}
	return true
}

const validateStep = (index: number) => {
	const key = visibleSteps.value[index]?.key
	switch (key) {
		case 'basic':
			return validateBasic()
		case 'locale':
			return validateLocale()
		case 'plans':
			return validatePlans()
		default:
			return true
	}
}

const handleNext = async () => {
if (!validateStep(currentStep.value)) {
		return
	}
	if (!isLastStep.value) {
		moveToStep(currentStep.value + 1)
		return
	}
	if (!validateBasic() || !validateLocale() || !validatePlans()) {
		return
	}
	const payload = {
		...allSteps[0].model,
		...allSteps[1].model,
		subscriptionPlans: allSteps[2].model.plans,
		tags: allSteps[0].model.tags || [],
	}
	try {
		submitting.value = true
		const detail = await store.create(payload)
		toast.add({ title: '草稿已保存', description: '可前往详情提交审批' })
		const target = detail?.id ? `/product/spus/${detail.id}` : '/product/spus'
		router.push(target)
	} catch (error) {
		console.error(error)
		toast.add({ title: '创建失败', description: '请稍后重试', color: 'red' })
	} finally {
		submitting.value = false
	}
}
</script>
