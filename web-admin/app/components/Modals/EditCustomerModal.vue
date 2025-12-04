<template>
  <UModal
    v-model:open="isOpen"
    title="编辑客户"
    description="修改客户信息和设置"
    :close="{ onClick: () => close(false) }"
    :ui="{
      content: 'w-full sm:max-w-2xl',
      body: 'p-0',
      footer: 'justify-between',
    }"
  >
    <!-- Body -->
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="space-y-6 p-4 sm:p-6">
          <!-- 基本信息 -->
          <div class="space-y-4">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <UFormField label="客户姓名" required>
                <UInput v-model="form.name" placeholder="输入客户姓名..." />
              </UFormField>

              <UFormField label="客户类型" required>
                <USelect
                  v-model="form.customerType"
                  :options="customerTypeOptions"
                  placeholder="选择客户类型"
                />
              </UFormField>
            </div>

            <!-- 联系信息 -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <UFormField label="手机号码" required>
                <UInput
                  v-model="form.phone"
                  placeholder="输入手机号码..."
                  type="tel"
                />
              </UFormField>

              <UFormField label="邮箱地址">
                <UInput
                  v-model="form.email"
                  placeholder="输入邮箱地址..."
                  type="email"
                />
              </UFormField>
            </div>

            <!-- 身份信息 -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <UFormField label="性别">
                <USelect
                  v-model="form.gender"
                  :options="genderOptions"
                  placeholder="选择性别"
                />
              </UFormField>

              <UFormField label="出生日期">
                <UInput
                  v-model="form.birthDate"
                  type="date"
                  placeholder="选择出生日期"
                />
              </UFormField>
            </div>
          </div>

          <!-- 地址信息 -->
          <UFormField label="详细地址">
            <UTextarea
              v-model="form.address"
              :rows="3"
              placeholder="输入详细地址..."
            />
          </UFormField>

          <!-- 会员信息 -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <UFormField label="会员等级">
              <USelect
                v-model="form.membershipTier"
                :options="membershipTierOptions"
                placeholder="选择会员等级"
              />
            </UFormField>

            <UFormField label="客户状态">
              <USelect
                v-model="form.status"
                :options="statusOptions"
                placeholder="选择客户状态"
              />
            </UFormField>
          </div>

          <!-- 标签 -->
          <UFormField label="客户标签">
            <div class="flex flex-wrap gap-2 mb-2">
              <UBadge
                v-for="(tag, i) in form.tags"
                :key="i"
                variant="soft"
                class="cursor-pointer"
                @click="removeTag(i)"
              >
                {{ tag }}
                <UIcon name="i-heroicons-x-mark" class="w-3 h-3 ml-1" />
              </UBadge>
            </div>
            <div class="flex gap-2">
              <UInput
                v-model="newTag"
                placeholder="添加标签..."
                size="sm"
                @keyup.enter="addTag"
              />
              <UButton
                size="sm"
                variant="outline"
                :disabled="!newTag.trim()"
                @click="addTag"
              >
                添加
              </UButton>
            </div>
          </UFormField>

          <!-- 备注 -->
          <UFormField label="备注信息">
            <UTextarea
              v-model="form.notes"
              :rows="3"
              placeholder="输入备注信息..."
            />
          </UFormField>

          <!-- 客户统计信息 -->
          <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 text-sm">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <span class="text-gray-500 dark:text-gray-400">注册时间：</span>
                <span class="text-gray-900 dark:text-white ml-2">{{
                  customer.registrationDate
                }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">客户状态：</span>
                <UBadge
                  :color="customer.status === 'active' ? 'success' : 'neutral'"
                  variant="soft"
                  class="ml-2"
                >
                  {{ customer.status === "active" ? "活跃" : "非活跃" }}
                </UBadge>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">订单总数：</span>
                <span class="text-gray-900 dark:text-white ml-2">{{
                  customer.totalOrders || 0
                }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">消费总额：</span>
                <span class="text-gray-900 dark:text-white ml-2"
                  >¥{{ (customer.totalSpent || 0).toLocaleString() }}</span
                >
              </div>
            </div>
          </div>
        </div>
      </UCard>
    </template>

    <!-- Footer -->
    <template #footer>
      <div class="flex items-center justify-between w-full">
        <div class="flex gap-2">
          <UButton
            variant="ghost"
            color="error"
            :loading="isDeleting"
            @click="deleteCustomer"
          >
            删除客户
          </UButton>
        </div>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="close(false)">取消</UButton>
          <UButton
            :loading="isSaving"
            :disabled="!isValid"
            @click="saveCustomer"
          >
            保存修改
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
interface Customer {
  id: string;
  name: string;
  email: string;
  phone: string;
  customerType: string;
  gender?: string;
  birthDate?: string;
  address?: string;
  membershipTier: string;
  source: string;
  tags: string[];
  notes?: string;
  registrationDate: string;
  status: "active" | "inactive";
  totalOrders?: number;
  totalSpent?: number;
}

const props = defineProps<{ open?: boolean; customer: Customer }>();
const emit = defineEmits<{
  "update:open": [boolean];
  updated: [customer: Customer];
  deleted: [customerId: string];
  close: [payload: any]; // overlay 用它来 resolve result
}>();

// 统一控制打开（本地 & overlay）
const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

function close(payload: any) {
  emit("close", payload); // -> overlay instance.result
  emit("update:open", false); // -> 本地 v-model:open
}

// 表单
const form = ref({
  name: "",
  email: "",
  phone: "",
  customerType: "",
  gender: "",
  birthDate: "",
  address: "",
  membershipTier: "",
  status: "" as "active" | "inactive" | "",
  tags: [] as string[],
  notes: "",
});

const newTag = ref("");
const isSaving = ref(false);
const isDeleting = ref(false);

// 选项数据
const customerTypeOptions = [
  { value: "individual", label: "个人客户" },
  { value: "enterprise", label: "企业客户" },
  { value: "vip", label: "VIP客户" },
];

const genderOptions = [
  { value: "male", label: "男" },
  { value: "female", label: "女" },
  { value: "other", label: "其他" },
];

const membershipTierOptions = [
  { value: "bronze", label: "青铜会员" },
  { value: "silver", label: "白银会员" },
  { value: "gold", label: "黄金会员" },
  { value: "platinum", label: "铂金会员" },
  { value: "diamond", label: "钻石会员" },
];

const statusOptions = [
  { value: "active", label: "活跃" },
  { value: "inactive", label: "非活跃" },
];

const isValid = computed(
  () =>
    form.value.name?.trim() &&
    form.value.phone?.trim() &&
    form.value.customerType &&
    form.value.membershipTier &&
    form.value.status
);

// 初始化/重置
function initFromProps() {
  const c = props.customer;
  form.value = {
    name: c.name || "",
    email: c.email || "",
    phone: c.phone || "",
    customerType: c.customerType || "",
    gender: c.gender || "",
    birthDate: c.birthDate || "",
    address: c.address || "",
    membershipTier: c.membershipTier || "",
    status: c.status || "active",
    tags: Array.isArray(c.tags) ? [...c.tags] : [],
    notes: c.notes || "",
  };
  newTag.value = "";
}

watch(() => props.customer, initFromProps, { immediate: true });
watch(isOpen, (v) => {
  if (v) initFromProps();
});

// 标签管理
const addTag = () => {
  const t = newTag.value.trim();
  if (t && !form.value.tags.includes(t)) {
    form.value.tags.push(t);
    newTag.value = "";
  }
};

const removeTag = (i: number) => form.value.tags.splice(i, 1);

// 保存客户
const saveCustomer = async () => {
  if (!isValid.value) return;
  isSaving.value = true;

  try {
    await new Promise((r) => setTimeout(r, 500));

    const updated: Customer = {
      ...props.customer,
      name: form.value.name,
      email: form.value.email,
      phone: form.value.phone,
      customerType: form.value.customerType,
      gender: form.value.gender || undefined,
      birthDate: form.value.birthDate || undefined,
      address: form.value.address || undefined,
      membershipTier: form.value.membershipTier,
      status: form.value.status as "active" | "inactive",
      tags: form.value.tags,
      notes: form.value.notes || undefined,
    };

    emit("updated", updated);
    close({ action: "update", customer: updated });
  } finally {
    isSaving.value = false;
  }
};

// 删除客户
const deleteCustomer = async () => {
  if (!confirm("确定要删除这个客户吗？此操作不可撤销。")) return;
  isDeleting.value = true;

  try {
    await new Promise((r) => setTimeout(r, 500));
    emit("deleted", props.customer.id);
    close({ action: "delete", customerId: props.customer.id });
  } finally {
    isDeleting.value = false;
  }
};
</script>
