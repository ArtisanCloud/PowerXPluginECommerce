import type { RouteLocationNormalizedLoaded, Router } from 'vue-router'

type QueryRecord = Record<string, string>
type ManagedQuerySyncEvent = {
  skipped: boolean;
  currentQuery: QueryRecord;
  nextQuery: QueryRecord;
}
type ManagedQuerySyncDebugStats = {
  page: string;
  logEvery?: number;
}

const normalizeQuery = (query: RouteLocationNormalizedLoaded['query']): QueryRecord => {
  const normalized: QueryRecord = {}
  for (const [key, value] of Object.entries(query)) {
    if (value != null) {
      normalized[key] = Array.isArray(value) ? String(value[0]) : String(value)
    }
  }
  return normalized
}

const isSameQuery = (left: QueryRecord, right: QueryRecord) => {
  const leftKeys = Object.keys(left)
  const rightKeys = Object.keys(right)
  if (leftKeys.length !== rightKeys.length) return false
  return rightKeys.every((key) => left[key] === right[key])
}

export const useManagedQuerySync = (options: {
  route: RouteLocationNormalizedLoaded;
  router: Router;
  buildNextQuery: () => QueryRecord;
  isReady?: () => boolean;
  onSynced?: (event: ManagedQuerySyncEvent) => void | Promise<void>;
  debugStats?: ManagedQuerySyncDebugStats;
}) => {
  let pending = false
  let runningTask: Promise<void> | null = null
  let syncedCount = 0
  let skippedCount = 0

  const trackDebugStats = (skipped: boolean) => {
    if (!import.meta.dev || !options.debugStats) return
    if (skipped) skippedCount += 1
    else syncedCount += 1
    const logEvery = Math.max(1, Number(options.debugStats.logEvery || 10))
    if ((syncedCount + skippedCount) % logEvery !== 0) return
    console.debug('[query-sync][stats]', {
      page: options.debugStats.page,
      synced: syncedCount,
      skipped: skippedCount,
    })
  }

  const syncManagedQuery = async () => {
    if (options.isReady && !options.isReady()) return
    const nextQuery = options.buildNextQuery()
    const currentQuery = normalizeQuery(options.route.query)
    if (isSameQuery(currentQuery, nextQuery)) {
      trackDebugStats(true)
      if (options.onSynced) {
        await options.onSynced({
          skipped: true,
          currentQuery,
          nextQuery,
        })
      }
      return
    }
    await options.router.replace({ query: nextQuery })
    trackDebugStats(false)
    if (options.onSynced) {
      await options.onSynced({
        skipped: false,
        currentQuery,
        nextQuery,
      })
    }
  }

  const queueManagedQuerySync = async () => {
    pending = true
    if (runningTask) {
      await runningTask
      return
    }
    runningTask = (async () => {
      try {
        while (pending) {
          pending = false
          await syncManagedQuery()
        }
      } finally {
        runningTask = null
      }
    })()
    await runningTask
  }

  return {
    syncManagedQuery,
    queueManagedQuerySync,
  }
}
