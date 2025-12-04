<template>
  <div class="p-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">退换货申请管理</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理客户的退换货申请</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-arrow-down-tray" @click="exportData">导出数据</UButton>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon name="i-heroicons-clock" class="w-6 h-6 text-yellow-600 dark:text-yellow-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">待审核</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ pendingCount }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-arrow-path" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">处理中</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ processingCount }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-check-circle" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">已完成</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ completedCount }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 标签页 -->
    <div class="mb-6">
      <UTabs v-model="activeTab" :items="tabs" />
    </div>

    <!-- 待审核 -->
    <UCard v-if="activeTab === 'pending'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">待审核</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="pendingSearchQuery" 
              placeholder="搜索申请..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton color="neutral" @click="resetPendingFilters">重置</UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请编号</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">客户信息</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">商品信息</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请类型</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请时间</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="application in pendingPageData" :key="application.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ application.applicationId }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <UAvatar :src="application.customer.avatar" :alt="application.customer.name" size="sm" class="mr-2" />
                  <div>
                    <div class="text-sm font-medium text-gray-900 dark:text-white">{{ application.customer.name }}</div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">{{ application.customer.phone }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm font-medium text-gray-900 dark:text-white">{{ application.product.name }}</div>
                <div class="text-sm text-gray-500 dark:text-gray-400">SKU: {{ application.product.sku }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="getApplicationTypeColor(application.type)" variant="soft">
                  {{ getApplicationTypeName(application.type) }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ formatDate(application.createdAt) }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-eye"
                    @click="viewApplication(application)"
                  >
                    查看
                  </UButton>
                  <UButton
                    color="success"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-check"
                    @click="approveApplication(application)"
                  >
                    通过
                  </UButton>
                  <UButton
                    color="red"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-x-mark"
                    @click="rejectApplication(application)"
                  >
                    拒绝
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
            显示第 {{ (pendingCurrentPage - 1) * pendingPageSize + 1 }}
            到 {{ Math.min(pendingCurrentPage * pendingPageSize, filteredPendingApplications.length) }} 条，共 {{ filteredPendingApplications.length }} 条
          </div>
          <UPagination
            v-model="pendingCurrentPage"
            :page-count="pendingPageCount"
            :total="filteredPendingApplications.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 处理中 -->
    <UCard v-if="activeTab === 'processing'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">处理中</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="processingSearchQuery" 
              placeholder="搜索申请..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton color="neutral" @click="resetProcessingFilters">重置</UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请编号</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">客户信息</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">商品信息</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请类型</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">处理进度</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请时间</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="application in processingPageData" :key="application.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ application.applicationId }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <UAvatar :src="application.customer.avatar" :alt="application.customer.name" size="sm" class="mr-2" />
                  <div>
                    <div class="text-sm font-medium text-gray-900 dark:text-white">{{ application.customer.name }}</div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">{{ application.customer.phone }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm font-medium text-gray-900 dark:text-white">{{ application.product.name }}</div>
                <div class="text-sm text-gray-500 dark:text-gray-400">SKU: {{ application.product.sku }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="getApplicationTypeColor(application.type)" variant="soft">
                  {{ getApplicationTypeName(application.type) }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <UProgress :value="application.progress" class="w-24 mr-2" />
                  <span class="text-sm text-gray-900 dark:text-white">{{ application.progress }}%</span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ formatDate(application.createdAt) }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <UButton
                  color="neutral"
                  variant="ghost"
                  size="sm"
                  icon="i-heroicons-eye"
                  @click="viewApplication(application)"
                >
                  查看
                </UButton>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (processingCurrentPage - 1) * processingPageSize + 1 }}
            到 {{ Math.min(processingCurrentPage * processingPageSize, filteredProcessingApplications.length) }} 条，共 {{ filteredProcessingApplications.length }} 条
          </div>
          <UPagination
            v-model="processingCurrentPage"
            :page-count="processingPageCount"
            :total="filteredProcessingApplications.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 已完成 -->
    <UCard v-if="activeTab === 'completed'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">已完成</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="completedSearchQuery" 
              placeholder="搜索申请..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton color="neutral" @click="resetCompletedFilters">重置</UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请编号</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">客户信息</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">商品信息</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">申请类型</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">处理结果</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">完成时间</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="application in completedPageData" :key="application.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ application.applicationId }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <UAvatar :src="application.customer.avatar" :alt="application.customer.name" size="sm" class="mr-2" />
                  <div>
                    <div class="text-sm font-medium text-gray-900 dark:text-white">{{ application.customer.name }}</div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">{{ application.customer.phone }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm font-medium text-gray-900 dark:text-white">{{ application.product.name }}</div>
                <div class="text-sm text-gray-500 dark:text-gray-400">SKU: {{ application.product.sku }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="getApplicationTypeColor(application.type)" variant="soft">
                  {{ getApplicationTypeName(application.type) }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="getResultColor(application.result)" variant="soft">
                  {{ getResultText(application.result) }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ formatDate(application.completedAt) }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <UButton
                  color="neutral"
                  variant="ghost"
                  size="sm"
                  icon="i-heroicons-eye"
                  @click="viewApplication(application)"
                >
                  查看
                </UButton>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (completedCurrentPage - 1) * completedPageSize + 1 }}
            到 {{ Math.min(completedCurrentPage * completedPageSize, filteredCompletedApplications.length) }} 条，共 {{ filteredCompletedApplications.length }} 条
          </div>
          <UPagination
            v-model="completedCurrentPage"
            :page-count="completedPageCount"
            :total="filteredCompletedApplications.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 申请详情模态框 -->
    <UModal
      v-model:open="showDetailModal"
      title="退换货申请详情"
      description="查看退换货申请的详细信息"
      :close="{ onClick: () => closeDetailModal() }"
      :ui="{
        content: 'w-full sm:max-w-3xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-6 p-4 sm:p-6">
            <!-- 基本信息 -->
            <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">基本信息</h3>
              
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <span class="text-gray-500 dark:text-gray-400">申请编号：</span>
                  <span class="text-gray-900 dark:text-white font-medium">{{ currentApplication?.applicationId }}</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">申请类型：</span>
                  <UBadge :color="getApplicationTypeColor(currentApplication?.type)" variant="soft">
                    {{ getApplicationTypeName(currentApplication?.type) }}
                  </UBadge>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">申请时间：</span>
                  <span class="text-gray-900 dark:text-white">{{ formatDate(currentApplication?.createdAt) }}</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">状态：</span>
                  <UBadge :color="getStatusColor(currentApplication?.status)" variant="soft">
                    {{ getStatusText(currentApplication?.status) }}
                  </UBadge>
                </div>
              </div>
            </div>

            <!-- 客户信息 -->
            <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">客户信息</h3>
              
              <div class="flex items-center gap-3 mb-4">
                <UAvatar :src="currentApplication?.customer?.avatar" :alt="currentApplication?.customer?.name" size="md" />
                <div>
                  <div class="font-medium text-gray-900 dark:text-white">{{ currentApplication?.customer?.name }}</div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">{{ currentApplication?.customer?.phone }}</div>
                </div>
              </div>
              
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <span class="text-gray-500 dark:text-gray-400">邮箱：</span>
                  <span class="text-gray-900 dark:text-white">{{ currentApplication?.customer?.email }}</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">客户ID：</span>
                  <span class="text-gray-900 dark:text-white">{{ currentApplication?.customer?.id }}</span>
                </div>
              </div>
            </div>

            <!-- 商品信息 -->
            <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">商品信息</h3>
              
              <div class="flex items-center gap-3 mb-4">
                <img :src="currentApplication?.product?.image" :alt="currentApplication?.product?.name" class="w-16 h-16 object-cover rounded" />
                <div>
                  <div class="font-medium text-gray-900 dark:text-white">{{ currentApplication?.product?.name }}</div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">SKU: {{ currentApplication?.product?.sku }}</div>
                </div>
              </div>
              
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <span class="text-gray-500 dark:text-gray-400">订单编号：</span>
                  <span class="text-gray-900 dark:text-white">{{ currentApplication?.order?.id }}</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">订单金额：</span>
                  <span class="text-gray-900 dark:text-white">¥{{ currentApplication?.order?.amount }}</span>
                </div>
              </div>
            </div>

            <!-- 申请详情 -->
            <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">申请详情</h3>
              
              <div class="space-y-3">
                <div>
                  <span class="text-gray-500 dark:text-gray-400">退换货原因：</span>
                  <span class="text-gray-900 dark:text-white">{{ currentApplication?.reason }}</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">问题描述：</span>
                  <span class="text-gray-900 dark:text-white">{{ currentApplication?.description }}</span>
                </div>
                <div v-if="currentApplication?.photos?.length">
                  <span class="text-gray-500 dark:text-gray-400">相关照片：</span>
                  <div class="flex flex-wrap gap-2 mt-2">
                    <img 
                      v-for="(photo, index) in currentApplication?.photos" 
                      :key="index" 
                      :src="photo" 
                      :alt="`照片${index + 1}`" 
                      class="w-16 h-16 object-cover rounded cursor-pointer"
                      @click="viewPhoto(photo)"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- 处理记录 -->
            <div v-if="currentApplication?.logs?.length">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">处理记录</h3>
              
              <div class="space-y-3">
                <div 
                  v-for="(log, index) in currentApplication?.logs" 
                  :key="index" 
                  class="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded"
                >
                  <div class="flex-shrink-0 w-2 h-2 bg-blue-500 rounded-full mt-2"></div>
                  <div>
                    <div class="text-sm font-medium text-gray-900 dark:text-white">{{ log.action }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ formatDate(log.timestamp) }} by {{ log.operator }}</div>
                    <div v-if="log.remark" class="text-sm text-gray-700 dark:text-gray-300 mt-1">{{ log.remark }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeDetailModal">关闭</UButton>
        </div>
      </template>
    </UModal>

    <!-- 拒绝申请模态框 -->
    <UModal
      v-model:open="showRejectModal"
      title="拒绝退换货申请"
      description="请输入拒绝申请的原因"
      :close="{ onClick: () => closeRejectModal() }"
      :ui="{
        content: 'w-full sm:max-w-md',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-4 p-4 sm:p-6">
            <div>
              <UTextarea
                v-model="rejectReason"
                label="拒绝原因"
                placeholder="请输入拒绝申请的具体原因"
                :rows="4"
              />
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeRejectModal">取消</UButton>
          <UButton color="red" @click="confirmReject">确认拒绝</UButton>
        </div>
      </template>
    </UModal>

    <!-- 照片查看模态框 -->
    <UModal
      v-model:open="showPhotoModal"
      title="查看照片"
      :close="{ onClick: () => closePhotoModal() }"
      :ui="{
        content: 'w-full sm:max-w-2xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <div class="p-4 sm:p-6">
          <img :src="currentPhoto" :alt="currentPhoto" class="w-full h-auto rounded" />
        </div>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closePhotoModal">关闭</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
// 标签页
const activeTab = ref("pending")

// 搜索查询
const pendingSearchQuery = ref("")
const processingSearchQuery = ref("")
const completedSearchQuery = ref("")

// 分页
const pendingCurrentPage = ref(1)
const pendingPageSize = ref(10)

const processingCurrentPage = ref(1)
const processingPageSize = ref(10)

const completedCurrentPage = ref(1)
const completedPageSize = ref(10)

// 模态框
const showDetailModal = ref(false)
const showRejectModal = ref(false)
const showPhotoModal = ref(false)

// 当前查看的数据
const currentApplication = ref(null)
const currentPhoto = ref("")

// 拒绝原因
const rejectReason = ref("")

// Tabs
const tabs = [
  { label: "待审核", value: "pending" },
  { label: "处理中", value: "processing" },
  { label: "已完成", value: "completed" },
]

// 统计数据
const pendingCount = ref(12)
const processingCount = ref(8)
const completedCount = ref(45)

// 模拟数据
const pendingApplications = ref([
  {
    id: "1",
    applicationId: "RMA20230915001",
    type: "RETURN",
    status: "pending",
    reason: "商品质量问题",
    description: "收到的商品有明显的划痕和损坏",
    photos: [
      "https://picsum.photos/200/200?random=1",
      "https://picsum.photos/200/200?random=2"
    ],
    customer: {
      id: "cust_001",
      name: "张三",
      phone: "13800138001",
      email: "zhangsan@example.com",
      avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=1"
    },
    product: {
      name: "无线蓝牙耳机",
      sku: "SKU001",
      image: "https://picsum.photos/100/100?random=3"
    },
    order: {
      id: "ORD20230910001",
      amount: 299.00
    },
    createdAt: "2023-09-15T14:30:25"
  },
  {
    id: "2",
    applicationId: "RMA20230915002",
    type: "EXCHANGE",
    status: "pending",
    reason: "尺寸不合适",
    description: "购买的T恤尺码偏小，需要更换大一号的",
    photos: [
      "https://picsum.photos/200/200?random=4"
    ],
    customer: {
      id: "cust_002",
      name: "李四",
      phone: "13800138002",
      email: "lisi@example.com",
      avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=2"
    },
    product: {
      name: "纯棉T恤",
      sku: "SKU002",
      image: "https://picsum.photos/100/100?random=5"
    },
    order: {
      id: "ORD20230912002",
      amount: 89.00
    },
    createdAt: "2023-09-15T16:45:12"
  }
])

const processingApplications = ref([
  {
    id: "3",
    applicationId: "RMA20230910001",
    type: "RETURN",
    status: "processing",
    progress: 60,
    reason: "商品质量问题",
    description: "商品在运输过程中损坏",
    photos: [
      "https://picsum.photos/200/200?random=6",
      "https://picsum.photos/200/200?random=7"
    ],
    customer: {
      id: "cust_003",
      name: "王五",
      phone: "13800138003",
      email: "wangwu@example.com",
      avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=3"
    },
    product: {
      name: "智能手表",
      sku: "SKU003",
      image: "https://picsum.photos/100/100?random=8"
    },
    order: {
      id: "ORD20230905003",
      amount: 1299.00
    },
    createdAt: "2023-09-10T10:15:30"
  }
])

const completedApplications = ref([
  {
    id: "4",
    applicationId: "RMA20230901001",
    type: "RETURN",
    status: "completed",
    result: "approved",
    reason: "商品质量问题",
    description: "商品存在功能缺陷",
    photos: [
      "https://picsum.photos/200/200?random=9"
    ],
    customer: {
      id: "cust_004",
      name: "赵六",
      phone: "13800138004",
      email: "zhaoliu@example.com",
      avatar: "https://api.dicebear.com/7.x/miniavs/svg?seed=4"
    },
    product: {
      name: "蓝牙音箱",
      sku: "SKU004",
      image: "https://picsum.photos/100/100?random=10"
    },
    order: {
      id: "ORD20230825004",
      amount: 199.00
    },
    createdAt: "2023-09-01T09:30:15",
    completedAt: "2023-09-05T14:20:30",
    logs: [
      {
        action: "申请提交",
        timestamp: "2023-09-01T09:30:15",
        operator: "张三",
        remark: "客户提交退货申请"
      },
      {
        action: "审核通过",
        timestamp: "2023-09-02T11:15:22",
        operator: "李四",
        remark: "确认商品质量问题，同意退货"
      },
      {
        action: "退货完成",
        timestamp: "2023-09-05T14:20:30",
        operator: "王五",
        remark: "收到退货商品，已完成退款"
      }
    ]
  }
])

// 过滤
const filteredPendingApplications = computed(() => {
  if (!pendingSearchQuery.value) return pendingApplications.value
  const q = pendingSearchQuery.value.toLowerCase()
  return pendingApplications.value.filter(app =>
    app.applicationId.toLowerCase().includes(q) ||
    app.customer.name.toLowerCase().includes(q) ||
    app.customer.phone.includes(q) ||
    app.product.name.toLowerCase().includes(q)
  )
})

const filteredProcessingApplications = computed(() => {
  if (!processingSearchQuery.value) return processingApplications.value
  const q = processingSearchQuery.value.toLowerCase()
  return processingApplications.value.filter(app =>
    app.applicationId.toLowerCase().includes(q) ||
    app.customer.name.toLowerCase().includes(q) ||
    app.customer.phone.includes(q) ||
    app.product.name.toLowerCase().includes(q)
  )
})

const filteredCompletedApplications = computed(() => {
  if (!completedSearchQuery.value) return completedApplications.value
  const q = completedSearchQuery.value.toLowerCase()
  return completedApplications.value.filter(app =>
    app.applicationId.toLowerCase().includes(q) ||
    app.customer.name.toLowerCase().includes(q) ||
    app.customer.phone.includes(q) ||
    app.product.name.toLowerCase().includes(q)
  )
})

// 分页数据
const pendingPageData = computed(() => {
  const start = (pendingCurrentPage.value - 1) * pendingPageSize.value
  const end = start + pendingPageSize.value
  return filteredPendingApplications.value.slice(start, end)
})

const processingPageData = computed(() => {
  const start = (processingCurrentPage.value - 1) * processingPageSize.value
  const end = start + processingPageSize.value
  return filteredProcessingApplications.value.slice(start, end)
})

const completedPageData = computed(() => {
  const start = (completedCurrentPage.value - 1) * completedPageSize.value
  const end = start + completedPageSize.value
  return filteredCompletedApplications.value.slice(start, end)
})

// 页数
const pendingPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredPendingApplications.value.length / pendingPageSize.value))
)

const processingPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredProcessingApplications.value.length / processingPageSize.value))
)

const completedPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredCompletedApplications.value.length / completedPageSize.value))
)

// 重置筛选
const resetPendingFilters = () => {
  pendingSearchQuery.value = ""
  pendingCurrentPage.value = 1
}

const resetProcessingFilters = () => {
  processingSearchQuery.value = ""
  processingCurrentPage.value = 1
}

const resetCompletedFilters = () => {
  completedSearchQuery.value = ""
  completedCurrentPage.value = 1
}

// 获取申请类型名称和颜色
const getApplicationTypeName = (type) => {
  const typeMap = {
    RETURN: "退货",
    EXCHANGE: "换货",
    REPAIR: "维修"
  }
  return typeMap[type] || type
}

const getApplicationTypeColor = (type) => {
  const colorMap = {
    RETURN: "red",
    EXCHANGE: "blue",
    REPAIR: "yellow"
  }
  return colorMap[type] || "neutral"
}

// 获取状态文本和颜色
const getStatusText = (status) => {
  const statusMap = {
    pending: "待审核",
    processing: "处理中",
    completed: "已完成"
  }
  return statusMap[status] || status
}

const getStatusColor = (status) => {
  const colorMap = {
    pending: "yellow",
    processing: "blue",
    completed: "green"
  }
  return colorMap[status] || "neutral"
}

// 获取结果文本和颜色
const getResultText = (result) => {
  const resultMap = {
    approved: "已批准",
    rejected: "已拒绝",
    completed: "已完成"
  }
  return resultMap[result] || result
}

const getResultColor = (result) => {
  const colorMap = {
    approved: "green",
    rejected: "red",
    completed: "green"
  }
  return colorMap[result] || "neutral"
}

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return "-"
  const date = new Date(dateString)
  return date.toLocaleString("zh-CN")
}

// 操作函数
const viewApplication = (application) => {
  currentApplication.value = application
  showDetailModal.value = true
}

const approveApplication = (application) => {
  if (confirm(`确定要通过申请 ${application.applicationId} 吗？`)) {
    // 在实际应用中，这里应该调用API进行审核通过操作
    alert(`已通过申请 ${application.applicationId}`)
    
    // 更新统计数据
    pendingCount.value--
    processingCount.value++
    
    // 从待审核列表中移除
    const index = pendingApplications.value.findIndex(app => app.id === application.id)
    if (index !== -1) {
      pendingApplications.value.splice(index, 1)
    }
    
    // 添加到处理中列表
    processingApplications.value.push({
      ...application,
      status: "processing",
      progress: 20
    })
  }
}

const rejectApplication = (application) => {
  currentApplication.value = application
  rejectReason.value = ""
  showRejectModal.value = true
}

const viewPhoto = (photo) => {
  currentPhoto.value = photo
  showPhotoModal.value = true
}

// 模态框操作
const closeDetailModal = () => {
  showDetailModal.value = false
  currentApplication.value = null
}

const closeRejectModal = () => {
  showRejectModal.value = false
  currentApplication.value = null
  rejectReason.value = ""
}

const closePhotoModal = () => {
  showPhotoModal.value = false
  currentPhoto.value = ""
}

// 确认拒绝
const confirmReject = () => {
  if (!rejectReason.value.trim()) {
    alert("请输入拒绝原因")
    return
  }

  if (currentApplication.value) {
    // 在实际应用中，这里应该调用API进行拒绝操作
    alert(`已拒绝申请 ${currentApplication.value.applicationId}，原因：${rejectReason.value}`)
    
    // 更新统计数据
    pendingCount.value--
    completedCount.value++
    
    // 从待审核列表中移除
    const index = pendingApplications.value.findIndex(app => app.id === currentApplication.value.id)
    if (index !== -1) {
      pendingApplications.value.splice(index, 1)
    }
    
    // 添加到已完成列表
    completedApplications.value.push({
      ...currentApplication.value,
      status: "completed",
      result: "rejected",
      completedAt: new Date().toISOString(),
      logs: [
        ...(currentApplication.value.logs || []),
        {
          action: "审核拒绝",
          timestamp: new Date().toISOString(),
          operator: "管理员",
          remark: rejectReason.value
        }
      ]
    })
  }

  closeRejectModal()
}

// 导出数据
const exportData = () => {
  alert("导出数据功能待实现")
}
</script>