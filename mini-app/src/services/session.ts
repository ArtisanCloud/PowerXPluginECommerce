export type CustomerSession = {
  token?: string;
  expiresAt?: string;
  customerId?: string;
  tenantUuid?: string;
  customerName?: string;
};

const TOKEN_KEY = "miniapp.customer.token";
const TENANT_KEY = "miniapp.tenant.uuid";
const CUSTOMER_ID_KEY = "miniapp.customer.id";
const CUSTOMER_NAME_KEY = "miniapp.customer.name";
const EXPIRES_AT_KEY = "miniapp.customer.expiresAt";
const IDENTIFIER_KEY = "miniapp.customer.identifier";

export function getCustomerToken() {
  return String(uni.getStorageSync(TOKEN_KEY) || "").trim();
}

export function getTenantUuid() {
  return String(uni.getStorageSync(TENANT_KEY) || "").trim();
}

export function getCustomerName() {
  return String(uni.getStorageSync(CUSTOMER_NAME_KEY) || "").trim();
}

export function getCustomerIdentifier() {
  return String(uni.getStorageSync(IDENTIFIER_KEY) || "").trim();
}

export function getSession(): CustomerSession {
  return {
    token: getCustomerToken() || undefined,
    tenantUuid: getTenantUuid() || undefined,
    customerId: String(uni.getStorageSync(CUSTOMER_ID_KEY) || "").trim() || undefined,
    customerName: getCustomerName() || undefined,
    expiresAt: String(uni.getStorageSync(EXPIRES_AT_KEY) || "").trim() || undefined,
  };
}

export function setSession(session: CustomerSession) {
  if (session.token) uni.setStorageSync(TOKEN_KEY, session.token);
  if (session.tenantUuid) uni.setStorageSync(TENANT_KEY, session.tenantUuid);
  if (session.customerId) uni.setStorageSync(CUSTOMER_ID_KEY, session.customerId);
  if (session.customerName) uni.setStorageSync(CUSTOMER_NAME_KEY, session.customerName);
  if (session.expiresAt) uni.setStorageSync(EXPIRES_AT_KEY, session.expiresAt);
}

export function setCustomerIdentifier(identifier: string) {
  const v = String(identifier || "").trim();
  if (!v) return;
  uni.setStorageSync(IDENTIFIER_KEY, v);
}

export function clearSession() {
  try {
    uni.removeStorageSync(TOKEN_KEY);
    uni.removeStorageSync(TENANT_KEY);
    uni.removeStorageSync(CUSTOMER_ID_KEY);
    uni.removeStorageSync(CUSTOMER_NAME_KEY);
    uni.removeStorageSync(EXPIRES_AT_KEY);
    uni.removeStorageSync(IDENTIFIER_KEY);
  } catch {}
}
