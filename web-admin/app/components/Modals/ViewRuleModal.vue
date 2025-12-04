<template>
  <UModal
    v-model:open="isOpen"
    title="规则详情"
    description="查看会员等级规则的完整信息"
    :close="{ onClick: () => close(false) }"
    :ui="{
      content: 'w-full sm:max-w-4xl',
      body: 'p-0',
      footer: 'justify-end',
    }"
  >
    <!-- Body -->
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="space-y-6 p-4 sm:p-6">
          <!-- 规则基本信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-4">
                <div
                  class="w-16 h-16 rounded-full flex items-center justify-center text-white text-xl font-bold"
                  :class="getRuleTypeClass(rule.type)"
                >
                  <UIcon :name="getRuleTypeIcon(rule.type)" class="w-8 h-8" />
                </div>
                <div>
                  <h3 class="text-2xl font-bold text-gray-900 dark:text-white">
                    {{ rule.name }}
                  </h3>
                  <p class="text-gray-600 dark:text-gray-400 mt-1">
                    {{ rule.description }}
                  </p>
                </div>
              </div>
              <div class="flex flex-col items-end gap-2">
                <UBadge
                  :color="rule.status === 'active' ? 'success' : 'neutral'"
                  variant="soft"
                  size="lg"
                >
                  {{ rule.status === "active" ? "启用中" : "已停用" }}
                </UBadge>
                <UBadge
                  :color="rule.type === 'promotion' ? 'success' : 'warning'"
                  variant="outline"
                >
                  {{ rule.type === "promotion" ? "晋升规则" : "降级规则" }}
                </UBadge>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
              <div
                class="text-center p-4 bg-blue-50 dark:bg-blue-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-blue-600 dark:text-blue-400"
                >
                  {{ rule.priority }}
                </div>
                <div class="text-sm text-blue-600 dark:text-blue-400 mt-1">
                  优先级
                </div>
              </div>

              <div
                class="text-center p-4 bg-green-50 dark:bg-green-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-green-600 dark:text-green-400"
                >
                  {{ rule.conditions.length }}
                </div>
                <div class="text-sm text-green-600 dark:text-green-400 mt-1">
                  条件数量
                </div>
              </div>

              <div
                class="text-center p-4 bg-purple-50 dark:bg-purple-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-purple-600 dark:text-purple-400"
                >
                  {{ rule.hasExpiry ? `${rule.expiryDays}天` : "永久" }}
                </div>
                <div class="text-sm text-purple-600 dark:text-purple-400 mt-1">
                  有效期
                </div>
              </div>
            </div>
          </div>

          <!-- 触发条件 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              触发条件
            </h3>

            <div class="space-y-4">
              <div
                v-for="(condition, index) in rule.conditions"
                :key="index"
                class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div class="flex items-center gap-3 mb-3">
                  <div
                    class="w-8 h-8 rounded-full flex items-center justify-center"
                    :class="getConditionIconClass(condition.type)"
                  >
                    <UIcon
                      :name="getConditionIcon(condition.type)"
                      class="w-4 h-4"
                    />
                  </div>
                  <span class="font-medium text-gray-900 dark:text-white">
                    {{ getConditionTitle(condition.type) }}
                  </span>
                  <UBadge variant="outline" size="sm">
                    条件 {{ index + 1 }}
                  </UBadge>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm">
                  <div>
                    <span class="text-gray-500 dark:text-gray-400"
                      >操作符：</span
                    >
                    <span class="text-gray-900 dark:text-white font-medium">
                      {{ getOperatorLabel(condition.operator) }}
                    </span>
                  </div>
                  <div>
                    <span class="text-gray-500 dark:text-gray-400"
                      >目标值：</span
                    >
                    <span class="text-gray-900 dark:text-white font-medium">
                      {{ formatConditionValue(condition) }}
                    </span>
                  </div>
                  <div v-if="condition.period">
                    <span class="text-gray-500 dark:text-gray-400"
                      >时间周期：</span
                    >
                    <span class="text-gray-900 dark:text-white font-medium">
                      {{ condition.period }} 天
                    </span>
                  </div>
                </div>

                <div class="mt-3 p-3 bg-gray-50 dark:bg-gray-900 rounded-lg">
                  <p class="text-sm text-gray-700 dark:text-gray-300">
                    {{ condition.description }}
                  </p>
                </div>
              </div>
            </div>

            <div class="mt-4 p-3 bg-blue-50 dark:bg-blue-950 rounded-lg">
              <div class="flex items-start gap-2">
                <UIcon
                  name="i-heroicons-information-circle"
                  class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5"
                />
                <div class="text-sm text-blue-700 dark:text-blue-300">
                  <strong>执行逻辑：</strong>
                  {{
                    rule.type === "promotion"
                      ? "满足所有条件时触发晋升"
                      : "满足任一条件时触发降级"
                  }}
                </div>
              </div>
            </div>
          </div>

          <!-- 有效期设置 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              有效期设置
            </h3>

            <div v-if="rule.hasExpiry" class="space-y-4">
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div
                  class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
                >
                  <div class="flex items-center gap-3 mb-2">
                    <UIcon
                      name="i-heroicons-clock"
                      class="w-5 h-5 text-purple-600"
                    />
                    <span class="font-medium text-gray-900 dark:text-white"
                      >有效期类型</span
                    >
                  </div>
                  <div class="text-lg font-bold text-gray-900 dark:text-white">
                    {{
                      rule.expiryType === "absolute" ? "绝对期限" : "相对期限"
                    }}
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                    {{
                      rule.expiryType === "absolute"
                        ? "从规则创建时开始计算"
                        : "从满足条件时开始计算"
                    }}
                  </div>
                </div>

                <div
                  class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
                >
                  <div class="flex items-center gap-3 mb-2">
                    <UIcon
                      name="i-heroicons-calendar-days"
                      class="w-5 h-5 text-green-600"
                    />
                    <span class="font-medium text-gray-900 dark:text-white"
                      >有效天数</span
                    >
                  </div>
                  <div class="text-lg font-bold text-gray-900 dark:text-white">
                    {{ rule.expiryDays }} 天
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                    约 {{ Math.round((rule.expiryDays || 0) / 30) }} 个月
                  </div>
                </div>
              </div>

              <div class="p-3 bg-yellow-50 dark:bg-yellow-950 rounded-lg">
                <div class="flex items-start gap-2">
                  <UIcon
                    name="i-heroicons-exclamation-triangle"
                    class="w-5 h-5 text-yellow-600 dark:text-yellow-400 mt-0.5"
                  />
                  <div class="text-sm text-yellow-700 dark:text-yellow-300">
                    <strong>到期处理：</strong>
                    {{
                      rule.type === "promotion"
                        ? "等级将在到期后自动降级到上一级别"
                        : "降级规则到期后将不再生效"
                    }}
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="text-center py-8">
              <UIcon
                name="i-heroicons-infinity"
                class="w-12 h-12 mx-auto mb-2 text-gray-400"
              />
              <p class="text-gray-500 dark:text-gray-400">
                此规则永久有效，无到期时间
              </p>
            </div>
          </div>

          <!-- 规则信息 -->
          <div>
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              规则信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
              <div>
                <span class="text-gray-500 dark:text-gray-400">规则ID：</span>
                <span class="text-gray-900 dark:text-white font-mono">{{
                  rule.id
                }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">优先级：</span>
                <span class="text-gray-900 dark:text-white">{{
                  rule.priority
                }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">创建时间：</span>
                <span class="text-gray-900 dark:text-white">{{
                  formatDate(rule.createdAt)
                }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">更新时间：</span>
                <span class="text-gray-900 dark:text-white">{{
                  formatDate(rule.updatedAt)
                }}</span>
              </div>
            </div>
          </div>
        </div>
      </UCard>
    </template>

    <!-- Footer -->
    <template #footer>
      <div class="flex gap-3">
        <UButton variant="ghost" @click="close(false)">关闭</UButton>
        <UButton color="primary" @click="editRule">编辑规则</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
interface MembershipRule {
  id: string;
  name: string;
  description: string;
  type: "promotion" | "demotion";
  conditions: Array<{
    type: "spend" | "growth" | "frequency";
    operator: "gte" | "lte" | "eq";
    value: number;
    period?: number;
    description: string;
  }>;
  hasExpiry: boolean;
  expiryType?: "absolute" | "relative";
  expiryDays?: number;
  status: "active" | "inactive";
  priority: number;
  createdAt: string;
  updatedAt: string;
}

const props = defineProps<{ open?: boolean; rule: MembershipRule }>();
const emit = defineEmits<{
  "update:open": [boolean];
  edit: [rule: MembershipRule];
  close: [payload: any];
}>();

// 统一控制打开（本地 & overlay）
const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

function close(payload: any) {
  emit("close", payload);
  emit("update:open", false);
}

// 获取规则类型图标
const getRuleTypeIcon = (type: string) => {
  return type === "promotion"
    ? "i-heroicons-arrow-trending-up"
    : "i-heroicons-arrow-trending-down";
};

// 获取规则类型样式
const getRuleTypeClass = (type: string) => {
  return type === "promotion"
    ? "bg-green-500 dark:bg-green-600"
    : "bg-orange-500 dark:bg-orange-600";
};

// 获取条件图标
const getConditionIcon = (type: string) => {
  const icons = {
    spend: "i-heroicons-currency-dollar",
    growth: "i-heroicons-chart-bar-square",
    frequency: "i-heroicons-arrow-path",
  };
  return icons[type as keyof typeof icons] || "i-heroicons-cog-6-tooth";
};

// 获取条件图标样式
const getConditionIconClass = (type: string) => {
  const classes = {
    spend: "bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-400",
    growth: "bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400",
    frequency:
      "bg-purple-100 dark:bg-purple-900 text-purple-600 dark:text-purple-400",
  };
  return (
    classes[type as keyof typeof classes] ||
    "bg-gray-100 dark:bg-gray-900 text-gray-600 dark:text-gray-400"
  );
};

// 获取条件标题
const getConditionTitle = (type: string) => {
  const titles = {
    spend: "消费金额条件",
    growth: "成长值条件",
    frequency: "消费频次条件",
  };
  return titles[type as keyof typeof titles] || "未知条件";
};

// 获取操作符标签
const getOperatorLabel = (operator: string) => {
  const labels = {
    gte: "大于等于 (≥)",
    lte: "小于等于 (≤)",
    eq: "等于 (=)",
  };
  return labels[operator as keyof typeof labels] || operator;
};

// 格式化条件值
const formatConditionValue = (condition: any) => {
  if (condition.type === "spend") {
    return `¥${condition.value.toLocaleString()}`;
  }
  return condition.value.toString();
};

// 日期格式化
const formatDate = (iso: string) => new Date(iso).toLocaleString("zh-CN");

// 编辑规则
const editRule = () => {
  emit("edit", props.rule);
  close({ action: "edit", rule: props.rule });
};
</script>
