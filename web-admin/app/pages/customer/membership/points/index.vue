<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          积分规则
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          配置积分获取和消耗规则
        </p>
      </div>
      <div class="flex gap-3">
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="openCreateRuleModal"
        >
          创建规则
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
              获取规则
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ earnRules.length }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon
              name="i-heroicons-arrow-trending-down"
              class="w-6 h-6 text-purple-600 dark:text-purple-400"
            />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600 dark:text-gray-400">
              消耗规则
            </p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ spendRules.length }}
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
            {{ activeTab === 'all' ? '所有规则' : activeTab === 'earn' ? '获取规则' : '消耗规则' }}
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

        <template #points-cell="{ row }">
          <div class="font-medium text-gray-900 dark:text-white">
            {{ row.points }} 积分
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

    <!-- 创建/编辑规则模态框 -->
    <UModal
      v-model:open="showRuleModal"
      :title="editingRule ? '编辑积分规则' : '创建积分规则'"
      :description="editingRule ? '修改积分规则配置' : '创建新的积分规则'"
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
                  规则类型 <span class="text-red-500">*</span>
                </label>
                <USelect
                  v-model="ruleForm.type"
                  :options="ruleTypeOptions"
                  option-attribute="label"
                  value-attribute="value"
                />
              </div>

              <div v-if="ruleForm.type !== 'expiration'">
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  积分数量 <span class="text-red-500">*</span>
                </label>
                <UInput
                  v-model.number="ruleForm.points"
                  type="number"
                  placeholder="请输入积分数"
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

              <!-- 特定规则类型的配置选项 -->
              <div class="sm:col-span-2" v-if="ruleForm.type === 'purchase_reward'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  购物奖励配置
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      奖励比例
                    </label>
                    <UInput
                      v-model="ruleForm.purchaseRatio"
                      type="number"
                      step="0.1"
                      placeholder="1元=X积分"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      最低消费金额
                    </label>
                    <UInput
                      v-model="ruleForm.minAmount"
                      type="number"
                      placeholder="最低消费金额"
                    />
                  </div>
                  <div class="sm:col-span-2">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      限制类目
                    </label>
                    <USelect
                      v-model="ruleForm.categories"
                      multiple
                      :options="[
                        { label: '全部类目', value: 'all' },
                        { label: '服装', value: 'clothing' },
                        { label: '电子产品', value: 'electronics' },
                        { label: '家居用品', value: 'home' },
                        { label: '美妆护肤', value: 'beauty' },
                      ]"
                      option-attribute="label"
                      value-attribute="value"
                    />
                  </div>
                </div>
              </div>

              <div class="sm:col-span-2" v-else-if="ruleForm.type === 'checkin_reward'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  签到奖励配置
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      基础奖励积分
                    </label>
                    <UInput
                      v-model="ruleForm.basePoints"
                      type="number"
                      placeholder="基础奖励积分"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      连续签到奖励递增
                    </label>
                    <UInput
                      v-model="ruleForm.incrementPoints"
                      type="number"
                      placeholder="每日递增积分"
                    />
                  </div>
                  <div class="sm:col-span-2">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      连续签到上限天数
                    </label>
                    <UInput
                      v-model="ruleForm.maxDays"
                      type="number"
                      placeholder="连续签到上限天数"
                    />
                  </div>
                </div>
              </div>

              <div class="sm:col-span-2" v-else-if="ruleForm.type === 'review_reward'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  评价奖励配置
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      文字评价奖励
                    </label>
                    <UInput
                      v-model="ruleForm.textReviewPoints"
                      type="number"
                      placeholder="文字评价奖励积分"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      图片评价奖励
                    </label>
                    <UInput
                      v-model="ruleForm.imageReviewPoints"
                      type="number"
                      placeholder="图片评价奖励积分"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      视频评价奖励
                    </label>
                    <UInput
                      v-model="ruleForm.videoReviewPoints"
                      type="number"
                      placeholder="视频评价奖励积分"
                    />
                  </div>
                </div>
              </div>

              <div class="sm:col-span-2" v-else-if="ruleForm.type === 'share_reward'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  分享奖励配置
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      分享奖励积分
                    </label>
                    <UInput
                      v-model="ruleForm.sharePoints"
                      type="number"
                      placeholder="分享奖励积分"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      每日分享上限
                    </label>
                    <UInput
                      v-model="ruleForm.dailyShareLimit"
                      type="number"
                      placeholder="每日分享上限"
                    />
                  </div>
                </div>
              </div>

              <div class="sm:col-span-2" v-else-if="ruleForm.type === 'order_deduction'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  订单抵扣配置
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      抵扣比例
                    </label>
                    <UInput
                      v-model="ruleForm.deductionRatio"
                      type="number"
                      step="0.1"
                      placeholder="X积分=1元"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      单笔订单最高抵扣
                    </label>
                    <UInput
                      v-model="ruleForm.maxDeductionPerOrder"
                      type="number"
                      placeholder="单笔订单最高抵扣金额"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      订单最低金额要求
                    </label>
                    <UInput
                      v-model="ruleForm.minOrderAmount"
                      type="number"
                      placeholder="订单最低金额要求"
                    />
                  </div>
                </div>
              </div>

              <div class="sm:col-span-2" v-else-if="ruleForm.type === 'expiration'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  过期机制配置
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      过期时间
                    </label>
                    <USelect
                      v-model="ruleForm.expirationType"
                      :options="[
                        { label: '固定日期', value: 'fixed_date' },
                        { label: '获得后N天', value: 'days_after' },
                      ]"
                      option-attribute="label"
                      value-attribute="value"
                    />
                  </div>
                  <div v-if="ruleForm.expirationType === 'fixed_date'">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      固定日期
                    </label>
                    <UInput
                      v-model="ruleForm.fixedDate"
                      type="date"
                    />
                  </div>
                  <div v-else-if="ruleForm.expirationType === 'days_after'">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      获得后天数
                    </label>
                    <UInput
                      v-model="ruleForm.daysAfter"
                      type="number"
                      placeholder="获得后多少天过期"
                    />
                  </div>
                  <div class="sm:col-span-2">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      通知设置
                    </label>
                    <div class="flex items-center gap-4">
                      <UCheckbox
                        v-model="ruleForm.notifyBeforeExpiration"
                        label="过期前通知"
                      />
                      <UInput
                        v-if="ruleForm.notifyBeforeExpiration"
                        v-model="ruleForm.notificationDays"
                        type="number"
                        placeholder="提前多少天通知"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <div class="sm:col-span-2" v-else-if="ruleForm.type === 'points_redemption' || ruleForm.type === 'coupon_redemption' || ruleForm.type === 'gift_redemption'">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  兑换配置
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      兑换比例
                    </label>
                    <div class="flex items-center gap-2">
                      <UInput
                        v-model="ruleForm.redemptionPoints"
                        type="number"
                        placeholder="所需积分"
                      />
                      <span class="text-gray-500">积分 =</span>
                      <UInput
                        v-model="ruleForm.redemptionValue"
                        type="text"
                        placeholder="兑换价值"
                      />
                    </div>
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      每人兑换上限
                    </label>
                    <UInput
                      v-model="ruleForm.redemptionLimit"
                      type="number"
                      placeholder="每人兑换上限"
                    />
                  </div>
                  <div class="sm:col-span-2">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      兑换时间限制
                    </label>
                    <div class="flex items-center gap-2">
                      <UInput
                        v-model="ruleForm.startTime"
                        type="datetime-local"
                      />
                      <span class="text-gray-500">至</span>
                      <UInput
                        v-model="ruleForm.endTime"
                        type="datetime-local"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <div class="sm:col-span-2">
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  适用条件
                </label>
                <UTextarea
                  v-model="ruleForm.conditions"
                  placeholder="请输入适用条件，每行一个条件"
                  rows="4"
                />
                <p class="text-gray-500 dark:text-gray-400 text-xs mt-1">
                  支持的条件变量: {order_amount}(订单金额), {product_category}(商品类目), {customer_level}(会员等级)
                </p>
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

// 当前编辑的规则
const editingRule = ref<any>(null);

// 活动标签页
const activeTab = ref("all");

// 搜索
const searchQuery = ref("");

// 分页
const currentPage = ref(1);
const pageSize = ref(10);
const pageCount = computed(() => Math.ceil(filteredRules.value.length / pageSize.value));

// 标签页配置
const tabs = [
  { label: "所有规则", value: "all" },
  { label: "获取规则", value: "earn" },
  { label: "消耗规则", value: "spend" },
];

// 规则类型选项
const ruleTypeOptions = [
  // 获取规则类型
  { label: "购物奖励", value: "purchase_reward", category: "earn" },
  { label: "签到奖励", value: "checkin_reward", category: "earn" },
  { label: "评价奖励", value: "review_reward", category: "earn" },
  { label: "分享奖励", value: "share_reward", category: "earn" },
  { label: "活动奖励", value: "event_reward", category: "earn" },
  { label: "推荐奖励", value: "referral_reward", category: "earn" },
  // 消耗规则类型
  { label: "订单抵扣", value: "order_deduction", category: "spend" },
  { label: "积分兑换", value: "points_redemption", category: "spend" },
  { label: "优惠券兑换", value: "coupon_redemption", category: "spend" },
  { label: "礼品兑换", value: "gift_redemption", category: "spend" },
  { label: "过期机制", value: "expiration", category: "spend" },
];

// 状态选项
const statusOptions = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

// 规则表单
const ruleForm = ref({
  name: "",
  type: "purchase_reward",
  points: 0,
  status: "active",
  description: "",
  conditions: "",
  // 购物奖励配置
  purchaseRatio: 1,
  minAmount: 0,
  categories: [],
  // 签到奖励配置
  basePoints: 10,
  incrementPoints: 0,
  maxDays: 7,
  // 评价奖励配置
  textReviewPoints: 50,
  imageReviewPoints: 100,
  videoReviewPoints: 200,
  // 分享奖励配置
  sharePoints: 20,
  dailyShareLimit: 5,
  // 订单抵扣配置
  deductionRatio: 100,
  maxDeductionPerOrder: 0,
  minOrderAmount: 0,
  // 过期机制配置
  expirationType: "fixed_date",
  fixedDate: "",
  daysAfter: 365,
  notifyBeforeExpiration: false,
  notificationDays: 7,
  // 兑换配置
  redemptionPoints: 100,
  redemptionValue: "1元",
  redemptionLimit: 0,
  startTime: "",
  endTime: "",
});

// 加载状态
const ruleLoading = ref(false);

// 总规则数
const totalRules = ref(11);

// 获取规则
const earnRules = ref([
  {
    id: 1,
    name: "购物奖励",
    type: "purchase_reward",
    points: 1,
    status: "active",
    description: "每消费1元获得1积分",
    purchaseRatio: 1,
    minAmount: 0,
    categories: ["all"]
  },
  {
    id: 2,
    name: "签到奖励",
    type: "checkin_reward",
    points: 10,
    status: "active",
    description: "每日签到获得10积分",
    basePoints: 10,
    incrementPoints: 2,
    maxDays: 7
  },
  {
    id: 3,
    name: "评价奖励",
    type: "review_reward",
    points: 50,
    status: "active",
    description: "商品评价获得50积分",
    textReviewPoints: 50,
    imageReviewPoints: 100,
    videoReviewPoints: 200
  },
  {
    id: 4,
    name: "分享奖励",
    type: "share_reward",
    points: 20,
    status: "active",
    description: "分享商品获得20积分",
    sharePoints: 20,
    dailyShareLimit: 5
  },
  {
    id: 5,
    name: "活动奖励",
    type: "event_reward",
    points: 100,
    status: "active",
    description: "参与活动获得100积分"
  },
  {
    id: 6,
    name: "推荐奖励",
    type: "referral_reward",
    points: 50,
    status: "inactive",
    description: "推荐好友注册获得50积分"
  },
]);

// 消耗规则
const spendRules = ref([
  {
    id: 7,
    name: "订单抵扣",
    type: "order_deduction",
    points: -100,
    status: "active",
    description: "100积分抵扣1元",
    deductionRatio: 100,
    maxDeductionPerOrder: 1000,
    minOrderAmount: 50
  },
  {
    id: 8,
    name: "积分兑换",
    type: "points_redemption",
    points: -100,
    status: "active",
    description: "100积分兑换1元余额",
    redemptionPoints: 100,
    redemptionValue: "1元",
    redemptionLimit: 0
  },
  {
    id: 9,
    name: "优惠券兑换",
    type: "coupon_redemption",
    points: -500,
    status: "active",
    description: "500积分兑换10元优惠券",
    redemptionPoints: 500,
    redemptionValue: "10元优惠券",
    redemptionLimit: 5
  },
  {
    id: 10,
    name: "礼品兑换",
    type: "gift_redemption",
    points: -2000,
    status: "active",
    description: "2000积分兑换精美礼品",
    redemptionPoints: 2000,
    redemptionValue: "精美礼品",
    redemptionLimit: 10
  },
  {
    id: 11,
    name: "过期机制",
    type: "expiration",
    points: 0,
    status: "active",
    description: "积分每年1月1日过期",
    expirationType: "fixed_date",
    fixedDate: "2024-01-01",
    notifyBeforeExpiration: true,
    notificationDays: 30
  },
]);

// 所有规则
const allRules = computed(() => [...earnRules.value, ...spendRules.value]);

// 过滤后的规则
const filteredRules = computed(() => {
  let result = [...allRules.value];

  // 根据标签页筛选
  if (activeTab.value === "earn") {
    result = [...earnRules.value];
  } else if (activeTab.value === "spend") {
    result = [...spendRules.value];
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
  { accessorKey: "type", header: "规则类型" },
  { accessorKey: "points", header: "积分" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
];

// 表单验证
const isRuleFormValid = computed(() => {
  // 基本验证
  if (ruleForm.value.name.trim() === "") return false;

  // 根据规则类型进行特定验证
  switch (ruleForm.value.type) {
    case 'expiration':
      // 过期机制不需要积分数量
      return true;
    case 'purchase_reward':
      return ruleForm.value.purchaseRatio > 0;
    case 'checkin_reward':
      return ruleForm.value.basePoints > 0;
    case 'review_reward':
      return ruleForm.value.textReviewPoints > 0 ||
             ruleForm.value.imageReviewPoints > 0 ||
             ruleForm.value.videoReviewPoints > 0;
    case 'share_reward':
      return ruleForm.value.sharePoints > 0;
    case 'order_deduction':
      return ruleForm.value.deductionRatio > 0;
    case 'points_redemption':
    case 'coupon_redemption':
    case 'gift_redemption':
      return ruleForm.value.redemptionPoints > 0 && ruleForm.value.redemptionValue.trim() !== "";
    default:
      return ruleForm.value.points !== 0;
  }
});

// 获取规则类型标签
const getRuleTypeLabel = (type: string) => {
  const labels = {
    purchase_reward: "购物奖励",
    checkin_reward: "签到奖励",
    review_reward: "评价奖励",
    share_reward: "分享奖励",
    event_reward: "活动奖励",
    referral_reward: "推荐奖励",
    order_deduction: "订单抵扣",
    points_redemption: "积分兑换",
    coupon_redemption: "优惠券兑换",
    gift_redemption: "礼品兑换",
    expiration: "过期机制",
  };
  return labels[type] || type;
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
    type: "purchase_reward",
    points: 0,
    status: "active",
    description: "",
    conditions: "",
    // 购物奖励配置
    purchaseRatio: 1,
    minAmount: 0,
    categories: [],
    // 签到奖励配置
    basePoints: 10,
    incrementPoints: 0,
    maxDays: 7,
    // 评价奖励配置
    textReviewPoints: 50,
    imageReviewPoints: 100,
    videoReviewPoints: 200,
    // 分享奖励配置
    sharePoints: 20,
    dailyShareLimit: 5,
    // 订单抵扣配置
    deductionRatio: 100,
    maxDeductionPerOrder: 0,
    minOrderAmount: 0,
    // 过期机制配置
    expirationType: "fixed_date",
    fixedDate: "",
    daysAfter: 365,
    notifyBeforeExpiration: false,
    notificationDays: 7,
    // 兑换配置
    redemptionPoints: 100,
    redemptionValue: "1元",
    redemptionLimit: 0,
    startTime: "",
    endTime: "",
  };
  showRuleModal.value = true;
};

// 查看规则
const viewRule = (rule: any) => {
  // 这里可以打开一个查看模态框或跳转到详情页
  alert(`查看规则: ${rule.name}`);
};

// 编辑规则
const editRule = (rule: any) => {
  editingRule.value = rule;
  ruleForm.value = {
    ...rule,
    // 确保所有字段都有默认值
    purchaseRatio: rule.purchaseRatio || 1,
    minAmount: rule.minAmount || 0,
    categories: rule.categories || [],
    basePoints: rule.basePoints || 10,
    incrementPoints: rule.incrementPoints || 0,
    maxDays: rule.maxDays || 7,
    textReviewPoints: rule.textReviewPoints || 50,
    imageReviewPoints: rule.imageReviewPoints || 100,
    videoReviewPoints: rule.videoReviewPoints || 200,
    sharePoints: rule.sharePoints || 20,
    dailyShareLimit: rule.dailyShareLimit || 5,
    deductionRatio: rule.deductionRatio || 100,
    maxDeductionPerOrder: rule.maxDeductionPerOrder || 0,
    minOrderAmount: rule.minOrderAmount || 0,
    expirationType: rule.expirationType || "fixed_date",
    fixedDate: rule.fixedDate || "",
    daysAfter: rule.daysAfter || 365,
    notifyBeforeExpiration: rule.notifyBeforeExpiration || false,
    notificationDays: rule.notificationDays || 7,
    redemptionPoints: rule.redemptionPoints || 100,
    redemptionValue: rule.redemptionValue || "1元",
    redemptionLimit: rule.redemptionLimit || 0,
    startTime: rule.startTime || "",
    endTime: rule.endTime || "",
  };
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

      if (ruleForm.value.points > 0 ||
          ['purchase_reward', 'checkin_reward', 'review_reward', 'share_reward', 'event_reward', 'referral_reward'].includes(ruleForm.value.type)) {
        earnRules.value.push(newRule);
      } else {
        spendRules.value.push(newRule);
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
    if (rule.points > 0) {
      const index = earnRules.value.findIndex((r) => r.id === rule.id);
      if (index !== -1) earnRules.value.splice(index, 1);
    } else {
      const index = spendRules.value.findIndex((r) => r.id === rule.id);
      if (index !== -1) spendRules.value.splice(index, 1);
    }

    totalRules.value--;

    // 显示成功消息
    alert("删除规则成功");
  } catch (error) {
    console.error("删除规则失败:", error);
    alert("删除规则失败，请重试");
  }
};
</script>
