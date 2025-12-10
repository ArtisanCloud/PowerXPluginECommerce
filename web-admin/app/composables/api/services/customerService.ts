import { useApiClient } from "../index";
import type {
  BulkActionPayload,
  BulkReminderPayload,
  Customer,
  CustomerCreatePayload,
  CustomerDeletePayload,
  CustomerExportPayload,
  CustomerListFilters,
  CustomerListResponse,
  CustomerUpdatePayload,
  JobStatus,
  MembershipFilters,
  MembershipListResponse,
} from "~/types/customer";

export const useCustomerService = () => {
  const { client } = useApiClient();
  const basePath = "/admin/customers";

  type ApiEnvelope<T> = {
    success?: boolean;
    data: T;
    message?: string;
    code?: number | string;
    request_id?: string;
    timestamp?: string;
    [key: string]: any;
  };

  const unwrap = <T>(result: ApiEnvelope<T> | T): T => {
    if (
      result &&
      typeof result === "object" &&
      "data" in result &&
      ("success" in (result as Record<string, any>) ||
        "code" in (result as Record<string, any>) ||
        "request_id" in (result as Record<string, any>))
    ) {
      return (result as ApiEnvelope<T>).data;
    }
    return result as T;
  };

  const listCustomers = async (filters?: CustomerListFilters) => {
    const response = await client<ApiEnvelope<CustomerListResponse>>(`${basePath}`, {
      method: "GET",
      query: filters,
    });
    return unwrap(response);
  };

  const createCustomer = async (payload: CustomerCreatePayload) => {
    const response = await client<ApiEnvelope<Customer>>(`${basePath}`, {
      method: "POST",
      body: payload,
    });
    return unwrap(response);
  };

  const updateCustomer = async (id: string, payload: CustomerUpdatePayload) => {
    if (!id) {
      throw new Error("customerId is required");
    }
    const response = await client<ApiEnvelope<Customer>>(`${basePath}/${id}`, {
      method: "PATCH",
      body: payload,
    });
    return unwrap(response);
  };

  const deleteCustomer = async (id: string, payload: CustomerDeletePayload) => {
    if (!id) {
      throw new Error("customerId is required");
    }
    const response = await client<ApiEnvelope<Customer>>(`${basePath}/${id}`, {
      method: "DELETE",
      body: payload,
    });
    return unwrap(response);
  };

  const getCustomer = async (id: string) => {
    if (!id) {
      throw new Error("customerId is required");
    }
    const response = await client<ApiEnvelope<Customer>>(`${basePath}/${id}`, { method: "GET" });
    return unwrap(response);
  };

  const listMembers = async (filters?: MembershipFilters) => {
    const response = await client<ApiEnvelope<MembershipListResponse>>(`${basePath}/members`, {
      method: "GET",
      query: filters,
    });
    return unwrap(response);
  };

  const runBulkAction = async (payload: BulkActionPayload) => {
    const response = await client<ApiEnvelope<{ taskId: string }>>(`${basePath}/bulk-actions`, {
      method: "POST",
      body: payload,
    });
    return unwrap(response);
  };

  const requestReminder = async (payload: BulkReminderPayload) => {
    const response = await client<ApiEnvelope<{ taskId: string }>>(`${basePath}/bulk-remind`, {
      method: "POST",
      body: payload,
    });
    return unwrap(response);
  };

  const requestExport = async (payload: CustomerExportPayload) => {
    const response = await client<ApiEnvelope<{ taskId: string }>>(`${basePath}/export`, {
      method: "POST",
      body: payload,
    });
    return unwrap(response);
  };

  const requestImport = async (file: File) => {
    const form = new FormData();
    form.append("file", file);
    const response = await client<ApiEnvelope<{ taskId: string }>>(`${basePath}/import`, {
      method: "POST",
      body: form,
    });
    return unwrap(response);
  };

  const fetchJobStatus = async (taskId: string) => {
    if (!taskId) {
      throw new Error("taskId is required");
    }
    const response = await client<ApiEnvelope<JobStatus>>(`/jobs/${taskId}`, { method: "GET" });
    return unwrap(response);
  };

  return {
    listCustomers,
    createCustomer,
    updateCustomer,
    deleteCustomer,
    getCustomer,
    listMembers,
    runBulkAction,
    requestReminder,
    requestExport,
    requestImport,
    fetchJobStatus,
  };
};
