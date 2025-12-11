import { defineStore } from 'pinia'
import type { SpuListParams, SpuSummary } from '~/composables/useSpuApi'
import { useSpuApi } from '~/composables/useSpuApi'

interface SpuState {
  items: SpuSummary[]
  total: number
  loading: boolean
}

export const useSpuStore = defineStore('product-spu', {
  state: (): SpuState => ({
    items: [],
    total: 0,
    loading: false,
  }),
  actions: {
    async fetchList(params?: SpuListParams) {
      const api = useSpuApi()
      this.loading = true
      try {
        const { items, meta } = await api.listSpus(params)
        this.items = items
        this.total = meta?.total ?? items.length
      } finally {
        this.loading = false
      }
    },
    async create(payload: Record<string, any>) {
      const api = useSpuApi()
      await api.createSpu(payload)
      await this.fetchList()
    },
  },
})
