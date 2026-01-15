import { getSession } from "./session";
import { syncCart, type CartDTO } from "./miniapp-cart";

// v2：不兼容旧的本地购物车数据（旧数据缺少价格等字段会导致合计不正确）
const CART_ITEMS_KEY = "miniapp.cart.v2.items";
const CART_UPDATED_AT_KEY = "miniapp.cart.v2.updatedAt";
const CART_LAST_SYNC_AT_KEY = "miniapp.cart.v2.lastSyncAt";

export type LocalCartItem = {
  skuId: string;
  spuId?: string;
  qty: number;
  selected?: boolean;
  title?: string;
  imageUrl?: string;
  skuLabel?: string;
  skuCode?: string;
  maxQty?: number;
  currency?: string;
  unitPrice?: number;
};

export type LocalCart = {
  items: LocalCartItem[];
  updatedAt?: string;
  lastSyncAt?: string;
};

function nowIso() {
  return new Date().toISOString();
}

function stripPriceFromSkuLabel(input?: string) {
  const raw = String(input || "").trim();
  if (!raw) return undefined;

  // 1) remove standalone currency lines like "¥299"/"￥299"/"CNY 299" (wrap may split into multiple lines)
  const lines = raw
    .split(/\r?\n/)
    .map((s) => s.trim())
    .filter(Boolean)
    .filter((line) => {
      const l = line.trim();
      if (!l) return false;
      if (/^([¥￥$€])\s*\d/.test(l)) return false;
      if (/^(CNY|RMB|USD|EUR)\s*[:：]?\s*\d/i.test(l)) return false;
      // also strip lines that are only a currency+number (with optional decimals)
      if (/^([¥￥$€])\s*\d+(\.\d+)?$/.test(l)) return false;
      if (/^(CNY|RMB|USD|EUR)\s*\d+(\.\d+)?$/i.test(l)) return false;
      return true;
    });

  let s = lines.join(" ").replace(/\s+/g, " ").trim();
  if (!s) return undefined;

  // 2) if label ends with "· <price>", strip the trailing price part (keep other content)
  // examples:
  //   "APP-JERSEY-001-M · ¥299" -> "APP-JERSEY-001-M"
  //   "APP-JERSEY-001-M · 299"  -> "APP-JERSEY-001-M"
  const parts = s.split("·").map((p) => p.trim()).filter(Boolean);
  if (parts.length >= 2) {
    const last = parts[parts.length - 1];
    if (/^[¥￥$€]\s*\d/.test(last) || /^(CNY|RMB|USD|EUR)\s*\d/i.test(last) || /^\d+(\.\d+)?\s*(CNY|RMB|USD|EUR)?$/i.test(last)) {
      parts.pop();
      s = parts.join(" · ").trim();
    }
  }
  return s || undefined;
}

function normalize(items: LocalCartItem[]) {
  const DEFAULT_MAX_QTY = 10;
  const map = new Map<string, LocalCartItem>();
  for (const it of items || []) {
    const skuId = String(it?.skuId || "").trim();
    const qty = Number(it?.qty || 0);
    if (!skuId || !Number.isFinite(qty) || qty <= 0) continue;
    const prev = map.get(skuId);
    if (!prev) {
      map.set(skuId, { ...it, skuId, qty, skuLabel: stripPriceFromSkuLabel(it.skuLabel) });
      continue;
    }
    map.set(skuId, {
      skuId,
      qty: (prev.qty || 0) + qty,
      selected: typeof it.selected === "boolean" ? it.selected : (typeof prev.selected === "boolean" ? prev.selected : true),
      spuId: it.spuId || prev.spuId,
      title: it.title || prev.title,
      imageUrl: it.imageUrl || prev.imageUrl,
      skuLabel: stripPriceFromSkuLabel(it.skuLabel || prev.skuLabel),
      skuCode: it.skuCode || prev.skuCode,
      maxQty: Number.isFinite(Number(it.maxQty)) ? Number(it.maxQty) : prev.maxQty,
      currency: String(it.currency || "").trim() || prev.currency,
      unitPrice: Number.isFinite(Number(it.unitPrice)) ? Number(it.unitPrice) : prev.unitPrice,
    });
  }
  return Array.from(map.values()).map((it) => {
    const maxQty = Number(it.maxQty);
    const clampedMax = Number.isFinite(maxQty) && maxQty > 0 ? Math.floor(maxQty) : undefined;
    const enforcedMax = Math.max(1, Math.min(99, clampedMax ?? DEFAULT_MAX_QTY));
    const nextQty = Math.max(1, Math.min(enforcedMax, Math.floor(Number(it.qty) || 1)));
    const unitPrice = Number(it.unitPrice);
    const normalizedPrice = Number.isFinite(unitPrice) && unitPrice >= 0 ? unitPrice : undefined;
    const currency = String(it.currency || "").trim() || undefined;
    const selected = typeof it.selected === "boolean" ? it.selected : true;
    return { ...it, qty: nextQty, selected, skuLabel: stripPriceFromSkuLabel(it.skuLabel), maxQty: enforcedMax, unitPrice: normalizedPrice, currency };
  });
}

export function getLocalCart(): LocalCart {
  try {
    const raw = String(uni.getStorageSync(CART_ITEMS_KEY) || "").trim();
    const items = raw ? (JSON.parse(raw) as LocalCartItem[]) : [];
    const normalized = normalize(items);
    // one-time migration: persist normalized items (without touching updatedAt)
    try {
      const nextRaw = JSON.stringify(normalized);
      if (raw && raw !== nextRaw) {
        uni.setStorageSync(CART_ITEMS_KEY, nextRaw);
      }
    } catch {}
    return {
      items: normalized,
      updatedAt: String(uni.getStorageSync(CART_UPDATED_AT_KEY) || "").trim() || undefined,
      lastSyncAt: String(uni.getStorageSync(CART_LAST_SYNC_AT_KEY) || "").trim() || undefined,
    };
  } catch {
    return { items: [] };
  }
}

export function setLocalCart(items: LocalCartItem[], updatedAt?: string) {
  const normalized = normalize(items);
  uni.setStorageSync(CART_ITEMS_KEY, JSON.stringify(normalized));
  uni.setStorageSync(CART_UPDATED_AT_KEY, updatedAt || nowIso());
}

export function clearLocalCart() {
  try {
    // 兼容清理旧 key（仅用于清除，不做读取兼容）
    uni.removeStorageSync("miniapp.cart.items");
    uni.removeStorageSync("miniapp.cart.updatedAt");
    uni.removeStorageSync("miniapp.cart.lastSyncAt");
    uni.removeStorageSync(CART_ITEMS_KEY);
    uni.removeStorageSync(CART_UPDATED_AT_KEY);
    uni.removeStorageSync(CART_LAST_SYNC_AT_KEY);
  } catch {}
}

export function cartCount(items?: LocalCartItem[]) {
  return (items || []).reduce((sum, it) => sum + Number(it.qty || 0), 0);
}

export function addCartItem(skuId: string, qty: number, meta: Partial<Omit<LocalCartItem, "skuId" | "qty">> = {}) {
  const id = String(skuId || "").trim();
  const q = Number(qty || 0);
  if (!id || !Number.isFinite(q) || q <= 0) return getLocalCart();
  const cart = getLocalCart();
  const next = normalize([...cart.items, { skuId: id, qty: q, ...meta }]);
  setLocalCart(next);
  return { ...cart, items: next, updatedAt: nowIso() };
}

export function setCartItemQty(skuId: string, qty: number) {
  const id = String(skuId || "").trim();
  const q = Number(qty || 0);
  if (!id) return getLocalCart();
  const cart = getLocalCart();
  const next = normalize(
    cart.items
      .map((it) => (it.skuId === id ? { ...it, skuId: it.skuId, qty: q } : it))
      .filter((it) => Number(it.qty || 0) > 0),
  );
  setLocalCart(next);
  return { ...cart, items: next, updatedAt: nowIso() };
}

export function removeCartItem(skuId: string) {
  const id = String(skuId || "").trim();
  const cart = getLocalCart();
  const next = normalize(cart.items.filter((it) => it.skuId !== id));
  setLocalCart(next);
  return { ...cart, items: next, updatedAt: nowIso() };
}

export async function syncCartHybrid(strategy: "max" | "overwrite" = "max"): Promise<CartDTO | null> {
  const session = getSession();
  if (!session?.token) return null;
  const local = getLocalCart();
  const selectedBySku = new Map(local.items.map((it) => [it.skuId, typeof it.selected === "boolean" ? it.selected : true] as const));
  const resp = await syncCart({ items: local.items.map(({ skuId, qty }) => ({ skuId, qty })), strategy });
  setLocalCart(
    resp.items.map((it) => ({ ...it, selected: selectedBySku.get(String(it.skuId || "").trim()) ?? true })),
    resp.updatedAt,
  );
  uni.setStorageSync(CART_LAST_SYNC_AT_KEY, nowIso());
  return resp;
}
