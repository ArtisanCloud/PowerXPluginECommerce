<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        {{ $t("inventory.stockTitle") }}
      </h1>
      <div class="flex space-x-2">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-heroicons-arrow-down-tray"
        >
          {{ $t("common.import") }}
        </UButton>
        <UButton
          color="primary"
          variant="solid"
          icon="i-heroicons-arrow-up-tray"
        >
          {{ $t("common.export") }}
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold">{{ $t("inventory.stockListTitle") }}</h3>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="searchQuery"
              :placeholder="$t('common.search')"
              icon="i-heroicons-magnifying-glass"
              class="w-64"
            />
            <USelect
              v-model="selectedStatus"
              :items="stockStatuses"
              :placeholder="$t('common.filter')"
              class="w-40"
            />
          </div>
        </div>
      </template>

      <UTable
        :data="filteredInventory"
        :columns="columns"
        :loading="loading"
        class="w-full"
      >
        <template #stock-cell="{ row }">
          <div class="flex items-center gap-2">
            <span>{{ row.original.stock }}</span>
            <UBadge
              :color="getStockColor(row.original.stock, row.original.minStock)"
              variant="subtle"
              size="xs"
            >
              {{ getStockStatus(row.original.stock, row.original.minStock) }}
            </UBadge>
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              size="xs"
              color="warning"
              variant="ghost"
              @click="openAdjust(row.original)"
            >
              调整库存
            </UButton>
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              icon="i-heroicons-eye"
              @click="viewSku(row.original)"
            >
              {{ $t("common.view") }}
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UModal
      v-model:open="adjustOpen"
      :title="'调整库存'"
      :description="adjustTarget ? `${adjustTarget.productId} · ${adjustTarget.productName} · ${adjustTarget.sku}` : '选择一个 SKU 后调整库存'"
      :prevent-close="adjustSubmitting"
      :ui="{ content: 'max-w-3xl w-full max-h-[calc(100dvh-2rem)] overflow-hidden' }"
    >
      <template #body>
        <UForm id="inventory-adjust-form" :state="{ adjustDelta }" class="space-y-4 p-1" @submit.prevent="submitAdjust">
          <UFormField label="库存增量（delta）" hint="正数增加，负数减少；结果不可为负数">
            <UInput v-model="adjustDelta" type="number" inputmode="numeric" step="1" />
          </UFormField>

          <UAlert v-if="adjustError" color="error" variant="soft" :title="$t('common.error')" :description="adjustError" />
        </UForm>
      </template>

      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end sm:gap-3">
          <UButton color="neutral" variant="outline" :disabled="adjustSubmitting" @click="closeAdjust">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton type="submit" form="inventory-adjust-form" color="primary" :loading="adjustSubmitting">
            {{ $t("common.confirm") }}
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { h } from "vue";
import type { TableColumn } from "@nuxt/ui";
import { useSkuApi } from "~/composables/api/useSku";

const { t } = useI18n();
const skuApi = useSkuApi();
const router = useRouter();

const searchQuery = ref("");
const ALL_STATUS = "__all__";
const selectedStatus = ref<string>(ALL_STATUS);
const loading = ref(false);
const adjustOpen = ref(false);
const adjustSubmitting = ref(false);
const adjustError = ref("");
const adjustDelta = ref<number | null>(null);

const stockStatuses = [
  // reka-ui SelectItem 不允许 value 为空字符串（空字符串用于“清空选择并显示 placeholder”）
  { label: "全部状态", value: ALL_STATUS },
  { label: "正常", value: "正常" },
  { label: "库存不足", value: "库存不足" },
  { label: "缺货", value: "缺货" },
];

type Inventory = {
  skuId: string;
  productId: string;
  productName: string;
  sku: string;
  stock: number;
  minStock: number;
  warehouse: string;
  lastUpdate: string;
};

const columns = computed<TableColumn<Inventory>[]>(() => [
  { accessorKey: "productId", header: t("inventory.productId") || "商品ID" },
  {
    accessorKey: "productName",
    header: t("inventory.productName") || "商品名称",
  },
  { accessorKey: "sku", header: t("inventory.sku") || "SKU" },
  { accessorKey: "stock", header: t("inventory.stock") || "当前库存" },
  { accessorKey: "minStock", header: t("inventory.minStock") || "最低库存" },
  { accessorKey: "warehouse", header: t("inventory.warehouse") || "仓库" },
  {
    accessorKey: "lastUpdate",
    header: t("inventory.lastUpdate") || "最后更新",
    cell: ({ getValue }) => {
      const v = String(getValue() || "");
      return h("span", {}, v.replace(" ", " "));
    },
  },
  { id: "actions", header: t("common.actions") || "操作" },
]);

const inventory = ref<Inventory[]>([]);
const adjustTarget = ref<Inventory | null>(null);

const toInventoryRow = async (raw: any): Promise<Inventory> => {
  const skuId = String(raw?.id || "");
  const skuCode = String(raw?.sku_code ?? raw?.skuCode ?? "");
  const spuName = String(raw?.spu_name ?? raw?.spuName ?? "");
  const specDisplay = String(raw?.spec_display ?? raw?.specDisplay ?? "");

  const snapshot = await skuApi.getInventorySnapshot(skuId);
  const warehouses = (snapshot as any)?.warehouses ?? [];
  const wh = warehouses.find((w: any) => String(w?.warehouse_id ?? w?.warehouseId ?? "") === "default");
  const stock = Number(wh?.available_qty ?? wh?.availableQty ?? 0);
  const minStock = Number(wh?.safety_stock ?? wh?.safetyStock ?? 0);
  const warehouse = String(wh?.warehouse_id ?? wh?.warehouseId ?? "default");
  const lastUpdate = String(
    wh?.last_synced_at ??
      wh?.lastSyncedAt ??
      (snapshot as any)?.last_synced_at ??
      (snapshot as any)?.lastSyncedAt ??
      ""
  );
  return {
    skuId,
    productId: skuCode || skuId,
    productName: spuName || "-",
    sku: specDisplay || skuCode,
    stock,
    minStock,
    warehouse,
    lastUpdate,
  };
};

const loadInventory = async () => {
  loading.value = true;
  try {
    const res = await skuApi.list({ page: 1, pageSize: 50, locale: "zh-CN" } as any);
    const items = (res?.items || []) as any[];
    const results = await Promise.allSettled(items.map((it) => toInventoryRow(it)));
    inventory.value = results
      .filter((r): r is PromiseFulfilledResult<Inventory> => r.status === "fulfilled")
      .map((r) => r.value);
  } catch (e) {
    console.error("Error loading inventory stock list:", e);
    inventory.value = [];
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  void loadInventory();
});

const viewSku = async (row: Inventory) => {
  if (!row?.skuId) return;
  await router.push(`/product/skus/${row.skuId}?from=inventory&tab=inventory`);
};

const openAdjust = (row: Inventory) => {
  adjustTarget.value = row;
  adjustError.value = "";
  adjustDelta.value = null;
  adjustOpen.value = true;
};

const closeAdjust = () => {
  if (typeof window !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur?.();
  }
  adjustOpen.value = false;
};

const submitAdjust = async () => {
  if (!adjustTarget.value?.skuId) return;
  const delta = Number(adjustDelta.value);
  if (!Number.isFinite(delta) || delta === 0) {
    adjustError.value = "请输入非 0 的整数（delta）";
    return;
  }

  adjustSubmitting.value = true;
  adjustError.value = "";
  try {
    const snap = await skuApi.adjustInventory(adjustTarget.value.skuId, { delta });
    const warehouses = (snap as any)?.warehouses ?? [];
    const wh = warehouses.find((w: any) => String(w?.warehouse_id ?? w?.warehouseId ?? "") === "default");
    const nextStock = Number(wh?.available_qty ?? wh?.availableQty ?? 0);
    const nextMinStock = Number(wh?.safety_stock ?? wh?.safetyStock ?? 0);
    const nextWarehouse = String(wh?.warehouse_id ?? wh?.warehouseId ?? "default");
    const nextLast = String(
      wh?.last_synced_at ??
        wh?.lastSyncedAt ??
        (snap as any)?.last_synced_at ??
        (snap as any)?.lastSyncedAt ??
        new Date().toISOString()
    );

    inventory.value = inventory.value.map((it) =>
      it.skuId === adjustTarget.value?.skuId
        ? { ...it, stock: nextStock, minStock: nextMinStock, warehouse: nextWarehouse, lastUpdate: nextLast }
        : it
    );
    closeAdjust();
  } catch (e: any) {
    adjustError.value = String(e?.message || e || "调整失败");
  } finally {
    adjustSubmitting.value = false;
  }
};

const filteredInventory = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  return inventory.value.filter((item) => {
    const passQ =
      !q ||
      item.productName.toLowerCase().includes(q) ||
      item.sku.toLowerCase().includes(q) ||
      item.productId.toLowerCase().includes(q);
    const status = getStockStatus(item.stock, item.minStock);
    const passStatus = selectedStatus.value === ALL_STATUS || status === selectedStatus.value;
    return passQ && passStatus;
  });
});

const getStockStatus = (stock: number, minStock: number) => {
  if (stock === 0) return "缺货";
  if (stock <= minStock) return "库存不足";
  return "正常";
};
const getStockColor = (stock: number, minStock: number) => {
  if (stock === 0) return "error";
  if (stock <= minStock) return "warning";
  return "success";
};
</script>
