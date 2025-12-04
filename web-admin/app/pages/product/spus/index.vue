<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          {{ $t("product.title") }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理SPU（标准化产品单元）信息
        </p>
      </div>
      <UButton
        color="primary"
        icon="i-heroicons-plus"
        @click="goToCreate"
      >
        {{ $t('product.add') }}
      </UButton>
    </div>

    <!-- 筛选和搜索栏 -->
    <UCard class="mb-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UInput
          v-model="searchQuery"
          :placeholder="$t('product.search')"
          icon="i-heroicons-magnifying-glass"
        />
        <USelect
          v-model="statusFilter"
          :options="statusOptions"
          :placeholder="$t('product.status')"
        />
        <USelect
          v-model="categoryFilter"
          :options="categoryOptions"
          :placeholder="$t('product.category')"
        />
        <USelect
          v-model="brandFilter"
          :options="brandOptions"
          :placeholder="$t('product.brand')"
        />
      </div>
      <div class="flex justify-end mt-4 space-x-2">
        <UButton @click="resetFilters" variant="ghost">{{ $t('common.reset') }}</UButton>
        <UButton @click="searchProducts">{{ $t('common.search') }}</UButton>
      </div>
    </UCard>

    <!-- 批量操作栏 -->
    <div class="flex justify-between items-center mb-4">
      <div class="flex items-center space-x-2">
        <UCheckbox v-model="selectAll" @change="toggleSelectAll" />
        <span class="text-sm text-gray-600 dark:text-gray-400">
          {{ selectedProducts.length }} {{ $t('common.selected') }}
        </span>
        <UButton
          v-if="selectedProducts.length > 0"
          color="primary"
          variant="outline"
          size="sm"
          @click="batchPublish"
        >
          {{ $t('product.publish') }}
        </UButton>
        <UButton
          v-if="selectedProducts.length > 0"
          color="primary"
          variant="outline"
          size="sm"
          @click="batchUnpublish"
        >
          {{ $t('product.unpublish') }}
        </UButton>
        <UButton
          v-if="selectedProducts.length > 0"
          color="success"
          variant="outline"
          size="sm"
          @click="batchEnable"
        >
          {{ $t('status.enable') }}
        </UButton>
        <UButton
          v-if="selectedProducts.length > 0"
          color="error"
          variant="outline"
          size="sm"
          @click="batchDisable"
        >
          {{ $t('status.disable') }}
        </UButton>
      </div>
      <div class="flex space-x-2">
        <UButton variant="outline" icon="i-heroicons-arrow-down-tray">{{ $t('common.export') }}</UButton>
        <UButton variant="outline" icon="i-heroicons-arrow-up-tray">{{ $t('common.import') }}</UButton>
      </div>
    </div>

    <!-- 商品表格 -->
    <UCard>
      <UTable
        v-model="selectedProducts"
        :data="filteredProducts"
        :columns="columns"
        :loading="loading"
        selectable
        class="w-full"
      >
        <!-- 状态列 -->
        <template #status-cell="{ getValue }">
          <UBadge
            :color="getValue() === 'active' ? 'success' : 'error'"
            variant="subtle"
          >
            {{ getValue() === "active" ? $t("status.active") : $t("status.inactive") }}
          </UBadge>
        </template>

        <!-- 发布状态列 -->
        <template #published-cell="{ getValue }">
          <UBadge
            :color="getValue() ? 'success' : 'warning'"
            variant="subtle"
          >
            {{ getValue() ? $t("status.published") : $t("status.draft") }}
          </UBadge>
        </template>

        <!-- 操作列 -->
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewProduct(row.original)"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
              @click="editProduct(row.original)"
            >
              {{ $t("common.edit") }}
            </UButton>
            <UButton
              v-if="row.original.published"
              color="warning"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye-slash"
              @click="unpublishProduct(row.original)"
            >
              {{ $t("product.unpublish") }}
            </UButton>
            <UButton
              v-else
              color="success"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="publishProduct(row.original)"
            >
              {{ $t("product.publish") }}
            </UButton>
            <UButton
              v-if="row.original.status === 'active'"
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-x-circle"
              @click="disableProduct(row.original)"
            >
              {{ $t("status.disable") }}
            </UButton>
            <UButton
              v-else
              color="success"
              variant="ghost"
              size="sm"
              icon="i-heroicons-check-circle"
              @click="enableProduct(row.original)"
            >
              {{ $t("status.enable") }}
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 分页 -->
    <div class="flex justify-between items-center mt-4">
      <div class="text-sm text-gray-600 dark:text-gray-400">
        {{ $t('common.total') }} {{ products.length }} {{ $t('product.products') }}
      </div>
      <UPagination
        v-model="page"
        :page-count="pageCount"
        :total="filteredProducts.length"
      />
    </div>

    <!-- 商品详情模态框 -->
    <UModal
      v-model:open="showDetailModal"
      :title="$t('product.detail')"
      :close="{ onClick: () => closeDetailModal() }"
      :ui="{
        content: 'w-full sm:max-w-4xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <div v-if="currentProduct" class="space-y-6 p-6">
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <!-- 左侧：商品图片 -->
            <div class="lg:col-span-1">
              <div class="space-y-4">
                <div class="aspect-square bg-gray-100 dark:bg-gray-800 rounded-lg overflow-hidden">
                  <img
                    v-if="currentProduct.images && currentProduct.images.length > 0"
                    :src="currentProduct.images[0]"
                    :alt="currentProduct.name"
                    class="w-full h-full object-cover"
                  />
                  <div
                    v-else
                    class="w-full h-full flex items-center justify-center text-gray-400"
                  >
                    <UIcon name="i-heroicons-photo" class="w-12 h-12" />
                  </div>
                </div>

                <div class="grid grid-cols-4 gap-2">
                  <div
                    v-for="(image, index) in currentProduct.images"
                    :key="index"
                    class="aspect-square bg-gray-100 dark:bg-gray-800 rounded cursor-pointer border-2"
                    :class="{ 'border-primary-500': index === currentImageIndex }"
                    @click="currentImageIndex = index"
                  >
                    <img
                      :src="image"
                      :alt="`Product image ${index + 1}`"
                      class="w-full h-full object-cover"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- 右侧：商品详情 -->
            <div class="lg:col-span-2 space-y-6">
              <!-- 商品名称和基本信息 -->
              <div class="flex justify-between items-start">
                <div>
                  <h2 class="text-2xl font-bold text-gray-900 dark:text-white">{{ currentProduct.name }}</h2>
                  <div class="mt-2">
                    <span class="text-3xl font-bold text-primary-600 dark:text-primary-400">
                      {{ formatCurrency(currentProduct.price) }}
                    </span>
                    <span v-if="currentProduct.originalPrice && currentProduct.originalPrice > currentProduct.price" class="ml-2 text-lg text-gray-500 line-through">
                      {{ formatCurrency(currentProduct.originalPrice) }}
                    </span>
                  </div>
                </div>
                <div class="text-right">
                  <div class="text-sm text-gray-600 dark:text-gray-400">{{ $t('product.stock') }}</div>
                  <div class="text-xl font-semibold" :class="currentProduct.stock > 0 ? 'text-green-600' : 'text-red-600'">
                    {{ currentProduct.stock > 0 ? currentProduct.stock : $t('product.outOfStock') }}
                  </div>
                </div>
              </div>

              <!-- 商品基本信息 -->
              <div class="grid grid-cols-2 gap-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <div class="flex justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ $t('product.productNumber') }}:</span>
                  <span class="font-medium">{{ currentProduct.productNumber }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ $t('product.category') }}:</span>
                  <span class="font-medium">{{ currentProduct.category }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ $t('product.brand') }}:</span>
                  <span class="font-medium">{{ currentProduct.brand }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ $t('product.barcode') }}:</span>
                  <span class="font-medium">{{ currentProduct.barcode }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ $t('product.status') }}:</span>
                  <UBadge :color="currentProduct.status === 'active' ? 'success' : 'error'">
                    {{ currentProduct.status === 'active' ? $t("status.active") : $t("status.inactive") }}
                  </UBadge>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ $t('product.published') }}:</span>
                  <UBadge :color="currentProduct.published ? 'success' : 'warning'">
                    {{ currentProduct.published ? $t("status.published") : $t("status.draft") }}
                  </UBadge>
                </div>
              </div>
            </div>
          </div>

          <!-- 商品描述 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              {{ $t('form.description') }}
            </h3>
            <div class="prose dark:prose-invert max-w-none">
              <p>{{ currentProduct.description || $t('common.noData') }}</p>
            </div>
          </div>

          <!-- 商品参数 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              {{ $t('product.specification') }}
            </h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div
                v-for="(attr, index) in currentProduct.attributes"
                :key="index"
                class="flex justify-between py-2 border-b border-gray-100 dark:border-gray-800"
              >
                <span class="text-gray-600 dark:text-gray-400">{{ attr.name }}:</span>
                <span class="font-medium">{{ attr.value }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end space-x-2">
          <UButton variant="ghost" @click="closeDetailModal">{{ $t('common.close') }}</UButton>
          <UButton color="primary" @click="editProduct(currentProduct)">{{ $t('common.edit') }}</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { h } from "vue";
import type { TableColumn } from "@nuxt/ui";

const { t } = useI18n();
const router = useRouter();

// 状态管理
const loading = ref(false);
const page = ref(1);
const pageSize = ref(10);
const searchQuery = ref("");
const statusFilter = ref("");
const categoryFilter = ref("");
const brandFilter = ref("");
const selectAll = ref(false);
const selectedProducts = ref([]);
const showDetailModal = ref(false);
const currentProduct = ref(null);
const currentImageIndex = ref(0);

// 筛选选项
const statusOptions = [
  { label: t("common.all"), value: "" },
  { label: t("status.active"), value: "active" },
  { label: t("status.inactive"), value: "inactive" }
];

const categoryOptions = [
  { label: t("common.all"), value: "" },
  { label: "手机", value: "手机" },
  { label: "电脑", value: "电脑" },
  { label: "配件", value: "配件" }
];

const brandOptions = [
  { label: t("common.all"), value: "" },
  { label: "Apple", value: "Apple" },
  { label: "Samsung", value: "Samsung" },
  { label: "Huawei", value: "Huawei" }
];

// 商品数据类型
type Product = {
  id: string;
  productNumber: string;
  name: string;
  category: string;
  brand: string;
  barcode: string;
  price: number;
  originalPrice?: number;
  stock: number;
  status: "active" | "inactive";
  published: boolean;
  description?: string;
  images?: string[];
  attributes?: { name: string; value: string }[];
};

// 商品数据
const products = ref<Product[]>([
  {
    id: "P001",
    productNumber: "SPU2023001",
    name: "iPhone 15 Pro",
    category: "手机",
    brand: "Apple",
    barcode: "1234567890123",
    price: 8999,
    originalPrice: 9999,
    stock: 50,
    status: "active",
    published: true,
    description: "最新款苹果手机，配备A17 Pro芯片，拥有钛金属设计和全新的操作按钮。支持5G网络，搭载超 Retina XDR 显示屏，带来惊艳的视觉体验。",
    images: ["https://via.placeholder.com/300x300", "https://via.placeholder.com/300x300/ff0000/ffffff"],
    attributes: [
      { name: "型号", value: "A2822" },
      { name: "颜色", value: "黑色, 白色, 蓝色" },
      { name: "存储容量", value: "128GB, 256GB, 512GB, 1TB" },
      { name: "屏幕尺寸", value: "6.1英寸" },
      { name: "处理器", value: "A17 Pro" }
    ]
  },
  {
    id: "P002",
    productNumber: "SPU2023002",
    name: "MacBook Pro",
    category: "电脑",
    brand: "Apple",
    barcode: "1234567890124",
    price: 12999,
    stock: 25,
    status: "active",
    published: true,
    description: "专业级笔记本电脑，搭载M2 Pro芯片，拥有强大的性能和出色的显示屏。",
    images: ["https://via.placeholder.com/300x300/00ff00/ffffff"],
    attributes: [
      { name: "型号", value: "MacBookPro18,3" },
      { name: "颜色", value: "银色, 深空灰" },
      { name: "存储容量", value: "512GB, 1TB, 2TB, 4TB" },
      { name: "屏幕尺寸", value: "14英寸, 16英寸" },
      { name: "处理器", value: "M2 Pro, M2 Max" }
    ]
  },
  {
    id: "P003",
    productNumber: "SPU2023003",
    name: "AirPods Pro",
    category: "配件",
    brand: "Apple",
    barcode: "1234567890125",
    price: 1899,
    stock: 0,
    status: "inactive",
    published: false,
    description: "主动降噪无线耳机，提供卓越的音质和舒适的佩戴体验。",
    images: ["https://via.placeholder.com/300x300/0000ff/ffffff"],
    attributes: [
      { name: "型号", value: "AirPodsPro2" },
      { name: "颜色", value: "白色" },
      { name: "电池续航", value: "6小时(耳机), 24小时(充电盒)" },
      { name: "充电方式", value: "Lightning, 无线充电" }
    ]
  }
]);

// 计算属性
const filteredProducts = computed(() => {
  return products.value.filter(product => {
    const matchesSearch = !searchQuery.value ||
      product.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      product.productNumber.toLowerCase().includes(searchQuery.value.toLowerCase());

    const matchesStatus = !statusFilter.value || product.status === statusFilter.value;
    const matchesCategory = !categoryFilter.value || product.category === categoryFilter.value;
    const matchesBrand = !brandFilter.value || product.brand === brandFilter.value;

    return matchesSearch && matchesStatus && matchesCategory && matchesBrand;
  });
});

const pageCount = computed(() =>
  Math.ceil(filteredProducts.value.length / pageSize.value)
);

// 列定义
const columns = computed<TableColumn<Product>[]>(() => [
  { accessorKey: "productNumber", header: t("product.productNumber") },
  { accessorKey: "name", header: t("product.productName") },
  { accessorKey: "category", header: t("product.category") },
  { accessorKey: "brand", header: t("product.brand") },
  {
    accessorKey: "price",
    header: t("product.price"),
    cell: ({ row, getValue }) => {
      const val = Number(getValue() ?? 0);
      const formatted = new Intl.NumberFormat("zh-CN", {
        style: "currency",
        currency: "CNY",
      }).format(val);
      return h("div", { class: "text-right font-medium" }, formatted);
    },
    meta: { class: { td: "text-right" } },
  },
  { accessorKey: "stock", header: t("product.stock") },
  { accessorKey: "status", header: t("product.status") },
  { accessorKey: "published", header: t("product.published") },
  { id: "actions", header: t("product.actions") },
]);

// 方法
const searchProducts = () => {
  // 搜索逻辑已在computed中实现
};

const resetFilters = () => {
  searchQuery.value = "";
  statusFilter.value = "";
  categoryFilter.value = "";
  brandFilter.value = "";
};

const toggleSelectAll = () => {
  if (selectAll.value) {
    selectedProducts.value = [...filteredProducts.value];
  } else {
    selectedProducts.value = [];
  }
};

const goToCreate = () => {
  router.push("/product/spus/create");
};

const viewProduct = (product: Product) => {
  currentProduct.value = product;
  showDetailModal.value = true;
};

const closeDetailModal = () => {
  showDetailModal.value = false;
  currentProduct.value = null;
};

const editProduct = (product: Product) => {
  router.push(`/product/spus/edit/${product.id}`);
};

const publishProduct = (product: Product) => {
  product.published = true;
  // 这里应该调用API更新商品状态
  alert(`${product.name} 已发布`);
};

const unpublishProduct = (product: Product) => {
  product.published = false;
  // 这里应该调用API更新商品状态
  alert(`${product.name} 已取消发布`);
};

const enableProduct = (product: Product) => {
  product.status = "active";
  // 这里应该调用API更新商品状态
  alert(`${product.name} 已启用`);
};

const disableProduct = (product: Product) => {
  product.status = "inactive";
  // 这里应该调用API更新商品状态
  alert(`${product.name} 已禁用`);
};

const batchPublish = () => {
  selectedProducts.value.forEach(product => {
    product.published = true;
  });
  selectedProducts.value = [];
  selectAll.value = false;
  // 这里应该调用API批量更新商品状态
};

const batchUnpublish = () => {
  selectedProducts.value.forEach(product => {
    product.published = false;
  });
  selectedProducts.value = [];
  selectAll.value = false;
  // 这里应该调用API批量更新商品状态
};

const batchEnable = () => {
  selectedProducts.value.forEach(product => {
    product.status = "active";
  });
  selectedProducts.value = [];
  selectAll.value = false;
  // 这里应该调用API批量更新商品状态
};

const batchDisable = () => {
  selectedProducts.value.forEach(product => {
    product.status = "inactive";
  });
  selectedProducts.value = [];
  selectAll.value = false;
  // 这里应该调用API批量更新商品状态
};

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "CNY",
  }).format(value);
};
</script>
