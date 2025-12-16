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
        <UButton
          color="primary"
          icon="i-heroicons-cog-6-tooth"
          :disabled="!canManageStrategy"
          @click="openConfig"
        >
          配置策略
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">基础信息</h3>
          <UBadge :color="statusBadge.color" variant="subtle">
            {{ statusBadge.label }}
          </UBadge>
        </div>
      </template>
      <dl class="grid gap-4 md:grid-cols-3">
        <div>
          <dt class="text-sm text-gray-500">负责人</dt>
          <dd class="text-gray-900 dark:text-white">{{ detail?.ownerUuid ?? '—' }}</dd>
        </div>
        <div>
          <dt class="text-sm text-gray-500">联系人</dt>
          <dd class="text-gray-900 dark:text-white">
            {{ detail?.contact?.name ?? '—' }} · {{ detail?.contact?.phone ?? '' }}
          </dd>
        </div>
        <div>
          <dt class="text-sm text-gray-500">上次同步</dt>
          <dd class="text-gray-900 dark:text-white">{{ detail?.syncHistory?.[0]?.createdAt ? formatDate(detail.syncHistory[0].createdAt) : '—' }}</dd>
        </div>
      </dl>
    </UCard>

    <ChannelTeamSection
      :team="detail?.team"
      :strategy="detail?.strategy"
      :can-edit="canManageStrategy"
      @edit="openConfig"
    />

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold">授权凭证</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">管理 OAuth/API Key/线下凭证。</p>
          </div>
          <UButton size="sm" icon="i-heroicons-plus" @click="openCredentialDrawer">
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
              <p class="text-xs text-gray-500">
                范围：{{ credential.scope?.length ? credential.scope.join(', ') : '全部' }}
              </p>
            </div>
            <UBadge :color="credentialStatusMeta(credential.status).color" variant="subtle">
              {{ credentialStatusMeta(credential.status).label }}
            </UBadge>
          </div>
          <div class="mt-2 text-sm text-gray-500">
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
      <div v-else class="py-8 text-center text-sm text-gray-500">
        暂无凭证，请点击“新增凭证”完成授权。
      </div>
    </UCard>

    <USlideover v-model="credentialDrawerOpen">
      <UCard class="flex h-full flex-col">
        <template #header>
          <div>
            <p class="text-sm text-gray-500">新增或刷新授权凭证</p>
            <h3 class="text-xl font-semibold text-gray-900 dark:text-white">凭证表单</h3>
          </div>
        </template>
        <ChannelCredentialDrawer
          v-model="credentialForm"
          :loading="credentialSaving"
          @submit="handleCredentialSubmit"
          @cancel="() => (credentialDrawerOpen = false)"
        />
      </UCard>
    </USlideover>

    <USlideover v-model="strategyDrawerOpen">
      <UCard class="flex h-full flex-col">
        <template #header>
          <div>
            <p class="text-sm text-gray-500">更新渠道策略与团队</p>
            <h3 class="text-xl font-semibold text-gray-900 dark:text-white">策略配置</h3>
          </div>
        </template>
        <ChannelStrategyForm
          v-model="strategyForm"
          :loading="strategySaving"
          @submit="handleStrategySubmit"
          @cancel="() => (strategyDrawerOpen = false)"
        />
      </UCard>
    </USlideover>

    <div class="grid gap-6 lg:grid-cols-3">
      <ChannelHealthCard
        class="lg:col-span-1"
        :score="detail?.health?.score"
        :labels="detail?.health?.labels"
      />
      <ChannelKpiTrend class="lg:col-span-2" :metrics="detail?.metrics" />
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <ChannelAlertTimeline
        :alerts="alerts"
        :loading="alertsLoading"
        @update="handleAlertUpdate"
      />
      <ChannelTaskPanel
        :tasks="tasks"
        :loading="tasksLoading"
        @link="handleTaskLink"
        @update="handleTaskUpdate"
        @remove="handleTaskRemove"
      />
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <ChannelNotePanel
        :notes="notes"
        :loading="notesLoading"
        @create="handleNoteCreate"
      />
      <ChannelSyncHistory
        :history="syncHistory"
        :loading="syncLoading"
        @trigger="handleManualSync"
      />
    </div>
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

const alertsLoading = ref(false)
const tasksLoading = ref(false)
const notesLoading = ref(false)
const syncLoading = ref(false)

const credentialDrawerOpen = ref(false)
const credentialSaving = ref(false)
const credentials = ref<ChannelCredential[]>([])
const credentialForm = ref<ChannelCredentialUpsertPayload>(createEmptyCredentialPayload())
const strategyDrawerOpen = ref(false)
const strategySaving = ref(false)
const strategyForm = ref<ChannelStrategyUpdatePayload>(createEmptyStrategyPayload())
const { hasPermission } = usePermissions()
const canManageStrategy = computed(() => hasPermission('com.powerx.plugin.ecommerce:channel.strategy:manage'))

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

const openCredentialDrawer = () => {
  credentialForm.value = createEmptyCredentialPayload()
  credentialDrawerOpen.value = true
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

const openConfig = () => {
  if (!canManageStrategy.value) {
    toast.add({
      title: '无权限',
      description: '需要 channel.strategy.manage 权限才能编辑策略',
      color: 'warning',
    })
    return
  }
  strategyForm.value = buildStrategyPayload()
  strategyDrawerOpen.value = true
}

const handleStrategySubmit = async (payload: ChannelStrategyUpdatePayload) => {
  strategySaving.value = true
  try {
    await store.saveStrategy(channelId.value, payload)
    toast.add({ title: '策略已更新' })
    strategyDrawerOpen.value = false
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
    credentialDrawerOpen.value = false
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

const formatDate = (value: string) => {
  try {
    return new Date(value).toLocaleString('zh-CN')
  } catch {
    return value
  }
}
</script>
