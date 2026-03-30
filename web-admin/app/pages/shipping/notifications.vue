<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">物流通知中心</h1>
        <p class="text-gray-500 dark:text-gray-400">管理通知模板并查看发送历史，支持失败重试与幂等发送。</p>
      </div>
      <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path" @click="loadAll">
        刷新
      </UButton>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">模板管理</h3>
      </template>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="tplForm.name" class="w-40" placeholder="模板名" />
        <USelect v-model="tplForm.event" class="w-40" :options="eventOptions" />
        <USelect v-model="tplForm.channel" class="w-28" :options="channelOptions" />
        <UInput v-model="tplForm.title" class="w-56" placeholder="标题（可选）" />
        <UInput v-model="tplForm.body" class="w-[420px]" placeholder="正文，如：运单 {{waybill_no}} 已签收" />
        <UButton color="primary" @click="saveTemplate">保存模板</UButton>
      </div>
      <UTable class="mt-4" :columns="tplColumns" :data="templateRows">
        <template #enabled-cell="{ getValue }">
          <UBadge :color="getValue() ? 'success' : 'neutral'" variant="subtle">
            {{ getValue() ? "启用" : "停用" }}
          </UBadge>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">发送通知</h3>
      </template>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="sendForm.waybillId" class="w-64" placeholder="运单ID" />
        <USelect v-model="sendForm.event" class="w-40" :options="eventOptions" />
        <UInput v-model="sendForm.idempotencyKey" class="w-56" placeholder="幂等键（可选）" />
        <UCheckbox v-model="sendForm.simulateFail" label="模拟失败" />
        <UButton color="primary" icon="i-heroicons-paper-airplane" @click="sendNow">发送</UButton>
      </div>
      <p v-if="sendResult" class="mt-2 text-sm text-gray-600 dark:text-gray-300">
        发送结果：{{ sendResult.idempotencyStatus }}，状态：{{ sendResult.record.status }}，尝试次数：{{ sendResult.record.attemptCount }}
      </p>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">发送历史</h3>
      </template>
      <UTable :columns="recordColumns" :data="recordRows">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <UButton
            size="xs"
            variant="ghost"
            color="warning"
            :disabled="row.original.status !== 'failed'"
            @click="retry(row.original.id)"
          >
            重试
          </UButton>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useLogisticsApi } from "~/composables/api";

definePageMeta({
  name: "shipping-notifications",
});

const logisticsApi = useLogisticsApi();
const templateRows = ref<any[]>([]);
const recordRows = ref<any[]>([]);
const sendResult = ref<any>(null);

const tplForm = reactive({
  name: "",
  event: "shipped",
  channel: "sms",
  title: "",
  body: "",
});

const sendForm = reactive({
  waybillId: "",
  event: "shipped",
  idempotencyKey: "",
  simulateFail: false,
});

const eventOptions = [
  { label: "发货", value: "shipped" },
  { label: "派送中", value: "out_for_delivery" },
  { label: "已签收", value: "signed" },
  { label: "异常", value: "exception" },
];

const channelOptions = [
  { label: "短信", value: "sms" },
  { label: "邮件", value: "email" },
  { label: "Webhook", value: "webhook" },
];

const tplColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "name", header: "模板名" },
  { accessorKey: "event", header: "事件" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "enabled", header: "状态" },
  { accessorKey: "body", header: "正文" },
]);

const recordColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "event", header: "事件" },
  { accessorKey: "waybillId", header: "运单ID" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "attemptCount", header: "尝试" },
  { accessorKey: "idempotencyKey", header: "幂等键" },
  { accessorKey: "lastError", header: "错误" },
  { id: "actions", header: "操作" },
]);

const statusMeta = (status: string) => {
  switch (status) {
    case "sent":
      return { label: "已发送", color: "success" as const };
    case "failed":
      return { label: "失败", color: "error" as const };
    default:
      return { label: "待处理", color: "neutral" as const };
  }
};

const loadTemplates = async () => {
  templateRows.value = await logisticsApi.listNotificationTemplates();
};

const loadRecords = async () => {
  recordRows.value = await logisticsApi.listNotificationRecords();
};

const loadAll = async () => {
  await loadTemplates();
  await loadRecords();
};

const saveTemplate = async () => {
  if (!tplForm.name || !tplForm.body) return;
  await logisticsApi.upsertNotificationTemplate({
    name: tplForm.name,
    event: tplForm.event,
    channel: tplForm.channel,
    title: tplForm.title || undefined,
    body: tplForm.body,
    enabled: true,
  });
  tplForm.name = "";
  tplForm.title = "";
  tplForm.body = "";
  await loadTemplates();
};

const sendNow = async () => {
  if (!sendForm.waybillId) return;
  sendResult.value = await logisticsApi.sendNotification({
    waybill_id: sendForm.waybillId,
    event: sendForm.event,
    idempotency_key: sendForm.idempotencyKey || undefined,
    payload: {
      simulate_fail: Boolean(sendForm.simulateFail),
    },
  });
  await loadRecords();
};

const retry = async (id: string) => {
  await logisticsApi.retryNotification(id);
  await loadRecords();
};

onMounted(() => {
  loadAll();
});
</script>
