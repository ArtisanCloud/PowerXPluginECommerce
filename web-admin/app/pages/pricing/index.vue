<template>
  <div class="p-6 space-y-6">
    <!-- 页面标题与操作 -->
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
        {{ $t("pricing.overview") }}
      </h1>
      <div class="flex gap-3">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-heroicons-arrow-path"
          @click="onRefresh"
        >
          {{ $t("common.refresh") }}
        </UButton>
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="calculatePrice"
        >
          {{ $t("pricing.priceSimulation") }}
        </UButton>
      </div>
    </div>

    <!-- 定价概览卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <UCard>
        <div class="flex items-center">
          <UIcon name="i-heroicons-book-open" class="w-8 h-8 text-blue-500" />
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">
              {{ $t("pricing.pricebooks") }}
            </p>
            <p class="text-2xl font-bold text-gray-900">15</p>
            <p class="text-xs text-green-600">+2 本月新增</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <UIcon
            name="i-heroicons-cog-6-tooth"
            class="w-8 h-8 text-green-500"
          />
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">
              {{ $t("pricing.rules") }}
            </p>
            <p class="text-2xl font-bold text-gray-900">48</p>
            <p class="text-xs text-blue-600">12 条活跃</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <UIcon
            name="i-heroicons-document-text"
            class="w-8 h-8 text-purple-500"
          />
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">
              {{ $t("pricing.contracts") }}
            </p>
            <p class="text-2xl font-bold text-gray-900">126</p>
            <p class="text-xs text-orange-600">8 即将到期</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <UIcon name="i-heroicons-chart-bar" class="w-8 h-8 text-yellow-500" />
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">
              {{ $t("pricing.marginAnalysis") }}
            </p>
            <p class="text-2xl font-bold text-gray-900">32.5%</p>
            <p class="text-xs text-green-600">+1.2% 环比</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 价格计算 & 溯源 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 价格计算引擎 -->
      <UCard>
        <template #header>
          <div class="flex justify-between items-center">
            <h3 class="text-lg font-semibold">
              {{ $t("pricing.priceCalculation") }}
            </h3>
            <UButton
              size="sm"
              color="primary"
              variant="outline"
              @click="calculatePrice"
            >
              {{ $t("pricing.priceSimulation") }}
            </UButton>
          </div>
        </template>

        <div class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1"
                >商品</label
              >
              <USelect
                v-model="selectedProduct"
                :options="productOptions"
                option-attribute="label"
                value-attribute="value"
                placeholder="选择商品"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1"
                >渠道</label
              >
              <USelect
                v-model="selectedChannel"
                :options="channelOptions"
                option-attribute="label"
                value-attribute="value"
                placeholder="选择渠道"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1"
                >客户层级</label
              >
              <USelect
                v-model="selectedTier"
                :options="tierOptions"
                option-attribute="label"
                value-attribute="value"
                placeholder="选择层级"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1"
                >数量</label
              >
              <UInput
                v-model.number="quantity"
                type="number"
                min="1"
                placeholder="输入数量"
              />
            </div>
          </div>

          <div
            v-if="calculatedPrice"
            class="bg-gray-50 dark:bg-gray-800/40 p-4 rounded-lg"
          >
            <div class="space-y-2">
              <div class="flex justify-between">
                <span class="text-sm text-gray-600">基准价格:</span>
                <span class="font-medium">{{
                  formatCNY(calculatedPrice.basePrice)
                }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-sm text-gray-600">折扣:</span>
                <span class="font-medium text-red-600"
                  >-{{ formatCNY(calculatedPrice.discount) }}</span
                >
              </div>
              <div class="flex justify-between border-t pt-2">
                <span class="font-semibold">最终价格:</span>
                <span class="font-bold text-lg text-green-600">{{
                  formatCNY(calculatedPrice.finalPrice)
                }}</span>
              </div>
            </div>
          </div>

          <UButton
            :loading="calculating"
            color="primary"
            block
            @click="calculatePrice"
          >
            计算价格
          </UButton>
        </div>
      </UCard>

      <!-- 价格溯源 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">{{ $t("pricing.priceTrace") }}</h3>
        </template>

        <div v-if="priceTrace.length" class="space-y-3">
          <div
            v-for="(step, idx) in priceTrace"
            :key="idx"
            class="flex items-center space-x-3 p-3 bg-gray-50 dark:bg-gray-800/40 rounded-lg"
          >
            <div
              class="w-6 h-6 bg-blue-500 text-white rounded-full flex items-center justify-center text-xs font-bold"
            >
              {{ idx + 1 }}
            </div>
            <div class="flex-1">
              <p class="text-sm font-medium text-gray-900 dark:text-gray-100">
                {{ step.rule }}
              </p>
              <p class="text-xs text-gray-500">{{ step.description }}</p>
            </div>
            <div class="text-right">
              <p
                class="text-sm font-bold"
                :class="step.change > 0 ? 'text-red-600' : 'text-green-600'"
              >
                {{ step.change > 0 ? "+" : ""
                }}{{ formatCNY(Math.abs(step.change)) }}
              </p>
              <p class="text-xs text-gray-500">{{ formatCNY(step.price) }}</p>
            </div>
          </div>
        </div>
        <div v-else class="text-center py-8 text-gray-500">
          <UIcon
            name="i-heroicons-magnifying-glass"
            class="w-12 h-12 mx-auto mb-3 text-gray-300"
          />
          <p>请先计算价格以查看溯源信息</p>
        </div>
      </UCard>
    </div>

    <!-- 最近活动：价格历史 & 待处理任务 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 价格变更历史 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">
            {{ $t("pricing.priceHistory") }}
          </h3>
        </template>

        <UTable :data="priceHistory" :columns="historyColumns">
          <!-- 列插槽名称与列id一致，使用 -cell 后缀 -->
          <template #changeType-cell="{ row }">
            <UBadge
              :variant="'subtle'"
              :color="row.original.changeType === '上调' ? 'error' : 'success'"
            >
              {{ row.original.changeType }}
            </UBadge>
          </template>

          <template #amount-cell="{ row }">
            <div
              class="text-right"
              :class="
                row.original.changeType === '上调'
                  ? 'text-red-600'
                  : 'text-green-600'
              "
            >
              {{ row.original.changeType === "上调" ? "+" : "-"
              }}{{ formatCNY(row.original.amount) }}
            </div>
          </template>
        </UTable>
      </UCard>

      <!-- 待处理任务 -->
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">待处理任务</h3>
        </template>

        <div class="space-y-3">
          <div
            v-for="task in pendingTasks"
            :key="task.id"
            class="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
          >
            <div class="flex items-center space-x-3">
              <UIcon
                :name="task.icon"
                class="w-5 h-5"
                :class="task.iconColor"
              />
              <div>
                <p class="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {{ task.title }}
                </p>
                <p class="text-xs text-gray-500">{{ task.description }}</p>
              </div>
            </div>
            <div class="flex items-center space-x-2">
              <UBadge
                :variant="'subtle'"
                :color="
                  task.priority === '高'
                    ? 'error'
                    : task.priority === '中'
                      ? 'warning'
                      : 'neutral'
                "
              >
                {{ task.priority }}
              </UBadge>
              <UButton size="xs" color="primary" variant="outline"
                >处理</UButton
              >
            </div>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { h } from "vue";
import type { TableColumn } from "@nuxt/ui";

const { t } = useI18n();
const toast = useToast();

/** ======================
 *  类型
 *  ====================== */
type Tier = "regular" | "vip" | "enterprise" | "distributor";
type TraceStep = {
  rule: string;
  description: string;
  change: number;
  price: number;
};
type CalcResult = { basePrice: number; discount: number; finalPrice: number };
type PriceHistoryRow = {
  product: string;
  channel: string;
  changeType: "上调" | "下调";
  amount: number;
  date: string;
};

/** ======================
 *  状态
 *  ====================== */
const selectedProduct = ref<string>("");
const selectedChannel = ref<string>("");
const selectedTier = ref<Tier | "">("");
const quantity = ref<number>(1);
const calculating = ref<boolean>(false);
const calculatedPrice = ref<CalcResult | null>(null);
const priceTrace = ref<TraceStep[]>([]);

/** ======================
 *  选项
 *  ====================== */
const productOptions = [
  { label: "iPhone 15 Pro", value: "iphone15pro" },
  { label: "MacBook Air M2", value: "macbookair" },
  { label: "AirPods Pro", value: "airpodspro" },
];
const channelOptions = [
  { label: "天猫旗舰店", value: "tmall" },
  { label: "京东自营", value: "jd" },
  { label: "线下门店", value: "offline" },
  { label: "企业直销", value: "b2b" },
];
const tierOptions = [
  { label: "普通客户", value: "regular" },
  { label: "VIP客户", value: "vip" },
  { label: "企业客户", value: "enterprise" },
  { label: "分销商", value: "distributor" },
];

/** ======================
 *  价格历史
 *  ====================== */
const priceHistory = ref<PriceHistoryRow[]>([
  {
    product: "iPhone 15 Pro",
    channel: "天猫旗舰店",
    changeType: "下调",
    amount: 200,
    date: "2024-01-15",
  },
  {
    product: "MacBook Air M2",
    channel: "京东自营",
    changeType: "上调",
    amount: 300,
    date: "2024-01-14",
  },
  {
    product: "AirPods Pro",
    channel: "线下门店",
    changeType: "下调",
    amount: 50,
    date: "2024-01-13",
  },
]);

const historyColumns: TableColumn<PriceHistoryRow>[] = [
  { accessorKey: "product", header: "商品" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "changeType", header: "变更类型" },
  {
    id: "amount",
    accessorKey: "amount",
    header: () => h("div", { class: "text-right" }, "变更金额"),
    meta: { class: { td: "text-right" } },
  },
  { accessorKey: "date", header: "变更时间" },
];

/** ======================
 *  待处理任务
 *  ====================== */
const pendingTasks = ref([
  {
    id: 1,
    title: "合同价格即将到期",
    description: "8个合同价格将在本月到期",
    icon: "i-heroicons-exclamation-triangle",
    iconColor: "text-yellow-500",
    priority: "高",
  },
  {
    id: 2,
    title: "价格规则冲突",
    description: "检测到3条价格规则存在冲突",
    icon: "i-heroicons-x-circle",
    iconColor: "text-red-500",
    priority: "高",
  },
  {
    id: 3,
    title: "竞品价格变动",
    description: "15个商品的竞品价格发生变化",
    icon: "i-heroicons-chart-bar",
    iconColor: "text-blue-500",
    priority: "中",
  },
]);

/** ======================
 *  工具函数
 *  ====================== */
const formatCNY = (n: number) =>
  new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "CNY",
    maximumFractionDigits: 0,
  }).format(n);

/** ======================
 *  行为
 *  ====================== */
function onRefresh() {
  selectedProduct.value = "";
  selectedChannel.value = "";
  selectedTier.value = "";
  quantity.value = 1;
  calculatedPrice.value = null;
  priceTrace.value = [];
  toast.add({ title: "已刷新", color: "neutral" });
}

async function calculatePrice() {
  if (!selectedProduct.value || !selectedChannel.value || !selectedTier.value) {
    toast.add({ title: "请选择商品、渠道与客户层级", color: "warning" });
    return;
  }
  if (!quantity.value || quantity.value < 1) quantity.value = 1;

  calculating.value = true;
  await new Promise((r) => setTimeout(r, 400));

  // Demo：基准价可按商品映射
  const basePriceMap: Record<string, number> = {
    iphone15pro: 8999,
    macbookair: 8999,
    airpodspro: 1899,
  };
  const basePrice = basePriceMap[selectedProduct.value] ?? 8999;

  // 层级折扣（示例：按比例并四舍五入到元）
  const tierRate: Record<Tier, number> = {
    regular: 0,
    vip: 0.05,
    enterprise: 0.08,
    distributor: 0.12,
  };
  const tRate = selectedTier.value
    ? (tierRate[selectedTier.value as Tier] ?? 0)
    : 0;
  const tierDiscount = Math.round(basePrice * tRate);

  // 批量折扣（示例：满10件立减200）
  const bulkDiscount = quantity.value >= 10 ? 200 : 0;

  const totalDiscount = tierDiscount + bulkDiscount;
  const finalPrice = Math.max(0, basePrice - totalDiscount);

  calculatedPrice.value = { basePrice, discount: totalDiscount, finalPrice };

  priceTrace.value = [
    {
      rule: "基准价格",
      description: "商品建议零售价",
      change: 0,
      price: basePrice,
    },
    {
      rule: "客户层级折扣",
      description:
        tRate > 0 ? `按层级折扣 ${Math.round(tRate * 100)}%` : "无折扣",
      change: -tierDiscount,
      price: basePrice - tierDiscount,
    },
  ];
  if (bulkDiscount) {
    priceTrace.value.push({
      rule: "批量折扣",
      description: "数量≥10件",
      change: -bulkDiscount,
      price: basePrice - tierDiscount - bulkDiscount,
    });
  }

  calculating.value = false;
}
</script>
