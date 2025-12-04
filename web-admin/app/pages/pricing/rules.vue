<template>
  <div class="p-6 space-y-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
        {{ $t("pricing.rules") }}
      </h1>
      <div class="flex gap-3">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-heroicons-play"
          @click="runBatch"
        >
          批量执行
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus" @click="onCreate">
          {{ $t("common.add") }}{{ $t("pricing.rules") }}
        </UButton>
      </div>
    </div>

    <!-- 统计 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <UCard>
        <div class="flex items-center">
          <UIcon name="i-heroicons-cog-6-tooth" class="w-8 h-8 text-blue-500" />
          <div class="ml-4">
            <p class="text-sm text-gray-500">总规则数</p>
            <p class="text-2xl font-bold">{{ rules.length }}</p>
          </div>
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center">
          <UIcon
            name="i-heroicons-check-circle"
            class="w-8 h-8 text-green-500"
          />
          <div class="ml-4">
            <p class="text-sm text-gray-500">活跃规则</p>
            <p class="text-2xl font-bold">
              {{ rules.filter((r) => r.status === "活跃").length }}
            </p>
          </div>
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center">
          <UIcon
            name="i-heroicons-pause-circle"
            class="w-8 h-8 text-yellow-500"
          />
          <div class="ml-4">
            <p class="text-sm text-gray-500">暂停规则</p>
            <p class="text-2xl font-bold">
              {{ rules.filter((r) => r.status === "暂停").length }}
            </p>
          </div>
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center">
          <UIcon
            name="i-heroicons-exclamation-triangle"
            class="w-8 h-8 text-red-500"
          />
          <div class="ml-4">
            <p class="text-sm text-gray-500">冲突规则</p>
            <p class="text-2xl font-bold">4</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 筛选 -->
    <UCard>
      <div class="grid grid-cols-1 md:grid-cols-5 gap-4">
        <UInput
          v-model="searchQuery"
          :placeholder="$t('common.search')"
          icon="i-heroicons-magnifying-glass"
        />
        <USelect
          v-model="selectedType"
          :options="ruleTypeOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="规则类型"
        />
        <USelect
          v-model="selectedChannel"
          :options="channelOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="适用渠道"
        />
        <USelect
          v-model="selectedStatus"
          :options="statusOptions"
          option-attribute="label"
          value-attribute="value"
          :placeholder="$t('common.status')"
        />
        <USelect
          v-model="selectedPriority"
          :options="priorityOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="优先级"
        />
      </div>
    </UCard>

    <!-- 列表 -->
    <UCard>
      <UTable :data="filteredRules" :columns="columns" :loading="loading">
        <!-- 类型 -->
        <template #type-cell="{ row }">
          <UBadge :color="getTypeColor(row.original.type)" variant="subtle">
            {{ row.original.type }}
          </UBadge>
        </template>

        <!-- 条件 -->
        <template #conditions-cell="{ row }">
          <div class="text-sm">
            <p class="font-medium">{{ row.original.conditions.primary }}</p>
            <p class="text-gray-500 text-xs">
              {{ row.original.conditions.secondary }}
            </p>
          </div>
        </template>

        <!-- 调整 -->
        <template #adjustment-cell="{ row }">
          <div class="text-sm">
            <span
              :class="
                row.original.adjustment.type === '折扣'
                  ? 'text-green-600'
                  : 'text-blue-600'
              "
            >
              {{ row.original.adjustment.value }}
            </span>
            <p class="text-xs text-gray-500">
              {{ row.original.adjustment.type }}
            </p>
          </div>
        </template>

        <!-- 优先级 -->
        <template #priority-cell="{ row }">
          <UBadge
            :color="getPriorityColor(row.original.priority)"
            variant="subtle"
          >
            {{ row.original.priority }}
          </UBadge>
        </template>

        <!-- 状态&开关 -->
        <template #status-cell="{ row }">
          <div class="flex items-center gap-2">
            <UBadge
              :color="getStatusColor(row.original.status)"
              variant="subtle"
            >
              {{ row.original.status }}
            </UBadge>
            <USwitch
              v-model="row.original.enabled"
              size="sm"
              @change="toggleRule(row.original)"
            />
          </div>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewRule(row.original)"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
              @click="editRule(row.original)"
            >
              {{ $t("common.edit") }}
            </UButton>
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-play"
              @click="openTest(row.original)"
            >
              测试
            </UButton>
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              @click="deleteRule(row.original)"
            >
              {{ $t("common.delete") }}
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 详情 / 测试 -->
    <UModal v-model="showDetailModal" :ui="{ width: 'max-w-4xl' }">
      <UCard>
        <template #header>
          <div class="flex justify-between items-center">
            <h3 class="text-lg font-semibold">{{ selectedRule?.name }}</h3>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="showDetailModal = false"
            />
          </div>
        </template>

        <div v-if="selectedRule" class="space-y-6">
          <!-- 基本信息 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h4 class="font-medium mb-3">基本信息</h4>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-gray-600">规则类型:</span>
                  <UBadge
                    :color="getTypeColor(selectedRule.type)"
                    variant="subtle"
                    >{{ selectedRule.type }}</UBadge
                  >
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">优先级:</span>
                  <UBadge
                    :color="getPriorityColor(selectedRule.priority)"
                    variant="subtle"
                    >{{ selectedRule.priority }}</UBadge
                  >
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">状态:</span>
                  <UBadge
                    :color="getStatusColor(selectedRule.status)"
                    variant="subtle"
                    >{{ selectedRule.status }}</UBadge
                  >
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">适用渠道:</span>
                  <span>{{ selectedRule.channels.join("，") }}</span>
                </div>
              </div>
            </div>
            <div>
              <h4 class="font-medium mb-3">执行统计</h4>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-gray-600">执行次数:</span
                  ><span>{{ selectedRule.executionCount }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">成功率:</span
                  ><span class="text-green-600"
                    >{{ selectedRule.successRate }}%</span
                  >
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">最后执行:</span
                  ><span>{{ selectedRule.lastExecuted }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">创建时间:</span
                  ><span>{{ selectedRule.createdAt }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 触发条件 -->
          <div>
            <h4 class="font-medium mb-3">触发条件</h4>
            <div class="bg-gray-50 p-4 rounded-lg space-y-3">
              <div
                v-for="(c, i) in selectedRule.detailedConditions"
                :key="i"
                class="flex items-center gap-3"
              >
                <div
                  class="w-6 h-6 bg-blue-500 text-white rounded-full flex items-center justify-center text-xs font-bold"
                >
                  {{ i + 1 }}
                </div>
                <div class="flex-1">
                  <p class="text-sm font-medium">{{ c.field }}</p>
                  <p class="text-xs text-gray-500">
                    {{ c.operator }} {{ c.value }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- 调整配置 -->
          <div>
            <h4 class="font-medium mb-3">价格调整</h4>
            <div
              class="bg-gray-50 p-4 rounded-lg grid grid-cols-1 md:grid-cols-3 gap-4 text-sm"
            >
              <div>
                <span class="text-gray-600">调整类型:</span>
                <p class="font-medium">{{ selectedRule.adjustment.type }}</p>
              </div>
              <div>
                <span class="text-gray-600">调整值:</span>
                <p class="font-medium">{{ selectedRule.adjustment.value }}</p>
              </div>
              <div>
                <span class="text-gray-600">舍入规则:</span>
                <p class="font-medium">
                  {{ selectedRule.adjustment.rounding }}
                </p>
              </div>
            </div>
          </div>

          <!-- 规则测试 -->
          <div>
            <h4 class="font-medium mb-3">规则测试</h4>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <USelect
                v-model="testProduct"
                :options="productOptions"
                option-attribute="label"
                value-attribute="value"
                placeholder="选择商品"
              />
              <USelect
                v-model="testChannel"
                :options="channelOptions"
                option-attribute="label"
                value-attribute="value"
                placeholder="选择渠道"
              />
              <UButton @click="runRuleTest" color="primary" block
                >执行测试</UButton
              >
            </div>

            <div v-if="testResult" class="mt-4 p-4 bg-blue-50 rounded-lg">
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                <div>
                  <span class="text-gray-600">原价格:</span>
                  <p class="font-medium">¥{{ testResult.originalPrice }}</p>
                </div>
                <div>
                  <span class="text-gray-600">调整后:</span>
                  <p class="font-medium text-green-600">
                    ¥{{ testResult.adjustedPrice }}
                  </p>
                </div>
                <div>
                  <span class="text-gray-600">差额:</span>
                  <p
                    class="font-medium"
                    :class="
                      testResult.difference > 0
                        ? 'text-red-600'
                        : 'text-green-600'
                    "
                  >
                    {{ testResult.difference > 0 ? "+" : "" }}¥{{
                      testResult.difference
                    }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

const { t } = useI18n();
const toast = useToast();

/** 类型 */
type Rule = {
  id: string;
  name: string;
  type: "批量折扣" | "客户层级" | "时间段" | "库存水位" | "竞品价格";
  conditions: { primary: string; secondary?: string };
  adjustment: {
    type: "折扣" | "加价" | "阶梯折扣";
    value: string;
    rounding: string;
  };
  priority: "高" | "中" | "低";
  status: "活跃" | "暂停" | "已停用";
  enabled: boolean;
  channels: string[];
  executionCount: number;
  successRate: number;
  lastExecuted: string;
  createdAt: string;
  detailedConditions: { field: string; operator: string; value: string }[];
};

/** 状态 */
const searchQuery = ref("");
const selectedType = ref<string>("");
const selectedChannel = ref<string>(""); // 与数据中文一致
const selectedStatus = ref<string>("");
const selectedPriority = ref<string>("");
const loading = ref(false);

const showDetailModal = ref(false);
const selectedRule = ref<Rule | null>(null);

const testProduct = ref<string>("");
const testChannel = ref<string>("");
const testResult = ref<{
  originalPrice: number;
  adjustedPrice: number;
  difference: number;
} | null>(null);

/** 选项（value 与数据字段保持一致） */
const ruleTypeOptions = [
  { label: "全部类型", value: "" },
  { label: "批量折扣", value: "批量折扣" },
  { label: "客户层级", value: "客户层级" },
  { label: "时间段", value: "时间段" },
  { label: "库存水位", value: "库存水位" },
  { label: "竞品价格", value: "竞品价格" },
];
const channelOptions = [
  { label: "全部渠道", value: "" },
  { label: "天猫旗舰店", value: "天猫旗舰店" },
  { label: "京东自营", value: "京东自营" },
  { label: "线下门店", value: "线下门店" },
  { label: "企业直销", value: "企业直销" },
];
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "活跃", value: "活跃" },
  { label: "暂停", value: "暂停" },
  { label: "已停用", value: "已停用" },
];
const priorityOptions = [
  { label: "全部优先级", value: "" },
  { label: "高", value: "高" },
  { label: "中", value: "中" },
  { label: "低", value: "低" },
];
const productOptions = [
  { label: "iPhone 15 Pro", value: "iphone15pro" },
  { label: "MacBook Air M2", value: "macbookair" },
  { label: "AirPods Pro", value: "airpodspro" },
];

/** 列 */
const columns: TableColumn<Rule>[] = [
  { accessorKey: "name", header: "规则名称" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "conditions", header: "触发条件" },
  { accessorKey: "adjustment", header: "价格调整" },
  { accessorKey: "priority", header: "优先级" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
];

/** 数据 */
const rules = ref<Rule[]>([
  {
    id: "PR001",
    name: "VIP客户5%折扣",
    type: "客户层级",
    conditions: { primary: "客户层级 = VIP", secondary: "订单金额 ≥ ¥1000" },
    adjustment: { type: "折扣", value: "5%", rounding: "四舍五入到元" },
    priority: "高",
    status: "活跃",
    enabled: true,
    channels: ["天猫旗舰店", "京东自营"],
    executionCount: 1250,
    successRate: 98.5,
    lastExecuted: "2024-01-15 14:30",
    createdAt: "2023-12-01",
    detailedConditions: [
      { field: "客户层级", operator: "等于", value: "VIP" },
      { field: "订单金额", operator: "大于等于", value: "¥1000" },
    ],
  },
  {
    id: "PR002",
    name: "批量采购阶梯折扣",
    type: "批量折扣",
    conditions: { primary: "数量 ≥ 10件", secondary: "单品价格 ≥ ¥500" },
    adjustment: {
      type: "阶梯折扣",
      value: "10件8折，50件7折",
      rounding: "四舍五入到角",
    },
    priority: "中",
    status: "活跃",
    enabled: true,
    channels: ["企业直销", "线下门店"],
    executionCount: 856,
    successRate: 96.2,
    lastExecuted: "2024-01-15 10:15",
    createdAt: "2023-11-15",
    detailedConditions: [
      { field: "购买数量", operator: "大于等于", value: "10" },
      { field: "单品价格", operator: "大于等于", value: "¥500" },
    ],
  },
  {
    id: "PR003",
    name: "夜间时段加价",
    type: "时间段",
    conditions: { primary: "时间 22:00-06:00", secondary: "配送服务 = 即时达" },
    adjustment: { type: "加价", value: "+¥20", rounding: "不舍入" },
    priority: "低",
    status: "暂停",
    enabled: false,
    channels: ["线下门店"],
    executionCount: 234,
    successRate: 89.7,
    lastExecuted: "2024-01-10 23:45",
    createdAt: "2023-10-20",
    detailedConditions: [
      { field: "下单时间", operator: "在范围内", value: "22:00-06:00" },
      { field: "配送服务", operator: "等于", value: "即时达" },
    ],
  },
  {
    id: "PR004",
    name: "库存清仓折扣",
    type: "库存水位",
    conditions: { primary: "库存 ≤ 10件", secondary: "商品上架 ≥ 90天" },
    adjustment: { type: "折扣", value: "30%", rounding: "向下取整到元" },
    priority: "高",
    status: "活跃",
    enabled: true,
    channels: ["天猫旗舰店", "京东自营", "线下门店"],
    executionCount: 445,
    successRate: 94.1,
    lastExecuted: "2024-01-14 16:20",
    createdAt: "2023-09-10",
    detailedConditions: [
      { field: "当前库存", operator: "小于等于", value: "10" },
      { field: "上架天数", operator: "大于等于", value: "90" },
    ],
  },
]);

/** 过滤 */
const filteredRules = computed(() => {
  let list = rules.value;

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (rule) =>
        rule.name.toLowerCase().includes(q) ||
        rule.type.toLowerCase().includes(q)
    );
  }
  if (selectedType.value) {
    list = list.filter((rule) => rule.type === selectedType.value);
  }
  if (selectedChannel.value) {
    list = list.filter((rule) => rule.channels.includes(selectedChannel.value));
  }
  if (selectedStatus.value) {
    list = list.filter((rule) => rule.status === selectedStatus.value);
  }
  if (selectedPriority.value) {
    list = list.filter((rule) => rule.priority === selectedPriority.value);
  }
  return list;
});

/** 颜色映射 */
function getTypeColor(type: Rule["type"]) {
  return (
    {
      客户层级: "blue",
      批量折扣: "green",
      时间段: "purple",
      库存水位: "orange",
      竞品价格: "red",
    }[type] || "neutral"
  );
}
function getPriorityColor(p: Rule["priority"]) {
  return { 高: "error", 中: "warning", 低: "neutral" }[p] || "neutral";
}
function getStatusColor(s: Rule["status"]) {
  return (
    { 活跃: "success", 暂停: "warning", 已停用: "neutral" }[s] || "neutral"
  );
}

/** 行为 */
function viewRule(rule: Rule) {
  selectedRule.value = rule;
  showDetailModal.value = true;
}
function editRule(rule: Rule) {
  toast.add({ title: `编辑：${rule.name}`, color: "primary" });
}
function openTest(rule: Rule) {
  selectedRule.value = rule;
  testProduct.value = "";
  testChannel.value = "";
  testResult.value = null;
  showDetailModal.value = true;
}
function deleteRule(rule: Rule) {
  rules.value = rules.value.filter((r) => r.id !== rule.id);
  toast.add({ title: "已删除规则", color: "error" });
}
function toggleRule(rule: Rule) {
  rule.status = rule.enabled ? "活跃" : "暂停";
  toast.add({
    title: `${rule.name} 已${rule.enabled ? "启用" : "暂停"}`,
    color: rule.enabled ? "success" : "warning",
  });
}
function runBatch() {
  toast.add({ title: "开始批量执行（示例）", color: "neutral" });
}
function onCreate() {
  toast.add({ title: "新建规则（示例）", color: "primary" });
}

/** 测试逻辑（示例） */
function runRuleTest() {
  if (!testProduct.value || !testChannel.value || !selectedRule.value) return;
  const original = 8999;
  let adjusted = original;

  // 简单演示：若规则为“折扣 5%”或包含百分号则应用
  const val = selectedRule.value.adjustment.value;
  if (val.includes("%")) {
    const pct = Number(val.replace("%", "")) / 100;
    adjusted = Math.round(original * (1 - pct));
  } else if (val.startsWith("+¥")) {
    adjusted = original + Number(val.replace("+¥", ""));
  }

  testResult.value = {
    originalPrice: original,
    adjustedPrice: adjusted,
    difference: adjusted - original,
  };
  toast.add({ title: "测试完成", color: "success" });
}
</script>
