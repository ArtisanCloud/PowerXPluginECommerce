<template>
  <UModal
    v-model:open="isOpen"
    title="佣金结算详情"
    description="查看佣金结算的详细信息"
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
          <!-- 分销员信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              分销员信息
            </h3>

            <div class="flex items-center gap-3">
              <UAvatar
                :src="commission?.distributor?.avatar || ''"
                :alt="commission?.distributor?.name || ''"
                size="lg"
                :ui="{ rounded: 'rounded-full' }"
              />
              <div>
                <div class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ commission?.distributor?.name || '' }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400">
                  ID: {{ commission?.distributor?.id || '' }}
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-4">
              <div>
                <span class="text-gray-500 dark:text-gray-400">手机号：</span>
                <span class="text-gray-900 dark:text-white">{{ commission?.distributor?.phone || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">邮箱：</span>
                <span class="text-gray-900 dark:text-white">{{ commission?.distributor?.email || '' }}</span>
              </div>
            </div>
          </div>

          <!-- 订单信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              订单信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span class="text-gray-500 dark:text-gray-400">订单编号：</span>
                <span class="text-gray-900 dark:text-white font-mono">{{ commission?.orderId || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">订单金额：</span>
                <span class="text-gray-900 dark:text-white">¥{{ (commission?.orderAmount || 0).toLocaleString() }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">下单时间：</span>
                <span class="text-gray-900 dark:text-white">{{ commission?.orderDate || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">佣金比例：</span>
                <span class="text-gray-900 dark:text-white">{{ commission?.commissionRate || 0 }}%</span>
              </div>
            </div>
          </div>

          <!-- 佣金信息 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              佣金信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">佣金金额</div>
                <div class="text-2xl font-bold text-green-600 dark:text-green-400">
                  ¥{{ (commission?.commissionAmount || 0).toLocaleString() }}
                </div>
              </div>
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">结算状态</div>
                <div class="mt-1">
                  <UBadge
                    v-if="commission"
                    :color="getStatusColor(commission.status)"
                    variant="soft"
                    size="sm"
                  >
                    {{ getStatusText(commission.status) }}
                  </UBadge>
                </div>
              </div>
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">结算时间</div>
                <div class="text-gray-900 dark:text-white">
                  {{ commission?.settledAt || '—' }}
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
        <UButton
          v-if="commission?.status === 'pending'"
          color="success"
          @click="settleAndClose"
        >
          确认结算
        </UButton>
        <UButton variant="ghost" @click="close(false)">关闭</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
const props = defineProps<{ open?: boolean; commission: any }>();
const emit = defineEmits<{
  "update:open": [boolean];
  close: [payload: any];
  settle: [commission: any];
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

// 状态相关的函数
const getStatusColor = (status: string) => {
  switch (status) {
    case "settled": return "success";
    case "pending": return "warning";
    case "withdrawable": return "blue";
    default: return "neutral";
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case "settled": return "已结算";
    case "pending": return "待结算";
    case "withdrawable": return "可提现";
    default: return "未知状态";
  }
};

// 结算并关闭
const settleAndClose = () => {
  if (props.commission) {
    emit("settle", props.commission);
    close(true);
  }
};
</script>