<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <p class="text-sm text-gray-500 dark:text-gray-400">渠道 ID · {{ channelId }}</p>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ channelProfile.name }}</h1>
        <p class="text-gray-500 dark:text-gray-400">{{ channelProfile.platform }} · {{ channelProfile.region }}</p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-uturn-left" @click="goBack">
          返回列表
        </UButton>
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path" @click="triggerSync">
          手动同步
        </UButton>
        <UButton color="primary" icon="i-heroicons-cog-6-tooth" @click="openConfig">
          配置策略
        </UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <UCard v-for="card in kpiCards" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">{{ card.value }}</div>
        <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-500' : 'text-rose-500'">
          {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% vs last period
        </p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">基础信息</h3>
          <UBadge :color="channelProfile.statusMeta.color" variant="subtle">
            {{ channelProfile.statusMeta.label }}
          </UBadge>
        </div>
      </template>
      <dl class="grid gap-4 md:grid-cols-3">
        <div>
          <dt class="text-sm text-gray-500">负责人</dt>
          <dd class="text-gray-900 dark:text-white">{{ channelProfile.owner }}</dd>
        </div>
        <div>
          <dt class="text-sm text-gray-500">联系人</dt>
          <dd class="text-gray-900 dark:text-white">{{ channelProfile.contact }}</dd>
        </div>
        <div>
          <dt class="text-sm text-gray-500">上次同步</dt>
          <dd class="text-gray-900 dark:text-white">{{ channelProfile.lastSync }}</dd>
        </div>
      </dl>
    </UCard>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold">KPI & 健康度</h3>
              <p class="text-sm text-gray-500">占位图表，后续将接入指标 API</p>
            </div>
            <UBadge color="info" variant="subtle">占位</UBadge>
          </div>
        </template>
        <ul class="space-y-3">
          <li v-for="kpi in kpiHighlights" :key="kpi.label" class="flex items-center justify-between">
            <span class="text-gray-500">{{ kpi.label }}</span>
            <span class="font-semibold text-gray-900 dark:text-white">{{ kpi.value }}</span>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">告警</h3>
            <UBadge color="warning" variant="subtle">{{ alerts.length }} 个</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li v-for="alert in alerts" :key="alert.id" class="border-b border-gray-200 pb-3 last:border-0 last:pb-0 dark:border-gray-800">
            <div class="flex items-center justify-between">
              <div>
                <p class="font-semibold text-gray-900 dark:text-white">{{ alert.title }}</p>
                <p class="text-sm text-gray-500">{{ alert.description }}</p>
              </div>
              <UBadge :color="alert.severity === 'critical' ? 'error' : 'warning'" variant="subtle">
                {{ alert.severity === 'critical' ? '严重' : '提醒' }}
              </UBadge>
            </div>
          </li>
        </ul>
      </UCard>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">任务关联</h3>
            <UButton size="xs" variant="ghost" @click="linkTask">关联任务</UButton>
          </div>
        </template>
        <ul class="space-y-3">
          <li v-for="task in taskLinks" :key="task.id" class="flex items-center justify-between">
            <div>
              <p class="font-semibold">{{ task.title }}</p>
              <p class="text-sm text-gray-500">{{ task.description }}</p>
            </div>
            <UBadge :color="task.status === 'done' ? 'success' : 'info'" variant="subtle">
              {{ task.status === 'done' ? '已完成' : '进行中' }}
            </UBadge>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">运营备注</h3>
            <UButton size="xs" variant="ghost" @click="addNote" :disabled="!newNote">
              保存备注
            </UButton>
          </div>
        </template>
        <UTextarea v-model="newNote" placeholder="记录运营决策、授权背景信息等" class="mb-4"/>
        <ul class="space-y-3">
          <li v-for="note in notes" :key="note.id" class="border-b border-gray-200 pb-3 last:border-0 last:pb-0 dark:border-gray-800">
            <p class="font-semibold text-gray-900 dark:text-white">{{ note.author }}</p>
            <p class="text-sm text-gray-500">{{ note.createdAt }}</p>
            <p class="mt-1 text-gray-800 dark:text-gray-200">{{ note.body }}</p>
          </li>
        </ul>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">同步历史</h3>
          <UBadge color="neutral" variant="subtle">展示最近 {{ syncHistory.length }} 条</UBadge>
        </div>
      </template>
      <UTable :columns="syncColumns" :data="syncHistory"/>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

const route = useRoute();
const router = useRouter();
const toast = useToast();

const channelId = computed(() => String(route.params.id ?? ""));

const channelProfile = reactive({
  name: "渠道占位名称",
  platform: "天猫国际",
  region: "全国",
  owner: "陈曦",
  contact: "张颖 · 188****5620",
  lastSync: "2024-02-12 09:30",
  statusMeta: { label: "未授权", color: "warning" as const },
});

const kpiCards = ref([
  { title: "GMV (30d)", value: "¥1.2M", trend: 6.3 },
  { title: "订单量", value: "8,420", trend: 3.1 },
  { title: "健康度", value: "78", trend: -1.8 },
]);

const kpiHighlights = ref([
  { label: "GMV 环比", value: "+8%" },
  { label: "库存覆盖", value: "95%" },
  { label: "错误率", value: "1.8%" },
  { label: "同步成功率", value: "99.2%" },
]);

const alerts = ref([
  { id: "al-1", title: "凭证将在 7 天内过期", description: "请提前刷新授权凭证", severity: "warning" },
  { id: "al-2", title: "库存覆盖跌破 90%", description: "建议触发备货任务", severity: "critical" },
]);

const taskLinks = ref([
  { id: "task-1", title: "渠道图片整改", description: "等待设计补齐详情页", status: "in_progress" },
  { id: "task-2", title: "客服脚本更新", description: "关联任务中心 #CS-2881", status: "done" },
]);

const notes = ref([
  { id: "note-1", author: "王芳", body: "等待品牌方确认 3 月营销档期。", createdAt: "2024-02-11" },
  { id: "note-2", author: "运营机器人", body: "昨晚自动同步成功。", createdAt: "2024-02-10" },
]);

const newNote = ref("");

const syncHistory = ref([
  { createdAt: "2024-02-12 09:30", triggerType: "manual", triggeredBy: "陈曦", duration: "38s", result: "成功" },
  { createdAt: "2024-02-11 21:00", triggerType: "scheduled", triggeredBy: "任务中心", duration: "42s", result: "成功" },
  { createdAt: "2024-02-10 21:00", triggerType: "scheduled", triggeredBy: "任务中心", duration: "41s", result: "失败" },
]);

const syncColumns = [
  { accessorKey: "createdAt", header: "触发时间" },
  { accessorKey: "triggerType", header: "方式" },
  { accessorKey: "triggeredBy", header: "触发人" },
  { accessorKey: "duration", header: "耗时" },
  { accessorKey: "result", header: "结果" },
] satisfies TableColumn<(typeof syncHistory)[number]>[];

const goBack = () => router.push("/channels");

const triggerSync = () => {
  toast.add({ title: "同步任务已创建", description: "稍后在同步历史中更新结果。" });
};

const openConfig = () => {
  toast.add({ title: "策略配置", description: "策略配置面板将在后续迭代提供。" });
};

const linkTask = () => {
  toast.add({ title: "任务关联", description: "即将接入任务中心 API。" });
};

const addNote = () => {
  if (!newNote.value) return;
  notes.value.unshift({
    id: `note-${Date.now()}`,
    author: "当前用户",
    body: newNote.value,
    createdAt: new Date().toISOString().slice(0, 10),
  });
  newNote.value = "";
  toast.add({ title: "备注已保存" });
};
</script>
