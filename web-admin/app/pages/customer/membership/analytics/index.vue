<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">客户分布分析</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">按会员等级统计客户分布情况</p>
      </div>
      <div class="flex gap-3">
        <UButton color="neutral" variant="outline" icon="i-heroicons-arrow-down-tray" @click="exportData">
          导出数据
        </UButton>
        <UButton color="primary" icon="i-heroicons-arrow-path" :loading="loading" @click="refreshData">
          刷新
        </UButton>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-users" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">总客户数</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ totalCustomers.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-star" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">平均等级</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ averageLevel.toFixed(1) }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon name="i-heroicons-arrow-trending-up" class="w-6 h-6 text-yellow-600 dark:text-yellow-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">最高等级客户</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ highestTierCustomers.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon name="i-heroicons-arrow-trending-down" class="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">最低等级客户</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ lowestTierCustomers.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
      <UCard>
        <template #header>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">客户等级分布（柱状图）</h2>
        </template>
        <ClientOnly>
          <VChart v-if="barChartData" :option="barChartData" :style="{ height: '320px' }" autoresize />
        </ClientOnly>
      </UCard>

      <UCard>
        <template #header>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">客户等级分布（饼图）</h2>
        </template>
        <ClientOnly>
          <VChart v-if="pieChartData" :option="pieChartData" :style="{ height: '320px' }" autoresize />
        </ClientOnly>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">等级分布详情</h2>
          <div class="flex gap-2">
            <UInput v-model="searchQuery" placeholder="搜索等级名称..." icon="i-heroicons-magnifying-glass" size="sm" />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredTiers" :loading="loading">
        <template #name-cell="{ row }">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-medium" :style="{ backgroundColor: row.color }">
              {{ (row.name || '').charAt(0) }}
            </div>
            <div>
              <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
              <div class="text-sm text-gray-500 dark:text-gray-400">{{ row.description }}</div>
            </div>
          </div>
        </template>

        <template #level-cell="{ row }">
          <UBadge variant="soft" size="sm">{{ row.level }}</UBadge>
        </template>

        <template #customerCount-cell="{ row }">
          <div class="text-center">
            <div class="font-medium text-gray-900 dark:text-white">{{ row.customerCount.toLocaleString() }}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400">占比 {{ row.percentage }}%</div>
          </div>
        </template>

        <template #growth-cell="{ row }">
          <div class="text-sm text-gray-500 dark:text-gray-400">{{ row.growth }}%</div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import VChart from "vue-echarts";
import { useCustomerApi } from "~/composables/api/useCustomer";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import type { MembershipInsight } from "~/types/customer";
import type { MembershipTier } from "~/types/membership";
import { useToast } from "#imports";

type TierDistribution = {
  id: string;
  name: string;
  description: string;
  color: string;
  level: number;
  customerCount: number;
  percentage: number;
  growth: number;
};

const customerApi = useCustomerApi();
const membershipApi = useMembershipAdminApi();
const toast = useToast();

const loading = ref(false);
const searchQuery = ref("");
const tiers = ref<TierDistribution[]>([]);

const totalCustomers = computed(() => tiers.value.reduce((sum, tier) => sum + tier.customerCount, 0));
const averageLevel = computed(() => {
  const total = totalCustomers.value;
  if (!total) return 0;
  const weighted = tiers.value.reduce((sum, tier) => sum + tier.level * tier.customerCount, 0);
  return weighted / total;
});
const highestTierCustomers = computed(() => {
  if (!tiers.value.length) return 0;
  const sorted = [...tiers.value].sort((a, b) => b.level - a.level);
  return sorted[0]?.customerCount || 0;
});
const lowestTierCustomers = computed(() => {
  if (!tiers.value.length) return 0;
  const sorted = [...tiers.value].sort((a, b) => a.level - b.level);
  return sorted[0]?.customerCount || 0;
});

const filteredTiers = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  if (!keyword) return tiers.value;
  return tiers.value.filter((tier) => tier.name.toLowerCase().includes(keyword));
});

const columns = [
  { accessorKey: "name", header: "等级名称" },
  { accessorKey: "level", header: "等级" },
  { accessorKey: "customerCount", header: "客户数" },
  { accessorKey: "growth", header: "增长率" },
];

const barChartData = computed(() => ({
  tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
  grid: { left: "3%", right: "4%", bottom: "3%", containLabel: true },
  xAxis: [{ type: "category", data: tiers.value.map((tier) => tier.name), axisTick: { alignWithLabel: true } }],
  yAxis: [{ type: "value" }],
  series: [
    {
      name: "客户数",
      type: "bar",
      barWidth: "60%",
      data: tiers.value.map((tier) => tier.customerCount),
      itemStyle: { color: (params: any) => tiers.value[params.dataIndex]?.color || "#93C5FD" },
    },
  ],
}));

const pieChartData = computed(() => ({
  tooltip: { trigger: "item" },
  legend: { orient: "vertical", left: "left" },
  series: [
    {
      name: "客户分布",
      type: "pie",
      radius: ["40%", "70%"],
      avoidLabelOverlap: false,
      label: { show: false, position: "center" },
      emphasis: {
        label: {
          show: true,
          fontSize: 20,
          fontWeight: "bold",
          formatter: (params: any) => `${params.name}\n${params.percent.toFixed(1)}%`,
        },
      },
      labelLine: { show: false },
      data: tiers.value.map((tier) => ({ value: tier.customerCount, name: tier.name, itemStyle: { color: tier.color } })),
    },
  ],
}));

const levelColors = ["#CD7F32", "#C0C0C0", "#FFD700", "#E5E4E2", "#B9F2FF", "#111827"];
const colorOf = (name: string) => {
  if (!name) return "#9CA3AF";
  let hash = 0;
  for (let i = 0; i < name.length; i += 1) hash = (hash << 5) - hash + name.charCodeAt(i);
  return levelColors[Math.abs(hash) % levelColors.length];
};

const deriveLevel = (tier: MembershipTier, index: number): number => {
  const rules = tier.rules && typeof tier.rules === "object" ? (tier.rules as Record<string, any>) : {};
  const explicit = Number(rules.level || rules.rank || rules.tierLevel || 0);
  if (Number.isFinite(explicit) && explicit > 0) return explicit;
  return index + 1;
};

const loadData = async () => {
  try {
    loading.value = true;
    const [membersResp, tiersResp] = await Promise.all([
      customerApi.listMembers({ page: 1, pageSize: 1000 }),
      membershipApi.listTiers(),
    ]);

    const members = (membersResp?.data || []) as MembershipInsight[];
    const tierItems = (tiersResp?.items || []) as MembershipTier[];

    const tierOrder = [...tierItems]
      .map((tier, index) => ({ tier, level: deriveLevel(tier, index) }))
      .sort((a, b) => a.level - b.level);

    const counts = new Map<string, number>();
    members.forEach((item) => {
      const tierName = String(item?.snapshot?.tier || item?.customer?.membershipTierLabel || item?.customer?.membershipTier || "未分层");
      counts.set(tierName, (counts.get(tierName) || 0) + 1);
    });

    const total = Math.max(1, members.length);
    const rows: TierDistribution[] = tierOrder.map(({ tier, level }) => {
      const count = counts.get(tier.name) || 0;
      counts.delete(tier.name);
      return {
        id: tier.id,
        name: tier.name,
        description: `状态：${tier.status || "draft"}`,
        color: colorOf(tier.name),
        level,
        customerCount: count,
        percentage: Number(((count / total) * 100).toFixed(1)),
        growth: 0,
      };
    });

    counts.forEach((count, name) => {
      rows.push({
        id: `extra-${name}`,
        name,
        description: "未匹配到等级配置",
        color: colorOf(name),
        level: rows.length + 1,
        customerCount: count,
        percentage: Number(((count / total) * 100).toFixed(1)),
        growth: 0,
      });
    });

    tiers.value = rows;
  } catch (error: any) {
    toast.add({ title: "加载分析数据失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    loading.value = false;
  }
};

const exportData = () => {
  const lines = ["等级,等级序号,客户数,占比"]; 
  tiers.value.forEach((tier) => {
    lines.push(`${tier.name},${tier.level},${tier.customerCount},${tier.percentage}%`);
  });
  const blob = new Blob(["\uFEFF" + lines.join("\n")], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "membership-tier-distribution.csv";
  a.click();
  URL.revokeObjectURL(url);
};

const refreshData = async () => {
  await loadData();
};

onMounted(() => {
  loadData();
});
</script>
