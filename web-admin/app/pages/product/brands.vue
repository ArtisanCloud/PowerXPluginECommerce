<template>
  <div class="p-6">
    <UAlert
      class="mb-6"
      color="neutral"
      variant="soft"
      title="提示"
      description="品牌功能后端接口尚未接入，本页暂不发起网络请求（避免 404 噪音）。"
    />
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">品牌管理</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理您的品牌信息、品牌馆和频道配置</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-plus" :disabled="!BRAND_API_ENABLED" @click="openBrandModal()"
          >新增品牌</UButton
        >
      </div>
    </div>

    <!-- 选项卡 -->
    <UTabs v-model="activeTab" :items="tabs" class="mb-6" />

    <!-- 品牌列表/维护 -->
    <UCard v-if="activeTab === 'list'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">品牌列表/维护</h2>
          <div class="flex gap-3">
            <UInput v-model="searchQuery" placeholder="搜索品牌..." icon="i-heroicons-magnifying-glass" size="sm" />
            <USelectMenu
              v-model="filterOrigin"
              :items="originOptions"
              value-key="value"
              label-key="label"
              :portal="false"
              placeholder="按产地筛选"
              size="sm"
            />
            <UButton color="neutral" @click="resetFilters">重置</UButton>
          </div>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">品牌名称</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">LOGO</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">产地</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">介绍</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="brand in brands" :key="brand.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{{ brand.name }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <UAvatar 
                  :src="brand.logo" 
                  :alt="brand.name"
                  size="sm"
                  class="border border-gray-200 dark:border-gray-700"
                />
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{{ brand.origin }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400 max-w-xs truncate">{{ brand.description }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-eye"
                    @click="viewBrand(brand)"
                  >
                    查看
                  </UButton>
                  <UButton
                    color="primary"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-pencil"
                    @click="openBrandModal(brand)"
                  >
                    编辑
                  </UButton>
                  <UButton
                    color="red"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="deleteBrand(brand.id)"
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
            显示第 {{ (currentPage - 1) * 10 + 1 }}
            到 {{ Math.min(currentPage * 10, brands.length) }} 条，共 {{ brands.length }} 条
          </div>
          <UPagination
            v-model="currentPage"
            :page-count="pageCount"
            :total="brands.length"
            :ui="{ rounded: 'rounded-full' }"
          />
        </div>
      </template>
    </UCard>

    <!-- 专区配置 -->
    <UCard v-if="activeTab === 'configuration'" class="mb-6">
      <template #header>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">品牌专区配置</h2>
      </template>

      <div class="space-y-6">
        <!-- 品牌馆配置 -->
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg">
          <div class="p-4 border-b border-gray-200 dark:border-gray-700">
            <h3 class="text-md font-medium text-gray-900 dark:text-white flex items-center">
              <UIcon name="i-heroicons-building-storefront" class="mr-2 w-5 h-5" />
              品牌馆配置
            </h3>
          </div>
          <div class="p-4 space-y-4">
            <UFormField label="启用品牌馆" description="允许顾客浏览品牌馆页面" class="flex items-center">
              <USwitch v-model="brandMallEnabled" class="ml-auto" />
            </UFormField>
            
            <UFormField label="默认品牌馆页面" description="选择默认展示的品牌馆页面">
              <USelectMenu
                v-model="defaultBrandMallPage"
                :items="brandMallPages"
                value-key="value"
                label-key="label"
                :portal="false"
                placeholder="选择页面"
              />
            </UFormField>

            <UFormField label="品牌馆展示样式" description="选择品牌馆的展示样式">
              <USelectMenu
                v-model="brandMallStyle"
                :items="styleOptions"
                value-key="value"
                label-key="label"
                :portal="false"
                placeholder="选择样式"
              />
            </UFormField>

            <UFormField label="品牌馆主题色" description="为品牌馆设置主题颜色">
              <div class="flex items-center space-x-4">
                <input
                  v-model="brandMallTheme"
                  type="color"
                  class="w-12 h-10 p-1 border border-gray-300 rounded"
                />
                <span class="text-sm text-gray-500">{{ brandMallTheme }}</span>
              </div>
            </UFormField>

            <div class="pt-4 flex justify-end">
              <UButton
                color="primary"
                variant="solid"
                @click="saveBrandMallSettings"
              >
                保存品牌馆设置
              </UButton>
            </div>
          </div>
        </div>

        <!-- 频道挂载配置 -->
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg">
          <div class="p-4 border-b border-gray-200 dark:border-gray-700">
            <h3 class="text-md font-medium text-gray-900 dark:text-white flex items-center">
              <UIcon name="i-heroicons-globe-alt" class="mr-2 w-5 h-5" />
              频道挂载配置
            </h3>
          </div>
          <div class="p-4 space-y-4">
            <UFormField label="品牌频道列表" description="选择可在频道中展示的品牌">
              <USelectMenu
                v-model="selectedBrandChannels"
                :items="allBrandItems"
                value-key="value"
                label-key="label"
                :portal="false"
                multiple
                placeholder="选择品牌"
              />
            </UFormField>
            
            <UFormField label="频道页面样式" description="选择频道页面的展示样式">
              <USelectMenu
                v-model="channelPageStyle"
                :items="styleOptions"
                value-key="value"
                label-key="label"
                :portal="false"
                placeholder="选择样式"
              />
            </UFormField>
            
            <UFormField label="品牌展示顺序" description="设置频道中品牌的展示顺序">
              <USelectMenu
                v-model="brandDisplayOrder"
                :items="orderOptions"
                value-key="value"
                label-key="label"
                :portal="false"
                placeholder="选择排序方式"
              />
            </UFormField>

            <UFormField label="频道页面头部" description="自定义频道页面的头部内容">
              <UTextarea
                v-model="channelHeaderContent"
                placeholder="输入频道页面头部自定义内容"
                rows="3"
              />
            </UFormField>

            <div class="pt-4 flex justify-end">
              <UButton
                color="primary"
                variant="solid"
                @click="saveChannelSettings"
              >
                保存频道设置
              </UButton>
            </div>
          </div>
        </div>

        <!-- 品牌商店集成 -->
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg">
          <div class="p-4 border-b border-gray-200 dark:border-gray-700">
            <h3 class="text-md font-medium text-gray-900 dark:text-white flex items-center">
              <UIcon name="i-heroicons-shopping-bag" class="mr-2 w-5 h-5" />
              品牌商店集成
            </h3>
          </div>
          <div class="p-4 space-y-4">
            <UFormField label="启用品牌商店" description="为每个品牌创建专属商店页面" class="flex items-center">
              <USwitch v-model="brandStoreEnabled" class="ml-auto" />
            </UFormField>
            
            <UFormField label="商店页面模板" description="选择品牌商店页面的模板">
              <USelectMenu
                v-model="brandStoreTemplate"
                :items="storeTemplates"
                value-key="value"
                label-key="label"
                :portal="false"
                placeholder="选择模板"
              />
            </UFormField>
            
            <UFormField label="商店导航配置" description="配置品牌商店的导航结构">
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                <UCheckbox
                  v-for="nav in storeNavigationOptions"
                  :key="nav.value"
                  v-model="selectedStoreNavigation"
                  :value="nav.value"
                  :label="nav.label"
                  size="md"
                />
              </div>
            </UFormField>
            
            <UFormField label="商店页面元素" description="选择在商店页面展示的元素">
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                <UCheckbox
                  v-for="element in storeElements"
                  :key="element.value"
                  v-model="selectedStoreElements"
                  :value="element.value"
                  :label="element.label"
                  size="md"
                />
              </div>
            </UFormField>

            <div class="pt-4 flex justify-end">
              <UButton
                color="primary"
                variant="solid"
                @click="saveStoreIntegrationSettings"
              >
                保存商店集成设置
              </UButton>
            </div>
          </div>
        </div>
      </div>
    </UCard>

    <!-- 品牌编辑模态框 -->
    <UModal
      v-model:open="showBrandModal"
      :title="editingBrand ? '编辑品牌' : '新增品牌'"
      :description="editingBrand ? '编辑现有品牌信息' : '创建新品牌'"
      :close="{ onClick: () => closeBrandModal() }"
      :ui="{
        content: 'w-full sm:max-w-3xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-6 p-4 sm:p-6">
            <UForm :schema="brandFormSchema" :state="brandFormState" @submit="saveBrand">
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <UFormField label="品牌名称 *" name="name" required :error="brandFormErrors.name">
                  <UInput v-model="brandFormState.name" placeholder="输入品牌名称" />
                </UFormField>

                <UFormField label="品牌产地" name="origin">
                  <UInput v-model="brandFormState.origin" placeholder="输入品牌产地，如：中国、法国等" />
                </UFormField>
              </div>

              <UFormField label="品牌LOGO" name="logo">
                <div class="flex flex-col sm:flex-row gap-4 items-start">
                  <UInput
                    v-model="brandFormState.logo"
                    placeholder="输入LOGO图片URL"
                    class="flex-1"
                  />
                  <div class="flex-shrink-0">
                    <UAvatar 
                      :src="brandFormState.logo" 
                      :alt="brandFormState.name" 
                      size="md" 
                      class="border border-gray-200 dark:border-gray-700"
                    />
                  </div>
                </div>
              </UFormField>

              <UFormField label="品牌介绍" name="description">
                <UTextarea
                  v-model="brandFormState.description"
                  placeholder="输入品牌介绍"
                  rows="3"
                />
              </UFormField>

              <div class="flex flex-col sm:flex-row justify-end space-y-3 sm:space-y-0 sm:space-x-3 pt-2">
                <UButton 
                  type="button" 
                  color="gray" 
                  variant="outline" 
                  @click="closeBrandModal"
                  class="w-full sm:w-auto"
                >
                  取消
                </UButton>
                <UButton 
                  type="submit" 
                  color="primary" 
                  variant="solid"
                  class="w-full sm:w-auto"
                >
                  保存
                </UButton>
              </div>
            </UForm>
          </div>
        </UCard>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useToastAlert } from '~/composables/useToastAlert'

// 选项卡配置
const activeTab = ref<'list' | 'configuration'>('list')
const tabs = [
  { label: '品牌列表/维护', value: 'list' },
  { label: '专区配置', value: 'configuration' }
]

// toast
const toast = useToastAlert()

// 品牌数据
const brands = ref<any[]>([])
const loading = ref(false)
const currentPage = ref(1)
const searchQuery = ref('')
const filterOrigin = ref(ORIGIN_ALL)

// 分页
const pageCount = computed(() => Math.ceil(brands.value.length / 10))

// 品牌表单
const brandFormSchema = ['name', 'logo', 'description', 'origin']

const brandFormState = reactive({
  id: null as number | null,
  name: '',
  logo: '',
  description: '',
  origin: ''
})

const brandFormErrors = reactive({
  name: ''
})

// 模态框 & 编辑状态
const showBrandModal = ref(false)
const editingBrand = ref<any>(null)

// 配置选项
const brandMallEnabled = ref(true)
const brandStoreEnabled = ref(true)
const defaultBrandMallPage = ref('grid')
const brandMallStyle = ref('modern')
const brandMallTheme = ref('#3b82f6')
const channelPageStyle = ref('grid')
const brandDisplayOrder = ref('alphabetic')
const selectedBrandChannels = ref<number[]>([])
const allBrands = ref<{ id: number; name: string }[]>([])
const channelHeaderContent = ref('')
const brandStoreTemplate = ref('default')
const selectedStoreNavigation = ref<string[]>(['products', 'about', 'contact'])
const selectedStoreElements = ref<string[]>(['featured', 'new_arrivals', 'best_sellers', 'promotions'])
const storeNavigationOptions = [
  { label: '产品分类', value: 'categories' },
  { label: '热销商品', value: 'products' },
  { label: '关于我们', value: 'about' },
  { label: '联系我们', value: 'contact' },
  { label: '客户服务', value: 'support' }
]
const storeElements = [
  { label: '推荐商品', value: 'featured' },
  { label: '新品上架', value: 'new_arrivals' },
  { label: '热销排行', value: 'best_sellers' },
  { label: '促销活动', value: 'promotions' },
  { label: '品牌故事', value: 'brand_story' },
  { label: '客户评价', value: 'reviews' }
]
const brandMallPages = [
  { label: '网格布局', value: 'grid' },
  { label: '列表布局', value: 'list' },
  { label: '画廊布局', value: 'gallery' }
]
const storeTemplates = [
  { label: '默认模板', value: 'default' },
  { label: '现代风格', value: 'modern' },
  { label: '经典风格', value: 'classic' },
  { label: '时尚风格', value: 'fashion' },
  { label: '简约风格', value: 'minimal' }
]
const styleOptions = [
  { label: '现代风格', value: 'modern' },
  { label: '经典风格', value: 'classic' },
  { label: '简约风格', value: 'minimal' }
]
const orderOptions = [
  { label: '按字母顺序', value: 'alphabetic' },
  { label: '按销量', value: 'sales' },
  { label: '按受欢迎程度', value: 'popularity' }
]
const BRAND_API_ENABLED = false
const ORIGIN_ALL = '__all__'

const originOptions = [
  { label: '全部产地', value: ORIGIN_ALL },
  { label: '中国', value: '中国' },
  { label: '美国', value: '美国' },
  { label: '法国', value: '法国' },
  { label: '德国', value: '德国' },
  { label: '日本', value: '日本' }
]

const allBrandItems = computed(() => (allBrands.value || []).map((b: any) => ({ label: b.name, value: b.id })))

// API calls for brand management
const loadBrands = async () => {
  if (!BRAND_API_ENABLED) {
    brands.value = []
    return
  }
  loading.value = true
  try {
    const response = await $fetch('/api/brands', {
      method: 'GET',
      params: {
        page: currentPage.value,
        limit: 10,
        search: searchQuery.value,
        origin: filterOrigin.value === ORIGIN_ALL ? '' : filterOrigin.value
      }
    })
    
    // 兜底，保证数据稳定
    brands.value = Array.isArray(response?.data) ? response.data : []
  } catch (error) {
    console.error('Error loading brands:', error)
    toast.add({ title: '加载品牌数据失败', color: 'red' })
    brands.value = []
  } finally {
    loading.value = false
  }
}

// Load brand configuration settings
const loadBrandConfig = async () => {
  if (!BRAND_API_ENABLED) {
    return
  }
  try {
    const response = await $fetch('/api/brands/config')
    const config = response?.data || {}

    brandMallEnabled.value = config.brandMallEnabled ?? true
    brandStoreEnabled.value = config.brandStoreEnabled ?? true
    defaultBrandMallPage.value = config.defaultBrandMallPage || 'grid'
    brandMallStyle.value = config.brandMallStyle || 'modern'
    brandMallTheme.value = config.brandMallTheme || '#3b82f6'
    channelPageStyle.value = config.channelPageStyle || 'grid'
    brandDisplayOrder.value = config.brandDisplayOrder || 'alphabetic'
    channelHeaderContent.value = config.channelHeaderContent || ''
    brandStoreTemplate.value = config.brandStoreTemplate || 'default'
    selectedStoreNavigation.value = config.selectedStoreNavigation || ['products', 'about', 'contact']
    selectedStoreElements.value = config.selectedStoreElements || ['featured', 'new_arrivals', 'best_sellers', 'promotions']
    selectedBrandChannels.value = config.selectedBrandChannels || []
  } catch (error) {
    console.error('Error loading brand config:', error)
    toast.add({ title: '加载品牌配置失败', color: 'red' })
  }
}

// 关闭品牌模态框
const closeBrandModal = () => {
  showBrandModal.value = false
  // 重置表单
  brandFormState.id = null
  brandFormState.name = ''
  brandFormState.logo = ''
  brandFormState.description = ''
  brandFormState.origin = ''
  editingBrand.value = null
  brandFormErrors.name = ''
}

// 保存品牌
const saveBrand = async () => {
  if (!BRAND_API_ENABLED) {
    toast.add({ title: '品牌功能未接入', color: 'yellow' })
    return
  }
  if (!brandFormState.name.trim()) {
    brandFormErrors.name = '品牌名称不能为空'
    return
  }
  try {
    if (editingBrand.value) {
      await $fetch(`/api/brands/${brandFormState.id}`, {
        method: 'PUT',
        body: {
          name: brandFormState.name,
          logo: brandFormState.logo,
          description: brandFormState.description,
          origin: brandFormState.origin
        }
      })
      toast.add({ title: '品牌更新成功', color: 'green' })
    } else {
      await $fetch('/api/brands', {
        method: 'POST',
        body: {
          name: brandFormState.name,
          logo: brandFormState.logo,
          description: brandFormState.description,
          origin: brandFormState.origin
        }
      })
      toast.add({ title: '品牌添加成功', color: 'green' })
    }
    closeBrandModal()
    await loadBrands()
  } catch (error) {
    console.error('Error saving brand:', error)
    toast.add({ title: editingBrand.value ? '更新品牌失败' : '添加品牌失败', color: 'red' })
  }
}

// 查看品牌
const viewBrand = (brand: any) => {
  alert(`查看品牌: ${brand.name}`)
}

// 删除品牌
const deleteBrand = async (id: number) => {
  if (!BRAND_API_ENABLED) {
    toast.add({ title: '品牌功能未接入', color: 'yellow' })
    return
  }
  if (!confirm('确定要删除这个品牌吗？')) return
  try {
    await $fetch(`/api/brands/${id}`, { method: 'DELETE' })
    toast.add({ title: '品牌删除成功', color: 'green' })
    await loadBrands()
  } catch (error) {
    console.error('Error deleting brand:', error)
    toast.add({ title: '删除品牌失败', color: 'red' })
  }
}

// 打开品牌编辑模态框
const openBrandModal = (brand: any = null) => {
  if (brand) {
    editingBrand.value = brand
    brandFormState.id = brand.id
    brandFormState.name = brand.name
    brandFormState.logo = brand.logo
    brandFormState.description = brand.description
    brandFormState.origin = brand.origin
  } else {
    editingBrand.value = null
    brandFormState.id = null
    brandFormState.name = ''
    brandFormState.logo = ''
    brandFormState.description = ''
    brandFormState.origin = ''
  }
  brandFormErrors.name = ''
  showBrandModal.value = true
}

// 重置筛选器
const resetFilters = () => {
  searchQuery.value = ''
  filterOrigin.value = ORIGIN_ALL
  loadBrands()
}

// 保存品牌馆设置
const saveBrandMallSettings = async () => {
  if (!BRAND_API_ENABLED) {
    toast.add({ title: '品牌功能未接入', color: 'yellow' })
    return
  }
  try {
    await $fetch('/api/brands/config', {
      method: 'PUT',
      body: {
        brandMallEnabled: brandMallEnabled.value,
        defaultBrandMallPage: defaultBrandMallPage.value,
        brandMallStyle: brandMallStyle.value,
        brandMallTheme: brandMallTheme.value
      }
    })
    toast.add({ title: '品牌馆设置保存成功', color: 'green' })
  } catch (error) {
    console.error('Error saving brand mall settings:', error)
    toast.add({ title: '保存品牌馆设置失败', color: 'red' })
  }
}

// 保存频道设置
const saveChannelSettings = async () => {
  if (!BRAND_API_ENABLED) {
    toast.add({ title: '品牌功能未接入', color: 'yellow' })
    return
  }
  try {
    await $fetch('/api/brands/config', {
      method: 'PUT',
      body: {
        channelPageStyle: channelPageStyle.value,
        brandDisplayOrder: brandDisplayOrder.value,
        channelHeaderContent: channelHeaderContent.value,
        selectedBrandChannels: selectedBrandChannels.value
      }
    })
    toast.add({ title: '频道设置保存成功', color: 'green' })
  } catch (error) {
    console.error('Error saving channel settings:', error)
    toast.add({ title: '保存频道设置失败', color: 'red' })
  }
}

// 保存商店集成设置
const saveStoreIntegrationSettings = async () => {
  if (!BRAND_API_ENABLED) {
    toast.add({ title: '品牌功能未接入', color: 'yellow' })
    return
  }
  try {
    await $fetch('/api/brands/config', {
      method: 'PUT',
      body: {
        brandStoreEnabled: brandStoreEnabled.value,
        brandStoreTemplate: brandStoreTemplate.value,
        selectedStoreNavigation: selectedStoreNavigation.value,
        selectedStoreElements: selectedStoreElements.value
      }
    })
    toast.add({ title: '商店集成设置保存成功', color: 'green' })
  } catch (error) {
    console.error('Error saving store integration settings:', error)
    toast.add({ title: '保存商店集成设置失败', color: 'red' })
  }
}

// 初始化数据
onMounted(async () => {
  filterOrigin.value = ORIGIN_ALL
  if (!BRAND_API_ENABLED) {
    brands.value = []
    allBrands.value = []
    return
  }
  await loadBrands()
  await loadBrandConfig()
  try {
    const response = await $fetch('/api/brands/all')
    allBrands.value = Array.isArray(response?.data) ? response.data : []
  } catch (error) {
    console.error('Error loading all brands:', error)
    allBrands.value = []
  }
})
</script>
