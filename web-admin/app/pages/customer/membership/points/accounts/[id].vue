<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between gap-3">
      <div class="flex items-center gap-3">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-left" @click="goBack">返回</UButton>
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">积分明细</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400">展示积分历史调整记录</p>
        </div>
      </div>
      <UButton color="neutral" variant="outline" icon="i-heroicons-arrow-down-tray" @click="exportData">导出</UButton>
    </div>

    <UCard>
      <div class="flex items-center gap-3">
        <UAvatar :src="account.avatar" :alt="account.customerName" size="lg" :ui="{ rounded: 'rounded-full' }" />
        <div>
          <div class="text-lg font-semibold text-gray-900 dark:text-white">{{ account.customerName }}</div>
          <div class="text-sm text-gray-500 dark:text-gray-400">{{ account.customerPhone }}</div>
        </div>
        <div class="ml-auto text-right">
          <div class="text-sm text-gray-500 dark:text-gray-400">当前积分余额</div>
          <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ account.balance.toLocaleString() }}</div>
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">积分记录</h2>
      </template>

      <UTable :columns="columns" :data="records" :loading="loading">
        <template #delta-cell="{ row }">
          <span :class="row.delta >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'">
            {{ row.delta >= 0 ? `+${row.delta}` : row.delta }}
          </span>
        </template>

        <template #sourceType-cell="{ row }">
          <UBadge color="neutral" variant="soft">{{ row.sourceType || '-' }}</UBadge>
        </template>

        <template #createdAt-cell="{ row }">
          <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDate(row.createdAt) }}</span>
        </template>
      </UTable>

      <UAlert v-if="!loading && !records.length" color="gray" class="mt-4">暂无积分记录</UAlert>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import { useRoute, useRouter } from 'vue-router'
import { useCustomerApi } from '~/composables/api/useCustomer'
import { useMembershipAdminApi } from '~/composables/api/useMembership'
import type { MembershipTokenTransaction } from '~/types/membership'
import { useToast } from '#imports'

const route = useRoute()
const router = useRouter()
const customerApi = useCustomerApi()
const membershipApi = useMembershipAdminApi()
const toast = useToast()

const accountId = String(route.params.id || '')

const account = ref({
  customerName: '未知客户',
  customerPhone: '-',
  avatar: '',
  balance: 0,
})
const loading = ref(false)
const records = ref<MembershipTokenTransaction[]>([])

const columns = computed<TableColumn<MembershipTokenTransaction>[]>(() => [
  { accessorKey: 'delta', header: '变动值' },
  { accessorKey: 'sourceType', header: '来源' },
  { accessorKey: 'sourceId', header: '来源ID' },
  { accessorKey: 'createdAt', header: '时间' },
])

const formatDate = (value?: string) => {
  if (!value) return '-'
  const ts = new Date(value)
  if (Number.isNaN(ts.getTime())) return value
  return ts.toLocaleString('zh-CN', { hour12: false })
}

const loadData = async () => {
  loading.value = true
  try {
    const [customer, tokenBalances, txnResp] = await Promise.all([
      customerApi.getCustomer(accountId),
      customerApi.getCustomerTokenBalances(accountId),
      membershipApi.listTokenTransactions({ customerId: accountId, tokenCode: 'points', page: 1, pageSize: 100 }),
    ])
    const snapshot = customer?.membershipSnapshot || {}
    const pointsToken = (tokenBalances?.items || []).find((item: any) => item.tokenCode === 'points')
    account.value = {
      customerName: customer?.name || '未知客户',
      customerPhone: customer?.phone || '-',
      avatar: `https://api.dicebear.com/7.x/miniavs/svg?seed=${encodeURIComponent(customer?.id || accountId)}`,
      balance: Number(pointsToken?.balance ?? snapshot?.points ?? customer?.points ?? 0) || 0,
    }
    records.value = txnResp?.items || []
  } catch (error: any) {
    toast.add({ title: '加载失败', description: error?.message || '请稍后重试', color: 'error' })
  } finally {
    loading.value = false
  }
}

const goBack = () => router.push('/customer/membership/points/accounts')

const exportData = () => {
  const lines = ['delta,sourceType,sourceId,createdAt']
  for (const item of records.value) {
    lines.push(`${item.delta},${item.sourceType || ''},${item.sourceId || ''},${item.createdAt || ''}`)
  }
  const blob = new Blob(["\uFEFF" + lines.join('\n')], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `points-transactions-${accountId}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(() => {
  loadData()
})
</script>
