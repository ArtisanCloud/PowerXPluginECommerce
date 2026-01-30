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
