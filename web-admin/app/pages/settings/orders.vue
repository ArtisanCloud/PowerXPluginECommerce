<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">订单与流程配置</h1>
        <p class="text-gray-500 dark:text-gray-400">
          自定义订单生命周期、自动化任务与售后流程。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton variant="ghost" color="neutral" icon="i-heroicons-arrow-path">
          重置
        </UButton>
        <UButton color="primary" icon="i-heroicons-check">
          保存配置
        </UButton>
      </div>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">订单状态流</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                配置订单状态与可执行动作。
              </p>
            </div>
            <UButton size="sm" variant="soft" icon="i-heroicons-plus">
              新增状态
            </UButton>
          </div>
        </template>

        <ul class="space-y-3">
          <li
            v-for="state in orderStates"
            :key="state.name"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <div>
                <p class="font-medium text-gray-900 dark:text-white">
                  {{ state.name }}
                </p>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  {{ state.desc }}
                </p>
              </div>
              <UButton
                size="sm"
                variant="ghost"
                icon="i-heroicons-arrows-right-left"
                @click="toggleTransitions(state)"
              >
                编辑动作
              </UButton>
            </div>
            <div class="mt-2 flex flex-wrap gap-2 text-xs text-gray-500 dark:text-gray-400">
              <UBadge
                v-for="action in state.actions"
                :key="action"
                color="neutral"
                variant="subtle"
              >
                {{ action }}
              </UBadge>
            </div>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">自动化规则</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                依据条件触发通知或动作。
              </p>
            </div>
            <UButton size="sm" variant="soft" icon="i-heroicons-plus">
              添加规则
            </UButton>
          </div>
        </template>

        <UTable :columns="ruleColumns" :data="automationRules">
          <template #enabled-cell="{ row }">
            <UToggle v-model="row.original.enabled" />
          </template>
          <template #actions-cell>
            <div class="flex gap-2">
              <UButton size="xs" variant="ghost">编辑</UButton>
              <UButton size="xs" variant="ghost" color="error">删除</UButton>
            </div>
          </template>
        </UTable>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">售后 SLA</h3>
          <UButton size="sm" variant="ghost">导出配置</UButton>
        </div>
      </template>

      <UTable :columns="slaColumns" :data="slaPolicies">
        <template #responseHours-cell="{ getValue }">
          {{ getValue() }} 小时
        </template>
        <template #resolveHours-cell="{ getValue }">
          {{ getValue() }} 小时
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "settings-orders",
});

type OrderState = {
  name: string;
  desc: string;
  actions: string[];
};

const orderStates = ref<OrderState[]>([
  {
    name: "待付款",
    desc: "创建订单但未完成支付",
    actions: ["取消订单", "提醒支付"],
  },
  {
    name: "待发货",
    desc: "支付完成进入备货流程",
    actions: ["批量配货", "标记发货"],
  },
  {
    name: "运输中",
    desc: "已发出等待签收",
    actions: ["重新推单", "通知客户"],
  },
]);

const toggleTransitions = (state: OrderState) => {
  state.actions = [...state.actions]; // 占位：未来可弹窗
};

type AutomationRule = {
  name: string;
  condition: string;
  action: string;
  enabled: boolean;
};

const automationRules = ref<AutomationRule[]>([
  {
    name: "长时间未支付",
    condition: "订单创建后 24h 未付款",
    action: "提醒客户 + 自动取消",
    enabled: true,
  },
  {
    name: "高价值订单",
    condition: "订单金额 > ¥10000",
    action: "通知运营负责人",
    enabled: true,
  },
]);

const ruleColumns = computed<TableColumn<AutomationRule>[]>(() => [
  { accessorKey: "name", header: "规则" },
  { accessorKey: "condition", header: "触发条件" },
  { accessorKey: "action", header: "执行动作" },
  { accessorKey: "enabled", header: "启用" },
  { id: "actions", header: "操作" },
]);

type SlaPolicy = {
  type: string;
  responseHours: number;
  resolveHours: number;
  channels: string[];
};

const slaPolicies = ref<SlaPolicy[]>([
  {
    type: "退货审核",
    responseHours: 12,
    resolveHours: 48,
    channels: ["天猫旗舰店", "京东自营"],
  },
  {
    type: "换货处理",
    responseHours: 8,
    resolveHours: 36,
    channels: ["抖音旗舰店", "小红书旗舰店"],
  },
]);

const slaColumns = computed<TableColumn<SlaPolicy>[]>(() => [
  { accessorKey: "type", header: "场景" },
  { accessorKey: "responseHours", header: "响应时限" },
  { accessorKey: "resolveHours", header: "解决时限" },
  {
    accessorKey: "channels",
    header: "适用渠道",
    cell: ({ getValue }) => (getValue() as string[]).join(" / "),
  },
]);
</script>
