type WarnUnknownManagedQueryKeysOptions = {
  scope: string;
  rawQuery: Record<string, unknown>;
  managedKeys: readonly string[];
  managedKeySet: Set<string>;
  keyPrefixes?: readonly string[];
  emitConsole?: boolean;
  onWarn?: (payload: WarnUnknownManagedQueryPayload) => void;
  debugStats?: {
    page: string;
    logEvery?: number;
  };
}

export type WarnUnknownManagedQueryPayload = {
  scope: string;
  unknownKeys: string[];
  managedKeys: string[];
  rawQueryKeys: string[];
}

export const readStringFromQueryValue = (value: unknown) => {
  if (Array.isArray(value)) return String(value[0] || '')
  return typeof value === 'string' ? value : ''
}

export const collectManagedQueryValues = <T extends string>(options: {
  rawQuery: Record<string, unknown>;
  managedKeys: readonly T[];
}) => {
  const managed: Partial<Record<T, string>> = {}
  options.managedKeys.forEach((key) => {
    managed[key] = readStringFromQueryValue(options.rawQuery[key])
  })
  return managed
}

export const createManagedQueryKeySet = <T extends string>(managedKeys: readonly T[]) =>
  new Set<string>(managedKeys)

export const normalizeQueryToStringRecord = (rawQuery: Record<string, unknown>) => {
  const normalized: Record<string, string> = {}
  Object.entries(rawQuery).forEach(([key, value]) => {
    const text = readStringFromQueryValue(value)
    if (!text) return
    normalized[key] = text
  })
  return normalized
}

export const parseAliasedQueryValue = (
  value: unknown,
  aliases: Record<string, string>,
): string | null => {
  const text = readStringFromQueryValue(value).trim().toLowerCase()
  if (!text) return null
  return aliases[text] || text
}

const lastWarningByScope = new Map<string, string>()
const warnedCountByScope = new Map<string, number>()

export const warnUnknownManagedQueryKeys = ({
  scope,
  rawQuery,
  managedKeys,
  managedKeySet,
  keyPrefixes = ['tx', 'er'],
  emitConsole = true,
  onWarn,
  debugStats,
}: WarnUnknownManagedQueryKeysOptions) => {
  if (!import.meta.dev) return
  const unknownKeys = Object.keys(rawQuery).filter((key) =>
    keyPrefixes.some((prefix) => key.startsWith(prefix)) && !managedKeySet.has(key),
  )
  if (!unknownKeys.length) return

  const warningKey = [...unknownKeys].sort().join('|')
  if (lastWarningByScope.get(scope) === warningKey) return
  lastWarningByScope.set(scope, warningKey)

  const payload = {
    unknownKeys,
    managedKeys: [...managedKeys],
    rawQueryKeys: Object.keys(rawQuery),
  }
  if (emitConsole) {
    console.warn(`[${scope}][query-sync][unknown-managed-keys]`, payload)
  }
  onWarn?.({
    scope,
    ...payload,
  })

  if (debugStats) {
    const warned = (warnedCountByScope.get(scope) || 0) + 1
    warnedCountByScope.set(scope, warned)
    const every = Math.max(1, Number(debugStats.logEvery || 1))
    if (warned % every === 0) {
      console.debug('[query-sync][unknown-managed-keys]', {
        page: debugStats.page,
        warned,
        unknownKeys: payload.unknownKeys,
      })
    }
  }
}
