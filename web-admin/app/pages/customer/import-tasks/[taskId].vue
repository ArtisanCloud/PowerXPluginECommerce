<template>
  <div class="space-y-4 p-4 md:p-6">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h1 class="text-lg font-semibold text-gray-900 dark:text-gray-100">客户导入结果</h1>
        <p class="text-xs text-gray-500 dark:text-gray-400">任务 ID：{{ taskId }}</p>
      </div>
      <div class="flex items-center gap-2">
        <UButton variant="subtle" color="neutral" icon="i-heroicons-arrow-left" @click="goBack">
          返回客户列表
        </UButton>
        <UButton color="primary" icon="i-heroicons-arrow-path" :loading="initializing" @click="refreshSnapshot">
          刷新状态
        </UButton>
      </div>
    </div>

    <UAlert
      v-if="streamError"
      color="warning"
      variant="soft"
      title="实时通道异常"
      :description="streamError"
    />

    <UCard>
      <template #header>
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-900 dark:text-gray-100">导入进度</span>
            <UBadge :color="statusBadgeColor" variant="soft">{{ statusLabel }}</UBadge>
          </div>
          <span class="text-sm text-gray-600 dark:text-gray-300">{{ progressValue }}%</span>
        </div>
      </template>

      <div class="space-y-4">
        <UProgress :model-value="progressValue" color="primary" size="md" />

        <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
          <div class="rounded-lg border border-gray-200/60 dark:border-white/10 p-3">
            <p class="text-xs text-gray-500 dark:text-gray-400">已处理</p>
            <p class="text-base font-semibold text-gray-900 dark:text-gray-100">{{ summary.processed }}</p>
          </div>
          <div class="rounded-lg border border-gray-200/60 dark:border-white/10 p-3">
            <p class="text-xs text-gray-500 dark:text-gray-400">成功导入</p>
            <p class="text-base font-semibold text-emerald-600 dark:text-emerald-400">{{ summary.imported }}</p>
          </div>
          <div class="rounded-lg border border-gray-200/60 dark:border-white/10 p-3">
            <p class="text-xs text-gray-500 dark:text-gray-400">冲突记录</p>
            <p class="text-base font-semibold text-orange-600 dark:text-orange-400">{{ summary.conflicts }}</p>
          </div>
        </div>

        <UAlert
          v-if="latestMessage"
          :color="statusColor"
          variant="soft"
          title="处理信息"
          :description="latestMessage"
        />

        <div class="flex flex-wrap items-center gap-2">
          <UButton
            v-if="downloadUrl"
            color="warning"
            variant="soft"
            icon="i-heroicons-arrow-down-tray"
            @click="downloadConflictReport"
          >
            下载冲突明细
          </UButton>
          <UButton color="primary" variant="outline" icon="i-heroicons-arrow-up-tray" @click="goImportAgain">
            再次导入
          </UButton>
        </div>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { navigateTo, useNuxtApp, useRoute } from "#imports";
import { useCustomerStore } from "~/stores/customer";

const TERMINAL_STATUSES = new Set(["success", "failed", "error", "cancelled"]);

const route = useRoute();
const store = useCustomerStore();
const nuxtApp = useNuxtApp();
const loading = (nuxtApp as any)?.$loadingIndicator;

const taskId = computed(() => String(route.params.taskId || "").trim());
const initializing = ref(false);
const streamError = ref("");

const getTaskFromStore = () => store.bulkTasks.find((item) => item.taskId === taskId.value);
const taskFromStore = computed(() => getTaskFromStore());

const status = computed(() => String(taskFromStore.value?.status || "queued").toLowerCase());
const latestMessage = computed(() => String(taskFromStore.value?.message || ""));
const progressValue = computed(() => {
  const raw = Number(taskFromStore.value?.progress ?? 0);
  if (!Number.isFinite(raw)) return 0;
  return Math.max(0, Math.min(100, Math.round(raw)));
});
const downloadUrl = computed(() => String(taskFromStore.value?.downloadUrl || "").trim());

const statusLabel = computed(() => {
  switch (status.value) {
    case "running":
      return "处理中";
    case "success":
      return "成功";
    case "failed":
    case "error":
      return "失败";
    case "cancelled":
      return "已取消";
    default:
      return "排队中";
  }
});

const statusColor = computed(() => {
  if (status.value === "success") return "success";
  if (status.value === "failed" || status.value === "error" || status.value === "cancelled") return "error";
  if (status.value === "running") return "primary";
  return "neutral";
});

const statusBadgeColor = computed(() => {
  if (status.value === "success") return "success";
  if (status.value === "failed" || status.value === "error" || status.value === "cancelled") return "error";
  if (status.value === "running") return "primary";
  return "neutral";
});

const summary = computed(() => {
  const msg = latestMessage.value;
  const imported = Number((msg.match(/成功导入\s*(\d+)\s*条客户/) || [])[1] || 0);
  const conflicts = Number((msg.match(/冲突\s*(\d+)\s*条/) || [])[1] || 0);
  const processed = imported + conflicts;
  return {
    imported,
    conflicts,
    processed,
  };
});

const updateGlobalLoading = () => {
  if (!loading) return;
  const p = progressValue.value;
  if (typeof loading.start === "function") loading.start();
  if (typeof loading.set === "function") loading.set(p, { force: true });
  if (TERMINAL_STATUSES.has(status.value) && typeof loading.finish === "function") {
    loading.finish();
  }
};

const refreshSnapshot = async () => {
  if (!taskId.value) return;
  initializing.value = true;
  try {
    await store.fetchTaskStatus(taskId.value);
  } catch (error: any) {
    streamError.value = error?.message || "任务状态拉取失败";
  } finally {
    initializing.value = false;
  }
};

const connectStream = () => {
  try {
    store.bindTaskStream();
    streamError.value = "";
  } catch (error: any) {
    streamError.value = error?.message || "实时通道连接失败";
  }
};

const toAbsoluteDownloadURL = (raw: string) => {
  if (!raw) return "";
  if (/^https?:\/\//i.test(raw)) return raw;
  if (typeof window === "undefined") return raw;
  return new URL(raw, window.location.origin).toString();
};

const downloadConflictReport = () => {
  const target = toAbsoluteDownloadURL(downloadUrl.value);
  if (!target || typeof window === "undefined") return;
  window.open(target, "_blank", "noopener,noreferrer");
};

const goBack = async () => {
  await navigateTo("/customer");
};

const goImportAgain = async () => {
  await navigateTo("/customer");
};

watch([status, progressValue], () => {
  updateGlobalLoading();
}, { immediate: true });

onMounted(async () => {
  connectStream();
  await refreshSnapshot();
});

onBeforeUnmount(() => {
  if (loading && typeof loading.finish === "function") {
    loading.finish();
  }
});
</script>
