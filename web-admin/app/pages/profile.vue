<template>
  <div class="space-y-5">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-xl font-semibold text-gray-950 dark:text-white">个人设置</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">当前登录账号与租户成员身份</p>
      </div>
      <UButton icon="i-heroicons-arrow-path" variant="soft" color="neutral" :loading="pending" @click="loadContext">
        刷新
      </UButton>
    </div>

    <UAlert
      v-if="errorMessage"
      color="error"
      variant="soft"
      icon="i-heroicons-exclamation-triangle"
      title="身份上下文加载失败"
      :description="errorMessage"
    />

    <div class="grid gap-4 xl:grid-cols-3">
      <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-800 dark:bg-gray-950">
        <div class="mb-4 flex items-center gap-2">
          <UIcon name="i-heroicons-user-circle" class="size-5 text-primary-500" />
          <h2 class="text-sm font-semibold text-gray-900 dark:text-white">User</h2>
        </div>
        <dl class="space-y-3 text-sm">
          <InfoRow label="User UUID" :value="userUuid" mono />
          <InfoRow label="User ID" :value="displayValue(user?.id)" mono />
          <InfoRow label="用户名" :value="user?.username" />
          <InfoRow label="显示名" :value="user?.display_name || user?.displayName" />
          <InfoRow label="邮箱" :value="user?.email" />
          <InfoRow label="手机号" :value="user?.phone" />
        </dl>
      </section>

      <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-800 dark:bg-gray-950">
        <div class="mb-4 flex items-center gap-2">
          <UIcon name="i-heroicons-identification" class="size-5 text-primary-500" />
          <h2 class="text-sm font-semibold text-gray-900 dark:text-white">Member</h2>
        </div>
        <dl class="space-y-3 text-sm">
          <InfoRow label="Member UUID" :value="memberUuid" mono />
          <InfoRow label="Member ID" :value="displayValue(currentMember?.member_id || context?.current_member_id)" mono />
          <InfoRow label="租户角色" :value="roleText" />
          <InfoRow label="成员租户" :value="currentMember?.tenant_name" />
          <InfoRow label="管理员" :value="currentMember?.is_admin ? '是' : '否'" />
        </dl>
      </section>

      <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-800 dark:bg-gray-950">
        <div class="mb-4 flex items-center gap-2">
          <UIcon name="i-heroicons-building-office-2" class="size-5 text-primary-500" />
          <h2 class="text-sm font-semibold text-gray-900 dark:text-white">Tenant</h2>
        </div>
        <dl class="space-y-3 text-sm">
          <InfoRow label="Tenant UUID" :value="tenantUuid" mono />
          <InfoRow label="Tenant Key" :value="tenant?.key" mono />
          <InfoRow label="租户名称" :value="tenant?.name || currentMember?.tenant_name" />
          <InfoRow label="Root" :value="context?.is_root ? '是' : '否'" />
        </dl>
      </section>
    </div>

    <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-800 dark:bg-gray-950">
      <div class="mb-4 flex items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <UIcon name="i-heroicons-users" class="size-5 text-primary-500" />
          <h2 class="text-sm font-semibold text-gray-900 dark:text-white">租户成员列表</h2>
        </div>
        <UBadge color="neutral" variant="soft">{{ members.length }}</UBadge>
      </div>
      <UTable :data="members" :columns="memberColumns" :loading="pending">
        <template #member_uuid-cell="{ row }">
          <span class="font-mono text-xs text-gray-700 dark:text-gray-300">{{ row.original.member_uuid || "-" }}</span>
        </template>
        <template #tenant_uuid-cell="{ row }">
          <span class="font-mono text-xs text-gray-700 dark:text-gray-300">{{ row.original.tenant_uuid || "-" }}</span>
        </template>
        <template #is_admin-cell="{ row }">
          <UBadge :color="row.original.is_admin ? 'success' : 'neutral'" variant="soft">
            {{ row.original.is_admin ? "是" : "否" }}
          </UBadge>
        </template>
      </UTable>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from "vue";
import { useAuthService, type AuthMeContextResponse } from "~/composables/api/services/authService";
import { useUserStore } from "~/stores/user";

definePageMeta({
  name: "profile",
});

const InfoRow = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], default: "" },
    mono: { type: Boolean, default: false },
  },
  setup(props) {
    return () =>
      h("div", { class: "grid grid-cols-[110px_minmax(0,1fr)] gap-3" }, [
        h("dt", { class: "text-gray-500 dark:text-gray-400" }, props.label),
        h(
          "dd",
          {
            class: [
              "min-w-0 break-all text-gray-900 dark:text-gray-100",
              props.mono ? "font-mono text-xs" : "",
            ],
          },
          String(props.value || "-")
        ),
      ]);
  },
});

const authApi = useAuthService();
const userStore = useUserStore();

const context = ref<AuthMeContextResponse | null>(null);
const pending = ref(false);
const errorMessage = ref("");

const user = computed(() => context.value?.user || {});
const tenant = computed(() => context.value?.tenant || {});
const members = computed(() => context.value?.members || []);
const tenantUuid = computed(() => context.value?.current_tenant_uuid || tenant.value?.uuid || "-");
const currentMember = computed(() => {
  const currentID = context.value?.current_member_id;
  const currentUUID = context.value?.current_member_uuid;
  return members.value.find((member) => {
    if (currentUUID && member.member_uuid === currentUUID) return true;
    if (currentID && Number(member.member_id) === Number(currentID)) return true;
    return member.tenant_uuid === tenantUuid.value;
  }) || members.value[0] || null;
});
const userUuid = computed(() => user.value?.uuid || user.value?.user_uuid || "-");
const memberUuid = computed(() => context.value?.current_member_uuid || currentMember.value?.member_uuid || "-");
const roleText = computed(() => (context.value?.roles || []).join(", ") || "-");

const memberColumns = [
  { accessorKey: "tenant_name", header: "租户" },
  { accessorKey: "tenant_uuid", header: "Tenant UUID" },
  { accessorKey: "member_id", header: "Member ID" },
  { accessorKey: "member_uuid", header: "Member UUID" },
  { accessorKey: "is_admin", header: "管理员" },
];

function displayValue(value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  return String(value);
}

async function loadContext() {
  pending.value = true;
  errorMessage.value = "";
  try {
    const resp = await authApi.getMeContext();
    const data = resp?.data || null;
    context.value = data;
    userStore.setContext(data);
    if (data?.user) {
      userStore.setUser(data.user);
    }
  } catch (error: any) {
    errorMessage.value = error?.response?._data?.message || error?.message || "无法获取当前用户上下文";
  } finally {
    pending.value = false;
  }
}

onMounted(() => {
  void loadContext();
});
</script>
