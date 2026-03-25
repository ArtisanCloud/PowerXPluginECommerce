<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between gap-3">
      <div class="flex items-center gap-3">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-left" @click="goBack">返回</UButton>
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">成长值账户详情</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400">查看并调整客户成长值余额</p>
        </div>
      </div>
      <div class="flex gap-2">
        <UButton color="primary" icon="i-heroicons-plus-circle" @click="openAdjust('add')">赠送成长值</UButton>
        <UButton color="warning" icon="i-heroicons-minus-circle" @click="openAdjust('deduct')">扣减成长值</UButton>
      </div>
    </div>

    <div class="border-b border-gray-200 dark:border-gray-800">
      <nav class="-mb-px flex space-x-8">
        <UButton :to="`/customer/membership/growth-value/accounts/${accountId}`" variant="ghost" color="neutral">账户信息</UButton>
        <UButton :to="`/customer/membership/growth-value/accounts/${accountId}/details`" variant="ghost" color="neutral">成长值明细</UButton>
      </nav>
    </div>

    <UAlert color="warning" variant="soft" title="流水接口暂未开放" description="当前仅展示实时余额与等级进度；成长值明细流水需后端补充查询接口。" />

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <UCard class="lg:col-span-2" :ui="{ body: 'space-y-4' }">
        <template #header>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">客户信息</h2>
        </template>
        <div class="flex items-center gap-3">
          <UAvatar :src="account.avatar" :alt="account.customerName" size="lg" :ui="{ rounded: 'rounded-full' }" />
          <div>
            <div class="text-lg font-semibold text-gray-900 dark:text-white">{{ account.customerName }}</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">{{ account.customerPhone || '-' }}</div>
          </div>
        </div>
        <div class="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
          <div>
            <div class="text-gray-500 dark:text-gray-400">客户ID</div>
            <div class="text-gray-900 dark:text-white">{{ account.customerId || '-' }}</div>
          </div>
          <div>
            <div class="text-gray-500 dark:text-gray-400">注册时间</div>
            <div class="text-gray-900 dark:text-white">{{ account.registeredAt || '-' }}</div>
          </div>
          <div>
            <div class="text-gray-500 dark:text-gray-400">当前等级</div>
            <div class="text-gray-900 dark:text-white">{{ account.levelName }}</div>
          </div>
          <div>
            <div class="text-gray-500 dark:text-gray-400">等级有效期</div>
            <div class="text-gray-900 dark:text-white">{{ account.levelExpiry }}</div>
          </div>
        </div>
      </UCard>

      <div class="space-y-6">
        <UCard>
          <div class="text-center">
            <div class="text-sm text-gray-500 dark:text-gray-400">当前成长值余额</div>
            <div class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{{ account.balance.toLocaleString() }}</div>
            <div class="mt-2"><UBadge color="primary" variant="soft">{{ account.levelName }}</UBadge></div>
          </div>
        </UCard>

        <UCard>
          <template #header>
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">等级晋升进度</h3>
          </template>
          <div v-if="nextLevel" class="space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">晋升到 {{ nextLevel.name }}</span>
              <span class="text-gray-900 dark:text-white">{{ account.balance }} / {{ nextLevel.threshold }}</span>
            </div>
            <UProgress :value="Math.min(100, (account.balance / Math.max(1, nextLevel.threshold)) * 100)" size="sm" />
            <div class="text-xs text-gray-500 dark:text-gray-400">还需 {{ Math.max(0, nextLevel.threshold - account.balance) }} 成长值</div>
          </div>
          <div v-else class="text-sm text-gray-500 dark:text-gray-400">当前已是最高等级或未配置阈值</div>
        </UCard>
      </div>
    </div>

    <UCard>
      <template #header>
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">最近成长值记录</h2>
      </template>
      <UAlert color="gray">暂无可查询的成长值流水记录</UAlert>
    </UCard>

    <UModal v-model:open="showAdjustModal">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ adjustType === 'add' ? '赠送成长值' : '扣减成长值' }}</h3>
          </template>
          <form class="space-y-4" @submit.prevent="saveAdjustment">
            <UFormField label="调整数量" required>
              <UInput v-model.number="adjustForm.growthValue" type="number" min="1" />
            </UFormField>
            <UFormField label="调整原因" required>
              <UTextarea v-model="adjustForm.reason" :rows="3" />
            </UFormField>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" type="button" @click="showAdjustModal = false">取消</UButton>
              <UButton color="primary" type="submit" :loading="adjustLoading" :disabled="!isAdjustFormValid">确认</UButton>
            </div>
          </form>
        </UCard>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCustomerApi } from '~/composables/api/useCustomer'
import { useMembershipAdminApi } from '~/composables/api/useMembership'
import { useToast } from '#imports'
import type { MembershipTier } from '~/types/membership'

const route = useRoute()
const router = useRouter()
const customerApi = useCustomerApi()
const membershipApi = useMembershipAdminApi()
const toast = useToast()

const accountId = String(route.params.id || '')

const account = ref({
  customerId: accountId,
  customerName: '未知客户',
  customerPhone: '-',
  avatar: '',
  balance: 0,
  levelName: '未分层',
  registeredAt: '-',
  levelExpiry: '-',
})

const thresholds = ref<Array<{ name: string; threshold: number }>>([])

const nextLevel = computed(() => thresholds.value.find((item) => item.threshold > account.value.balance) || null)

const showAdjustModal = ref(false)
const adjustType = ref<'add' | 'deduct'>('add')
const adjustLoading = ref(false)
const adjustForm = ref({ growthValue: 0, reason: '' })
const isAdjustFormValid = computed(() => adjustForm.value.growthValue > 0 && adjustForm.value.reason.trim().length > 0)

const getThresholdFromTier = (tier: MembershipTier): number => {
  const rules = tier.rules && typeof tier.rules === 'object' ? (tier.rules as Record<string, any>) : {}
  const upgrade = rules.upgrade && typeof rules.upgrade === 'object' ? rules.upgrade : {}
  const value = Number(upgrade.minGrowthValue ?? rules.growthValueThreshold ?? rules.threshold ?? 0)
  return Number.isFinite(value) ? value : 0
}

const loadData = async () => {
  try {
    const [customer, tokenBalances, tierResp] = await Promise.all([
      customerApi.getCustomer(accountId),
      customerApi.getCustomerTokenBalances(accountId),
      membershipApi.listTiers(),
    ])

    const snapshot = customer?.membershipSnapshot || {}
    const growthToken = (tokenBalances?.items || []).find((item: any) => item.tokenCode === 'growth_value')
    const balance = Number(growthToken?.balance ?? snapshot?.growthValue ?? customer?.growthValue ?? 0) || 0

    account.value = {
      customerId: customer?.id || accountId,
      customerName: customer?.name || '未知客户',
      customerPhone: customer?.phone || '-',
      avatar: `https://api.dicebear.com/7.x/miniavs/svg?seed=${encodeURIComponent(customer?.id || accountId)}`,
      balance,
      levelName: snapshot?.tier || customer?.membershipTierLabel || customer?.membershipTier || '未分层',
      registeredAt: customer?.createdAt || '-',
      levelExpiry: '-',
    }

    thresholds.value = (tierResp?.items || [])
      .map((tier: MembershipTier) => ({ name: tier.name, threshold: getThresholdFromTier(tier) }))
      .filter((item) => item.threshold > 0)
      .sort((a, b) => a.threshold - b.threshold)
  } catch (error: any) {
    toast.add({ title: '加载账户详情失败', description: error?.message || '请稍后重试', color: 'error' })
  }
}

const goBack = () => router.push('/customer/membership/growth-value')

const openAdjust = (type: 'add' | 'deduct') => {
  adjustType.value = type
  adjustForm.value = { growthValue: 0, reason: '' }
  showAdjustModal.value = true
}

const saveAdjustment = async () => {
  if (!isAdjustFormValid.value) return
  try {
    adjustLoading.value = true
    const delta = adjustType.value === 'add' ? Math.abs(adjustForm.value.growthValue) : -Math.abs(adjustForm.value.growthValue)
    await membershipApi.adjustToken({
      customerId: account.value.customerId,
      tokenCode: 'growth_value',
      delta,
      reason: adjustForm.value.reason.trim(),
    })
    showAdjustModal.value = false
    await loadData()
    toast.add({ title: '调整成功', color: 'success' })
  } catch (error: any) {
    toast.add({ title: '调整失败', description: error?.message || '请稍后重试', color: 'error' })
  } finally {
    adjustLoading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
