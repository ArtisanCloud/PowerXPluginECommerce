<template>
  <UModal
    v-model:open="isOpen"
    :title="isReviewMode ? '提现申请详情' : '提现记录详情'"
    :description="isReviewMode ? '审核分销员的提现申请' : '查看提现记录的详细信息'"
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
                :src="withdrawal?.distributor?.avatar || ''"
                :alt="withdrawal?.distributor?.name || ''"
                size="lg"
                :ui="{ rounded: 'rounded-full' }"
              />
              <div>
                <div class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ withdrawal?.distributor?.name || '' }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400">
                  ID: {{ withdrawal?.distributor?.id || '' }}
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-4">
              <div>
                <span class="text-gray-500 dark:text-gray-400">手机号：</span>
                <span class="text-gray-900 dark:text-white">{{ withdrawal?.distributor?.phone || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">邮箱：</span>
                <span class="text-gray-900 dark:text-white">{{ withdrawal?.distributor?.email || '' }}</span>
              </div>
            </div>
          </div>

          <!-- 提现信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              提现信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span class="text-gray-500 dark:text-gray-400">申请金额：</span>
                <span class="text-gray-900 dark:text-white font-bold text-lg">¥{{ (withdrawal?.amount || 0).toLocaleString() }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">申请时间：</span>
                <span class="text-gray-900 dark:text-white">{{ withdrawal?.appliedAt || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">银行账户：</span>
                <span class="text-gray-900 dark:text-white">{{ withdrawal?.bankAccount || '—' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">账户姓名：</span>
                <span class="text-gray-900 dark:text-white">{{ withdrawal?.accountName || '—' }}</span>
              </div>
            </div>
          </div>

          <!-- 财务信息 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              财务信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">累计佣金</div>
                <div class="text-2xl font-bold text-gray-900 dark:text-white">
                  ¥{{ (withdrawal?.totalCommission || 0).toLocaleString() }}
                </div>
              </div>
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">已提现金额</div>
                <div class="text-2xl font-bold text-gray-900 dark:text-white">
                  ¥{{ (withdrawal?.withdrawnAmount || 0).toLocaleString() }}
                </div>
              </div>
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">剩余可提现</div>
                <div class="text-2xl font-bold text-blue-600 dark:text-blue-400">
                  ¥{{ (withdrawal?.withdrawableAmount || 0).toLocaleString() }}
                </div>
              </div>
            </div>
          </div>

          <!-- 审核信息（仅在审核模式下显示） -->
          <div v-if="isReviewMode && withdrawal?.status === 'rejected'" class="border-t border-gray-200 dark:border-gray-700 pt-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              拒绝原因
            </h3>
            <div class="text-gray-900 dark:text-white bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
              {{ withdrawal?.rejectReason || '无拒绝原因' }}
            </div>
          </div>
        </div>
      </UCard>
    </template>

    <!-- Footer -->
    <template #footer>
      <div class="flex gap-3">
        <UButton
          v-if="isReviewMode && withdrawal?.status === 'pending'"
          color="success"
          @click="approveAndClose"
        >
          通过申请
        </UButton>
        <UButton
          v-if="isReviewMode && withdrawal?.status === 'pending'"
          color="red"
          @click="rejectAndClose"
        >
          拒绝申请
        </UButton>
        <UButton variant="ghost" @click="close(false)">关闭</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
const props = defineProps<{ open?: boolean; withdrawal: any; isReviewMode?: boolean }>();
const emit = defineEmits<{
  "update:open": [boolean];
  close: [payload: any];
  approve: [withdrawal: any];
  reject: [withdrawal: any];
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

// 通过并关闭
const approveAndClose = () => {
  if (props.withdrawal) {
    emit("approve", props.withdrawal);
    close(true);
  }
};

// 拒绝并关闭
const rejectAndClose = () => {
  if (props.withdrawal) {
    emit("reject", props.withdrawal);
    close(true);
  }
};
</script>