<template>
  <div class="p-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">RMA 入口配置</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">配置客户自助退换货申请表单字段和退换货类型</p>
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
            <p class="text-sm text-gray-600 dark:text-gray-400">表单字段</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ formFields.length }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-arrow-uturn-left" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">退换货类型</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ rmaTypes.length }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon name="i-heroicons-cog-6-tooth" class="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">启用状态</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <UBadge :color="configStatus === 'enabled' ? 'success' : 'neutral'" variant="soft">
                {{ configStatus === 'enabled' ? '已启用' : '已禁用' }}
              </UBadge>
            </p>
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
            RMA入口状态
          </label>
          <USelect
            v-model="configStatus"
            :options="[
              { label: '启用', value: 'enabled' },
              { label: '禁用', value: 'disabled' }
            ]"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            控制客户是否可以访问自助退换货门户
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            申请时间限制
          </label>
          <div class="flex items-center space-x-2">
            <UInput
              v-model.number="timeLimit"
              type="number"
              min="1"
              class="w-24"
            />
            <span class="text-gray-700 dark:text-gray-300">天</span>
          </div>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            订单完成后多少天内可以申请退换货
          </p>
        </div>

        <div class="md:col-span-2">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            说明文本
          </label>
          <UTextarea
            v-model="instructions"
            placeholder="请输入退换货申请的说明文本"
            :rows="3"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            显示在申请表单顶部的说明文本
          </p>
        </div>
      </div>
    </UCard>

    <!-- 表单字段配置 -->
    <UCard v-if="activeTab === 'fields'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">表单字段配置</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="fieldSearchQuery" 
              placeholder="搜索字段..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton 
              color="primary" 
              variant="outline" 
              size="sm" 
              icon="i-heroicons-plus"
              @click="addFormField"
            >
              添加字段
            </UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">字段标签</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">字段类型</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">描述</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">必填</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="(field, index) in filteredFields" :key="field.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ field.label }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ getFieldTypeName(field.type) }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ field.description || '-' }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="field.required ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ field.required ? '是' : '否' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="field.enabled ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ field.enabled ? '启用' : '禁用' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-pencil"
                    @click="editFormField(field)"
                  >
                    编辑
                  </UButton>
                  <UButton
                    color="error"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="removeFormField(index)"
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
            显示第 {{ (fieldCurrentPage - 1) * fieldPageSize + 1 }}
            到 {{ Math.min(fieldCurrentPage * fieldPageSize, filteredFields.length) }} 条，共 {{ filteredFields.length }} 条
          </div>
          <UPagination
            v-model="fieldCurrentPage"
            :page-count="fieldPageCount"
            :total="filteredFields.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 退换货类型配置 -->
    <UCard v-if="activeTab === 'types'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">退换货类型配置</h2>
          <div class="flex gap-3">
            <UInput 
              v-model="typeSearchQuery" 
              placeholder="搜索类型..." 
              icon="i-heroicons-magnifying-glass" 
              size="sm" 
            />
            <UButton 
              color="primary" 
              variant="outline" 
              size="sm" 
              icon="i-heroicons-plus"
              @click="addRmaType"
            >
              添加类型
            </UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">类型名称</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">类型代码</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">描述</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">退货要求</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">照片要求</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">状态</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="(type, index) in filteredTypes" :key="type.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ type.name }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ type.code }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ type.description || '-' }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="type.requiresReturn ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ type.requiresReturn ? '需要' : '不需要' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="type.requiresPhotos ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ type.requiresPhotos ? '需要' : '不需要' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UBadge :color="type.enabled ? 'success' : 'neutral'" variant="soft" size="xs">
                  {{ type.enabled ? '启用' : '禁用' }}
                </UBadge>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-pencil"
                    @click="editRmaType(type)"
                  >
                    编辑
                  </UButton>
                  <UButton
                    color="error"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="removeRmaType(index)"
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
            显示第 {{ (typeCurrentPage - 1) * typePageSize + 1 }}
            到 {{ Math.min(typeCurrentPage * typePageSize, filteredTypes.length) }} 条，共 {{ filteredTypes.length }} 条
          </div>
          <UPagination
            v-model="typeCurrentPage"
            :page-count="typePageCount"
            :total="filteredTypes.length"
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

    <!-- 表单字段编辑模态框 -->
    <UModal
      v-model:open="showFieldModal"
      :title="editingField ? '编辑表单字段' : '添加表单字段'"
      description="配置退换货申请表单字段"
      :close="{ onClick: () => closeFieldModal() }"
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
                v-model="currentField.label"
                label="字段标签"
                placeholder="请输入字段标签"
                :error="!!fieldErrors.label"
              />
              <USelect
                v-model="currentField.type"
                label="字段类型"
                :options="fieldTypeOptions"
                :error="!!fieldErrors.type"
              />
              <div class="sm:col-span-2">
                <UTextarea
                  v-model="currentField.description"
                  label="字段描述"
                  placeholder="请输入字段描述（可选）"
                  :rows="2"
                />
              </div>
              <div>
                <UCheckbox v-model="currentField.required" label="是否必填" />
              </div>
              <div>
                <UCheckbox v-model="currentField.enabled" label="是否启用" />
              </div>
              <div class="sm:col-span-2" v-if="currentField.type === 'select' || currentField.type === 'radio'">
                <UTextarea
                  v-model="currentField.options"
                  label="选项值（每行一个选项）"
                  placeholder="选项1
选项2
选项3"
                  :rows="3"
                />
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  每行输入一个选项值
                </p>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeFieldModal">取消</UButton>
          <UButton color="primary" @click="saveFormField">保存</UButton>
        </div>
      </template>
    </UModal>

    <!-- 退换货类型编辑模态框 -->
    <UModal
      v-model:open="showTypeModal"
      :title="editingType ? '编辑退换货类型' : '添加退换货类型'"
      description="配置支持的退换货类型"
      :close="{ onClick: () => closeTypeModal() }"
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
                v-model="currentType.name"
                label="类型名称"
                placeholder="请输入类型名称"
                :error="!!typeErrors.name"
              />
              <UInput
                v-model="currentType.code"
                label="类型代码"
                placeholder="请输入类型代码"
                :error="!!typeErrors.code"
              />
              <div class="sm:col-span-2">
                <UTextarea
                  v-model="currentType.description"
                  label="类型描述"
                  placeholder="请输入类型描述（可选）"
                  :rows="2"
                />
              </div>
              <div>
                <UCheckbox v-model="currentType.requiresReturn" label="需要退货" />
              </div>
              <div>
                <UCheckbox v-model="currentType.requiresPhotos" label="需要照片" />
              </div>
              <div>
                <UCheckbox v-model="currentType.enabled" label="是否启用" />
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeTypeModal">取消</UButton>
          <UButton color="primary" @click="saveRmaType">保存</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
// 标签页
const activeTab = ref("global")

// 搜索查询
const fieldSearchQuery = ref("")
const typeSearchQuery = ref("")

// 分页
const fieldCurrentPage = ref(1)
const fieldPageSize = ref(10)

const typeCurrentPage = ref(1)
const typePageSize = ref(10)

// 配置状态
const configStatus = ref("enabled")
const timeLimit = ref(30)
const instructions = ref("请填写以下信息申请退换货，我们会尽快处理您的申请。")

// 表单字段
const formFields = ref([
  {
    id: "1",
    label: "退换货原因",
    type: "select",
    description: "请选择退换货的具体原因",
    required: true,
    enabled: true,
    options: "商品质量问题\n尺寸不合适\n颜色不喜欢\n其他原因"
  },
  {
    id: "2",
    label: "问题描述",
    type: "textarea",
    description: "请详细描述您遇到的问题",
    required: true,
    enabled: true
  },
  {
    id: "3",
    label: "照片上传",
    type: "file",
    description: "请上传相关照片（最多5张）",
    required: false,
    enabled: true
  },
  {
    id: "4",
    label: "SKU信息",
    type: "text",
    description: "请输入商品SKU编号",
    required: true,
    enabled: true
  }
])

// 退换货类型
const rmaTypes = ref([
  {
    id: "1",
    name: "退货",
    code: "RETURN",
    description: "申请退货退款",
    requiresReturn: true,
    requiresPhotos: true,
    enabled: true
  },
  {
    id: "2",
    name: "换货",
    code: "EXCHANGE",
    description: "申请更换商品",
    requiresReturn: true,
    requiresPhotos: true,
    enabled: true
  },
  {
    id: "3",
    name: "维修",
    code: "REPAIR",
    description: "申请商品维修",
    requiresReturn: false,
    requiresPhotos: true,
    enabled: true
  }
])

// 模态框
const showFieldModal = ref(false)
const showTypeModal = ref(false)

// 当前编辑的数据
const editingField = ref(false)
const currentField = ref({
  id: "",
  label: "",
  type: "text",
  description: "",
  required: false,
  enabled: true,
  options: ""
})

const editingType = ref(false)
const currentType = ref({
  id: "",
  name: "",
  code: "",
  description: "",
  requiresReturn: false,
  requiresPhotos: false,
  enabled: true
})

// 错误信息
const fieldErrors = ref({})
const typeErrors = ref({})

// Tabs
const tabs = [
  { label: "全局配置", value: "global" },
  { label: "表单字段", value: "fields" },
  { label: "退换货类型", value: "types" },
]

// 字段类型选项
const fieldTypeOptions = [
  { label: "文本输入", value: "text" },
  { label: "多行文本", value: "textarea" },
  { label: "下拉选择", value: "select" },
  { label: "单选框", value: "radio" },
  { label: "复选框", value: "checkbox" },
  { label: "文件上传", value: "file" },
  { label: "日期选择", value: "date" }
]

// 过滤
const filteredFields = computed(() => {
  if (!fieldSearchQuery.value) return formFields.value
  const q = fieldSearchQuery.value.toLowerCase()
  return formFields.value.filter(field =>
    field.label.toLowerCase().includes(q) ||
    field.description.toLowerCase().includes(q)
  )
})

const filteredTypes = computed(() => {
  if (!typeSearchQuery.value) return rmaTypes.value
  const q = typeSearchQuery.value.toLowerCase()
  return rmaTypes.value.filter(type =>
    type.name.toLowerCase().includes(q) ||
    type.code.toLowerCase().includes(q) ||
    type.description.toLowerCase().includes(q)
  )
})

// 分页数据
const fieldPageData = computed(() => {
  const start = (fieldCurrentPage.value - 1) * fieldPageSize.value
  const end = start + fieldPageSize.value
  return filteredFields.value.slice(start, end)
})

const typePageData = computed(() => {
  const start = (typeCurrentPage.value - 1) * typePageSize.value
  const end = start + typePageSize.value
  return filteredTypes.value.slice(start, end)
})

// 页数
const fieldPageCount = computed(() =>
  Math.max(1, Math.ceil(filteredFields.value.length / fieldPageSize.value))
)

const typePageCount = computed(() =>
  Math.max(1, Math.ceil(filteredTypes.value.length / typePageSize.value))
)

// 获取字段类型名称
const getFieldTypeName = (type) => {
  const typeMap = {
    text: "文本输入",
    textarea: "多行文本",
    select: "下拉选择",
    radio: "单选框",
    checkbox: "复选框",
    file: "文件上传",
    date: "日期选择"
  }
  return typeMap[type] || type
}

// 表单字段操作
const addFormField = () => {
  editingField.value = false
  currentField.value = {
    id: "",
    label: "",
    type: "text",
    description: "",
    required: false,
    enabled: true,
    options: ""
  }
  fieldErrors.value = {}
  showFieldModal.value = true
}

const editFormField = (field) => {
  editingField.value = true
  currentField.value = { ...field }
  fieldErrors.value = {}
  showFieldModal.value = true
}

const removeFormField = (index) => {
  if (confirm("确定要删除这个表单字段吗？")) {
    formFields.value.splice(index, 1)
  }
}

const closeFieldModal = () => {
  showFieldModal.value = false
  currentField.value = {
    id: "",
    label: "",
    type: "text",
    description: "",
    required: false,
    enabled: true,
    options: ""
  }
  fieldErrors.value = {}
}

const saveFormField = () => {
  // 验证表单
  fieldErrors.value = {}
  
  if (!currentField.value.label.trim()) {
    fieldErrors.value.label = "字段标签不能为空"
  }
  
  if (!currentField.value.type) {
    fieldErrors.value.type = "请选择字段类型"
  }
  
  if (Object.keys(fieldErrors.value).length > 0) {
    return
  }

  if (editingField.value) {
    const index = formFields.value.findIndex(f => f.id === currentField.value.id)
    if (index !== -1) {
      formFields.value[index] = { ...currentField.value }
    }
  } else {
    currentField.value.id = `field_${Date.now()}`
    formFields.value.push({ ...currentField.value })
  }

  closeFieldModal()
}

// 退换货类型操作
const addRmaType = () => {
  editingType.value = false
  currentType.value = {
    id: "",
    name: "",
    code: "",
    description: "",
    requiresReturn: false,
    requiresPhotos: false,
    enabled: true
  }
  typeErrors.value = {}
  showTypeModal.value = true
}

const editRmaType = (type) => {
  editingType.value = true
  currentType.value = { ...type }
  typeErrors.value = {}
  showTypeModal.value = true
}

const removeRmaType = (index) => {
  if (confirm("确定要删除这个退换货类型吗？")) {
    rmaTypes.value.splice(index, 1)
  }
}

const closeTypeModal = () => {
  showTypeModal.value = false
  currentType.value = {
    id: "",
    name: "",
    code: "",
    description: "",
    requiresReturn: false,
    requiresPhotos: false,
    enabled: true
  }
  typeErrors.value = {}
}

const saveRmaType = () => {
  // 验证表单
  typeErrors.value = {}
  
  if (!currentType.value.name.trim()) {
    typeErrors.value.name = "类型名称不能为空"
  }
  
  if (!currentType.value.code.trim()) {
    typeErrors.value.code = "类型代码不能为空"
  }
  
  if (Object.keys(typeErrors.value).length > 0) {
    return
  }

  if (editingType.value) {
    const index = rmaTypes.value.findIndex(t => t.id === currentType.value.id)
    if (index !== -1) {
      rmaTypes.value[index] = { ...currentType.value }
    }
  } else {
    currentType.value.id = `type_${Date.now()}`
    rmaTypes.value.push({ ...currentType.value })
  }

  closeTypeModal()
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