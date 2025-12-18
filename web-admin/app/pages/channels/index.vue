<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">店铺 / 渠道</h1>
        <p class="text-gray-500 dark:text-gray-400">
          管理多平台店铺、跟踪授权状态，并在此发起入驻审批。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path" @click="refreshChannels">
          刷新
        </UButton>
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-heroicons-clipboard-document-check"
          @click="goToApproval"
        >
          审批台
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus" @click="startChannelWizard">
          接入新渠道
        </UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <UCard v-for="card in summaryCards" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
          {{ card.value }}
        </div>
        <p class="text-xs text-gray-400">
          {{ card.helper }}
        </p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">渠道列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              支持按平台、状态搜索，抽屉即可完成创建/编辑。
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model.trim="keyword"
              class="w-56"
              placeholder="搜索渠道 / 负责人"
              icon="i-heroicons-magnifying-glass"
            />
            <USelectMenu
              v-model="platformFilter"
              class="w-44"
              :options="platformFilterOptions"
              value-attribute="value"
              option-attribute="label"
              placeholder="全部平台"
            />
            <USelectMenu
              v-model="statusFilter"
              class="w-44"
              :options="statusOptions"
              value-attribute="value"
              option-attribute="label"
              placeholder="授权状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="displayedChannels" :loading="loading">
        <template #name-cell="{ row }">
          <div>
            <p class="font-semibold text-gray-900 dark:text-white">{{ row.original.name }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ row.original.platform }} · {{ row.original.storeId || 'Store N/A' }}
            </p>
          </div>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="statusMeta(row.original.status).color" variant="subtle">
            {{ statusMeta(row.original.status).label }}
          </UBadge>
        </template>
        <template #tags-cell="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge
              v-for="tag in row.original.tags"
              :key="`${row.original.id}-${tag}`"
              color="neutral"
              variant="soft"
              size="xs"
            >
              {{ tag }}
            </UBadge>
            <span v-if="!row.original.tags?.length" class="text-xs text-gray-400">-</span>
          </div>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex flex-wrap gap-2">
            <UButton size="xs" variant="ghost" @click="openChannelDetail(row.original.id)">
              详情
            </UButton>
            <UButton size="xs" variant="ghost" @click="editChannel(row.original)">
              编辑
            </UButton>
            <UButton
              v-if="canSubmit(row.original)"
              size="xs"
              color="primary"
              variant="soft"
              @click="submitForApproval(row.original)"
            >
              提交审批
            </UButton>
          </div>
        </template>
      </UTable>

      <template #footer>
        <div class="flex flex-col gap-3 text-sm text-gray-500 dark:text-gray-400 lg:flex-row lg:items-center lg:justify-between">
          <span>共 {{ total }} 条渠道</span>
          <UPagination v-model="page" :total="total" :page-count="pageSize" show-first show-last/>
        </div>
      </template>
    </UCard>

    <UModal
      v-if="modalOpen"
      v-model:open="modalOpen"
      :title="modalTitle"
      :description="modalDescription"
      :ui="{ content: 'max-w-6xl w-[90vw] mx-auto' }"
      :prevent-close="saving"
    >
      <template #close>
        <UButton
          icon="i-heroicons-x-mark"
          variant="ghost"
          color="neutral"
          square
          :disabled="saving"
          @click="closeModal"
        />
      </template>
      <template #body>
        <div class="p-5">
          <ChannelForm
            v-model="formModel"
            :mode="formMode"
            :loading="saving"
            :platform-options="formPlatformOptions"
            :channel-type-options="formChannelTypeOptions"
            :country-options="formCountryOptions"
            :owner-options="formOwnerOptions"
            :owner-loading="ownersLoading"
            @search-owner="handleOwnerSearch"
            @submit="handleFormSubmit"
            @cancel="closeModal"
          />
        </div>
      </template>
    </UModal>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">渠道上线 Checklist</h3>
            <UBadge color="info" variant="subtle">{{ checklist.length }} 项</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="item in checklist"
            :key="item.id"
            class="flex items-start gap-3 rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <UIcon
              :name="item.done ? 'i-heroicons-check-circle-solid' : 'i-heroicons-clock'"
              :class="[
                'h-5 w-5',
                item.done ? 'text-emerald-500' : 'text-gray-400 dark:text-gray-500',
              ]"
            />
            <div>
              <p class="font-medium text-gray-900 dark:text-white">{{ item.title }}</p>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ item.desc }}</p>
            </div>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                渠道运营提醒
              </h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                预警低库存/活动节点，提醒提前准备。
              </p>
            </div>
            <UBadge color="warning" variant="subtle">重要</UBadge>
          </div>
        </template>

        <ul class="space-y-4">
          <li
            v-for="alert in alerts"
            :key="alert.id"
            class="rounded-xl border border-amber-100 p-4 dark:border-amber-900/40"
          >
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ alert.title }}</span>
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ alert.date }}</span>
            </div>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ alert.desc }}</p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">查看详情</UButton>
              <UButton size="xs" variant="ghost">忽略</UButton>
            </div>
          </li>
        </ul>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import type { TableColumn } from '@nuxt/ui'
import ChannelForm from '~/components/channels/ChannelForm.vue'
import { useChannelsStore } from '~/stores/channels'
import type { ChannelDraftPayload, ChannelSummary, ChannelStatus } from '~/types/channels'
import { createEmptyChannelPayload } from '~/types/channels'

definePageMeta({
  name: 'channels',
})

const router = useRouter()
const toast = useToast()

const store = useChannelsStore()
const { items, total, loading, saving, platforms, channelTypes, countries, owners, ownersLoading } =
  storeToRefs(store)

const keyword = ref('')
const platformFilter = ref('')
const statusFilter = ref('')
const page = ref(1)
const pageSize = ref(10)
const modalOpen = ref(false)
const editingChannel = ref<ChannelSummary | null>(null)
const formModel = ref<ChannelDraftPayload>(createEmptyChannelPayload())
const modalTitle = computed(() => (editingChannel.value ? editingChannel.value.name : '渠道信息'))
const modalDescription = computed(() =>
  editingChannel.value ? '更新渠道资料 / Edit Channel' : '创建新渠道 / Create Channel',
)

const fallbackPlatformOptions = [
  { label: '天猫 Tmall', value: 'tmall' },
  { label: '京东 JD', value: 'jd' },
  { label: '抖音 Douyin', value: 'douyin' },
  { label: '线下 Offline', value: 'offline' },
]

const fallbackChannelTypeOptions = [
  { label: '平台授权 / Platform OAuth', value: 'platform_oauth' },
  { label: '手动凭证 / Manual Credential', value: 'platform_manual' },
  { label: '线下渠道 / Offline', value: 'offline' },
]

const fallbackCountryOptions = [
  {
    code: 'CN',
    label: '中国 China',
    cities: [
      { code: 'cn-beijing', label: '北京 Beijing' },
      { code: 'cn-shanghai', label: '上海 Shanghai' },
      { code: 'cn-shenzhen', label: '深圳 Shenzhen' },
    ],
  },
  {
    code: 'SG',
    label: '新加坡 Singapore',
    cities: [{ code: 'sg-singapore', label: '新加坡 Singapore' }],
  },
]

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '草稿 Draft', value: 'draft' },
  { label: '待审核 Pending', value: 'pending_review' },
  { label: '驳回 Rejected', value: 'rejected' },
  { label: '未授权 Unauthorized', value: 'unauthorized' },
  { label: '已授权 Authorized', value: 'authorized' },
]

const resolvedPlatformOptions = computed(() =>
  platforms.value.length
    ? platforms.value.map((platform) => ({ label: platform.label, value: platform.code }))
    : fallbackPlatformOptions,
)

const resolvedChannelTypeOptions = computed(() =>
  channelTypes.value.length
    ? channelTypes.value.map((type) => ({ label: type.label, value: type.code }))
    : fallbackChannelTypeOptions,
)

const resolvedCountryOptions = computed(() =>
  countries.value.length ? countries.value : fallbackCountryOptions,
)

const platformFilterOptions = computed(() => [
  { label: '全部平台', value: '' },
  ...resolvedPlatformOptions.value,
])

const formPlatformOptions = resolvedPlatformOptions

const formChannelTypeOptions = resolvedChannelTypeOptions

const formCountryOptions = resolvedCountryOptions

const formOwnerOptions = computed(() =>
  owners.value.length
    ? owners.value.map((owner) => ({
        label: owner.displayName || owner.username,
        value: owner.username || String(owner.id),
        description: owner.email,
      }))
    : [],
)

const columns: TableColumn<ChannelSummary>[] = [
  { accessorKey: 'name', header: '渠道' },
  { accessorKey: 'region', header: '区域' },
  { accessorKey: 'ownerUuid', header: '负责人' },
  { accessorKey: 'status', header: '状态' },
  { accessorKey: 'tags', header: '标签' },
  { id: 'actions', header: '操作' },
]

const displayedChannels = computed(() => items.value ?? [])

const summaryCards = computed(() => {
  const pending = displayedChannels.value.filter((c) => c.status === 'pending_review').length
  const unauthorized = displayedChannels.value.filter((c) => c.status === 'unauthorized').length
  const offline = displayedChannels.value.filter((c) => c.channelType === 'offline').length
  return [
    { title: '渠道总数 Channels', value: total.value, helper: '含全部平台 (取自分页数据)' },
    { title: '待审批 Pending', value: pending, helper: '当前页中等待审批的渠道' },
    { title: '线下渠道 Offline', value: offline, helper: `未授权：${unauthorized}` },
  ]
})

const statusMeta = (status: ChannelStatus) => {
  switch (status) {
    case 'draft':
      return { label: '草稿', color: 'neutral' }
    case 'pending_review':
      return { label: '待审核', color: 'warning' }
    case 'rejected':
      return { label: '已退回', color: 'error' }
    case 'unauthorized':
      return { label: '未授权', color: 'info' }
    case 'authorized':
      return { label: '已授权', color: 'success' }
    case 'disabled':
      return { label: '已停用', color: 'neutral' }
    default:
      return { label: status, color: 'neutral' }
  }
}

const loadChannels = async () => {
  try {
    await store.fetchList({
      keyword: keyword.value || undefined,
      platform: platformFilter.value || undefined,
      status: statusFilter.value ? [statusFilter.value] : undefined,
      page: page.value,
      pageSize: pageSize.value,
    })
  } catch (error: any) {
    toast.add({
      title: '渠道列表加载失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const loadPlatformCatalog = async () => {
  try {
    await store.fetchPlatformCatalog()
  } catch (error: any) {
    toast.add({
      title: '平台枚举加载失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const loadOwnerOptions = async () => {
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

const handleOwnerSearch = async (keyword: string) => {
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

await Promise.all([loadChannels(), loadPlatformCatalog(), loadOwnerOptions()])

watch(
  [keyword, platformFilter, statusFilter],
  () => {
    page.value = 1
    loadChannels()
  },
  { deep: true },
)

watch(page, (val, old) => {
  if (val === old) return
  loadChannels()
})

const refreshChannels = () => {
  loadChannels()
}

const goToApproval = () => router.push('/channels/approval')

const formMode = computed(() => (editingChannel.value ? 'edit' : 'create'))

const startChannelWizard = () => {
  editingChannel.value = null
  formModel.value = createEmptyChannelPayload()
  modalOpen.value = true
}

const mapSummaryToDraft = (channel: ChannelSummary): ChannelDraftPayload => ({
  ...createEmptyChannelPayload(),
  name: channel.name,
  platform: channel.platform,
  storeId: channel.storeId || '',
  region: channel.region,
  ownerUuid: channel.ownerUuid,
  channelType: channel.channelType,
  tags: [...(channel.tags ?? [])],
})

const editChannel = (channel: ChannelSummary) => {
  editingChannel.value = channel
  formModel.value = mapSummaryToDraft(channel)
  modalOpen.value = true
}

const closeModal = () => {
  ;(document.activeElement as HTMLElement | null)?.blur?.()
  modalOpen.value = false
}

const handleFormSubmit = async (payload: ChannelDraftPayload) => {
  try {
    if (editingChannel.value) {
      await store.updateChannel(editingChannel.value.id, payload)
      toast.add({ title: '渠道已更新', description: editingChannel.value.name })
    } else {
      await store.createChannel(payload)
      toast.add({ title: '渠道已创建', description: payload.name })
    }
    modalOpen.value = false
    await loadChannels()
  } catch (error: any) {
    toast.add({
      title: '保存失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const canSubmit = (channel: ChannelSummary) => channel.status === 'draft'

const submitForApproval = async (channel: ChannelSummary) => {
  try {
    await store.submitChannel(channel.id)
    toast.add({ title: '已提交审批', description: `${channel.name} 正在等待审批` })
    await loadChannels()
  } catch (error: any) {
    toast.add({
      title: '提交失败',
      description: error?.message ?? '请稍后重试',
      color: 'error',
    })
  }
}

const openChannelDetail = (id: string) => {
  router.push(`/channels/${id}`)
}

const checklist = [
  { id: 1, title: '完成资料校验', desc: '法人信息、联系人、域名已补充', done: true },
  { id: 2, title: '上传授权凭证', desc: '等待平台授权回调', done: false },
  { id: 3, title: '确认负责人与审批链路', desc: '运营负责人和审批人已配置', done: true },
]

const alerts = [
  { id: 1, title: '凭证即将过期', date: '今天', desc: '京东自营旗舰凭证 7 天后过期' },
  { id: 2, title: '渠道 GMV 下滑', date: '昨天', desc: '天猫国际 GMV 同比 -12%，请关注' },
]
</script>
