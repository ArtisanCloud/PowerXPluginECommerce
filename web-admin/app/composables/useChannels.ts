import { apiDel, apiGet, apiPatch, apiPost } from '~/composables/api'
import type {
  ChannelAlert,
  ChannelAlertUpdatePayload,
  ChannelApprovalPayload,
  ChannelCredential,
  ChannelCredentialTestPayload,
  ChannelCredentialUpsertPayload,
  ChannelDetail,
  ChannelDraftPayload,
  ChannelListApiResponse,
  ChannelListParams,
  ChannelListResponse,
  ChannelNote,
  ChannelNotePayload,
  ChannelSubmitPayload,
  ChannelSyncHistoryItem,
  ChannelTaskLink,
  ChannelTaskLinkPayload,
  ChannelSyncTriggerPayload,
  ChannelStrategySnapshot,
  ChannelStrategyUpdatePayload,
  ChannelPlatformCatalog,
  ChannelOwnerListResponse,
} from '~/types/channels'

const API_PREFIX = 'admin/channels'

const unwrapApiData = <T>(response: T | { data?: T } | null | undefined): T => {
  if (response && typeof response === 'object' && 'data' in response) {
    const data = (response as { data?: T }).data
    if (typeof data !== 'undefined') {
      return data as T
    }
  }
  return response as T
}

export const useChannelsApi = () => {
  const listChannels = async (params?: ChannelListParams) => {
    const response = await apiGet<ChannelListApiResponse>(API_PREFIX, params)
    return unwrapApiData<ChannelListResponse>(response)
  }

  const createChannel = async (payload: ChannelDraftPayload) => {
    const response = await apiPost<ChannelDetail | { data?: ChannelDetail }>(API_PREFIX, payload)
    return unwrapApiData<ChannelDetail>(response)
  }

  const updateChannel = async (id: string, payload: ChannelDraftPayload) => {
    const response = await apiPatch<ChannelDetail | { data?: ChannelDetail }>(`${API_PREFIX}/${id}`, payload)
    return unwrapApiData<ChannelDetail>(response)
  }

  const submitChannel = async (id: string, payload?: ChannelSubmitPayload) => {
    const response = await apiPost<ChannelDetail | { data?: ChannelDetail }>(`${API_PREFIX}/${id}/submit`, payload ?? {})
    return unwrapApiData<ChannelDetail>(response)
  }

  const decideApproval = async (id: string, payload: ChannelApprovalPayload) => {
    const response = await apiPost<ChannelDetail | { data?: ChannelDetail }>(`${API_PREFIX}/${id}/approval`, payload)
    return unwrapApiData<ChannelDetail>(response)
  }

  const getChannelDetail = async (id: string) => {
    const response = await apiGet<ChannelDetail | { data?: ChannelDetail }>(`${API_PREFIX}/${id}`)
    return unwrapApiData<ChannelDetail>(response)
  }

  const listChannelCredentials = async (channelId: string) => {
    const response = await apiGet<{ items?: ChannelCredential[] } | { data?: { items?: ChannelCredential[] } }>(
      `${API_PREFIX}/${channelId}/credentials`,
    )
    const payload = unwrapApiData<{ items?: ChannelCredential[] }>(response)
    return { items: payload?.items ?? [] }
  }

  const upsertChannelCredential = async (channelId: string, payload: ChannelCredentialUpsertPayload) => {
    const response = await apiPost<ChannelCredential | { data?: ChannelCredential }>(
      `${API_PREFIX}/${channelId}/credentials`,
      payload,
    )
    return unwrapApiData<ChannelCredential>(response)
  }

  const testChannelCredential = async (channelId: string, payload: ChannelCredentialTestPayload) => {
    const response = await apiPost<ChannelCredential | { data?: ChannelCredential }>(
      `${API_PREFIX}/${channelId}/credentials/test`,
      payload,
    )
    return unwrapApiData<ChannelCredential>(response)
  }

  const listChannelAlerts = async (channelId: string) => {
    const response = await apiGet<{ items?: ChannelAlert[] } | { data?: { items?: ChannelAlert[] } }>(
      `${API_PREFIX}/${channelId}/alerts`,
    )
    const payload = unwrapApiData<{ items?: ChannelAlert[] }>(response)
    return { items: payload?.items ?? [] }
  }

  const updateChannelAlert = async (channelId: string, alertId: string, payload: ChannelAlertUpdatePayload) => {
    return apiPatch(`${API_PREFIX}/${channelId}/alerts/${alertId}`, payload)
  }

  const listChannelTasks = async (channelId: string) => {
    const response = await apiGet<{ items?: ChannelTaskLink[] } | { data?: { items?: ChannelTaskLink[] } }>(
      `${API_PREFIX}/${channelId}/tasks`,
    )
    const payload = unwrapApiData<{ items?: ChannelTaskLink[] }>(response)
    return { items: payload?.items ?? [] }
  }

  const linkChannelTask = async (channelId: string, payload: ChannelTaskLinkPayload) => {
    const response = await apiPost<ChannelTaskLink | { data?: ChannelTaskLink }>(`${API_PREFIX}/${channelId}/tasks`, payload)
    return unwrapApiData<ChannelTaskLink>(response)
  }

  const updateChannelTask = async (channelId: string, taskLinkId: string, payload: ChannelTaskLinkPayload) => {
    return apiPatch(`${API_PREFIX}/${channelId}/tasks/${taskLinkId}`, payload)
  }

  const removeChannelTask = async (channelId: string, taskLinkId: string) => {
    return apiDel(`${API_PREFIX}/${channelId}/tasks/${taskLinkId}`)
  }

  const listChannelNotes = async (channelId: string) => {
    const response = await apiGet<{ items?: ChannelNote[] } | { data?: { items?: ChannelNote[] } }>(
      `${API_PREFIX}/${channelId}/notes`,
    )
    const payload = unwrapApiData<{ items?: ChannelNote[] }>(response)
    return { items: payload?.items ?? [] }
  }

  const createChannelNote = async (channelId: string, payload: ChannelNotePayload) => {
    const response = await apiPost<ChannelNote | { data?: ChannelNote }>(`${API_PREFIX}/${channelId}/notes`, payload)
    return unwrapApiData<ChannelNote>(response)
  }

  const triggerChannelSync = async (channelId: string, payload?: ChannelSyncTriggerPayload) => {
    return apiPost(`${API_PREFIX}/${channelId}/sync`, payload ?? {})
  }

  const listChannelSyncHistory = async (channelId: string) => {
    const response = await apiGet<{ items?: ChannelSyncHistoryItem[] } | { data?: { items?: ChannelSyncHistoryItem[] } }>(
      `${API_PREFIX}/${channelId}/sync-history`,
    )
    const payload = unwrapApiData<{ items?: ChannelSyncHistoryItem[] }>(response)
    return { items: payload?.items ?? [] }
  }

  const getChannelStrategy = async (channelId: string) => {
    const response = await apiGet<ChannelStrategySnapshot | { data?: ChannelStrategySnapshot }>(
      `${API_PREFIX}/${channelId}/strategy`,
    )
    return unwrapApiData<ChannelStrategySnapshot>(response)
  }

  const updateChannelStrategy = async (channelId: string, payload: ChannelStrategyUpdatePayload) => {
    const response = await apiPatch<ChannelStrategySnapshot | { data?: ChannelStrategySnapshot }>(
      `${API_PREFIX}/${channelId}/strategy`,
      payload,
    )
    return unwrapApiData<ChannelStrategySnapshot>(response)
  }

  const listChannelPlatforms = async () => {
    const response = await apiGet<ChannelPlatformCatalog | { data?: ChannelPlatformCatalog }>(`${API_PREFIX}/platforms`)
    return unwrapApiData<ChannelPlatformCatalog>(response)
  }

  const listChannelOwners = async (params?: { keyword?: string; limit?: number }) => {
    return apiGet<ChannelOwnerListResponse>(`${API_PREFIX}/owners`, params)
  }

  return {
    listChannels,
    createChannel,
    updateChannel,
    submitChannel,
    decideApproval,
    getChannelDetail,
    listChannelCredentials,
    upsertChannelCredential,
    testChannelCredential,
    listChannelAlerts,
    updateChannelAlert,
    listChannelTasks,
    linkChannelTask,
    updateChannelTask,
    removeChannelTask,
    listChannelNotes,
    createChannelNote,
    triggerChannelSync,
    listChannelSyncHistory,
    getChannelStrategy,
    updateChannelStrategy,
    listChannelPlatforms,
    listChannelOwners,
  }
}
