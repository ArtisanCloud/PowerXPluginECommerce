<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">会籍权益</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          查看并创建会籍权益配置
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" @click="createOpen = true">
        创建权益
      </UButton>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">权益列表</h2>
          <UInput
            v-model="searchQuery"
            placeholder="搜索权益名称"
            icon="i-heroicons-magnifying-glass"
            size="sm"
          />
        </div>
      </template>

      <UTable :columns="columns" :data="filteredBenefits" :loading="loading">
        <template #status-cell="{ row }">
          <UBadge :color="row.status === 'active' ? 'success' : 'neutral'" variant="soft">
            {{ row.status || '-' }}
          </UBadge>
        </template>
        <template #items-cell="{ row }">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ formatItems(row.items) }}
          </span>
        </template>
        <template #createdAt-cell="{ row }">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ formatDate(row.createdAt) }}
          </span>
        </template>
      </UTable>

      <UAlert v-if="!filteredBenefits.length && !loading" color="gray" class="mt-4">
        暂无权益配置
      </UAlert>
    </UCard>

    <UModal v-model:open="createOpen">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">创建会籍权益</h3>
          </template>
          <form class="space-y-4" @submit.prevent="handleCreateBenefit">
            <UFormField label="权益名称" required>
              <UInput v-model="createForm.name" placeholder="例如：免邮权益" />
            </UFormField>
            <UFormField label="类型">
              <USelect
                v-model="createForm.type"
                :items="typeOptions"
                option-attribute="label"
                value-attribute="value"
              />
            </UFormField>
            <UFormField label="状态">
              <USelect
                v-model="createForm.status"
                :items="statusOptions"
                option-attribute="label"
                value-attribute="value"
              />
            </UFormField>
            <UFormField label="内容(JSON)">
              <UTextarea
                v-model="createForm.itemsText"
                :rows="6"
                class="font-mono text-xs"
                placeholder='[{"serviceCode":"shipping.free","quantity":-1}]'
              />
            </UFormField>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" type="button" :disabled="saving" @click="createOpen = false">
                取消
              </UButton>
              <UButton color="primary" type="submit" :loading="saving">
                保存
              </UButton>
            </div>
          </form>
        </UCard>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useMembershipAdminApi } from "~/composables/api/useMembership";
import type { MembershipBenefit } from "~/types/membership";
import type { TableColumn } from "@nuxt/ui";
import { useToast } from "#imports";

const api = useMembershipAdminApi();
const toast = useToast();
const loading = ref(false);
const saving = ref(false);
const searchQuery = ref("");
const benefits = ref<MembershipBenefit[]>([]);
const createOpen = ref(false);
const createForm = reactive({
  name: "",
  type: "single",
  status: "active",
  itemsText: "[]",
});

const typeOptions = [
  { label: "单体权益", value: "single" },
  { label: "组合权益", value: "bundle" },
];

const statusOptions = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

const columns = computed<TableColumn<MembershipBenefit>[]>(() => [
  { accessorKey: "name", header: "权益名称" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "items", header: "内容" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "创建时间" },
]);

const filteredBenefits = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  if (!keyword) return benefits.value;
  return benefits.value.filter((benefit) =>
    String(benefit.name || "").toLowerCase().includes(keyword),
  );
});

const formatDate = (value?: string) => {
  if (!value) return "-";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "short",
    day: "numeric",
  }).format(parsed);
};

const formatItems = (items?: any) => {
  if (!items) return "-";
  if (Array.isArray(items)) return `${items.length} 项`;
  return "已配置";
};

const loadBenefits = async () => {
  try {
    loading.value = true;
    const resp = await api.listBenefits();
    benefits.value = resp?.items ?? [];
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadBenefits();
});

const handleCreateBenefit = async () => {
  const name = createForm.name.trim();
  if (!name) {
    toast.add({ title: "权益名称不能为空", color: "warning" });
    return;
  }

  let items: any = [];
  const rawItems = createForm.itemsText.trim();
  if (rawItems) {
    try {
      items = JSON.parse(rawItems);
    } catch (error: any) {
      toast.add({ title: "内容 JSON 解析失败", description: error?.message, color: "error" });
      return;
    }
  }

  try {
    saving.value = true;
    await api.createBenefit({
      name,
      type: createForm.type,
      status: createForm.status,
      items,
    });
    toast.add({ title: "会籍权益已创建", color: "success" });
    createOpen.value = false;
    createForm.name = "";
    createForm.type = "single";
    createForm.status = "active";
    createForm.itemsText = "[]";
    await loadBenefits();
  } catch (error: any) {
    toast.add({ title: "创建失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    saving.value = false;
  }
};
</script>
