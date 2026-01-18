// 统一导出所有 API 相关功能

export * from "./useTemplate";
export * from "./useStream";
export * from "./useCategory";
export * from "./useCategoryTemplate";
export * from "./useCategoryMapping";
export * from "./useCustomer";
export * from "./useOrder";
export * from "./usePricebook";
export * from "./useProductSpec";
export * from "./usePayments";
export {
  useApiClient,
  apiGet,
  apiPost,
  apiPut,
  apiPatch,
  apiDel,
} from "./_client";
export { resolveApiBase, getAuthToken, getTenantUuid } from "./_base";
export type { Page, ApiResponse, ListQuery } from "./_base";
