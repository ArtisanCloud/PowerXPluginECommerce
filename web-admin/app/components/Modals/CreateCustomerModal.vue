<template>
  <UModal
    v-model:open="isOpen"
    title="添加客户"
    description="创建新的客户信息"
    :ui="{ body: 'p-0', footer: 'justify-end' }"
  >
    <!-- 表单主体 -->
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="space-y-4 p-4 sm:p-6">
          <!-- 基本信息 -->
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

            <UFormField label="客户来源">
              <USelect
                v-model="form.source"
                :options="sourceOptions"
                placeholder="选择客户来源"
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
                size="sm"
                placeholder="添加标签..."
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
        </div>
      </UCard>
    </template>

    <!-- 操作按钮 -->
    <template #footer>
      <UButton variant="ghost" @click="handleCancel">取消</UButton>
      <UButton :loading="loading" :disabled="!valid" @click="submit">
        创建客户
      </UButton>
    </template>
  </UModal>
</template>

<script setup lang="ts">
/** 允许本地 v-model:open，也允许 useOverlay 打开后默认即为 true */
const props = defineProps<{ open?: boolean }>();
const emit = defineEmits<{
  "update:open": [boolean];
  created: [customer: Customer];
  close: [payload: any]; // useOverlay 通过它拿 result
}>();

// 统一的打开状态
const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

function close(payload: any) {
  emit("close", payload); // 给 overlay：resolve(instance.result)
  emit("update:open", false); // 给本地：关闭
}

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
  totalOrders: number;
  totalSpent: number;
}

const form = ref({
  name: "",
  email: "",
  phone: "",
  customerType: "",
  gender: "",
  birthDate: "",
  address: "",
  membershipTier: "",
  source: "",
  tags: [] as string[],
  notes: "",
});

const newTag = ref("");
const loading = ref(false);

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

const sourceOptions = [
  { value: "website", label: "官网注册" },
  { value: "mobile_app", label: "手机APP" },
  { value: "wechat", label: "微信小程序" },
  { value: "referral", label: "朋友推荐" },
  { value: "advertisement", label: "广告投放" },
  { value: "offline_store", label: "线下门店" },
  { value: "social_media", label: "社交媒体" },
  { value: "other", label: "其他渠道" },
];

// 表单验证
const valid = computed(
  () =>
    form.value.name.trim() &&
    form.value.phone.trim() &&
    form.value.customerType &&
    form.value.membershipTier &&
    form.value.source
);

// 标签管理
const addTag = () => {
  const t = newTag.value.trim();
  if (t && !form.value.tags.includes(t)) {
    form.value.tags.push(t);
    newTag.value = "";
  }
};

const removeTag = (i: number) => form.value.tags.splice(i, 1);

// 取消操作
function handleCancel() {
  close(false); // 给 overlay 一个 falsy 结果
}

// 提交表单
async function submit() {
  if (!valid.value) return;
  loading.value = true;
  
  try {
    // 模拟API调用
    await new Promise((r) => setTimeout(r, 800));
    
    const now = new Date().toISOString();
    const customer: Customer = {
      id: `C${Date.now()}`,
      name: form.value.name,
      email: form.value.email,
      phone: form.value.phone,
      customerType: form.value.customerType,
      gender: form.value.gender || undefined,
      birthDate: form.value.birthDate || undefined,
      address: form.value.address || undefined,
      membershipTier: form.value.membershipTier,
      source: form.value.source,
      tags: form.value.tags,
      notes: form.value.notes || undefined,
      registrationDate: now.split('T')[0], // 只取日期部分
      status: "active",
      totalOrders: 0,
      totalSpent: 0,
    };

    emit("created", customer); // 本地使用
    close({ action: "create", customer }); // overlay 拿 result 的关键
    reset();
  } catch (error) {
    console.error("创建客户失败:", error);
  } finally {
    loading.value = false;
  }
}

// 重置表单
function reset() {
  form.value = {
    name: "",
    email: "",
    phone: "",
    customerType: "",
    gender: "",
    birthDate: "",
    address: "",
    membershipTier: "",
    source: "",
    tags: [],
    notes: "",
  };
  newTag.value = "";
}
</script>