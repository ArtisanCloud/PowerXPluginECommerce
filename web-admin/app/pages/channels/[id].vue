<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <p class="text-sm text-gray-500 dark:text-gray-400">渠道 ID · {{ channelId }}</p>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ detail?.name ?? '加载中…' }}
        </h1>
        <p class="text-gray-500 dark:text-gray-400">
          {{ detail?.platform ?? '--' }} · {{ detail?.region ?? '未知' }}
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-uturn-left" @click="goBack">
          返回列表
        </UButton>
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path" :loading="syncLoading" @click="handleManualSync">
          手动同步
        </UButton>
      </div>
    </div>

    <UCard :class="cardClass">
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">基础信息</h3>
          <UBadge :color="statusBadge.color" variant="subtle">
            {{ statusBadge.label }}
          </UBadge>
        </div>
      </template>
      <dl class="grid gap-4 md:grid-cols-3">
        <div>
          <dt class="text-sm text-gray-500 dark:text-gray-400">负责人</dt>
          <dd class="text-gray-900 dark:text-white">{{ detail?.ownerUuid ?? '—' }}</dd>
        </div>
        <div>
          <dt class="text-sm text-gray-500 dark:text-gray-400">联系人</dt>
          <dd class="text-gray-900 dark:text-white">
            {{ detail?.contact?.name ?? '—' }} · {{ detail?.contact?.phone ?? '' }}
          </dd>
        </div>
        <div>
          <dt class="text-sm text-gray-500 dark:text-gray-400">上次同步</dt>
          <dd class="text-gray-900 dark:text-white">{{ detail?.syncHistory?.[0]?.createdAt ? formatDate(detail.syncHistory[0].createdAt) : '—' }}</dd>
        </div>
      </dl>
    </UCard>

    <ChannelTeamSection
      :class="cardClass"
      :team="detail?.team"
      :strategy="detail?.strategy"
      :can-edit="canManageStrategy"
      @edit="openConfig"
    />

    <UCard :class="cardClass">
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">授权凭证</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">管理 OAuth/API Key/线下凭证。</p>
          </div>
          <UButton size="sm" icon="i-heroicons-plus" @click="openCredentialModal">
            新增凭证
          </UButton>
        </div>
      </template>
      <div v-if="credentials.length" class="space-y-4">
        <div
          v-for="credential in credentials"
          :key="credential.id"
          class="rounded-lg border border-gray-100 p-4 dark:border-gray-800"
        >
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="font-semibold text-gray-900 dark:text-white">{{ credential.type }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                范围：{{ credential.scope?.length ? credential.scope.join(', ') : '全部' }}
              </p>
            </div>
            <UBadge :color="credentialStatusMeta(credential.status).color" variant="subtle">
              {{ credentialStatusMeta(credential.status).label }}
            </UBadge>
          </div>
          <div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
            <p>
              到期时间：{{ credential.expiresAt ? formatDate(credential.expiresAt) : '未设置' }}
            </p>
            <p>
              最近巡检：{{ credential.lastTestedAt ? formatDate(credential.lastTestedAt) : '未测试' }}
            </p>
          </div>
          <div v-if="credential.attachmentUrl" class="mt-2 text-sm">
            <a :href="credential.attachmentUrl" class="text-primary-600 hover:underline dark:text-primary-400" target="_blank" rel="noopener">
              查看线下附件
            </a>
          </div>
          <div class="mt-3 flex flex-wrap gap-2">
            <UButton size="xs" variant="ghost" @click="markCredentialTested(credential, true)">标记成功</UButton>
            <UButton size="xs" variant="ghost" color="warning" @click="markCredentialTested(credential, false)">
              标记失败
            </UButton>
          </div>
        </div>
      </div>
      <div v-else class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
        暂无凭证，请点击“新增凭证”完成授权。
      </div>
    </UCard>

    <UModal
      v-model:open="credentialModalOpen"
      :prevent-close="credentialSaving"
      :close="!credentialSaving"
      :title="credentialModalTitle"
      :description="credentialModalDescription"
      :ui="{ content: 'max-w-3xl w-[90vw]' }"
    >
      <template #body>
        <ChannelCredentialDrawer
          v-model="credentialForm"
          :loading="credentialSaving"
          @submit="handleCredentialSubmit"
          @cancel="closeCredentialModal"
        />
      </template>
    </UModal>

    <UModal
      v-model:open="strategyModalOpen"
      :prevent-close="strategySaving"
      :close="!strategySaving"
      :title="strategyModalTitle"
      :description="strategyModalDescription"
      :ui="{ content: 'max-w-4xl w-[90vw]' }"
    >
      <template #body>
        <ChannelStrategyForm
          v-model="strategyForm"
          :loading="strategySaving"
          :owner-options="ownerOptions"
          :owner-loading="store.ownersLoading"
          @submit="handleStrategySubmit"
          @cancel="closeStrategyModal"
          @search-owner="handleStrategyOwnerSearch"
        />
      </template>
    </UModal>

    <div class="grid gap-6 lg:grid-cols-3">
      <ChannelHealthCard
        :class="['lg:col-span-1', cardClass]"
        :score="detail?.health?.score"
        :labels="detail?.health?.labels"
      />
      <ChannelKpiTrend :class="['lg:col-span-2', cardClass]" :metrics="detail?.metrics" />
    </div>

    <div class="grid gap-4 lg:grid-cols-2">
      <UCard :class="cardClass">
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500 dark:text-gray-400">告警总览</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">
                {{ alerts.length }} 条
              </p>
            </div>
            <UBadge color="warning" variant="subtle">{{ alerts.length }}</UBadge>
          </div>
        </template>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ latestAlertTitle }}
        </p>
        <div class="mt-4 flex justify-end">
          <UButton size="sm" variant="soft" @click="alertsModalOpen = true">
            查看告警时间轴
          </UButton>
        </div>
      </UCard>

      <UCard :class="cardClass">
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500 dark:text-gray-400">任务中心</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">
                {{ pendingTaskCount }} / {{ tasks.length }}
              </p>
            </div>
            <UBadge color="info" variant="subtle">
              进行中
            </UBadge>
          </div>
        </template>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          跟踪渠道接入、巡检和审批相关任务
        </p>
        <div class="mt-4 flex justify-end">
          <UButton size="sm" variant="soft" @click="tasksModalOpen = true">
            管理任务
          </UButton>
        </div>
      </UCard>

      <UCard :class="cardClass">
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500 dark:text-gray-400">内部备注</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">
                {{ notes.length }} 条
              </p>
            </div>
            <UBadge color="neutral" variant="subtle">
              最新
            </UBadge>
          </div>
        </template>
        <p class="text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
          {{ latestNotePreview }}
        </p>
        <div class="mt-4 flex justify-end">
          <UButton size="sm" variant="soft" @click="notesModalOpen = true">
            查看备注
          </UButton>
        </div>
      </UCard>

      <UCard :class="cardClass">
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500 dark:text-gray-400">同步历史</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">
                {{ latestSyncLabel }}
              </p>
            </div>
            <UBadge color="primary" variant="subtle">
              {{ latestSyncResult }}
            </UBadge>
          </div>
        </template>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          查看渠道与外部平台的同步日志
        </p>
        <div class="mt-4 flex justify-end gap-2">
          <UButton size="sm" variant="ghost" :loading="syncLoading" @click="handleManualSync">
            手动同步
          </UButton>
          <UButton size="sm" variant="soft" @click="syncModalOpen = true">
            查看记录
          </UButton>
        </div>
      </UCard>
    </div>

    <UModal
      v-model:open="alertsModalOpen"
      title="告警时间轴"
      description="查看渠道的历史告警与状态"
      :ui="{ content: 'max-w-5xl w-[90vw]' }"
    >
      <template #body>
        <ChannelAlertTimeline
          :alerts="alerts"
          :loading="alertsLoading"
          @update="handleAlertUpdate"
        />
      </template>
    </UModal>

    <UModal
      v-model:open="tasksModalOpen"
      title="任务管理"
      description="跟踪渠道接入、巡检和审批相关任务"
      :ui="{ content: 'max-w-5xl w-[90vw]' }"
    >
      <template #body>
        <ChannelTaskPanel
          :tasks="tasks"
          :loading="tasksLoading"
          @link="handleTaskLink"
          @update="handleTaskUpdate"
          @remove="handleTaskRemove"
        />
      </template>
    </UModal>

    <UModal
      v-model:open="notesModalOpen"
      title="渠道备注"
      description="查看与维护当前渠道的内部备注"
      :ui="{ content: 'max-w-4xl w-[90vw]' }"
    >
      <template #body>
        <ChannelNotePanel
          :notes="notes"
          :loading="notesLoading"
          @create="handleNoteCreate"
        />
      </template>
    </UModal>

    <UModal
      v-model:open="syncModalOpen"
      title="同步历史"
      description="查看渠道与外部平台的同步记录"
      :ui="{ content: 'max-w-4xl w-[90vw]' }"
    >
      <template #body>
        <ChannelSyncHistory
          :history="syncHistory"
          :loading="syncLoading"
          @trigger="handleManualSync"
        />
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useChannelsStore } from '~/stores/channels'
import ChannelCredentialDrawer from '~/components/channels/ChannelCredentialDrawer.vue'
import ChannelHealthCard from '~/components/channels/ChannelHealthCard.vue'
import ChannelKpiTrend from '~/components/channels/ChannelKpiTrend.vue'
import ChannelAlertTimeline from '~/components/channels/ChannelAlertTimeline.vue'
import ChannelTaskPanel from '~/components/channels/ChannelTaskPanel.vue'
import ChannelNotePanel from '~/components/channels/ChannelNotePanel.vue'
import ChannelSyncHistory from '~/components/channels/ChannelSyncHistory.vue'
import ChannelStrategyForm from '~/components/channels/ChannelStrategyForm.vue'
import ChannelTeamSection from '~/components/channels/ChannelTeamSection.vue'
import type {
  ChannelCredential,
  ChannelCredentialUpsertPayload,
  ChannelAlertUpdatePayload,
  ChannelTaskLinkPayload,
  ChannelNotePayload,
  ChannelStrategyUpdatePayload,
} from '~/types/channels'
import { createEmptyCredentialPayload, createEmptyStrategyPayload } from '~/types/channels'
import { usePermissions } from '~/composables/usePermissions'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const store = useChannelsStore()
const nuxtApp = useNuxtApp()

const channelId = computed(() => String(route.params.id ?? ''))

const detail = computed(() => store.current)
const alerts = computed(() => store.alerts)
const tasks = computed(() => store.tasks)
const notes = computed(() => store.notes)
const syncHistory = computed(() => store.syncHistory)
const ownerOptions = computed(() => store.owners)

const alertsLoading = ref(false)
const tasksLoading = ref(false)
const notesLoading = ref(false)
const syncLoading = ref(false)

const credentialModalOpen = ref(false)
const credentialSaving = ref(false)
const credentials = ref<ChannelCredential[]>([])
const credentialForm = ref<ChannelCredentialUpsertPayload>(createEmptyCredentialPayload())
const strategyModalOpen = ref(false)
const strategySaving = ref(false)
const strategyForm = ref<ChannelStrategyUpdatePayload>(createEmptyStrategyPayload())
const { hasPermission } = usePermissions()
const canManageStrategy = computed(() => hasPermission('com.powerx.plugin.ecommerce:channel.strategy:manage'))
const alertsModalOpen = ref(false)
const tasksModalOpen = ref(false)
const notesModalOpen = ref(false)
const syncModalOpen = ref(false)
const cardClass = 'channel-panel-card'
const credentialModalTitle = '凭证表单'
const credentialModalDescription = '新增或刷新渠道授权凭证'
const strategyModalTitle = '策略配置'
const strategyModalDescription = '更新渠道策略与团队'

const blurActiveElement = () => {
  if (typeof document === 'undefined') return
  const active = document.activeElement as HTMLElement | null
  if (active && typeof active.blur === 'function') {
    active.blur()
  }
}

const closeCredentialModal = () => {
  blurActiveElement()
  credentialModalOpen.value = false
}

const closeStrategyModal = () => {
  blurActiveElement()
  strategyModalOpen.value = false
}

const statusBadge = computed(() => {
  const status = detail.value?.status ?? 'unknown'
  const map: Record<string, { label: string; color: string }> = {
    draft: { label: '草稿', color: 'neutral' },
    pending_review: { label: '待审批', color: 'warning' },
    rejected: { label: '已驳回', color: 'error' },
    unauthorized: { label: '未授权', color: 'warning' },
    authorized: { label: '已授权', color: 'success' },
    disabled: { label: '已停用', color: 'neutral' },
  }
  return map[status] ?? { label: status, color: 'neutral' }
})

const formatDate = (value: string) => {
  try {
    return new Date(value).toLocaleString('zh-CN')
  } catch {
    return value
  }
}

const latestAlertTitle = computed(() => alerts.value[0]?.title ?? '暂无告警')
const pendingTaskCount = computed(() => tasks.value.filter((task) => task.status !== 'done').length)
const latestNotePreview = computed(() => notes.value[0]?.body ?? '暂无备注')
const latestSyncLabel = computed(() => {
  const record = syncHistory.value[0]
  return record?.createdAt ? formatDate(record.createdAt) : '暂无同步记录'
})
const latestSyncResult = computed(() => syncHistory.value[0]?.result ?? '暂无执行结果')

const measureKpiLoad = async () => {
  const supportsPerf = typeof performance !== 'undefined'
  const start = supportsPerf ? performance.now() : 0
  await store.fetchDetail(channelId.value)
  if (supportsPerf && nuxtApp.$perf) {
    const duration = performance.now() - start
    nuxtApp.$perf.logKpiLoad(duration)
  }
}

const loadDetail = async () => {
  try {
    await measureKpiLoad()
    await Promise.all([
      loadCredentials(),
      store.refreshAlerts(channelId.value),
      store.fetchTasks(channelId.value),
      store.fetchNotes(channelId.value),
      store.fetchSyncHistory(channelId.value),
    ])
  } catch (error: any) {
    toast.add({
      title: '加载详情失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

watch(channelId, () => {
  loadDetail()
})

onMounted(() => {
  loadDetail()
})

const loadCredentials = async () => {
  try {
    credentials.value = await store.fetchCredentials(channelId.value)
  } catch (error: any) {
    toast.add({
      title: '加载凭证失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const openCredentialModal = () => {
  credentialForm.value = createEmptyCredentialPayload()
  credentialModalOpen.value = true
}

const buildStrategyPayload = (): ChannelStrategyUpdatePayload => {
  const payload = createEmptyStrategyPayload()
  if (detail.value?.strategy) {
    Object.assign(payload.strategy, detail.value.strategy)
  }
  if (detail.value?.team) {
    Object.assign(payload.team, detail.value.team)
    payload.team.operators = [...(detail.value.team.operators ?? [])]
  }
  if (typeof payload.strategy.feeRate !== 'number') {
    payload.strategy.feeRate = detail.value?.strategy?.feeRate ?? 0
  }
  return payload
}

const ensureOwnerOptions = async () => {
  if (ownerOptions.value.length) {
    return
  }
  try {
    await store.fetchOwners()
  } catch (error: any) {
    toast.add({
      title: '负责人列表加载失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const handleStrategyOwnerSearch = async (keyword: string) => {
  try {
    await store.fetchOwners(keyword)
  } catch (error: any) {
    toast.add({
      title: '负责人搜索失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const openConfig = async () => {
  if (!canManageStrategy.value) {
    toast.add({
      title: '无权限',
      description: '需要 channel.strategy.manage 权限才能编辑策略',
      color: 'warning',
    })
    return
  }
  await ensureOwnerOptions()
  strategyForm.value = buildStrategyPayload()
  strategyModalOpen.value = true
}

const handleStrategySubmit = async (payload: ChannelStrategyUpdatePayload) => {
  strategySaving.value = true
  try {
    await store.saveStrategy(channelId.value, payload)
    toast.add({ title: '策略已更新' })
    closeStrategyModal()
  } catch (error: any) {
    toast.add({
      title: '更新策略失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    strategySaving.value = false
  }
}

const handleCredentialSubmit = async (payload: ChannelCredentialUpsertPayload) => {
  credentialSaving.value = true
  try {
    await store.saveCredential(channelId.value, payload)
    toast.add({ title: '凭证已保存' })
    closeCredentialModal()
    await loadCredentials()
  } catch (error: any) {
    toast.add({
      title: '保存凭证失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    credentialSaving.value = false
  }
}

const markCredentialTested = async (credential: ChannelCredential, succeeded: boolean) => {
  try {
    await store.testCredential(channelId.value, {
      type: credential.type,
      succeeded,
      result: { tested_at: new Date().toISOString() },
    })
    toast.add({
      title: succeeded ? '测试通过' : '测试失败',
      description: credential.type,
      color: succeeded ? 'success' : 'warning',
    })
    await loadCredentials()
  } catch (error: any) {
    toast.add({
      title: '提交巡检结果失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const credentialStatusMeta = (status: string) => {
  switch (status) {
    case 'expiring':
      return { label: '即将到期', color: 'warning' }
    case 'expired':
      return { label: '已过期', color: 'error' }
    case 'test_failed':
      return { label: '测试失败', color: 'error' }
    case 'valid':
      return { label: '有效', color: 'success' }
    default:
      return { label: status, color: 'neutral' }
  }
}

const handleAlertUpdate = async ({ alertId, status }: { alertId: string; status: string }) => {
  alertsLoading.value = true
  try {
    const payload: ChannelAlertUpdatePayload = { status }
    await store.updateAlert(channelId.value, alertId, payload)
    toast.add({ title: '告警已更新' })
  } catch (error: any) {
    toast.add({
      title: '更新告警失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    alertsLoading.value = false
  }
}

const handleTaskLink = async (payload: ChannelTaskLinkPayload) => {
  tasksLoading.value = true
  try {
    await store.linkTask(channelId.value, payload)
    toast.add({ title: '任务已关联' })
  } catch (error: any) {
    toast.add({
      title: '关联任务失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    tasksLoading.value = false
  }
}

const handleTaskUpdate = async (payload: { id: string; status: string }) => {
  tasksLoading.value = true
  try {
    await store.updateTask(channelId.value, payload.id, { status: payload.status })
    toast.add({ title: '任务状态已更新' })
  } catch (error: any) {
    toast.add({
      title: '更新任务失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    tasksLoading.value = false
  }
}

const handleTaskRemove = async (taskLinkId: string) => {
  tasksLoading.value = true
  try {
    await store.removeTask(channelId.value, taskLinkId)
    toast.add({ title: '任务已解除' })
  } catch (error: any) {
    toast.add({
      title: '解除任务失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    tasksLoading.value = false
  }
}

const handleNoteCreate = async (payload: ChannelNotePayload) => {
  notesLoading.value = true
  try {
    await store.createNote(channelId.value, payload)
    toast.add({ title: '备注已保存' })
  } catch (error: any) {
    toast.add({
      title: '新增备注失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    notesLoading.value = false
  }
}

const handleManualSync = async () => {
  syncLoading.value = true
  try {
    await store.triggerSync(channelId.value)
    toast.add({ title: '同步任务已创建' })
  } catch (error: any) {
    toast.add({
      title: '创建同步任务失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  } finally {
    syncLoading.value = false
  }
}

const goBack = () => router.push('/channels')
</script>

<style scoped>
:global(.channel-panel-card) {
  position: relative;
  border-radius: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.35) !important;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.97), rgba(248, 250, 252, 0.92)) !important;
  box-shadow: 0 25px 50px rgba(15, 23, 42, 0.15);
  backdrop-filter: blur(24px);
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

:global(.channel-panel-card:hover) {
  border-color: rgba(59, 130, 246, 0.35) !important;
  box-shadow: 0 30px 70px rgba(15, 23, 42, 0.25);
}

:global(.dark .channel-panel-card) {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.88), rgba(15, 23, 42, 0.95)) !important;
  border-color: rgba(148, 163, 184, 0.55) !important;
  box-shadow: 0 30px 70px rgba(2, 6, 23, 0.85);
}

:global(.dark .channel-panel-card:hover) {
  border-color: rgba(129, 140, 248, 0.65) !important;
}

:global(.channel-panel-card h3) {
  color: #0f172a;
}

:global(.dark .channel-panel-card h3) {
  color: #f8fafc;
}
</style>
