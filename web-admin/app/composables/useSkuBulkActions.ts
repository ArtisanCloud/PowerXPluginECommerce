import { useToast } from '#imports'
import { useProductSkuStore } from '~/stores/productSku'
import { useSkuApi } from '~/composables/api/useSku'
import type {
  BulkOperationType,
  SkuBulkTask,
  SkuBulkTaskRequest,
  SkuBulkTaskStatus,
  SkuExportPayload,
  SkuImportMode,
} from '~/types/product/sku'

const normalizeTask = (
  task: any,
  operationType?: BulkOperationType,
  explicitTaskType?: SkuBulkTask["taskType"],
): SkuBulkTask => {
  const taskId = task?.taskId ?? task?.task_id ?? ''
  const status = (task?.status ?? 'pending') as SkuBulkTaskStatus
  const approvalRequired = Boolean(task?.approvalRequired ?? task?.approval_required)
  const approvalState = task?.approvalState ?? task?.approval_state
  const stats = task?.stats
  const errorReportUrl = task?.errorReportUrl ?? task?.error_report
  const derivedTaskType =
    explicitTaskType ??
    task?.taskType ??
    task?.task_type ??
    (operationType
      ? operationType.startsWith('inventory')
        ? 'inventory_adjustment'
        : 'price_adjustment'
      : 'import')
  return {
    taskId,
    taskType: derivedTaskType,
    scope: task?.scope,
    operation: task?.operation,
    status,
    approvalRequired,
    approvalState,
    approvalReason: task?.approvalReason ?? task?.approval_reason,
    approvalThreshold: task?.approvalThreshold ?? task?.approval_threshold,
    affectedCount: task?.affectedCount ?? task?.affected_count,
    submittedBy: task?.submittedBy ?? task?.submitted_by,
    approvedBy: task?.approvedBy ?? task?.approved_by,
    approvedAt: task?.approvedAt ?? task?.approved_at,
    stats,
    errorReportUrl,
    result: task?.result,
    createdAt: task?.createdAt ?? task?.created_at,
    updatedAt: task?.updatedAt ?? task?.updated_at,
  }
}

export const useSkuBulkActions = () => {
  const api = useSkuApi()
  const store = useProductSkuStore()
  const toast = useToast()

  const submitAdjustment = async (payload: SkuBulkTaskRequest) => {
    const response = await api.submitBulkTask(payload).catch((error) => {
      throw error
    })
    let normalized: SkuBulkTask | null = null
    if (response) {
      normalized = normalizeTask(response, payload.operation.type)
      if (normalized.taskId) {
        store.trackBulkTask(normalized)
      }
    }
    toast.add({
      title: '批量任务已创建',
      description: '可前往任务中心查看执行情况',
      color: 'primary',
    })
    return normalized
  }

  const submitImport = async (input: { file: File; mode?: SkuImportMode }) => {
    if (!input?.file) {
      throw new Error('Missing import file')
    }
    const response = await api.importByFile(input.file, input.mode ?? 'upsert').catch((error) => {
      throw error
    })
    let normalized: SkuBulkTask | null = null
    if (response) {
      normalized = normalizeTask(response, undefined, 'import')
      if (normalized.taskId) {
        store.trackBulkTask(normalized)
      }
    }
    toast.add({
      title: '导入任务已创建',
      description: '文件校验完成后会自动执行，可在任务中心跟进',
      color: 'primary',
    })
    return normalized
  }

  const submitExport = async (payload: SkuExportPayload) => {
    const response = await api.exportSkus(payload).catch((error) => {
      throw error
    })
    let normalized: SkuBulkTask | null = null
    if (response) {
      normalized = normalizeTask(response, undefined, 'export')
      if (normalized.taskId) {
        store.trackBulkTask(normalized)
      }
    }
    toast.add({
      title: '导出任务已创建',
      description: '系统会生成下载链接，可在任务中心或本页获取。',
      color: 'primary',
    })
    return normalized
  }

  const decideApproval = async (input: { taskId: string; decision: 'approve' | 'reject'; note?: string }) => {
    if (!input.taskId) {
      throw new Error('taskId is required')
    }
    const response = await api.decideBulkTask(input.taskId, {
      decision: input.decision,
      note: input.note,
    })
    const normalized = normalizeTask(response)
    if (normalized.taskId) {
      store.trackBulkTask(normalized)
    }
    toast.add({
      title: input.decision === 'approve' ? '任务审批通过' : '任务已驳回',
      description: input.note,
      color: input.decision === 'approve' ? 'success' : 'warning',
    })
  }

  const retryTask = async (taskId: string) => {
    if (!taskId) return
    const response = await api.retryBulkTask(taskId)
    const normalized = normalizeTask(response)
    if (normalized.taskId) {
      store.trackBulkTask(normalized)
    }
    toast.add({
      title: '已触发重试',
      description: `任务 ${taskId} 将重新执行`,
      color: 'primary',
    })
  }

  return {
    submitAdjustment,
    submitImport,
    submitExport,
    decideApproval,
    retryTask,
  }
}
