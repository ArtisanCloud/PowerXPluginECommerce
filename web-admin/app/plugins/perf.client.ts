export default defineNuxtPlugin(() => {
  const logKpiLoad = (durationMs: number) => {
    const payload = {
      durationMs: Math.round(durationMs),
      thresholdMs: 3000,
      passed: durationMs <= 3000,
      timestamp: new Date().toISOString(),
    }
    if (process.dev) {
      console.info('[ChannelMaster][KPI]', payload)
    }
    if (typeof window !== 'undefined') {
      // expose latest payload for quick inspection
      ;(window as any).__channelPerf = {
        ...(window as any).__channelPerf,
        lastKpiLoad: payload,
      }
      window.dispatchEvent(new CustomEvent('channel-master:kpi-load', { detail: payload }))
    }
  }

  return {
    provide: {
      perf: {
        logKpiLoad,
      },
    },
  }
})
