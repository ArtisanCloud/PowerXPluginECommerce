<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        {{ $t("nav.members") }}
      </h1>
      <UButton color="primary" icon="i-heroicons-plus">
        {{ $t("common.add") }}
      </UButton>
    </div>

    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold">会员列表</h3>
          <div class="flex space-x-2">
            <UInput
              v-model="searchQuery"
              :placeholder="$t('common.search')"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="selectedLevel"
              :options="memberLevels"
              :placeholder="$t('common.filter')"
              option-attribute="label"
              value-attribute="value"
              class="w-40"
            />
          </div>
        </div>
      </template>

      <UTable
        :data="filteredMembers"
        :columns="columns"
        :loading="loading"
        class="w-full"
      >
        <!-- v3: 单元格插槽用 -cell -->
        <template #level-cell="{ getValue }">
          <UBadge :color="getLevelColor(getValue())" variant="subtle">
            {{ getValue() }}
          </UBadge>
        </template>

        <template #status-cell="{ getValue }">
          <UBadge
            :color="getValue() === '活跃' ? 'success' : 'neutral'"
            variant="subtle"
          >
            {{ getValue() }}
          </UBadge>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex space-x-2">
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              icon="i-heroicons-eye"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              size="xs"
              color="primary"
              variant="ghost"
              icon="i-heroicons-pencil-square"
            >
              {{ $t("common.edit") }}
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
const { t } = useI18n();

// 响应式数据
const searchQuery = ref("");
const selectedLevel = ref<string | "">("");
const loading = ref(false);

// 会员等级选项
const memberLevels = [
  { label: "全部等级", value: "" },
  { label: "普通会员", value: "普通会员" },
  { label: "银卡会员", value: "银卡会员" },
  { label: "金卡会员", value: "金卡会员" },
  { label: "VIP会员", value: "VIP会员" },
];

// 模拟会员数据
type Member = {
  memberId: string;
  name: string;
  level: "普通会员" | "银卡会员" | "金卡会员" | "VIP会员";
  points: number;
  totalSpent: string;
  joinDate: string;
  status: "活跃" | "不活跃";
};

const members = ref<Member[]>([
  {
    memberId: "M001",
    name: "张三",
    level: "VIP会员",
    points: 2580,
    totalSpent: "¥25,800",
    joinDate: "2023-01-15",
    status: "活跃",
  },
  {
    memberId: "M002",
    name: "李四",
    level: "金卡会员",
    points: 1200,
    totalSpent: "¥12,000",
    joinDate: "2023-03-20",
    status: "活跃",
  },
  {
    memberId: "M003",
    name: "王五",
    level: "银卡会员",
    points: 680,
    totalSpent: "¥6,800",
    joinDate: "2023-06-10",
    status: "活跃",
  },
  {
    memberId: "M004",
    name: "赵六",
    level: "普通会员",
    points: 320,
    totalSpent: "¥3,200",
    joinDate: "2023-08-05",
    status: "不活跃",
  },
]);

// v3：列定义（用 accessorKey / header / cell）
// 用 computed 包起来，i18n 切换时表头会更新
const columns = computed<TableColumn<Member>[]>(() => [
  { accessorKey: "memberId", header: t("members.id") || "会员ID" },
  { accessorKey: "name", header: t("members.name") || "会员姓名" },
  { accessorKey: "level", header: t("members.level") || "会员等级" },
  { accessorKey: "points", header: t("members.points") || "积分" },
  { accessorKey: "totalSpent", header: t("members.totalSpent") || "累计消费" },
  { accessorKey: "joinDate", header: t("members.joinDate") || "加入日期" },
  { accessorKey: "status", header: t("members.status") || "状态" },
  { id: "actions", header: t("common.actions") || "操作" },
]);

// 过滤
const filteredMembers = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  return members.value.filter((m) => {
    const passQ =
      !q ||
      m.name.toLowerCase().includes(q) ||
      m.memberId.toLowerCase().includes(q);
    const passLevel = !selectedLevel.value || m.level === selectedLevel.value;
    return passQ && passLevel;
  });
});

// 等级 -> 语义色映射（v3 推荐用语义色）
const getLevelColor = (level: string) => {
  const map: Record<string, "primary" | "secondary" | "warning" | "neutral"> = {
    VIP会员: "primary",
    金卡会员: "warning",
    银卡会员: "secondary",
    普通会员: "neutral",
  };
  return map[level] || "neutral";
};
</script>
