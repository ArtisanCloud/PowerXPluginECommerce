import { useApiClient } from "../index";
import type {
  AuditContext,
  BulkActionPayload,
  BulkReminderPayload,
  Customer,
  CustomerExportPayload,
  CustomerListFilters,
  CustomerListResponse,
  JobStatus,
  MembershipFilters,
  MembershipListResponse,
} from "~/types/customer";

type RequestOptions = {
  headers?: HeadersInit;
  audit?: AuditContext;
};

const withOptions = (options?: RequestOptions) => (options ? { ...options } : {});

export const useCustomerService = () => {
  const { client } = useApiClient();
  const basePath = "/customers";

  const listCustomers = (filters?: CustomerListFilters) => {
    return client<CustomerListResponse>(`${basePath}`, {
      method: "GET",
      query: filters,
    });
  };

  const getCustomer = (id: string) => {
    if (!id) {
      throw new Error("customerId is required");
    }
    return client<Customer>(`${basePath}/${id}`, { method: "GET" });
  };

  const listMembers = (filters?: MembershipFilters) => {
    return client<MembershipListResponse>(`${basePath}/members`, {
      method: "GET",
      query: filters,
    });
  };

  const runBulkAction = (payload: BulkActionPayload, options?: RequestOptions) => {
    return client<{ taskId: string }>(`${basePath}/bulk-actions`, {
      method: "POST",
      body: payload,
      ...withOptions(options),
    });
  };

  const requestReminder = (payload: BulkReminderPayload, options?: RequestOptions) => {
    return client<{ taskId: string }>(`${basePath}/bulk-remind`, {
      method: "POST",
      body: payload,
      ...withOptions(options),
    });
  };

  const requestExport = (payload: CustomerExportPayload, options?: RequestOptions) => {
    return client<{ taskId: string }>(`${basePath}/export`, {
      method: "POST",
      body: payload,
      ...withOptions(options),
    });
  };

  const requestImport = (file: File, options?: RequestOptions) => {
    const form = new FormData();
    form.append("file", file);
    return client<{ taskId: string }>(`${basePath}/import`, {
      method: "POST",
      body: form,
      ...withOptions(options),
    });
  };

  const fetchJobStatus = (taskId: string) => {
    if (!taskId) {
      throw new Error("taskId is required");
    }
    return client<JobStatus>(`/jobs/${taskId}`, { method: "GET" });
  };

  return {
    listCustomers,
    getCustomer,
    listMembers,
    runBulkAction,
    requestReminder,
    requestExport,
    requestImport,
    fetchJobStatus,
  };
};
