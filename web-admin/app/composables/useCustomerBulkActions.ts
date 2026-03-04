import { useCustomerApi } from "./api";
import type {
  BulkActionPayload,
  BulkReminderPayload,
  CustomerExportPayload,
} from "~/types/customer";
import { useCustomerMetrics } from "./useCustomerMetrics";
import { useToastAlert } from "./useToastAlert";

export const useCustomerBulkActions = () => {
  const api = useCustomerApi();
  const metrics = useCustomerMetrics();
  const toast = useToastAlert();

  const submitBulkAction = async (payload: BulkActionPayload) => {
    if (!payload.ids?.length) {
      throw new Error("请至少选择一名客户");
    }
    const recorder = metrics.trackJob("customer_bulk_action", {
      action: payload.action,
      count: payload.ids.length,
    });
    try {
      const response = await api.runBulkAction(
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
      const response = await api.requestReminder(
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
      const response = await api.requestExport(
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

  const submitImport = async (params: { file: File; conflictStrategy?: "fail" | "skip" }) => {
    if (!params.file) {
      throw new Error("请上传导入文件");
    }
    const tracker = metrics.trackJob("customer_import_task");
    try {
      const response = await api.requestImport(params.file, params.conflictStrategy || "fail");
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

  return {
    submitBulkAction,
    submitReminder,
    submitExport,
    submitImport,
  };
};
