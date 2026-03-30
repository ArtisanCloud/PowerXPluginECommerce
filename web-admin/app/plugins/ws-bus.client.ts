import { defineNuxtPlugin } from '#imports'
import { useWsBusClient } from '~/composables/useWsBusClient'

declare global {
  interface Window {
    __PX_WS_BUS_BOOTSTRAPPED__?: boolean
  }
}

export default defineNuxtPlugin(() => {
  if (!process.client) return

  if (window.__PX_WS_BUS_BOOTSTRAPPED__) {
    return
  }
  window.__PX_WS_BUS_BOOTSTRAPPED__ = true

  const wsBus = useWsBusClient()
  wsBus.connect()
})
