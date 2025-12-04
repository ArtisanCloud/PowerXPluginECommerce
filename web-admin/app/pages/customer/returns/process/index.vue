<template>
  <div class="p-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">RMA 流程配置</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">配置退换货流程规则和地址信息</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-arrow-down-tray" @click="exportConfig">导出配置</UButton>
      </div>
    </div>

    <!-- 配置卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-document-text" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">审核规则</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ reviewRules.length }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-map-pin" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">退货地址</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ returnAddresses.length }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon name="i-heroicons-truck" class="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">物流规则</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ logisticsRules.length }}</p>
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
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">全局配置</h2>
      </template>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            RMA流程状态
          </label>
          <USelect
            v-model="processStatus"
            :options="[
              { label: '启用', value: 'enabled' },
              { label: '禁用', value: 'disabled' }
            ]"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            控制整个RMA流程是否启用
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            默认退货地址
          </label>
          <USelect
            v-model="defaultReturnAddress"
            :options="returnAddressOptions"
            option-attribute="name"
            value-attribute="id"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            客户未指定时使用的默认退货地址
          </p>
        </div>

        <div class="md:col-span-2">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            说明文本
          </label>
          <UTextarea
            v-model="instructions"
            placeholder="请输入RMA流程的说明文本"
            :rows="3"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            显示在RMA流程页面的说明文本
          </p>
        </div>
      </div>
    </UCard>

    <!-- 审核规则配置 -->
    <UCard v-if="activeTab === 'review'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">审核规则配置</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="reviewSearchQuery" 
              placeholder="搜索规则..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton 
              color="primary" 
              variant="outline" 
              size="sm" 
              icon="i-heroicons-plus"
              @click="addReviewRule"
            >
              添加规则
            </UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">规则名称</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">适用类目</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">退货期限</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">无理由退货</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="(rule, index) in filteredReviewRules" :key="rule.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ rule.name }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ rule.category || '所有类目' }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ rule.returnPeriod }}天</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="rule.noReasonReturn ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ rule.noReasonReturn ? '支持' : '不支持' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="rule.enabled ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ rule.enabled ? '启用' : '禁用' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-pencil"
                    @click="editReviewRule(rule)"
                  >
                    编辑
                  </UButton>
                  <UButton
                    color="error"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="removeReviewRule(index)"
                  >
                    删除
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
            到 {{ Math.min(reviewCurrentPage * reviewPageSize, filteredReviewRules.length) }} 条，共 {{ filteredReviewRules.length }} 条
          </div>
          <UPagination
            v-model="reviewCurrentPage"
            :page-count="reviewPageCount"
            :total="filteredReviewRules.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 退货地址配置 -->
    <UCard v-if="activeTab === 'address'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">退货地址配置</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="addressSearchQuery" 
              placeholder="搜索地址..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton 
              color="primary" 
              variant="outline" 
              size="sm" 
              icon="i-heroicons-plus"
              @click="addReturnAddress"
            >
              添加地址
            </UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">地址名称</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">联系人</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">联系电话</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">地址</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">适用地区</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="(address, index) in filteredReturnAddresses" :key="address.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ address.name }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ address.contact }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ address.phone }}</td>
              <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400 w-64">
                {{ address.province }}{{ address.city }}{{ address.district }}{{ address.address }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ address.applicableRegions || '全国' }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="address.enabled ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ address.enabled ? '启用' : '禁用' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-pencil"
                    @click="editReturnAddress(address)"
                  >
                    编辑
                  </UButton>
                  <UButton
                    color="error"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="removeReturnAddress(index)"
                  >
                    删除
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
            显示第 {{ (addressCurrentPage - 1) * addressPageSize + 1 }}
            到 {{ Math.min(addressCurrentPage * addressPageSize, filteredReturnAddresses.length) }} 条，共 {{ filteredReturnAddresses.length }} 条
          </div>
          <UPagination
            v-model="addressCurrentPage"
            :page-count="addressPageCount"
            :total="filteredReturnAddresses.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 退货物流规则配置 -->
    <UCard v-if="activeTab === 'logistics'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">退货物流规则配置</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="logisticsSearchQuery" 
              placeholder="搜索规则..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton 
              color="primary" 
              variant="outline" 
              size="sm" 
              icon="i-heroicons-plus"
              @click="addLogisticsRule"
            >
              添加规则
            </UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">规则名称</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">承运商</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">运费承担</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">追踪要求</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="(rule, index) in filteredLogisticsRules" :key="rule.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ rule.name }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ rule.carrier }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                {{ rule.freightPayer === 'customer' ? '客户承担' : '商家承担' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="rule.trackingRequired ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ rule.trackingRequired ? '需要' : '不需要' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="rule.enabled ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ rule.enabled ? '启用' : '禁用' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-pencil"
                    @click="editLogisticsRule(rule)"
                  >
                    编辑
                  </UButton>
                  <UButton
                    color="error"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="removeLogisticsRule(index)"
                  >
                    删除
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
            显示第 {{ (logisticsCurrentPage - 1) * logisticsPageSize + 1 }}
            到 {{ Math.min(logisticsCurrentPage * logisticsPageSize, filteredLogisticsRules.length) }} 条，共 {{ filteredLogisticsRules.length }} 条
          </div>
          <UPagination
            v-model="logisticsCurrentPage"
            :page-count="logisticsPageCount"
            :total="filteredLogisticsRules.length"
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

    <!-- 审核规则编辑模态框 -->
    <UModal
      v-model:open="showReviewRuleModal"
      :title="editingReviewRule ? '编辑审核规则' : '添加审核规则'"
      description="配置退换货审核规则"
      :close="{ onClick: () => closeReviewRuleModal() }"
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
                v-model="currentReviewRule.name"
                label="规则名称"
                placeholder="请输入规则名称"
                :error="!!reviewRuleErrors.name"
              />
              <USelect
                v-model="currentReviewRule.category"
                label="适用商品类目"
                :options="categoryOptions"
                placeholder="选择商品类目"
              />
              <UInput
                v-model.number="currentReviewRule.returnPeriod"
                label="退货期限（天）"
                type="number"
                min="1"
                :error="!!reviewRuleErrors.returnPeriod"
              />
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  无理由退货
                </label>
                <USelect
                  v-model="currentReviewRule.noReasonReturn"
                  :options="[
                    { label: '支持', value: true },
                    { label: '不支持', value: false }
                  ]"
                />
              </div>
              <div class="sm:col-span-2">
                <UTextarea
                  v-model="currentReviewRule.description"
                  label="规则描述"
                  placeholder="请输入规则描述（可选）"
                  :rows="2"
                />
              </div>
              <div>
                <UCheckbox v-model="currentReviewRule.enabled" label="是否启用" />
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeReviewRuleModal">取消</UButton>
          <UButton color="primary" @click="saveReviewRule">保存</UButton>
        </div>
      </template>
    </UModal>

    <!-- 退货地址编辑模态框 -->
    <UModal
      v-model:open="showReturnAddressModal"
      :title="editingReturnAddress ? '编辑退货地址' : '添加退货地址'"
      description="配置退货地址信息"
      :close="{ onClick: () => closeReturnAddressModal() }"
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
                v-model="currentReturnAddress.name"
                label="地址名称"
                placeholder="请输入地址名称"
                :error="!!returnAddressErrors.name"
              />
              <UInput
                v-model="currentReturnAddress.contact"
                label="联系人"
                placeholder="请输入联系人姓名"
                :error="!!returnAddressErrors.contact"
              />
              <UInput
                v-model="currentReturnAddress.phone"
                label="联系电话"
                placeholder="请输入联系电话"
                :error="!!returnAddressErrors.phone"
              />
              <UInput
                v-model="currentReturnAddress.postcode"
                label="邮政编码"
                placeholder="请输入邮政编码"
              />
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  省份
                </label>
                <USelect
                  v-model="currentReturnAddress.province"
                  :options="provinceOptions"
                  placeholder="选择省份"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  城市
                </label>
                <USelect
                  v-model="currentReturnAddress.city"
                  :options="cityOptions"
                  placeholder="选择城市"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  区县
                </label>
                <USelect
                  v-model="currentReturnAddress.district"
                  :options="districtOptions"
                  placeholder="选择区县"
                />
              </div>
              <div class="sm:col-span-2">
                <UInput
                  v-model="currentReturnAddress.address"
                  label="详细地址"
                  placeholder="请输入详细地址"
                  :error="!!returnAddressErrors.address"
                />
              </div>
              <div class="sm:col-span-2">
                <UTextarea
                  v-model="currentReturnAddress.applicableRegions"
                  label="适用地区"
                  placeholder="请输入适用地区（多个地区用逗号分隔）"
                  :rows="2"
                />
              </div>
              <div>
                <UCheckbox v-model="currentReturnAddress.enabled" label="是否启用" />
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeReturnAddressModal">取消</UButton>
          <UButton color="primary" @click="saveReturnAddress">保存</UButton>
        </div>
      </template>
    </UModal>

    <!-- 物流规则编辑模态框 -->
    <UModal
      v-model:open="showLogisticsRuleModal"
      :title="editingLogisticsRule ? '编辑物流规则' : '添加物流规则'"
      description="配置退货物流规则"
      :close="{ onClick: () => closeLogisticsRuleModal() }"
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
                v-model="currentLogisticsRule.name"
                label="规则名称"
                placeholder="请输入规则名称"
                :error="!!logisticsRuleErrors.name"
              />
              <USelect
                v-model="currentLogisticsRule.carrier"
                label="承运商"
                :options="carrierOptions"
                placeholder="选择承运商"
              />
              <USelect
                v-model="currentLogisticsRule.freightPayer"
                label="运费承担"
                :options="freightPayerOptions"
              />
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  追踪要求
                </label>
                <USelect
                  v-model="currentLogisticsRule.trackingRequired"
                  :options="[
                    { label: '需要', value: true },
                    { label: '不需要', value: false }
                  ]"
                />
              </div>
              <div class="sm:col-span-2">
                <UTextarea
                  v-model="currentLogisticsRule.description"
                  label="规则描述"
                  placeholder="请输入规则描述（可选）"
                  :rows="2"
                />
              </div>
              <div>
                <UCheckbox v-model="currentLogisticsRule.enabled" label="是否启用" />
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeLogisticsRuleModal">取消</UButton>
          <UButton color="primary" @click="saveLogisticsRule">保存</UButton>
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
const addressSearchQuery = ref("")
const logisticsSearchQuery = ref("")

// 分页
const reviewCurrentPage = ref(1)
const reviewPageSize = ref(10)

const addressCurrentPage = ref(1)
const addressPageSize = ref(10)

const logisticsCurrentPage = ref(1)
const logisticsPageSize = ref(10)

// 配置状态
const processStatus = ref("enabled")
const defaultReturnAddress = ref("")
const instructions = ref("请按照以下流程进行退换货操作，我们将尽快处理您的申请。")

// 模态框
const showReviewRuleModal = ref(false)
const showReturnAddressModal = ref(false)
const showLogisticsRuleModal = ref(false)

// 当前编辑的数据
const editingReviewRule = ref(false)
const currentReviewRule = ref({
  id: "",
  name: "",
  category: "",
  returnPeriod: 7,
  noReasonReturn: true,
  description: "",
  enabled: true
})

const editingReturnAddress = ref(false)
const currentReturnAddress = ref({
  id: "",
  name: "",
  contact: "",
  phone: "",
  province: "",
  city: "",
  district: "",
  address: "",
  postcode: "",
  applicableRegions: "",
  enabled: true
})

const editingLogisticsRule = ref(false)
const currentLogisticsRule = ref({
  id: "",
  name: "",
  carrier: "",
  freightPayer: "customer",
  trackingRequired: true,
  description: "",
  enabled: true
})

// 错误信息
const reviewRuleErrors = ref({})
const returnAddressErrors = ref({})
const logisticsRuleErrors = ref({})

// Tabs
const tabs = [
  { label: "全局配置", value: "global" },
  { label: "审核规则", value: "review" },
  { label: "退货地址", value: "address" },
  { label: "物流规则", value: "logistics" },
]

// 模拟数据
const reviewRules = ref([
  {
    id: "1",
    name: "7天无理由退货",
    category: "",
    returnPeriod: 7,
    noReasonReturn: true,
    description: "支持7天内无理由退货",
    enabled: true
  },
  {
    id: "2",
    name: "15天质量问题退货",
    category: "电子产品",
    returnPeriod: 15,
    noReasonReturn: false,
    description: "电子产品支持15天内质量问题退货",
    enabled: true
  }
])

const returnAddresses = ref([
  {
    id: "1",
    name: "华东退货中心",
    contact: "张经理",
    phone: "021-12345678",
    province: "上海市",
    city: "上海市",
    district: "浦东新区",
    address: "张江高科技园区XX路XX号",
    postcode: "201203",
    applicableRegions: "上海,江苏,浙江,安徽",
    enabled: true
  },
  {
    id: "2",
    name: "华北退货中心",
    contact: "李经理",
    phone: "010-87654321",
    province: "北京市",
    city: "北京市",
    district: "朝阳区",
    address: "望京XX大厦XX层",
    postcode: "100102",
    applicableRegions: "北京,天津,河北,山西,内蒙古",
    enabled: true
  }
])

const logisticsRules = ref([
  {
    id: "1",
    name: "标准退货物流",
    carrier: "顺丰速运",
    freightPayer: "customer",
    trackingRequired: true,
    description: "使用顺丰速运，运费由客户承担",
    enabled: true
  },
  {
    id: "2",
    name: "VIP客户物流",
    carrier: "京东物流",
    freightPayer: "merchant",
    trackingRequired: true,
    description: "VIP客户使用京东物流，运费由商家承担",
    enabled: true
  }
])

// 过滤
const filteredReviewRules = computed(() => {
  if (!reviewSearchQuery.value) return reviewRules.value
  const q = reviewSearchQuery.value.toLowerCase()
  return reviewRules.value.filter(rule =>
    rule.name.toLowerCase().includes(q) ||
    rule.category.toLowerCase().includes(q) ||
    rule.description.toLowerCase().includes(q)
  )
})

const filteredReturnAddresses = computed(() => {
  if (!addressSearchQuery.value) return returnAddresses.value
  const q = addressSearchQuery.value.toLowerCase()
  return returnAddresses.value.filter(addr =>
    addr.name.toLowerCase().includes(q) ||
    addr.contact.toLowerCase().includes(q) ||
    addr.phone.includes(q) ||
    addr.province.includes(q) ||
    addr.city.includes(q) ||
    addr.district.includes(q) ||
    addr.address.toLowerCase().includes(q) ||
    (addr.applicableRegions && addr.applicableRegions.toLowerCase().includes(q))
  )
})

const filteredLogisticsRules = computed(() => {
  if (!logisticsSearchQuery.value) return logisticsRules.value
  const q = logisticsSearchQuery.value.toLowerCase()
  return logisticsRules.value.filter(rule =>
    rule.name.toLowerCase().includes(q) ||
    rule.carrier.toLowerCase().includes(q) ||
    rule.description.toLowerCase().includes(q)
  )
})

// 分页数据
const reviewPageData = computed(() => {
  const start = (reviewCurrentPage.value - 1) * reviewPageSize.value
  const end = start + reviewPageSize.value
  return filteredReviewRules.value.slice(start, end)
})

const addressPageData = computed(() => {
  const start = (addressCurrentPage.value - 1) * addressPageSize.value
  const end = start + addressPageSize.value
  return filteredReturnAddresses.value.slice(start, end)
})

const logisticsPageData = computed(() => {
  const start = (logisticsCurrentPage.value - 1) * logisticsPageSize.value
  const end = start + logisticsPageSize.value
  return filteredLogisticsRules.value.slice(start, end)
})

// 页数
const reviewPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredReviewRules.value.length / reviewPageSize.value))
)

const addressPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredReturnAddresses.value.length / addressPageSize.value))
)

const logisticsPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredLogisticsRules.value.length / logisticsPageSize.value))
)

// 选项数据
const categoryOptions = [
  { label: "所有类目", value: "" },
  { label: "电子产品", value: "电子产品" },
  { label: "服装鞋帽", value: "服装鞋帽" },
  { label: "家居用品", value: "家居用品" },
  { label: "食品饮料", value: "食品饮料" }
]

const provinceOptions = [
  { label: "北京市", value: "北京市" },
  { label: "上海市", value: "上海市" },
  { label: "广东省", value: "广东省" },
  { label: "江苏省", value: "江苏省" },
  { label: "浙江省", value: "浙江省" }
]

const cityOptions = [
  { label: "请选择省份", value: "" }
]

const districtOptions = [
  { label: "请选择城市", value: "" }
]

const carrierOptions = [
  { label: "顺丰速运", value: "顺丰速运" },
  { label: "京东物流", value: "京东物流" },
  { label: "中通快递", value: "中通快递" },
  { label: "圆通速递", value: "圆通速递" },
  { label: "韵达快递", value: "韵达快递" }
]

const freightPayerOptions = [
  { label: "客户承担", value: "customer" },
  { label: "商家承担", value: "merchant" }
]

// 计算属性
const returnAddressOptions = computed(() => {
  return returnAddresses.value.map(addr => ({
    id: addr.id,
    name: `${addr.name} (${addr.contact})`
  }))
})

// 重置筛选
const resetReviewFilters = () => {
  reviewSearchQuery.value = ""
  reviewCurrentPage.value = 1
}

const resetAddressFilters = () => {
  addressSearchQuery.value = ""
  addressCurrentPage.value = 1
}

const resetLogisticsFilters = () => {
  logisticsSearchQuery.value = ""
  logisticsCurrentPage.value = 1
}

// 审核规则操作
const addReviewRule = () => {
  editingReviewRule.value = false
  currentReviewRule.value = {
    id: "",
    name: "",
    category: "",
    returnPeriod: 7,
    noReasonReturn: true,
    description: "",
    enabled: true
  }
  reviewRuleErrors.value = {}
  showReviewRuleModal.value = true
}

const editReviewRule = (rule) => {
  editingReviewRule.value = true
  currentReviewRule.value = { ...rule }
  reviewRuleErrors.value = {}
  showReviewRuleModal.value = true
}

const removeReviewRule = (index) => {
  if (confirm("确定要删除这个审核规则吗？")) {
    reviewRules.value.splice(index, 1)
  }
}

const closeReviewRuleModal = () => {
  showReviewRuleModal.value = false
  currentReviewRule.value = {
    id: "",
    name: "",
    category: "",
    returnPeriod: 7,
    noReasonReturn: true,
    description: "",
    enabled: true
  }
  reviewRuleErrors.value = {}
}

const saveReviewRule = () => {
  // 验证表单
  reviewRuleErrors.value = {}
  
  if (!currentReviewRule.value.name.trim()) {
    reviewRuleErrors.value.name = "规则名称不能为空"
  }
  
  if (!currentReviewRule.value.returnPeriod || currentReviewRule.value.returnPeriod <= 0) {
    reviewRuleErrors.value.returnPeriod = "退货期限必须大于0"
  }
  
  if (Object.keys(reviewRuleErrors.value).length > 0) {
    return
  }

  if (editingReviewRule.value) {
    const index = reviewRules.value.findIndex(r => r.id === currentReviewRule.value.id)
    if (index !== -1) {
      reviewRules.value[index] = { ...currentReviewRule.value }
    }
  } else {
    currentReviewRule.value.id = `rule_${Date.now()}`
    reviewRules.value.push({ ...currentReviewRule.value })
  }

  closeReviewRuleModal()
}

// 退货地址操作
const addReturnAddress = () => {
  editingReturnAddress.value = false
  currentReturnAddress.value = {
    id: "",
    name: "",
    contact: "",
    phone: "",
    province: "",
    city: "",
    district: "",
    address: "",
    postcode: "",
    applicableRegions: "",
    enabled: true
  }
  returnAddressErrors.value = {}
  showReturnAddressModal.value = true
}

const editReturnAddress = (address) => {
  editingReturnAddress.value = true
  currentReturnAddress.value = { ...address }
  returnAddressErrors.value = {}
  showReturnAddressModal.value = true
}

const removeReturnAddress = (index) => {
  if (confirm("确定要删除这个退货地址吗？")) {
    returnAddresses.value.splice(index, 1)
  }
}

const closeReturnAddressModal = () => {
  showReturnAddressModal.value = false
  currentReturnAddress.value = {
    id: "",
    name: "",
    contact: "",
    phone: "",
    province: "",
    city: "",
    district: "",
    address: "",
    postcode: "",
    applicableRegions: "",
    enabled: true
  }
  returnAddressErrors.value = {}
}

const saveReturnAddress = () => {
  // 验证表单
  returnAddressErrors.value = {}
  
  if (!currentReturnAddress.value.name.trim()) {
    returnAddressErrors.value.name = "地址名称不能为空"
  }
  
  if (!currentReturnAddress.value.contact.trim()) {
    returnAddressErrors.value.contact = "联系人不能为空"
  }
  
  if (!currentReturnAddress.value.phone.trim()) {
    returnAddressErrors.value.phone = "联系电话不能为空"
  }
  
  if (!currentReturnAddress.value.address.trim()) {
    returnAddressErrors.value.address = "详细地址不能为空"
  }
  
  if (Object.keys(returnAddressErrors.value).length > 0) {
    return
  }

  if (editingReturnAddress.value) {
    const index = returnAddresses.value.findIndex(a => a.id === currentReturnAddress.value.id)
    if (index !== -1) {
      returnAddresses.value[index] = { ...currentReturnAddress.value }
    }
  } else {
    currentReturnAddress.value.id = `addr_${Date.now()}`
    returnAddresses.value.push({ ...currentReturnAddress.value })
  }

  closeReturnAddressModal()
}

// 物流规则操作
const addLogisticsRule = () => {
  editingLogisticsRule.value = false
  currentLogisticsRule.value = {
    id: "",
    name: "",
    carrier: "",
    freightPayer: "customer",
    trackingRequired: true,
    description: "",
    enabled: true
  }
  logisticsRuleErrors.value = {}
  showLogisticsRuleModal.value = true
}

const editLogisticsRule = (rule) => {
  editingLogisticsRule.value = true
  currentLogisticsRule.value = { ...rule }
  logisticsRuleErrors.value = {}
  showLogisticsRuleModal.value = true
}

const removeLogisticsRule = (index) => {
  if (confirm("确定要删除这个物流规则吗？")) {
    logisticsRules.value.splice(index, 1)
  }
}

const closeLogisticsRuleModal = () => {
  showLogisticsRuleModal.value = false
  currentLogisticsRule.value = {
    id: "",
    name: "",
    carrier: "",
    freightPayer: "customer",
    trackingRequired: true,
    description: "",
    enabled: true
  }
  logisticsRuleErrors.value = {}
}

const saveLogisticsRule = () => {
  // 验证表单
  logisticsRuleErrors.value = {}
  
  if (!currentLogisticsRule.value.name.trim()) {
    logisticsRuleErrors.value.name = "规则名称不能为空"
  }
  
  if (!currentLogisticsRule.value.carrier) {
    logisticsRuleErrors.value.carrier = "请选择承运商"
  }
  
  if (Object.keys(logisticsRuleErrors.value).length > 0) {
    return
  }

  if (editingLogisticsRule.value) {
    const index = logisticsRules.value.findIndex(r => r.id === currentLogisticsRule.value.id)
    if (index !== -1) {
      logisticsRules.value[index] = { ...currentLogisticsRule.value }
    }
  } else {
    currentLogisticsRule.value.id = `log_${Date.now()}`
    logisticsRules.value.push({ ...currentLogisticsRule.value })
  }

  closeLogisticsRuleModal()
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

const exportConfig = () => {
  alert("配置已导出")
}
</script>