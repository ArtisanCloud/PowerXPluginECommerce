<template>
  <UModal
    v-model:open="isOpen"
    :title="title"
    :description="description"
    :ui="{
      content: 'w-full sm:max-w-2xl',
      body: 'p-0',
      footer: 'justify-end',
    }"
  >
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="space-y-6 p-4 sm:p-6">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
              >
                调整类型 <span class="text-red-500">*</span>
              </label>
              <USelect
                v-model="form.type"
                :options="adjustTypeOptions"
                option-attribute="label"
                value-attribute="value"
              />
            </div>

            <div>
              <label
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
              >
                成长值数量 <span class="text-red-500">*</span>
              </label>
              <UInput
                v-model.number="form.growthValue"
                type="number"
                placeholder="请输入成长值数量"
              />
            </div>

            <div class="sm:col-span-2">
              <label
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
              >
                调整原因 <span class="text-red-500">*</span>
              </label>
              <UTextarea
                v-model="form.reason"
                placeholder="请输入调整原因"
                rows="3"
              />
            </div>

            <div class="sm:col-span-2" v-if="mode === 'batch'">
              <label
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
              >
                选择客户
              </label>
              <UInput
                v-model="customerSearch"
                placeholder="搜索客户姓名或手机号..."
                icon="i-heroicons-magnifying-glass"
                class="mb-3"
              />
              <div class="border border-gray-200 dark:border-gray-700 rounded-lg max-h-60 overflow-y-auto">
                <div
                  v-for="customer in filteredCustomers"
                  :key="customer.id"
                  class="flex items-center justify-between p-3 border-b border-gray-200 dark:border-gray-700 last:border-b-0"
                >
                  <div class="flex items-center gap-3">
                    <UCheckbox
                      v-model="customer.selected"
                      :value="customer.id"
                    />
                    <UAvatar
                      :src="customer.avatar"
                      :alt="customer.name"
                      size="sm"
                      :ui="{ rounded: 'rounded-full' }"
                    />
                    <div>
                      <div class="font-medium text-gray-900 dark:text-white">
                        {{ customer.name }}
                      </div>
                      <div class="text-sm text-gray-500 dark:text-gray-400">
                        {{ customer.phone }}
                      </div>
                    </div>
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">
                    成长值: {{ customer.growthValue }}
                  </div>
                </div>
              </div>
              <div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
                已选择 {{ selectedCustomers.length }} 个客户
              </div>
            </div>

            <div class="sm:col-span-2" v-if="mode === 'single' && account">
              <label
                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
              >
                客户信息
              </label>
              <div class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-900 rounded-lg">
                <UAvatar
                  :src="account.avatar"
                  :alt="account.customerName"
                  size="md"
                  :ui="{ rounded: 'rounded-full' }"
                />
                <div>
                  <div class="font-medium text-gray-900 dark:text-white">
                    {{ account.customerName }}
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">
                    当前成长值: {{ account.balance }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </UCard>
    </template>

    <template #footer>
      <div class="flex gap-3">
        <UButton variant="ghost" @click="close(false)">取消</UButton>
        <UButton
          color="primary"
          :loading="loading"
          :disabled="!isFormValid"
          @click="submit"
        >
          确认调整
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
const props = defineProps({
  open: {
    type: Boolean,
    default: false
  },
  mode: {
    type: String,
    default: 'single', // 'single' 或 'batch'
    validator: (value: string) => ['single', 'batch'].includes(value)
  },
  account: {
    type: Object,
    default: null
  },
  customers: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['update:open', 'submit', 'close']);

// 控制模态框打开状态
const isOpen = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
});

// 模态框标题和描述
const title = computed(() => {
  return props.mode === 'batch' ? '批量调整成长值' : 
         form.value.type === 'add' ? '赠送成长值' : '扣减成长值';
});

const description = computed(() => {
  return props.mode === 'batch' ? '批量为多个客户账户增加或扣减成长值' :
         form.value.type === 'add' ? '为客户账户增加成长值' : '为客户账户扣减成长值';
});

// 表单数据
const form = ref({
  type: 'add',
  growthValue: 0,
  reason: ''
});

// 客户搜索
const customerSearch = ref('');

// 加载状态
const loading = ref(false);

// 调整类型选项
const adjustTypeOptions = [
  { label: '赠送成长值', value: 'add' },
  { label: '扣减成长值', value: 'deduct' }
];

// 过滤后的客户列表
const filteredCustomers = computed(() => {
  if (!customerSearch.value) return props.customers;
  return props.customers.filter(
    (customer) =>
      customer.name.toLowerCase().includes(customerSearch.value.toLowerCase()) ||
      customer.phone.includes(customerSearch.value)
  );
});

// 选中的客户
const selectedCustomers = computed(() => {
  return props.customers.filter((customer) => customer.selected);
});

// 表单验证
const isFormValid = computed(() => {
  if (props.mode === 'batch') {
    return (
      form.value.growthValue > 0 &&
      form.value.reason.trim() !== '' &&
      selectedCustomers.value.length > 0
    );
  } else {
    return (
      form.value.growthValue > 0 &&
      form.value.reason.trim() !== ''
    );
  }
});

// 关闭模态框
function close(payload) {
  emit('close', payload);
  emit('update:open', false);
  
  // 重置表单
  form.value = {
    type: 'add',
    growthValue: 0,
    reason: ''
  };
  
  // 重置客户搜索
  customerSearch.value = '';
  
  // 重置客户选择状态（如果是批量模式）
  if (props.mode === 'batch') {
    props.customers.forEach((customer) => {
      customer.selected = false;
    });
  }
}

// 提交表单
const submit = async () => {
  if (!isFormValid.value) return;

  loading.value = true;

  try {
    // 准备提交数据
    const data = {
      ...form.value,
      ...(props.mode === 'batch' && { customerIds: selectedCustomers.value.map(c => c.id) }),
      ...(props.mode === 'single' && { accountId: props.account?.id })
    };

    // 触发提交事件
    emit('submit', data);
    
    // 关闭模态框
    close({ action: 'submitted', data });
  } catch (error) {
    console.error('调整成长值失败:', error);
    alert('调整成长值失败，请重试');
  } finally {
    loading.value = false;
  }
};
</script>