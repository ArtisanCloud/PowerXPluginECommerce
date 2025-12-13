import type {
  ChannelDetail,
  ChannelDraftPayload,
  ChannelListParams,
  ChannelListResponse,
} from '~/app/types/channels'

const API_PREFIX = '/channels'

export const useChannelsApi = () => {
  const runtimeConfig = useRuntimeConfig()
  const baseURL = runtimeConfig.public?.apiBaseUrl ?? '/_p/com.powerx.plugin.ecommerce/api/v1'
  const client = $fetch.create({ baseURL })

  const listChannels = async (params?: ChannelListParams) => {
    return client<ChannelListResponse>(API_PREFIX, { method: 'GET', query: params })
  }

  const createChannel = async (payload: ChannelDraftPayload) => {
    return client<ChannelDetail>(API_PREFIX, { method: 'POST', body: payload })
  }

  return { listChannels, createChannel }
}
