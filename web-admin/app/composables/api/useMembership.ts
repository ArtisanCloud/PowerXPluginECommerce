import { apiGet, apiPost } from "./_client";
import type { ApiResponse } from "./_base";
import type {
  AdjustTokenRequest,
  GrantEntitlementRequest,
  MembershipBenefitList,
  MembershipTierList,
  RevokeEntitlementRequest,
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

    listBenefits: (init?: any) =>
      unwrap(apiGet<ApiEnvelope<MembershipBenefitList>>(`${basePath}/benefits`, undefined, init)),

    grantEntitlement: (payload: GrantEntitlementRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<any>>(`${basePath}/entitlements/grant`, payload, init)),

    revokeEntitlement: (payload: RevokeEntitlementRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<any>>(`${basePath}/entitlements/revoke`, payload, init)),

    adjustToken: (payload: AdjustTokenRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<any>>(`${basePath}/tokens/adjust`, payload, init)),
  };
}
