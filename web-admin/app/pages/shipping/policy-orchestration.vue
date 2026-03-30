<template>
  <div class="space-y-6 p-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">策略编排中心</h1>
        <p class="text-gray-500 dark:text-gray-400">管理策略流、冲突预检、灰度发布与版本回滚。</p>
      </div>
      <UButton color="primary" :loading="loading" @click="refreshAll">刷新</UButton>
    </div>

    <UAlert v-if="message" :color="messageType === 'error' ? 'error' : 'success'" variant="soft" :description="message" />

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">流程编辑</h3>
      </template>
      <div class="grid gap-2 md:grid-cols-4">
        <UInput v-model="form.name" placeholder="流程名称" />
        <UInput v-model.number="form.priority" type="number" placeholder="优先级" />
        <USelect v-model="form.status" :options="statusOptions" />
        <UInput v-model="form.conflictRelationsText" placeholder="冲突键，逗号分隔" />
      </div>
      <div class="mt-2">
        <UTextarea v-model="form.flowDefinitionText" :rows="5" placeholder='流程定义 JSON，例如 {"steps":["risk","route","publish"]}' />
      </div>
      <div class="mt-3 flex flex-wrap gap-2">
        <UButton color="info" variant="soft" :loading="saving" @click="saveFlow">保存流程</UButton>
        <UButton color="warning" variant="soft" :loading="previewing" @click="previewConflicts">冲突预检</UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">流程列表</h3>
      </template>
      <UTable :columns="flowColumns" :data="flows" />
      <div class="mt-3 flex flex-wrap gap-2">
        <USelect v-model="selectedFlowId" :options="flowOptions" placeholder="选择流程" />
        <UInput v-model="publishRequestKey" placeholder="发布请求键（可选）" />
        <UButton color="primary" :loading="publishing" @click="publishFlow">发布</UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">版本记录</h3>
      </template>
      <div class="flex flex-wrap gap-2">
        <USelect v-model="rollbackTargetVersionId" :options="versionOptions" placeholder="回滚目标版本" />
        <UInput v-model="rollbackReason" placeholder="回滚原因" />
        <UButton color="error" variant="soft" :loading="rollingBack" @click="rollbackFlow">执行回滚</UButton>
      </div>
      <UTable class="mt-3" :columns="versionColumns" :data="versions" />
    </UCard>

    <UCard v-if="conflictPreview">
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">冲突预检结果</h3>
      </template>
      <UAlert
        :color="conflictPreview.hasConflict ? 'warning' : 'success'"
        variant="soft"
        :description="conflictPreview.hasConflict ? '检测到冲突，请处理后再发布。' : '未发现冲突，可发布。'"
      />
      <UTable class="mt-3" :columns="conflictColumns" :data="conflictPreview.items" />
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import {
  type LogisticsPolicyOrchestrationConflictItem,
  type LogisticsPolicyOrchestrationConflictPreview,
  type LogisticsPolicyOrchestrationFlow,
  type LogisticsPolicyOrchestrationVersion,
  useLogisticsApi,
} from "~/composables/api";

definePageMeta({ name: "shipping-policy-orchestration" });

const api = useLogisticsApi();
const loading = ref(false);
const saving = ref(false);
const previewing = ref(false);
const publishing = ref(false);
const rollingBack = ref(false);
const message = ref("");
const messageType = ref<"success" | "error">("success");
const flows = ref<LogisticsPolicyOrchestrationFlow[]>([]);
const versions = ref<LogisticsPolicyOrchestrationVersion[]>([]);
const conflictPreview = ref<LogisticsPolicyOrchestrationConflictPreview | null>(null);
const selectedFlowId = ref("");
const publishRequestKey = ref("");
const rollbackTargetVersionId = ref("");
const rollbackReason = ref("");

const form = reactive({
  id: "",
  name: "",
  priority: 100,
  status: "draft",
  conflictRelationsText: "",
  flowDefinitionText: '{"steps":["risk","route","publish"]}',
});

const statusOptions = [
  { label: "草稿", value: "draft" },
  { label: "启用", value: "active" },
  { label: "归档", value: "archived" },
];

const flowColumns = computed<TableColumn<LogisticsPolicyOrchestrationFlow>[]>(() => [
  { accessorKey: "name", header: "流程" },
  { accessorKey: "priority", header: "优先级" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "publishedVersionId", header: "当前版本ID" },
  { accessorKey: "updatedAt", header: "更新时间" },
]);

const versionColumns = computed<TableColumn<LogisticsPolicyOrchestrationVersion>[]>(() => [
  { accessorKey: "versionNo", header: "版本号" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "requestKey", header: "请求键" },
  { accessorKey: "changeSummary", header: "变更说明" },
  { accessorKey: "publishedAt", header: "发布时间" },
  { accessorKey: "createdAt", header: "创建时间" },
]);

const conflictColumns = computed<TableColumn<LogisticsPolicyOrchestrationConflictItem>[]>(() => [
  { accessorKey: "flowName", header: "冲突流程" },
  { accessorKey: "conflictCount", header: "冲突数" },
  { accessorKey: "conflictKeys", header: "冲突键" },
  { accessorKey: "reason", header: "说明" },
]);

const flowOptions = computed(() =>
  flows.value.map((item) => ({
    label: `${item.name} (#${item.priority})`,
    value: item.id,
  })),
);

const versionOptions = computed(() =>
  versions.value.map((item) => ({
    label: `v${item.versionNo} | ${item.status} | ${item.id}`,
    value: item.id,
  })),
);

const showError = (err: unknown) => {
  messageType.value = "error";
  message.value = err instanceof Error ? err.message : "请求失败";
};

const parseFlowDefinition = (): Record<string, any> => {
  const raw = form.flowDefinitionText.trim();
  if (!raw) return {};
  return JSON.parse(raw);
};

const parseConflictRelations = (): string[] =>
  form.conflictRelationsText
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

const refreshFlows = async () => {
  flows.value = await api.listPolicyOrchestrationFlows({ limit: 100 });
  if (!selectedFlowId.value && flows.value.length > 0) {
    selectedFlowId.value = flows.value[0].id;
  }
};

const refreshVersions = async () => {
  if (!selectedFlowId.value) {
    versions.value = [];
    return;
  }
  versions.value = await api.listPolicyOrchestrationVersions(selectedFlowId.value, { limit: 100 });
  if (!rollbackTargetVersionId.value && versions.value.length > 0) {
    rollbackTargetVersionId.value = versions.value[0].id;
  }
};

const refreshAll = async () => {
  loading.value = true;
  message.value = "";
  try {
    await refreshFlows();
    await refreshVersions();
  } catch (err) {
    showError(err);
  } finally {
    loading.value = false;
  }
};

const saveFlow = async () => {
  saving.value = true;
  message.value = "";
  try {
    const row = await api.upsertPolicyOrchestrationFlow({
      id: form.id || undefined,
      name: form.name || "默认策略流程",
      priority: Number(form.priority || 100),
      status: form.status || "draft",
      conflict_relations: parseConflictRelations(),
      flow_definition: parseFlowDefinition(),
    });
    form.id = row.id;
    selectedFlowId.value = row.id;
    await refreshFlows();
    await refreshVersions();
    messageType.value = "success";
    message.value = "流程已保存";
  } catch (err) {
    showError(err);
  } finally {
    saving.value = false;
  }
};

const previewConflicts = async () => {
  previewing.value = true;
  message.value = "";
  try {
    conflictPreview.value = await api.previewPolicyOrchestrationConflicts({
      flow_id: form.id || selectedFlowId.value || undefined,
      name: form.name || undefined,
      conflict_relations: parseConflictRelations(),
      flow_definition: parseFlowDefinition(),
    });
  } catch (err) {
    showError(err);
  } finally {
    previewing.value = false;
  }
};

const publishFlow = async () => {
  if (!selectedFlowId.value) {
    messageType.value = "error";
    message.value = "请先选择流程";
    return;
  }
  publishing.value = true;
  message.value = "";
  try {
    const result = await api.publishPolicyOrchestrationFlow(selectedFlowId.value, {
      request_key: publishRequestKey.value || undefined,
      change_summary: "控制台发布",
      force: Boolean(conflictPreview.value?.hasConflict),
    });
    messageType.value = "success";
    message.value = result.idempotencyStatus === "replayed" ? "重复发布已幂等回放" : "发布成功";
    await refreshFlows();
    await refreshVersions();
  } catch (err) {
    showError(err);
  } finally {
    publishing.value = false;
  }
};

const rollbackFlow = async () => {
  if (!selectedFlowId.value || !rollbackTargetVersionId.value) {
    messageType.value = "error";
    message.value = "请选择流程与目标版本";
    return;
  }
  rollingBack.value = true;
  message.value = "";
  try {
    await api.rollbackPolicyOrchestrationFlow(selectedFlowId.value, {
      target_version_id: rollbackTargetVersionId.value,
      reason: rollbackReason.value || undefined,
    });
    messageType.value = "success";
    message.value = "回滚完成";
    await refreshFlows();
    await refreshVersions();
  } catch (err) {
    showError(err);
  } finally {
    rollingBack.value = false;
  }
};

watch(selectedFlowId, async () => {
  await refreshVersions();
});

onMounted(refreshAll);
</script>
