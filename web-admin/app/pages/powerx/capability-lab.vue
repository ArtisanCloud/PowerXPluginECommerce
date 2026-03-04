<template>
  <div class="space-y-6">
    <header class="space-y-1">
      <h1 class="text-2xl font-semibold">PowerX Capability Lab</h1>
      <p class="text-sm text-gray-500">通过插件后端统一入口调用能力：`/api/v1/integration/capabilities/invoke`</p>
    </header>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between gap-3">
          <span class="font-medium">调用参数</span>
          <div class="text-xs text-gray-500">API Base: {{ apiBase }}</div>
        </div>
      </template>

      <div class="grid gap-4 md:grid-cols-2">
        <UFormField label="Capability ID" required>
          <UInput v-model="form.capabilityId" placeholder="com.corex.media.assets.manage" />
        </UFormField>
        <UFormField label="Action" required>
          <UInput v-model="form.action" placeholder="Create" />
        </UFormField>
        <UFormField label="Preferred Protocol">
          <USelect v-model="form.preferredProtocol" :items="protocolOptions" />
        </UFormField>
        <UFormField label="X-Request-ID">
          <UInput v-model="form.requestId" placeholder="可选，不填自动生成" />
        </UFormField>
        <UFormField label="X-PX-Use-Mock">
          <UInput v-model="form.mockModule" placeholder="可选，例如 media" />
        </UFormField>
      </div>

      <div class="mt-4">
        <UFormField label="Payload (JSON)" required>
          <UTextarea
            v-model="form.payloadText"
            :rows="14"
            class="font-mono text-xs"
            placeholder='{"method":"POST","endpoint":"/api/v1/media/assets","body":{}}'
          />
        </UFormField>
      </div>

      <template #footer>
        <div class="flex items-center gap-3">
          <UButton color="primary" :loading="submitting" @click="invokeCapability">调用</UButton>
          <UButton variant="soft" color="neutral" @click="resetTemplate">重置模板</UButton>
        </div>
      </template>
    </UCard>

    <UCard>
      <template #header>
        <span class="font-medium">调用结果</span>
      </template>

      <div v-if="errorMessage" class="mb-3 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
        {{ errorMessage }}
      </div>

      <div class="grid gap-3 md:grid-cols-3">
        <div class="rounded border p-3 text-sm">
          <div class="text-gray-500">Trace ID</div>
          <div class="break-all font-mono">{{ result?.traceId || "-" }}</div>
        </div>
        <div class="rounded border p-3 text-sm">
          <div class="text-gray-500">Status</div>
          <div class="font-medium">{{ result?.status || "-" }}</div>
        </div>
        <div class="rounded border p-3 text-sm">
          <div class="text-gray-500">Warnings</div>
          <div class="break-all">{{ result?.warnings?.join(" | ") || "-" }}</div>
        </div>
      </div>

      <div class="mt-4">
        <UFormField label="Raw Response">
          <UTextarea :model-value="resultText" :rows="16" readonly class="font-mono text-xs" />
        </UFormField>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { resolveApiBase } from "~/composables/api/_base";
import { PowerXCapabilityBridgeError } from "~/plugins/powerx-capability.client";

definePageMeta({
  title: "PowerXCapabilityLab",
});

const { invoke } = usePowerXCapability();
const apiBase = computed(() => resolveApiBase());

const protocolOptions = ["", "rest", "grpc", "workflow", "agent"];

const defaultPayload = () =>
  JSON.stringify(
    {
      method: "POST",
      endpoint: "/api/v1/media/assets",
      headers: {
        "Content-Type": "application/json",
      },
      query: {
        page: 1,
        page_size: 20,
      },
      body: {
        assetName: "demo.pdf",
        uploadMethod: "presign_upload",
      },
    },
    null,
    2
  );

const form = reactive({
  capabilityId: "",
  action: "",
  preferredProtocol: "rest",
  requestId: "",
  mockModule: "",
  payloadText: defaultPayload(),
});

const submitting = ref(false);
const errorMessage = ref("");
const result = ref<any | null>(null);

const resultText = computed(() => {
  if (!result.value) {
    return "";
  }
  try {
    return JSON.stringify(result.value, null, 2);
  } catch {
    return String(result.value);
  }
});

function resetTemplate() {
  form.payloadText = defaultPayload();
}

async function invokeCapability() {
  errorMessage.value = "";
  result.value = null;
  let payload: Record<string, any>;
  try {
    payload = JSON.parse(form.payloadText || "{}");
  } catch (error) {
    errorMessage.value = `Payload 不是合法 JSON: ${error instanceof Error ? error.message : "parse error"}`;
    return;
  }

  const headers: Record<string, string> = {};
  if (form.mockModule.trim()) {
    headers["X-PX-Use-Mock"] = form.mockModule.trim();
  }

  submitting.value = true;
  try {
    const response = await invoke({
      capabilityId: form.capabilityId,
      action: form.action,
      preferredProtocol: form.preferredProtocol || undefined,
      requestId: form.requestId || undefined,
      payload,
      headers,
      apiBase: apiBase.value,
    });
    result.value = response;
  } catch (error) {
    if (error instanceof PowerXCapabilityBridgeError) {
      errorMessage.value = `${error.message}${error.traceId ? ` (traceId: ${error.traceId})` : ""}`;
      result.value = {
        status: error.status,
        traceId: error.traceId,
        details: error.details,
        warnings: error.warnings,
      };
      return;
    }
    errorMessage.value = error instanceof Error ? error.message : "capability invoke failed";
  } finally {
    submitting.value = false;
  }
}
</script>
