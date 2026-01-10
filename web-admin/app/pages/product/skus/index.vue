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
      <div class="flex flex-wrap justify-end gap-2">
        <UButton
          variant="outline"
          color="neutral"
          icon="i-heroicons-arrow-up-tray"
          @click="openImportModal"
        >
          {{ $t('product.sku.batchImport') }}
        </UButton>
        <UButton
          variant="outline"
          color="neutral"
          icon="i-heroicons-arrow-down-tray"
          @click="openExportModal"
        >
          {{ $t('product.sku.batchExport') }}
        </UButton>
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
          :items="statusOptions"
          :placeholder="$t('product.sku.status')"
        />
        <USelectMenu
          v-model="spuFilterModel"
          :items="spuOptionsWithAll"
          value-key="value"
          label-key="label"
          :portal="false"
          :placeholder="$t('product.sku.spu')"
          class="w-full"
        />
      </div>
      <div class="flex justify-end mt-4 space-x-2">
        <UButton @click="resetFilters" variant="ghost">{{ $t('common.reset') }}</UButton>
        <UButton @click="searchSkus">{{ $t('common.search') }}</UButton>
      </div>
    </UCard>

    <BulkTaskStatusList
      class="mb-6"
      :tasks="bulkTasks"
      :selected-task-id="highlightedTaskId || undefined"
      @select="handleTaskSelect"
      @review="handleReviewRequest"
    />

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
          size="sm"
          icon="i-heroicons-adjustments-horizontal"
          @click="openBulkAdjustModal"
        >
          {{ $t('product.sku.bulkAdjust') }}
        </UButton>
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
        {{ $t('common.total') }} {{ totalSkus || skus.length }} {{ $t('product.sku.skus') }}
      </div>
      <UPagination
        v-model="page"
        :page-count="pageCount"
        :total="totalSkus || filteredSkus.length"
      />
    </div>

    <!-- 批量导入模态框 -->
    <UModal
      v-model:open="showImportModal"
      :title="$t('product.sku.batchImport')"
      :close="{ onClick: () => closeImportModal() }"
      :ui="{
        content: 'w-full sm:max-w-4xl',
        body: 'p-0',
      }"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <BulkImportUploader @submitted="handleImportSubmitted" />
        </div>
      </template>
    </UModal>

    <!-- 批量导出模态框 -->
    <UModal
      v-model:open="showExportModal"
      :title="$t('product.sku.batchExport')"
      :close="{ onClick: () => closeExportModal() }"
      :ui="{
        content: 'w-full sm:max-w-3xl',
        body: 'p-0',
      }"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <BulkExportPanel
            :keyword="searchQuery"
            :status="statusFilter"
            @exported="handleExportCompleted"
          />
        </div>
      </template>
    </UModal>

    <!-- 创建SKU模态框 -->
    <UModal
      v-model:open="showCreateModal"
      :title="$t('product.sku.add')"
      :description="$t('product.sku.createDescription')"
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
                <USelectMenu
                  v-model="currentSkuSpuIdModel"
                  :items="spuOptions"
                  value-key="value"
                  label-key="label"
                  :placeholder="$t('product.sku.selectSpu')"
                  :portal="false"
                  searchable
                  :ui="{ content: 'z-[80]' }"
                  class="w-full"
                />
              </UFormField>

              <UFormField :label="$t('product.sku.specifications')">
                <div class="space-y-2">
                  <p class="text-sm text-gray-600 dark:text-gray-400">
                    SKU 规格来自 SPU 的规格定义，请在 SPU 页面通过「关联 SKU / 批量生成」创建变体。
                  </p>
                  <UButton
                    v-if="currentSku.spuId"
                    color="primary"
                    variant="outline"
                    size="sm"
                    icon="i-heroicons-arrow-top-right-on-square"
                    @click="goToSpuSkuGenerator(currentSku.spuId)"
                  >
                    去该 SPU 生成 SKU
                  </UButton>
                </div>
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
          <UButton color="primary" :loading="savingSku" @click="saveSku">{{ $t('common.save') }}</UButton>
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
                <p class="text-gray-900 dark:text-white">
                  {{ getSpuLabel(currentSku.spuId) || currentSku.spu || "-" }}
                </p>
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

    <BulkAdjustModal
      v-model="showBulkAdjustModal"
      :selection="bulkSelection"
      @submitted="handleBulkTaskSubmitted"
    />
    <ApprovalDrawer
      v-model="approvalDrawerOpen"
      :task="approvalTarget"
      :loading="approvalProcessing"
      @approve="(note) => handleApprovalDecision('approve', note)"
      @reject="(note) => handleApprovalDecision('reject', note)"
    />
  </div>
</template>

<script setup lang="ts">
import { h } from "vue";
import { storeToRefs } from "pinia";
import type { TableColumn } from "@nuxt/ui";
import { useToastAlert } from "~/composables/useToastAlert";
import BulkAdjustModal from "~/components/product/sku/BulkAdjustModal.vue";
import BulkImportUploader from "~/components/product/sku/BulkImportUploader.vue";
import BulkExportPanel from "~/components/product/sku/BulkExportPanel.vue";
import BulkTaskStatusList from "~/components/product/sku/BulkTaskStatusList.vue";
import ApprovalDrawer from "~/components/product/sku/ApprovalDrawer.vue";
import { useSkuBulkTaskTracker } from "~/composables/useSkuBulkTaskTracker";
import { useSkuBulkActions } from "~/composables/useSkuBulkActions";
import { useSkuApi } from "~/composables/api/useSku";
import { useSpuApi } from "~/composables/api/useSpu";
import { useProductSkuStore } from "~/stores/productSku";
import type { ProductSku, SkuBulkTask, SkuGeneratorDefaults, SkuUpsertRequest } from "~/types/product/sku";

const { t } = useI18n();
const router = useRouter();
const toast = useToastAlert();
const bulkActions = useSkuBulkActions();
const { tasks: bulkTasks } = useSkuBulkTaskTracker();
const skuApi = useSkuApi();
const spuApi = useSpuApi();
const skuStore = useProductSkuStore();
const { items: storeSkus, total, loading: storeLoading } = storeToRefs(skuStore);

// 状态管理
const loading = computed(() => storeLoading.value);
const page = ref(1);
const pageSize = ref(10);
const searchQuery = ref("");
const statusFilter = ref<string>("__all__");
const spuFilter = ref<string>("__all__");
const selectAll = ref(false);
const selectedSkus = ref<Sku[]>([]);
const showCreateModal = ref(false);
const showDetailModal = ref(false);
const showBulkAdjustModal = ref(false);
const showImportModal = ref(false);
const showExportModal = ref(false);
const approvalDrawerOpen = ref(false);
const approvalTarget = ref<SkuBulkTask | null>(null);
const approvalProcessing = ref(false);
const highlightedTaskId = ref<string | undefined>(undefined);
const activeTab = ref("basic");
const detailTab = ref("basic");
const savingSku = ref(false);

// 当前编辑的SKU
const currentSku = ref({
  id: "",
  skuCode: "",
  barcode: "",
  spuId: "",
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

const totalSkus = computed(() => total.value || 0);
const spuOptions = ref<{ label: string; value: string }[]>([]);
const spuLookup = ref<Record<string, string>>({});
const spuOptionsWithAll = computed(() => [
  { label: t("common.all"), value: "__all__" },
  ...spuOptions.value,
]);

const getSpuLabel = (spuId?: string) => {
  if (!spuId) {
    return "";
  }
  return spuLookup.value[spuId] ?? "";
};

const normalizeSelectValueToString = (value: any) => {
  if (!value) {
    return "";
  }
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "object") {
    const candidate = (value as any).value ?? (value as any).id;
    if (typeof candidate === "string") {
      return candidate;
    }
  }
  return "";
};

const spuFilterModel = computed({
  get: () => spuFilter.value,
  set: (value) => {
    const normalized = normalizeSelectValueToString(value);
    spuFilter.value = normalized || "__all__";
  },
});

const currentSkuSpuIdModel = computed({
  get: () => currentSku.value.spuId,
  set: (value) => {
    currentSku.value.spuId = normalizeSelectValueToString(value);
  },
});

// 筛选选项
const statusOptions = [
  { label: t("common.all"), value: "__all__" },
  { label: t("status.active"), value: "active" },
  { label: t("status.inactive"), value: "inactive" }
];

// SKU数据类型
type Sku = {
  id: string;
  skuCode: string;
  barcode: string;
  spuId?: string;
  spuName?: string;
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
  entity?: ProductSku;
};

// SKU数据
const skus = computed(() => storeSkus.value.map((item) => formatSkuRow(item)));

const mapApiStatusToUi = (status?: string): "active" | "inactive" => {
  if (!status) {
    return "inactive";
  }
  return status === "online" || status === "ready" ? "active" : "inactive";
};

const mapUiStatusToApi = (status?: string) => {
  if (!status || status === "__all__") {
    return undefined;
  }
  if (status === "active") {
    return "online";
  }
  if (status === "inactive") {
    return "offline";
  }
  return undefined;
};

const buildDefaultValues = (): SkuGeneratorDefaults => {
  const defaults: SkuGeneratorDefaults = {};
  if (currentSku.value.moq) {
    defaults.minOrderQty = currentSku.value.moq;
  }
  if (currentSku.value.weight) {
    defaults.weight = Number(currentSku.value.weight);
  }
  if (currentSku.value.packageDimensions) {
    defaults.dimensions = currentSku.value.packageDimensions;
  }
  return defaults;
};

const formatSpecs = (specs?: ProductSku["specs"]) => {
  if (!specs?.length) {
    return "";
  }
  return specs
    .map((spec) => {
      const name = spec.specName || spec.specId;
      const value = spec.valueName || spec.valueId;
      if (name && value) {
        return `${name}: ${value}`;
      }
      return value || name || "";
    })
    .filter(Boolean)
    .join(" / ");
};

const normalizeApiSku = (item: any): ProductSku => {
  if (!item) {
    return {
      id: "",
      tenantUuid: "",
      spuId: "",
      skuCode: "",
      specs: [],
      status: "draft",
    };
  }
  const specs = Array.isArray(item.specs)
    ? item.specs.map((spec: any) => ({
        specId: spec?.spec_id ?? spec?.specId ?? "",
        specName: spec?.spec_name ?? spec?.specName ?? "",
        valueId: spec?.value_id ?? spec?.valueId ?? "",
        valueName: spec?.value_name ?? spec?.valueName ?? "",
      })).filter((spec: any) => spec.specId || spec.valueId)
    : [];
  const priceRefs = item.price_refs ?? item.priceRefs;
  const salePrice = typeof item.sale_price === "number" ? item.sale_price : typeof item.salePrice === "number" ? item.salePrice : undefined;
  const currency = typeof item.currency === "string" ? item.currency : typeof item.currencyCode === "string" ? item.currencyCode : undefined;
  const fallbackPriceRefs =
    !priceRefs && typeof salePrice === "number" && Number.isFinite(salePrice) && salePrice > 0
      ? [
          {
            priceListId: "base",
            currency: currency || "CNY",
            tiers: [{ minQty: 1, price: salePrice }],
          },
        ]
      : undefined;
  return {
    id: item.id ?? "",
    tenantUuid: item.tenant_uuid ?? item.tenantUuid ?? "",
    spuId: item.spu_id ?? item.spuId ?? "",
    spuName: item.spu_name ?? item.spuName,
    skuCode: item.sku_code ?? item.skuCode ?? "",
    specs,
    specDisplay: item.spec_display ?? item.specDisplay,
    barcode: item.barcode,
    status: item.status ?? "draft",
    minOrderQty: item.min_order_qty ?? item.minOrderQty,
    priceRefs: priceRefs ?? fallbackPriceRefs,
    logistics: item.logistics ?? item.default_values ?? item.logisticsSnapshot,
    inventory: item.inventory ?? [],
    createdAt: item.created_at ?? item.createdAt,
    updatedAt: item.updated_at ?? item.updatedAt,
  };
};

const formatSkuRow = (rawItem: ProductSku | any): Sku => {
  const item = normalizeApiSku(rawItem);
  const inventory = item.inventory?.[0];
  const availableStock = inventory?.availableQty ?? 0;
  const safetyStock = inventory?.safetyStock ?? 0;
  const priceRef = item.priceRefs?.[0];
  const priceTier = priceRef?.tiers?.[0];
  const label = item.spuName || getSpuLabel(item.spuId) || item.spuId;
  const weight = item.logistics?.weight ?? 0;
  const packageDimensions = item.logistics?.dimensions ?? "";
  return {
    id: item.id,
    skuCode: item.skuCode,
    barcode: item.barcode ?? "",
    spuId: item.spuId,
    spuName: item.spuName,
    spu: label,
    specifications: item.specDisplay || formatSpecs(item.specs),
    description: "",
    price: priceTier?.price ?? 0,
    availableStock,
    safetyStock,
    warningThreshold: safetyStock,
    enableSerialNumber: false,
    enableBatchNumber: false,
    weight,
    volume: 0,
    packageDimensions,
    moq: item.minOrderQty ?? 1,
    purchaseLimit: 0,
    status: mapApiStatusToUi(item.status),
    inventoryStatus: availableStock <= safetyStock ? "lowStock" : "inStock",
    entity: item,
  };
};

const loadSkus = async () => {
  try {
    await skuStore.fetchList({
      spuId: spuFilter.value === "__all__" ? undefined : spuFilter.value || undefined,
      status: mapUiStatusToApi(statusFilter.value) || undefined,
      page: page.value,
      pageSize: pageSize.value,
    });
    selectAll.value = false;
    selectedSkus.value = [];
  } catch (error: any) {
    toast.add({
      title: t("message.error"),
      description: error?.message ?? "Load SKU failed",
      color: "red",
    });
  }
};

const loadSpuOptions = async () => {
  try {
    const response = await spuApi
      .listSpus({
        pageSize: 100,
      })
      .catch(() => null);
    const items = response?.items ?? [];
    const options = items.map((item) => ({
      label: item.name || item.code,
      value: item.id,
    }));
    const lookup: Record<string, string> = {};
    options.forEach((opt) => {
      lookup[opt.value] = opt.label;
    });
    spuOptions.value = options;
    spuLookup.value = lookup;
  } catch (error: any) {
    toast.add({
      title: t("message.error"),
      description: error?.message ?? "Load SPU failed",
      color: "red",
    });
  }
};

onMounted(() => {
  loadSpuOptions();
  loadSkus();
});

watch(
  () => [page.value, pageSize.value],
  () => {
    loadSkus();
  },
);

watch(
  () => [statusFilter.value, spuFilter.value],
  () => {
    const changed = page.value !== 1;
    page.value = 1;
    if (!changed) {
      loadSkus();
    }
  },
);

// 计算属性
const filteredSkus = computed(() => {
  return skus.value.filter(sku => {
    const matchesSearch = !searchQuery.value ||
      sku.skuCode.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      sku.barcode.includes(searchQuery.value) ||
      sku.spu.toLowerCase().includes(searchQuery.value.toLowerCase());

    const matchesStatus = statusFilter.value === "__all__" || sku.status === statusFilter.value;
    const matchesSpu = spuFilter.value === "__all__" || sku.spuId === spuFilter.value;

    return matchesSearch && matchesStatus && matchesSpu;
  });
});

const pageCount = computed(() => {
  if (!pageSize.value) {
    return 1;
  }
  const total = totalSkus.value || filteredSkus.value.length || 0;
  return Math.max(1, Math.ceil(total / pageSize.value));
});

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

const blurActiveElement = () => {
  if (typeof document === "undefined") {
    return;
  }
  const active = document.activeElement as HTMLElement | null;
  if (active && typeof active.blur === "function") {
    active.blur();
  }
};

const openImportModal = () => {
  blurActiveElement();
  showImportModal.value = true;
};

const closeImportModal = () => {
  blurActiveElement();
  showImportModal.value = false;
};

const openExportModal = () => {
  blurActiveElement();
  showExportModal.value = true;
};

const closeExportModal = () => {
  blurActiveElement();
  showExportModal.value = false;
};

const handleImportSubmitted = (taskId?: string) => {
  highlightedTaskId.value = taskId;
  closeImportModal();
};

const handleExportCompleted = (taskId?: string) => {
  highlightedTaskId.value = taskId;
  closeExportModal();
};

// 方法
const searchSkus = () => {
  const changed = page.value !== 1;
  page.value = 1;
  if (!changed) {
    loadSkus();
  }
};

const resetFilters = () => {
  searchQuery.value = "";
  statusFilter.value = "__all__";
  spuFilter.value = "__all__";
  const changed = page.value !== 1;
  page.value = 1;
  if (!changed) {
    loadSkus();
  }
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
    spuId: "",
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
    spuId: "",
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
  if (sku.id) {
    router.push(`/product/skus/${sku.id}`);
    return;
  }
  openDetailModal(sku);
};

const editSku = (sku: Sku) => {
  currentSku.value = { ...sku };
  activeTab.value = "basic";
  showCreateModal.value = true;
};

const deleteSku = (sku: Sku) => {
  if (confirm(`${t("message.confirm.delete")} SKU: ${sku.skuCode}?`)) {
    // TODO: 接入真实删除接口
  }
};

const saveSku = async () => {
  if (savingSku.value) {
    return;
  }
  const trimmedCode = currentSku.value.skuCode.trim();
  if (!trimmedCode) {
    toast.add({ title: t("product.sku.toast.codeRequired"), color: "error" });
    return;
  }
  if (!currentSku.value.spuId) {
    toast.add({ title: t("product.sku.toast.spuRequired"), color: "error" });
    return;
  }
  savingSku.value = true;
  try {
    const statusForApi = mapUiStatusToApi(currentSku.value.status) || "draft";
    const payload: SkuUpsertRequest = {
      skus: [
        {
          spuId: currentSku.value.spuId,
          skuCode: trimmedCode,
          barcode: currentSku.value.barcode?.trim() || undefined,
          status: statusForApi,
          minOrderQty: currentSku.value.moq || undefined,
          defaultValues: buildDefaultValues(),
          specs: [],
        },
      ],
    };
    const result = await skuStore.createSkus(payload);
    const created = result?.created ?? 0;
    if (created > 0) {
      toast.add({ title: t("product.sku.toast.saveSuccess"), color: "success" });
      closeCreateModal();
      await loadSkus();
    } else {
      const skipped = result?.skipped?.[0];
      toast.add({
        title: t("product.sku.toast.duplicateTitle"),
        description: skipped ? t("product.sku.toast.duplicateDesc", { code: skipped }) : t("product.sku.toast.saveFailed"),
        color: "warning",
      });
    }
  } catch (error: any) {
    toast.add({
      title: t("product.sku.toast.saveFailed"),
      description: error?.message || "",
      color: "error",
    });
  } finally {
    savingSku.value = false;
  }
};

const mapSkuToProductSku = (sku: Sku): ProductSku => {
  if (sku.entity) {
    return sku.entity;
  }
  return {
    id: sku.id,
    tenantUuid: "",
    spuId: sku.spuId || sku.spu,
    skuCode: sku.skuCode,
    specs: [],
    barcode: sku.barcode,
    status: sku.status === "inactive" ? "offline" : "online",
    minOrderQty: sku.moq,
    priceRefs: sku.price
      ? [
          {
            priceListId: "default",
            currency: "CNY",
            tiers: [{ minQty: 1, price: sku.price }],
          },
        ]
      : undefined,
    inventory: [
      {
        warehouseId: "default",
        availableQty: sku.availableStock ?? 0,
        lockedQty: 0,
        inTransitQty: 0,
        safetyStock: sku.safetyStock ?? 0,
      },
    ],
  };
};

const goToSpuSkuGenerator = (spuId: string) => {
  if (!spuId) {
    return;
  }
  router.push({ path: `/product/spus/edit/${spuId}`, query: { panel: "sku" } });
};

const bulkSelection = computed<ProductSku[]>(() =>
  selectedSkus.value.map(mapSkuToProductSku)
);

const openBulkAdjustModal = () => {
  if (!selectedSkus.value.length) {
    toast.add({ title: t("product.sku.bulkSelectHint"), color: "red" });
    return;
  }
  showBulkAdjustModal.value = true;
};

const handleBulkTaskSubmitted = () => {
  showBulkAdjustModal.value = false;
  selectAll.value = false;
  selectedSkus.value = [];
};

watch(
  () => bulkTasks.value.length,
  () => {
    if (!bulkTasks.value.length) {
      highlightedTaskId.value = undefined;
    } else if (!highlightedTaskId.value) {
      highlightedTaskId.value = bulkTasks.value[0]?.taskId;
    }
  },
  { immediate: true },
);

const handleTaskSelect = (task: SkuBulkTask) => {
  highlightedTaskId.value = task.taskId;
};

const handleReviewRequest = (task: SkuBulkTask) => {
  highlightedTaskId.value = task.taskId;
  approvalTarget.value = task;
  approvalDrawerOpen.value = true;
};

const handleApprovalDecision = async (decision: "approve" | "reject", note: string) => {
  if (!approvalTarget.value) return;
  approvalProcessing.value = true;
  try {
    await bulkActions.decideApproval({
      taskId: approvalTarget.value.taskId,
      decision,
      note,
    });
    toast.add({
      title:
        decision === "approve"
          ? t("product.sku.approval.toastApproved")
          : t("product.sku.approval.toastRejected"),
      description: note,
      color: decision === "approve" ? "success" : "warning",
    });
    approvalDrawerOpen.value = false;
  } catch (error: any) {
    toast.add({
      title: t("product.sku.approval.toastFailed"),
      description: error?.message ?? "",
      color: "red",
    });
  } finally {
    approvalProcessing.value = false;
  }
};
</script>
