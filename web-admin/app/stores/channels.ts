import { defineStore } from 'pinia'
import type {
  ChannelAlert,
  ChannelApprovalPayload,
  ChannelCredentialTestPayload,
  ChannelCredentialUpsertPayload,
  ChannelDetail,
  ChannelDraftPayload,
  ChannelListParams,
  ChannelNote,
  ChannelNotePayload,
  ChannelSubmitPayload,
  ChannelSummary,
  ChannelSyncHistoryItem,
  ChannelTaskLink,
  ChannelTaskLinkPayload,
  ChannelAlertUpdatePayload,
  ChannelSyncTriggerPayload,
  ChannelStrategySnapshot,
  ChannelStrategyUpdatePayload,
  ChannelPlatform,
  ChannelChannelType,
  ChannelRegion,
  ChannelCountry,
  ChannelOwnerOption,
} from '~/types/channels'
import { useChannelsApi } from '~/composables/useChannels'

interface ChannelState {
  items: ChannelSummary[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  saving: boolean
  current: ChannelDetail | null
  detailLoading: boolean
  alerts: ChannelAlert[]
  tasks: ChannelTaskLink[]
  notes: ChannelNote[]
  syncHistory: ChannelSyncHistoryItem[]
  filters: ChannelListParams
  platforms: ChannelPlatform[]
  channelTypes: ChannelChannelType[]
  regions: ChannelRegion[]
  countries: ChannelCountry[]
  owners: ChannelOwnerOption[]
  ownersLoading: boolean
  platformCatalogLoading: boolean
}

export const useChannelsStore = defineStore('channels', {
    state: (): ChannelState => ({
      items: [],
      total: 0,
      page: 1,
      pageSize: 10,
      loading: false,
      saving: false,
      current: null,
      detailLoading: false,
      alerts: [],
      tasks: [],
      notes: [],
      syncHistory: [],
      filters: {},
      platforms: [],
      channelTypes: [],
      regions: [],
      countries: [],
      owners: [],
      ownersLoading: false,
      platformCatalogLoading: false,
    }),
  actions: {
    async fetchPlatformCatalog(force = false) {
      if (
        !force &&
        this.platforms.length > 0 &&
        this.channelTypes.length > 0 &&
        this.regions.length > 0 &&
        this.countries.length > 0
      ) {
        return {
          platforms: this.platforms,
          channelTypes: this.channelTypes,
          regions: this.regions,
          countries: this.countries,
        }
      }
      const api = useChannelsApi()
      this.platformCatalogLoading = true
      try {
        const catalog = await api.listChannelPlatforms()
        this.platforms = catalog.platforms ?? []
        this.channelTypes = catalog.channelTypes ?? []
        this.regions = catalog.regions ?? []
        this.countries = catalog.countries ?? []
        return catalog
      } finally {
        this.platformCatalogLoading = false
      }
    },
    async fetchList(params: ChannelListParams = {}) {
      const api = useChannelsApi()
      this.loading = true
      try {
        const merged: ChannelListParams = {
          ...this.filters,
          page: params.page ?? this.page,
          pageSize: params.pageSize ?? this.pageSize,
          ...params,
        }
        const { items, meta } = await api.listChannels(merged)
        this.items = items ?? []
        this.total = meta?.total ?? this.items.length
        this.page = meta?.page ?? merged.page ?? 1
        this.pageSize = meta?.pageSize ?? merged.pageSize ?? this.pageSize
        this.filters = { ...this.filters, ...params }
        return this.items
      } finally {
        this.loading = false
      }
    },
    async fetchOwners(keyword = '', limit = 20) {
      const api = useChannelsApi()
      this.ownersLoading = true
      try {
        const response = await api.listChannelOwners({
          keyword: keyword || undefined,
          limit,
        })
        const resolved = Array.isArray(response?.items)
          ? response.items
          : Array.isArray(response?.data?.items)
            ? response.data.items ?? []
            : []
        this.owners = resolved
        return this.owners
      } finally {
        this.ownersLoading = false
      }
    },
    setPage(page: number) {
      this.page = page
    },
    setPageSize(size: number) {
      this.pageSize = size
    },
    setFilters(partial: ChannelListParams) {
      this.filters = { ...this.filters, ...partial }
    },
    async createChannel(payload: ChannelDraftPayload) {
      const api = useChannelsApi()
      this.saving = true
      try {
        const detail = await api.createChannel(payload)
        this.current = detail
        await this.fetchList({ page: this.page })
        return detail
      } finally {
        this.saving = false
      }
    },
    async updateChannel(channelId: string, payload: ChannelDraftPayload) {
      const api = useChannelsApi()
      this.saving = true
      try {
        const detail = await api.updateChannel(channelId, payload)
        this.current = detail
        await this.fetchList({ page: this.page })
        return detail
      } finally {
        this.saving = false
      }
    },
    async submitChannel(channelId: string, payload?: ChannelSubmitPayload) {
      const api = useChannelsApi()
      this.saving = true
      try {
        const detail = await api.submitChannel(channelId, payload)
        this.current = detail
        await this.fetchList({ page: this.page })
        return detail
      } finally {
        this.saving = false
      }
    },
    async decideApproval(channelId: string, payload: ChannelApprovalPayload) {
      const api = useChannelsApi()
      this.saving = true
      try {
        const detail = await api.decideApproval(channelId, payload)
        return detail
      } finally {
        this.saving = false
      }
    },
    async fetchApprovals(params: ChannelListParams = {}) {
      const api = useChannelsApi()
      const query: ChannelListParams = {
        status: ['pending_review'],
        pageSize: 20,
        ...params,
      }
      return api.listChannels(query)
    },
    async fetchCredentials(channelId: string) {
      const api = useChannelsApi()
      const { items } = await api.listChannelCredentials(channelId)
      return items ?? []
    },
    async saveCredential(channelId: string, payload: ChannelCredentialUpsertPayload) {
      const api = useChannelsApi()
      return api.upsertChannelCredential(channelId, payload)
    },
    async testCredential(channelId: string, payload: ChannelCredentialTestPayload) {
      const api = useChannelsApi()
      return api.testChannelCredential(channelId, payload)
    },
    async fetchDetail(channelId: string) {
      const api = useChannelsApi()
      this.detailLoading = true
      try {
        const detail = await api.getChannelDetail(channelId)
        this.current = detail
        this.alerts = detail.alerts ?? []
        this.tasks = detail.tasks ?? []
        this.notes = detail.notes ?? []
        this.syncHistory = detail.syncHistory ?? []
        return detail
      } finally {
        this.detailLoading = false
      }
    },
    async refreshAlerts(channelId: string) {
      const api = useChannelsApi()
      const { items } = await api.listChannelAlerts(channelId)
      this.alerts = items ?? []
      return this.alerts
    },
    async updateAlert(channelId: string, alertId: string, payload: ChannelAlertUpdatePayload) {
      const api = useChannelsApi()
      await api.updateChannelAlert(channelId, alertId, payload)
      await this.refreshAlerts(channelId)
    },
    async fetchTasks(channelId: string) {
      const api = useChannelsApi()
      const { items } = await api.listChannelTasks(channelId)
      this.tasks = items ?? []
      return this.tasks
    },
    async linkTask(channelId: string, payload: ChannelTaskLinkPayload) {
      const api = useChannelsApi()
      await api.linkChannelTask(channelId, payload)
      await this.fetchTasks(channelId)
    },
    async updateTask(channelId: string, taskLinkId: string, payload: ChannelTaskLinkPayload) {
      const api = useChannelsApi()
      await api.updateChannelTask(channelId, taskLinkId, payload)
      await this.fetchTasks(channelId)
    },
    async removeTask(channelId: string, taskLinkId: string) {
      const api = useChannelsApi()
      await api.removeChannelTask(channelId, taskLinkId)
      await this.fetchTasks(channelId)
    },
    async fetchNotes(channelId: string) {
      const api = useChannelsApi()
      const { items } = await api.listChannelNotes(channelId)
      this.notes = items ?? []
      return this.notes
    },
    async createNote(channelId: string, payload: ChannelNotePayload) {
      const api = useChannelsApi()
      await api.createChannelNote(channelId, payload)
      await this.fetchNotes(channelId)
    },
    async triggerSync(channelId: string, payload?: ChannelSyncTriggerPayload) {
      const api = useChannelsApi()
      await api.triggerChannelSync(channelId, payload)
      await this.fetchSyncHistory(channelId)
    },
    async fetchSyncHistory(channelId: string) {
      const api = useChannelsApi()
      const { items } = await api.listChannelSyncHistory(channelId)
      this.syncHistory = items ?? []
      return this.syncHistory
    },
    async fetchStrategy(channelId: string) {
      const api = useChannelsApi()
      const snapshot = await api.getChannelStrategy(channelId)
      if (this.current) {
        this.current.strategy = snapshot.strategy
        this.current.team = snapshot.team
      }
      return snapshot
    },
    async saveStrategy(channelId: string, payload: ChannelStrategyUpdatePayload) {
      const api = useChannelsApi()
      const snapshot = await api.updateChannelStrategy(channelId, payload)
      if (this.current) {
        this.current.strategy = snapshot.strategy
        this.current.team = snapshot.team
      }
      return snapshot
    },
    setCurrent(detail: ChannelDetail | null) {
      this.current = detail
    },
  },
})
