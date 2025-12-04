<template>
  <UModal
    v-model:open="isOpen"
    title="等级详情"
    description="查看会员等级的完整信息"
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
          <!-- 等级基本信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-4">
                <div
                  class="w-16 h-16 rounded-full flex items-center justify-center text-white text-xl font-bold"
                  :style="{ backgroundColor: tier.color }"
                >
                  {{ tier.name.charAt(0) }}
                </div>
                <div>
                  <h3 class="text-2xl font-bold text-gray-900 dark:text-white">
                    {{ tier.name }}
                  </h3>
                  <p class="text-gray-600 dark:text-gray-400 mt-1">
                    {{ tier.description }}
                  </p>
                </div>
              </div>
              <UBadge
                :color="tier.status === 'active' ? 'success' : 'neutral'"
                variant="soft"
                size="lg"
              >
                {{ tier.status === "active" ? "启用中" : "已停用" }}
              </UBadge>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
              <div
                class="text-center p-4 bg-blue-50 dark:bg-blue-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-blue-600 dark:text-blue-400"
                >
                  {{ tier.level }}
                </div>
                <div class="text-sm text-blue-600 dark:text-blue-400 mt-1">
                  等级级别
                </div>
              </div>

              <div
                class="text-center p-4 bg-green-50 dark:bg-green-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-green-600 dark:text-green-400"
                >
                  {{ (tier.memberCount || 0).toLocaleString() }}
                </div>
                <div class="text-sm text-green-600 dark:text-green-400 mt-1">
                  当前会员数
                </div>
              </div>

              <div
                class="text-center p-4 bg-purple-50 dark:bg-purple-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-purple-600 dark:text-purple-400"
                >
                  {{ tier.benefits?.length || 0 }}
                </div>
                <div class="text-sm text-purple-600 dark:text-purple-400 mt-1">
                  权益数量
                </div>
              </div>
            </div>
          </div>

          <!-- 升级门槛 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              升级门槛
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <div
                class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div class="flex items-center gap-3 mb-2">
                  <UIcon
                    name="i-heroicons-currency-dollar"
                    class="w-5 h-5 text-green-600"
                  />
                  <span class="font-medium text-gray-900 dark:text-white"
                    >消费门槛</span
                  >
                </div>
                <div class="text-2xl font-bold text-gray-900 dark:text-white">
                  ¥{{ (tier.spendThreshold || 0).toLocaleString() }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  累计消费金额达到此数值可升级
                </div>
              </div>

              <div
                class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div class="flex items-center gap-3 mb-2">
                  <UIcon
                    name="i-heroicons-star"
                    class="w-5 h-5 text-yellow-600"
                  />
                  <span class="font-medium text-gray-900 dark:text-white"
                    >积分门槛</span
                  >
                </div>
                <div class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ (tier.pointsThreshold || 0).toLocaleString() }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  累计积分达到此数值可升级
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
                  <strong>升级规则：</strong
                  >满足消费门槛或积分门槛任一条件即可升级到此等级。
                </div>
              </div>
            </div>
          </div>

          <!-- 会员权益 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              会员权益
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div
                v-for="benefit in tier.benefits"
                :key="benefit.type"
                class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div class="flex items-center gap-3 mb-2">
                  <div
                    class="w-8 h-8 rounded-full flex items-center justify-center"
                    :class="getBenefitIconClass(benefit.type)"
                  >
                    <UIcon
                      :name="getBenefitIcon(benefit.type)"
                      class="w-4 h-4"
                    />
                  </div>
                  <span class="font-medium text-gray-900 dark:text-white">
                    {{ getBenefitTitle(benefit.type) }}
                  </span>
                </div>

                <div
                  class="text-lg font-bold text-gray-900 dark:text-white mb-1"
                >
                  {{ formatBenefitValue(benefit.type, benefit.value) }}
                </div>

                <div class="text-sm text-gray-500 dark:text-gray-400">
                  {{ benefit.description }}
                </div>
              </div>
            </div>
          </div>

          <!-- 等级规则 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              等级规则
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div
                v-for="rule in tier.rules"
                :key="rule.type"
                class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div class="flex items-center gap-3 mb-2">
                  <div
                    class="w-8 h-8 rounded-full flex items-center justify-center bg-orange-100 dark:bg-orange-900 text-orange-600 dark:text-orange-400"
                  >
                    <UIcon
                      name="i-heroicons-document-text"
                      class="w-4 h-4"
                    />
                  </div>
                  <span class="font-medium text-gray-900 dark:text-white">
                    {{ getRuleTitle(rule.type) }}
                  </span>
                </div>

                <div
                  class="text-lg font-bold text-gray-900 dark:text-white mb-1"
                >
                  {{ rule.condition }} {{ rule.value }}
                </div>

                <div class="text-sm text-gray-500 dark:text-gray-400">
                  {{ rule.description }}
                </div>
              </div>
              
              <div
                v-if="!tier.rules || tier.rules.length === 0"
                class="col-span-2 text-center py-8 text-gray-500 dark:text-gray-400"
              >
                <UIcon name="i-heroicons-document-text" class="w-12 h-12 mx-auto mb-2" />
                <p>暂无规则</p>
              </div>
            </div>
          </div>

          <!-- 等级信息 -->
          <div>
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              等级信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
              <div>
                <span class="text-gray-500 dark:text-gray-400">等级ID：</span>
                <span class="text-gray-900 dark:text-white font-mono">{{
                  tier.id
                }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">等级颜色：</span>
                <div class="inline-flex items-center gap-2">
                  <div
                    class="w-4 h-4 rounded border border-gray-300"
                    :style="{ backgroundColor: tier.color }"
                  ></div>
                  <span class="text-gray-900 dark:text-white font-mono">{{
                    tier.color
                  }}</span>
                </div>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">创建时间：</span>
                <span class="text-gray-900 dark:text-white">{{
                  formatDate(tier.createdAt)
                }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">更新时间：</span>
                <span class="text-gray-900 dark:text-white">{{
                  formatDate(tier.updatedAt)
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
        <UButton color="primary" @click="editTier">编辑等级</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
interface MembershipTier {
  id: string;
  name: string;
  description: string;
  color: string;
  level: number;
  spendThreshold?: number;
  pointsThreshold?: number;
  benefits: Array<{
    type: string;
    value: number | string;
    description: string;
  }>;
  rules: Array<{
    type: string;
    condition: string;
    value: number | string;
    description: string;
  }>;
  memberCount?: number;
  status: "active" | "inactive";
  createdAt: string;
  updatedAt: string;
}

const props = defineProps<{ open?: boolean; tier: MembershipTier }>();
const emit = defineEmits<{
  "update:open": [boolean];
  edit: [tier: MembershipTier];
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

// 权益相关函数
const getBenefitIcon = (type: string) => {
  const icons = {
    discount: "i-heroicons-receipt-percent",
    points: "i-heroicons-star",
    shipping: "i-heroicons-truck",
    priority: "i-heroicons-bolt",
    exclusive: "i-heroicons-sparkles",
    personal: "i-heroicons-user-circle",
  };
  return icons[type as keyof typeof icons] || "i-heroicons-gift";
};

const getBenefitIconClass = (type: string) => {
  const classes = {
    discount:
      "bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-400",
    points:
      "bg-yellow-100 dark:bg-yellow-900 text-yellow-600 dark:text-yellow-400",
    shipping: "bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400",
    priority:
      "bg-purple-100 dark:bg-purple-900 text-purple-600 dark:text-purple-400",
    exclusive: "bg-pink-100 dark:bg-pink-900 text-pink-600 dark:text-pink-400",
    personal:
      "bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-400",
  };
  return (
    classes[type as keyof typeof classes] ||
    "bg-gray-100 dark:bg-gray-900 text-gray-600 dark:text-gray-400"
  );
};

const getBenefitTitle = (type: string) => {
  const titles = {
    discount: "购物折扣",
    points: "积分倍率",
    shipping: "免运费",
    priority: "优先服务",
    exclusive: "专属活动",
    personal: "专属客服",
  };
  return titles[type as keyof typeof titles] || "特殊权益";
};

const getRuleTitle = (type: string) => {
  const titles = {
    spend_limit: "消费限制",
    points_limit: "积分限制",
    purchase_count: "购买次数",
    special_event: "特殊活动",
    region_limit: "地区限制",
  };
  return titles[type as keyof typeof titles] || "自定义规则";
};

const formatBenefitValue = (type: string, value: number | string) => {
  switch (type) {
    case "discount":
      return `${10 - Number(value) / 10}折`;
    case "points":
      return `${value}x`;
    case "shipping":
      return "免费";
    default:
      return value.toString();
  }
};

// 日期格式化
const formatDate = (iso: string) => new Date(iso).toLocaleString("zh-CN");

// 编辑等级
const editTier = () => {
  emit("edit", props.tier);
  close({ action: "edit", tier: props.tier });
};
</script>
