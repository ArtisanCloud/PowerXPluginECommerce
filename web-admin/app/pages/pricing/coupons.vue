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
            {{ row.original.status }}
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

    <UModal v-model:open="editOpen" :title="editing?.id ? '编辑模板' : '新建模板'" :close="true">
      <template #body>
        <div class="p-4 space-y-3">
          <UFormField label="模板编码"><UInput v-model="form.code" :disabled="!!editing?.id" /></UFormField>
          <UFormField label="模板名称"><UInput v-model="form.name" /></UFormField>
          <UFormField label="类型"><UInput v-model="form.coupon_type" placeholder="fixed/discount" /></UFormField>
          <UFormField label="状态"><USelect v-model="form.status" :items="statusItems" /></UFormField>
          <UFormField label="生效开始"><UInput v-model="form.valid_from" placeholder="2026-04-03T00:00:00Z" /></UFormField>
          <UFormField label="生效结束"><UInput v-model="form.valid_to" placeholder="2026-04-30T23:59:59Z" /></UFormField>
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
          <UFormField label="用户ID列表（逗号分隔）">
            <UInput v-model="issueForm.user_ids_raw" placeholder="u-1,u-2" />
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

const api = useCouponsApi();
const toast = useToast();
const loading = ref(false);
const saving = ref(false);
const issuing = ref(false);
const templates = ref<CouponTemplate[]>([]);
const total = ref(0);

const filters = reactive({
  keyword: "",
  status: "",
  page: 1,
  pageSize: 20,
});

const statusItems = [
  { label: "全部", value: "" },
  { label: "draft", value: "draft" },
  { label: "active", value: "active" },
  { label: "inactive", value: "inactive" },
  { label: "disabled", value: "disabled" },
  { label: "archived", value: "archived" },
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
  coupon_type: "fixed",
  status: "draft",
  valid_from: "",
  valid_to: "",
});

const issueOpen = ref(false);
const issueTemplate = ref<CouponTemplate | null>(null);
const issueForm = reactive({
  user_ids_raw: "",
  quantity_per_user: 1,
});

const issueTemplateLabel = computed(() => {
  if (!issueTemplate.value) return "";
  return `${issueTemplate.value.code} / ${issueTemplate.value.name}`;
});

const refresh = async () => {
  loading.value = true;
  try {
    const res = await api.listTemplates({
      keyword: filters.keyword || undefined,
      status: filters.status || undefined,
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
  form.coupon_type = "fixed";
  form.status = "draft";
  form.valid_from = "";
  form.valid_to = "";
  editOpen.value = true;
};

const openEdit = (row: CouponTemplate) => {
  editing.value = row;
  form.code = row.code;
  form.name = row.name;
  form.coupon_type = row.coupon_type;
  form.status = row.status;
  form.valid_from = row.valid_from;
  form.valid_to = row.valid_to;
  editOpen.value = true;
};

const saveTemplate = async () => {
  saving.value = true;
  try {
    const payload = {
      code: form.code,
      name: form.name,
      coupon_type: form.coupon_type,
      status: form.status,
      valid_from: form.valid_from,
      valid_to: form.valid_to,
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
  issueForm.user_ids_raw = "";
  issueForm.quantity_per_user = 1;
  issueOpen.value = true;
};

const submitIssue = async () => {
  if (!issueTemplate.value) return;
  issuing.value = true;
  try {
    const userIDs = issueForm.user_ids_raw
      .split(",")
      .map((v) => v.trim())
      .filter(Boolean);
    const res = await api.issueCoupons({
      template_id: issueTemplate.value.id,
      user_ids: userIDs,
      quantity_per_user: issueForm.quantity_per_user || 1,
    });
    issueOpen.value = false;
    toast.add({ color: "success", title: "发券成功", description: `已发放 ${res.issued} 张` });
  } catch (error: any) {
    toast.add({ color: "error", title: "发券失败", description: error?.message || "请求失败" });
  } finally {
    issuing.value = false;
  }
};

onMounted(refresh);
</script>
