<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div class="text-sm text-gray-600 dark:text-gray-400">
        <span class="font-medium text-gray-900 dark:text-white">收货地址</span>
        <span class="ml-2 text-xs text-gray-400">最多 20 条</span>
      </div>
      <div class="flex items-center gap-2">
        <UButton size="sm" variant="ghost" icon="i-heroicons-arrow-path" :loading="loading" @click="refresh">
          刷新
        </UButton>
        <UButton v-if="canManage" size="sm" color="primary" icon="i-heroicons-plus" @click="openCreate">
          新增地址
        </UButton>
      </div>
    </div>

    <UAlert
      v-if="error"
      color="red"
      title="加载失败"
      :description="error"
    />

    <div v-if="!loading && !addresses.length" class="rounded-xl border border-dashed border-gray-200 p-6 text-center text-sm text-gray-500 dark:border-gray-800 dark:text-gray-400">
      暂无收货地址
    </div>

    <div class="space-y-3">
      <div
        v-for="a in addresses"
        :key="a.id"
        class="rounded-xl border border-gray-100 bg-white/70 p-4 shadow-sm dark:border-gray-800 dark:bg-gray-900/60"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0 space-y-1">
            <div class="flex flex-wrap items-center gap-2">
              <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                {{ a.shippingAddress.recipientName }}
              </p>
              <p class="text-sm text-gray-600 dark:text-gray-300">
                {{ a.shippingAddress.recipientPhone || "—" }}
              </p>
              <UBadge v-if="a.isDefault" color="primary" variant="soft" size="xs">默认</UBadge>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-300">
              {{ formatAddress(a.shippingAddress) || "—" }}
            </p>
            <p class="text-xs text-gray-400">
              更新：{{ formatTime(a.updatedAt) }}
            </p>
          </div>

          <div class="flex shrink-0 flex-wrap items-center gap-2">
            <UButton
              v-if="canManage && !a.isDefault"
              size="xs"
              variant="outline"
              :loading="busyId === a.id"
              @click="setDefault(a)"
            >
              设为默认
            </UButton>
            <UButton
              v-if="canManage"
              size="xs"
              variant="ghost"
              icon="i-heroicons-pencil-square"
              @click="openEdit(a)"
            >
              编辑
            </UButton>
            <UButton
              v-if="canManage"
              size="xs"
              color="red"
              variant="ghost"
              icon="i-heroicons-trash"
              :loading="busyId === a.id"
              @click="confirmDelete(a)"
            >
              删除
            </UButton>
          </div>
        </div>
      </div>
    </div>

    <UModal
      v-model:open="editOpen"
      :title="editing?.id ? '编辑地址' : '新增地址'"
      :ui="{ content: 'sm:max-w-xl', footer: 'justify-end gap-2' }"
    >
      <template #body>
        <div class="space-y-4 p-4">
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <UFormField label="收货人" required>
              <UInput v-model="form.recipientName" placeholder="姓名" />
            </UFormField>
            <UFormField label="手机号" required>
              <UInput v-model="form.recipientPhone" placeholder="手机号" />
            </UFormField>
          </div>
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <UFormField label="省">
              <UInput v-model="form.province" placeholder="省" />
            </UFormField>
            <UFormField label="市">
              <UInput v-model="form.city" placeholder="市" />
            </UFormField>
            <UFormField label="区">
              <UInput v-model="form.district" placeholder="区" />
            </UFormField>
          </div>
          <UFormField label="详细地址" required>
            <UInput v-model="form.address1" placeholder="街道/门牌号" />
          </UFormField>
          <UFormField label="地址补充">
            <UInput v-model="form.address2" placeholder="楼栋/单元/门禁等（可选）" />
          </UFormField>
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <UFormField label="邮编">
              <UInput v-model="form.postalCode" placeholder="邮编（可选）" />
            </UFormField>
            <UFormField label="默认地址">
              <USelect v-model="form.isDefault" :items="defaultItems" class="w-full" />
            </UFormField>
          </div>
        </div>
      </template>
      <template #footer>
        <UButton variant="ghost" :disabled="saving" @click="editOpen = false">取消</UButton>
        <UButton color="primary" :loading="saving" :disabled="saving || !isFormValid" @click="save">
          保存
        </UButton>
      </template>
    </UModal>

    <UModal
      v-model:open="deleteOpen"
      title="删除地址"
      :ui="{ content: 'sm:max-w-lg', footer: 'justify-end gap-2' }"
    >
      <template #body>
        <div class="space-y-2 p-4 text-sm text-gray-600 dark:text-gray-300">
          <p>确认删除该地址？</p>
          <p class="text-xs text-gray-400">
            {{ deletingTarget ? `${deletingTarget.shippingAddress.recipientName} · ${formatAddress(deletingTarget.shippingAddress)}` : "" }}
          </p>
        </div>
      </template>
      <template #footer>
        <UButton variant="ghost" :disabled="busyId !== ''" @click="deleteOpen = false">取消</UButton>
        <UButton color="red" :loading="busyId === deletingTarget?.id" :disabled="!deletingTarget" @click="doDelete">
          删除
        </UButton>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useToastAlert } from "~/composables/useToastAlert";
import { useCustomerAddressService, type CustomerAddressDTO, type ShippingAddress } from "~/composables/api/services/customerAddressService";

const props = defineProps<{
  customerId?: string | null;
  canManage?: boolean;
}>();

const toast = useToastAlert();
const svc = useCustomerAddressService();

const canManage = computed(() => Boolean(props.canManage));
const customerId = computed(() => String(props.customerId || "").trim());

const addresses = ref<CustomerAddressDTO[]>([]);
const loading = ref(false);
const error = ref("");
const busyId = ref<string>("");

const editOpen = ref(false);
const deleteOpen = ref(false);
const editing = ref<CustomerAddressDTO | null>(null);
const deletingTarget = ref<CustomerAddressDTO | null>(null);
const saving = ref(false);

const defaultItems = [
  { label: "是", value: true },
  { label: "否", value: false },
];

const form = ref({
  recipientName: "",
  recipientPhone: "",
  province: "",
  city: "",
  district: "",
  address1: "",
  address2: "",
  postalCode: "",
  isDefault: false,
});

const isFormValid = computed(() => {
  return Boolean(form.value.recipientName.trim() && form.value.recipientPhone.trim() && form.value.address1.trim());
});

function formatAddress(a: ShippingAddress) {
  const parts = [a.province, a.city, a.district, a.address1, a.address2].filter(Boolean);
  return parts.join("");
}

function formatTime(value?: string) {
  const s = String(value || "").trim();
  if (!s) return "—";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(d);
}

async function refresh() {
  const id = customerId.value;
  if (!id) return;
  loading.value = true;
  error.value = "";
  try {
    const list = await svc.listCustomerAddresses(id);
    addresses.value = (list || []).slice(0, 20);
  } catch (e: any) {
    error.value = e?.data?.message || e?.message || "请求失败";
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  form.value = {
    recipientName: "",
    recipientPhone: "",
    province: "",
    city: "",
    district: "",
    address1: "",
    address2: "",
    postalCode: "",
    isDefault: addresses.value.length === 0,
  };
}

function openCreate() {
  if (!canManage.value) return;
  editing.value = null;
  resetForm();
  editOpen.value = true;
}

function openEdit(a: CustomerAddressDTO) {
  if (!canManage.value) return;
  editing.value = a;
  form.value = {
    recipientName: a.shippingAddress.recipientName || "",
    recipientPhone: a.shippingAddress.recipientPhone || "",
    province: a.shippingAddress.province || "",
    city: a.shippingAddress.city || "",
    district: a.shippingAddress.district || "",
    address1: a.shippingAddress.address1 || "",
    address2: a.shippingAddress.address2 || "",
    postalCode: a.shippingAddress.postalCode || "",
    isDefault: Boolean(a.isDefault),
  };
  editOpen.value = true;
}

async function save() {
  const id = customerId.value;
  if (!id) return;
  if (!isFormValid.value) return;
  saving.value = true;
  try {
    const payload = {
      isDefault: Boolean(form.value.isDefault),
      shippingAddress: {
        recipientName: form.value.recipientName.trim(),
        recipientPhone: form.value.recipientPhone.trim(),
        countryCode: "CN",
        province: form.value.province.trim() || undefined,
        city: form.value.city.trim() || undefined,
        district: form.value.district.trim() || undefined,
        address1: form.value.address1.trim(),
        address2: form.value.address2.trim() || undefined,
        postalCode: form.value.postalCode.trim() || undefined,
      },
    };

    if (editing.value?.id) {
      await svc.updateCustomerAddress(id, editing.value.id, payload);
    } else {
      if (addresses.value.length >= 20) {
        toast.add({ title: "最多可保存 20 个地址", color: "red" });
        return;
      }
      await svc.createCustomerAddress(id, payload);
    }
    editOpen.value = false;
    await refresh();
    toast.add({ title: "已保存", color: "green" });
  } catch (e: any) {
    toast.add({ title: "保存失败", description: e?.data?.message || e?.message, color: "red" });
  } finally {
    saving.value = false;
  }
}

async function setDefault(a: CustomerAddressDTO) {
  const id = customerId.value;
  if (!id) return;
  busyId.value = a.id;
  try {
    await svc.setDefaultCustomerAddress(id, a.id);
    await refresh();
    toast.add({ title: "已设为默认", color: "green" });
  } catch (e: any) {
    toast.add({ title: "设置失败", description: e?.data?.message || e?.message, color: "red" });
  } finally {
    busyId.value = "";
  }
}

function confirmDelete(a: CustomerAddressDTO) {
  deletingTarget.value = a;
  deleteOpen.value = true;
}

async function doDelete() {
  const id = customerId.value;
  const target = deletingTarget.value;
  if (!id || !target?.id) return;
  busyId.value = target.id;
  try {
    await svc.deleteCustomerAddress(id, target.id);
    deleteOpen.value = false;
    deletingTarget.value = null;
    await refresh();
    toast.add({ title: "已删除", color: "green" });
  } catch (e: any) {
    toast.add({ title: "删除失败", description: e?.data?.message || e?.message, color: "red" });
  } finally {
    busyId.value = "";
  }
}

watch(
  () => customerId.value,
  () => {
    addresses.value = [];
    if (customerId.value) void refresh();
  },
  { immediate: true },
);

onMounted(() => {
  if (customerId.value) void refresh();
});
</script>

