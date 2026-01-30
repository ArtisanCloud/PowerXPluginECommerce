type MiniAppRequestOptions = {
  method: "GET" | "POST";
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

export type PaymentCreateRequest = {
  orderId: string;
  payMethod: string;
  idempotencyKey: string;
  providerId: number;
  client?: string;
};

export type WechatPayParams = {
  appId: string;
  timeStamp: string;
  nonceStr: string;
  package: string;
  signType: string;
  paySign: string;
};

export type PaymentCreateResponse = {
  transactionId: string;
  status: string;
  wechat?: WechatPayParams;
};

export type PaymentStatusResponse = {
  transactionId: string;
  orderId: string;
  orderNo: string;
  status: string;
  failureReason?: string;
  updatedAt?: string;
};

export type PaymentOutcome = {
  status: "success" | "cancel" | "fail";
  message?: string;
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

function getMiniAppBaseUrl() {
  return getBaseUrl();
}

function getTenantUUID() {
  const v = String(uni.getStorageSync("miniapp.tenant.uuid") || "").trim();
  return v || DEFAULT_TENANT_UUID;
}

function getCustomerToken() {
  return String(uni.getStorageSync("miniapp.customer.token") || "").trim();
}

export function getMiniAppOpenID() {
  return String(uni.getStorageSync("miniapp.customer.openid") || "").trim();
}

export function getWechatProviderId() {
  const stored = String(uni.getStorageSync("miniapp.wechat.providerId") || "").trim();
  const value = stored || String(import.meta.env.VITE_MINIAPP_WECHAT_PROVIDER_ID || "").trim();
  return value;
}

export function getWechatProviderIdNumber() {
  const raw = String(getWechatProviderId() || "").trim();
  const num = Number(raw);
  return Number.isFinite(num) && num > 0 ? num : 0;
}

export function buildIdempotencyKey(prefix = "pay") {
  const rand = Math.random().toString(36).slice(2, 10);
  return `${prefix}-${Date.now()}-${rand}`;
}

async function miniAppRequest<T>(opts: MiniAppRequestOptions): Promise<T> {
  const url = `${getMiniAppBaseUrl()}${opts.path}`;
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-Tenant-UUID": getTenantUUID(),
    ...(opts.headers || {}),
  };
  const token = getCustomerToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;

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
          return reject(new Error(errMsg));
        }

        if (httpStatus >= 200 && httpStatus < 300) return resolve(body as T);
        reject(new Error((body as any)?.message || (body as any)?.error || "request failed"));
      },
      fail: (err: any) => reject(new Error(err?.errMsg || "request failed")),
    });
  });
}

export async function createPaymentTransaction(req: PaymentCreateRequest) {
  return await miniAppRequest<PaymentCreateResponse>({
    method: "POST",
    path: "/payments/transactions",
    data: req,
  });
}

export async function getPaymentTransactionStatus(id: string) {
  const tid = encodeURIComponent(String(id || "").trim());
  return await miniAppRequest<PaymentStatusResponse>({
    method: "GET",
    path: `/payments/transactions/${tid}`,
  });
}

export async function requestMiniAppPayment(params: WechatPayParams): Promise<PaymentOutcome> {
  if (!params || !params.appId) return { status: "fail", message: "支付参数缺失" };
  return await new Promise<PaymentOutcome>((resolve) => {
    uni.requestPayment({
      provider: "wxpay",
      ...params,
      success: () => resolve({ status: "success" }),
      fail: (err: any) => {
        const msg = String(err?.errMsg || "支付失败");
        const lower = msg.toLowerCase();
        if (lower.includes("cancel")) return resolve({ status: "cancel", message: msg });
        resolve({ status: "fail", message: msg });
      },
    });
  });
}
