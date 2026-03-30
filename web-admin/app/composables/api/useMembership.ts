import { apiDel, apiGet, apiPatch, apiPost } from "./_client";
import type { ApiResponse } from "./_base";
import type {
  AdjustTokenRequest,
  CreateBenefitRequest,
  CreateTierRequest,
  GrantEntitlementRequest,
  MembershipBenefitList,
  MembershipTokenTransactionList,
  MembershipTierList,
  RedeemPointsBenefitResult,
  RedeemPointsBenefitRequest,
  RevokeEntitlementRequest,
  UpdateTierStatusRequest,
} from "~/types/membership";

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

export function useMembershipAdminApi() {
  const basePath = "/admin/membership";

  return {
    listTiers: (init?: any) =>
      unwrap(apiGet<ApiEnvelope<MembershipTierList>>(`${basePath}/tiers`, undefined, init)),
    createTier: (payload: CreateTierRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<any>>(`${basePath}/tiers`, payload, init)),
    updateTierStatus: (id: string, payload: UpdateTierStatusRequest, init?: any) =>
      unwrap(apiPatch<ApiEnvelope<any>>(`${basePath}/tiers/${id}/status`, payload, init)),
    deleteTier: (id: string, init?: any) =>
      unwrap(apiDel<ApiEnvelope<any>>(`${basePath}/tiers/${id}`, init)),

    listBenefits: (init?: any) =>
      unwrap(apiGet<ApiEnvelope<MembershipBenefitList>>(`${basePath}/benefits`, undefined, init)),
    createBenefit: (payload: CreateBenefitRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<any>>(`${basePath}/benefits`, payload, init)),

    grantEntitlement: (payload: GrantEntitlementRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<any>>(`${basePath}/entitlements/grant`, payload, init)),

    revokeEntitlement: (payload: RevokeEntitlementRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<any>>(`${basePath}/entitlements/revoke`, payload, init)),

    adjustToken: (payload: AdjustTokenRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<any>>(`${basePath}/tokens/adjust`, payload, init)),
    listTokenTransactions: (
      query: {
        customerId?: string;
        tokenCode?: string;
        sourceType?: string;
        sourceId?: string;
        createdFrom?: string;
        createdTo?: string;
        page?: number;
        pageSize?: number;
      },
      init?: any,
    ) =>
      unwrap(apiGet<ApiEnvelope<MembershipTokenTransactionList>>(`${basePath}/tokens/transactions`, query, init)),
    redeemPointsBenefit: (payload: RedeemPointsBenefitRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<RedeemPointsBenefitResult>>(`${basePath}/points/redeem`, payload, init)),
  };
}
