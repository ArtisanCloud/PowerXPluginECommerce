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

type AuthResponse = {
  token?: string;
  expiresAt?: string;
  customerId?: string;
  tenantUuid?: string;
  customerName?: string;
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
