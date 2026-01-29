<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">支付渠道</h1>
        <p class="text-gray-500 dark:text-gray-400">
          管理线上收单、分期、钱包等支付服务，监控费率与 SLA。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-document-arrow-down">
          导出配置
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus" @click="openCreate">新增渠道</UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <UCard v-for="card in summaryCards" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
          {{ card.value }}
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ card.subtitle }}</p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">支付服务商</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">查看接入状态、费率及 SLA。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-52"
              placeholder="搜索渠道/联系人"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="typeFilter"
              class="w-40"
              :items="typeItems"
              placeholder="支付类型"
            />
            <USelect
              v-model="statusFilter"
              class="w-40"
              :items="statusItems"
              placeholder="服务状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredProviders" :loading="loading">
        <template #name-cell="{ getValue }">
          <span>{{ getValue() }}</span>
        </template>
        <template #mchId-cell="{ getValue }">
          <span>{{ getValue() || "-" }}</span>
        </template>
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #feeRate-cell="{ getValue }">{{ formatFeeRate(getValue()) }}</template>
        <template #updatedAt-cell="{ getValue }">{{ fmtDT(getValue()) }}</template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              size="xs"
              variant="ghost"
              :disabled="!isWechat(row.original)"
              @click="openConfig(row.original)"
            >
              配置
            </UButton>
            <UButton size="xs" variant="ghost" color="primary">查看账单</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UModal
      v-model:open="configOpen"
      :title="'微信支付配置'"
      :description="'仅微信支付可配置，其他渠道暂不可用。'"
      :close="{ onClick: closeConfig }"
      :prevent-close="saving"
      :ui="{ content: 'max-w-4xl w-full', body: 'p-4 sm:p-5', footer: 'justify-end' }"
    >
      <template #body>
        <UForm id="wechat-provider-form" class="space-y-4" :state="configForm" @submit="saveConfig">
          <UAlert v-if="configError" color="red" variant="subtle" :title="configError" />
          <div class="grid gap-4 md:grid-cols-2">
            <UFormField label="渠道名称">
              <UInput v-model="configForm.name" disabled />
            </UFormField>
            <UFormField label="状态">
              <USelect v-model="configForm.status" :items="statusOptions" />
            </UFormField>
            <UFormField label="AppID">
              <UInput v-model="configForm.appId" placeholder="wx..." />
            </UFormField>
            <UFormField label="AppSecret">
              <div class="flex items-center gap-2">
                <UInput v-model="configForm.appSecret" type="password" placeholder="********" />
                <UButton
                  color="neutral"
                  variant="outline"
                  size="xs"
                  :loading="testingMiniApp"
                  @click="testMiniAppConfig"
                >
                  校验
                </UButton>
                <UBadge v-if="appSecretConfigured" color="success" variant="subtle">已配置</UBadge>
              </div>
            </UFormField>
            <UFormField label="HTTP 调试">
              <div class="flex items-center gap-2">
                <USwitch v-model="configForm.httpDebug" />
                <span class="text-xs text-gray-500 dark:text-gray-400">仅该渠道生效</span>
              </div>
            </UFormField>
            <UFormField label="商户号 (mchId)">
              <UInput v-model="configForm.mchId" placeholder="1900000109" />
            </UFormField>
            <UFormField label="证书序列号 (serialNo)">
              <div class="flex items-center gap-2">
                <UInput v-model="configForm.serialNo" placeholder="XXXXXXXXXXXX" />
                <UButton
                  color="neutral"
                  variant="outline"
                  size="xs"
                  :loading="testingCert"
                  @click="testCertSerial"
                >
                  校验
                </UButton>
              </div>
            </UFormField>
            <UFormField label="API v3 Key">
              <UInput v-model="configForm.apiV3Key" type="password" size="lg" placeholder="****************" />
            </UFormField>
            <UFormField label="回调 Host">
              <UInput
                v-model="configForm.notifyHost"
                placeholder="https://debug-ecommerce.artisan-cloud.com"
              />
            </UFormField>
            <UFormField label="回调地址 (notifyUrl)" class="md:col-span-2 w-full">
              <div class="w-full">
                <UInput
                  v-model="configForm.notifyUrl"
                  class="w-full"
                  size="lg"
                  placeholder="https://example.com/api/v1/mini-app/payments/providers/wechat/1616273230/wxc1ebbc8236ffae2b/callback"
                />
                <div class="mt-2 space-y-2">
                  <div class="text-xs text-gray-500 dark:text-gray-400">回调 URL 预览（输入 host 自动生成）</div>
                  <div class="flex flex-wrap items-center gap-2">
                    <div v-if="notifyUrlPreview" class="flex-1 min-w-[220px] text-xs text-gray-300">
                      {{ notifyUrlPreview }}
                    </div>
                    <div v-else class="flex-1 min-w-[220px] text-xs text-gray-500">—</div>
                    <UButton
                      color="neutral"
                      variant="outline"
                      size="xs"
                      :disabled="!notifyUrlPreview"
                      @click="applyNotifyUrlPreview"
                    >
                      填入
                    </UButton>
                    <UButton
                      color="neutral"
                      variant="outline"
                      size="xs"
                      :disabled="!notifyUrlPreview"
                      @click="copyNotifyUrlPreview"
                    >
                      复制
                    </UButton>
                  </div>
                </div>
              </div>
            </UFormField>
            <div class="md:col-span-2 grid gap-6 md:grid-cols-2">
              <div class="rounded-lg border border-gray-200/60 p-4 dark:border-gray-700/60">
                <div class="flex items-center justify-between">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">商户私钥 (privateKeyPem)</span>
                  <UBadge v-if="privateKeyConfigured" color="success" variant="subtle">已配置</UBadge>
                </div>
                <UTextarea
                  v-model="configForm.privateKeyPem"
                  :rows="6"
                  class="mt-2"
                  placeholder="-----BEGIN PRIVATE KEY-----"
                />
                <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <span>文本粘贴与文件上传二选一</span>
                </div>
                <div class="mt-3 flex flex-wrap items-center gap-2">
                  <input
                    ref="privateKeyFileInput"
                    type="file"
                    class="hidden"
                    accept=".pem,.key,text/plain"
                    @change="onPrivateKeyFileChange"
                  />
                  <UButton color="neutral" variant="outline" @click="triggerPrivateKeyFile">
                    选择文件
                  </UButton>
                  <span class="text-xs text-gray-500 dark:text-gray-400">
                    {{ privateKeyFileName || "未选择文件" }}
                  </span>
                </div>
              </div>
              <div class="rounded-lg border border-gray-200/60 p-4 dark:border-gray-700/60">
                <div class="flex items-center justify-between">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">商户证书 (certPem)</span>
                  <UBadge v-if="certConfigured" color="success" variant="subtle">已配置</UBadge>
                </div>
                <UTextarea
                  v-model="configForm.certPem"
                  :rows="6"
                  class="mt-2"
                  placeholder="-----BEGIN CERTIFICATE-----"
                />
                <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <span>文本粘贴与文件上传二选一</span>
                </div>
                <div class="mt-3 flex flex-wrap items-center gap-2">
                  <input
                    ref="certFileInput"
                    type="file"
                    class="hidden"
                    accept=".pem,.crt,text/plain"
                    @change="onCertFileChange"
                  />
                  <UButton color="neutral" variant="outline" @click="triggerCertFile">
                    选择文件
                  </UButton>
                  <span class="text-xs text-gray-500 dark:text-gray-400">
                    {{ certFileName || "未选择文件" }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </UForm>
      </template>

      <template #footer>
        <div class="flex items-center justify-end gap-2">
          <UButton color="neutral" variant="ghost" @click="closeConfig">取消</UButton>
          <UButton color="primary" :loading="saving" type="submit" form="wechat-provider-form">保存配置</UButton>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="createOpen"
      :title="'新增支付渠道'"
      :description="'新渠道接入暂未开放，仅提供表单占位。'"
      :close="{ onClick: closeCreate }"
      :ui="{ content: 'max-w-2xl w-full', body: 'p-4 sm:p-5', footer: 'justify-end' }"
    >
      <template #body>
        <UForm id="create-provider-form" class="space-y-4" :state="createForm" @submit="saveCreate">
          <div class="grid gap-4 md:grid-cols-2">
            <UFormField label="渠道名称">
              <UInput v-model="createForm.name" placeholder="请输入渠道名称" />
            </UFormField>
            <UFormField label="类型">
              <USelect v-model="createForm.type" :items="createTypeOptions" />
            </UFormField>
            <UFormField label="状态">
              <USelect v-model="createForm.status" :items="statusOptions" />
            </UFormField>
            <UFormField label="费率(%)">
              <UInput v-model="createForm.feeRate" placeholder="0.60" />
            </UFormField>
            <UFormField label="币种">
              <UInput v-model="createForm.currency" placeholder="CNY" />
            </UFormField>
            <UFormField label="结算周期">
              <UInput v-model="createForm.settlementCycle" placeholder="daily" />
            </UFormField>
          </div>
        </UForm>
      </template>

      <template #footer>
        <div class="flex items-center justify-end gap-2">
          <UButton color="neutral" variant="ghost" @click="closeCreate">取消</UButton>
          <UButton color="primary" type="submit" form="create-provider-form" :loading="createSaving">
            保存
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { usePaymentsApi } from "~/composables/api";
import { useToastAlert } from "~/composables/useToastAlert";
import type { PaymentProvider, PaymentProviderDetail } from "~/types/payments";

definePageMeta({
  name: "payments-providers",
});

const { listProviders, getProviderDetail, updateProvider, testMiniApp, testWechatCertSerial } = usePaymentsApi();
const toast = useToastAlert();
const loading = ref(false);
const providers = ref<PaymentProvider[]>([]);
const configOpen = ref(false);
const saving = ref(false);
const testingMiniApp = ref(false);
const testingCert = ref(false);
const configError = ref("");
const selectedProvider = ref<PaymentProvider | null>(null);
const createOpen = ref(false);
const createSaving = ref(false);
const createForm = reactive({
  name: "",
  type: "wechat",
  status: "active",
  feeRate: "",
  currency: "CNY",
  settlementCycle: "daily",
});
const configForm = reactive({
  name: "",
  status: "active",
  appId: "",
  appSecret: "",
  httpDebug: false,
  mchId: "",
  serialNo: "",
  apiV3Key: "",
  privateKeyPem: "",
  certPem: "",
  notifyHost: "",
  notifyUrl: "",
});
const privateKeyConfigured = ref(false);
const certConfigured = ref(false);
const appSecretConfigured = ref(false);
const lastAutoNotifyUrl = ref("");
const privateKeyFileInput = ref<HTMLInputElement | null>(null);
const privateKeyFileName = ref("");
const certFileInput = ref<HTMLInputElement | null>(null);
const certFileName = ref("");

const keyword = ref("");
const ALL_FILTER = "all";
const typeFilter = ref(ALL_FILTER);
const statusFilter = ref(ALL_FILTER);

const typeItems = computed(() => {
  const types = Array.from(
    new Set(
      providers.value
        .map((provider) => String(provider.type || "").trim())
        .filter((type) => type.length > 0),
    ),
  );
  return [{ label: "全部类型", value: ALL_FILTER }].concat(
    types.map((type) => ({ label: type, value: type })),
  );
});

const statusItems = [
  { label: "全部状态", value: ALL_FILTER },
  { label: "合作中", value: "active" },
  { label: "观察中", value: "monitor" },
  { label: "暂停", value: "paused" },
  { label: "停用", value: "disabled" },
  { label: "停用(旧)", value: "inactive" },
];

const statusOptions = [
  { label: "启用", value: "active" },
  { label: "禁用", value: "disabled" },
];

const createTypeOptions = [
  { label: "微信支付", value: "wechat" },
  { label: "支付宝(暂不可用)", value: "alipay" },
  { label: "银联(暂不可用)", value: "unionpay" },
];

const columns = computed<TableColumn<PaymentProvider>[]>(() => [
  { accessorKey: "name", header: "渠道" },
  { accessorKey: "mchId", header: "商户号" },
  { accessorKey: "feeRate", header: "费率" },
  { accessorKey: "currency", header: "币种" },
  { accessorKey: "settlementCycle", header: "结算周期" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "updatedAt", header: "更新时间" },
  { id: "actions", header: "操作" },
]);

const filteredProviders = computed(() =>
  providers.value.filter((provider) => {
    const matchesKeyword =
      !keyword.value ||
      provider.name.includes(keyword.value);
    const matchesType =
      typeFilter.value === ALL_FILTER || provider.type === typeFilter.value;
    const matchesStatus =
      statusFilter.value === ALL_FILTER || provider.status === statusFilter.value;
    return matchesKeyword && matchesType && matchesStatus;
  }),
);

const statusMeta = (status: string | "") => {
  switch (status) {
    case "active":
      return { label: "合作中", color: "success" as const };
    case "monitor":
      return { label: "观察中", color: "warning" as const };
    case "paused":
      return { label: "暂停", color: "neutral" as const };
    case "disabled":
      return { label: "停用", color: "neutral" as const };
    case "inactive":
      return { label: "停用", color: "neutral" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const summaryCards = computed(() => [
  {
    title: "支付渠道总数",
    value: providers.value.length,
    subtitle: "全部已接入渠道",
  },
  {
    title: "合作中渠道",
    value: providers.value.filter((p) => p.status === "active").length,
    subtitle: "已上线可用",
  },
  {
    title: "平均费率",
    value: formatFeeRate(
      providers.value.length
        ? providers.value.reduce((sum, provider) => sum + provider.feeRate, 0) /
            providers.value.length
        : 0,
    ),
    subtitle: "根据当前配置",
  },
]);

const fmtDT = (s: string) =>
  s ? new Date(s).toLocaleString("zh-CN", { hour12: false }) : "-";

const formatFeeRate = (value: number) => {
  const rate = Number(value) || 0;
  if (rate <= 0) return "0%";
  const percent = rate <= 1 ? rate * 100 : rate;
  return `${percent.toFixed(2)}%`;
};

const isWechat = (provider: PaymentProvider) =>
  String(provider.type || "").toLowerCase() === "wechat";

const normalizeHost = (raw: string) => {
  const host = String(raw || "").trim();
  if (!host) return "";
  if (/^https?:\/\//i.test(host)) return host.replace(/\/+$/, "");
  return `https://${host.replace(/\/+$/, "")}`;
};

const notifyUrlPreview = computed(() => {
  const host = normalizeHost(configForm.notifyHost) || extractNotifyHost(configForm.notifyUrl);
  if (!host) return "";
  const mchId = String(configForm.mchId || "").trim();
  const appId = String(configForm.appId || "").trim();
  if (!mchId || !appId) return "";
  return `${host}/api/v1/mini-app/payments/providers/wechat/${encodeURIComponent(mchId)}/${encodeURIComponent(appId)}/callback`;
});

const applyNotifyUrlPreview = () => {
  if (notifyUrlPreview.value) configForm.notifyUrl = notifyUrlPreview.value;
};

const copyNotifyUrlPreview = async () => {
  if (!notifyUrlPreview.value) return;
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(notifyUrlPreview.value);
      toast.add({ title: "已复制回调地址", color: "green" });
      return;
    }
  } catch {}
  const el = document.createElement("textarea");
  el.value = notifyUrlPreview.value;
  el.style.position = "fixed";
  el.style.left = "-9999px";
  document.body.appendChild(el);
  el.select();
  document.execCommand("copy");
  document.body.removeChild(el);
  toast.add({ title: "已复制回调地址", color: "green" });
};

const extractNotifyHost = (url: string) => {
  const raw = String(url || "").trim();
  if (!raw) return "";
  try {
    const parsed = new URL(raw);
    return parsed.origin;
  } catch {
    return "";
  }
};

const fillConfigForm = (detail: PaymentProviderDetail) => {
  const creds = detail.credentials || {};
  configForm.name = detail.name || "";
  configForm.status = detail.status || "active";
  configForm.appId = String(detail.appId || creds.appId || creds.appid || creds.app_id || "").trim();
  appSecretConfigured.value = Boolean(
    String(creds.appSecret || creds.app_secret || creds.secret || "").trim(),
  );
  configForm.appSecret = "";
  configForm.httpDebug = Boolean(
    creds.httpDebug ?? creds.http_debug ?? false,
  );
  configForm.mchId = String(detail.mchId || creds.mchId || creds.mchid || creds.mch_id || "").trim();
  configForm.serialNo = String(detail.serialNo || creds.serialNo || creds.serial_no || "").trim();
  const apiV3Raw = String(creds.apiV3Key || creds.api_v3_key || "").trim();
  const apiV3Mask = "****************";
  configForm.apiV3Key = apiV3Raw ? (apiV3Raw === "*" ? apiV3Mask : apiV3Raw) : "";
  privateKeyConfigured.value = Boolean(
    String(creds.privateKeyPem || creds.private_key_pem || "").trim(),
  );
  configForm.privateKeyPem = "";
  certConfigured.value = Boolean(
    String(creds.certPem || creds.cert_pem || "").trim(),
  );
  configForm.certPem = "";
  configForm.notifyUrl = String(detail.notifyUrl || creds.notifyUrl || creds.notify_url || "").trim();
  configForm.notifyHost = extractNotifyHost(configForm.notifyUrl);
  lastAutoNotifyUrl.value = notifyUrlPreview.value || configForm.notifyUrl;
};

watch(
  () => notifyUrlPreview.value,
  (next) => {
    if (!next) return;
    if (!configForm.notifyUrl || configForm.notifyUrl === lastAutoNotifyUrl.value) {
      configForm.notifyUrl = next;
    }
    lastAutoNotifyUrl.value = next;
  },
);

watch(
  () => configForm.notifyUrl,
  (next) => {
    const hostFromUrl = extractNotifyHost(next);
    const fallbackHost = normalizeHost(next);
    const host = hostFromUrl || fallbackHost;
    if (host && host !== configForm.notifyHost) {
      configForm.notifyHost = host;
    }
  },
);

const openConfig = async (provider: PaymentProvider) => {
  if (!isWechat(provider)) return;
  selectedProvider.value = provider;
  configOpen.value = true;
  configError.value = "";
  saving.value = false;
  try {
    const detail = await getProviderDetail(provider.id);
    fillConfigForm(detail);
  } catch (error: any) {
    configError.value = error?.message || "加载配置失败";
  }
};

const closeConfig = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
  configOpen.value = false;
};

const openCreate = () => {
  createOpen.value = true;
};

const closeCreate = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
  createOpen.value = false;
};

const saveCreate = async () => {
  createSaving.value = true;
  try {
    toast.add({
      title: "暂未开放",
      description: "新增支付渠道功能暂未开放。",
      color: "yellow",
    });
    closeCreate();
  } finally {
    createSaving.value = false;
  }
};

const saveConfig = async () => {
  if (!selectedProvider.value) return;
  saving.value = true;
  configError.value = "";
  try {
    if (configForm.privateKeyPem.includes("BEGIN CERTIFICATE")) {
      configError.value = "商户私钥内容疑似为证书，请上传 apiclient_key.pem";
      toast.add({
        title: "保存失败",
        description: configError.value,
        color: "red",
      });
      return;
    }
    if (configForm.certPem.includes("BEGIN PRIVATE KEY")) {
      configError.value = "商户证书内容疑似为私钥，请上传 apiclient_cert.pem";
      toast.add({
        title: "保存失败",
        description: configError.value,
        color: "red",
      });
      return;
    }
    const credentials: Record<string, any> = {
      appId: configForm.appId,
      appSecret: configForm.appSecret,
      httpDebug: configForm.httpDebug,
      mchId: configForm.mchId,
      serialNo: configForm.serialNo,
      apiV3Key: configForm.apiV3Key,
      notifyUrl: configForm.notifyUrl,
    };
    if (!configForm.appSecret || configForm.appSecret === "*") {
      delete credentials.appSecret;
    }
    if (!configForm.apiV3Key || configForm.apiV3Key === "****************") {
      delete credentials.apiV3Key;
    }
    if (configForm.privateKeyPem) {
      credentials.privateKeyPem = configForm.privateKeyPem;
    }
    if (configForm.certPem) {
      credentials.certPem = configForm.certPem;
    }
    await updateProvider(selectedProvider.value.id, {
      status: configForm.status,
      credentials,
    });
    await loadProviders();
    toast.add({
      title: "配置已保存",
      description: "支付渠道配置已更新",
      color: "green",
    });
    closeConfig();
  } catch (error: any) {
    configError.value = error?.message || "保存失败";
    toast.add({
      title: "保存失败",
      description: configError.value,
      color: "red",
    });
  } finally {
    saving.value = false;
  }
};

const testMiniAppConfig = async () => {
  if (!selectedProvider.value) return;
  if (!configForm.appId) {
    toast.add({
      title: "校验失败",
      description: "请先填写 AppID",
      color: "red",
    });
    return;
  }
  if (!configForm.appSecret && !appSecretConfigured.value) {
    toast.add({
      title: "校验失败",
      description: "请先填写 AppSecret",
      color: "red",
    });
    return;
  }
  testingMiniApp.value = true;
  try {
    await testMiniApp(selectedProvider.value.id, {
      appId: configForm.appId,
      appSecret: configForm.appSecret,
    });
    toast.add({
      title: "校验成功",
      description: "AppID 与 AppSecret 可用。",
      color: "green",
    });
  } catch (error: any) {
    toast.add({
      title: "校验失败",
      description: error?.message || "AppID 与 AppSecret 校验失败",
      color: "red",
    });
  } finally {
    testingMiniApp.value = false;
  }
};

const testCertSerial = async () => {
  if (!selectedProvider.value) return;
  if (!configForm.serialNo) {
    toast.add({
      title: "校验失败",
      description: "请先填写证书序列号",
      color: "red",
    });
    return;
  }
  if (!configForm.certPem && !certConfigured.value) {
    toast.add({
      title: "校验失败",
      description: "请先上传商户证书",
      color: "red",
    });
    return;
  }
  testingCert.value = true;
  try {
    await testWechatCertSerial(selectedProvider.value.id, {
      serialNo: configForm.serialNo,
      certPem: configForm.certPem,
    });
    toast.add({
      title: "校验成功",
      description: "证书序列号匹配。",
      color: "green",
    });
  } catch (error: any) {
    toast.add({
      title: "校验失败",
      description: error?.message || "证书序列号校验失败",
      color: "red",
    });
  } finally {
    testingCert.value = false;
  }
};

const triggerPrivateKeyFile = () => {
  privateKeyFileInput.value?.click();
};

const onPrivateKeyFileChange = (event: Event) => {
  const input = event.target as HTMLInputElement | null;
  const file = input?.files?.[0];
  if (!file) return;
  privateKeyFileName.value = file.name;
  const reader = new FileReader();
  reader.onload = () => {
    const text = String(reader.result || "").trim();
    if (text) configForm.privateKeyPem = text;
  };
  reader.readAsText(file);
};

const triggerCertFile = () => {
  certFileInput.value?.click();
};

const onCertFileChange = (event: Event) => {
  const input = event.target as HTMLInputElement | null;
  const file = input?.files?.[0];
  if (!file) return;
  certFileName.value = file.name;
  const reader = new FileReader();
  reader.onload = () => {
    const text = String(reader.result || "").trim();
    if (text) configForm.certPem = text;
  };
  reader.readAsText(file);
};

const loadProviders = async () => {
  loading.value = true;
  try {
    providers.value = await listProviders();
  } catch (error: any) {
    toast.add({
      title: "获取支付渠道失败",
      description: error?.message || "请稍后重试",
      color: "red",
    });
  } finally {
    loading.value = false;
  }
};

onMounted(loadProviders);
</script>
