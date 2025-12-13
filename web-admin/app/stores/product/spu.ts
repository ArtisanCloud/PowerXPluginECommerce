import { defineStore } from 'pinia'
import type {
	ApprovalActionPayload,
	PublishSpuPayload,
	RollbackPayload,
	SpuDetail,
	SpuListParams,
	SpuSkuLink,
	SpuSummary,
	SpuVersionDetail,
	SpuVersionSummary,
	SubmitSpuPayload,
	WithdrawSpuPayload,
	DeleteSpuPayload,
	VersionListParams,
} from '~/composables/api/useSpu'
import { useSpuApi } from '~/composables/api/useSpu'

interface SpuState {
	items: SpuSummary[]
	total: number
	loading: boolean
	detailLoading: boolean
	current: SpuDetail | null
	skus: SpuSkuLink[]
	skusLoading: boolean
	versions: SpuVersionSummary[]
	versionsLoading: boolean
	versionLoading: boolean
	versionDetail: SpuVersionDetail | null
}

export const useSpuStore = defineStore('product-spu', {
	state: (): SpuState => ({
		items: [],
		total: 0,
		loading: false,
		detailLoading: false,
		current: null,
		skus: [],
		skusLoading: false,
		versions: [],
		versionsLoading: false,
		versionLoading: false,
		versionDetail: null,
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
		async fetchDetail(id: string) {
			const api = useSpuApi()
			this.detailLoading = true
			try {
				const detail = await api.getSpu(id)
				this.current = detail
				return detail
			} finally {
				this.detailLoading = false
			}
		},
		async create(payload: Record<string, any>) {
			const api = useSpuApi()
			const detail = await api.createSpu(payload)
			await this.fetchList()
			return detail
		},
		async update(id: string, payload: Record<string, any>) {
			const api = useSpuApi()
			const detail = await api.updateSpu(id, payload)
			this.current = detail
			await this.fetchList()
			return detail
		},
		async submit(id: string, payload?: SubmitSpuPayload) {
			const api = useSpuApi()
			const detail = await api.submitSpu(id, payload)
			this.current = detail
			await this.fetchList()
			return detail
		},
		async publish(id: string, payload: PublishSpuPayload) {
			const api = useSpuApi()
			const detail = await api.publishSpu(id, payload)
			this.current = detail
			await this.fetchList()
			return detail
		},
		async withdraw(id: string, payload: WithdrawSpuPayload) {
			const api = useSpuApi()
			const detail = await api.withdrawSpu(id, payload)
			this.current = detail
			await this.fetchList()
			return detail
		},
		async delete(id: string, payload: DeleteSpuPayload) {
			const api = useSpuApi()
			const detail = await api.deleteSpu(id, payload)
			this.current = null
			await this.fetchList()
			return detail
		},
		async fetchSkus(id: string) {
			const api = useSpuApi()
			this.skusLoading = true
			try {
				const { items } = await api.listSpuSkus(id)
				this.skus = items ?? []
				return this.skus
			} finally {
				this.skusLoading = false
			}
		},
		async saveSkus(id: string, items: SpuSkuLink[]) {
			const api = useSpuApi()
			const { items: linked } = await api.replaceSpuSkus(id, items)
			this.skus = linked ?? []
			return this.skus
		},
		async fetchVersions(id: string, params?: VersionListParams) {
			const api = useSpuApi()
			this.versionsLoading = true
			try {
				const list = await api.listSpuVersions(id, params)
				this.versions = list?.items ?? []
				return this.versions
			} finally {
				this.versionsLoading = false
			}
		},
		async fetchVersionDetail(id: string, versionId: string) {
			const api = useSpuApi()
			this.versionLoading = true
			try {
				const detail = await api.getSpuVersion(id, versionId)
				this.versionDetail = detail
				return detail
			} finally {
				this.versionLoading = false
			}
		},
		async approveVersion(id: string, versionId: string, payload?: ApprovalActionPayload) {
			const api = useSpuApi()
			const detail = await api.approveSpuVersion(id, versionId, payload)
			this.versionDetail = detail
			await this.fetchVersions(id)
			return detail
		},
		async rejectVersion(id: string, versionId: string, payload?: ApprovalActionPayload) {
			const api = useSpuApi()
			const detail = await api.rejectSpuVersion(id, versionId, payload)
			this.versionDetail = detail
			await this.fetchVersions(id)
			return detail
		},
		async rollbackVersion(id: string, versionId: string, payload: RollbackPayload) {
			const api = useSpuApi()
			const detail = await api.rollbackSpuVersion(id, versionId, payload)
			this.versionDetail = detail
			await this.fetchVersions(id)
			await this.fetchDetail(id)
			return detail
		},
	},
})
