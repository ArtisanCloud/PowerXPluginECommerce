import { getAuthToken } from '~/composables/api/_base'

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

let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempts = 0
const reconnectDelayMs = 3000

const DEFAULT_TOPICS = [
  'task.progress',
  'powerx.task.progress.v1',
  'worker.task.updated',
  'org_sync.progress',
  'powerx.org_sync.progress.v1',
] as const

const topicHandlers = new Map<string, Set<WsEventHandler>>()
const subscribedTopics = new Set<string>()

const normalizeTopic = (topic?: string) => String(topic || '').trim()

const resolveBrowserLocation = () => {
  if (typeof window === 'undefined') return null
  const location = (window as any)?.location
  if (!location || typeof location !== 'object') return null
  const protocol = String(location.protocol || '').trim()
  const host = String(location.host || '').trim()
  const pathname = String(location.pathname || '')
  if (!protocol || !host) return null
  return { protocol, host, pathname }
}

const isHostMode = () => {
  try {
    const cfg = useRuntimeConfig()
    if (Boolean(cfg.public?.insidePowerX)) return true
  } catch {
    // ignore
  }

  const browserLocation = resolveBrowserLocation()
  if (browserLocation) {
    const p = browserLocation.pathname
    // 兼容 env 未注入 insidePowerX 的场景：根据宿主嵌入路径兜底识别
    if (/\/_p\/[^/]+\/admin(?:\/|$)/.test(p)) {
      return true
    }
  }

  try {
    const cfg = useRuntimeConfig()
    const base = String(cfg.public?.pluginAdminBase || '').trim()
    if (base && browserLocation) {
      return browserLocation.pathname.startsWith(base.replace(/\/+$/, ''))
    }
  } catch {
    // ignore
  }

  return false
}

const resolveWsPath = () => '/api/ws'

const toBase64Url = (raw: string) => {
  try {
    return btoa(raw).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
  } catch {
    return ''
  }
}

const buildWsBearerProtocol = (token?: string | null) => {
  const raw = String(token || '').trim()
  if (!raw) return ''
  const normalized = /^Bearer\s+/i.test(raw) ? raw.replace(/^Bearer\s+/i, '').trim() : raw
  if (!normalized) return ''
  const encoded = toBase64Url(normalized)
  if (!encoded) return ''
  return `bearer.${encoded}`
}

const redactWsURL = (raw: string) => {
  try {
    const url = new URL(raw)
    if (url.searchParams.has('authorization')) {
      url.searchParams.set('authorization', 'Bearer ***')
    }
    return url.toString()
  } catch {
    return raw
  }
}

const wsOriginFromApiBase = (apiBase?: string | null) => {
  const raw = String(apiBase || '').trim()
  if (!raw) return null
  if (!/^https?:\/\//i.test(raw)) return null
  try {
    const apiURL = new URL(raw)
    const wsProtocol = apiURL.protocol === 'https:' ? 'wss:' : 'ws:'
    let pathname = apiURL.pathname.replace(/\/+$/, '')
    if (pathname.endsWith('/api/v1')) {
      pathname = pathname.replace(/\/api\/v1$/, '/api/ws')
    } else if (pathname.endsWith('/api')) {
      pathname = `${pathname}/ws`
    } else if (pathname.length > 0) {
      pathname = `${pathname}/ws`
    } else {
      pathname = '/api/ws'
    }
    return `${wsProtocol}//${apiURL.host}${pathname}`
  } catch {
    return null
  }
}

const resolveWsURL = () => {
  const browserLocation = resolveBrowserLocation()
  if (!browserLocation) return ''

  const protocol = browserLocation.protocol === 'https:' ? 'wss:' : 'ws:'
  // 与其他插件对齐：默认同源 /api/ws，经由当前站点代理转发。
  let wsURL = new URL(resolveWsPath(), `${protocol}//${browserLocation.host}`).toString()
  let mode = isHostMode() ? 'host-same-origin' : 'same-origin'

  // 允许通过 runtimeConfig.public.wsBaseUrl 显式覆盖（仅在明确配置时生效）。
  try {
    const cfg = useRuntimeConfig()
    const runtimeWsBase = String((cfg.public as any)?.wsBaseUrl || '').trim()
    if (runtimeWsBase) {
      const derived = wsOriginFromApiBase(runtimeWsBase) || runtimeWsBase
      wsURL = derived
      mode = 'runtime-ws-base'
    }
  } catch {
    // ignore
  }

  const resolved = String(wsURL || '').trim()
  console.info('[ws-bus] resolved ws url', { resolved: redactWsURL(resolved), mode, hostMode: isHostMode() })
  return resolved
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

const sendSubscribe = (topics: string[]) => {
  if (!socket || socket.readyState !== WebSocket.OPEN || topics.length === 0) {
    return
  }
  const normalized = Array.from(new Set(topics.map((topic) => normalizeTopic(topic)).filter(Boolean)))
  if (normalized.length === 0) return
  socket.send(
    JSON.stringify({
      type: 'subscribe',
      topics: normalized,
    }),
  )
  normalized.forEach((topic) => subscribedTopics.add(topic))
  console.info('[ws-bus] subscribe sent', { topics: normalized })
}

const ensureSubscriptions = () => {
  const wanted = desiredTopics()
  const missing = wanted.filter((topic) => !subscribedTopics.has(topic))
  if (missing.length > 0) {
    sendSubscribe(missing)
  }
}

const dispatchEvent = (message: WsBusEvent) => {
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

const bindSocket = () => {
  const wsURL = resolveWsURL()
  if (!wsURL) return
  const bearerProtocol = buildWsBearerProtocol(getAuthToken())
  const wsProtocols = bearerProtocol ? [bearerProtocol] : undefined

  console.info('[ws-bus] connecting', {
    wsURL: redactWsURL(wsURL),
    topics: desiredTopics(),
    authMode: bearerProtocol ? 'sec-websocket-protocol' : 'none',
  })
  socket = wsProtocols ? new WebSocket(wsURL, wsProtocols) : new WebSocket(wsURL)

  socket.onopen = () => {
    reconnectAttempts = 0
    subscribedTopics.clear()
    console.info('[ws-bus] connected', { wsURL: redactWsURL(wsURL), topics: desiredTopics() })
    ensureSubscriptions()
  }

  socket.onmessage = (event) => {
    let data: WsBusEvent | null = null
    try {
      data = JSON.parse(event.data)
    } catch {
      return
    }

    if (!data) return
    if (data.type === 'event' || data.topic) {
      dispatchEvent(data)
    }
  }

  socket.onclose = (event) => {
    socket = null
    subscribedTopics.clear()
    reconnectAttempts += 1
    const delay = Math.min(reconnectDelayMs * Math.max(reconnectAttempts, 1), 15000)
    console.warn('[ws-bus] closed', { code: event.code, reason: event.reason, reconnectAttempts, delay })
    if (reconnectTimer) clearTimeout(reconnectTimer)
    reconnectTimer = setTimeout(() => {
      bindSocket()
    }, delay)
  }

  socket.onerror = (event) => {
    console.warn('[ws-bus] error', event)
    // onclose 负责重连
  }
}

export function useWsBusClient(): WsBusClient {
  const connect = () => {
    if (typeof window === 'undefined') return
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      return
    }
    bindSocket()
  }

  const disconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempts = 0
    subscribedTopics.clear()
    if (socket) {
      socket.close()
      socket = null
    }
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
    }
  }

  const isConnected = () => Boolean(socket && socket.readyState === WebSocket.OPEN)

  return {
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    isConnected,
  }
}
