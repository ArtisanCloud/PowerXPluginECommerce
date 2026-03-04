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

  const normalizeCustomer = (item: Record<string, any>): Customer => ({
    id: String(item?.id ?? "").trim(),
    name: String(item?.name ?? "").trim(),
    type: item?.type ?? "individual",
    email: item?.email ?? item?.email_address,
    phone: item?.phone ?? item?.phone_number,
    country: item?.country,
    region: item?.region,
    membershipTier: item?.membershipTier ?? item?.membership_tier,
    membershipTierLabel: item?.membershipTierLabel ?? item?.membership_tier_label,
    growthValue: item?.growthValue ?? item?.growth_value,
    points: item?.points,
    lastOrderAmount: item?.lastOrderAmount ?? item?.last_order_amount,
    lastOrderAt: item?.lastOrderAt ?? item?.last_order_at,
    status: item?.status,
    riskLevel: item?.riskLevel ?? item?.risk_level,
    source: item?.source,
    accountManager: item?.accountManager ?? item?.account_manager,
    tags: item?.tags ?? [],
    createdAt: item?.createdAt ?? item?.created_at,
    updatedAt: item?.updatedAt ?? item?.updated_at,
    notes: item?.notes,
    maskedFields: item?.maskedFields ?? item?.masked_fields,
    membershipSnapshot: item?.membershipSnapshot ?? item?.membership_snapshot,
    metadata: item?.metadata ?? item?.meta,
  });

  const normalizeCustomerList = (payload: CustomerListResponse): CustomerListResponse => ({
    ...payload,
    data: Array.isArray(payload?.data) ? payload.data.map((item) => normalizeCustomer(item as Record<string, any>)) : [],
  });

  const listCustomers = async (filters?: CustomerListFilters) => {
    const response = await client<ApiEnvelope<CustomerListResponse>>(`${basePath}`, {
      method: "GET",
      query: filters,
    });
    return normalizeCustomerList(unwrap(response));
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
    return normalizeCustomer(unwrap(response) as Record<string, any>);
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

  const requestImport = async (file: File, conflictStrategy: "fail" | "skip" = "fail") => {
    const form = new FormData();
    form.append("file", file);
    form.append("conflict_strategy", conflictStrategy);
    const response = await client<ApiEnvelope<{ taskId: string }>>(`${basePath}/import`, {
      method: "POST",
      body: form,
    });
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
  };
};
