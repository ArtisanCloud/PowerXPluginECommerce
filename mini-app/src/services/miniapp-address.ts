import { miniAppRequest } from "./miniapp-client";

export type ShippingAddress = {
  label?: string;
  recipientName: string;
  recipientPhone: string;
  countryCode?: string;
  province?: string;
  city?: string;
  district?: string;
  address1: string;
  address2?: string;
  postalCode?: string;
  metadata?: Record<string, any>;
};

export type CustomerAddressDTO = {
  id: string;
  customerId: string;
  isDefault: boolean;
  shippingAddress: ShippingAddress;
  createdAt: string;
  updatedAt: string;
};

export type AddressUpsertRequest = {
  isDefault?: boolean;
  shippingAddress: ShippingAddress;
};

export function formatShippingAddress(a: ShippingAddress) {
  const parts = [a.province, a.city, a.district, a.address1, a.address2].filter(Boolean);
  return parts.join("");
}

export async function listMyAddresses() {
  return await miniAppRequest<CustomerAddressDTO[]>({ method: "GET", path: "/me/addresses" });
}

export async function createMyAddress(req: AddressUpsertRequest) {
  return await miniAppRequest<CustomerAddressDTO>({ method: "POST", path: "/me/addresses", data: req });
}

export async function updateMyAddress(addressId: string, req: AddressUpsertRequest) {
  const id = String(addressId || "").trim();
  return await miniAppRequest<CustomerAddressDTO>({ method: "PATCH", path: `/me/addresses/${encodeURIComponent(id)}`, data: req });
}

export async function deleteMyAddress(addressId: string) {
  const id = String(addressId || "").trim();
  return await miniAppRequest<{ success: boolean }>({ method: "DELETE", path: `/me/addresses/${encodeURIComponent(id)}` });
}

export async function setDefaultMyAddress(addressId: string) {
  const id = String(addressId || "").trim();
  return await miniAppRequest<CustomerAddressDTO>({ method: "POST", path: `/me/addresses/${encodeURIComponent(id)}/default` });
}

