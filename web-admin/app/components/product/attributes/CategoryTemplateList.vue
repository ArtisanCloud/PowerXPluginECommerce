<template>
  <UCard>
    <template #header>
      <div class="flex justify-between items-center">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">类目模板列表</h2>
        <div class="flex items-center gap-2">
          <UButton icon="i-heroicons-plus" color="primary" @click="openCreateModal">新建模板</UButton>
          <UButton icon="i-heroicons-arrow-up-on-square" color="gray">导入</UButton>
          <UButton icon="i-heroicons-arrow-down-on-square" color="gray">导出</UButton>
        </div>
      </div>
    </template>

    <!-- Filters -->
    <div class="flex items-center gap-4 mb-4">
      <UInput v-model="filters.keyword" placeholder="名称/编码..." class="flex-1" />
      <USelectMenu v-model="filters.status" :options="statusOptions" multiple placeholder="状态" />
      <USelectMenu v-model="filters.category" :options="categoryOptionsForFilter" multiple placeholder="所属类目" />
    </div>

    <!-- Table -->
    <UTable :columns="columns" :rows="filteredRows" :loading="pending">
      <template #name-data="{ row }">
        <div class="font-medium">{{ row.name }}</div>
        <div class="text-xs text-gray-500">{{ row.code }}</div>
      </template>
      <template #category-data="{ row }">
        <span>{{ row.category }}</span>
      </template>
      <template #attributeGroupCount-data="{ row }">
        <span class="text-sm bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">{{ row.attributeGroupCount }} 个属性组</span>
      </template>
      <template #status-data="{ row }">
        <UBadge :color="row.status === 'enabled' ? 'primary' : 'gray'" variant="soft">{{ row.status }}</UBadge>
      </template>
      <template #actions-data="{ row }">
        <UDropdownMenu :items="getActionItems(row)">
          <UButton color="gray" variant="ghost" icon="i-heroicons-ellipsis-horizontal-20-solid" />
        </UDropdownMenu>
      </template>
    </UTable>

    <!-- Modals -->
    <CreateCategoryTemplateModal 
      v-model:open="isCreateModalOpen" 
      :template-to-copy="templateToCopy"
      :available-attribute-groups="availableAttributeGroups"
      :category-options="categoryOptions"
      @created="onTemplateCreated"
    />
    <EditCategoryTemplateModal 
      v-if="editingTemplate"
      v-model:open="isEditModalOpen" 
      :template="editingTemplate"
      :available-attribute-groups="availableAttributeGroups"
      :category-options="categoryOptions"
      @updated="onTemplateUpdated"
    />
  </UCard>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import type { DropdownItem } from '@nuxt/ui/dist/runtime/types'
import CreateCategoryTemplateModal from '~/components/modals/CreateCategoryTemplateModal.vue'
import EditCategoryTemplateModal from '~/components/modals/EditCategoryTemplateModal.vue'

const pending = ref(false)
const isCreateModalOpen = ref(false)
const isEditModalOpen = ref(false)
const editingTemplate = ref<any>(null)
const templateToCopy = ref<any>(null)

// --- Options for Selects and Filters ---
const statusOptions = ['enabled', 'disabled', 'deprecated']
const categoryOptions = [
  { label: '服装', value: '服装' },
  { label: '数码', value: '数码' },
  { label: '家居', value: '家居' },
  { label: '食品', value: '食品' }
]
const categoryOptionsForFilter = ['服装', '数码', '家居', '食品']

// --- Table Columns ---
const columns = [
  { key: 'name', label: '名称/编码', id: 'name' },
  { key: 'category', label: '所属类目', id: 'category' },
  { key: 'attributeGroupCount', label: '属性组数', id: 'attributeGroupCount' },
  { key: 'status', label: '状态', id: 'status' },
  { key: 'updatedAt', label: '最近更新', id: 'updatedAt' },
  { key: 'actions', label: '操作', id: 'actions' }
]

// --- Mock Data ---
const availableAttributeGroups = ref([
  { id: 1, name: '基础属性', code: 'basic_attr' },
  { id: 2, name: '规格参数', code: 'spec_params' },
  { id: 3, name: '包装信息', code: 'packaging' },
  { id: 4, name: '安全信息', code: 'safety' },
  { id: 5, name: '认证信息', code: 'certification' },
])

const mockData = ref([
  { id: 1, name: '服装类目模板', code: 'clothing_template', category: '服装', attributeGroupCount: 3, status: 'enabled', updatedAt: '2025-09-27', sortOrder: 1, attributeGroups: [{id: 1, name: '基础属性', code: 'basic_attr', sort: 1}] },
  { id: 2, name: '数码产品模板', code: 'electronics_template', category: '数码', attributeGroupCount: 4, status: 'enabled', updatedAt: '2025-09-26', sortOrder: 2, attributeGroups: [{id: 2, name: '规格参数', code: 'spec_params', sort: 1}] },
  { id: 3, name: '家居用品模板', code: 'home_template', category: '家居', attributeGroupCount: 2, status: 'disabled', updatedAt: '2025-09-25', sortOrder: 3, attributeGroups: [] },
  { id: 4, name: '食品类目模板', code: 'food_template', category: '食品', attributeGroupCount: 3, status: 'enabled', updatedAt: '2025-09-24', sortOrder: 4, attributeGroups: [] },
])

// --- Filtering Logic ---
const filters = reactive({
  keyword: '',
  status: [],
  category: []
})

const filteredRows = computed(() => {
  let data = mockData.value
  if (filters.keyword) {
    data = data.filter(item => 
      item.name.includes(filters.keyword) || item.code.includes(filters.keyword)
    )
  }
  if (filters.status.length) {
    data = data.filter(item => filters.status.includes(item.status))
  }
  if (filters.category.length) {
    data = data.filter(item => filters.category.includes(item.category))
  }
  return data
})

// --- Actions ---
const getActionItems = (row: any): DropdownItem[][] => [
  [
    { label: '编辑', icon: 'i-heroicons-pencil-square-20-solid', click: () => openEditModal(row) },
    { label: '复制', icon: 'i-heroicons-document-duplicate-20-solid', click: () => openCopyModal(row) }
  ],
  [
    { label: '禁用', icon: 'i-heroicons-pause-circle-20-solid', click: () => toggleStatus(row) },
    { label: '删除', icon: 'i-heroicons-trash-20-solid', click: () => deleteTemplate(row), class: 'text-red-500' }
  ]
]

const openCreateModal = () => {
  templateToCopy.value = null;
  isCreateModalOpen.value = true;
}

const openEditModal = (template: any) => {
  editingTemplate.value = template;
  isEditModalOpen.value = true;
}

const openCopyModal = (template: any) => {
  templateToCopy.value = template;
  isCreateModalOpen.value = true;
}

const deleteTemplate = (template: any) => {
  if (confirm(`确定要删除类目模板 "${template.name}" 吗？`)) {
    const index = mockData.value.findIndex(item => item.id === template.id);
    if (index !== -1) {
      mockData.value.splice(index, 1);
    }
  }
}

const toggleStatus = (template: any) => {
  const index = mockData.value.findIndex(item => item.id === template.id);
  if (index !== -1) {
    mockData.value[index].status = mockData.value[index].status === 'enabled' ? 'disabled' : 'enabled';
  }
}

const onTemplateCreated = (newTemplate: any) => {
  mockData.value.unshift(newTemplate);
  isCreateModalOpen.value = false;
}

const onTemplateUpdated = (updatedTemplate: any) => {
  const index = mockData.value.findIndex(item => item.id === updatedTemplate.id);
  if (index !== -1) {
    mockData.value[index] = { ...mockData.value[index], ...updatedTemplate, updatedAt: new Date().toISOString().split('T')[0] };
  }
  isEditModalOpen.value = false;
}

</script>