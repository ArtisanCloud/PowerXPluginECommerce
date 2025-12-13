import { defineStore } from 'pinia'
import type { ChannelDetail, ChannelDraftPayload, ChannelListParams, ChannelSummary } from '~/app/types/channels'
import { useChannelsApi } from '~/app/composables/useChannels'

interface ChannelState {
  items: ChannelSummary[]
  total: number
  loading: boolean
  creating: boolean
  current: ChannelDetail | null
}

export const useChannelsStore = defineStore('channels', {
  state: (): ChannelState => ({
    items: [],
    total: 0,
    loading: false,
    creating: false,
    current: null,
  }),
  actions: {
    async fetchList(params?: ChannelListParams) {
      const api = useChannelsApi()
      this.loading = true
      try {
        const { items, meta } = await api.listChannels(params)
        this.items = items ?? []
        this.total = meta?.total ?? this.items.length
        return this.items
      } finally {
        this.loading = false
      }
    },
    async createChannel(payload: ChannelDraftPayload) {
      const api = useChannelsApi()
      this.creating = true
      try {
        const detail = await api.createChannel(payload)
        this.current = detail
        await this.fetchList()
        return detail
      } finally {
        this.creating = false
      }
    },
    setCurrent(detail: ChannelDetail | null) {
      this.current = detail
    },
  },
})
