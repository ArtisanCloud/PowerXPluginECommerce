import type { FetchOptions } from 'ofetch'

export interface SpuListParams {
  keyword?: string
  status?: string
  page?: number
  pageSize?: number
}

export interface SpuSummary {
  id: string
  name: string
  code: string
  status: string
  type: string
  updatedAt?: string
}

export const useSpuApi = () => {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBaseUrl || '/api'

  const request = async <T>(url: string, options: FetchOptions<'json'> = {}) => {
    return await $fetch<T>(url, {
      baseURL,
      ...options,
    })
  }

  return {
    listSpus: (params?: SpuListParams) => request<{ items: SpuSummary[]; meta: { total: number } }>('product/spus', { params }),
    createSpu: (payload: Record<string, any>) => request('product/spus', { method: 'POST', body: payload }),
    getSpu: (id: string) => request<SpuSummary>(`product/spus/${id}`),
  }
}
