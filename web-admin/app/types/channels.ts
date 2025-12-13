export type ChannelStatus = 'draft' | 'pending_review' | 'rejected' | 'unauthorized' | 'authorized' | 'disabled'

export interface ChannelSummary {
  id: string
  name: string
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

export interface ChannelListParams {
  keyword?: string
  platform?: string
  status?: string[]
  owner?: string
  region?: string
  page?: number
  pageSize?: number
}

export interface ChannelDraftPayload {
  name: string
  platform: string
  storeId: string
  region: string
  ownerUuid: string
  channelType: string
  tags?: string[]
}

export interface ChannelDetail extends ChannelSummary {
  contactName?: string
  contactPhone?: string
  contactEmail?: string
  lastSyncAt?: string
  syncStatus?: string
}

export interface ChannelTaskLink {
  id: string
  title: string
  description?: string
  status: 'open' | 'in_progress' | 'done'
}

export interface ChannelNote {
  id: string
  author: string
  body: string
  createdAt: string
}

export interface ChannelSyncHistoryItem {
  createdAt: string
  triggerType: string
  triggeredBy: string
  duration: string
  result: string
}
