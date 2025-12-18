import type { NuxtApp } from '#app'

type PerfLogger = {
  logKpiLoad: (durationMs: number) => void
}

declare module '#app' {
  interface NuxtApp {
    $perf: PerfLogger
  }
}

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $perf: PerfLogger
  }
}

export {}
