<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">物流风控中心</h1>
        <p class="text-gray-500 dark:text-gray-400">维护命中规则与黑名单，处理误拦截放行，降低异常派送风险。</p>
      </div>
      <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path" @click="loadAll">刷新</UButton>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">风控规则</h3>
      </template>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="ruleForm.name" class="w-40" placeholder="规则名" />
        <USelect v-model="ruleForm.matchField" class="w-36" :options="matchFieldOptions" />
        <USelect v-model="ruleForm.matchMode" class="w-36" :options="matchModeOptions" />
        <UInput v-model="ruleForm.pattern" class="w-56" placeholder="匹配内容" />
        <USelect v-model="ruleForm.decision" class="w-32" :options="decisionOptions" />
        <UInput v-model.number="ruleForm.priority" class="w-28" type="number" placeholder="优先级" />
        <UButton color="primary" @click="saveRule">保存规则</UButton>
      </div>
      <UTable class="mt-4" :columns="ruleColumns" :data="ruleRows">
        <template #enabled-cell="{ getValue }">
          <UBadge :color="getValue() ? 'success' : 'neutral'" variant="subtle">{{ getValue() ? "启用" : "停用" }}</UBadge>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">黑名单</h3>
      </template>
      <div class="flex flex-wrap gap-2">
        <USelect v-model="blackForm.entryType" class="w-32" :options="entryTypeOptions" />
        <UInput v-model="blackForm.recipientName" class="w-36" placeholder="收件人（可选）" />
        <UInput v-model="blackForm.recipientPhone" class="w-40" placeholder="手机号（可选）" />
        <UInput v-model="blackForm.addressLine" class="w-72" placeholder="地址（可选）" />
        <UInput v-model="blackForm.reason" class="w-48" placeholder="原因" />
        <UButton color="warning" @click="saveBlacklist">加入黑名单</UButton>
      </div>
      <UTable class="mt-4" :columns="blacklistColumns" :data="blacklistRows">
        <template #status-cell="{ getValue }">
          <UBadge :color="getValue() === 'active' ? 'error' : 'neutral'" variant="subtle">
            {{ getValue() === "active" ? "生效中" : "失效" }}
          </UBadge>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">命中记录与放行</h3>
      </template>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="evalForm.waybillNo" class="w-36" placeholder="运单号（可选）" />
        <UInput v-model="evalForm.recipientName" class="w-36" placeholder="收件人" />
        <UInput v-model="evalForm.recipientPhone" class="w-40" placeholder="手机号" />
        <UInput v-model="evalForm.destinationLine" class="w-[460px]" placeholder="详细地址" />
        <UButton color="primary" icon="i-heroicons-shield-exclamation" @click="runEvaluate">即时评估</UButton>
      </div>
      <p v-if="evaluateResult" class="mt-3 text-sm text-gray-600 dark:text-gray-300">
        评估结果：{{ evaluateResult.decision }}，风险分：{{ evaluateResult.score }}，拦截：{{ evaluateResult.blocked ? "是" : "否" }}
      </p>
      <UTable class="mt-4" :columns="hitColumns" :data="hitRows">
        <template #status-cell="{ getValue }">
          <UBadge :color="getValue() === 'released' ? 'success' : 'warning'" variant="subtle">
            {{ getValue() === "released" ? "已放行" : "待处理" }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <UButton
            size="xs"
            variant="ghost"
            color="success"
            :disabled="row.original.status === 'released'"
            @click="release(row.original.id)"
          >
            人工放行
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
  name: "shipping-risk-control",
});

const logisticsApi = useLogisticsApi();
const ruleRows = ref<any[]>([]);
const blacklistRows = ref<any[]>([]);
const hitRows = ref<any[]>([]);
const evaluateResult = ref<any>(null);

const ruleForm = reactive({
  name: "",
  matchField: "address",
  matchMode: "contains",
  pattern: "",
  decision: "review",
  priority: 100,
});

const blackForm = reactive({
  entryType: "recipient",
  recipientName: "",
  recipientPhone: "",
  addressLine: "",
  reason: "",
});

const evalForm = reactive({
  waybillNo: "",
  recipientName: "",
  recipientPhone: "",
  destinationLine: "",
});

const matchFieldOptions = [
  { label: "地址", value: "address" },
  { label: "收件人", value: "recipient" },
  { label: "手机号", value: "phone" },
];

const matchModeOptions = [
  { label: "包含", value: "contains" },
  { label: "精确", value: "exact" },
  { label: "前缀", value: "prefix" },
  { label: "正则", value: "regex" },
];

const decisionOptions = [
  { label: "审核", value: "review" },
  { label: "拦截", value: "block" },
];

const entryTypeOptions = [
  { label: "收件人", value: "recipient" },
  { label: "地址", value: "address" },
  { label: "手机号", value: "phone" },
];

const ruleColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "name", header: "规则" },
  { accessorKey: "matchField", header: "字段" },
  { accessorKey: "matchMode", header: "模式" },
  { accessorKey: "pattern", header: "匹配值" },
  { accessorKey: "decision", header: "动作" },
  { accessorKey: "priority", header: "优先级" },
  { accessorKey: "enabled", header: "状态" },
]);

const blacklistColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "entryType", header: "类型" },
  { accessorKey: "recipientName", header: "收件人" },
  { accessorKey: "recipientPhone", header: "手机号" },
  { accessorKey: "addressLine", header: "地址" },
  { accessorKey: "reason", header: "原因" },
  { accessorKey: "status", header: "状态" },
]);

const hitColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "source", header: "来源" },
  { accessorKey: "decision", header: "判定" },
  { accessorKey: "riskLevel", header: "风险级别" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "description", header: "说明" },
  { id: "actions", header: "操作" },
]);

const loadRules = async () => {
  ruleRows.value = await logisticsApi.listRiskRules();
};

const loadBlacklist = async () => {
  blacklistRows.value = await logisticsApi.listRiskBlacklist();
};

const loadHits = async () => {
  hitRows.value = await logisticsApi.listRiskHits({ status: "open" });
};

const loadAll = async () => {
  await Promise.all([loadRules(), loadBlacklist(), loadHits()]);
};

const saveRule = async () => {
  if (!ruleForm.name || !ruleForm.pattern) return;
  await logisticsApi.upsertRiskRule({
    name: ruleForm.name,
    match_field: ruleForm.matchField,
    match_mode: ruleForm.matchMode,
    pattern: ruleForm.pattern,
    decision: ruleForm.decision,
    priority: Number(ruleForm.priority || 100),
    enabled: true,
  });
  ruleForm.name = "";
  ruleForm.pattern = "";
  await loadRules();
};

const saveBlacklist = async () => {
  if (!blackForm.recipientName && !blackForm.recipientPhone && !blackForm.addressLine) return;
  await logisticsApi.upsertRiskBlacklist({
    entry_type: blackForm.entryType,
    recipient_name: blackForm.recipientName || undefined,
    recipient_phone: blackForm.recipientPhone || undefined,
    address_line: blackForm.addressLine || undefined,
    reason: blackForm.reason || undefined,
    status: "active",
  });
  blackForm.recipientName = "";
  blackForm.recipientPhone = "";
  blackForm.addressLine = "";
  blackForm.reason = "";
  await loadBlacklist();
};

const runEvaluate = async () => {
  if (!evalForm.recipientName && !evalForm.recipientPhone && !evalForm.destinationLine) return;
  evaluateResult.value = await logisticsApi.evaluateRisk({
    waybill_no: evalForm.waybillNo || undefined,
    recipient_name: evalForm.recipientName || undefined,
    recipient_phone: evalForm.recipientPhone || undefined,
    destination_line: evalForm.destinationLine || undefined,
  });
  await loadHits();
};

const release = async (id: string) => {
  await logisticsApi.releaseRiskHit(id, {
    operator_id: "admin",
    reason: "人工复核通过",
  });
  await loadHits();
};

onMounted(() => {
  loadAll();
});
</script>
