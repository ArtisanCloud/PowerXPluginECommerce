export const RECONCILIATION_ERROR_MESSAGES: Record<string, string> = {
  RECONCILIATION_BATCH_NOT_FOUND: "未找到对账批次，请刷新后重试",
  RECONCILIATION_DELTA_NOT_FOUND: "未找到差异记录，请确认数据是否已变更",
  RECONCILIATION_TASK_ALREADY_OPEN: "该差异已有未关闭处置任务，请直接跟进现有任务",
  RECONCILIATION_UNSUPPORTED_ACTION: "当前操作不被支持，请检查请求参数或状态",
  RECONCILIATION_INVALID_BILLING_CYCLE: "账期格式错误，请使用 YYYY-MM-DD",
}

export const resolveReconciliationErrorMessage = (code?: string, fallback = "操作失败，请稍后重试"): string => {
  if (!code) return fallback
  return RECONCILIATION_ERROR_MESSAGES[code] || fallback
}

