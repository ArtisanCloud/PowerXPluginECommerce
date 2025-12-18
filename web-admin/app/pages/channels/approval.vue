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
            <div v-if="request.approvalHistory?.length" class="rounded-lg border border-gray-200/70 bg-white/70 p-3 dark:border-gray-800 dark:bg-slate-900/60">
              <p class="text-xs font-medium uppercase text-gray-400">最近记录</p>
              <div class="mt-1 space-y-1">
                <p class="text-sm text-gray-900 dark:text-gray-100">
                  {{ formatHistoryLabel(latestHistoryEntry(request)) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ formatHistoryTime(latestHistoryEntry(request)?.at) }} · {{ latestHistoryEntry(request)?.actor || 'system' }}
                </p>
                <p v-if="latestHistoryEntry(request)?.reason" class="text-xs text-rose-400">
                  原因：{{ latestHistoryEntry(request)?.reason }}
                </p>
              </div>
            </div>
            <div class="flex flex-wrap gap-2">
              <UButton color="neutral" variant="ghost" @click="viewDetail(request.id)">
                查看详情
              </UButton>
              <UButton color="primary" icon="i-heroicons-clipboard-document-check" @click="openApprovalModal(request)">
                处理申请
              </UButton>
            </div>
          </div>
        </UCard>
        <div v-if="!approvals.length" class="py-6 text-center text-sm text-gray-500">
          当前没有待审批的渠道。
        </div>
      </div>
    </UCard>
    <UModal
      v-model:open="approvalModalOpen"
      title="申请记录与审批操作"
      description="参考历史记录并提交审批意见"
      :prevent-close="store.saving"
      :ui="{ content: 'max-w-3xl w-[90vw]' }"
    >
      <template #body>
        <div class="space-y-6" v-if="activeApproval">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ activeApproval.name }}</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ activeApproval.platform }} · {{ activeApproval.region }}
            </p>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">申请记录</p>
            <div v-if="modalHistory.length" class="mt-3 space-y-3">
              <div
                v-for="entry in modalHistory"
                :key="entry.at + entry.event"
                class="rounded-lg border border-gray-200/80 p-3 dark:border-gray-700"
              >
                <div class="flex items-center justify-between">
                  <span class="text-sm font-semibold text-gray-900 dark:text-white">
                    {{ formatHistoryLabel(entry) }}
                  </span>
                  <UBadge variant="soft">{{ entry.actor || 'system' }}</UBadge>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ formatHistoryTime(entry.at) }}</p>
                <p v-if="entry.reason" class="mt-1 text-sm text-rose-400">原因：{{ entry.reason }}</p>
                <p v-if="entry.note && entry.event === 'submitted'" class="mt-1 text-xs text-gray-500">
                  备注：{{ entry.note }}
                </p>
              </div>
            </div>
            <p v-else class="mt-3 text-sm text-gray-500 dark:text-gray-400">暂无历史记录，当前为首次提审。</p>
          </div>
          <div class="space-y-3">
            <UFormField label="审批备注" help="驳回时为必填，将记录在审计日志中。">
              <template #default="{ id }">
                <UTextarea
                  :id="id"
                  v-model="modalNote"
                  placeholder="例如：资料缺少授权书，请补齐后重新提交"
                  :rows="3"
                />
              </template>
            </UFormField>
            <div class="flex flex-wrap gap-2">
              <UButton color="success" :loading="store.saving" @click="submitDecision('approve')">
                通过
              </UButton>
              <UButton color="error" variant="soft" :loading="store.saving" @click="submitDecision('reject')">
                驳回
              </UButton>
              <UButton color="neutral" variant="ghost" @click="closeApprovalModal">
                取消
              </UButton>
            </div>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { ChannelSummary, ChannelApprovalHistoryEntry } from '~/types/channels'
import { useChannelsStore } from '~/stores/channels'

const router = useRouter()
const toast = useToast()
const store = useChannelsStore()

const approvals = ref<ChannelSummary[]>([])
const loading = ref(false)
const platformFilter = ref('')
const decisionNotes = reactive<Record<string, string>>({})
const approvalModalOpen = ref(false)
const activeApproval = ref<ChannelSummary | null>(null)
const modalNote = ref('')

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

const modalHistory = computed(() => activeApproval.value?.approvalHistory ?? [])

const latestHistoryEntry = (request: ChannelSummary) => {
  const history = request.approvalHistory ?? []
  return history.length ? history[history.length - 1] : undefined
}

const formatHistoryLabel = (entry?: ChannelApprovalHistoryEntry) => {
  if (!entry) {
    return '暂无记录'
  }
  const mapping: Record<string, string> = {
    submitted: '发起申请',
    approve: '审批通过',
    approved: '审批通过',
    reject: '审批驳回',
    rejected: '审批驳回',
  }
  return mapping[entry.event] || entry.event
}

const formatHistoryTime = (value?: string) => {
  if (!value) return '--'
  try {
    return new Date(value).toLocaleString('zh-CN')
  } catch {
    return value
  }
}

const openApprovalModal = (request: ChannelSummary) => {
  activeApproval.value = request
  modalNote.value = decisionNotes[request.id] || ''
  approvalModalOpen.value = true
}

const closeApprovalModal = () => {
  approvalModalOpen.value = false
  activeApproval.value = null
  modalNote.value = ''
}

watch(
  () => activeApproval.value?.id,
  (id) => {
    if (!id) {
      modalNote.value = ''
      return
    }
    modalNote.value = decisionNotes[id] || ''
  },
)

watch(modalNote, (val) => {
  if (activeApproval.value) {
    decisionNotes[activeApproval.value.id] = val
  }
})

const decide = async (request: ChannelSummary, decision: 'approve' | 'reject', options: { closeModal?: boolean } = {}) => {
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
    if (options.closeModal) {
      closeApprovalModal()
    }
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

const submitDecision = async (decision: 'approve' | 'reject') => {
  if (!activeApproval.value) return
  await decide(activeApproval.value, decision, { closeModal: true })
}
</script>
