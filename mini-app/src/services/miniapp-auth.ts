import { miniAppRequest } from "./miniapp-client";
import { setCustomerIdentifier, setSession } from "./session";

export type MiniAppAuthLoginInput = {
  identifier: string;
  password: string;
};

export type MiniAppAuthRegisterInput = {
  name: string;
  identifier: string;
  password: string;
  email?: string;
  phone?: string;
};

export type MiniAppAuthWechatLoginInput = {
  providerId: string | number;
  code: string;
  nickname?: string;
  avatarUrl?: string;
};

type AuthResponse = {
  token?: string;
  expiresAt?: string;
  customerId?: string;
  tenantUuid?: string;
  customerName?: string;
  openid?: string;
  unionid?: string;
};

export async function miniAppAuthLogin(input: MiniAppAuthLoginInput) {
  const data = await miniAppRequest<AuthResponse>({
    method: "POST",
    path: "/auth/login",
    data: input,
  });
  setCustomerIdentifier(input.identifier);
  setSession({
    token: data?.token,
    tenantUuid: data?.tenantUuid,
    customerId: data?.customerId,
    customerName: data?.customerName,
    expiresAt: data?.expiresAt,
  });
  return data;
}

export async function miniAppAuthRegister(input: MiniAppAuthRegisterInput) {
  const data = await miniAppRequest<AuthResponse>({
    method: "POST",
    path: "/auth/register",
    data: input,
  });
  setCustomerIdentifier(input.identifier);
  setSession({
    token: data?.token,
    tenantUuid: data?.tenantUuid,
    customerId: data?.customerId,
    customerName: data?.customerName,
    expiresAt: data?.expiresAt,
  });
  return data;
}

export async function miniAppAuthWechatLogin(input: MiniAppAuthWechatLoginInput) {
  const data = await miniAppRequest<AuthResponse>({
    method: "POST",
    path: "/auth/wechat/login",
    data: input,
  });
  if (data?.openid) setCustomerIdentifier(`wechat:${data.openid}`);
  if (data?.openid) uni.setStorageSync("miniapp.customer.openid", String(data.openid));
  setSession({
    token: data?.token,
    tenantUuid: data?.tenantUuid,
    customerId: data?.customerId,
    customerName: data?.customerName,
    expiresAt: data?.expiresAt,
  });
  return data;
}
