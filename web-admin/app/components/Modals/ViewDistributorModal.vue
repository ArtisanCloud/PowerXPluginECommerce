<template>
  <UModal
    v-model:open="isOpen"
    title="分销员详情"
    description="查看分销员的完整信息"
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
          <!-- 分销员基本信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              基本信息
            </h3>

            <div class="flex items-center gap-3 mb-6">
              <UAvatar
                :src="distributor?.avatar || ''"
                :alt="distributor?.name || ''"
                size="lg"
                :ui="{ rounded: 'rounded-full' }"
              />
              <div>
                <div class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ distributor?.name || '' }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400">
                  ID: {{ distributor?.id || '' }}
                </div>
                <UBadge
                  v-if="distributor"
                  :color="getStatusColor(distributor.status)"
                  variant="soft"
                  size="sm"
                  class="mt-1"
                >
                  {{ getStatusText(distributor.status) }}
                </UBadge>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span class="text-gray-500 dark:text-gray-400">手机号：</span>
                <span class="text-gray-900 dark:text-white">{{ distributor?.phone || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">邮箱：</span>
                <span class="text-gray-900 dark:text-white">{{ distributor?.email || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">身份证号：</span>
                <span class="text-gray-900 dark:text-white">{{ distributor?.idCard || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">微信号：</span>
                <span class="text-gray-900 dark:text-white">{{ distributor?.wechat || '' }}</span>
              </div>
              <div class="sm:col-span-2">
                <span class="text-gray-500 dark:text-gray-400">地址：</span>
                <span class="text-gray-900 dark:text-white">{{ distributor?.address || '' }}</span>
              </div>
            </div>
          </div>

          <!-- 推广信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              推广信息
            </h3>

            <div class="space-y-4">
              <div>
                <span class="text-gray-500 dark:text-gray-400">推广码：</span>
                <span class="text-gray-900 dark:text-white font-mono">{{ distributor?.promotionCode || '' }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">推广链接：</span>
                <div class="flex items-center gap-2 mt-1">
                  <span class="text-gray-900 dark:text-white break-all">{{ distributor?.promotionLink || '' }}</span>
                  <UButton
                    v-if="distributor?.promotionLink"
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-clipboard"
                    @click="copyToClipboard(distributor.promotionLink)"
                  >
                    复制
                  </UButton>
                </div>
              </div>
            </div>
          </div>

          <!-- 业绩统计 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              业绩统计
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">累计销售额</div>
                <div class="text-2xl font-bold text-gray-900 dark:text-white">
                  ¥{{ (distributor?.totalSales || 0).toLocaleString() }}
                </div>
              </div>
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">累计佣金</div>
                <div class="text-2xl font-bold text-green-600 dark:text-green-400">
                  ¥{{ (distributor?.totalCommission || 0).toLocaleString() }}
                </div>
              </div>
              <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <div class="text-sm text-gray-500 dark:text-gray-400">邀请人数</div>
                <div class="text-2xl font-bold text-blue-600 dark:text-blue-400">
                  {{ distributor?.invitedCount || 0 }}
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
          v-if="distributor?.status === 'pending'"
          color="success"
          @click="approveAndClose"
        >
          审核通过
        </UButton>
        <UButton
          v-else-if="distributor?.status === 'active'"
          color="warning"
          @click="disableAndClose"
        >
          禁用
        </UButton>
        <UButton
          v-else-if="distributor?.status === 'inactive'"
          color="success"
          @click="enableAndClose"
        >
          启用
        </UButton>
        <UButton variant="ghost" @click="close(false)">关闭</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
const props = defineProps<{ open?: boolean; distributor: any }>();
const emit = defineEmits<{
  "update:open": [boolean];
  close: [payload: any];
  approve: [distributor: any];
  disable: [distributor: any];
  enable: [distributor: any];
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
    case "active": return "success";
    case "pending": return "warning";
    case "inactive": return "error";
    default: return "neutral";
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case "active": return "已启用";
    case "pending": return "待审核";
    case "inactive": return "已禁用";
    default: return "未知状态";
  }
};

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    alert("已复制到剪贴板");
  });
};

// 审核通过并关闭
const approveAndClose = () => {
  if (props.distributor) {
    emit("approve", props.distributor);
    close(true);
  }
};

// 禁用并关闭
const disableAndClose = () => {
  if (props.distributor) {
    emit("disable", props.distributor);
    close(true);
  }
};

// 启用并关闭
const enableAndClose = () => {
  if (props.distributor) {
    emit("enable", props.distributor);
    close(true);
  }
};
</script>