<template>
  <div class="p-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">规则配置</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">配置推荐/分销的奖励规则</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-plus" @click="createNewStrategy">新建策略</UButton>
      </div>
    </div>

    <!-- 策略列表 -->
    <UCard class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">分销策略</h2>
          <div class="flex gap-3">
            <UInput v-model="searchQuery" placeholder="搜索策略..." icon="i-heroicons-magnifying-glass" size="sm" />
            <UButton color="neutral" @click="resetFilters">重置</UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">策略名称</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">奖励对象</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">奖励内容</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">触发条件</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">分销比例</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="strategy in filteredStrategies" :key="strategy.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ strategy.name }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ strategy.rewardTarget?.name || '未设置' }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ strategy.rewardContent?.value || '未设置' }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ strategy.triggerCondition?.value || '未设置' }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ strategy.distributionRatio?.ratio ? strategy.distributionRatio.ratio + '%' : '未设置' }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="strategy.status === 'enabled' ? 'success' : 'neutral'" variant="soft">
                  {{ strategy.status === 'enabled' ? '启用' : '禁用' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-eye"
                    @click="viewStrategy(strategy)"
                  >
                    查看
                  </UButton>
                  <UButton
                    color="primary"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-pencil"
                    @click="editStrategy(strategy)"
                  >
                    编辑
                  </UButton>
                  <UButton
                    v-if="strategy.status === 'enabled'"
                    color="orange"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-x-mark"
                    @click="disableStrategy(strategy)"
                  >
                    禁用
                  </UButton>
                  <UButton
                    v-else
                    color="success"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-check"
                    @click="enableStrategy(strategy)"
                  >
                    启用
                  </UButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (currentPage - 1) * pageSize + 1 }}
            到 {{ Math.min(currentPage * pageSize, filteredStrategies.length) }} 条，共 {{ filteredStrategies.length }} 条
          </div>
          <UPagination
            v-model="currentPage"
            :page-count="pageCount"
            :total="filteredStrategies.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 策略配置模态框 -->
    <UModal
      v-model:open="showStrategyModal"
      :title="editingStrategy ? '编辑策略' : '新建策略'"
      description="配置分销策略的各个规则组件"
      :close="{ onClick: () => closeStrategyModal() }"
      :ui="{
        content: 'w-full sm:max-w-4xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-6 p-4 sm:p-6">
            <!-- 策略基本信息 -->
            <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                基本信息
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <UInput
                  v-model="currentStrategy.name"
                  label="策略名称"
                  placeholder="请输入策略名称"
                  :error="!!errors.name"
                />
                <USelect
                  v-model="currentStrategy.status"
                  label="状态"
                  :options="[{ label: '启用', value: 'enabled' }, { label: '禁用', value: 'disabled' }]"
                />
                <div class="sm:col-span-2">
                  <UTextarea
                    v-model="currentStrategy.description"
                    label="策略描述"
                    placeholder="请输入策略描述"
                  />
                </div>
              </div>
            </div>

            <!-- 规则配置 -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <!-- 奖励对象 -->
              <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div class="flex items-center justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">奖励对象</h4>
                  <UButton
                    color="primary"
                    variant="outline"
                    size="xs"
                    @click="openRuleModal('rewardTarget')"
                  >
                    选择
                  </UButton>
                </div>
                <div v-if="currentStrategy.rewardTarget" class="text-sm text-gray-700 dark:text-gray-300">
                  <div>{{ currentStrategy.rewardTarget.name }}</div>
                  <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">{{ currentStrategy.rewardTarget.description }}</div>
                </div>
                <div v-else class="text-sm text-gray-500 dark:text-gray-400">
                  未选择奖励对象
                </div>
              </div>

              <!-- 奖励内容 -->
              <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div class="flex items-center justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">奖励内容</h4>
                  <UButton
                    color="primary"
                    variant="outline"
                    size="xs"
                    @click="openRuleModal('rewardContent')"
                  >
                    选择
                  </UButton>
                </div>
                <div v-if="currentStrategy.rewardContent" class="text-sm text-gray-700 dark:text-gray-300">
                  <div>{{ currentStrategy.rewardContent.type }}: {{ currentStrategy.rewardContent.value }}</div>
                  <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">{{ currentStrategy.rewardContent.description }}</div>
                </div>
                <div v-else class="text-sm text-gray-500 dark:text-gray-400">
                  未选择奖励内容
                </div>
              </div>

              <!-- 触发条件 -->
              <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div class="flex items-center justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">触发条件</h4>
                  <UButton
                    color="primary"
                    variant="outline"
                    size="xs"
                    @click="openRuleModal('triggerCondition')"
                  >
                    选择
                  </UButton>
                </div>
                <div v-if="currentStrategy.triggerCondition" class="text-sm text-gray-700 dark:text-gray-300">
                  <div>{{ currentStrategy.triggerCondition.type }}: {{ currentStrategy.triggerCondition.value }}</div>
                  <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">{{ currentStrategy.triggerCondition.description }}</div>
                </div>
                <div v-else class="text-sm text-gray-500 dark:text-gray-400">
                  未选择触发条件
                </div>
              </div>

              <!-- 分销比例 -->
              <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div class="flex items-center justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">分销比例</h4>
                  <UButton
                    color="primary"
                    variant="outline"
                    size="xs"
                    @click="openRuleModal('distributionRatio')"
                  >
                    选择
                  </UButton>
                </div>
                <div v-if="currentStrategy.distributionRatio" class="text-sm text-gray-700 dark:text-gray-300">
                  <div>{{ currentStrategy.distributionRatio.level }}: {{ currentStrategy.distributionRatio.ratio }}%</div>
                  <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">{{ currentStrategy.distributionRatio.description }}</div>
                </div>
                <div v-else class="text-sm text-gray-500 dark:text-gray-400">
                  未选择分销比例
                </div>
              </div>

              <!-- 结算方式 -->
              <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div class="flex items-center justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">结算方式</h4>
                  <UButton
                    color="primary"
                    variant="outline"
                    size="xs"
                    @click="openRuleModal('settlementMethod')"
                  >
                    选择
                  </UButton>
                </div>
                <div v-if="currentStrategy.settlementMethod" class="text-sm text-gray-700 dark:text-gray-300">
                  <div>{{ currentStrategy.settlementMethod.name }} ({{ currentStrategy.settlementMethod.period }})</div>
                  <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">{{ currentStrategy.settlementMethod.description }}</div>
                </div>
                <div v-else class="text-sm text-gray-500 dark:text-gray-400">
                  未选择结算方式
                </div>
              </div>

              <!-- 提现规则 -->
              <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div class="flex items-center justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">提现规则</h4>
                  <UButton
                    color="primary"
                    variant="outline"
                    size="xs"
                    @click="openRuleModal('withdrawalRule')"
                  >
                    选择
                  </UButton>
                </div>
                <div v-if="currentStrategy.withdrawalRule" class="text-sm text-gray-700 dark:text-gray-300">
                  <div>{{ currentStrategy.withdrawalRule.name }}</div>
                  <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">最低¥{{ currentStrategy.withdrawalRule.minAmount }}, 手续费{{ currentStrategy.withdrawalRule.fee }}%</div>
                </div>
                <div v-else class="text-sm text-gray-500 dark:text-gray-400">
                  未选择提现规则
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeStrategyModal">取消</UButton>
          <UButton color="primary" @click="saveStrategy">保存策略</UButton>
        </div>
      </template>
    </UModal>

    <!-- 规则选择模态框 -->
    <UModal
      v-model:open="showRuleModal"
      :title="getRuleModalTitle()"
      :description="`选择或创建${getRuleModalTitle()}`"
      :close="{ onClick: () => closeRuleModal() }"
      :ui="{
        content: 'w-full sm:max-w-3xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="p-4 sm:p-6">
            <div class="flex justify-between items-center mb-4">
              <UInput
                v-model="ruleSearchQuery"
                placeholder="搜索规则..."
                icon="i-heroicons-magnifying-glass"
                size="sm"
              />
              <UButton
                color="primary"
                variant="outline"
                size="sm"
                icon="i-heroicons-plus"
                @click="createNewRule"
              >
                新建
              </UButton>
            </div>

            <div class="overflow-x-auto max-h-96">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead class="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th scope="col" class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider"></th>
                    <th scope="col" class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">名称/值</th>
                    <th scope="col" class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">描述</th>
                    <th scope="col" class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">状态</th>
                  </tr>
                </thead>
                <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  <tr 
                    v-for="rule in filteredRules" 
                    :key="rule.id" 
                    class="cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700"
                    @click="selectRule(rule)"
                  >
                    <td class="px-4 py-2 whitespace-nowrap">
                      <URadio v-model="selectedRuleId" :value="rule.id" />
                    </td>
                    <td class="px-4 py-2 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                      {{ getRuleDisplayValue(rule) }}
                    </td>
                    <td class="px-4 py-2 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {{ rule.description || '无描述' }}
                    </td>
                    <td class="px-4 py-2 whitespace-nowrap">
                      <UBadge :color="rule.status === 'enabled' ? 'success' : 'neutral'" variant="soft" size="xs">
                        {{ rule.status === 'enabled' ? '启用' : '禁用' }}
                      </UBadge>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeRuleModal">取消</UButton>
          <UButton color="primary" @click="confirmRuleSelection">确定</UButton>
        </div>
      </template>
    </UModal>

    <!-- 规则编辑模态框 -->
    <UModal
      v-model:open="showRuleEditModal"
      :title="editingRule ? '编辑规则' : '新建规则'"
      :description="getRuleEditDescription()"
      :close="{ onClick: () => closeRuleEditModal() }"
      :ui="{
        content: 'w-full sm:max-w-2xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-4 p-4 sm:p-6">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <UInput
                v-model="currentRule.name"
                label="名称"
                placeholder="请输入名称"
                :error="!!ruleErrors.name"
              />
              <UInput
                v-model="currentRule.value"
                label="值"
                placeholder="请输入值"
                :error="!!ruleErrors.value"
              />
              <USelect
                v-model="currentRule.status"
                label="状态"
                :options="[{ label: '启用', value: 'enabled' }, { label: '禁用', value: 'disabled' }]"
              />
              <div class="sm:col-span-2">
                <UTextarea
                  v-model="currentRule.description"
                  label="描述"
                  placeholder="请输入描述"
                  rows="3"
                />
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeRuleEditModal">取消</UButton>
          <UButton color="primary" @click="saveRule">保存</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
// 分页
const currentPage = ref(1);
const pageSize = ref(10);
const searchQuery = ref("");

// 模态框
const showStrategyModal = ref(false);
const showRuleModal = ref(false);
const showRuleEditModal = ref(false);

// 当前编辑的数据
const editingStrategy = ref(false);
const currentStrategy = ref({
  id: "",
  name: "",
  description: "",
  status: "enabled",
  rewardTarget: null as any,
  rewardContent: null as any,
  triggerCondition: null as any,
  distributionRatio: null as any,
  settlementMethod: null as any,
  withdrawalRule: null as any
});

// 当前编辑的规则
const editingRule = ref(false);
const currentRule = ref({
  id: "",
  type: "",
  name: "",
  value: "",
  status: "enabled",
  description: ""
});

// 当前选择的规则类型
const currentRuleType = ref("");

// 当前选中的规则ID
const selectedRuleId = ref("");

// 搜索查询
const ruleSearchQuery = ref("");

// 错误信息
const errors = ref<Record<string, string>>({});
const ruleErrors = ref<Record<string, string>>({});

// 模拟数据
const strategies = ref([
  {
    id: "1",
    name: "基础分销策略",
    description: "适用于所有用户的分销策略",
    status: "enabled",
    rewardTarget: { id: "1", type: "用户", name: "所有用户", description: "所有注册用户都可以参与", status: "enabled" },
    rewardContent: { id: "2", type: "现金", value: "10.00", description: "奖励10元现金", status: "enabled" },
    triggerCondition: { id: "1", type: "注册", value: "完成注册", description: "用户完成注册后触发", status: "enabled" },
    distributionRatio: { id: "1", level: "一级分销", ratio: "10", description: "一级分销员获得10%佣金", status: "enabled" },
    settlementMethod: { id: "1", name: "实时结算", period: "实时", description: "订单完成后立即结算", status: "enabled" },
    withdrawalRule: { id: "1", name: "最低提现", minAmount: "100.00", fee: "0", description: "最低提现100元，无手续费", status: "enabled" }
  }
]);

const rewardTargets = ref([
  { id: "1", type: "用户", name: "所有用户", description: "所有注册用户都可以参与", status: "enabled" },
  { id: "2", type: "分销员", name: "一级分销员", description: "一级分销员可获得奖励", status: "enabled" },
  { id: "3", type: "分销员", name: "二级分销员", description: "二级分销员可获得奖励", status: "enabled" },
]);

const rewardContents = ref([
  { id: "1", type: "积分", value: "100", description: "奖励100积分", status: "enabled" },
  { id: "2", type: "现金", value: "10.00", description: "奖励10元现金", status: "enabled" },
  { id: "3", type: "优惠券", value: "20.00", description: "奖励20元优惠券", status: "enabled" },
]);

const triggerConditions = ref([
  { id: "1", type: "注册", value: "完成注册", description: "用户完成注册后触发", status: "enabled" },
  { id: "2", type: "购买", value: "首次购买", description: "用户首次购买后触发", status: "enabled" },
  { id: "3", type: "购买", value: "满100元", description: "用户单笔订单满100元触发", status: "enabled" },
]);

const distributionRatios = ref([
  { id: "1", level: "一级分销", ratio: "10", description: "一级分销员获得10%佣金", status: "enabled" },
  { id: "2", level: "二级分销", ratio: "5", description: "二级分销员获得5%佣金", status: "enabled" },
]);

const settlementMethods = ref([
  { id: "1", name: "实时结算", period: "实时", description: "订单完成后立即结算", status: "enabled" },
  { id: "2", name: "日结", period: "每日", description: "每日结算前一日佣金", status: "enabled" },
  { id: "3", name: "周结", period: "每周", description: "每周一结算上周佣金", status: "disabled" },
]);

const withdrawalRules = ref([
  { id: "1", name: "最低提现", minAmount: "100.00", fee: "0", description: "最低提现100元，无手续费", status: "enabled" },
  { id: "2", name: "手续费规则", minAmount: "50.00", fee: "1", description: "最低提现50元，1%手续费", status: "enabled" },
]);

// 过滤
const filteredStrategies = computed(() => {
  if (!searchQuery.value) return strategies.value;
  const q = searchQuery.value.toLowerCase();
  return strategies.value.filter(strategy =>
    strategy.name.toLowerCase().includes(q) ||
    strategy.description.toLowerCase().includes(q)
  );
});

const filteredRules = computed(() => {
  const allRules = getRulesByType(currentRuleType.value);
  if (!ruleSearchQuery.value) return allRules;
  const q = ruleSearchQuery.value.toLowerCase();
  return allRules.filter((rule: any) =>
    rule.name?.toLowerCase().includes(q) ||
    rule.value?.toLowerCase().includes(q) ||
    rule.description?.toLowerCase().includes(q)
  );
});

// 页数
const pageCount = computed(() =>
  Math.max(1, Math.ceil(filteredStrategies.value.length / pageSize.value))
);

// 重置筛选
const resetFilters = () => {
  searchQuery.value = "";
  currentPage.value = 1;
};

// 根据类型获取规则列表
const getRulesByType = (type: string) => {
  switch (type) {
    case "rewardTarget": return rewardTargets.value;
    case "rewardContent": return rewardContents.value;
    case "triggerCondition": return triggerConditions.value;
    case "distributionRatio": return distributionRatios.value;
    case "settlementMethod": return settlementMethods.value;
    case "withdrawalRule": return withdrawalRules.value;
    default: return [];
  }
};

// 获取规则显示值
const getRuleDisplayValue = (rule: any) => {
  if (rule.level && rule.ratio) return `${rule.level}: ${rule.ratio}%`;
  if (rule.name && rule.period) return `${rule.name} (${rule.period})`;
  if (rule.name && rule.minAmount) return `${rule.name} (¥${rule.minAmount})`;
  if (rule.type && rule.value) return `${rule.type}: ${rule.value}`;
  return rule.name || rule.value || "未命名";
};

// 获取规则模态框标题
const getRuleModalTitle = () => {
  const titles: Record<string, string> = {
    rewardTarget: "奖励对象",
    rewardContent: "奖励内容",
    triggerCondition: "触发条件",
    distributionRatio: "分销比例",
    settlementMethod: "结算方式",
    withdrawalRule: "提现规则"
  };
  return titles[currentRuleType.value] || "规则";
};

// 获取规则编辑描述
const getRuleEditDescription = () => {
  const descriptions: Record<string, string> = {
    rewardTarget: "配置奖励对象规则",
    rewardContent: "配置奖励内容规则",
    triggerCondition: "配置触发条件规则",
    distributionRatio: "配置分销比例规则",
    settlementMethod: "配置结算方式规则",
    withdrawalRule: "配置提现规则"
  };
  return descriptions[currentRuleType.value] || "配置规则";
};

// 策略操作
const createNewStrategy = () => {
  editingStrategy.value = false;
  currentStrategy.value = {
    id: "",
    name: "",
    description: "",
    status: "enabled",
    rewardTarget: null,
    rewardContent: null,
    triggerCondition: null,
    distributionRatio: null,
    settlementMethod: null,
    withdrawalRule: null
  };
  showStrategyModal.value = true;
};

const editStrategy = (strategy: any) => {
  editingStrategy.value = true;
  currentStrategy.value = { ...strategy };
  showStrategyModal.value = true;
};

const viewStrategy = (strategy: any) => {
  alert(`查看策略: ${strategy.name}`);
};

const enableStrategy = (strategy: any) => {
  strategy.status = "enabled";
  alert(`已启用策略: ${strategy.name}`);
};

const disableStrategy = (strategy: any) => {
  strategy.status = "disabled";
  alert(`已禁用策略: ${strategy.name}`);
};

const closeStrategyModal = () => {
  showStrategyModal.value = false;
  currentStrategy.value = {
    id: "",
    name: "",
    description: "",
    status: "enabled",
    rewardTarget: null,
    rewardContent: null,
    triggerCondition: null,
    distributionRatio: null,
    settlementMethod: null,
    withdrawalRule: null
  };
  errors.value = {};
};

const saveStrategy = () => {
  if (!currentStrategy.value.name.trim()) {
    errors.value.name = "策略名称不能为空";
    return;
  }

  if (editingStrategy.value) {
    alert(`已更新策略: ${currentStrategy.value.name}`);
  } else {
    alert(`已创建新策略: ${currentStrategy.value.name}`);
  }

  closeStrategyModal();
};

// 规则操作
const openRuleModal = (ruleType: string) => {
  currentRuleType.value = ruleType;
  selectedRuleId.value = getCurrentRuleId(ruleType);
  ruleSearchQuery.value = "";
  showRuleModal.value = true;
};

const closeRuleModal = () => {
  showRuleModal.value = false;
  currentRuleType.value = "";
  selectedRuleId.value = "";
};

const getCurrentRuleId = (ruleType: string) => {
  const strategy = currentStrategy.value;
  switch (ruleType) {
    case "rewardTarget": return strategy.rewardTarget?.id || "";
    case "rewardContent": return strategy.rewardContent?.id || "";
    case "triggerCondition": return strategy.triggerCondition?.id || "";
    case "distributionRatio": return strategy.distributionRatio?.id || "";
    case "settlementMethod": return strategy.settlementMethod?.id || "";
    case "withdrawalRule": return strategy.withdrawalRule?.id || "";
    default: return "";
  }
};

const selectRule = (rule: any) => {
  selectedRuleId.value = rule.id;
};

const confirmRuleSelection = () => {
  const rules = getRulesByType(currentRuleType.value);
  const selectedRule = rules.find((rule: any) => rule.id === selectedRuleId.value);
  
  if (selectedRule) {
    switch (currentRuleType.value) {
      case "rewardTarget":
        currentStrategy.value.rewardTarget = selectedRule;
        break;
      case "rewardContent":
        currentStrategy.value.rewardContent = selectedRule;
        break;
      case "triggerCondition":
        currentStrategy.value.triggerCondition = selectedRule;
        break;
      case "distributionRatio":
        currentStrategy.value.distributionRatio = selectedRule;
        break;
      case "settlementMethod":
        currentStrategy.value.settlementMethod = selectedRule;
        break;
      case "withdrawalRule":
        currentStrategy.value.withdrawalRule = selectedRule;
        break;
    }
  }
  
  closeRuleModal();
};

const createNewRule = () => {
  editingRule.value = false;
  currentRule.value = {
    id: "",
    type: currentRuleType.value,
    name: "",
    value: "",
    status: "enabled",
    description: ""
  };
  ruleErrors.value = {};
  showRuleEditModal.value = true;
};

const closeRuleEditModal = () => {
  showRuleEditModal.value = false;
  currentRule.value = {
    id: "",
    type: "",
    name: "",
    value: "",
    status: "enabled",
    description: ""
  };
  ruleErrors.value = {};
};

const saveRule = () => {
  if (!currentRule.value.name.trim()) {
    ruleErrors.value.name = "名称不能为空";
    return;
  }

  if (!currentRule.value.value.trim()) {
    ruleErrors.value.value = "值不能为空";
    return;
  }

  // 添加或更新规则到对应列表
  const rulesList = getRulesListByType(currentRuleType.value);
  if (editingRule.value) {
    const index = rulesList.findIndex((r: any) => r.id === currentRule.value.id);
    if (index !== -1) rulesList[index] = { ...currentRule.value };
  } else {
    currentRule.value.id = `rule_${Date.now()}`;
    rulesList.push({ ...currentRule.value });
  }

  alert(`${editingRule.value ? '更新' : '创建'}规则成功`);
  closeRuleEditModal();
};

const getRulesListByType = (type: string) => {
  switch (type) {
    case "rewardTarget": return rewardTargets.value;
    case "rewardContent": return rewardContents.value;
    case "triggerCondition": return triggerConditions.value;
    case "distributionRatio": return distributionRatios.value;
    case "settlementMethod": return settlementMethods.value;
    case "withdrawalRule": return withdrawalRules.value;
    default: return [];
  }
};
</script>