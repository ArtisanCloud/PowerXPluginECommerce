import { useNuxtApp } from "#imports";

export interface CustomerMetricEvent {
  name: string;
  durationMs?: number;
  count?: number;
  status?: string;
  category?: string;
  metadata?: Record<string, any>;
  timestamp?: number;
}

export interface CustomerMetricsRecorder {
  record: (event: CustomerMetricEvent) => void;
}

const noopRecorder: CustomerMetricsRecorder = {
  record: () => undefined,
};

const now = () =>
  typeof performance !== "undefined" && typeof performance.now === "function"
    ? performance.now()
    : Date.now();

export const useCustomerMetrics = () => {
  const nuxtApp = useNuxtApp();
  const recorder = nuxtApp.$customerMetrics || noopRecorder;

  const recordEvent = (event: CustomerMetricEvent) => {
    recorder.record({
      ...event,
      category: event.category || "customer_ops",
    });
  };

  const startLatencyTimer = (
    name: string,
    metadata?: Record<string, any>
  ): (() => void) => {
    const start = now();
    return () => {
      const end = now();
      recordEvent({
        name,
        durationMs: end - start,
        metadata,
      });
    };
  };

  const trackJob = (
    name: string,
    metadata?: Record<string, any>
  ): ((status: "success" | "failed", extra?: Record<string, any>) => void) => {
    const start = Date.now();
    return (status, extra) => {
      recordEvent({
        name,
        status,
        durationMs: Date.now() - start,
        metadata: { ...metadata, ...extra },
      });
    };
  };

  const recordReminderResult = (params: {
    channel: string;
    total: number;
    success: number;
  }) => {
    recordEvent({
      name: "customer_membership_reminder",
      metadata: params,
    });
  };

  return {
    recordEvent,
    startLatencyTimer,
    trackJob,
    recordReminderResult,
  };
};
