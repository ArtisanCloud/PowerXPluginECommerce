import { useApiClient } from "../index";

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

export const useCustomerAddressService = () => {
  const { client } = useApiClient();

  type ApiEnvelope<T> = {
    success?: boolean;
    data: T;
    message?: string;
    code?: number | string;
    request_id?: string;
    timestamp?: string;
    [key: string]: any;
  };

  const unwrap = <T>(result: ApiEnvelope<T> | T): T => {
    if (
      result &&
      typeof result === "object" &&
      "data" in result &&
      ("success" in (result as Record<string, any>) ||
        "code" in (result as Record<string, any>) ||
        "request_id" in (result as Record<string, any>))
    ) {
      return (result as ApiEnvelope<T>).data;
    }
    return result as T;
  };

  const normalizeShippingAddress = (address: Record<string, any>): ShippingAddress => ({
    label: address?.label,
    recipientName: String(address?.recipientName ?? address?.recipient_name ?? "").trim(),
    recipientPhone: String(address?.recipientPhone ?? address?.recipient_phone ?? "").trim(),
    countryCode: address?.countryCode ?? address?.country_code,
    province: address?.province,
    city: address?.city,
    district: address?.district,
    address1: String(address?.address1 ?? address?.address_1 ?? address?.address_line1 ?? "").trim(),
    address2: address?.address2 ?? address?.address_2 ?? address?.address_line2,
    postalCode: address?.postalCode ?? address?.postal_code,
    metadata: address?.metadata ?? address?.meta,
  });

  const normalizeCustomerAddress = (item: Record<string, any>): CustomerAddressDTO => ({
    id: String(item?.id ?? "").trim(),
    customerId: String(item?.customerId ?? item?.customer_id ?? "").trim(),
    isDefault: Boolean(item?.isDefault ?? item?.is_default),
    shippingAddress: normalizeShippingAddress(item?.shippingAddress ?? item?.shipping_address ?? {}),
    createdAt: item?.createdAt ?? item?.created_at,
    updatedAt: item?.updatedAt ?? item?.updated_at,
  });

  const normalizeAddressList = (items: CustomerAddressDTO[] | any): CustomerAddressDTO[] =>
    Array.isArray(items) ? items.map((item) => normalizeCustomerAddress(item as Record<string, any>)) : [];

  const basePath = "/admin/customers";

  const listCustomerAddresses = async (customerId: string) => {
    const id = String(customerId || "").trim();
    if (!id) throw new Error("customerId is required");
    const response = await client<ApiEnvelope<CustomerAddressDTO[]>>(`${basePath}/${id}/addresses`, {
      method: "GET",
    });
    return normalizeAddressList(unwrap(response) as any);
  };

  const createCustomerAddress = async (customerId: string, payload: AddressUpsertRequest) => {
    const id = String(customerId || "").trim();
    if (!id) throw new Error("customerId is required");
    const response = await client<ApiEnvelope<CustomerAddressDTO>>(`${basePath}/${id}/addresses`, {
      method: "POST",
      body: payload,
    });
    return normalizeCustomerAddress(unwrap(response) as any);
  };

  const updateCustomerAddress = async (customerId: string, addressId: string, payload: AddressUpsertRequest) => {
    const id = String(customerId || "").trim();
    const aid = String(addressId || "").trim();
    if (!id) throw new Error("customerId is required");
    if (!aid) throw new Error("addressId is required");
    const response = await client<ApiEnvelope<CustomerAddressDTO>>(`${basePath}/${id}/addresses/${aid}`, {
      method: "PATCH",
      body: payload,
    });
    return normalizeCustomerAddress(unwrap(response) as any);
  };

  const deleteCustomerAddress = async (customerId: string, addressId: string) => {
    const id = String(customerId || "").trim();
    const aid = String(addressId || "").trim();
    if (!id) throw new Error("customerId is required");
    if (!aid) throw new Error("addressId is required");
    const response = await client<ApiEnvelope<{ success: boolean }>>(`${basePath}/${id}/addresses/${aid}`, {
      method: "DELETE",
    });
    return unwrap(response);
  };

  const setDefaultCustomerAddress = async (customerId: string, addressId: string) => {
    const id = String(customerId || "").trim();
    const aid = String(addressId || "").trim();
    if (!id) throw new Error("customerId is required");
    if (!aid) throw new Error("addressId is required");
    const response = await client<ApiEnvelope<CustomerAddressDTO>>(`${basePath}/${id}/addresses/${aid}/default`, {
      method: "POST",
    });
    return normalizeCustomerAddress(unwrap(response) as any);
  };

  return {
    listCustomerAddresses,
    createCustomerAddress,
    updateCustomerAddress,
    deleteCustomerAddress,
    setDefaultCustomerAddress,
  };
};
