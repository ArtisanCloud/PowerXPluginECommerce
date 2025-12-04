<template>
  <UModal
    v-model:open="isOpen"
    title="奖励详情"
    description="查看奖励发放的完整信息"
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
          <!-- 奖励基本信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              奖励信息
            </h3>

            <div class="flex items-center gap-4 mb-4">
              <UAvatar
                :src="reward.customer.avatar"
                :alt="reward.customer.name"
                size="lg"
                :ui="{ rounded: 'rounded-full' }"
              />
              <div>
                <div class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ reward.customer.name }}
                </div>
                <div class="text-gray-600 dark:text-gray-400">
                  {{ reward.customer.phone }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  客户ID: {{ reward.customer.id }}
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-4">
              <div class="p-3 bg-green-50 dark:bg-green-950 rounded-lg">
                <div class="text-sm text-green-600 dark:text-green-400">奖励金额</div>
                <div class="text-2xl font-bold text-green-700 dark:text-green-300">
                  ¥{{ reward.amount.toLocaleString() }}
                </div>
              </div>

              <div class="p-3 bg-blue-50 dark:bg-blue-950 rounded-lg">
                <div class="text-sm text-blue-600 dark:text-blue-400">奖励类型</div>
                <div class="text-lg font-bold text-blue-700 dark:text-blue-300">
                  {{ reward.type }}
                </div>
              </div>
            </div>
          </div>

          <!-- 奖励详情 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              奖励详情
            </h3>

            <div class="space-y-3">
              <div>
                <span class="text-gray-500 dark:text-gray-400">关联记录：</span>
                <span class="text-gray-900 dark:text-white">{{ reward.relatedRecord }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">奖励ID：</span>
                <span class="text-gray-900 dark:text-white font-mono">{{ reward.id }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">创建时间：</span>
                <span class="text-gray-900 dark:text-white">{{ reward.createdAt }}</span>
              </div>
              <div v-if="reward.paidAt">
                <span class="text-gray-500 dark:text-gray-400">发放时间：</span>
                <span class="text-gray-900 dark:text-white">{{ reward.paidAt }}</span>
              </div>
            </div>
          </div>

          <!-- 奖励状态 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              奖励状态
            </h3>

            <div class="flex items-center gap-3">
              <UBadge
                :color="getStatusColor(reward.status)"
                variant="soft"
                size="lg"
              >
                {{ getStatusText(reward.status) }}
              </UBadge>
            </div>

            <div class="mt-4 p-3 rounded-lg" :class="getStatusBgClass(reward.status)">
              <div class="flex items-start gap-2">
                <UIcon
                  :name="getStatusIcon(reward.status)"
                  class="w-5 h-5 mt-0.5"
                  :class="getStatusIconClass(reward.status)"
                />
                <div class="text-sm" :class="getStatusTextClass(reward.status)">
                  {{ getStatusDescription(reward.status) }}
                </div>
              </div>
            </div>
          </div>

          <!-- 相关邀请信息 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              相关邀请信息
            </h3>

            <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
              <div class="text-center py-4" v-if="!relatedInvitation">
                <UIcon name="i-heroicons-user-group" class="w-12 h-12 mx-auto text-gray-400 mb-2" />
                <p class="text-gray-500 dark:text-gray-400">暂无相关邀请记录</p>
              </div>

              <div v-else class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <UAvatar
                    :src="relatedInvitation.inviter.avatar"
                    :alt="relatedInvitation.inviter.name"
                    size="md"
                    :ui="{ rounded: 'rounded-full' }"
                  />
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">
                      {{ relatedInvitation.inviter.name }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      邀请了 {{ relatedInvitation.invitee.name }}
                    </div>
                  </div>
                </div>
                <div class="text-right">
                  <div class="text-sm text-gray-500 dark:text-gray-400">
                    邀请时间
                  </div>
                  <div class="text-sm text-gray-900 dark:text-white">
                    {{ relatedInvitation.invitedAt }}
                  </div>
                </div>
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
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
interface Reward {
  id: string;
  customer: {
    id: string;
    name: string;
    phone: string;
    avatar: string;
  };
  amount: number;
  type: string;
  relatedRecord: string;
  createdAt: string;
  paidAt: string | null;
  status: "paid" | "pending" | "cancelled";
}

interface Invitation {
  id: string;
  inviter: {
    id: string;
    name: string;
    phone: string;
    avatar: string;
  };
  invitee: {
    id: string;
    name: string;
    phone: string;
    avatar: string;
  };
  invitedAt: string;
  status: "active" | "inactive";
}

const props = defineProps<{ open?: boolean; reward: Reward }>();
const emit = defineEmits<{
  "update:open": [boolean];
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

// 模拟相关邀请数据
const relatedInvitation = computed(() => {
  // 在实际应用中，这应该通过API获取与当前奖励相关的邀请记录
  // 这里我们模拟一些数据
  return {
    id: "inv_1",
    inviter: props.reward.customer,
    invitee: {
      id: "cust_2",
      name: "李四",
      phone: "13800138002",
      avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2",
    },
    invitedAt: "2023-09-15 14:30:25",
    status: "active",
  };
});

// 状态相关函数
const getStatusColor = (status: string) => {
  switch (status) {
    case "paid":
      return "success";
    case "pending":
      return "warning";
    case "cancelled":
      return "error";
    default:
      return "neutral";
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case "paid":
      return "已发放";
    case "pending":
      return "待发放";
    case "cancelled":
      return "已取消";
    default:
      return "未知状态";
  }
};

const getStatusBgClass = (status: string) => {
  switch (status) {
    case "paid":
      return "bg-green-50 dark:bg-green-950";
    case "pending":
      return "bg-yellow-50 dark:bg-yellow-950";
    case "cancelled":
      return "bg-red-50 dark:bg-red-950";
    default:
      return "bg-gray-50 dark:bg-gray-950";
  }
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case "paid":
      return "i-heroicons-check-circle";
    case "pending":
      return "i-heroicons-clock";
    case "cancelled":
      return "i-heroicons-x-circle";
    default:
      return "i-heroicons-information-circle";
  }
};

const getStatusIconClass = (status: string) => {
  switch (status) {
    case "paid":
      return "text-green-600 dark:text-green-400";
    case "pending":
      return "text-yellow-600 dark:text-yellow-400";
    case "cancelled":
      return "text-red-600 dark:text-red-400";
    default:
      return "text-gray-600 dark:text-gray-400";
  }
};

const getStatusTextClass = (status: string) => {
  switch (status) {
    case "paid":
      return "text-green-700 dark:text-green-300";
    case "pending":
      return "text-yellow-700 dark:text-yellow-300";
    case "cancelled":
      return "text-red-700 dark:text-red-300";
    default:
      return "text-gray-700 dark:text-gray-300";
  }
};

const getStatusDescription = (status: string) => {
  switch (status) {
    case "paid":
      return "奖励已成功发放至客户账户。";
    case "pending":
      return "奖励正在处理中，将在下一个发放周期内完成发放。";
    case "cancelled":
      return "奖励发放已被取消，可能由于客户账户异常或其他原因。";
    default:
      return "奖励状态未知。";
  }
};
</script>
