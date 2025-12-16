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

export const useChannelsApi = () => {
  const listChannels = async (params?: ChannelListParams) => {
    const response = await apiGet<ChannelListApiResponse>(API_PREFIX, params)
    if (response && typeof response === 'object' && 'data' in response && response.data) {
      return response.data
    }
    return response as ChannelListResponse
  }

  const createChannel = async (payload: ChannelDraftPayload) => {
    return apiPost<ChannelDetail>(API_PREFIX, payload)
  }

  const updateChannel = async (id: string, payload: ChannelDraftPayload) => {
    return apiPatch<ChannelDetail>(`${API_PREFIX}/${id}`, payload)
  }

  const submitChannel = async (id: string, payload?: ChannelSubmitPayload) => {
    return apiPost<ChannelDetail>(`${API_PREFIX}/${id}/submit`, payload ?? {})
  }

  const decideApproval = async (id: string, payload: ChannelApprovalPayload) => {
    return apiPost<ChannelDetail>(`${API_PREFIX}/${id}/approval`, payload)
  }

  const getChannelDetail = async (id: string) => {
    return apiGet<ChannelDetail>(`${API_PREFIX}/${id}`)
  }

  const listChannelCredentials = async (channelId: string) => {
    return apiGet<{ items: ChannelCredential[] }>(`${API_PREFIX}/${channelId}/credentials`)
  }

  const upsertChannelCredential = async (channelId: string, payload: ChannelCredentialUpsertPayload) => {
    return apiPost<ChannelCredential>(`${API_PREFIX}/${channelId}/credentials`, payload)
  }

  const testChannelCredential = async (channelId: string, payload: ChannelCredentialTestPayload) => {
    return apiPost<ChannelCredential>(`${API_PREFIX}/${channelId}/credentials/test`, payload)
  }

  const listChannelAlerts = async (channelId: string) => {
    return apiGet<{ items: ChannelAlert[] }>(`${API_PREFIX}/${channelId}/alerts`)
  }

  const updateChannelAlert = async (channelId: string, alertId: string, payload: ChannelAlertUpdatePayload) => {
    return apiPatch(`${API_PREFIX}/${channelId}/alerts/${alertId}`, payload)
  }

  const listChannelTasks = async (channelId: string) => {
    return apiGet<{ items: ChannelTaskLink[] }>(`${API_PREFIX}/${channelId}/tasks`)
  }

  const linkChannelTask = async (channelId: string, payload: ChannelTaskLinkPayload) => {
    return apiPost<ChannelTaskLink>(`${API_PREFIX}/${channelId}/tasks`, payload)
  }

  const updateChannelTask = async (channelId: string, taskLinkId: string, payload: ChannelTaskLinkPayload) => {
    return apiPatch(`${API_PREFIX}/${channelId}/tasks/${taskLinkId}`, payload)
  }

  const removeChannelTask = async (channelId: string, taskLinkId: string) => {
    return apiDel(`${API_PREFIX}/${channelId}/tasks/${taskLinkId}`)
  }

  const listChannelNotes = async (channelId: string) => {
    return apiGet<{ items: ChannelNote[] }>(`${API_PREFIX}/${channelId}/notes`)
  }

  const createChannelNote = async (channelId: string, payload: ChannelNotePayload) => {
    return apiPost<ChannelNote>(`${API_PREFIX}/${channelId}/notes`, payload)
  }

  const triggerChannelSync = async (channelId: string, payload?: ChannelSyncTriggerPayload) => {
    return apiPost(`${API_PREFIX}/${channelId}/sync`, payload ?? {})
  }

  const listChannelSyncHistory = async (channelId: string) => {
    return apiGet<{ items: ChannelSyncHistoryItem[] }>(`${API_PREFIX}/${channelId}/sync-history`)
  }

  const getChannelStrategy = async (channelId: string) => {
    return apiGet<ChannelStrategySnapshot>(`${API_PREFIX}/${channelId}/strategy`)
  }

  const updateChannelStrategy = async (channelId: string, payload: ChannelStrategyUpdatePayload) => {
    return apiPatch<ChannelStrategySnapshot>(`${API_PREFIX}/${channelId}/strategy`, payload)
  }

  const listChannelPlatforms = async () => {
    return apiGet<ChannelPlatformCatalog>(`${API_PREFIX}/platforms`)
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
