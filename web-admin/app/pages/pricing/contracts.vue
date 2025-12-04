<template>
  <div class="p-6 space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
        {{ $t("pricing.contracts") }}
      </h1>
      <div class="flex gap-3">
        <UButton
          color="gray"
          variant="outline"
          icon="i-heroicons-document-arrow-down"
          @click="exportContracts"
        >
          导出合同
        </UButton>
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="createContract"
        >
          新建合同价格
        </UButton>
      </div>
    </div>

    <!-- 统计卡片（动态） -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <UCard>
        <div class="flex items-center">
          <UIcon
            name="i-heroicons-document-text"
            class="w-8 h-8 text-blue-500"
          />
          <div class="ml-4">
            <p class="text-sm text-gray-500">总合同数</p>
            <p class="text-2xl font-bold">{{ contracts.length }}</p>
          </div>
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center">
          <UIcon
            name="i-heroicons-check-circle"
            class="w-8 h-8 text-green-500"
          />
          <div class="ml-4">
            <p class="text-sm text-gray-500">生效中</p>
            <p class="text-2xl font-bold">{{ activeCount }}</p>
          </div>
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center">
          <UIcon name="i-heroicons-clock" class="w-8 h-8 text-yellow-500" />
          <div class="ml-4">
            <p class="text-sm text-gray-500">即将到期</p>
            <p class="text-2xl font-bold">{{ expiringSoonCount }}</p>
          </div>
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center">
          <UIcon
            name="i-heroicons-currency-yen"
            class="w-8 h-8 text-purple-500"
          />
          <div class="ml-4">
            <p class="text-sm text-gray-500">本月合同额</p>
            <p class="text-2xl font-bold">¥{{ monthlyAmount }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 筛选 -->
    <UCard>
      <div class="grid grid-cols-1 md:grid-cols-5 gap-4">
        <UInput
          v-model="searchQuery"
          placeholder="搜索客户或合同"
          icon="i-heroicons-magnifying-glass"
        />
        <USelect
          v-model="selectedCustomerType"
          :options="customerTypeOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="客户类型"
        />
        <USelect
          v-model="selectedStatus"
          :options="statusOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="合同状态"
        />
        <USelect
          v-model="selectedChannel"
          :options="channelOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="适用渠道"
        />
        <UDatePicker v-model="selectedDateRange" range placeholder="到期时间" />
      </div>
    </UCard>

    <!-- 列表 -->
    <UCard>
      <UTable :data="filteredContracts" :columns="columns" :loading="loading">
        <!-- 客户信息 -->
        <template #customer-cell="{ row }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="row.original.customer.avatar"
              :alt="row.original.customer.name"
              size="sm"
            />
            <div>
              <p class="font-medium">{{ row.original.customer.name }}</p>
              <p class="text-xs text-gray-500">
                {{ row.original.customer.code }}
              </p>
            </div>
          </div>
        </template>

        <!-- 合同类型 -->
        <template #contractType-cell="{ row }">
          <UBadge
            :color="getContractTypeColor(row.original.contractType)"
            variant="subtle"
          >
            {{ row.original.contractType }}
          </UBadge>
        </template>

        <!-- 价格信息 -->
        <template #priceInfo-cell="{ row }">
          <div class="text-sm">
            <p class="font-medium">{{ row.original.priceInfo.type }}</p>
            <p class="text-xs text-gray-500">
              {{ row.original.priceInfo.value }}
            </p>
          </div>
        </template>

        <!-- 订购量 -->
        <template #volume-cell="{ row }">
          <div class="text-sm">
            <p class="font-medium">{{ row.original.volume.committed }}</p>
            <div class="w-full bg-gray-200 rounded-full h-2 mt-1">
              <div
                class="bg-blue-500 h-2 rounded-full"
                :style="{ width: `${row.original.volume.progress}%` }"
              ></div>
            </div>
            <p class="text-xs text-gray-500 mt-1">
              {{ row.original.volume.progress }}% 完成
            </p>
          </div>
        </template>

        <!-- 状态 -->
        <template #status-cell="{ row }">
          <div class="flex items-center gap-2">
            <UBadge
              :color="getStatusColor(row.original.status)"
              variant="subtle"
            >
              {{ row.original.status }}
            </UBadge>
            <span
              v-if="
                row.original.daysToExpiry <= EXPIRY_THRESHOLD &&
                row.original.daysToExpiry >= 0
              "
              class="text-xs text-red-500"
            >
              {{ row.original.daysToExpiry }}天后到期
            </span>
          </div>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="gray"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewContract(row.original)"
              >查看</UButton
            >
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
              @click="editContract(row.original)"
              >编辑</UButton
            >
            <UButton
              color="green"
              variant="ghost"
              size="sm"
              icon="i-heroicons-arrow-path"
              @click="renewContract(row.original)"
              >续约</UButton
            >
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 详情 -->
    <UModal v-model="showDetailModal" :ui="{ width: 'max-w-5xl' }">
      <UCard>
        <template #header>
          <div class="flex justify-between items-center">
            <h3 class="text-lg font-semibold">
              {{ selectedContract?.contractNumber }}
            </h3>
            <UButton
              color="gray"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="showDetailModal = false"
            />
          </div>
        </template>

        <div v-if="selectedContract" class="space-y-6">
          <!-- 基本信息 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h4 class="font-medium mb-3">合同信息</h4>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-gray-600">合同编号:</span
                  ><span>{{ selectedContract.contractNumber }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">合同类型:</span>
                  <UBadge
                    :color="getContractTypeColor(selectedContract.contractType)"
                    variant="subtle"
                    >{{ selectedContract.contractType }}</UBadge
                  >
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">签约日期:</span
                  ><span>{{ selectedContract.signDate }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">生效日期:</span
                  ><span>{{ selectedContract.effectiveDate }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">到期日期:</span
                  ><span>{{ selectedContract.expiryDate }}</span>
                </div>
              </div>
            </div>
            <div>
              <h4 class="font-medium mb-3">客户信息</h4>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-gray-600">客户名称:</span
                  ><span>{{ selectedContract.customer.name }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">客户代码:</span
                  ><span>{{ selectedContract.customer.code }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">客户类型:</span
                  ><span>{{ selectedContract.customer.type }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">信用等级:</span>
                  <UBadge
                    :color="
                      getCreditColor(selectedContract.customer.creditRating)
                    "
                    variant="subtle"
                    >{{ selectedContract.customer.creditRating }}</UBadge
                  >
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">销售代表:</span
                  ><span>{{ selectedContract.customer.salesRep }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 价格条款 -->
          <div>
            <h4 class="font-medium mb-3">价格条款</h4>
            <div class="bg-gray-50 p-4 rounded-lg">
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                <div>
                  <span class="text-gray-600">定价类型:</span>
                  <p class="font-medium">
                    {{ selectedContract.priceInfo.type }}
                  </p>
                </div>
                <div>
                  <span class="text-gray-600">价格值:</span>
                  <p class="font-medium">
                    {{ selectedContract.priceInfo.value }}
                  </p>
                </div>
                <div>
                  <span class="text-gray-600">基准价格:</span>
                  <p class="font-medium">
                    {{ selectedContract.priceInfo.basePrice }}
                  </p>
                </div>
              </div>
              <div class="mt-4 grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                <div>
                  <span class="text-gray-600">最低订购量:</span>
                  <p class="font-medium">
                    {{ selectedContract.volume.minimum }}
                  </p>
                </div>
                <div>
                  <span class="text-gray-600">承诺订购量:</span>
                  <p class="font-medium">
                    {{ selectedContract.volume.committed }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- 商品清单 -->
          <div>
            <h4 class="font-medium mb-3">
              适用商品 ({{ contractProducts.length }})
            </h4>
            <UTable
              :data="contractProducts"
              :columns="productColumns"
              class="border"
            >
              <template #product-cell="{ row }">
                <div class="flex items-center gap-3">
                  <img
                    :src="row.original.image"
                    :alt="row.original.name"
                    class="w-10 h-10 rounded object-cover"
                  />
                  <div>
                    <p class="font-medium">{{ row.original.name }}</p>
                    <p class="text-xs text-gray-500">{{ row.original.sku }}</p>
                  </div>
                </div>
              </template>
              <template #contractPrice-cell="{ row }">
                <div class="text-right">
                  <p class="font-medium">¥{{ row.original.contractPrice }}</p>
                  <p class="text-xs text-gray-500">
                    市场价: ¥{{ row.original.marketPrice }}
                  </p>
                </div>
              </template>
              <template #discount-cell="{ row }">
                <div class="text-right">
                  <span class="text-green-600 font-medium"
                    >{{ row.original.discount }}%</span
                  >
                </div>
              </template>
              <template #volume-cell="{ row }">
                <div class="text-right">
                  <p class="font-medium">{{ row.original.orderedQty }}</p>
                  <p class="text-xs text-gray-500">
                    / {{ row.original.committedQty }}
                  </p>
                </div>
              </template>
            </UTable>
          </div>

          <!-- 执行情况 -->
          <div>
            <h4 class="font-medium mb-3">执行情况</h4>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <UCard
                ><div class="text-center">
                  <p class="text-2xl font-bold text-blue-600">
                    {{ selectedContract.performance.totalOrders }}
                  </p>
                  <p class="text-sm text-gray-500">总订单数</p>
                </div></UCard
              >
              <UCard
                ><div class="text-center">
                  <p class="text-2xl font-bold text-green-600">
                    ¥{{ selectedContract.performance.totalAmount }}
                  </p>
                  <p class="text-sm text-gray-500">总金额</p>
                </div></UCard
              >
              <UCard
                ><div class="text-center">
                  <p class="text-2xl font-bold text-purple-600">
                    {{ selectedContract.performance.avgOrderValue }}
                  </p>
                  <p class="text-sm text-gray-500">平均订单价值</p>
                </div></UCard
              >
              <UCard
                ><div class="text-center">
                  <p class="text-2xl font-bold text-orange-600">
                    {{ selectedContract.performance.fulfillmentRate }}%
                  </p>
                  <p class="text-sm text-gray-500">履约率</p>
                </div></UCard
              >
            </div>
          </div>
        </div>
      </UCard>
    </UModal>
    <ContractPricingWizard
      v-model="showWizard"
      :initial-data="wizardInitial"
      @saved="
        (payload) => {
          // 这里可以把新建/编辑结果合并到 contracts 列表，或触发刷新
          // console.log('saved payload', payload)
        }
      "
    />
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import ContractPricingWizard from "~/components/ContractPricingWizard.vue";

const { t } = useI18n();
const toast = useToast();

/** 常量 */
const EXPIRY_THRESHOLD = 30; // 天

/** 响应式状态 */
const searchQuery = ref("");
const selectedCustomerType = ref("");
const selectedStatus = ref("");
const selectedChannel = ref("");
const selectedDateRange = ref<any>(null);
const loading = ref(false);
const showDetailModal = ref(false);
const selectedContract = ref<any>(null);

/** 选项 */
const customerTypeOptions = [
  { label: "全部类型", value: "" },
  { label: "企业客户", value: "企业客户" },
  { label: "经销商", value: "经销商" },
  { label: "代理商", value: "代理商" },
  { label: "零售商", value: "零售商" },
];
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "生效中", value: "生效中" },
  { label: "即将到期", value: "即将到期" },
  { label: "已到期", value: "已到期" },
  { label: "已终止", value: "已终止" },
];
const channelOptions = [
  { label: "全部渠道", value: "" },
  { label: "直销", value: "直销" },
  { label: "经销", value: "经销" },
  { label: "在线", value: "在线" },
  { label: "零售", value: "零售" },
];

/** 列定义 */
type ContractRow = (typeof contracts.value)[number];
const columns: TableColumn<ContractRow>[] = [
  { accessorKey: "contractNumber", header: "合同编号" },
  { accessorKey: "customer", header: "客户信息" },
  { accessorKey: "contractType", header: "合同类型" },
  { accessorKey: "priceInfo", header: "价格信息" },
  { accessorKey: "volume", header: "订购量" },
  { accessorKey: "expiryDate", header: "到期日期" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
];

const productColumns: TableColumn<any>[] = [
  { accessorKey: "product", header: "商品" },
  { accessorKey: "contractPrice", header: "合同价格" },
  { accessorKey: "discount", header: "折扣率" },
  { accessorKey: "volume", header: "订购量" },
];

/** 数据（示例） */
const contracts = ref([
  {
    id: "CT001",
    contractNumber: "CT-2024-001",
    customer: {
      name: "华为技术有限公司",
      code: "HW001",
      type: "企业客户",
      creditRating: "AAA",
      salesRep: "张经理",
      avatar: "https://api.dicebear.com/7.x/initials/svg?seed=HW",
    },
    contractType: "年度框架协议",
    priceInfo: { type: "阶梯折扣", value: "5%-15%", basePrice: "市场价" },
    volume: { minimum: "1000万", committed: "5000万", progress: 68 },
    signDate: "2024-01-01",
    effectiveDate: "2024-01-01",
    expiryDate: "2024-12-31",
    status: "生效中",
    daysToExpiry: 342,
    channels: ["直销", "在线"],
    performance: {
      totalOrders: 156,
      totalAmount: "3,400万",
      avgOrderValue: "21.8万",
      fulfillmentRate: 95.2,
    },
  },
  {
    id: "CT002",
    contractNumber: "CT-2024-002",
    customer: {
      name: "小米科技有限公司",
      code: "MI001",
      type: "企业客户",
      creditRating: "AA+",
      salesRep: "李经理",
      avatar: "https://api.dicebear.com/7.x/initials/svg?seed=MI",
    },
    contractType: "季度采购协议",
    priceInfo: { type: "固定折扣", value: "8%", basePrice: "标准价" },
    volume: { minimum: "500万", committed: "2000万", progress: 45 },
    signDate: "2024-01-15",
    effectiveDate: "2024-01-15",
    expiryDate: "2024-04-15",
    status: "即将到期",
    daysToExpiry: 25,
    channels: ["直销"],
    performance: {
      totalOrders: 89,
      totalAmount: "900万",
      avgOrderValue: "10.1万",
      fulfillmentRate: 88.7,
    },
  },
  {
    id: "CT003",
    contractNumber: "CT-2023-089",
    customer: {
      name: "苏宁易购集团",
      code: "SN001",
      type: "经销商",
      creditRating: "AA",
      salesRep: "王经理",
      avatar: "https://api.dicebear.com/7.x/initials/svg?seed=SN",
    },
    contractType: "经销商协议",
    priceInfo: { type: "批发价", value: "标准批发价-3%", basePrice: "批发价" },
    volume: { minimum: "2000万", committed: "8000万", progress: 92 },
    signDate: "2023-03-01",
    effectiveDate: "2023-03-01",
    expiryDate: "2024-02-29",
    status: "已到期",
    daysToExpiry: -15,
    channels: ["经销", "零售"],
    performance: {
      totalOrders: 234,
      totalAmount: "7,360万",
      avgOrderValue: "31.5万",
      fulfillmentRate: 96.8,
    },
  },
]);

const contractProducts = ref([
  {
    id: "P001",
    name: "iPhone 15 Pro Max",
    sku: "IP15PM-256-NT",
    image: "https://via.placeholder.com/40x40",
    contractPrice: "8,999",
    marketPrice: "9,999",
    discount: 10,
    committedQty: "1000台",
    orderedQty: "680台",
  },
  {
    id: "P002",
    name: "MacBook Air M2",
    sku: "MBA-M2-256-SG",
    image: "https://via.placeholder.com/40x40",
    contractPrice: "7,999",
    marketPrice: "8,999",
    discount: 11.1,
    committedQty: "500台",
    orderedQty: "340台",
  },
  {
    id: "P003",
    name: "AirPods Pro 2",
    sku: "APP2-USB-C",
    image: "https://via.placeholder.com/40x40",
    contractPrice: "1,799",
    marketPrice: "1,999",
    discount: 10,
    committedQty: "2000台",
    orderedQty: "1,360台",
  },
]);

/** 过滤 */
const filteredContracts = computed(() => {
  let list = contracts.value;

  // 搜索
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(
      (c) =>
        c.customer.name.toLowerCase().includes(q) ||
        c.contractNumber.toLowerCase().includes(q)
    );
  }
  // 客户类型
  if (selectedCustomerType.value) {
    list = list.filter((c) => c.customer.type === selectedCustomerType.value);
  }
  // 状态
  if (selectedStatus.value) {
    list = list.filter((c) => c.status === selectedStatus.value);
  }
  // 渠道
  if (selectedChannel.value) {
    list = list.filter((c) => c.channels.includes(selectedChannel.value));
  }
  // 到期时间范围（UDatePicker range: [startDate, endDate]）
  if (
    selectedDateRange.value?.length === 2 &&
    selectedDateRange.value[0] &&
    selectedDateRange.value[1]
  ) {
    const [start, end] = selectedDateRange.value;
    const s = new Date(start as string).getTime();
    const e = new Date(end as string).getTime();
    list = list.filter((c) => {
      const d = new Date(c.expiryDate).getTime();
      return d >= s && d <= e;
    });
  }

  return list;
});

/** 派生指标 */
const activeCount = computed(
  () => contracts.value.filter((c) => c.status === "生效中").length
);
const expiringSoonCount = computed(
  () =>
    contracts.value.filter(
      (c) => c.daysToExpiry >= 0 && c.daysToExpiry <= EXPIRY_THRESHOLD
    ).length
);
const monthlyAmount = computed(() => {
  // demo：取本月“生效中”的 totalAmount 字符串求和（仅示意）
  const pickup = contracts.value.filter((c) => c.status === "生效中");
  const sum = pickup.reduce(
    (acc, c) => acc + toNumber(c.performance.totalAmount),
    0
  );
  return formatNumber(sum);
});

/** 工具函数 */
function getContractTypeColor(type: string) {
  return (
    {
      年度框架协议: "blue",
      季度采购协议: "green",
      经销商协议: "purple",
      零售协议: "orange",
    }[type] || "neutral"
  );
}
function getStatusColor(status: string) {
  return (
    {
      生效中: "success",
      即将到期: "warning",
      已到期: "error",
      已终止: "neutral",
    }[status] || "neutral"
  );
}
function getCreditColor(rating: string) {
  return (
    {
      AAA: "success",
      "AA+": "success",
      AA: "warning",
      "A+": "warning",
      A: "neutral",
    }[rating] || "neutral"
  );
}
function toNumber(s: string) {
  // 简易把“3,400万” => 34000000
  if (s.endsWith("万")) return Number(s.replace(/[,万]/g, "")) * 10000;
  return Number(s.replace(/,/g, "")) || 0;
}
function formatNumber(n: number) {
  return n.toLocaleString();
}

/** 行为 */
function viewContract(contract: any) {
  selectedContract.value = contract;
  showDetailModal.value = true;
}

function renewContract(contract: any) {
  toast.add({ title: `续约：${contract.contractNumber}`, color: "green" });
}
function exportContracts() {
  toast.add({ title: "已开始导出（示例）", color: "gray" });
}

const showWizard = ref(false);
function createContract() {
  showWizard.value = true;
}
function editContract(contract: any) {
  // 这里把 contract 映射为向导 initialData 结构（示例简单映射）
  const initial = {
    customer: contract.customer,
    channels: contract.channels,
    dateRange: [contract.effectiveDate, contract.expiryDate],
    currency: "CNY",
    pricing: {
      mode: "固定折扣",
      value: "8",
      basePrice: "标准价",
      tiers: [{ min: 10, discount: 10 }],
      rounding: "四舍五入到元",
    },
    products: [] as any[],
    approval: {
      approver: "定价经理",
      priority: "中",
      autoActivate: true,
      notes: "",
    },
  };
  wizardInitial.value = initial;
  showWizard.value = true;
}
const wizardInitial = ref<any>(null);
</script>
