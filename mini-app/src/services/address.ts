export type AddressItem = {
  id: string;
  name: string;
  phone: string;
  province?: string;
  city?: string;
  district?: string;
  detail: string;
  isDefault?: boolean;
};

const KEY_LIST = "miniapp.address.list";
const KEY_SELECTED = "miniapp.address.selected";

function safeParse<T>(raw: string): T | null {
  try {
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

function normalizePhone(phone: string) {
  return String(phone || "").replace(/\s+/g, "").trim();
}

function normalizeList(list: AddressItem[]) {
  const items = (Array.isArray(list) ? list : [])
    .map((it) => ({
      id: String(it?.id || "").trim(),
      name: String(it?.name || "").trim(),
      phone: normalizePhone(String(it?.phone || "")),
      province: String(it?.province || "").trim() || undefined,
      city: String(it?.city || "").trim() || undefined,
      district: String(it?.district || "").trim() || undefined,
      detail: String(it?.detail || "").trim(),
      isDefault: Boolean(it?.isDefault),
    }))
    .filter((it) => it.id && it.name && it.phone && it.detail);

  // ensure only one default
  const firstDefault = items.find((x) => x.isDefault);
  if (!firstDefault && items.length) items[0].isDefault = true;
  if (firstDefault) {
    items.forEach((x) => {
      if (x.id !== firstDefault.id) x.isDefault = false;
    });
  }
  return items;
}

export function listAddresses(): AddressItem[] {
  const raw = String(uni.getStorageSync(KEY_LIST) || "").trim();
  const parsed = raw ? safeParse<AddressItem[]>(raw) : null;
  return normalizeList(parsed || []);
}

export function saveAddresses(items: AddressItem[]) {
  const normalized = normalizeList(items);
  uni.setStorageSync(KEY_LIST, JSON.stringify(normalized));
  return normalized;
}

export function getSelectedAddressId(): string {
  return String(uni.getStorageSync(KEY_SELECTED) || "").trim();
}

export function setSelectedAddressId(id: string) {
  uni.setStorageSync(KEY_SELECTED, String(id || "").trim());
}

export function getSelectedAddress(): AddressItem | null {
  const list = listAddresses();
  const selected = getSelectedAddressId();
  if (selected) {
    const hit = list.find((x) => x.id === selected);
    if (hit) return hit;
  }
  const def = list.find((x) => x.isDefault);
  return def || (list[0] || null);
}

export function maskPhone(phone: string) {
  const s = normalizePhone(phone);
  if (s.length < 7) return s || "—";
  return `${s.slice(0, 3)}****${s.slice(-4)}`;
}

export function formatFullAddress(a: AddressItem) {
  const parts = [a.province, a.city, a.district, a.detail].filter(Boolean);
  return parts.join("");
}

