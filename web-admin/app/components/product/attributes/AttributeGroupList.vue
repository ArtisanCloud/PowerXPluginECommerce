<template>
  <UCard>
    <template #header>
      <div class="flex justify-between items-center">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">属性组管理</h2>
        <div class="flex items-center gap-2">
          <UButton icon="i-heroicons-plus" color="primary" @click="openCreateModal">新建属性组</UButton>
          <UButton icon="i-heroicons-arrow-up-on-square" color="gray">导入</UButton>
          <UButton icon="i-heroicons-arrow-down-on-square" color="gray">导出</UButton>
        </div>
      </div>
    </template>

    <!-- Filters -->
    <div class="flex items-center gap-4 mb-4">
      <UInput v-model="filters.keyword" placeholder="名称/编码..." class="flex-1" />
      <USelectMenu v-model="filters.status" :options="statusOptions" multiple placeholder="状态" />
      <USelectMenu v-model="filters.scope" :options="scopeOptions" multiple placeholder="作用域" />
    </div>

    <!-- Table -->
    <UTable :columns="columns" :rows="filteredRows" :loading="pending">
      <template #name-data="{ row }">
        <div class="font-medium">{{ row.name }}</div>
        <div class="text-xs text-gray-500">{{ row.code }}</div>
      </template>
      <template #scope-data="{ row }">
        <span>{{ row.scope }}</span>
      </template>
      <template #attributeCount-data="{ row }">
        <span class="text-sm bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">{{ row.attributeCount }} 个属性</span>
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
    <CreateAttributeGroupModal 
      v-model:open="isCreateModalOpen" 
      :group-to-copy="groupToCopy"
      :available-attributes="availableAttributes"
      @created="onGroupCreated"
    />
    <EditAttributeGroupModal 
      v-if="editingGroup"
      v-model:open="isEditModalOpen" 
      :group="editingGroup"
      :available-attributes="availableAttributes"
      @updated="onGroupUpdated"
    />
  </UCard>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import type { DropdownItem } from '@nuxt/ui/dist/runtime/types'
import CreateAttributeGroupModal from '~/components/Modals/CreateAttributeGroupModal.vue'
import EditAttributeGroupModal from '~/components/Modals/EditAttributeGroupModal.vue'

const pending = ref(false)
const isCreateModalOpen = ref(false)
const isEditModalOpen = ref(false)
const editingGroup = ref<any>(null)
const groupToCopy = ref<any>(null)

// --- Options for Selects and Filters ---
const statusOptions = ['enabled', 'disabled', 'deprecated']
const scopeOptions = ['SPU', 'SKU', 'Both']

// --- Table Columns ---
const columns = [
  { key: 'name', label: '名称/编码', id: 'name' },
  { key: 'scope', label: '作用域', id: 'scope' },
  { key: 'attributeCount', label: '属性数', id: 'attributeCount' },
  { key: 'status', label: '状态', id: 'status' },
  { key: 'updatedAt', label: '最近更新', id: 'updatedAt' },
  { key: 'actions', label: '操作', id: 'actions' }
]

// --- Mock Data ---
const availableAttributes = ref([
  { id: 1, name: '材质', code: 'material' },
  { id: 2, name: '净含量', code: 'net_weight' },
  { id: 3, name: '上市年份', code: 'release_year' },
  { id: 4, name: '商品毛重', code: 'gross_weight' },
  { id: 5, name: '颜色', code: 'color' },
  { id: 6, name: '尺寸', code: 'size' },
  { id: 7, name: '容量', code: 'capacity' },
  { id: 8, name: '重量', code: 'weight' },
])

const mockData = ref([
  { id: 1, name: '基础属性', code: 'basic_attr', scope: 'SPU', attributeCount: 4, isSearchable: true, status: 'enabled', updatedAt: '2025-09-27', sortOrder: 1, attributes: [ { id: 1, name: '材质', code: 'material', required: true, sort: 1 }, { id: 3, name: '上市年份', code: 'release_year', required: false, sort: 2 } ] },
  { id: 2, name: '规格参数', code: 'spec_params', scope: 'SKU', attributeCount: 4, isSearchable: true, status: 'enabled', updatedAt: '2025-09-26', sortOrder: 2, attributes: [ { id: 5, name: '颜色', code: 'color', required: true, sort: 1 }, { id: 6, name: '尺寸', code: 'size', required: true, sort: 2 } ] },
  { id: 3, name: '包装信息', code: 'packaging', scope: 'SPU', attributeCount: 2, isSearchable: false, status: 'disabled', updatedAt: '2025-09-25', sortOrder: 3, attributes: [ { id: 4, name: '商品毛重', code: 'gross_weight', required: false, sort: 1 } ] },
])

// --- Filtering Logic ---
const filters = reactive({
  keyword: '',
  status: [],
  scope: []
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
  if (filters.scope.length) {
    data = data.filter(item => filters.scope.includes(item.scope))
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
    { label: '删除', icon: 'i-heroicons-trash-20-solid', click: () => deleteGroup(row), class: 'text-red-500' }
  ]
]

const openCreateModal = () => {
  groupToCopy.value = null;
  isCreateModalOpen.value = true;
}

const openEditModal = (group: any) => {
  editingGroup.value = group;
  isEditModalOpen.value = true;
}

const openCopyModal = (group: any) => {
  groupToCopy.value = group;
  isCreateModalOpen.value = true;
}

const deleteGroup = (group: any) => {
  if (confirm(`确定要删除属性组 "${group.name}" 吗？`)) {
    const index = mockData.value.findIndex(item => item.id === group.id);
    if (index !== -1) {
      mockData.value.splice(index, 1);
    }
  }
}

const toggleStatus = (group: any) => {
  const index = mockData.value.findIndex(item => item.id === group.id);
  if (index !== -1) {
    mockData.value[index].status = mockData.value[index].status === 'enabled' ? 'disabled' : 'enabled';
  }
}

const onGroupCreated = (newGroup: any) => {
  mockData.value.unshift(newGroup);
  isCreateModalOpen.value = false;
}

const onGroupUpdated = (updatedGroup: any) => {
  const index = mockData.value.findIndex(item => item.id === updatedGroup.id);
  if (index !== -1) {
    mockData.value[index] = { ...mockData.value[index], ...updatedGroup, updatedAt: new Date().toISOString().split('T')[0] };
  }
  isEditModalOpen.value = false;
}

</script>
