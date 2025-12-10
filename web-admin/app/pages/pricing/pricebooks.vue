<template>
  <div class="p-6 space-y-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
        {{ $t("pricing.pricebooks") }}
      </h1>
      <div class="flex gap-3">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-heroicons-arrow-down-tray"
          @click="onExport"
        >
          {{ $t("common.export") }}
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus" @click="onCreate">
          {{ $t("common.add") }}{{ $t("pricing.pricebooks") }}
        </UButton>
      </div>
    </div>

    <!-- 筛选 -->
    <UCard>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UInput
          v-model="searchQuery"
          :placeholder="$t('common.search')"
          icon="i-heroicons-magnifying-glass"
        />
        <USelect
          v-model="selectedChannel"
          :options="channelOptions"
          option-attribute="label"
          value-attribute="value"
          :placeholder="$t('channel.title')"
        />
        <USelect
          v-model="selectedRegion"
          :options="regionOptions"
          option-attribute="label"
          value-attribute="value"
          :placeholder="$t('pricing.region')"
        />
        <USelect
          v-model="selectedStatus"
          :options="statusOptions"
          option-attribute="label"
          value-attribute="value"
          :placeholder="$t('common.status')"
        />
      </div>
    </UCard>

    <!-- 列表 -->
    <UCard>
      <UTable :data="filteredPricebooks" :columns="columns" :loading="loading">
        <!-- 状态 -->
        <template #status-cell="{ row }">
          <UBadge :color="getStatusColor(row.original.status)" variant="subtle">
            {{ row.original.status }}
          </UBadge>
        </template>

        <!-- 有效期 -->
        <template #effectiveRange-cell="{ row }">
          <div>
            <p class="text-sm">{{ row.original.effectiveDate }}</p>
            <p class="text-xs text-gray-500">
              至 {{ row.original.expiryDate }}
            </p>
          </div>
        </template>

        <!-- 商品数量 -->
        <template #productCount-cell="{ row }">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium">{{
              row.original.productCount
            }}</span>
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              @click="viewProducts(row.original)"
            >
              查看
            </UButton>
          </div>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewPricebook(row.original)"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
              @click="editPricebook(row.original)"
            >
              {{ $t("common.edit") }}
            </UButton>
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-document-duplicate"
              @click="duplicatePricebook(row.original)"
            >
              复制
            </UButton>
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              @click="deletePricebook(row.original)"
            >
              {{ $t("common.delete") }}
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 详情弹窗 -->
    <UModal v-model="showDetailModal" :ui="{ width: 'max-w-4xl' }">
      <UCard>
        <template #header>
          <div class="flex justify-between items-center">
            <h3 class="text-lg font-semibold">{{ selectedPricebook?.name }}</h3>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="showDetailModal = false"
            />
          </div>
        </template>

        <div v-if="selectedPricebook" class="space-y-6">
          <!-- 基本信息 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h4 class="font-medium mb-3">基本信息</h4>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-gray-600">渠道:</span>
                  <span>{{ selectedPricebook.channel }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">地区:</span>
                  <span>{{ selectedPricebook.region }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">币种:</span>
                  <span>{{ selectedPricebook.currency }}</span>
                </div>
                <div class="flex justify-between items-center">
                  <span class="text-gray-600">状态:</span>
                  <UBadge
                    :color="getStatusColor(selectedPricebook.status)"
                    variant="subtle"
                  >
                    {{ selectedPricebook.status }}
                  </UBadge>
                </div>
              </div>
            </div>
            <div>
              <h4 class="font-medium mb-3">有效期</h4>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-gray-600">生效日期:</span>
                  <span>{{ selectedPricebook.effectiveDate }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">失效日期:</span>
                  <span>{{ selectedPricebook.expiryDate }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">创建时间:</span>
                  <span>{{ selectedPricebook.createdAt }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">更新时间:</span>
                  <span>{{ selectedPricebook.updatedAt }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 商品价格 -->
          <div>
            <h4 class="font-medium mb-3">
              商品价格 ({{ pricebookProducts.length }})
            </h4>
            <UTable
              :data="pricebookProducts"
              :columns="productColumns"
              class="border"
            >
              <template #price-cell="{ row }">
                <div class="text-right">
                  <p class="font-medium">¥{{ row.original.price }}</p>
                  <p class="text-xs text-gray-500">
                    MSRP: ¥{{ row.original.msrp }}
                  </p>
                </div>
              </template>
              <template #margin-cell="{ row }">
                <div class="text-right">
                  <span
                    :class="
                      row.original.margin >= 30
                        ? 'text-green-600'
                        : row.original.margin >= 20
                          ? 'text-yellow-600'
                          : 'text-red-600'
                    "
                  >
                    {{ row.original.margin }}%
                  </span>
                </div>
              </template>
            </UTable>
          </div>
        </div>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { h } from "vue";
import type { TableColumn } from "@nuxt/ui";
import { useToastAlert } from "~/composables/useToastAlert";

const { t } = useI18n();
const toast = useToastAlert();

/** 类型 */
type Pricebook = {
  id: string;
  name: string;
  channel: string;
  region: string;
  currency: string;
  productCount: number;
  effectiveDate: string;
  expiryDate: string;
  status: "生效中" | "待生效" | "已失效" | "草稿";
  createdAt: string;
  updatedAt: string;
};
type PricebookProduct = {
  sku: string;
  name: string;
  category: string;
  price: string;
  msrp: string;
  margin: number;
};

/** 状态 */
const searchQuery = ref("");
const selectedChannel = ref<string>(""); // 注意：值与数据里中文一致
const selectedRegion = ref<string>(""); // 同上
const selectedStatus = ref<string>(""); // 同上
const loading = ref(false);

const showDetailModal = ref(false);
const selectedPricebook = ref<Pricebook | null>(null);

/** 选项（value 与数据字段保持一致，避免匹配失败） */
const channelOptions = [
  { label: "全部渠道", value: "" },
  { label: "天猫旗舰店", value: "天猫旗舰店" },
  { label: "京东自营", value: "京东自营" },
  { label: "线下门店", value: "线下门店" },
  { label: "企业直销", value: "企业直销" },
];
const regionOptions = [
  { label: "全部地区", value: "" },
  { label: "全国", value: "全国" },
  { label: "华东", value: "华东" },
  { label: "华南", value: "华南" },
  { label: "华北", value: "华北" },
];
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "生效中", value: "生效中" },
  { label: "待生效", value: "待生效" },
  { label: "已失效", value: "已失效" },
  { label: "草稿", value: "草稿" },
];

/** 列定义（使用 accessorKey + -cell 插槽） */
const columns: TableColumn<Pricebook>[] = [
  { accessorKey: "name", header: "价目表名称" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "region", header: "地区" },
  { accessorKey: "currency", header: "币种" },
  { accessorKey: "productCount", header: "商品数量" },
  { id: "effectiveRange", header: "有效期" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
];

const productColumns: TableColumn<PricebookProduct>[] = [
  { accessorKey: "sku", header: "SKU" },
  { accessorKey: "name", header: "商品名称" },
  { accessorKey: "category", header: "分类" },
  {
    id: "price",
    accessorKey: "price",
    header: () => h("div", { class: "text-right" }, "价格"),
    meta: { class: { td: "text-right" } },
  },
  {
    id: "margin",
    accessorKey: "margin",
    header: () => h("div", { class: "text-right" }, "毛利率"),
    meta: { class: { td: "text-right" } },
  },
];

/** 数据 */
const pricebooks = ref<Pricebook[]>([
  {
    id: "PB001",
    name: "天猫旗舰店标准价目表",
    channel: "天猫旗舰店",
    region: "全国",
    currency: "CNY",
    productCount: 156,
    effectiveDate: "2024-01-01",
    expiryDate: "2024-12-31",
    status: "生效中",
    createdAt: "2023-12-15",
    updatedAt: "2024-01-10",
  },
  {
    id: "PB002",
    name: "京东自营促销价目表",
    channel: "京东自营",
    region: "全国",
    currency: "CNY",
    productCount: 89,
    effectiveDate: "2024-02-01",
    expiryDate: "2024-02-29",
    status: "生效中",
    createdAt: "2024-01-20",
    updatedAt: "2024-01-25",
  },
  {
    id: "PB003",
    name: "华东地区线下门店价目表",
    channel: "线下门店",
    region: "华东",
    currency: "CNY",
    productCount: 234,
    effectiveDate: "2024-03-01",
    expiryDate: "2024-05-31",
    status: "待生效",
    createdAt: "2024-02-10",
    updatedAt: "2024-02-15",
  },
  {
    id: "PB004",
    name: "企业客户批发价目表",
    channel: "企业直销",
    region: "全国",
    currency: "CNY",
    productCount: 67,
    effectiveDate: "2023-12-01",
    expiryDate: "2023-12-31",
    status: "已失效",
    createdAt: "2023-11-15",
    updatedAt: "2023-11-20",
  },
  {
    id: "PB005",
    name: "春节特惠价目表",
    channel: "天猫旗舰店",
    region: "全国",
    currency: "CNY",
    productCount: 45,
    effectiveDate: "2024-02-10",
    expiryDate: "2024-02-17",
    status: "草稿",
    createdAt: "2024-01-30",
    updatedAt: "2024-02-01",
  },
]);

const pricebookProducts = ref<PricebookProduct[]>([
  {
    sku: "IP15P-128-BLK",
    name: "iPhone 15 Pro 128GB 黑色",
    category: "智能手机",
    price: "8999",
    msrp: "9999",
    margin: 28,
  },
  {
    sku: "MBA-M2-256-SLV",
    name: "MacBook Air M2 256GB 银色",
    category: "笔记本电脑",
    price: "9499",
    msrp: "10499",
    margin: 32,
  },
  {
    sku: "APP-3G-WHT",
    name: "AirPods Pro 3代 白色",
    category: "音频设备",
    price: "1899",
    msrp: "2199",
    margin: 25,
  },
]);

/** 过滤 */
const filteredPricebooks = computed(() => {
  let list = pricebooks.value;

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (pb) =>
        pb.name.toLowerCase().includes(q) ||
        pb.channel.toLowerCase().includes(q)
    );
  }
  if (selectedChannel.value) {
    list = list.filter((pb) => pb.channel === selectedChannel.value);
  }
  if (selectedRegion.value) {
    list = list.filter((pb) => pb.region === selectedRegion.value);
  }
  if (selectedStatus.value) {
    list = list.filter((pb) => pb.status === selectedStatus.value);
  }
  return list;
});

/** 工具 */
function getStatusColor(status: Pricebook["status"]) {
  const map: Record<Pricebook["status"], string> = {
    生效中: "green",
    待生效: "blue",
    已失效: "gray",
    草稿: "yellow",
  };
  return map[status] ?? "gray";
}

/** 行为 */
function viewPricebook(pb: Pricebook) {
  selectedPricebook.value = pb;
  showDetailModal.value = true;
}
function editPricebook(pb: Pricebook) {
  toast.add({ title: `编辑：${pb.name}`, color: "primary" });
}
function duplicatePricebook(pb: Pricebook) {
  // 深拷贝一份为草稿
  const copy: Pricebook = {
    ...pb,
    id: `${pb.id}-COPY`,
    name: `${pb.name}（副本）`,
    status: "草稿",
    createdAt: new Date().toISOString().slice(0, 10),
    updatedAt: new Date().toISOString().slice(0, 10),
  };
  pricebooks.value.unshift(copy);
  toast.add({ title: "已复制为草稿", color: "neutral" });
}
function deletePricebook(pb: Pricebook) {
  pricebooks.value = pricebooks.value.filter((p) => p.id !== pb.id);
  toast.add({ title: "已删除", color: "error" });
}
function viewProducts(pb: Pricebook) {
  viewPricebook(pb);
}
function onExport() {
  toast.add({ title: "开始导出价目表（示例）", color: "neutral" });
}
function onCreate() {
  toast.add({ title: "新建价目表（示例）", color: "primary" });
}
</script>
