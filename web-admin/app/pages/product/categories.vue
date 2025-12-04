<template>
  <div class="max-w-6xl mx-auto">
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          {{ $t('product.category.title') }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理商品类目树、属性模板绑定和运营设置
        </p>
      </div>
      <div class="flex space-x-2">
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="openCreateModal"
        >
          {{ $t('product.category.add') }}
        </UButton>
      </div>
    </div>

    <!-- 类目树 -->
    <UCard class="mb-6">
      <template #header>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ $t('product.category.categoryTree') }}
        </h2>
      </template>

      <div class="space-y-2">
        <draggable
          v-model="categoryTree"
          tag="ul"
          :group="{ name: 'categories', pull: true, put: true }"
          :animation="200"
          handle=".drag-handle"
          item-key="id"
          class="list-group"
          @end="onDragEnd"
        >
          <template #item="{ element, index }">
            <li class="list-group-item">
              <CategoryTreeNode
                :category="element"
                :level="0"
                @edit="editCategory"
                @delete="deleteCategory"
                @add-subcategory="addSubcategory"
              />
            </li>
          </template>
        </draggable>
      </div>

      <template #footer>
        <div class="flex justify-end space-x-2">
          <UButton variant="ghost" @click="resetTree">{{ $t('common.reset') }}</UButton>
          <UButton color="primary" @click="saveTree">{{ $t('common.save') }}</UButton>
        </div>
      </template>
    </UCard>

    <!-- 属性模板绑定 -->
    <UCard class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ $t('product.category.attributeTemplateBinding') }}
          </h2>
          <UButton
            color="primary"
            variant="outline"
            size="sm"
            icon="i-heroicons-plus"
            @click="openAttributeTemplateModal"
          >
            {{ $t('product.category.bindAttributeTemplate') }}
          </UButton>
        </div>
      </template>

      <div class="overflow-x-auto">
        <UTable
          :data="boundAttributeTemplates"
          :columns="attributeTemplateColumns"
          class="w-full"
        >
          <template #required-cell="{ getValue }">
            <UBadge :color="getValue() ? 'success' : 'neutral'">
              {{ getValue() ? $t('common.yes') : $t('common.no') }}
            </UBadge>
          </template>

          <template #actions-cell="{ row }">
            <div class="flex gap-2">
              <UButton
                color="neutral"
                variant="ghost"
                size="sm"
                icon="i-heroicons-pencil"
                @click="editAttributeTemplateBinding(row.original)"
              >
                {{ $t("common.edit") }}
              </UButton>
              <UButton
                color="error"
                variant="ghost"
                size="sm"
                icon="i-heroicons-trash"
                @click="unbindAttributeTemplate(row.original)"
              >
                {{ $t("common.delete") }}
              </UButton>
            </div>
          </template>
        </UTable>
      </div>
    </UCard>

    <!-- 运营设置 -->
    <UCard>
      <template #header>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ $t('product.category.operationSettings') }}
        </h2>
      </template>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <h3 class="text-md font-medium text-gray-900 dark:text-white mb-3">
            {{ $t('product.category.recommendedCategories') }}
          </h3>
          <div class="space-y-2">
            <div
              v-for="(category, index) in recommendedCategories"
              :key="index"
              class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
            >
              <div class="flex items-center">
                <span class="font-medium">{{ category.name }}</span>
                <UBadge
                  v-if="category.isRecommended"
                  color="success"
                  class="ml-2"
                >
                  {{ $t('product.category.recommended') }}
                </UBadge>
              </div>
              <div class="flex space-x-2">
                <UToggle
                  v-model="category.isRecommended"
                  @change="updateRecommendedCategory(category)"
                />
                <UButton
                  color="error"
                  variant="ghost"
                  size="sm"
                  icon="i-heroicons-x-mark"
                  @click="removeRecommendedCategory(index)"
                />
              </div>
            </div>
            <UButton
              color="primary"
              variant="outline"
              icon="i-heroicons-plus"
              @click="addRecommendedCategory"
            >
              {{ $t('product.category.addRecommendedCategory') }}
            </UButton>
          </div>
        </div>

        <div>
          <h3 class="text-md font-medium text-gray-900 dark:text-white mb-3">
            {{ $t('product.category.displayControl') }}
          </h3>
          <div class="space-y-4">
            <UFormField :label="$t('product.category.displayMode')">
              <USelect
                v-model="displaySettings.displayMode"
                :options="displayModeOptions"
              />
            </UFormField>

            <UFormField :label="$t('product.category.itemsPerPage')">
              <UInput
                v-model.number="displaySettings.itemsPerPage"
                type="number"
                min="1"
                max="100"
              />
            </UFormField>

            <UFormField :label="$t('product.category.showSubcategoryCount')">
              <UToggle v-model="displaySettings.showSubcategoryCount" />
            </UFormField>

            <UFormField :label="$t('product.category.enableFilter')">
              <UToggle v-model="displaySettings.enableFilter" />
            </UFormField>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <UButton color="primary" @click="saveOperationSettings">{{ $t('common.save') }}</UButton>
        </div>
      </template>
    </UCard>

    <!-- 创建/编辑类目模态框 -->
    <UModal
      v-model:open="showCategoryModal"
      :title="editingCategory ? $t('product.category.edit') : $t('product.category.add')"
      :description="editingCategory ? $t('product.category.editCategoryDescription') : $t('product.category.addCategoryDescription')"
      :close="{ onClick: () => closeCategoryModal() }"
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
              <UFormField :label="$t('form.name')" required>
                <UInput v-model="currentCategory.name" :placeholder="$t('form.name')" />
              </UFormField>

              <UFormField :label="$t('product.category.parentCategory')">
                <USelect
                  v-model="currentCategory.parentId"
                  :options="parentCategoryOptions"
                  :placeholder="$t('product.category.selectParentCategory')"
                />
              </UFormField>

              <UFormField :label="$t('product.category.sortOrder')">
                <UInput
                  v-model.number="currentCategory.sortOrder"
                  type="number"
                  min="0"
                  :placeholder="$t('product.category.sortOrder')"
                />
              </UFormField>

              <UFormField :label="$t('product.category.icon')">
                <UInput v-model="currentCategory.icon" :placeholder="$t('product.category.icon')" />
              </UFormField>

              <div class="sm:col-span-2">
                <UFormField :label="$t('form.description')">
                  <UTextarea
                    v-model="currentCategory.description"
                    :placeholder="$t('form.description')"
                    :rows="3"
                  />
                </UFormField>
              </div>

              <div class="sm:col-span-2">
                <UFormField :label="$t('product.category.seoTitle')">
                  <UInput v-model="currentCategory.seoTitle" :placeholder="$t('product.category.seoTitle')" />
                </UFormField>
              </div>

              <div class="sm:col-span-2">
                <UFormField :label="$t('product.category.seoDescription')">
                  <UTextarea
                    v-model="currentCategory.seoDescription"
                    :placeholder="$t('product.category.seoDescription')"
                    :rows="3"
                  />
                </UFormField>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeCategoryModal">{{ $t('common.cancel') }}</UButton>
          <UButton color="primary" @click="saveCategory">{{ $t('common.save') }}</UButton>
        </div>
      </template>
    </UModal>

    <!-- 属性模板绑定模态框 -->
    <UModal
      v-model:open="showAttributeTemplateModal"
      :title="editingAttributeTemplate ? $t('product.category.editAttributeTemplateBinding') : $t('product.category.bindAttributeTemplate')"
      :description="editingAttributeTemplate ? $t('product.category.editAttributeTemplateDescription') : $t('product.category.addAttributeTemplateDescription')"
      :close="{ onClick: () => closeAttributeTemplateModal() }"
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
              <UFormField :label="$t('product.category.attributeTemplate')" required>
                <USelect
                  v-model="currentAttributeTemplate.templateId"
                  :options="attributeTemplateOptions"
                  :placeholder="$t('product.category.selectAttributeTemplate')"
                />
              </UFormField>

              <UFormField :label="$t('product.category.required')">
                <UToggle v-model="currentAttributeTemplate.required" />
              </UFormField>

              <div class="sm:col-span-2">
                <UFormField :label="$t('product.category.description')">
                  <UTextarea
                    v-model="currentAttributeTemplate.description"
                    :placeholder="$t('form.description')"
                    :rows="3"
                  />
                </UFormField>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeAttributeTemplateModal">{{ $t('common.cancel') }}</UButton>
          <UButton color="primary" @click="saveAttributeTemplateBinding">{{ $t('common.save') }}</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import draggable from 'vuedraggable'

// 类型定义
type Category = {
  id: string
  name: string
  parentId: string | null
  sortOrder: number
  icon?: string
  description?: string
  seoTitle?: string
  seoDescription?: string
  children?: Category[]
}

type AttributeTemplateBinding = {
  id: string
  categoryId: string
  templateId: string
  templateName: string
  required: boolean
  description?: string
}

type RecommendedCategory = {
  id: string
  name: string
  isRecommended: boolean
}

// 状态管理
const showCategoryModal = ref(false)
const showAttributeTemplateModal = ref(false)
const editingCategory = ref(false)
const editingAttributeTemplate = ref(false)
const currentCategory = ref({
  id: '',
  name: '',
  parentId: null as string | null,
  sortOrder: 0,
  icon: '',
  description: '',
  seoTitle: '',
  seoDescription: ''
})

const currentAttributeTemplate = ref({
  id: '',
  categoryId: '',
  templateId: '',
  templateName: '',
  required: false,
  description: ''
})

const categoryTree = ref<Category[]>([
  {
    id: 'cat1',
    name: '电子产品',
    parentId: null,
    sortOrder: 1,
    icon: 'i-heroicons-computer-desktop',
    description: '各种电子设备和配件',
    children: [
      {
        id: 'cat1-1',
        name: '手机',
        parentId: 'cat1',
        sortOrder: 1,
        icon: 'i-heroicons-device-phone-mobile',
        description: '智能手机和平板电脑'
      },
      {
        id: 'cat1-2',
        name: '电脑',
        parentId: 'cat1',
        sortOrder: 2,
        icon: 'i-heroicons-computer-desktop',
        description: '台式机和笔记本电脑'
      }
    ]
  },
  {
    id: 'cat2',
    name: '服装',
    parentId: null,
    sortOrder: 2,
    icon: 'i-heroicons-shirt',
    description: '各种服装和配饰',
    children: [
      {
        id: 'cat2-1',
        name: '男装',
        parentId: 'cat2',
        sortOrder: 1,
        icon: 'i-heroicons-user',
        description: '男士服装'
      },
      {
        id: 'cat2-2',
        name: '女装',
        parentId: 'cat2',
        sortOrder: 2,
        icon: 'i-heroicons-user-circle',
        description: '女士服装'
      }
    ]
  }
])

const boundAttributeTemplates = ref<AttributeTemplateBinding[]>([
  {
    id: 'binding1',
    categoryId: 'cat1',
    templateId: 'tpl1',
    templateName: '电子产品属性模板',
    required: true,
    description: '所有电子产品都需要绑定此模板'
  },
  {
    id: 'binding2',
    categoryId: 'cat2',
    templateId: 'tpl2',
    templateName: '服装属性模板',
    required: true,
    description: '所有服装都需要绑定此模板'
  }
])

const recommendedCategories = ref<RecommendedCategory[]>([
  { id: 'cat1', name: '电子产品', isRecommended: true },
  { id: 'cat2', name: '服装', isRecommended: true }
])

const displaySettings = ref({
  displayMode: 'grid',
  itemsPerPage: 20,
  showSubcategoryCount: true,
  enableFilter: true
})

// 选项数据
const parentCategoryOptions = [
  { label: $t('common.none'), value: null },
  { label: '电子产品', value: 'cat1' },
  { label: '服装', value: 'cat2' }
]

const attributeTemplateOptions = [
  { label: '电子产品属性模板', value: 'tpl1' },
  { label: '服装属性模板', value: 'tpl2' },
  { label: '家居用品属性模板', value: 'tpl3' }
]

const displayModeOptions = [
  { label: $t('product.category.grid'), value: 'grid' },
  { label: $t('product.category.list'), value: 'list' },
  { label: $t('product.category.carousel'), value: 'carousel' }
]

// 列定义
const attributeTemplateColumns = [
  { accessorKey: 'templateName', header: $t('product.category.attributeTemplate') },
  { accessorKey: 'required', header: $t('product.category.required') },
  { accessorKey: 'description', header: $t('form.description') },
  { id: 'actions', header: $t('common.actions') }
]

// 方法
const openCreateModal = () => {
  editingCategory.value = false
  currentCategory.value = {
    id: '',
    name: '',
    parentId: null,
    sortOrder: 0,
    icon: '',
    description: '',
    seoTitle: '',
    seoDescription: ''
  }
  showCategoryModal.value = true
}

const closeCategoryModal = () => {
  showCategoryModal.value = false
  currentCategory.value = {
    id: '',
    name: '',
    parentId: null,
    sortOrder: 0,
    icon: '',
    description: '',
    seoTitle: '',
    seoDescription: ''
  }
}

const editCategory = (category: Category) => {
  editingCategory.value = true
  currentCategory.value = { ...category }
  showCategoryModal.value = true
}

const deleteCategory = (category: Category) => {
  if (confirm(`${$t('message.confirm.delete')} ${category.name}?`)) {
    // 删除类目逻辑
    alert($t('message.success.deleted'))
  }
}

const addSubcategory = (parentId: string) => {
  editingCategory.value = false
  currentCategory.value = {
    id: '',
    name: '',
    parentId: parentId,
    sortOrder: 0,
    icon: '',
    description: '',
    seoTitle: '',
    seoDescription: ''
  }
  showCategoryModal.value = true
}

const saveCategory = () => {
  // 保存类目逻辑
  alert($t('message.success.saved'))
  closeCategoryModal()
}

const onDragEnd = () => {
  // 拖拽结束后的处理逻辑
  alert($t('message.success.updated'))
}

const resetTree = () => {
  // 重置类目树逻辑
  alert($t('message.success.reset'))
}

const saveTree = () => {
  // 保存类目树逻辑
  alert($t('message.success.saved'))
}

const openAttributeTemplateModal = () => {
  editingAttributeTemplate.value = false
  currentAttributeTemplate.value = {
    id: '',
    categoryId: '',
    templateId: '',
    templateName: '',
    required: false,
    description: ''
  }
  showAttributeTemplateModal.value = true
}

const closeAttributeTemplateModal = () => {
  showAttributeTemplateModal.value = false
  currentAttributeTemplate.value = {
    id: '',
    categoryId: '',
    templateId: '',
    templateName: '',
    required: false,
    description: ''
  }
}

const editAttributeTemplateBinding = (binding: AttributeTemplateBinding) => {
  editingAttributeTemplate.value = true
  currentAttributeTemplate.value = { ...binding }
  showAttributeTemplateModal.value = true
}

const unbindAttributeTemplate = (binding: AttributeTemplateBinding) => {
  if (confirm(`${$t('message.confirm.delete')} ${binding.templateName}?`)) {
    // 解绑属性模板逻辑
    boundAttributeTemplates.value = boundAttributeTemplates.value.filter(b => b.id !== binding.id)
    alert($t('message.success.deleted'))
  }
}

const saveAttributeTemplateBinding = () => {
  // 保存属性模板绑定逻辑
  alert($t('message.success.saved'))
  closeAttributeTemplateModal()
}

const addRecommendedCategory = () => {
  // 添加推荐类目逻辑
  alert($t('product.category.addRecommendedCategory'))
}

const removeRecommendedCategory = (index: number) => {
  recommendedCategories.value.splice(index, 1)
  alert($t('message.success.deleted'))
}

const updateRecommendedCategory = (category: RecommendedCategory) => {
  // 更新推荐类目逻辑
  alert(`${category.name} ${category.isRecommended ? $t('product.category.recommended') : $t('product.category.notRecommended')}`)
}

const saveOperationSettings = () => {
  // 保存运营设置逻辑
  alert($t('message.success.saved'))
}
</script>
