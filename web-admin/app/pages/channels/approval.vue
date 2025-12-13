<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">渠道审批台</h1>
        <p class="text-gray-500 dark:text-gray-400">审核渠道入驻与变更，保障资料一致性。</p>
      </div>
      <div class="flex gap-2">
        <UButton to="/channels" variant="ghost" color="neutral" icon="i-heroicons-arrow-left">
          返回列表
        </UButton>
        <USelect v-model="platformFilter" :options="platformOptions" class="w-48" placeholder="全部平台"/>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold">待处理申请</h3>
            <p class="text-sm text-gray-500">依优先级排序的占位数据</p>
          </div>
          <UBadge color="info" variant="subtle">{{ filteredApprovals.length }} 条</UBadge>
        </div>
      </template>
      <div class="space-y-4">
        <UCard
          v-for="request in filteredApprovals"
          :key="request.id"
          class="bg-slate-50/80 dark:bg-slate-900/40"
        >
          <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div>
              <p class="text-sm text-gray-500">{{ request.platform }} · {{ request.region }}</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">{{ request.name }}</p>
              <p class="text-sm text-gray-500">提交人：{{ request.owner }} · {{ request.submittedAt }}</p>
              <p class="mt-1 text-sm text-gray-500">更新内容：{{ request.summary }}</p>
            </div>
            <div class="flex gap-2">
              <UButton color="neutral" variant="ghost" @click="viewDetail(request.id)">
                查看详情
              </UButton>
              <UButton color="success" icon="i-heroicons-check" @click="decide(request.id, 'approve')">
                通过
              </UButton>
              <UButton color="error" variant="outline" icon="i-heroicons-x-mark" @click="decide(request.id, 'reject')">
                驳回
              </UButton>
            </div>
          </div>
        </UCard>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">审批日志</h3>
          <UBadge color="neutral" variant="subtle">占位</UBadge>
        </div>
      </template>
      <p class="text-sm text-gray-500">
        审批结果与评论将写入 channel_audit_logs，当前页面展示静态样例。
      </p>
    </UCard>
  </div>
</template>

<script setup lang="ts">
type ApprovalStatus = "pending" | "approved" | "rejected";

type ApprovalRequest = {
  id: string;
  name: string;
  platform: string;
  region: string;
  owner: string;
  summary: string;
  status: ApprovalStatus;
  submittedAt: string;
};

const router = useRouter();
const toast = useToast();

const approvals = ref<ApprovalRequest[]>([
  {
    id: "TMALL-NEW",
    name: "天猫国际-潮玩旗舰店",
    platform: "天猫国际",
    region: "全国",
    owner: "李倩",
    summary: "新渠道入驻，待审批 owner/approver 配置",
    status: "pending",
    submittedAt: "今天 09:20",
  },
  {
    id: "JD-REFRESH",
    name: "京东官方旗舰",
    platform: "京东",
    region: "全国",
    owner: "王帆",
    summary: "更新客服负责人并补充客服 SLA",
    status: "pending",
    submittedAt: "昨天 19:05",
  },
]);

const platformFilter = ref("");
const platformOptions = computed(() => [
  { label: "全部平台", value: "" },
  ...Array.from(new Set(approvals.value.map((item) => item.platform))).map((platform) => ({
    label: platform,
    value: platform,
  })),
]);

const filteredApprovals = computed(() =>
  approvals.value.filter((request) => !platformFilter.value || request.platform === platformFilter.value),
);

const decide = (id: string, decision: "approve" | "reject") => {
  approvals.value = approvals.value.filter((item) => item.id !== id);
  toast.add({
    title: decision === "approve" ? "审批已通过" : "审批已驳回",
    description: `渠道 ${id} ${decision === 'approve' ? '进入授权环节' : '已退回修改'}`,
  });
};

const viewDetail = (id: string) => {
  router.push(`/channels/${id}`);
};
</script>
