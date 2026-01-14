import { clearSession } from "./session";

type MiniAppRequestOptions = {
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
  data?: any;
  headers?: Record<string, string>;
};

type Envelope<T> = {
  code?: number;
  message?: string;
  data?: T;
};

type ApiError = {
  code?: string;
  message?: string;
  details?: any;
};

type ApiResponse<T> = {
  success?: boolean;
  data?: T;
  message?: string;
  error?: ApiError | any;
  timestamp?: string;
  requestId?: string;
};

const DEFAULT_TENANT_UUID =
  (import.meta.env.VITE_MINIAPP_TENANT_UUID as string | undefined) ||
  "00000000-0000-0000-0000-000000000001";
const DEFAULT_BASE =
  (import.meta.env.VITE_MINIAPP_API_BASE as string | undefined) ||
  "http://localhost:8078/api/v1/mini-app";

function getBaseUrl() {
  const v = String(uni.getStorageSync("miniapp.api.base") || "").trim();
  return v || DEFAULT_BASE;
}

function getTenantUUID() {
  const v = String(uni.getStorageSync("miniapp.tenant.uuid") || "").trim();
  return v || DEFAULT_TENANT_UUID;
}

function getCustomerToken() {
  return String(uni.getStorageSync("miniapp.customer.token") || "").trim();
}

export async function miniAppRequest<T>(opts: MiniAppRequestOptions): Promise<T> {
  const url = `${getBaseUrl()}${opts.path}`;
  const isAuthEndpoint = String(opts.path || "").startsWith("/auth/");
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-Tenant-UUID": getTenantUUID(),
    ...(opts.headers || {}),
  };
  const token = getCustomerToken();
  // 登录/注册接口不应携带旧 token，避免误导排查（401 也应被视为“凭证错误”，而非“未登录”）。
  if (token && !isAuthEndpoint) headers["Authorization"] = `Bearer ${token}`;

  return await new Promise<T>((resolve, reject) => {
    uni.request({
      url,
      method: opts.method as any,
      data: opts.data,
      header: headers,
      success: (res: any) => {
        const body: Envelope<T> | T = res?.data;
        const httpStatus = Number(res?.statusCode || 0);

        if (body && typeof body === "object" && "code" in (body as any) && "data" in (body as any)) {
          const env = body as Envelope<T>;
          if ((env.code ?? 0) === 0) return resolve(env.data as T);
          if ((env.code ?? 0) === 401) {
            if (isAuthEndpoint) return reject(new Error("手机号或密码错误"));
            clearSession();
            return reject(new Error("请先登录"));
          }
          return reject(new Error(env.message || "request failed"));
        }

        if (body && typeof body === "object" && "success" in (body as any) && ("data" in (body as any) || "error" in (body as any))) {
          const api = body as ApiResponse<T>;
          if (api.success) return resolve(api.data as T);
          const errMsg =
            (api.error && typeof api.error === "object" ? (api.error as ApiError).message : "") ||
            (typeof api.error === "string" ? api.error : "") ||
            api.message ||
            "request failed";
          if (httpStatus === 401 || String(errMsg).toLowerCase().includes("unauthorized")) {
            if (isAuthEndpoint) return reject(new Error("手机号或密码错误"));
            clearSession();
            return reject(new Error("请先登录"));
          }
          return reject(new Error(errMsg));
        }

        if (httpStatus >= 200 && httpStatus < 300) return resolve(body as T);
        if (httpStatus === 401) {
          if (isAuthEndpoint) return reject(new Error("手机号或密码错误"));
          clearSession();
          return reject(new Error("请先登录"));
        }
        reject(new Error((body as any)?.message || (body as any)?.error || "request failed"));
      },
      fail: (err: any) => reject(new Error(err?.errMsg || "request failed")),
    });
  });
}
