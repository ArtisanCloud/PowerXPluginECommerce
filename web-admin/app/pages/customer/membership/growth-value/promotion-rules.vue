<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          等级晋升逻辑
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          配置会员等级晋升规则和条件
        </p>
      </div>
      <div class="flex gap-3">
        <UDropdown :items="exportItems" :popper="{ placement: 'bottom-start' }">
          <UButton
            color="neutral"
            variant="outline"
            icon="i-heroicons-arrow-down-tray"
          >
            导出
          </UButton>
        </UDropdown>
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="openCreateRuleModal"
        >
          创建晋升规则
        </UButton>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon
              name="i-heroicons-cog-6-tooth"
              class="w-6 h-6 text-blue-600 dark:text-blue-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              总规则数
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ totalRules }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon
              name="i-heroicons-arrow-trending-up"
              class="w-6 h-6 text-green-600 dark:text-green-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              自动晋升规则
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ autoPromotionRules.length }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon
              name="i-heroicons-user-group"
              class="w-6 h-6 text-purple-600 dark:text-purple-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              手动晋升规则
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ manualPromotionRules.length }}
            </p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 规则分类标签 -->
    <div class="mb-6">
      <UTabs
        v-model="activeTab"
        :items="tabs"
        @update:model-value="onTabChange"
      />
    </div>

    <!-- 规则列表 -->
    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ activeTab === 'all' ? '所有规则' : activeTab === 'auto' ? '自动晋升规则' : '手动晋升规则' }}
          </h2>
          <div class="flex gap-2">
            <UInput
              v-model="searchQuery"
              placeholder="搜索规则名称..."
              icon="i-heroicons-magnifying-glass"
              size="sm"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredRules">
        <template #status-cell="{ row }">
          <UBadge
            :color="row.status === 'active' ? 'success' : 'neutral'"
            variant="soft"
          >
            {{ row.status === "active" ? "启用" : "停用" }}
          </UBadge>
        </template>

        <template #type-cell="{ row }">
          <div class="font-medium text-gray-900 dark:text-white">
            {{ getRuleTypeLabel(row.type) }}
          </div>
        </template>

        <template #targetLevel-cell="{ row }">
          <div class="flex items-center gap-2">
            <div
              class="w-3 h-3 rounded-full"
              :style="{ backgroundColor: row.targetLevelColor }"
            ></div>
            <span>{{ row.targetLevelName }}</span>
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewRule(row)"
            >
              查看
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
              @click="editRule(row)"
            >
              编辑
            </UButton>
            <UButton
              :color="row.status === 'active' ? 'warning' : 'success'"
              variant="ghost"
              size="sm"
              :icon="row.status === 'active' ? 'i-heroicons-pause' : 'i-heroicons-play'"
              @click="toggleRuleStatus(row)"
            >
              {{ row.status === "active" ? "停用" : "启用" }}
            </UButton>
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              @click="deleteRule(row)"
            >
              删除
            </UButton>
          </div>
        </template>
      </UTable>

      <!-- 分页 -->
      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (currentPage - 1) * pageSize + 1 }} 到 {{ Math.min(currentPage * pageSize, filteredRules.length) }} 条，共 {{ filteredRules.length }} 条
          </div>
          <UPagination
            v-model="currentPage"
            :page-count="pageCount"
            :total="filteredRules.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 查看规则详情模态框 -->
    <UModal
      v-model:open="showViewRuleModal"
      title="晋升规则详情"
      description="查看会员等级晋升规则详细信息"
      :ui="{
        content: 'w-full sm:max-w-2xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-6 p-4 sm:p-6">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                  规则名称
                </label>
                <div class="mt-1 text-gray-900 dark:text-white">
                  {{ viewingRule.name }}
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                  晋升类型
                </label>
                <div class="mt-1 text-gray-900 dark:text-white">
                  {{ getRuleTypeLabel(viewingRule.type) }}
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                  目标等级
                </label>
                <div class="mt-1 flex items-center gap-2">
                  <div
                    class="w-3 h-3 rounded-full"
                    :style="{ backgroundColor: viewingRule.targetLevelColor }"
                  ></div>
                  <span class="text-gray-900 dark:text-white">{{ viewingRule.targetLevelName }}</span>
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                  状态
                </label>
                <div class="mt-1">
                  <UBadge
                    :color="viewingRule.status === 'active' ? 'success' : 'neutral'"
                    variant="soft"
                  >
                    {{ viewingRule.status === "active" ? "启用" : "停用" }}
                  </UBadge>
                </div>
              </div>

              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                  规则描述
                </label>
                <div class="mt-1 text-gray-900 dark:text-white">
                  {{ viewingRule.description || '无描述' }}
                </div>
              </div>
            </div>

            <div class="border-t border-gray-200 dark:border-gray-800 pt-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                {{ viewingRule.type === 'auto' ? '自动晋升条件' : '手动晋升条件' }}
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4" v-if="viewingRule.type === 'auto'">
                <div>
                  <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                    成长值要求
                  </label>
                  <div class="mt-1 text-gray-900 dark:text-white">
                    {{ viewingRule.growthValueThreshold || 0 }}
                  </div>
                </div>

                <div>
                  <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                    消费金额要求
                  </label>
                  <div class="mt-1 text-gray-900 dark:text-white">
                    {{ viewingRule.spendThreshold ? `¥${viewingRule.spendThreshold}` : '无要求' }}
                  </div>
                </div>

                <div>
                  <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                    购买次数要求
                  </label>
                  <div class="mt-1 text-gray-900 dark:text-white">
                    {{ viewingRule.purchaseCountThreshold || 0 }}
                  </div>
                </div>

                <div>
                  <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                    会员时长要求（月）
                  </label>
                  <div class="mt-1 text-gray-900 dark:text-white">
                    {{ viewingRule.membershipDuration || 0 }}
                  </div>
                </div>
              </div>

              <div class="space-y-4" v-else>
                <div>
                  <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                    审批流程
                  </label>
                  <div class="mt-1 text-gray-900 dark:text-white">
                    {{ getApprovalProcessLabel(viewingRule.approvalProcess) }}
                  </div>
                </div>

                <div v-if="viewingRule.approvalProcess !== 'none'">
                  <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                    审批人
                  </label>
                  <div class="mt-1 text-gray-900 dark:text-white">
                    {{ viewingRule.approvers || '未指定' }}
                  </div>
                </div>
              </div>
            </div>

            <div class="border-t border-gray-200 dark:border-gray-800 pt-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                晋升奖励
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                    奖励积分
                  </label>
                  <div class="mt-1 text-gray-900 dark:text-white">
                    {{ viewingRule.rewardPoints || 0 }}
                  </div>
                </div>

                <div>
                  <label class="block text-sm font-medium text-gray-500 dark:text-gray-400">
                    奖励优惠券
                  </label>
                  <div class="mt-1 text-gray-900 dark:text-white">
                    {{ viewingRule.rewardCoupon || '无' }}
                  </div>
                </div>
              </div>
            </div>

            <!-- 关联成长值账户统计 -->
            <div class="border-t border-gray-200 dark:border-gray-800 pt-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                关联成长值账户统计
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div class="bg-blue-50 dark:bg-blue-900/30 rounded-lg p-4">
                  <div class="text-sm text-blue-700 dark:text-blue-300">符合条件账户数</div>
                  <div class="text-2xl font-bold text-blue-900 dark:text-blue-100 mt-1">
                    {{ qualifiedAccountsCount }}
                  </div>
                </div>

                <div class="bg-green-50 dark:bg-green-900/30 rounded-lg p-4">
                  <div class="text-sm text-green-700 dark:text-green-300">已晋升账户数</div>
                  <div class="text-2xl font-bold text-green-900 dark:text-green-100 mt-1">
                    {{ promotedAccountsCount }}
                  </div>
                </div>

                <div class="bg-purple-50 dark:bg-purple-900/30 rounded-lg p-4">
                  <div class="text-sm text-purple-700 dark:text-purple-300">晋升率</div>
                  <div class="text-2xl font-bold text-purple-900 dark:text-purple-100 mt-1">
                    {{ promotionRate }}%
                  </div>
                </div>
              </div>

              <div class="mt-4">
                <UButton
                  to="/customer/membership/growth-value"
                  variant="outline"
                  size="sm"
                  trailing-icon="i-heroicons-arrow-right"
                >
                  查看关联账户
                </UButton>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <UButton @click="showViewRuleModal = false">关闭</UButton>
      </template>
    </UModal>

    <!-- 创建/编辑规则模态框 -->
    <UModal
      v-model:open="showRuleModal"
      :title="editingRule ? '编辑晋升规则' : '创建晋升规则'"
      :description="editingRule ? '修改会员等级晋升规则' : '创建新的会员等级晋升规则'"
      :ui="{
        content: 'w-full sm:max-w-3xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-6 p-4 sm:p-6">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  规则名称 <span class="text-red-500">*</span>
                </label>
                <UInput
                  v-model="ruleForm.name"
                  placeholder="请输入规则名称"
                />
              </div>

              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  晋升类型 <span class="text-red-500">*</span>
                </label>
                <USelect
                  v-model="ruleForm.type"
                  :options="ruleTypeOptions"
                  option-attribute="label"
                  value-attribute="value"
                />
              </div>

              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  目标等级 <span class="text-red-500">*</span>
                </label>
                <USelect
                  v-model="ruleForm.targetLevel"
                  :options="membershipLevels"
                  option-attribute="name"
                  value-attribute="id"
                />
              </div>

              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  状态
                </label>
                <USelect
                  v-model="ruleForm.status"
                  :options="statusOptions"
                  option-attribute="label"
                  value-attribute="value"
                />
              </div>

              <div class="sm:col-span-2">
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  规则描述
                </label>
                <UTextarea
                  v-model="ruleForm.description"
                  placeholder="请输入规则描述"
                  rows="3"
                />
              </div>

              <!-- 自动晋升规则配置 -->
              <div class="sm:col-span-2" v-if="ruleForm.type === 'auto'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  自动晋升条件
                </h3>
                <div class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      成长值要求
                    </label>
                    <UInput
                      v-model.number="ruleForm.growthValueThreshold"
                      type="number"
                      placeholder="成长值要求"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      消费金额要求
                    </label>
                    <UInput
                      v-model.number="ruleForm.spendThreshold"
                      type="number"
                      placeholder="消费金额要求"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      购买次数要求
                    </label>
                    <UInput
                      v-model.number="ruleForm.purchaseCountThreshold"
                      type="number"
                      placeholder="购买次数要求"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      会员时长要求（月）
                    </label>
                    <UInput
                      v-model.number="ruleForm.membershipDuration"
                      type="number"
                      placeholder="会员时长要求"
                    />
                  </div>
                </div>
              </div>

              <!-- 手动晋升规则配置 -->
              <div class="sm:col-span-2" v-if="ruleForm.type === 'manual'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  手动晋升条件
                </h3>
                <div class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      审批流程
                    </label>
                    <USelect
                      v-model="ruleForm.approvalProcess"
                      :options="[
                        { label: '无需审批', value: 'none' },
                        { label: '一级审批', value: 'level1' },
                        { label: '二级审批', value: 'level2' },
                      ]"
                      option-attribute="label"
                      value-attribute="value"
                    />
                  </div>
                  <div v-if="ruleForm.approvalProcess !== 'none'">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      审批人
                    </label>
                    <UInput
                      v-model="ruleForm.approvers"
                      placeholder="请输入审批人（多个用逗号分隔）"
                    />
                  </div>
                </div>
              </div>

              <div class="sm:col-span-2">
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  晋升奖励
                </label>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      奖励积分
                    </label>
                    <UInput
                      v-model.number="ruleForm.rewardPoints"
                      type="number"
                      placeholder="奖励积分"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      奖励优惠券
                    </label>
                    <UInput
                      v-model="ruleForm.rewardCoupon"
                      placeholder="奖励优惠券ID"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeRuleModal">取消</UButton>
          <UButton
            color="primary"
            :loading="ruleLoading"
            :disabled="!isRuleFormValid"
            @click="saveRule"
          >
            保存规则
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
// 模态框状态
const showRuleModal = ref(false);
const showViewRuleModal = ref(false);

// 当前编辑的规则
const editingRule = ref<any>(null);

// 当前查看的规则
const viewingRule = ref<any>({});

// 关联账户统计
const qualifiedAccountsCount = computed(() => {
  // 模拟符合条件的账户数
  return viewingRule.value.growthValueThreshold ?
    Math.floor(10000 / viewingRule.value.growthValueThreshold) : 0;
});

const promotedAccountsCount = computed(() => {
  // 模拟已晋升账户数
  return Math.floor(qualifiedAccountsCount.value * 0.7);
});

const promotionRate = computed(() => {
  // 计算晋升率
  if (qualifiedAccountsCount.value === 0) return 0;
  return Math.round((promotedAccountsCount.value / qualifiedAccountsCount.value) * 100);
});

// 活动标签页
const activeTab = ref("all");

// 搜索
const searchQuery = ref("");

// 导出选项
const exportItems = [
  [
    {
      label: "导出为CSV",
      icon: "i-heroicons-table-cells",
      click: () => exportRules("csv")
    },
    {
      label: "导出为Excel",
      icon: "i-heroicons-document-chart-bar",
      click: () => exportRules("excel")
    }
  ]
];

// 分页
const currentPage = ref(1);
const pageSize = ref(10);
const pageCount = computed(() => Math.ceil(filteredRules.value.length / pageSize.value));

// 标签页配置
const tabs = [
  { label: "所有规则", value: "all" },
  { label: "自动晋升规则", value: "auto" },
  { label: "手动晋升规则", value: "manual" },
];

// 规则类型选项
const ruleTypeOptions = [
  { label: "自动晋升", value: "auto" },
  { label: "手动晋升", value: "manual" },
];

// 状态选项
const statusOptions = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

// 会员等级选项
const membershipLevels = ref([
  { id: "level_1", name: "青铜会员", color: "#CD7F32" },
  { id: "level_2", name: "白银会员", color: "#C0C0C0" },
  { id: "level_3", name: "黄金会员", color: "#FFD700" },
  { id: "level_4", name: "铂金会员", color: "#E5E4E2" },
  { id: "level_5", name: "钻石会员", color: "#B9F2FF" },
  { id: "level_6", name: "黑金会员", color: "#000000" },
]);

// 规则表单
const ruleForm = ref({
  name: "",
  type: "auto",
  targetLevel: "level_1",
  status: "active",
  description: "",
  // 自动晋升条件
  growthValueThreshold: 0,
  spendThreshold: 0,
  purchaseCountThreshold: 0,
  membershipDuration: 0,
  // 手动晋升条件
  approvalProcess: "none",
  approvers: "",
  // 晋升奖励
  rewardPoints: 0,
  rewardCoupon: "",
});

// 加载状态
const ruleLoading = ref(false);

// 总规则数
const totalRules = ref(12);

// 自动晋升规则
const autoPromotionRules = ref([
  {
    id: 1,
    name: "青铜到白银自动晋升",
    type: "auto",
    targetLevel: "level_2",
    targetLevelName: "白银会员",
    targetLevelColor: "#C0C0C0",
    status: "active",
    description: "满足条件自动晋升到白银会员",
    growthValueThreshold: 500,
    spendThreshold: 1000,
    purchaseCountThreshold: 5,
    membershipDuration: 1
  },
  {
    id: 2,
    name: "白银到黄金自动晋升",
    type: "auto",
    targetLevel: "level_3",
    targetLevelName: "黄金会员",
    targetLevelColor: "#FFD700",
    status: "active",
    description: "满足条件自动晋升到黄金会员",
    growthValueThreshold: 2000,
    spendThreshold: 5000,
    purchaseCountThreshold: 10,
    membershipDuration: 3
  },
  {
    id: 3,
    name: "黄金到铂金自动晋升",
    type: "auto",
    targetLevel: "level_4",
    targetLevelName: "铂金会员",
    targetLevelColor: "#E5E4E2",
    status: "active",
    description: "满足条件自动晋升到铂金会员",
    growthValueThreshold: 5000,
    spendThreshold: 15000,
    purchaseCountThreshold: 20,
    membershipDuration: 6
  },
  {
    id: 4,
    name: "铂金到钻石自动晋升",
    type: "auto",
    targetLevel: "level_5",
    targetLevelName: "钻石会员",
    targetLevelColor: "#B9F2FF",
    status: "active",
    description: "满足条件自动晋升到钻石会员",
    growthValueThreshold: 15000,
    spendThreshold: 50000,
    purchaseCountThreshold: 50,
    membershipDuration: 12
  },
  {
    id: 9,
    name: "新用户快速晋升",
    type: "auto",
    targetLevel: "level_2",
    targetLevelName: "白银会员",
    targetLevelColor: "#C0C0C0",
    status: "active",
    description: "新用户首次购物即可晋升",
    growthValueThreshold: 100,
    spendThreshold: 200,
    purchaseCountThreshold: 1,
    membershipDuration: 0
  },
  {
    id: 10,
    name: "活跃用户加速晋升",
    type: "auto",
    targetLevel: "level_3",
    targetLevelName: "黄金会员",
    targetLevelColor: "#FFD700",
    status: "active",
    description: "月活跃用户加速晋升规则",
    growthValueThreshold: 1500,
    spendThreshold: 3000,
    purchaseCountThreshold: 8,
    membershipDuration: 1
  },
]);

// 手动晋升规则
const manualPromotionRules = ref([
  {
    id: 5,
    name: "特殊贡献晋升",
    type: "manual",
    targetLevel: "level_3",
    targetLevelName: "黄金会员",
    targetLevelColor: "#FFD700",
    status: "active",
    description: "特殊贡献客户手动晋升",
    approvalProcess: "level2",
    approvers: "manager1,manager2"
  },
  {
    id: 6,
    name: "VIP客户晋升",
    type: "manual",
    targetLevel: "level_4",
    targetLevelName: "铂金会员",
    targetLevelColor: "#E5E4E2",
    status: "active",
    description: "VIP客户手动晋升",
    approvalProcess: "level1",
    approvers: "vip_manager"
  },
  {
    id: 7,
    name: "合作伙伴晋升",
    type: "manual",
    targetLevel: "level_5",
    targetLevelName: "钻石会员",
    targetLevelColor: "#B9F2FF",
    status: "inactive",
    description: "合作伙伴手动晋升",
    approvalProcess: "level2",
    approvers: "partner_manager,ceo"
  },
  {
    id: 8,
    name: "员工推荐晋升",
    type: "manual",
    targetLevel: "level_2",
    targetLevelName: "白银会员",
    targetLevelColor: "#C0C0C0",
    status: "active",
    description: "员工推荐客户手动晋升",
    approvalProcess: "level1",
    approvers: "hr_manager"
  },
  {
    id: 11,
    name: "客服推荐晋升",
    type: "manual",
    targetLevel: "level_3",
    targetLevelName: "黄金会员",
    targetLevelColor: "#FFD700",
    status: "active",
    description: "客服推荐优质客户晋升",
    approvalProcess: "level1",
    approvers: "cs_manager"
  },
  {
    id: 12,
    name: "品牌大使晋升",
    type: "manual",
    targetLevel: "level_5",
    targetLevelName: "钻石会员",
    targetLevelColor: "#B9F2FF",
    status: "active",
    description: "品牌大使特殊晋升",
    approvalProcess: "level2",
    approvers: "marketing_director,ceo"
  },
]);

// 所有规则
const allRules = computed(() => [...autoPromotionRules.value, ...manualPromotionRules.value]);

// 过滤后的规则
const filteredRules = computed(() => {
  let result = [...allRules.value];

  // 根据标签页筛选
  if (activeTab.value === "auto") {
    result = [...autoPromotionRules.value];
  } else if (activeTab.value === "manual") {
    result = [...manualPromotionRules.value];
  }

  // 根据搜索关键词筛选
  if (searchQuery.value) {
    result = result.filter(
      (rule) =>
        rule.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        rule.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    );
  }

  return result;
});

// 表格列定义
const columns = [
  { accessorKey: "name", header: "规则名称" },
  { accessorKey: "type", header: "晋升类型" },
  { accessorKey: "targetLevel", header: "目标等级" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
];

// 表单验证
const isRuleFormValid = computed(() => {
  return ruleForm.value.name.trim() !== "" && ruleForm.value.targetLevel !== "";
});

// 获取规则类型标签
const getRuleTypeLabel = (type: string) => {
  const labels = {
    auto: "自动晋升",
    manual: "手动晋升",
  };
  return labels[type] || type;
};

// 获取审批流程标签
const getApprovalProcessLabel = (process: string) => {
  const labels = {
    none: "无需审批",
    level1: "一级审批",
    level2: "二级审批",
  };
  return labels[process] || process;
};

// 标签页切换
const onTabChange = (tab: string) => {
  activeTab.value = tab;
  currentPage.value = 1;
};

// 打开创建规则模态框
const openCreateRuleModal = () => {
  editingRule.value = null;
  ruleForm.value = {
    name: "",
    type: "auto",
    targetLevel: "level_1",
    status: "active",
    description: "",
    // 自动晋升条件
    growthValueThreshold: 0,
    spendThreshold: 0,
    purchaseCountThreshold: 0,
    membershipDuration: 0,
    // 手动晋升条件
    approvalProcess: "none",
    approvers: "",
    // 晋升奖励
    rewardPoints: 0,
    rewardCoupon: "",
  };
  showRuleModal.value = true;
};

// 查看规则
const viewRule = (rule: any) => {
  viewingRule.value = { ...rule };
  showViewRuleModal.value = true;
};

// 编辑规则
const editRule = (rule: any) => {
  editingRule.value = rule;
  ruleForm.value = { ...rule };
  showRuleModal.value = true;
};

// 关闭规则模态框
const closeRuleModal = () => {
  showRuleModal.value = false;
  editingRule.value = null;
};

// 保存规则
const saveRule = async () => {
  if (!isRuleFormValid.value) return;

  ruleLoading.value = true;

  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    if (editingRule.value) {
      // 更新现有规则
      const rule = allRules.value.find((r) => r.id === editingRule.value.id);
      if (rule) {
        Object.assign(rule, ruleForm.value);
      }
    } else {
      // 创建新规则
      const newRule = {
        id: Date.now(),
        ...ruleForm.value,
      };

      // 添加目标等级名称和颜色
      const targetLevel = membershipLevels.value.find(level => level.id === ruleForm.value.targetLevel);
      if (targetLevel) {
        newRule.targetLevelName = targetLevel.name;
        newRule.targetLevelColor = targetLevel.color;
      }

      if (ruleForm.value.type === "auto") {
        autoPromotionRules.value.push(newRule);
      } else {
        manualPromotionRules.value.push(newRule);
      }

      totalRules.value++;
    }

    // 显示成功消息
    alert(`${editingRule.value ? "更新" : "创建"}规则成功`);

    // 关闭模态框
    closeRuleModal();
  } catch (error) {
    console.error("保存规则失败:", error);
    alert("保存规则失败，请重试");
  } finally {
    ruleLoading.value = false;
  }
};

// 切换规则状态
const toggleRuleStatus = async (rule: any) => {
  const newStatus = rule.status === "active" ? "inactive" : "active";
  const action = newStatus === "active" ? "启用" : "停用";

  if (!confirm(`确定要${action}规则 "${rule.name}" 吗？`)) {
    return;
  }

  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 500));

    rule.status = newStatus;

    // 显示成功消息
    alert(`${action}规则成功`);
  } catch (error) {
    console.error(`${action}规则失败:`, error);
    alert(`${action}规则失败，请重试`);

    // 恢复原状态
    rule.status = rule.status === "active" ? "inactive" : "active";
  }
};

// 删除规则
const deleteRule = async (rule: any) => {
  if (!confirm(`确定要删除规则 "${rule.name}" 吗？此操作不可撤销。`)) {
    return;
  }

  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 500));

    // 从相应数组中删除
    if (rule.type === "auto") {
      const index = autoPromotionRules.value.findIndex((r) => r.id === rule.id);
      if (index !== -1) autoPromotionRules.value.splice(index, 1);
    } else {
      const index = manualPromotionRules.value.findIndex((r) => r.id === rule.id);
      if (index !== -1) manualPromotionRules.value.splice(index, 1);
    }

    totalRules.value--;

    // 显示成功消息
    alert("删除规则成功");
  } catch (error) {
    console.error("删除规则失败:", error);
    alert("删除规则失败，请重试");
  }
};

// 导出规则
const exportRules = (format: string) => {
  const formatLabel = format === "csv" ? "CSV" : "Excel";
  alert(`导出规则为${formatLabel}功能待实现`);
  console.log(`Exporting rules as ${format}`);
};
</script>
