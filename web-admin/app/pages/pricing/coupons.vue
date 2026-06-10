<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">优惠券模板与发券</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          管理优惠券模板，并对指定用户批量发券。
        </p>
      </div>
      <UButton color="neutral" variant="outline" icon="i-heroicons-arrow-path" :loading="loading" @click="refresh">
        刷新
      </UButton>
    </div>

    <UCard>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UFormField label="关键词">
          <UInput v-model="filters.keyword" placeholder="模板编码/名称" @keydown.enter="refresh" />
        </UFormField>
        <UFormField label="状态">
          <USelect v-model="filters.status" :items="statusItems" />
        </UFormField>
        <UFormField label="页码">
          <UInput v-model.number="filters.page" type="number" min="1" />
        </UFormField>
        <UFormField label="每页">
          <UInput v-model.number="filters.pageSize" type="number" min="1" />
        </UFormField>
      </div>
      <div class="mt-4 flex gap-2">
        <UButton color="primary" icon="i-heroicons-magnifying-glass" @click="refresh">查询</UButton>
        <UButton color="primary" variant="outline" icon="i-heroicons-plus" @click="openCreate">新建模板</UButton>
      </div>
    </UCard>

    <UCard>
      <UTable :data="templates" :columns="columns" :loading="loading">
        <template #status-cell="{ row }">
          <UBadge :color="row.original.status === 'active' ? 'success' : 'neutral'" variant="subtle">
            {{ statusLabel(row.original.status) }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="sm" color="primary" variant="ghost" icon="i-heroicons-pencil-square" @click="openEdit(row.original)">
              编辑
            </UButton>
            <UButton size="sm" color="neutral" variant="ghost" icon="i-heroicons-ticket" @click="openIssue(row.original)">
              发券
            </UButton>
          </div>
        </template>
      </UTable>
      <div class="mt-4 text-sm text-gray-500">总数：{{ total }}</div>
    </UCard>

    <UModal
      v-model:open="editOpen"
      :title="editing?.id ? '编辑模板' : '新建模板'"
      :close="true"
      :ui="{ content: 'w-full sm:max-w-6xl' }"
    >
      <template #body>
        <div class="max-h-[70vh] overflow-y-auto px-10 py-6">
          <div class="space-y-6">
            <section class="space-y-3">
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-200">基础信息</h3>
              <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-x-6 gap-y-4">
                <UFormField label="模板编码" class="md:col-span-2 xl:col-span-2">
                  <div class="flex gap-2">
                    <UInput v-model="form.code" class="min-w-0 flex-1" :disabled="!!editing?.id" />
                    <UButton
                      v-if="!editing?.id"
                      color="neutral"
                      variant="outline"
                      icon="i-heroicons-sparkles"
                      @click="generateTemplateCode"
                    >
                      随机生成
                    </UButton>
                  </div>
                </UFormField>
                <UFormField label="模板名称">
                  <UInput v-model="form.name" class="w-full" />
                </UFormField>
                <UFormField label="状态">
                  <USelect v-model="form.status" :items="templateStatusItems" class="w-full" />
                </UFormField>
              </div>
            </section>

            <USeparator />

            <section class="space-y-3">
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-200">优惠规则</h3>
              <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-x-6 gap-y-4">
                <UFormField label="类型">
                  <USelect v-model="form.coupon_type" :items="couponTypeItems" class="w-full" />
                </UFormField>
                <UFormField label="最低消费金额">
                  <UInput
                    v-model.number="form.min_amount_yuan"
                    type="number"
                    min="0"
                    step="0.01"
                    class="w-full"
                    placeholder="0 表示无门槛"
                  />
                </UFormField>
                <UFormField v-if="form.coupon_type === 'amount'" label="优惠金额">
                  <UInput v-model.number="form.amount_yuan" type="number" min="0" step="0.01" class="w-full" />
                </UFormField>
                <UFormField v-else label="折扣比例">
                  <UInput
                    v-model.number="form.percent"
                    type="number"
                    min="0.01"
                    max="100"
                    step="0.01"
                    class="w-full"
                    placeholder="例如 85 表示 85 折"
                  />
                </UFormField>
                <UFormField label="优惠层级">
                  <USelect v-model="form.level" :items="levelItems" class="w-full" />
                </UFormField>
              </div>
            </section>

            <section class="space-y-3">
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-200">有效期与适用范围</h3>
              <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-x-6 gap-y-4">
                <UFormField label="生效开始">
                  <UInput v-model="form.valid_from" type="datetime-local" step="1" class="w-full" />
                </UFormField>
                <UFormField label="生效结束">
                  <UInput v-model="form.valid_to" type="datetime-local" step="1" class="w-full" />
                </UFormField>
                <UFormField label="适用 SKU" class="md:col-span-2 xl:col-span-2">
                  <USelectMenu
                    v-model="form.sku_ids"
                    :items="skuItems"
                    value-key="value"
                    label-key="label"
                    multiple
                    searchable
                    :loading="skuLoading"
                    :portal="false"
                    :ui="{ content: 'z-[60] max-h-72 overflow-auto' }"
                    class="w-full"
                    placeholder="搜索并选择 SKU，留空表示全场"
                    @update:search-term="handleSkuSearch"
                  >
                    <template #default>
                      <div v-if="selectedSkuItems.length" class="flex min-h-6 flex-1 flex-wrap items-center gap-1.5 py-0.5">
                        <UBadge
                          v-for="item in selectedSkuItems"
                          :key="item.value"
                          color="neutral"
                          variant="subtle"
                          class="max-w-full"
                        >
                          <span class="max-w-56 truncate">{{ item.label }}</span>
                        </UBadge>
                      </div>
                      <span v-else class="text-dimmed">搜索并选择 SKU，留空表示全场</span>
                    </template>
                  </USelectMenu>
                </UFormField>
              </div>
            </section>

            <USeparator />

            <section class="space-y-3">
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-200">叠加与退款</h3>
              <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-x-6 gap-y-4">
                <UFormField label="优先级">
                  <UInput v-model.number="form.priority" type="number" min="1" step="1" class="w-full" />
                </UFormField>
                <UFormField label="互斥组">
                  <UInput v-model="form.exclusion_group" class="w-full" placeholder="同组只允许使用一张，可留空" />
                </UFormField>
                <UFormField label="退款策略">
                  <USelect v-model="form.refund_policy" :items="refundPolicyItems" class="w-full" />
                </UFormField>
                <UFormField label="开关">
                  <div class="flex h-10 items-center gap-8 rounded-md border border-gray-200 px-3 dark:border-gray-800">
                    <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                      <USwitch v-model="form.stackable" />
                      <span>允许叠加</span>
                    </label>
                    <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                      <USwitch v-model="form.refund_return_coupon" />
                      <span>退款返券</span>
                    </label>
                  </div>
                </UFormField>
              </div>
            </section>
          </div>
        </div>
      </template>
      <template #footer>
        <div class="p-4 flex w-full justify-end gap-2">
          <UButton color="neutral" variant="outline" @click="editOpen = false">取消</UButton>
          <UButton color="primary" :loading="saving" @click="saveTemplate">保存</UButton>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="issueOpen" title="批量发券" :close="true">
      <template #body>
        <div class="p-4 space-y-3">
          <UFormField label="模板">
            <UInput :model-value="issueTemplateLabel" disabled />
          </UFormField>
          <UFormField label="发券用户">
            <USelectMenu
              v-model="issueForm.user_ids"
              :items="customerItems"
              value-key="value"
              label-key="label"
              multiple
              searchable
              :loading="customerLoading"
              :portal="false"
              :ui="{ content: 'z-[60] max-h-72 overflow-auto' }"
              class="w-full"
              placeholder="搜索并选择用户"
              @update:search-term="handleCustomerSearch"
            >
              <template #default>
                <div v-if="selectedCustomerItems.length" class="flex min-h-6 flex-1 flex-wrap items-center gap-1.5 py-0.5">
                  <UBadge
                    v-for="item in selectedCustomerItems"
                    :key="item.value"
                    color="neutral"
                    variant="subtle"
                    class="max-w-full"
                  >
                    <span class="max-w-56 truncate">{{ item.label }}</span>
                  </UBadge>
                </div>
                <span v-else class="text-dimmed">搜索并选择用户</span>
              </template>
            </USelectMenu>
          </UFormField>
          <UFormField label="每个用户发放数量">
            <UInput v-model.number="issueForm.quantity_per_user" type="number" min="1" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="p-4 flex w-full justify-end gap-2">
          <UButton color="neutral" variant="outline" @click="issueOpen = false">取消</UButton>
          <UButton color="primary" :loading="issuing" @click="submitIssue">确认发券</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useCouponsApi, type CouponTemplate } from "~/composables/api/useCoupons";
import { useCustomerApi } from "~/composables/api/useCustomer";
import { useSkuApi } from "~/composables/api/useSku";

const api = useCouponsApi();
const customerApi = useCustomerApi();
const skuApi = useSkuApi();
const toast = useToast();
const loading = ref(false);
const saving = ref(false);
const issuing = ref(false);
const customerLoading = ref(false);
const skuLoading = ref(false);
const templates = ref<CouponTemplate[]>([]);
const total = ref(0);
const customerItems = ref<{ label: string; value: string }[]>([]);
const skuItems = ref<{ label: string; value: string }[]>([]);

const filters = reactive({
  keyword: "",
  status: "all",
  page: 1,
  pageSize: 20,
});

const statusItems = [
  { label: "全部", value: "all" },
  { label: "草稿", value: "draft" },
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
  { label: "禁用", value: "disabled" },
  { label: "归档", value: "archived" },
];

const templateStatusItems = statusItems.filter((item) => item.value !== "all");

const statusLabel = (status: string) => {
  return templateStatusItems.find((item) => item.value === status)?.label || status;
};

const couponTypeItems = [
  { label: "固定金额券", value: "amount" },
  { label: "折扣券", value: "percent" },
];

const levelItems = [
  { label: "订单级", value: "order" },
  { label: "商品级", value: "item" },
];

const refundPolicyItems = [
  { label: "按订单退款策略处理", value: "default" },
  { label: "退款返券", value: "return_coupon" },
  { label: "退款不返券", value: "no_return" },
];

const columns: TableColumn<CouponTemplate>[] = [
  { accessorKey: "code", header: "编码" },
  { accessorKey: "name", header: "名称" },
  { accessorKey: "coupon_type", header: "类型" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "valid_from", header: "开始时间" },
  { accessorKey: "valid_to", header: "结束时间" },
  { id: "actions", header: "操作" },
];

const editOpen = ref(false);
const editing = ref<CouponTemplate | null>(null);
const form = reactive({
  code: "",
  name: "",
  coupon_type: "amount",
  status: "draft",
  valid_from: "",
  valid_to: "",
  min_amount_yuan: 0,
  amount_yuan: 0,
  percent: 100,
  level: "order",
  sku_ids: [] as string[],
  priority: 100,
  stackable: true,
  exclusion_group: "",
  refund_return_coupon: false,
  refund_policy: "default",
});

const issueOpen = ref(false);
const issueTemplate = ref<CouponTemplate | null>(null);
const issueForm = reactive({
  user_ids: [] as string[],
  quantity_per_user: 1,
});

const issueTemplateLabel = computed(() => {
  if (!issueTemplate.value) return "";
  return `${issueTemplate.value.code} / ${issueTemplate.value.name}`;
});

const selectedSkuItems = computed(() => form.sku_ids.map((value) => {
  return skuItems.value.find((item) => item.value === value) || { label: value, value };
}));

const selectedCustomerItems = computed(() => issueForm.user_ids.map((value) => {
  return customerItems.value.find((item) => item.value === value) || { label: value, value };
}));

const refresh = async () => {
  loading.value = true;
  try {
    const res = await api.listTemplates({
      keyword: filters.keyword || undefined,
      status: filters.status === "all" ? undefined : filters.status,
      page: filters.page,
      pageSize: filters.pageSize,
    });
    templates.value = res.items || [];
    total.value = Number(res.total || 0);
  } catch (error: any) {
    toast.add({ color: "error", title: "查询失败", description: error?.message || "请求失败" });
  } finally {
    loading.value = false;
  }
};

const openCreate = () => {
  editing.value = null;
  form.code = "";
  form.name = "";
  form.coupon_type = "amount";
  form.status = "draft";
  form.valid_from = "";
  form.valid_to = "";
  resetRuleForm();
  editOpen.value = true;
};

const generateTemplateCode = () => {
  const typePrefix = form.coupon_type === "percent" ? "PCT" : "AMT";
  const date = new Date();
  const ymd = [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, "0"),
    String(date.getDate()).padStart(2, "0"),
  ].join("");
  const random = Math.random().toString(36).slice(2, 8).toUpperCase();
  form.code = `CPN-${typePrefix}-${ymd}-${random}`;
};

const centsToYuan = (value: unknown) => {
  const cents = Number(value || 0);
  if (!Number.isFinite(cents) || cents <= 0) return 0;
  return Math.round(cents) / 100;
};

const yuanToCents = (value: unknown) => {
  const yuan = Number(value || 0);
  if (!Number.isFinite(yuan) || yuan <= 0) return 0;
  return Math.round(yuan * 100);
};

const normalizePercent = (value: unknown) => {
  const percent = Number(value || 0);
  if (!Number.isFinite(percent) || percent <= 0) return 0;
  return Math.min(Math.round(percent * 100), 10000);
};

const bpsToPercent = (value: unknown) => {
  const bps = Number(value || 0);
  if (!Number.isFinite(bps) || bps <= 0) return 100;
  return Math.round((bps / 100) * 100) / 100;
};

const resetRuleForm = () => {
  form.min_amount_yuan = 0;
  form.amount_yuan = 0;
  form.percent = 100;
  form.level = "order";
  form.sku_ids = [];
  form.priority = 100;
  form.stackable = true;
  form.exclusion_group = "";
  form.refund_return_coupon = false;
  form.refund_policy = "default";
};

let skuSearchTimer: ReturnType<typeof setTimeout> | null = null;
let customerSearchTimer: ReturnType<typeof setTimeout> | null = null;

const mergeSkuItems = (items: { label: string; value: string }[]) => {
  const byValue = new Map<string, { label: string; value: string }>();
  for (const item of [...skuItems.value, ...items]) {
    if (item.value) {
      byValue.set(item.value, item);
    }
  }
  skuItems.value = Array.from(byValue.values());
};

const ensureSelectedSkuItems = (values: string[]) => {
  mergeSkuItems(values.map((value) => ({ label: value, value })));
};

const loadSkuItems = async (keyword = "") => {
  skuLoading.value = true;
  try {
    const res = await skuApi.list({
      keyword: keyword || undefined,
      page: 1,
      pageSize: 30,
    });
    mergeSkuItems((res.items || [])
      .map((sku) => {
        const value = sku.skuCode || sku.id;
        const suffix = sku.spuName || sku.specDisplay;
        return {
          value,
          label: suffix ? `${value} / ${suffix}` : value,
        };
      })
      .filter((item) => item.value));
  } catch (error: any) {
    toast.add({ color: "error", title: "SKU 查询失败", description: error?.message || "请求失败" });
  } finally {
    skuLoading.value = false;
  }
};

const handleSkuSearch = (term: string) => {
  if (skuSearchTimer) {
    clearTimeout(skuSearchTimer);
  }
  skuSearchTimer = setTimeout(() => {
    loadSkuItems(term.trim());
  }, 250);
};

const mergeCustomerItems = (items: { label: string; value: string }[]) => {
  const byValue = new Map<string, { label: string; value: string }>();
  for (const item of [...customerItems.value, ...items]) {
    if (item.value) {
      byValue.set(item.value, item);
    }
  }
  customerItems.value = Array.from(byValue.values());
};

const loadCustomerItems = async (keyword = "") => {
  customerLoading.value = true;
  try {
    const res = await customerApi.listCustomers({
      keyword: keyword || undefined,
      page: 1,
      pageSize: 30,
    });
    mergeCustomerItems((res.data || [])
      .map((customer) => {
        const suffix = [customer.phone, customer.email].filter(Boolean).join(" / ");
        return {
          value: customer.id,
          label: suffix ? `${customer.name || customer.id} / ${suffix}` : `${customer.name || customer.id}`,
        };
      })
      .filter((item) => item.value));
  } catch (error: any) {
    toast.add({ color: "error", title: "用户查询失败", description: error?.message || "请求失败" });
  } finally {
    customerLoading.value = false;
  }
};

const handleCustomerSearch = (term: string) => {
  if (customerSearchTimer) {
    clearTimeout(customerSearchTimer);
  }
  customerSearchTimer = setTimeout(() => {
    loadCustomerItems(term.trim());
  }, 250);
};

const toDateTimeLocal = (value?: string | null) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return [
    date.getFullYear(),
    pad(date.getMonth() + 1),
    pad(date.getDate()),
  ].join("-") + `T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
};

const fromDateTimeLocal = (value: string) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toISOString();
};

const openEdit = (row: CouponTemplate) => {
  editing.value = row;
  form.code = row.code;
  form.name = row.name;
  form.coupon_type = row.coupon_type;
  form.status = row.status;
  form.valid_from = toDateTimeLocal(row.valid_from);
  form.valid_to = toDateTimeLocal(row.valid_to);
  const thresholdRule = row.threshold_rule || {};
  const scopeRule = row.scope_rule || {};
  const stackingRule = row.stacking_rule || {};
  const refundRule = row.refund_rule || {};
  form.min_amount_yuan = centsToYuan(thresholdRule.min_amount_minor);
  form.amount_yuan = centsToYuan(thresholdRule.amount_minor ?? scopeRule.amount_minor);
  form.percent = bpsToPercent(thresholdRule.percent_bps);
  form.level = String(scopeRule.level || stackingRule.level || "order");
  form.sku_ids = Array.isArray(scopeRule.sku_ids) ? scopeRule.sku_ids.map((item: any) => String(item)).filter(Boolean) : [];
  ensureSelectedSkuItems(form.sku_ids);
  form.priority = Number(stackingRule.priority || 100);
  form.stackable = typeof stackingRule.stackable === "boolean" ? stackingRule.stackable : true;
  form.exclusion_group = String(stackingRule.exclusion_group || "");
  form.refund_return_coupon = Boolean(refundRule.return_coupon);
  form.refund_policy = String(refundRule.policy || (refundRule.return_coupon ? "return_coupon" : "default"));
  editOpen.value = true;
};

const buildRulePayload = () => {
  const thresholdRule: Record<string, any> = {};
  const minAmount = yuanToCents(form.min_amount_yuan);
  if (minAmount > 0) {
    thresholdRule.min_amount_minor = minAmount;
  }
  if (form.coupon_type === "amount") {
    thresholdRule.amount_minor = yuanToCents(form.amount_yuan);
  } else {
    thresholdRule.percent_bps = normalizePercent(form.percent);
  }

  const skuIDs = form.sku_ids.map((item) => String(item).trim()).filter(Boolean);
  const scopeRule: Record<string, any> = { level: form.level };
  if (skuIDs.length > 0) {
    scopeRule.sku_ids = skuIDs;
  }

  const stackingRule: Record<string, any> = {
    level: form.level,
    priority: Number(form.priority || 100),
    stackable: Boolean(form.stackable),
  };
  if (form.exclusion_group.trim()) {
    stackingRule.exclusion_group = form.exclusion_group.trim();
  }

  const refundRule: Record<string, any> = {
    policy: form.refund_policy,
    return_coupon: Boolean(form.refund_return_coupon || form.refund_policy === "return_coupon"),
  };

  return { thresholdRule, scopeRule, stackingRule, refundRule };
};

const saveTemplate = async () => {
  saving.value = true;
  try {
    const { thresholdRule, scopeRule, stackingRule, refundRule } = buildRulePayload();
    const payload = {
      code: form.code,
      name: form.name,
      coupon_type: form.coupon_type,
      status: form.status,
      valid_from: fromDateTimeLocal(form.valid_from),
      valid_to: fromDateTimeLocal(form.valid_to),
      threshold_rule: thresholdRule,
      scope_rule: scopeRule,
      stacking_rule: stackingRule,
      refund_rule: refundRule,
    };
    if (editing.value?.id) {
      await api.updateTemplate(editing.value.id, payload);
    } else {
      await api.createTemplate(payload);
    }
    editOpen.value = false;
    toast.add({ color: "success", title: "保存成功" });
    await refresh();
  } catch (error: any) {
    toast.add({ color: "error", title: "保存失败", description: error?.message || "请求失败" });
  } finally {
    saving.value = false;
  }
};

const openIssue = (row: CouponTemplate) => {
  issueTemplate.value = row;
  issueForm.user_ids = [];
  issueForm.quantity_per_user = 1;
  loadCustomerItems();
  issueOpen.value = true;
};

const submitIssue = async () => {
  if (!issueTemplate.value) return;
  if (issueForm.user_ids.length === 0) {
    toast.add({ color: "warning", title: "请选择发券用户" });
    return;
  }
  issuing.value = true;
  try {
    const res = await api.issueCoupons({
      template_id: issueTemplate.value.id,
      user_ids: issueForm.user_ids,
      quantity_per_user: issueForm.quantity_per_user || 1,
    });
    const templateID = issueTemplate.value.id;
    issueOpen.value = false;
    toast.add({ color: "success", title: "发券成功", description: `已发放 ${res.issued} 张` });
    await navigateTo({
      path: "/pricing/coupon-usages",
      query: {
        templateId: templateID,
        status: "available",
      },
    });
  } catch (error: any) {
    toast.add({ color: "error", title: "发券失败", description: error?.message || "请求失败" });
  } finally {
    issuing.value = false;
  }
};

onMounted(() => {
  refresh();
  loadSkuItems();
});
</script>
