import { apiDel, apiGet, apiPatch, apiPost } from "./_client";
import type { ApiResponse } from "./_base";
import type {
  BulkActionPayload,
  BulkReminderPayload,
  Customer,
  CustomerCreatePayload,
  CustomerDeletePayload,
  CustomerEntitlementList,
  CustomerExportPayload,
  CustomerListFilters,
  CustomerListResponse,
  CustomerTokenBalanceList,
  CustomerUpdatePayload,
  JobStatus,
  MembershipFilters,
  MembershipListResponse,
} from "~/types/customer";

type ApiEnvelope<T> = ApiResponse<T> & {
  request_id?: string;
  timestamp?: string;
  [key: string]: any;
};

const unwrap = async <T>(promise: Promise<ApiEnvelope<T> | T>): Promise<T> => {
  const resp = await promise;
  if (resp && typeof resp === "object" && "data" in (resp as Record<string, any>)) {
    return (resp as ApiEnvelope<T>).data;
  }
  return resp as T;
};

export function useCustomerApi() {
  const basePath = "admin/customers";

  return {
    listCustomers: (filters?: CustomerListFilters, init?: any) =>
      unwrap(apiGet<ApiEnvelope<CustomerListResponse>>(basePath, filters, init)),

    getCustomer: (id: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<Customer>>(`${basePath}/${id}`, undefined, init)),

    getCustomerEntitlements: (id: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<CustomerEntitlementList>>(`${basePath}/${id}/entitlements`, undefined, init)),

    getCustomerTokenBalances: (id: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<CustomerTokenBalanceList>>(`${basePath}/${id}/tokens`, undefined, init)),

    createCustomer: (payload: CustomerCreatePayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<Customer>>(basePath, payload, init)),

    updateCustomer: (id: string, payload: CustomerUpdatePayload, init?: any) =>
      unwrap(apiPatch<ApiEnvelope<Customer>>(`${basePath}/${id}`, payload, init)),

    deleteCustomer: (id: string, payload: CustomerDeletePayload, init?: any) =>
      unwrap(
        apiDel<ApiEnvelope<Customer>>(`${basePath}/${id}`, {
          body: JSON.stringify(payload),
          ...(init || {}),
        }),
      ),

    listMembers: (filters?: MembershipFilters, init?: any) =>
      unwrap(apiGet<ApiEnvelope<MembershipListResponse>>(`${basePath}/members`, filters, init)),

    runBulkAction: (payload: BulkActionPayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<{ taskId: string }>>(`${basePath}/bulk-actions`, payload, init)),

    requestReminder: (payload: BulkReminderPayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<{ taskId: string }>>(`${basePath}/bulk-remind`, payload, init)),

    requestExport: (payload: CustomerExportPayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<{ taskId: string }>>(`${basePath}/export`, payload, init)),

    requestImport: (file: File, init?: any) => {
      const form = new FormData();
      form.append("file", file);
      return unwrap(apiPost<ApiEnvelope<{ taskId: string }>>(`${basePath}/import`, form, init));
    },

    fetchJobStatus: (taskId: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<JobStatus>>(`jobs/${taskId}`, undefined, init)),
  };
}
