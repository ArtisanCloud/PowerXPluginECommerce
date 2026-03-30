<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">会籍等级</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          查看并创建会籍等级配置
        </p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" @click="createOpen = true">
        创建等级
      </UButton>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">等级列表</h2>
          <UInput
            v-model="searchQuery"
            placeholder="搜索等级名称/编码"
            icon="i-heroicons-magnifying-glass"
            size="sm"
          />
        </div>
      </template>

      <UTable :columns="columns" :data="filteredTiers" :loading="loading">
        <template #status-cell="{ row }">
          <UBadge :color="row.status === 'active' ? 'success' : 'neutral'" variant="soft">
            {{ row.status || '-' }}
          </UBadge>
        </template>
        <template #createdAt-cell="{ row }">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ formatDate(row.createdAt) }}
          </span>
        </template>
      </UTable>

      <UAlert v-if="!filteredTiers.length && !loading" color="gray" class="mt-4">
        暂无会籍等级配置
      </UAlert>
    </UCard>

    <UModal v-model:open="createOpen">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">创建会籍等级</h3>
          </template>
          <form class="space-y-4" @submit.prevent="handleCreateTier">
            <UFormField label="等级名称" required>
              <UInput v-model="createForm.name" placeholder="例如：白金会员" />
            </UFormField>
            <UFormField label="等级编码" required>
              <UInput v-model="createForm.code" placeholder="例如：platinum" />
            </UFormField>
            <UFormField label="状态">
              <USelect
                v-model="createForm.status"
                :items="statusOptions"
                option-attribute="label"
                value-attribute="value"
              />
            </UFormField>
            <UFormField label="规则(JSON)">
              <UTextarea
                v-model="createForm.rulesText"
                :rows="5"
                class="font-mono text-xs"
                placeholder='{"upgrade":{"minSpend": 1000}}'
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
import type { MembershipTier } from "~/types/membership";
import type { TableColumn } from "@nuxt/ui";
import { useToast } from "#imports";

const api = useMembershipAdminApi();
const toast = useToast();
const loading = ref(false);
const saving = ref(false);
const searchQuery = ref("");
const tiers = ref<MembershipTier[]>([]);
const createOpen = ref(false);
const createForm = reactive({
  name: "",
  code: "",
  status: "draft",
  rulesText: "{}",
});

const statusOptions = [
  { label: "草稿", value: "draft" },
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

const columns = computed<TableColumn<MembershipTier>[]>(() => [
  { accessorKey: "name", header: "等级名称" },
  { accessorKey: "code", header: "等级编码" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "创建时间" },
]);

const filteredTiers = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  if (!keyword) return tiers.value;
  return tiers.value.filter((tier) => {
    return (
      String(tier.name || "").toLowerCase().includes(keyword) ||
      String(tier.code || "").toLowerCase().includes(keyword)
    );
  });
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

const loadTiers = async () => {
  try {
    loading.value = true;
    const resp = await api.listTiers();
    tiers.value = resp?.items ?? [];
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadTiers();
});

const handleCreateTier = async () => {
  const name = createForm.name.trim();
  const code = createForm.code.trim();
  if (!name || !code) {
    toast.add({ title: "名称和编码不能为空", color: "warning" });
    return;
  }

  let rules: Record<string, any> | null = null;
  const rawRules = createForm.rulesText.trim();
  if (rawRules) {
    try {
      rules = JSON.parse(rawRules);
    } catch (error: any) {
      toast.add({ title: "规则 JSON 解析失败", description: error?.message, color: "error" });
      return;
    }
  }

  try {
    saving.value = true;
    await api.createTier({
      name,
      code,
      status: createForm.status,
      rules,
    });
    toast.add({ title: "会籍等级已创建", color: "success" });
    createOpen.value = false;
    createForm.name = "";
    createForm.code = "";
    createForm.status = "draft";
    createForm.rulesText = "{}";
    await loadTiers();
  } catch (error: any) {
    toast.add({ title: "创建失败", description: error?.message || "请稍后重试", color: "error" });
  } finally {
    saving.value = false;
  }
};
</script>
