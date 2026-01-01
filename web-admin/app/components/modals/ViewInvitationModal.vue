<template>
  <UModal
    v-model:open="isOpen"
    title="邀请详情"
    description="查看邀请关系的完整信息"
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
          <!-- 邀请关系基本信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              邀请关系信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <!-- 邀请人信息 -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-white mb-3">
                  邀请人
                </h4>
                <div class="flex items-center gap-3">
                  <UAvatar
                    :src="invitation?.inviter?.avatar || ''"
                    :alt="invitation?.inviter?.name || ''"
                    size="md"
                    :ui="{ rounded: 'rounded-full' }"
                  />
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">
                      {{ invitation.inviter.name }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      {{ invitation.inviter.phone }}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      ID: {{ invitation.inviter.id }}
                    </div>
                  </div>
                </div>
              </div>

              <!-- 被邀请人信息 -->
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 class="font-medium text-gray-900 dark:text-white mb-3">
                  被邀请人
                </h4>
                <div class="flex items-center gap-3">
                  <UAvatar
                    :src="invitation?.invitee?.avatar || ''"
                    :alt="invitation?.invitee?.name || ''"
                    size="md"
                    :ui="{ rounded: 'rounded-full' }"
                  />
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">
                      {{ invitation.invitee.name }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      {{ invitation.invitee.phone }}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      ID: {{ invitation.invitee.id }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="mt-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span class="text-gray-500 dark:text-gray-400">邀请时间：</span>
                <span class="text-gray-900 dark:text-white">{{ invitation.invitedAt }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">关系ID：</span>
                <span class="text-gray-900 dark:text-white font-mono">{{ invitation.id }}</span>
              </div>
            </div>
          </div>

          <!-- 邀请状态 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              邀请状态
            </h3>

            <div class="flex items-center gap-3">
              <UBadge
                :color="invitation.status === 'active' ? 'success' : 'neutral'"
                variant="soft"
                size="lg"
              >
                {{ invitation.status === "active" ? "有效" : "无效" }}
              </UBadge>
              <div class="text-sm text-gray-600 dark:text-gray-400">
                {{ invitation.status === "active" ? "该邀请关系有效" : "该邀请关系已失效" }}
              </div>
            </div>

            <div class="mt-4 p-3 bg-blue-50 dark:bg-blue-950 rounded-lg">
              <div class="flex items-start gap-2">
                <UIcon
                  name="i-heroicons-information-circle"
                  class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5"
                />
                <div class="text-sm text-blue-700 dark:text-blue-300">
                  <strong>状态说明：</strong>
                  有效邀请表示被邀请人已完成注册并满足相关条件。无效邀请可能表示被邀请人未完成注册或邀请已过期。
                </div>
              </div>
            </div>
          </div>

          <!-- 相关奖励信息 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              相关奖励
            </h3>

            <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
              <div class="text-center py-4" v-if="!relatedRewards || relatedRewards.length === 0">
                <UIcon name="i-heroicons-gift" class="w-12 h-12 mx-auto text-gray-400 mb-2" />
                <p class="text-gray-500 dark:text-gray-400">暂无相关奖励记录</p>
              </div>

              <div v-else>
                <div
                  v-for="reward in relatedRewards"
                  :key="reward.id"
                  class="flex items-center justify-between py-3 border-b border-gray-100 dark:border-gray-800 last:border-0"
                >
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">
                      {{ reward.type }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      {{ reward.relatedRecord }}
                    </div>
                  </div>
                  <div class="text-right">
                    <div class="font-medium text-gray-900 dark:text-white">
                      ¥{{ reward.amount.toLocaleString() }}
                    </div>
                    <div class="text-sm" :class="getRewardStatusClass(reward.status)">
                      {{ getRewardStatusText(reward.status) }}
                    </div>
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

const props = defineProps<{ open?: boolean; invitation: Invitation }>();
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

// 模拟相关奖励数据
const relatedRewards = computed(() => {
  // 在实际应用中，这应该通过API获取与当前邀请相关的奖励记录
  // 这里我们模拟一些数据
  return [
    {
      id: "rew_1",
      customer: props.invitation.inviter,
      amount: 1200,
      type: "邀请奖励",
      relatedRecord: `邀请${props.invitation.invitee.name}注册`,
      createdAt: "2023-09-15 14:30:25",
      paidAt: "2023-09-16 09:15:30",
      status: "paid",
    }
  ];
});

// 奖励状态相关函数
const getRewardStatusClass = (status: string) => {
  switch (status) {
    case "paid":
      return "text-green-600 dark:text-green-400";
    case "pending":
      return "text-yellow-600 dark:text-yellow-400";
    case "cancelled":
      return "text-red-600 dark:text-red-400";
    default:
      return "text-gray-500 dark:text-gray-400";
  }
};

const getRewardStatusText = (status: string) => {
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
</script>
