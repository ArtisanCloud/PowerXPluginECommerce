import { ref } from "vue";
import { useCustomerService } from "./api/services/customerService";
import type {
  BulkActionPayload,
  BulkReminderPayload,
  CustomerExportPayload,
  JobStatus,
} from "~/types/customer";
import { useCustomerMetrics } from "./useCustomerMetrics";
import { useToastAlert } from "./useToastAlert";

const sleep = (ms: number) =>
  new Promise((resolve) => {
    setTimeout(resolve, ms);
  });

export const useCustomerBulkActions = () => {
  const service = useCustomerService();
  const metrics = useCustomerMetrics();
  const toast = useToastAlert();
  const pollingTasks = ref<Record<string, boolean>>({});

  const submitBulkAction = async (payload: BulkActionPayload) => {
    if (!payload.ids?.length) {
      throw new Error("请至少选择一名客户");
    }
    const recorder = metrics.trackJob("customer_bulk_action", {
      action: payload.action,
      count: payload.ids.length,
    });
    try {
      const response = await service.runBulkAction(
        {
          action: payload.action,
          ids: payload.ids,
          payload: payload.payload,
        }
      );
      recorder("success", { taskId: response.taskId });
      toast.add({
        title: "批量任务已提交",
        description: "可在任务中心查看进度",
        color: "primary",
      });
      return response;
    } catch (error) {
      recorder("failed");
      throw error;
    }
  };

  const submitReminder = async (payload: BulkReminderPayload) => {
    if (!payload.ids?.length) {
      throw new Error("请至少选择一个目标客户");
    }
    const tracker = metrics.trackJob("customer_reminder_task", {
      channel: payload.channel,
      count: payload.ids.length,
    });
    try {
      const response = await service.requestReminder(
        {
          ids: payload.ids,
          channel: payload.channel,
          templateId: payload.templateId,
          metadata: payload.metadata,
        }
      );
      tracker("success", { taskId: response.taskId });
      toast.add({
        title: "保级提醒任务已提交",
        description: "任务中心将跟踪发送结果",
        color: "primary",
      });
      return response;
    } catch (error) {
      tracker("failed");
      throw error;
    }
  };

  const submitExport = async (payload: CustomerExportPayload) => {
    const tracker = metrics.trackJob("customer_export_task");
    try {
      const response = await service.requestExport(
        {
          filters: payload.filters,
          fields: payload.fields,
        }
      );
      tracker("success", { taskId: response.taskId });
      toast.add({
        title: "导出任务已创建",
        description: "完成后可在通知中心下载",
        color: "primary",
      });
      return response;
    } catch (error) {
      tracker("failed");
      throw error;
    }
  };

  const submitImport = async (params: { file: File }) => {
    if (!params.file) {
      throw new Error("请上传导入文件");
    }
    const tracker = metrics.trackJob("customer_import_task");
    try {
      const response = await service.requestImport(params.file);
      tracker("success", { taskId: response.taskId });
      toast.add({
        title: "导入任务已提交",
        description: "可在任务中心获取进度",
        color: "primary",
      });
      return response;
    } catch (error) {
      tracker("failed");
      throw error;
    }
  };

  const pollJobOnce = async (taskId: string) => {
    return service.fetchJobStatus(taskId);
  };

  const pollJobUntilFinished = async (
    taskId: string,
    options?: { intervalMs?: number; timeoutMs?: number }
  ): Promise<JobStatus> => {
    const intervalMs = options?.intervalMs ?? 2_000;
    const timeoutMs = options?.timeoutMs ?? 120_000;
    const start = Date.now();
    pollingTasks.value[taskId] = true;

    try {
      while (true) {
        const status = await pollJobOnce(taskId);
        const normalized = (status.status || "").toLowerCase();
        if (["success", "failed", "error"].includes(normalized)) {
          return status;
        }
        if (Date.now() - start > timeoutMs) {
          throw new Error("任务执行超时");
        }
        await sleep(intervalMs);
      }
    } finally {
      delete pollingTasks.value[taskId];
    }
  };

  return {
    submitBulkAction,
    submitReminder,
    submitExport,
    submitImport,
    pollJobOnce,
    pollJobUntilFinished,
    pollingTasks,
  };
};
