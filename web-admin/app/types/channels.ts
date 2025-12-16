export type ChannelStatus = 'draft' | 'pending_review' | 'rejected' | 'unauthorized' | 'authorized' | 'disabled'

export interface ChannelContactPayload {
  name: string
  phone: string
  email: string
}

export interface ChannelSummary {
  id: string
  name: string
  storeId?: string
  platform: string
  region: string
  ownerUuid: string
  status: ChannelStatus
  channelType: string
  tags: string[]
}

export interface ChannelListMeta {
  total: number
  page: number
  pageSize: number
}

export interface ChannelListResponse {
  items: ChannelSummary[]
  meta: ChannelListMeta
}

export type ChannelListApiResponse =
  | ChannelListResponse
  | {
      success?: boolean
      data?: ChannelListResponse
    }

export interface ChannelListParams {
  keyword?: string
  platform?: string
  status?: string[]
  owner?: string
  region?: string
  tags?: string[]
  minHealth?: number
  minGmv?: number
  page?: number
  pageSize?: number
}

export interface ChannelPlatform {
  code: string
  label: string
  description?: string
  channelTypes: string[]
  regions?: string[]
  deprecated?: boolean
}

export interface ChannelChannelType {
  code: string
  label: string
  description?: string
}

export interface ChannelRegion {
  code: string
  label: string
  description?: string
}

export interface ChannelCity {
  code: string
  label: string
  countryCode?: string
}

export interface ChannelCountry {
  code: string
  label: string
  description?: string
  cities: ChannelCity[]
}

export interface ChannelOwnerOption {
  id: number
  username: string
  displayName: string
  email?: string
}

export interface ChannelOwnerListResponse {
  success?: boolean
  request_id?: string
  timestamp?: string
  items?: ChannelOwnerOption[]
  data?: {
    items?: ChannelOwnerOption[]
  }
}

export interface ChannelPlatformCatalog {
  platforms: ChannelPlatform[]
  channelTypes: ChannelChannelType[]
  regions: ChannelRegion[]
  countries: ChannelCountry[]
}

export interface ChannelDraftPayload {
  name: string
  platform: string
  storeId: string
  region: string
  ownerUuid: string
  approverUuid?: string
  channelType: string
  domain?: string
  tags?: string[]
  contact: ChannelContactPayload
}

export interface ChannelStrategyConfig {
  pricebookId?: string
  inventoryStrategyId?: string
  logisticsStrategyId?: string
  csSlaId?: string
  feeRate?: number
  settlementCycle?: string
  paymentTerms?: string
  notes?: string
  effectiveAt?: string
}

export interface ChannelTeamConfig {
  ownerUuid: string
  approverUuid?: string
  operators?: string[]
}

export interface ChannelStrategySnapshot {
  strategy: ChannelStrategyConfig
  team: ChannelTeamConfig
}

export interface ChannelStrategyUpdatePayload {
  strategy: ChannelStrategyConfig
  team: ChannelTeamConfig
}

export interface ChannelSubmitPayload {
  note?: string
  offlineEvidenceUrl?: string
}

export type ChannelApprovalDecision = 'approve' | 'reject'

export interface ChannelApprovalPayload {
  decision: ChannelApprovalDecision
  reason?: string
}

export interface ChannelMetric {
  window: string
  gmv: number
  orders: number
  gmvGrowthRate: number
  inventoryCoverage: number
  errorRate: number
  syncSuccessRate: number
  healthScore: number
  sourceTimestamp?: string
}

export interface ChannelHealth {
  score: number
  labels: string[]
}

export interface ChannelAlert {
  id: string
  type: string
  severity: string
  title: string
  description?: string
  status: string
  triggeredAt: string
  resolvedAt?: string
}

export interface ChannelTaskLink {
  id: string
  taskId: string
  taskSource: string
  status: 'open' | 'in_progress' | 'done'
  note?: string
  linkedBy: string
  linkedAt: string
}

export interface ChannelNote {
  id: string
  authorUuid: string
  visibility: string
  body: string
  createdAt: string
}

export interface ChannelSyncHistoryItem {
  id: string
  triggerType: string
  triggeredBy: string
  durationMs: number
  result: string
  createdAt: string
}

export interface ChannelDetail extends ChannelSummary {
  domain?: string
  approverUuid?: string
  contact?: ChannelContactPayload
  lastSyncAt?: string
  syncStatus?: string
  health?: ChannelHealth
  metrics?: ChannelMetric[]
  alerts?: ChannelAlert[]
  tasks?: ChannelTaskLink[]
  notes?: ChannelNote[]
  syncHistory?: ChannelSyncHistoryItem[]
  strategy?: ChannelStrategyConfig
  team?: ChannelTeamConfig
}

export const createEmptyChannelPayload = (): ChannelDraftPayload => ({
  name: '',
  platform: '',
  storeId: '',
  region: '',
  ownerUuid: '',
  approverUuid: '',
  channelType: '',
  domain: '',
  tags: [],
  contact: {
    name: '',
    phone: '',
    email: '',
  },
})

export interface ChannelCredential {
  id: string
  type: string
  status: string
  scope: string[]
  expiresAt?: string
  lastRefreshedAt?: string
  lastTestedAt?: string
  testResult?: Record<string, any>
  attachmentUrl?: string
}

export interface ChannelCredentialUpsertPayload {
  type: string
  payload: Record<string, any>
  scope?: string[]
  expiresAt?: string
  metadata?: Record<string, any>
  attachmentUrl?: string
}

export interface ChannelCredentialTestPayload {
  type: string
  succeeded?: boolean
  result?: Record<string, any>
}

export const createEmptyCredentialPayload = (): ChannelCredentialUpsertPayload => ({
  type: 'oauth',
  payload: {},
  scope: [],
  metadata: {},
  attachmentUrl: '',
})

export const createEmptyStrategyPayload = (): ChannelStrategyUpdatePayload => ({
  strategy: {
    pricebookId: '',
    inventoryStrategyId: '',
    logisticsStrategyId: '',
    csSlaId: '',
    feeRate: 0,
    settlementCycle: '',
    paymentTerms: '',
    notes: '',
    effectiveAt: '',
  },
  team: {
    ownerUuid: '',
    approverUuid: '',
    operators: [],
  },
})

export interface ChannelAlertUpdatePayload {
  status?: string
  assigneeUuid?: string
  taskId?: string
  note?: string
}

export interface ChannelTaskLinkPayload {
  taskId: string
  taskSource?: string
  status?: string
  note?: string
}

export interface ChannelNotePayload {
  body: string
  visibility?: string
}

export interface ChannelSyncTriggerPayload {
  triggerType?: string
  payload?: Record<string, any>
}
