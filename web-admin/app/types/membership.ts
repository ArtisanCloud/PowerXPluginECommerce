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

export type CreateTierRequest = {
  name: string;
  code: string;
  status?: string;
  rules?: Record<string, any> | null;
};

export type UpdateTierStatusRequest = {
  status: string;
};

export type CreateBenefitRequest = {
  name: string;
  type?: string;
  status?: string;
  items?: any;
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

export type MembershipTokenTransaction = {
  id: string;
  customerId: string;
  tokenCode: string;
  delta: number;
  sourceType: string;
  sourceId: string;
  createdAt: string;
};

export type MembershipTokenTransactionList = {
  items: MembershipTokenTransaction[];
  total?: number;
  page?: number;
  pageSize?: number;
};

export type RedeemPointsBenefitRequest = {
  customerId: string;
  benefitId: string;
  pointsCost: number;
  reason?: string;
  sourceId?: string;
};

export type RedeemPointsBenefitResult = {
  benefitId: string;
  customerId: string;
  transactionId: string;
  balance: number;
  grantedServices: string[];
};
