<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">渠道审批台</h1>
        <p class="text-gray-500 dark:text-gray-400">审核渠道入驻与变更，保障资料一致性。</p>
      </div>
      <div class="flex gap-2">
        <UButton to="/channels" variant="ghost" color="neutral" icon="i-heroicons-arrow-left">
          返回列表
        </UButton>
        <USelectMenu
          v-model="platformFilter"
          :options="platformOptions"
          class="w-48"
          value-attribute="value"
          option-attribute="label"
          placeholder="全部平台"
        />
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold">待处理申请</h3>
            <p class="text-sm text-gray-500">数据来自 /channels?status=pending_review</p>
          </div>
          <UBadge color="info" variant="subtle">{{ approvals.length }} 条</UBadge>
        </div>
      </template>
      <div v-if="loading" class="flex items-center justify-center py-12 text-gray-500">
        正在加载...
      </div>
      <div v-else class="space-y-4">
        <UCard
          v-for="request in approvals"
          :key="request.id"
          class="bg-slate-50/80 dark:bg-slate-900/40"
        >
          <div class="flex flex-col gap-4">
            <div class="flex flex-col gap-1">
              <p class="text-sm text-gray-500">
                {{ request.platform }} · {{ request.region }}
              </p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">{{ request.name }}</p>
              <p class="text-sm text-gray-500">
                负责人：{{ request.ownerUuid }} · 渠道类型：{{ request.channelType }}
              </p>
            </div>
            <div class="space-y-2">
              <p class="text-sm text-gray-500">
                审批备注（可选）：
              </p>
              <UTextarea
                v-model="decisionNotes[request.id]"
                placeholder="驳回时请注明原因，审批意见将写入审计日志"
              />
            </div>
            <div class="flex flex-wrap gap-2">
              <UButton color="neutral" variant="ghost" @click="viewDetail(request.id)">
                查看详情
              </UButton>
              <UButton color="success" icon="i-heroicons-check" @click="decide(request, 'approve')">
                通过
              </UButton>
              <UButton color="error" variant="outline" icon="i-heroicons-x-mark" @click="decide(request, 'reject')">
                驳回
              </UButton>
            </div>
          </div>
        </UCard>
        <div v-if="!approvals.length" class="py-6 text-center text-sm text-gray-500">
          当前没有待审批的渠道。
        </div>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { ChannelSummary } from '~/types/channels'
import { useChannelsStore } from '~/stores/channels'

const router = useRouter()
const toast = useToast()
const store = useChannelsStore()

const approvals = ref<ChannelSummary[]>([])
const loading = ref(false)
const platformFilter = ref('')
const decisionNotes = reactive<Record<string, string>>({})

const loadApprovals = async () => {
  loading.value = true
  try {
    const { items } = await store.fetchApprovals({
      platform: platformFilter.value || undefined,
    })
    approvals.value = items ?? []
  } catch (error: any) {
    toast.add({
      title: '加载审批列表失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    loading.value = false
  }
}

await loadApprovals()

watch(platformFilter, () => {
  loadApprovals()
})

const platformOptions = computed(() => {
  const unique = new Set(approvals.value.map((item) => item.platform))
  return [{ label: '全部平台', value: '' }, ...Array.from(unique).map((platform) => ({ label: platform, value: platform }))]
})

const decide = async (request: ChannelSummary, decision: 'approve' | 'reject') => {
  if (decision === 'reject' && !decisionNotes[request.id]) {
    toast.add({ title: '请填写驳回原因', color: 'warning' })
    return
  }
  try {
    await store.decideApproval(request.id, {
      decision,
      reason: decision === 'reject' ? decisionNotes[request.id] : undefined,
    })
    toast.add({
      title: decision === 'approve' ? '审批已通过' : '审批已驳回',
      description: request.name,
      color: decision === 'approve' ? 'success' : 'warning',
    })
    await loadApprovals()
  } catch (error: any) {
    toast.add({
      title: '提交审批结果失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const viewDetail = (id: string) => {
  router.push(`/channels/${id}`)
}
</script>
