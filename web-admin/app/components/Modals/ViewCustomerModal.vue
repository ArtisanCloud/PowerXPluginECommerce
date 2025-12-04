<template>
  <UModal
    v-model:open="isOpen"
    title="客户详情"
    description="查看客户完整信息"
    :close="{ onClick: () => close(false) }"
    :ui="{
      content: 'w-full sm:max-w-3xl',
      body: 'p-0',
      footer: 'justify-end',
    }"
  >
    <!-- Body -->
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="space-y-6 p-4 sm:p-6">
          <!-- 客户基本信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                基本信息
              </h3>
              <UBadge
                :color="customer.status === 'active' ? 'success' : 'neutral'"
                variant="soft"
                size="lg"
              >
                {{ customer.status === "active" ? "活跃客户" : "非活跃客户" }}
              </UBadge>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  客户ID
                </label>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ customer.id }}
                </p>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  客户姓名
                </label>
                <p
                  class="mt-1 text-sm text-gray-900 dark:text-white font-medium"
                >
                  {{ customer.name }}
                </p>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  客户类型
                </label>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ getCustomerTypeLabel(customer.customerType) }}
                </p>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  手机号码
                </label>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ customer.phone }}
                </p>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  邮箱地址
                </label>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ customer.email || "未填写" }}
                </p>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  性别
                </label>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ getGenderLabel(customer.gender) }}
                </p>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  出生日期
                </label>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ customer.birthDate || "未填写" }}
                </p>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  注册日期
                </label>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ customer.registrationDate }}
                </p>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  客户来源
                </label>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ getSourceLabel(customer.source) }}
                </p>
              </div>
            </div>

            <!-- 地址信息 -->
            <div v-if="customer.address" class="mt-4">
              <label
                class="text-sm font-medium text-gray-500 dark:text-gray-400"
              >
                详细地址
              </label>
              <p class="mt-1 text-sm text-gray-900 dark:text-white">
                {{ customer.address }}
              </p>
            </div>
          </div>

          <!-- 会员信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              会员信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  会员等级
                </label>
                <div class="mt-1">
                  <UBadge
                    :color="getMembershipColor(customer.membershipTier)"
                    variant="soft"
                  >
                    {{ getMembershipLabel(customer.membershipTier) }}
                  </UBadge>
                </div>
              </div>

              <div>
                <label
                  class="text-sm font-medium text-gray-500 dark:text-gray-400"
                >
                  客户标签
                </label>
                <div class="mt-1 flex flex-wrap gap-2">
                  <UBadge
                    v-for="tag in customer.tags"
                    :key="tag"
                    variant="outline"
                    size="sm"
                  >
                    {{ tag }}
                  </UBadge>
                  <span
                    v-if="!customer.tags?.length"
                    class="text-sm text-gray-500 dark:text-gray-400"
                  >
                    暂无标签
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 交易统计 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              交易统计
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
              <div
                class="text-center p-4 bg-blue-50 dark:bg-blue-950 rounded-lg"
              >
                <div
                  class="text-2xl font-bold text-blue-600 dark:text-blue-400"
                >
                  {{ customer.totalOrders || 0 }}
                </div>
                <div class="text-sm text-blue-600 dark:text-blue-400 mt-1">
                  订单总数
                </div>
              </div>

              <div
                class="text-center p-4 bg-green-50 dark:bg-green-950 rounded-lg"
              >
                <div
                  class="text-2xl font-bold text-green-600 dark:text-green-400"
                >
                  ¥{{ (customer.totalSpent || 0).toLocaleString() }}
                </div>
                <div class="text-sm text-green-600 dark:text-green-400 mt-1">
                  消费总额
                </div>
              </div>

              <div
                class="text-center p-4 bg-purple-50 dark:bg-purple-950 rounded-lg"
              >
                <div
                  class="text-2xl font-bold text-purple-600 dark:text-purple-400"
                >
                  ¥{{ getAverageOrderValue() }}
                </div>
                <div class="text-sm text-purple-600 dark:text-purple-400 mt-1">
                  平均订单价值
                </div>
              </div>
            </div>
          </div>

          <!-- 备注信息 -->
          <div v-if="customer.notes">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              备注信息
            </h3>
            <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
              <p
                class="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap"
              >
                {{ customer.notes }}
              </p>
            </div>
          </div>
        </div>
      </UCard>
    </template>

    <!-- Footer -->
    <template #footer>
      <div class="flex gap-3">
        <UButton variant="ghost" @click="close(false)">关闭</UButton>
        <UButton color="primary" @click="editCustomer">编辑客户</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
interface Customer {
  id: string;
  name: string;
  email: string;
  phone: string;
  customerType?: string;
  gender?: string;
  birthDate?: string;
  address?: string;
  membershipTier?: string;
  source?: string;
  tags?: string[];
  notes?: string;
  registrationDate: string;
  status: "active" | "inactive";
  totalOrders?: number;
  totalSpent?: number;
}

const props = defineProps<{ open?: boolean; customer: Customer }>();
const emit = defineEmits<{
  "update:open": [boolean];
  edit: [customer: Customer];
  close: [payload: any]; // overlay 用它来 resolve result
}>();

// 统一控制打开（本地 & overlay）
const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

function close(payload: any) {
  emit("close", payload); // -> overlay instance.result
  emit("update:open", false); // -> 本地 v-model:open
}

// 标签转换函数
const getCustomerTypeLabel = (type?: string) => {
  const labels = {
    individual: "个人客户",
    enterprise: "企业客户",
    vip: "VIP客户",
  };
  return labels[type as keyof typeof labels] || type || "未知";
};

const getGenderLabel = (gender?: string) => {
  const labels = {
    male: "男",
    female: "女",
    other: "其他",
  };
  return labels[gender as keyof typeof labels] || "未填写";
};

const getSourceLabel = (source?: string) => {
  const labels = {
    website: "官网注册",
    mobile_app: "手机APP",
    wechat: "微信小程序",
    referral: "朋友推荐",
    advertisement: "广告投放",
    offline_store: "线下门店",
    social_media: "社交媒体",
    other: "其他渠道",
  };
  return labels[source as keyof typeof labels] || source || "未知";
};

const getMembershipLabel = (tier?: string) => {
  const labels = {
    bronze: "青铜会员",
    silver: "白银会员",
    gold: "黄金会员",
    platinum: "铂金会员",
    diamond: "钻石会员",
  };
  return labels[tier as keyof typeof labels] || tier || "普通会员";
};

const getMembershipColor = (
  tier?: string
):
  | "primary"
  | "secondary"
  | "success"
  | "info"
  | "warning"
  | "error"
  | "neutral" => {
  const colors = {
    bronze: "warning" as const,
    silver: "neutral" as const,
    gold: "warning" as const,
    platinum: "info" as const,
    diamond: "primary" as const,
  };
  return colors[tier as keyof typeof colors] || "neutral";
};

// 计算平均订单价值
const getAverageOrderValue = () => {
  const orders = props.customer.totalOrders || 0;
  const spent = props.customer.totalSpent || 0;
  if (orders === 0) return "0";
  return Math.round(spent / orders).toLocaleString();
};

// 编辑客户
const editCustomer = () => {
  emit("edit", props.customer);
  close({ action: "edit", customer: props.customer });
};
</script>
