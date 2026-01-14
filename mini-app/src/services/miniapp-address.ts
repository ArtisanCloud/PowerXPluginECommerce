import { miniAppRequest } from "./miniapp-client";

export type MiniAppShippingAddress = {
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

export type MiniAppCustomerAddress = {
  id: string;
  customerId: string;
  isDefault: boolean;
  shippingAddress: MiniAppShippingAddress;
  createdAt: string;
  updatedAt: string;
};

export type MiniAppAddressUpsertRequest = {
  isDefault?: boolean;
  shippingAddress: MiniAppShippingAddress;
};

export async function miniAppListMyAddresses() {
  return await miniAppRequest<MiniAppCustomerAddress[]>({
    method: "GET",
    path: "/me/addresses",
  });
}

export async function miniAppCreateMyAddress(req: MiniAppAddressUpsertRequest) {
  return await miniAppRequest<MiniAppCustomerAddress>({
    method: "POST",
    path: "/me/addresses",
    data: req,
  });
}

export async function miniAppUpdateMyAddress(id: string, req: MiniAppAddressUpsertRequest) {
  return await miniAppRequest<MiniAppCustomerAddress>({
    method: "PATCH",
    path: `/me/addresses/${encodeURIComponent(String(id))}`,
    data: req,
  });
}

export async function miniAppDeleteMyAddress(id: string) {
  return await miniAppRequest<{ success: boolean }>({
    method: "DELETE",
    path: `/me/addresses/${encodeURIComponent(String(id))}`,
  });
}

export async function miniAppSetMyAddressDefault(id: string) {
  return await miniAppRequest<MiniAppCustomerAddress>({
    method: "POST",
    path: `/me/addresses/${encodeURIComponent(String(id))}/default`,
  });
}

const KEY_SELECTED = "miniapp.address.selectedId";

export function getSelectedAddressId() {
  return String(uni.getStorageSync(KEY_SELECTED) || "").trim();
}

export function setSelectedAddressId(id: string) {
  uni.setStorageSync(KEY_SELECTED, String(id || "").trim());
}

export function maskPhone(phone: string) {
  const s = String(phone || "").replace(/\s+/g, "").trim();
  if (s.length < 7) return s || "—";
  return `${s.slice(0, 3)}****${s.slice(-4)}`;
}

export function formatFullAddress(a: MiniAppShippingAddress) {
  const parts = [a.province, a.city, a.district, a.address1, a.address2].filter(Boolean);
  return parts.join("");
}

