<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">权限与角色</h1>
        <p class="text-gray-500 dark:text-gray-400">
          管理后台角色、页面权限以及操作审计。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton variant="ghost" color="neutral" icon="i-heroicons-shield-check">
          权限对照
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">
          新建角色
        </UButton>
      </div>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">角色列表</h3>
            <UInput
              v-model="roleKeyword"
              class="w-48"
              placeholder="搜索角色"
              icon="i-heroicons-magnifying-glass"
            />
          </div>
        </template>

        <UTable :columns="roleColumns" :data="filteredRoles">
          <template #members-cell="{ getValue }">
            {{ getValue() }} 人
          </template>
          <template #actions-cell>
            <div class="flex gap-2">
              <UButton size="xs" variant="ghost">分配成员</UButton>
              <UButton size="xs" variant="ghost" color="neutral">复制</UButton>
            </div>
          </template>
        </UTable>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">权限矩阵</h3>
            <USelect v-model="selectedRoleId" :options="roleOptions" class="w-48" />
          </div>
        </template>

        <div class="space-y-4">
          <div
            v-for="module in permissionMatrix"
            :key="module.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <div>
                <p class="font-medium text-gray-900 dark:text-white">
                  {{ module.name }}
                </p>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  {{ module.desc }}
                </p>
              </div>
              <UToggle v-model="module.enabled" />
            </div>
            <div class="mt-3 grid gap-2 md:grid-cols-2">
              <label
                v-for="action in module.actions"
                :key="action.label"
                class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300"
              >
                <UCheckbox v-model="action.allowed" />
                {{ action.label }}
              </label>
            </div>
          </div>
        </div>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">操作审计</h3>
          <UButton size="sm" variant="ghost">导出</UButton>
        </div>
      </template>

      <UTable :columns="auditColumns" :data="auditLogs">
        <template #time-cell="{ getValue }">
          {{ getValue() }}
        </template>
        <template #result-cell="{ getValue }">
          <UBadge :color="getValue() === '成功' ? 'success' : 'error'" variant="subtle">
            {{ getValue() }}
          </UBadge>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "settings-roles",
});

type Role = {
  id: string;
  name: string;
  desc: string;
  members: number;
};

const roles = ref<Role[]>([
  { id: "ops", name: "运营管理员", desc: "商品、营销、订单管理", members: 8 },
  { id: "finance", name: "财务结算", desc: "支付与对账权限", members: 3 },
  { id: "viewer", name: "只读访客", desc: "仅可查看报表", members: 12 },
]);

const roleKeyword = ref("");
const filteredRoles = computed(() =>
  roles.value.filter((role) =>
    role.name.toLowerCase().includes(roleKeyword.value.trim().toLowerCase()),
  ),
);

const roleColumns = computed<TableColumn<Role>[]>(() => [
  { accessorKey: "name", header: "角色" },
  { accessorKey: "desc", header: "说明" },
  { accessorKey: "members", header: "成员数" },
  { id: "actions", header: "操作" },
]);

const roleOptions = computed(() =>
  roles.value.map((role) => ({ label: role.name, value: role.id })),
);
const selectedRoleId = ref(roleOptions.value[0]?.value ?? "");

type PermissionModule = {
  id: string;
  name: string;
  desc: string;
  enabled: boolean;
  actions: { label: string; allowed: boolean }[];
};

const permissionMatrix = ref<PermissionModule[]>([
  {
    id: "product",
    name: "商品管理",
    desc: "SPU/SKU、库存、定价",
    enabled: true,
    actions: [
      { label: "查看", allowed: true },
      { label: "编辑", allowed: true },
      { label: "发布", allowed: false },
    ],
  },
  {
    id: "pricing",
    name: "定价中心",
    desc: "价目表与合同价",
    enabled: true,
    actions: [
      { label: "查看", allowed: true },
      { label: "编辑", allowed: false },
      { label: "审批", allowed: false },
    ],
  },
  {
    id: "finance",
    name: "财务与结算",
    desc: "支付渠道、对账、税务",
    enabled: false,
    actions: [
      { label: "查看", allowed: false },
      { label: "编辑", allowed: false },
    ],
  },
]);

type AuditLog = {
  user: string;
  role: string;
  action: string;
  time: string;
  result: "成功" | "失败";
};

const auditLogs = ref<AuditLog[]>([
  {
    user: "王小明",
    role: "运营管理员",
    action: "修改商品价格",
    time: "2024-02-12 14:20",
    result: "成功",
  },
  {
    user: "李静",
    role: "财务结算",
    action: "导出结算单",
    time: "2024-02-12 10:05",
    result: "成功",
  },
  {
    user: "陈晨",
    role: "运营管理员",
    action: "删除营销活动",
    time: "2024-02-11 18:42",
    result: "失败",
  },
]);

const auditColumns = computed<TableColumn<AuditLog>[]>(() => [
  { accessorKey: "user", header: "用户" },
  { accessorKey: "role", header: "角色" },
  { accessorKey: "action", header: "操作" },
  { accessorKey: "time", header: "时间" },
  { accessorKey: "result", header: "结果" },
]);
</script>
