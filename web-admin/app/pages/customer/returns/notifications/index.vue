<template>
  <div class="p-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">客户通知</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理退换货相关的客户通知和提醒</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-arrow-down-tray" @click="exportNotifications">导出通知</UButton>
      </div>
    </div>

    <!-- 通知统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-bell" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">待发送通知</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ pendingNotifications.length }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-check-badge" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">已发送通知</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ sentNotifications.length }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon name="i-heroicons-clock" class="w-6 h-6 text-yellow-600 dark:text-yellow-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">审核提醒</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ reviewReminders.length }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon name="i-heroicons-currency-dollar" class="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">退款提醒</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ refundReminders.length }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 标签页 -->
    <div class="mb-6">
      <UTabs v-model="activeTab" :items="tabs" />
    </div>

    <!-- 全局配置 -->
    <UCard v-if="activeTab === 'global'" class="mb-6">
      <template #header>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">通知配置</h2>
      </template>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            通知功能状态
          </label>
          <USelect
            v-model="notificationStatus"
            :options="[
              { label: '启用', value: 'enabled' },
              { label: '禁用', value: 'disabled' }
            ]"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            控制是否向客户发送退换货相关通知
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            通知发送方式
          </label>
          <USelect
            v-model="notificationChannels"
            multiple
            :options="[
              { label: '站内信', value: 'in-site' },
              { label: '短信', value: 'sms' },
              { label: '邮件', value: 'email' },
              { label: '微信', value: 'wechat' }
            ]"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            选择通知发送的方式（可多选）
          </p>
        </div>

        <div class="md:col-span-2">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            通知模板
          </label>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <UButton 
                color="primary" 
                variant="outline" 
                block 
                @click="editTemplate('review')"
                class="mb-2"
              >
                审核结果通知模板
              </UButton>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                审核通过/拒绝时发送给客户的提醒
              </p>
            </div>
            <div>
              <UButton 
                color="primary" 
                variant="outline" 
                block 
                @click="editTemplate('refund')"
                class="mb-2"
              >
                退款进度通知模板
              </UButton>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                退款状态更新时发送给客户的提醒
              </p>
            </div>
          </div>
        </div>
      </div>
    </UCard>

    <!-- 审核结果提醒 -->
    <UCard v-if="activeTab === 'review'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">审核结果提醒</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="reviewSearchQuery" 
              placeholder="搜索提醒..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton 
              color="primary" 
              variant="outline" 
              size="sm" 
              icon="i-heroicons-plus"
              @click="addReviewReminder"
            >
              添加提醒
            </UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请编号</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">客户</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">审核状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">审核时间</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">通知状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="(reminder, index) in filteredReviewReminders" :key="reminder.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ reminder.applicationId }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ reminder.customerName }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="getReviewStatusColor(reminder.reviewStatus)" variant="soft" size="xs">
                  {{ getReviewStatusText(reminder.reviewStatus) }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ reminder.reviewTime }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="getNotificationStatusColor(reminder.notificationStatus)" variant="soft" size="xs">
                  {{ getNotificationStatusText(reminder.notificationStatus) }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-eye"
                    @click="viewReminder(reminder)"
                  >
                    查看
                  </UButton>
                  <UButton
                    v-if="reminder.notificationStatus !== 'sent'"
                    color="primary"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-paper-airplane"
                    @click="sendReminder(reminder)"
                  >
                    发送
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
            显示第 {{ (reviewCurrentPage - 1) * reviewPageSize + 1 }}
            到 {{ Math.min(reviewCurrentPage * reviewPageSize, filteredReviewReminders.length) }} 条，共 {{ filteredReviewReminders.length }} 条
          </div>
          <UPagination
            v-model="reviewCurrentPage"
            :page-count="reviewPageCount"
            :total="filteredReviewReminders.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 退款进度提醒 -->
    <UCard v-if="activeTab === 'refund'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">退款进度提醒</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="refundSearchQuery" 
              placeholder="搜索提醒..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton 
              color="primary" 
              variant="outline" 
              size="sm" 
              icon="i-heroicons-plus"
              @click="addRefundReminder"
            >
              添加提醒
            </UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请编号</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">客户</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">退款状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">更新时间</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">通知状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="(reminder, index) in filteredRefundReminders" :key="reminder.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ reminder.applicationId }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ reminder.customerName }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="getRefundStatusColor(reminder.refundStatus)" variant="soft" size="xs">
                  {{ getRefundStatusText(reminder.refundStatus) }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ reminder.updateTime }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="getNotificationStatusColor(reminder.notificationStatus)" variant="soft" size="xs">
                  {{ getNotificationStatusText(reminder.notificationStatus) }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-eye"
                    @click="viewReminder(reminder)"
                  >
                    查看
                  </UButton>
                  <UButton
                    v-if="reminder.notificationStatus !== 'sent'"
                    color="primary"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-paper-airplane"
                    @click="sendReminder(reminder)"
                  >
                    发送
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
            显示第 {{ (refundCurrentPage - 1) * refundPageSize + 1 }}
            到 {{ Math.min(refundCurrentPage * refundPageSize, filteredRefundReminders.length) }} 条，共 {{ filteredRefundReminders.length }} 条
          </div>
          <UPagination
            v-model="refundCurrentPage"
            :page-count="refundPageCount"
            :total="filteredRefundReminders.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 操作按钮 -->
    <div class="flex justify-end space-x-3 mt-6">
      <UButton variant="ghost" @click="resetConfig">重置</UButton>
      <UButton color="primary" @click="saveConfig">保存配置</UButton>
    </div>

    <!-- 模板编辑模态框 -->
    <UModal
      v-model:open="showTemplateModal"
      :title="currentTemplateType === 'review' ? '审核结果通知模板' : '退款进度通知模板'"
      :description="currentTemplateType === 'review' ? '编辑审核结果通知的模板内容' : '编辑退款进度通知的模板内容'"
      :close="{ onClick: () => closeTemplateModal() }"
      :ui="{
        content: 'w-full sm:max-w-2xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-4 p-4 sm:p-6">
            <div class="grid grid-cols-1 gap-4">
              <UTextarea
                v-model="currentTemplate.content"
                label="通知内容"
                placeholder="请输入通知内容"
                :rows="5"
              />
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  可用变量
                </label>
                <div class="flex flex-wrap gap-2">
                  <UBadge 
                    v-for="variable in templateVariables" 
                    :key="variable.key" 
                    color="primary" 
                    variant="soft"
                    size="xs"
                    @click="insertVariable(variable.key)"
                    class="cursor-pointer"
                  >
                    {{ variable.label }} ({{ variable.key }})
                  </UBadge>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  点击变量可插入到通知内容中
                </p>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeTemplateModal">取消</UButton>
          <UButton color="primary" @click="saveTemplate">保存</UButton>
        </div>
      </template>
    </UModal>

    <!-- 提醒详情模态框 -->
    <UModal
      v-model:open="showDetailModal"
      title="提醒详情"
      description="查看通知提醒的详细信息"
      :close="{ onClick: () => closeDetailModal() }"
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
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  申请编号
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentReminder.applicationId }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  客户姓名
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentReminder.customerName }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  联系方式
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentReminder.contact }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  申请时间
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentReminder.applicationTime }}</p>
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  申请类型
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentReminder.applicationType }}</p>
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  通知内容
                </label>
                <div class="bg-gray-50 dark:bg-gray-700 p-3 rounded">
                  <p class="text-gray-900 dark:text-white whitespace-pre-wrap">{{ currentReminder.content }}</p>
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeDetailModal">关闭</UButton>
          <UButton 
            v-if="currentReminder.notificationStatus !== 'sent'"
            color="primary" 
            @click="sendReminder(currentReminder)"
          >
            发送通知
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
// 标签页
const activeTab = ref("global")

// 搜索查询
const reviewSearchQuery = ref("")
const refundSearchQuery = ref("")

// 分页
const reviewCurrentPage = ref(1)
const reviewPageSize = ref(10)

const refundCurrentPage = ref(1)
const refundPageSize = ref(10)

// 配置状态
const notificationStatus = ref("enabled")
const notificationChannels = ref(["in-site", "sms"])

// 模态框
const showTemplateModal = ref(false)
const showDetailModal = ref(false)

// 当前编辑的数据
const currentTemplateType = ref("")
const currentTemplate = ref({
  content: ""
})

const currentReminder = ref({
  id: "",
  applicationId: "",
  customerName: "",
  contact: "",
  applicationTime: "",
  applicationType: "",
  reviewStatus: "",
  refundStatus: "",
  reviewTime: "",
  updateTime: "",
  notificationStatus: "",
  content: ""
})

// Tabs
const tabs = [
  { label: "全局配置", value: "global" },
  { label: "审核结果提醒", value: "review" },
  { label: "退款进度提醒", value: "refund" },
]

// 模板变量
const templateVariables = [
  { key: "{{customer_name}}", label: "客户姓名" },
  { key: "{{application_id}}", label: "申请编号" },
  { key: "{{application_type}}", label: "申请类型" },
  { key: "{{review_status}}", label: "审核状态" },
  { key: "{{refund_status}}", label: "退款状态" },
  { key: "{{review_time}}", label: "审核时间" },
  { key: "{{update_time}}", label: "更新时间" }
]

// 模拟数据
const pendingNotifications = ref([
  { id: "1" },
  { id: "2" },
  { id: "3" }
])

const sentNotifications = ref([
  { id: "1" },
  { id: "2" },
  { id: "3" },
  { id: "4" },
  { id: "5" }
])

const reviewReminders = ref([
  {
    id: "1",
    applicationId: "RMA2023001",
    customerName: "张三",
    contact: "13800138000",
    applicationTime: "2023-05-01 10:30",
    applicationType: "退货",
    reviewStatus: "approved",
    reviewTime: "2023-05-02 14:20",
    notificationStatus: "pending",
    content: "尊敬的张三，您的退货申请已通过审核，请按指引寄回商品。"
  },
  {
    id: "2",
    applicationId: "RMA2023002",
    customerName: "李四",
    contact: "13900139000",
    applicationTime: "2023-05-03 09:15",
    applicationType: "换货",
    reviewStatus: "rejected",
    reviewTime: "2023-05-04 11:45",
    notificationStatus: "sent",
    content: "尊敬的李四，您的换货申请未通过审核，原因是商品已过换货期。"
  }
])

const refundReminders = ref([
  {
    id: "1",
    applicationId: "RMA2023003",
    customerName: "王五",
    contact: "13700137000",
    applicationTime: "2023-05-05 16:20",
    applicationType: "退货",
    refundStatus: "processing",
    updateTime: "2023-05-06 09:30",
    notificationStatus: "pending",
    content: "尊敬的王五，您的退款正在处理中，预计1-3个工作日内到账。"
  },
  {
    id: "2",
    applicationId: "RMA2023004",
    customerName: "赵六",
    contact: "13600136000",
    applicationTime: "2023-05-07 14:10",
    applicationType: "退货",
    refundStatus: "completed",
    updateTime: "2023-05-10 15:45",
    notificationStatus: "sent",
    content: "尊敬的赵六，您的退款已完成，金额已原路返回。"
  }
])

// 过滤
const filteredReviewReminders = computed(() => {
  if (!reviewSearchQuery.value) return reviewReminders.value
  const q = reviewSearchQuery.value.toLowerCase()
  return reviewReminders.value.filter(reminder =>
    reminder.applicationId.toLowerCase().includes(q) ||
    reminder.customerName.toLowerCase().includes(q)
  )
})

const filteredRefundReminders = computed(() => {
  if (!refundSearchQuery.value) return refundReminders.value
  const q = refundSearchQuery.value.toLowerCase()
  return refundReminders.value.filter(reminder =>
    reminder.applicationId.toLowerCase().includes(q) ||
    reminder.customerName.toLowerCase().includes(q)
  )
})

// 分页数据
const reviewPageData = computed(() => {
  const start = (reviewCurrentPage.value - 1) * reviewPageSize.value
  const end = start + reviewPageSize.value
  return filteredReviewReminders.value.slice(start, end)
})

const refundPageData = computed(() => {
  const start = (refundCurrentPage.value - 1) * refundPageSize.value
  const end = start + refundPageSize.value
  return filteredRefundReminders.value.slice(start, end)
})

// 页数
const reviewPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredReviewReminders.value.length / reviewPageSize.value))
)

const refundPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredRefundReminders.value.length / refundPageSize.value))
)

// 状态文本映射
const getReviewStatusText = (status) => {
  const statusMap = {
    approved: "审核通过",
    rejected: "审核拒绝",
    pending: "待审核"
  }
  return statusMap[status] || status
}

const getReviewStatusColor = (status) => {
  const colorMap = {
    approved: "success",
    rejected: "error",
    pending: "warning"
  }
  return colorMap[status] || "neutral"
}

const getRefundStatusText = (status) => {
  const statusMap = {
    pending: "待处理",
    processing: "处理中",
    completed: "已完成",
    failed: "失败"
  }
  return statusMap[status] || status
}

const getRefundStatusColor = (status) => {
  const colorMap = {
    pending: "warning",
    processing: "primary",
    completed: "success",
    failed: "error"
  }
  return colorMap[status] || "neutral"
}

const getNotificationStatusText = (status) => {
  const statusMap = {
    pending: "待发送",
    sent: "已发送",
    failed: "发送失败"
  }
  return statusMap[status] || status
}

const getNotificationStatusColor = (status) => {
  const colorMap = {
    pending: "warning",
    sent: "success",
    failed: "error"
  }
  return colorMap[status] || "neutral"
}

// 模板操作
const editTemplate = (type) => {
  currentTemplateType.value = type
  // 模拟加载模板内容
  if (type === 'review') {
    currentTemplate.value.content = "尊敬的{{customer_name}}，您的{{application_type}}申请（编号：{{application_id}}）审核{{review_status}}，请登录查看详细信息。"
  } else {
    currentTemplate.value.content = "尊敬的{{customer_name}}，您的退款申请（编号：{{application_id}}）状态已更新为{{refund_status}}，{{update_time}}。"
  }
  showTemplateModal.value = true
}

const closeTemplateModal = () => {
  showTemplateModal.value = false
  currentTemplate.value = {
    content: ""
  }
}

const saveTemplate = () => {
  // 保存模板逻辑
  alert(`${currentTemplateType.value === 'review' ? '审核结果' : '退款进度'}通知模板已保存`)
  closeTemplateModal()
}

const insertVariable = (variable) => {
  const content = currentTemplate.value.content
  const cursorPosition = document.activeElement?.selectionStart || content.length
  currentTemplate.value.content = content.slice(0, cursorPosition) + variable + content.slice(cursorPosition)
}

// 提醒操作
const addReviewReminder = () => {
  // 添加审核提醒逻辑
  alert("添加审核提醒")
}

const addRefundReminder = () => {
  // 添加退款提醒逻辑
  alert("添加退款提醒")
}

const viewReminder = (reminder) => {
  currentReminder.value = { ...reminder }
  showDetailModal.value = true
}

const closeDetailModal = () => {
  showDetailModal.value = false
  currentReminder.value = {
    id: "",
    applicationId: "",
    customerName: "",
    contact: "",
    applicationTime: "",
    applicationType: "",
    reviewStatus: "",
    refundStatus: "",
    reviewTime: "",
    updateTime: "",
    notificationStatus: "",
    content: ""
  }
}

const sendReminder = (reminder) => {
  // 发送提醒逻辑
  alert(`通知已发送给 ${reminder.customerName}`)
  // 更新状态
  if (reviewReminders.value.find(r => r.id === reminder.id)) {
    reviewReminders.value = reviewReminders.value.map(r => 
      r.id === reminder.id ? { ...r, notificationStatus: "sent" } : r
    )
  }
  if (refundReminders.value.find(r => r.id === reminder.id)) {
    refundReminders.value = refundReminders.value.map(r => 
      r.id === reminder.id ? { ...r, notificationStatus: "sent" } : r
    )
  }
}

// 配置操作
const saveConfig = () => {
  alert("配置已保存")
}

const resetConfig = () => {
  if (confirm("确定要重置所有配置吗？此操作不可恢复。")) {
    // 重置逻辑
    alert("配置已重置")
  }
}

const exportNotifications = () => {
  alert("通知记录已导出")
}
</script>