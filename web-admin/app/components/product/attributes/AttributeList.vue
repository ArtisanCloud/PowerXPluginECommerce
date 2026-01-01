<template>
  <UCard>
    <template #header>
      <div class="flex justify-between items-center">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">属性管理</h2>
        <div class="flex items-center gap-2">
          <UButton icon="i-heroicons-plus" color="primary" @click="openCreateModal">新建属性</UButton>
          <UButton icon="i-heroicons-arrow-up-on-square" color="gray">导入</UButton>
          <UButton icon="i-heroicons-arrow-down-on-square" color="gray">导出</UButton>
        </div>
      </div>
    </template>

    <!-- Filters -->
    <div class="flex items-center gap-4 mb-4">
      <UInput v-model="filters.keyword" placeholder="名称/编码..." class="flex-1" />
      <USelectMenu v-model="filters.status" :options="statusOptions" multiple placeholder="状态" />
      <USelectMenu v-model="filters.dataType" :options="dataTypeOptionsForFilter" multiple placeholder="数据类型" />
      <USelectMenu v-model="filters.scope" :options="scopeOptions" multiple placeholder="作用域" />
    </div>

    <!-- Table -->
    <UTable :columns="columns" :rows="filteredRows" :loading="pending">
      <template #name-data="{ row }">
        <div class="font-medium">{{ row.name }}</div>
        <div class="text-xs text-gray-500">{{ row.code }}</div>
      </template>
      <template #dataType-data="{ row }">
        <UBadge variant="soft">{{ row.dataType }}</UBadge>
      </template>
      <template #scope-data="{ row }">
        <span>{{ row.scope }}</span>
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
    <CreateAttributeModal 
      v-model:open="isCreateModalOpen" 
      :attribute-to-copy="attributeToCopy"
      @created="onAttributeCreated"
    />
    <EditAttributeModal 
      v-if="editingAttribute"
      v-model:open="isEditModalOpen" 
      :attribute="editingAttribute"
      @updated="onAttributeUpdated"
    />
  </UCard>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import type { DropdownItem } from '@nuxt/ui/dist/runtime/types'
import CreateAttributeModal from '~/components/Modals/CreateAttributeModal.vue'
import EditAttributeModal from '~/components/Modals/EditAttributeModal.vue'

const pending = ref(false)
const isCreateModalOpen = ref(false)
const isEditModalOpen = ref(false)
const editingAttribute = ref<any>(null)
const attributeToCopy = ref<any>(null)

// --- Options for Selects and Filters ---
const statusOptions = ['enabled', 'disabled', 'deprecated']
const scopeOptions = ['SPU', 'SKU', 'Both']
const dataTypeOptionsForFilter = ['text', 'number', 'boolean', 'enum', 'multi_enum']

// --- Table Columns ---
const columns = [
  { key: 'name', label: '名称/编码', id: 'name' },
  { key: 'dataType', label: '数据类型', id: 'dataType' },
  { key: 'scope', label: '作用域', id: 'scope' },
  { key: 'status', label: '状态', id: 'status' },
  { key: 'updatedAt', label: '最近更新', id: 'updatedAt' },
  { key: 'refCount', label: '引用数', id: 'refCount' },
  { key: 'actions', label: '操作', id: 'actions' }
]

// --- Mock Data ---
const mockData = ref([
  { id: 1, name: '材质', code: 'material', dataType: 'enum', scope: 'SPU', isSearchable: true, isAggregatable: true, isSortable: false, status: 'enabled', updatedAt: '2025-09-27', refCount: 12, validation: { enumOptions: [{value: '棉', alias:'cotton', sort:1},{value: '涤纶', alias:'polyester', sort:2}] } },
  { id: 2, name: '净含量', code: 'net_weight', dataType: 'number', scope: 'SKU', isSearchable: true, isAggregatable: false, isSortable: true, status: 'enabled', updatedAt: '2025-09-26', refCount: 5, validation: { min: 0, max: 1000 } },
  { id: 3, name: '上市年份', code: 'release_year', dataType: 'date', scope: 'SPU', isSearchable: true, isAggregatable: true, isSortable: true, status: 'disabled', updatedAt: '2025-09-25', refCount: 20 },
  { id: 4, name: '商品毛重', code: 'gross_weight', dataType: 'number', scope: 'Both', isSearchable: false, isAggregatable: false, isSortable: false, status: 'enabled', updatedAt: '2025-09-28', refCount: 8 },
])

// --- Filtering Logic ---
const filters = reactive({
  keyword: '',
  status: [],
  dataType: [],
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
  if (filters.dataType.length) {
    data = data.filter(item => filters.dataType.includes(item.dataType))
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
    { label: '删除', icon: 'i-heroicons-trash-20-solid', click: () => deleteAttribute(row), class: 'text-red-500' }
  ]
]

const openCreateModal = () => {
  attributeToCopy.value = null;
  isCreateModalOpen.value = true;
}

const openEditModal = (attr: any) => {
  editingAttribute.value = attr;
  isEditModalOpen.value = true;
}

const openCopyModal = (attr: any) => {
  attributeToCopy.value = attr;
  isCreateModalOpen.value = true;
}

const deleteAttribute = (attr: any) => {
  if (confirm(`确定要删除属性 "${attr.name}" 吗？`)) {
    const index = mockData.value.findIndex(item => item.id === attr.id);
    if (index !== -1) {
      mockData.value.splice(index, 1);
    }
  }
}

const toggleStatus = (attr: any) => {
  const index = mockData.value.findIndex(item => item.id === attr.id);
  if (index !== -1) {
    mockData.value[index].status = mockData.value[index].status === 'enabled' ? 'disabled' : 'enabled';
  }
}

const onAttributeCreated = (newAttribute: any) => {
  mockData.value.unshift(newAttribute);
  isCreateModalOpen.value = false;
}

const onAttributeUpdated = (updatedAttribute: any) => {
  const index = mockData.value.findIndex(item => item.id === updatedAttribute.id);
  if (index !== -1) {
    mockData.value[index] = { ...mockData.value[index], ...updatedAttribute, updatedAt: new Date().toISOString().split('T')[0] };
  }
  isEditModalOpen.value = false;
}

</script>
