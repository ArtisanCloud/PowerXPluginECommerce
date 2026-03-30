import { miniAppRequest } from "./miniapp-client";

export type EntitlementItem = {
  id: string;
  serviceCode: string;
  quantity: number;
  validFrom?: string;
  validTo?: string;
  stackPolicy: string;
  sourceType: string;
  sourceId: string;
};

export type TokenBalanceItem = {
  tokenCode: string;
  balance: number;
};

export type MembershipBenefitItem = {
  id: string;
  name: string;
  type?: string;
  items?: any;
  status?: string;
};

export type MembershipProfile = {
  customerId?: string;
  customerName?: string;
  membershipTier?: string;
  tierName?: string;
  avatarUrl?: string;
};

export async function miniAppListMembershipBenefits() {
  return await miniAppRequest<{ items: MembershipBenefitItem[] }>({
    method: "GET",
    path: "/membership/benefits",
  });
}

export async function miniAppGetMembershipProfile() {
  return await miniAppRequest<MembershipProfile>({
    method: "GET",
    path: "/membership/profile",
  });
}

export async function miniAppGetEntitlements() {
  return await miniAppRequest<{ items: EntitlementItem[] }>({
    method: "GET",
    path: "/membership/entitlements",
  });
}

export async function miniAppGetTokenBalances() {
  return await miniAppRequest<{ items: TokenBalanceItem[] }>({
    method: "GET",
    path: "/membership/tokens",
  });
}
