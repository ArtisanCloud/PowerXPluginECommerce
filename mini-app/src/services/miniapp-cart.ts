import { miniAppRequest } from "./miniapp-client";

export type CartItem = {
  skuId: string;
  qty: number;
};

export type CartDTO = {
  items: CartItem[];
  updatedAt: string;
};

export type CartSyncRequest = {
  items: CartItem[];
  strategy?: "max" | "overwrite";
};

export async function getCart() {
  return await miniAppRequest<CartDTO>({ method: "GET", path: "/cart" });
}

export async function syncCart(payload: CartSyncRequest) {
  return await miniAppRequest<CartDTO>({
    method: "POST",
    path: "/cart/sync",
    data: payload,
  });
}

