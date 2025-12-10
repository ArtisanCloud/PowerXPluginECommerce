import type {
  CustomerMetricEvent,
  CustomerMetricsRecorder,
} from "~/composables/useCustomerMetrics";

declare global {
  interface Window {
    __pxMetrics?: {
      events: CustomerMetricEvent[];
    };
  }
}

const createRecorder = (): CustomerMetricsRecorder => {
  if (typeof window === "undefined") {
    return {
      record: () => undefined,
    };
  }
  const container =
    (window.__pxMetrics ||= { events: [] as CustomerMetricEvent[] });
  return {
    record: (event: CustomerMetricEvent) => {
      container.events.push({
        ...event,
        timestamp: event.timestamp ?? Date.now(),
        category: event.category ?? "customer_ops",
      });
    },
  };
};

declare module "#app" {
  interface NuxtApp {
    $customerMetrics: CustomerMetricsRecorder;
  }
}

declare module "vue" {
  interface ComponentCustomProperties {
    $customerMetrics: CustomerMetricsRecorder;
  }
}

export default defineNuxtPlugin((nuxtApp) => {
  const recorder = createRecorder();
  nuxtApp.provide("customerMetrics", recorder);

  if (typeof window === "undefined" || typeof performance === "undefined") {
    return;
  }

  const startTime = performance.now();
  let recorded = false;

  nuxtApp.hook("page:finish", () => {
    const path = nuxtApp.$router?.currentRoute.value.fullPath || "";
    if (!recorded && path.includes("/admin/integration/marketplace/dashboard")) {
      const duration = performance.now() - startTime;
      recorder.record({
        name: "marketplace_dashboard_first_paint",
        durationMs: duration,
        category: "marketplace",
      });
      recorded = true;
    }
  });
});
