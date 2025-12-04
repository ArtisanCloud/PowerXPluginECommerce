<template>
  <div class="max-w-6xl mx-auto">
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          {{ $t('product.sku.title') }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理商品SKU和变体信息
        </p>
      </div>
      <div class="flex space-x-2">
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="openCreateModal"
        >
          {{ $t('product.sku.add') }}
        </UButton>
      </div>
    </div>

    <!-- 筛选和搜索栏 -->
    <UCard class="mb-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UInput
          v-model="searchQuery"
          :placeholder="$t('product.sku.search')"
          icon="i-heroicons-magnifying-glass"
        />
        <USelect
          v-model="statusFilter"
          :options="statusOptions"
          :placeholder="$t('product.sku.status')"
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
        <UButton @click="searchSkus">{{ $t('common.search') }}</UButton>
      </div>
    </UCard>

    <!-- 批量操作栏 -->
    <div class="flex justify-between items-center mb-4">
      <div class="flex items-center space-x-2">
        <UCheckbox v-model="selectAll" @change="toggleSelectAll" />
        <span class="text-sm text-gray-600 dark:text-gray-400">
          {{ selectedSkus.length }} {{ $t('common.selected') }}
        </span>
        <UButton
          v-if="selectedSkus.length > 0"
          color="primary"
          variant="outline"
          size="sm"
          @click="batchImport"
        >
          {{ $t('product.sku.batchImport') }}
        </UButton>
        <UButton
          v-if="selectedSkus.length > 0"
          color="primary"
          variant="outline"
          size="sm"
          @click="batchExport"
        >
          {{ $t('product.sku.batchExport') }}
        </UButton>
      </div>
      <div class="flex space-x-2">
        <UButton variant="outline" icon="i-heroicons-arrow-down-tray">{{ $t('common.export') }}</UButton>
        <UButton variant="outline" icon="i-heroicons-arrow-up-tray">{{ $t('common.import') }}</UButton>
      </div>
    </div>

    <!-- SKU表格 -->
    <UCard>
      <UTable
        v-model="selectedSkus"
        :data="filteredSkus"
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

        <!-- 库存状态列 -->
        <template #inventoryStatus-cell="{ getValue }">
          <UBadge
            :color="getValue() === 'inStock' ? 'success' : 'warning'"
            variant="subtle"
          >
            {{ getValue() === "inStock" ? $t("product.sku.inStock") : $t("product.sku.lowStock") }}
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
              @click="viewSku(row.original)"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
              @click="editSku(row.original)"
            >
              {{ $t("common.edit") }}
            </UButton>
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              @click="deleteSku(row.original)"
            >
              {{ $t("common.delete") }}
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 分页 -->
    <div class="flex justify-between items-center mt-4">
      <div class="text-sm text-gray-600 dark:text-gray-400">
        {{ $t('common.total') }} {{ skus.length }} {{ $t('product.sku.skus') }}
      </div>
      <UPagination
        v-model="page"
        :page-count="pageCount"
        :total="filteredSkus.length"
      />
    </div>

    <!-- 创建SKU模态框 -->
    <UModal
      v-model:open="showCreateModal"
      :title="$t('product.sku.add')"
      :close="{ onClick: () => closeCreateModal() }"
      :ui="{
        content: 'w-full sm:max-w-4xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <div class="space-y-6 p-6">
          <!-- Tabs -->
          <UTabs v-model="activeTab" :items="tabs" class="mb-6" />

          <!-- 基础信息 -->
          <div v-if="activeTab === 'basic'">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <UFormField :label="$t('product.sku.skuCode')" required>
                <UInput v-model="currentSku.skuCode" :placeholder="$t('product.sku.skuCode')" />
              </UFormField>

              <UFormField :label="$t('product.sku.barcode')">
                <UInput v-model="currentSku.barcode" :placeholder="$t('product.sku.barcode')" />
              </UFormField>

              <UFormField :label="$t('product.sku.spu')">
                <USelect
                  v-model="currentSku.spu"
                  :options="spuOptions"
                  :placeholder="$t('product.sku.selectSpu')"
                />
              </UFormField>

              <UFormField :label="$t('product.sku.specifications')">
                <UInput v-model="currentSku.specifications" :placeholder="$t('product.sku.specifications')" />
              </UFormField>

              <div class="md:col-span-2">
                <UFormField :label="$t('product.sku.description')">
                  <UTextarea
                    v-model="currentSku.description"
                    :placeholder="$t('product.sku.description')"
                    :rows="3"
                  />
                </UFormField>
              </div>
            </div>
          </div>

          <!-- 库存信息 -->
          <div v-if="activeTab === 'inventory'">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <UFormField :label="$t('product.sku.availableStock')">
                <UInput
                  v-model.number="currentSku.availableStock"
                  type="number"
                  min="0"
                  :placeholder="$t('product.sku.availableStock')"
                />
              </UFormField>

              <UFormField :label="$t('product.sku.safetyStock')">
                <UInput
                  v-model.number="currentSku.safetyStock"
                  type="number"
                  min="0"
                  :placeholder="$t('product.sku.safetyStock')"
                />
              </UFormField>

              <UFormField :label="$t('product.sku.warningThreshold')">
                <UInput
                  v-model.number="currentSku.warningThreshold"
                  type="number"
                  min="0"
                  :placeholder="$t('product.sku.warningThreshold')"
                />
              </UFormField>
            </div>

            <div class="mt-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                {{ $t('product.sku.serialBatchManagement') }}
              </h3>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <UFormField :label="$t('product.sku.enableSerialNumber')">
                  <USwitch v-model="currentSku.enableSerialNumber" />
                </UFormField>

                <UFormField :label="$t('product.sku.enableBatchNumber')">
                  <USwitch v-model="currentSku.enableBatchNumber" />
                </UFormField>
              </div>
            </div>
          </div>

          <!-- 物流信息 -->
          <div v-if="activeTab === 'logistics'">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <UFormField :label="$t('product.sku.weight')">
                <UInput
                  v-model.number="currentSku.weight"
                  type="number"
                  min="0"
                  step="0.01"
                  :placeholder="$t('product.sku.weight')"
                />
              </UFormField>

              <UFormField :label="$t('product.sku.volume')">
                <UInput
                  v-model.number="currentSku.volume"
                  type="number"
                  min="0"
                  step="0.01"
                  :placeholder="$t('product.sku.volume')"
                />
              </UFormField>

              <UFormField :label="$t('product.sku.packageDimensions')">
                <UInput
                  v-model="currentSku.packageDimensions"
                  :placeholder="$t('product.sku.packageDimensions')"
                />
              </UFormField>
            </div>
          </div>

          <!-- 约束信息 -->
          <div v-if="activeTab === 'constraints'">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <UFormField :label="$t('product.sku.moq')">
                <UInput
                  v-model.number="currentSku.moq"
                  type="number"
                  min="1"
                  :placeholder="$t('product.sku.moq')"
                />
              </UFormField>

              <UFormField :label="$t('product.sku.purchaseLimit')">
                <UInput
                  v-model.number="currentSku.purchaseLimit"
                  type="number"
                  min="0"
                  :placeholder="$t('product.sku.purchaseLimit')"
                />
              </UFormField>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end space-x-2">
          <UButton variant="ghost" @click="closeCreateModal">{{ $t('common.cancel') }}</UButton>
          <UButton color="primary" @click="saveSku">{{ $t('common.save') }}</UButton>
        </div>
      </template>
    </UModal>

    <!-- 查看SKU详情模态框 -->
    <UModal
      v-model:open="showDetailModal"
      :title="$t('product.sku.detail')"
      :close="{ onClick: () => closeDetailModal() }"
      :ui="{
        content: 'w-full sm:max-w-4xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <div v-if="currentSku" class="space-y-6 p-6">
          <UTabs v-model="detailTab" :items="detailTabs" class="mb-6" />

          <!-- 基础信息详情 -->
          <div v-if="detailTab === 'basic'">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.skuCode') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.skuCode }}</p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.barcode') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.barcode }}</p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.spu') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.spu }}</p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.specifications') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.specifications }}</p>
              </div>

              <div class="md:col-span-2">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.description') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.description }}</p>
              </div>
            </div>
          </div>

          <!-- 库存信息详情 -->
          <div v-if="detailTab === 'inventory'">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.availableStock') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.availableStock }}</p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.safetyStock') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.safetyStock }}</p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.warningThreshold') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.warningThreshold }}</p>
              </div>
            </div>

            <div class="mt-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                {{ $t('product.sku.serialBatchManagement') }}
              </h3>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    {{ $t('product.sku.enableSerialNumber') }}
                  </label>
                  <UBadge :color="currentSku.enableSerialNumber ? 'success' : 'neutral'">
                    {{ currentSku.enableSerialNumber ? $t('common.yes') : $t('common.no') }}
                  </UBadge>
                </div>

                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    {{ $t('product.sku.enableBatchNumber') }}
                  </label>
                  <UBadge :color="currentSku.enableBatchNumber ? 'success' : 'neutral'">
                    {{ currentSku.enableBatchNumber ? $t('common.yes') : $t('common.no') }}
                  </UBadge>
                </div>
              </div>
            </div>
          </div>

          <!-- 物流信息详情 -->
          <div v-if="detailTab === 'logistics'">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.weight') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.weight }} kg</p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.volume') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.volume }} m³</p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.packageDimensions') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.packageDimensions }}</p>
              </div>
            </div>
          </div>

          <!-- 约束信息详情 -->
          <div v-if="detailTab === 'constraints'">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.moq') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.moq }}</p>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ $t('product.sku.purchaseLimit') }}
                </label>
                <p class="text-gray-900 dark:text-white">{{ currentSku.purchaseLimit }}</p>
              </div>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end space-x-2">
          <UButton variant="ghost" @click="closeDetailModal">{{ $t('common.close') }}</UButton>
          <UButton color="primary" @click="editSku(currentSku)">{{ $t('common.edit') }}</UButton>
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
const selectedSkus = ref([]);
const showCreateModal = ref(false);
const showDetailModal = ref(false);
const activeTab = ref("basic");
const detailTab = ref("basic");

// 当前编辑的SKU
const currentSku = ref({
  id: "",
  skuCode: "",
  barcode: "",
  spu: "",
  specifications: "",
  description: "",
  availableStock: 0,
  safetyStock: 0,
  warningThreshold: 0,
  enableSerialNumber: false,
  enableBatchNumber: false,
  weight: 0,
  volume: 0,
  packageDimensions: "",
  moq: 1,
  purchaseLimit: 0,
  status: "active",
  inventoryStatus: "inStock"
});

// Tabs
const tabs = [
  { label: t("product.sku.basicInfo"), value: "basic" },
  { label: t("product.sku.inventory"), value: "inventory" },
  { label: t("product.sku.logistics"), value: "logistics" },
  { label: t("product.sku.constraints"), value: "constraints" },
];

const detailTabs = [
  { label: t("product.sku.basicInfo"), value: "basic" },
  { label: t("product.sku.inventory"), value: "inventory" },
  { label: t("product.sku.logistics"), value: "logistics" },
  { label: t("product.sku.constraints"), value: "constraints" },
];

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

const spuOptions = [
  { label: "iPhone 15 Pro", value: "IP15P" },
  { label: "MacBook Pro", value: "MBP" },
  { label: "AirPods Pro", value: "APP" }
];

// SKU数据类型
type Sku = {
  id: string;
  skuCode: string;
  barcode: string;
  spu: string;
  specifications: string;
  description: string;
  price: number;
  availableStock: number;
  safetyStock: number;
  warningThreshold: number;
  enableSerialNumber: boolean;
  enableBatchNumber: boolean;
  weight: number;
  volume: number;
  packageDimensions: string;
  moq: number;
  purchaseLimit: number;
  status: "active" | "inactive";
  inventoryStatus: "inStock" | "lowStock";
};

// SKU数据
const skus = ref<Sku[]>([
  {
    id: "SKU001",
    skuCode: "IP15P-128-BLK",
    barcode: "1234567890123",
    spu: "iPhone 15 Pro",
    specifications: "128GB, 黑色",
    description: "iPhone 15 Pro 128GB 黑色版本",
    price: 8999,
    availableStock: 20,
    safetyStock: 5,
    warningThreshold: 10,
    enableSerialNumber: true,
    enableBatchNumber: false,
    weight: 0.195,
    volume: 0.0001,
    packageDimensions: "150x80x20mm",
    moq: 1,
    purchaseLimit: 5,
    status: "active",
    inventoryStatus: "inStock"
  },
  {
    id: "SKU002",
    skuCode: "IP15P-256-BLK",
    barcode: "1234567890124",
    spu: "iPhone 15 Pro",
    specifications: "256GB, 黑色",
    description: "iPhone 15 Pro 256GB 黑色版本",
    price: 9999,
    availableStock: 15,
    safetyStock: 5,
    warningThreshold: 10,
    enableSerialNumber: true,
    enableBatchNumber: false,
    weight: 0.195,
    volume: 0.0001,
    packageDimensions: "150x80x20mm",
    moq: 1,
    purchaseLimit: 5,
    status: "active",
    inventoryStatus: "inStock"
  },
  {
    id: "SKU003",
    skuCode: "MBP-512-SLV",
    barcode: "1234567890125",
    spu: "MacBook Pro",
    specifications: "512GB, 银色",
    description: "MacBook Pro 512GB 银色版本",
    price: 12999,
    availableStock: 8,
    safetyStock: 3,
    warningThreshold: 5,
    enableSerialNumber: true,
    enableBatchNumber: false,
    weight: 1.5,
    volume: 0.001,
    packageDimensions: "300x220x20mm",
    moq: 1,
    purchaseLimit: 3,
    status: "active",
    inventoryStatus: "lowStock"
  }
]);

// 计算属性
const filteredSkus = computed(() => {
  return skus.value.filter(sku => {
    const matchesSearch = !searchQuery.value ||
      sku.skuCode.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      sku.barcode.includes(searchQuery.value) ||
      sku.spu.toLowerCase().includes(searchQuery.value.toLowerCase());

    const matchesStatus = !statusFilter.value || sku.status === statusFilter.value;
    const matchesCategory = !categoryFilter.value || sku.spu.includes(categoryFilter.value);
    const matchesBrand = !brandFilter.value || sku.spu.includes(brandFilter.value);

    return matchesSearch && matchesStatus && matchesCategory && matchesBrand;
  });
});

const pageCount = computed(() =>
  Math.ceil(filteredSkus.value.length / pageSize.value)
);

// 列定义
const columns = computed<TableColumn<Sku>[]>(() => [
  { accessorKey: "skuCode", header: t("product.sku.skuCode") },
  { accessorKey: "spu", header: t("product.sku.spu") },
  { accessorKey: "specifications", header: t("product.sku.specifications") },
  {
    accessorKey: "price",
    header: t("product.sku.price"),
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
  { accessorKey: "availableStock", header: t("product.sku.availableStock") },
  { accessorKey: "inventoryStatus", header: t("product.sku.inventoryStatus") },
  { accessorKey: "status", header: t("product.sku.status") },
  { id: "actions", header: t("product.sku.actions") },
]);

// 方法
const searchSkus = () => {
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
    selectedSkus.value = [...filteredSkus.value];
  } else {
    selectedSkus.value = [];
  }
};

const openCreateModal = () => {
  currentSku.value = {
    id: "",
    skuCode: "",
    barcode: "",
    spu: "",
    specifications: "",
    description: "",
    availableStock: 0,
    safetyStock: 0,
    warningThreshold: 0,
    enableSerialNumber: false,
    enableBatchNumber: false,
    weight: 0,
    volume: 0,
    packageDimensions: "",
    moq: 1,
    purchaseLimit: 0,
    status: "active",
    inventoryStatus: "inStock"
  };
  activeTab.value = "basic";
  showCreateModal.value = true;
};

const closeCreateModal = () => {
  showCreateModal.value = false;
};

const openDetailModal = (sku: Sku) => {
  currentSku.value = { ...sku };
  detailTab.value = "basic";
  showDetailModal.value = true;
};

const closeDetailModal = () => {
  showDetailModal.value = false;
  currentSku.value = {
    id: "",
    skuCode: "",
    barcode: "",
    spu: "",
    specifications: "",
    description: "",
    availableStock: 0,
    safetyStock: 0,
    warningThreshold: 0,
    enableSerialNumber: false,
    enableBatchNumber: false,
    weight: 0,
    volume: 0,
    packageDimensions: "",
    moq: 1,
    purchaseLimit: 0,
    status: "active",
    inventoryStatus: "inStock"
  };
};

const viewSku = (sku: Sku) => {
  openDetailModal(sku);
};

const editSku = (sku: Sku) => {
  currentSku.value = { ...sku };
  activeTab.value = "basic";
  showCreateModal.value = true;
};

const deleteSku = (sku: Sku) => {
  if (confirm(`${t("message.confirm.delete")} SKU: ${sku.skuCode}?`)) {
    skus.value = skus.value.filter(s => s.id !== sku.id);
  }
};

const saveSku = () => {
  // 这里应该调用API保存SKU
  if (currentSku.value.id) {
    // 更新现有SKU
    const index = skus.value.findIndex(s => s.id === currentSku.value.id);
    if (index !== -1) {
      skus.value[index] = { ...currentSku.value };
    }
  } else {
    // 创建新SKU
    currentSku.value.id = `SKU${Date.now()}`;
    skus.value.push({ ...currentSku.value });
  }
  closeCreateModal();
  alert(t("message.success.saved"));
};

const batchImport = () => {
  alert(t("product.sku.batchImport"));
};

const batchExport = () => {
  alert(t("product.sku.batchExport"));
};
</script>
