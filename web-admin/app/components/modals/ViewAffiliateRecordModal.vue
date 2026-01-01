<template>
  <UModal
    v-model:open="isOpen"
    :title="modalTitle"
    :description="modalDescription"
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
          <!-- 邀请记录基本信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-4">
                <UAvatar
                  :src="record.avatar"
                  :alt="record.name"
                  size="lg"
                  :ui="{ rounded: 'rounded-full' }"
                />
                <div>
                  <h3 class="text-2xl font-bold text-gray-900 dark:text-white">
                    {{ record.name }}
                  </h3>
                  <p class="text-gray-600 dark:text-gray-400 mt-1">
                    {{ record.phone }}
                  </p>
                </div>
              </div>
              <UBadge
                :color="record.status === 'active' ? 'success' : 'neutral'"
                variant="soft"
                size="lg"
              >
                {{ record.status === "active" ? "有效" : "无效" }}
              </UBadge>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
              <div
                class="text-center p-4 bg-blue-50 dark:bg-blue-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-blue-600 dark:text-blue-400"
                >
                  {{ record.id }}
                </div>
                <div class="text-sm text-blue-600 dark:text-blue-400 mt-1">
                  记录ID
                </div>
              </div>

              <div
                class="text-center p-4 bg-green-50 dark:bg-green-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-green-600 dark:text-green-400"
                >
                  {{ formatDate(record.createdAt) }}
                </div>
                <div class="text-sm text-green-600 dark:text-green-400 mt-1">
                  创建时间
                </div>
              </div>

              <div
                class="text-center p-4 bg-purple-50 dark:bg-purple-950 rounded-lg"
              >
                <div
                  class="text-3xl font-bold text-purple-600 dark:text-purple-400"
                >
                  {{ record.type === 'invitation' ? '邀请记录' : '奖励记录' }}
                </div>
                <div class="text-sm text-purple-600 dark:text-purple-400 mt-1">
                  记录类型
                </div>
              </div>
            </div>
          </div>

          <!-- 详细信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              {{ record.type === 'invitation' ? '邀请详情' : '奖励详情' }}
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <div v-if="record.type === 'invitation'">
                <div class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  邀请人
                </div>
                <div class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <UAvatar
                    :src="record.inviter.avatar"
                    :alt="record.inviter.name"
                    size="md"
                    :ui="{ rounded: 'rounded-full' }"
                  />
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">
                      {{ record.inviter.name }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      {{ record.inviter.phone }}
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="record.type === 'invitation'">
                <div class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  被邀请人
                </div>
                <div class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <UAvatar
                    :src="record.invitee.avatar"
                    :alt="record.invitee.name"
                    size="md"
                    :ui="{ rounded: 'rounded-full' }"
                  />
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">
                      {{ record.invitee.name }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      {{ record.invitee.phone }}
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="record.type === 'reward'">
                <div class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  客户
                </div>
                <div class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <UAvatar
                    :src="record.customer.avatar"
                    :alt="record.customer.name"
                    size="md"
                    :ui="{ rounded: 'rounded-full' }"
                  />
                  <div>
                    <div class="font-medium text-gray-900 dark:text-white">
                      {{ record.customer.name }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      {{ record.customer.phone }}
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="record.type === 'reward'">
                <div class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  奖励金额
                </div>
                <div class="text-2xl font-bold text-gray-900 dark:text-white p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  ¥{{ record.amount.toLocaleString() }}
                </div>
              </div>

              <div>
                <div class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  {{ record.type === 'invitation' ? '邀请时间' : '创建时间' }}
                </div>
                <div class="text-gray-900 dark:text-white p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  {{ formatDate(record.createdAt) }}
                </div>
              </div>

              <div v-if="record.type === 'reward'">
                <div class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  发放时间
                </div>
                <div class="text-gray-900 dark:text-white p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  {{ record.paidAt ? formatDate(record.paidAt) : '未发放' }}
                </div>
              </div>

              <div v-if="record.type === 'reward'">
                <div class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  奖励类型
                </div>
                <div class="text-gray-900 dark:text-white p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  {{ record.rewardType }}
                </div>
              </div>

              <div v-if="record.type === 'reward'">
                <div class="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  关联记录
                </div>
                <div class="text-gray-900 dark:text-white p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  {{ record.relatedRecord }}
                </div>
              </div>
            </div>
          </div>

          <!-- 状态信息 -->
          <div>
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              状态信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
              <div>
                <span class="text-gray-500 dark:text-gray-400">记录ID：</span>
                <span class="text-gray-900 dark:text-white font-mono">{{ record.id }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">状态：</span>
                <UBadge
                  :color="record.status === 'active' || record.status === 'paid' ? 'success' : record.status === 'pending' ? 'warning' : 'error'"
                  variant="soft"
                >
                  {{ record.status === "active" ? "有效" : record.status === "paid" ? "已发放" : record.status === "pending" ? "待发放" : "已取消" }}
                </UBadge>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">创建时间：</span>
                <span class="text-gray-900 dark:text-white">{{ formatDate(record.createdAt) }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">更新时间：</span>
                <span class="text-gray-900 dark:text-white">{{ formatDate(record.updatedAt || record.createdAt) }}</span>
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
interface InvitationRecord {
  id: string;
  name: string;
  phone: string;
  avatar: string;
  type: 'invitation';
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
  createdAt: string;
  updatedAt?: string;
  status: 'active' | 'inactive';
}

interface RewardRecord {
  id: string;
  name: string;
  phone: string;
  avatar: string;
  type: 'reward';
  customer: {
    id: string;
    name: string;
    phone: string;
    avatar: string;
  };
  amount: number;
  rewardType: string;
  relatedRecord: string;
  createdAt: string;
  paidAt?: string;
  updatedAt?: string;
  status: 'paid' | 'pending' | 'cancelled';
}

type RecordDetail = InvitationRecord | RewardRecord;

const props = defineProps<{ open?: boolean; record: RecordDetail }>();
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

// 计算标题和描述
const modalTitle = computed(() => {
  return props.record.type === 'invitation' ? '邀请记录详情' : '奖励记录详情';
});

const modalDescription = computed(() => {
  return props.record.type === 'invitation' ? '查看邀请关系的完整信息' : '查看奖励发放的完整信息';
});

// 日期格式化
const formatDate = (iso: string) => {
  if (!iso) return 'N/A';
  return new Date(iso).toLocaleString("zh-CN");
};
</script>
