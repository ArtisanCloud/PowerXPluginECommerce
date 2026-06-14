export interface PowerXCapabilityRequest {
  capabilityId: string;
  action: string;
  payload?: Record<string, any> | null;
  headers?: Record<string, string>;
  requestId?: string;
  signal?: AbortSignal;
  apiBase?: string;
  endpoint?: string;
  preferredProtocol?: string;
  metadata?: Record<string, any>;
}

export interface PowerXCapabilityResponse {
  traceId?: string;
  status?: string;
  data?: Record<string, any> | null;
  errors?: Record<string, any> | Record<string, any>[] | null;
  warnings?: string[] | null;
  raw?: Record<string, any> | Record<string, any>[] | string | null;
}

export class PowerXCapabilityBridgeError extends Error {
  status?: number;
  traceId?: string;
  details?: any;
  warnings?: string[];

  constructor(
    message: string,
    opts: { status?: number; traceId?: string; details?: any; warnings?: string[] } = {}
  ) {
    super(message);
    this.name = "PowerXCapabilityBridgeError";
    this.status = opts.status;
    this.traceId = opts.traceId;
    this.details = opts.details;
    this.warnings = opts.warnings;
  }
}

export interface PowerXCapabilityBridge {
  invoke(request: PowerXCapabilityRequest): Promise<PowerXCapabilityResponse>;
}

