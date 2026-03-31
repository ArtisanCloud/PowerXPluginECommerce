<template>
  <div class="space-y-4 px-4 py-6 sm:px-6">
    <UCard>
      <template #header>
        <div class="space-y-1">
          <h1 class="text-xl font-semibold">售后服务</h1>
          <p class="text-sm text-gray-500">提交退款/退货退款/换货申请并查看进度。</p>
        </div>
      </template>

      <div class="grid gap-3 sm:grid-cols-2">
        <UFormField label="订单号" required>
          <UInput v-model.trim="form.orderId" placeholder="请输入订单号" />
        </UFormField>
        <UFormField label="订单明细 ID" required>
          <UInput v-model.trim="form.orderItemId" placeholder="请输入订单明细 ID" />
        </UFormField>
        <UFormField label="售后类型" required>
          <USelect v-model="form.caseType" :items="caseTypeOptions" class="w-full" />
        </UFormField>
        <UFormField label="原因编码" required>
          <UInput v-model.trim="form.reasonCode" placeholder="例如：damaged" />
        </UFormField>
      </div>

      <UFormField label="补充说明" class="mt-3">
        <UTextarea v-model.trim="form.reasonDetail" :rows="3" placeholder="可选" />
      </UFormField>

      <div class="mt-4 flex gap-2">
        <UButton :loading="submitting" @click="submitCase">提交申请</UButton>
        <UButton color="neutral" variant="soft" :loading="loading" @click="loadCases">刷新列表</UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-base font-semibold">我的售后申请</h2>
          <span class="text-xs text-gray-500">共 {{ cases.length }} 条</span>
        </div>
      </template>

      <div v-if="!cases.length" class="text-sm text-gray-500">暂无售后申请记录。</div>
      <ul v-else class="space-y-2">
        <li v-for="item in cases" :key="item.id" class="rounded-md border border-gray-200 p-3 dark:border-gray-700">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-medium">{{ item.caseNo || item.id }}</span>
            <UBadge color="neutral" variant="subtle">{{ item.caseType }}</UBadge>
            <UBadge color="primary" variant="subtle">{{ item.status }}</UBadge>
          </div>
          <div class="mt-1 text-xs text-gray-500">
            订单 {{ item.orderId }} / 明细 {{ item.orderItemId }}
          </div>
        </li>
      </ul>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { useAfterSalesApi, type AfterSaleCaseType } from '~/composables/api/useAfterSales'

const toast = useToastAlert()
const api = useAfterSalesApi()

const loading = ref(false)
const submitting = ref(false)
const cases = ref<Array<{ id: string; caseNo: string; orderId: string; orderItemId: string; caseType: string; status: string }>>([])

const form = reactive({
  orderId: '',
  orderItemId: '',
  caseType: 'refund_only' as AfterSaleCaseType,
  reasonCode: '',
  reasonDetail: '',
})

const caseTypeOptions = [
  { label: '仅退款', value: 'refund_only' },
  { label: '退货退款', value: 'return_refund' },
  { label: '换货', value: 'exchange' },
]

const loadCases = async () => {
  loading.value = true
  try {
    const resp = await api.listMiniAppCases({ page: 1, pageSize: 20 })
    cases.value = resp.items || []
  } catch (error: any) {
    toast.add({ title: '获取售后列表失败', description: error?.message || '请稍后重试', color: 'error' })
  } finally {
    loading.value = false
  }
}

const submitCase = async () => {
  if (!form.orderId || !form.orderItemId || !form.reasonCode) {
    toast.add({ title: '请填写完整必填字段', color: 'warning' })
    return
  }

  submitting.value = true
  try {
    await api.createMiniAppCase({
      orderId: form.orderId,
      orderItemId: form.orderItemId,
      caseType: form.caseType,
      reasonCode: form.reasonCode,
      reasonDetail: form.reasonDetail || undefined,
    })
    toast.add({ title: '售后申请已提交', color: 'success' })
    form.reasonCode = ''
    form.reasonDetail = ''
    await loadCases()
  } catch (error: any) {
    toast.add({ title: '提交失败', description: error?.message || '请稍后重试', color: 'error' })
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadCases()
})
</script>
