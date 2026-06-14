import { createPluginWsBusClient, type PluginWsBusClient } from '@artisan-cloud/plugin-framework-client'
import { getAuthToken, getTenantUuid, resolveApiBase } from '~/composables/api/_base'
import { PLUGIN_ID } from '~/utils/powerx-bridge'

export type WsBusEvent = {
  type?: string
  topic?: string
  payload?: Record<string, any>
  [key: string]: any
}

type WsEventHandler = (event: WsBusEvent) => void

type WsBusClient = {
  connect: () => void
  disconnect: () => void
  subscribe: (topic: string, handler: WsEventHandler) => void
  unsubscribe: (topic: string, handler: WsEventHandler) => void
  isConnected: () => boolean
}

const DEFAULT_TOPICS = [
  'task.progress',
  'powerx.task.progress.v1',
  'worker.task.updated',
  'org_sync.progress',
  'powerx.org_sync.progress.v1',
] as const

const topicHandlers = new Map<string, Set<WsEventHandler>>()
const subscribedTopics = new Set<string>()
let client: PluginWsBusClient | null = null

const normalizeTopic = (topic?: string) => String(topic || '').trim()

const createClient = () => {
  if (client) return client
  const cfg = typeof useRuntimeConfig === 'function' ? useRuntimeConfig() : undefined
  const publicCfg = (cfg?.public || {}) as Record<string, any>
  client = createPluginWsBusClient({
    pluginId: PLUGIN_ID,
    apiBaseURL: resolveApiBase(),
    hostBaseURL: String(publicCfg.powerxCoreBase || publicCfg.apiBaseUrl || ''),
    wsBaseURL: String(publicCfg.wsBaseUrl || ''),
    wsPath: '/api/ws',
    insidePowerX: Boolean(publicCfg.insidePowerX),
    token: getAuthToken(),
    tenantUuid: getTenantUuid(),
    reconnectIntervalMs: 3000,
    onEvent: dispatchEvent,
  })
  return client
}

const desiredTopics = () => {
  const merged = new Set<string>()
  DEFAULT_TOPICS.forEach((topic) => {
    const normalized = normalizeTopic(topic)
    if (normalized) merged.add(normalized)
  })
  Array.from(topicHandlers.keys()).forEach((topic) => {
    const normalized = normalizeTopic(topic)
    if (normalized) merged.add(normalized)
  })
  return Array.from(merged)
}

const refreshContext = () => {
  const bus = createClient()
  bus.setContext({
    token: getAuthToken(),
    tenantUuid: getTenantUuid(),
  })
  return bus
}

const ensureSubscriptions = () => {
  const bus = refreshContext()
  const missing = desiredTopics().filter((topic) => !subscribedTopics.has(topic))
  if (missing.length === 0) return
  bus.subscribe(missing)
  missing.forEach((topic) => subscribedTopics.add(topic))
}

function dispatchEvent(message: WsBusEvent) {
  const topic = normalizeTopic(message.topic)
  if (!topic) return
  const handlers = topicHandlers.get(topic)
  if (!handlers || handlers.size === 0) return
  handlers.forEach((handler) => {
    try {
      handler(message)
    } catch (error) {
      console.warn('[ws-bus] handler failed', error)
    }
  })
}

export function useWsBusClient(): WsBusClient {
  const connect = () => {
    if (typeof window === 'undefined') return
    const bus = refreshContext()
    bus.connect()
    ensureSubscriptions()
  }

  const disconnect = () => {
    client?.disconnect()
    subscribedTopics.clear()
  }

  const subscribe = (topic: string, handler: WsEventHandler) => {
    const normalized = normalizeTopic(topic)
    if (!normalized || !handler) return
    if (!topicHandlers.has(normalized)) {
      topicHandlers.set(normalized, new Set<WsEventHandler>())
    }
    topicHandlers.get(normalized)!.add(handler)
    ensureSubscriptions()
  }

  const unsubscribe = (topic: string, handler: WsEventHandler) => {
    const normalized = normalizeTopic(topic)
    if (!normalized) return
    const handlers = topicHandlers.get(normalized)
    if (!handlers) return
    handlers.delete(handler)
    if (handlers.size === 0) {
      topicHandlers.delete(normalized)
      subscribedTopics.delete(normalized)
      client?.unsubscribe([normalized])
    }
  }

  const isConnected = () => Boolean(client?.getState().connected)

  return {
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    isConnected,
  }
}
