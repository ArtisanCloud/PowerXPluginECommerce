// 统一创建 $fetch 实例（单例）+ 便捷方法

import { resolveApiBase, getAuthToken, getTenantUuid } from "./_base";
import { useAuth } from "~/composables/useAuth";
import { useRouter } from "vue-router";
import { useToastAlert } from "../useToastAlert";
import { useHostCtxStore } from "~/stores/hostCtx";
import { PLUGIN_ID } from "~/utils/powerx-bridge";

type Json = Record<string, any>;

let _client: typeof $fetch | null = null;
let _baseURL: string | null = null;
let _clientEnv: "client" | "server" | null = null;

export function useApiClient() {
  const env = typeof window === "undefined" ? "server" : "client";
  if (_client && _clientEnv === env) {
    return { client: _client, baseURL: _baseURL! };
  }

  const baseURL = resolveApiBase();
  const router = (() => {
    if (typeof window === "undefined") return null;
    try {
      return useRouter();
    } catch {
      return null;
    }
  })();
  _baseURL = baseURL;

  const baseClient = $fetch.create({
    baseURL,
    timeout: 30_000,
  });
  const hostCtxStore = process.client ? useHostCtxStore() : null;

  const prepareOptions = async (options?: Record<string, any>) => {
    const next: Record<string, any> = options ? { ...options } : {};
    const headers =
      options?.headers instanceof Headers
        ? new Headers(options.headers)
        : new Headers((options?.headers as HeadersInit) || undefined);
    next.headers = headers;
    const skipAuth = Boolean((options as any)?.skipAuth);
    const auth = useAuth();
    const pluginOrigin =
      typeof window !== "undefined" ? window.location.origin : "plugin";
    const requestPluginId =
      (next as any)?.pluginId ||
      process.env.NUXT_PUBLIC_PLUGIN_ID ||
      PLUGIN_ID;
    const ctxKey =
      hostCtxStore && typeof window !== "undefined"
        ? `${pluginOrigin}::${requestPluginId}`
        : null;

    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }

    const isFormData =
      next.body &&
      typeof FormData !== "undefined" &&
      next.body instanceof FormData;

    if (!isFormData && !headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }

    let authToken: string | null = null;
    if (!skipAuth) {
      authToken = (next as any).authToken || (next as any).token;
      if (!authToken) {
        authToken = (await auth.ensureFreshToken()) || getAuthToken() || null;
      }
    }

    const existingAuthHeader = headers.get("Authorization");
    if (authToken && (!existingAuthHeader || !existingAuthHeader.trim())) {
      headers.set(
        "Authorization",
        /^Bearer\\s/i.test(String(authToken))
          ? String(authToken)
          : `Bearer ${authToken}`
      );
    }

    const ctxPayload = ctxKey ? hostCtxStore?.getCtx(ctxKey) : null;
    const debugCtx =
      process.env.NUXT_PUBLIC_BRIDGE_DEBUG === "true" ||
      (typeof window !== "undefined" && (window as any).__PX_DEBUG_CTX__);
    if (ctxPayload?.ctx && !headers.has("X-PowerX-CTX")) {
      headers.set("X-PowerX-CTX", ctxPayload.ctx);
    }
    if (ctxPayload?.ctxSig && !headers.has("X-PowerX-CTX-SIG")) {
      headers.set("X-PowerX-CTX-SIG", ctxPayload.ctxSig);
    }
    if (ctxPayload?.ctxJwt && !headers.has("X-PowerX-CTX-JWT")) {
      headers.set("X-PowerX-CTX-JWT", ctxPayload.ctxJwt);
    }
    if (debugCtx && ctxKey) {
      console.info("[Plugin][api] ctx headers", {
        key: ctxKey,
        hasCtx: Boolean(ctxPayload?.ctx),
        hasCtxSig: Boolean(ctxPayload?.ctxSig),
        hasCtxJwt: Boolean(ctxPayload?.ctxJwt),
        headers: {
          ctx: headers.get("X-PowerX-CTX") ? "yes" : "no",
          ctxSig: headers.get("X-PowerX-CTX-SIG") ? "yes" : "no",
          ctxJwt: headers.get("X-PowerX-CTX-JWT") ? "yes" : "no",
        },
      });
    }

    if (!headers.has("X-Tenant-UUID")) {
      const tenant = (next as any).tenantUuid || getTenantUuid();
      if (tenant) {
        headers.set("X-Tenant-UUID", String(tenant));
      }
    }

    const debugAuth =
      process.env.NUXT_PUBLIC_BRIDGE_DEBUG === "true" ||
      (typeof window !== "undefined" && (window as any).__PX_DEBUG_AUTH__);
    if (debugAuth && typeof window !== "undefined") {
      const lsToken = (() => {
        try {
          return window.localStorage?.getItem("access_token") || "";
        } catch {
          return "";
        }
      })();
      console.info("[Plugin][api] auth header prepared", {
        hasAuthorization: Boolean(headers.get("Authorization")?.trim()),
        hasLocalStorageToken: Boolean(lsToken),
        tokenPreview: lsToken ? `${lsToken.slice(0, 4)}...${lsToken.slice(-4)}` : "",
        url: String((next as any)?.url || ""),
      });
    }

    return next;
  };

  const toast = process.client ? useToastAlert() : null;

  const handleAuthError = (response?: { status?: number; _data?: any }) => {
    if (!response) return;
    const auth = useAuth();
    if (response.status === 503) {
      console.error("API error:", response.status, response._data);
      const message = response._data?.message || "宿主认证不可用，请稍后重试";
      auth.failClosed?.(message);
      if (process.client) {
        const redirect = window.location.pathname + window.location.search;
        router?.push({ path: "/users/login", query: { redirect } });
      }
      return;
    }
    if (response.status === 401) {
      console.error("API error:", response.status, response._data);
      auth.clearAuth();
      if (process.client) {
        toast?.add?.({
          title: "登录状态已失效",
          description: "请重新登录后再试",
          color: "red",
        });
      }
    }
  };

  const invokeClient = async (request: any, options?: any, raw = false) => {
    const prepared = await prepareOptions(options);
    try {
      return await (raw
        ? baseClient.raw(request, prepared)
        : baseClient(request, prepared));
    } catch (error: any) {
      handleAuthError(error?.response);
      const responseMessage =
        error?.response?._data?.message ||
        error?.response?._data?.error?.message ||
        error?.data?.message;
      if (responseMessage) {
        error.message = responseMessage;
        if (!error.data) {
          error.data = {} as any;
        }
        if (!error.data.message) {
          error.data.message = responseMessage;
        }
      }
      throw error;
    }
  };

  const client = ((request: any, options?: any) =>
    invokeClient(request, options)) as typeof baseClient;

  client.raw = (request: any, options?: any) => invokeClient(request, options, true);
  client.native = baseClient.native;

  _client = client;
  _clientEnv = env;
  return { client, baseURL };
}

// 常用 CRUD 便捷封装
export function apiGet<T>(path: string, query?: Json, init?: any) {
  const { client } = useApiClient();
  return client<T>(path, { method: "GET", query, ...init });
}
export function apiPost<T>(path: string, body?: any, init?: any) {
  const { client } = useApiClient();
  const payload =
    body instanceof FormData ? body : body ? JSON.stringify(body) : undefined;
  return client<T>(path, { method: "POST", body: payload, ...init });
}
export function apiPut<T>(path: string, body?: any, init?: any) {
  const { client } = useApiClient();
  const payload =
    body instanceof FormData ? body : body ? JSON.stringify(body) : undefined;
  return client<T>(path, { method: "PUT", body: payload, ...init });
}
export function apiPatch<T>(path: string, body?: any, init?: any) {
  const { client } = useApiClient();
  const payload =
    body instanceof FormData ? body : body ? JSON.stringify(body) : undefined;
  return client<T>(path, { method: "PATCH", body: payload, ...init });
}
export function apiDel<T>(path: string, init?: any) {
  const { client } = useApiClient();
  return client<T>(path, { method: "DELETE", ...init });
}
