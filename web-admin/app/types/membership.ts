export type MembershipTier = {
  id: string;
  name: string;
  code: string;
  status: string;
  rules?: Record<string, any> | null;
  createdAt?: string;
  updatedAt?: string;
};

export type MembershipBenefit = {
  id: string;
  name: string;
  type: string;
  items?: Record<string, any> | null;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

export type MembershipTierList = {
  items: MembershipTier[];
};

export type MembershipBenefitList = {
  items: MembershipBenefit[];
};

export type GrantEntitlementRequest = {
  customerId: string;
  serviceCode: string;
  quantity: number;
  validFrom?: string;
  validTo?: string;
  stackPolicy?: string;
  reason?: string;
};

export type RevokeEntitlementRequest = {
  entitlementId: string;
  reason?: string;
};

export type AdjustTokenRequest = {
  customerId: string;
  tokenCode: string;
  delta: number;
  reason?: string;
};
