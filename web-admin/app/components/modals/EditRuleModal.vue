<template>
  <UModal
    v-model:open="isOpen"
    :title="isEditing ? '编辑规则' : '新建规则'"
    :description="getModalDescription()"
    :close="{ onClick: () => close(false) }"
    :ui="{
      content: 'w-full sm:max-w-2xl',
      body: 'p-0',
      footer: 'justify-end',
    }"
  >
    <!-- Body -->
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="space-y-6 p-4 sm:p-6">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                规则类型 <span class="text-red-500">*</span>
              </label>
              <USelect
                v-model="form.type"
                :options="ruleTypeOptions"
                option-attribute="label"
                value-attribute="value"
                :disabled="isEditing"
                :error="!!errors.type"
              />
              <p v-if="errors.type" class="text-red-500 text-xs mt-1">
                {{ errors.type }}
              </p>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                规则名称 <span class="text-red-500">*</span>
              </label>
              <UInput
                v-model="form.name"
                placeholder="请输入规则名称"
                :error="!!errors.name"
              />
              <p v-if="errors.name" class="text-red-500 text-xs mt-1">
                {{ errors.name }}
              </p>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                规则值 <span class="text-red-500">*</span>
              </label>
              <UInput
                v-model="form.value"
                placeholder="请输入规则值"
                :error="!!errors.value"
              />
              <p v-if="errors.value" class="text-red-500 text-xs mt-1">
                {{ errors.value }}
              </p>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                状态
              </label>
              <USelect
                v-model="form.status"
                :options="statusOptions"
                option-attribute="label"
                value-attribute="value"
              />
            </div>

            <div class="sm:col-span-2">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                描述
              </label>
              <UTextarea
                v-model="form.description"
                placeholder="请输入规则描述"
                rows="3"
              />
            </div>
          </div>
        </div>
      </UCard>
    </template>

    <!-- Footer -->
    <template #footer>
      <div class="flex gap-3">
        <UButton variant="ghost" @click="close(false)">取消</UButton>
        <UButton
          color="primary"
          :loading="loading"
          :disabled="!isFormValid"
          @click="saveRule"
        >
          {{ isEditing ? '保存' : '创建' }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
interface Rule {
  id: string;
  type: string;
  name: string;
  value: string;
  status: string;
  description: string;
}

const props = defineProps<{ 
  open?: boolean; 
  rule?: Rule;
  ruleType: string;
}>();

const emit = defineEmits<{
  "update:open": [boolean];
  saved: [rule: Rule];
  close: [payload: any];
}>();

// 统一控制打开（本地 & overlay）
const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

const loading = ref(false);
const isEditing = computed(() => !!props.rule?.id);

// 表单数据
const form = ref({
  id: "",
  type: props.ruleType,
  name: "",
  value: "",
  status: "enabled",
  description: ""
});

// 表单验证错误
const errors = ref<Record<string, string>>({});

// 规则类型选项
const ruleTypeOptions = [
  { label: "奖励对象", value: "rewardTargets" },
  { label: "奖励内容", value: "rewardContents" },
  { label: "触发条件", value: "triggerConditions" },
  { label: "分销比例", value: "distributionRatios" },
  { label: "结算方式", value: "settlementMethods" },
  { label: "提现规则", value: "withdrawalRules" },
];

// 状态选项
const statusOptions = [
  { label: "启用", value: "enabled" },
  { label: "禁用", value: "disabled" },
];

// 表单验证
const isFormValid = computed(() => {
  return (
    form.value.type !== "" &&
    form.value.name.trim() !== "" &&
    form.value.value !== "" &&
    Object.keys(errors.value).length === 0
  );
});

// 验证表单
const validateForm = () => {
  errors.value = {};

  if (!form.value.type) {
    errors.value.type = "请选择规则类型";
  }

  if (!form.value.name.trim()) {
    errors.value.name = "规则名称不能为空";
  }

  if (!form.value.value) {
    errors.value.value = "规则值不能为空";
  }
};

// 监听表单变化进行验证
watch(
  () => form.value,
  () => {
    validateForm();
  },
  { deep: true }
);

// 获取模态框描述
const getModalDescription = () => {
  const typeLabels: Record<string, string> = {
    rewardTargets: "配置奖励对象规则",
    rewardContents: "配置奖励内容规则",
    triggerConditions: "配置触发条件规则",
    distributionRatios: "配置分销比例规则",
    settlementMethods: "配置结算方式规则",
    withdrawalRules: "配置提现规则"
  };
  return typeLabels[form.value.type] || "配置规则";
};

// 保存规则
const saveRule = async () => {
  validateForm();
  if (!isFormValid.value) return;

  loading.value = true;

  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 500));

    const ruleData: Rule = {
      id: isEditing.value ? form.value.id : `rule_${Date.now()}`,
      type: form.value.type,
      name: form.value.name,
      value: form.value.value,
      status: form.value.status,
      description: form.value.description
    };

    emit("saved", ruleData);
    close({ action: isEditing.value ? "updated" : "created", rule: ruleData });
  } catch (error) {
    console.error("保存规则失败:", error);
  } finally {
    loading.value = false;
  }
};

// 初始化表单
watch(
  () => props.rule,
  (newRule) => {
    if (newRule) {
      form.value = { ...newRule };
    } else {
      form.value = {
        id: "",
        type: props.ruleType,
        name: "",
        value: "",
        status: "enabled",
        description: ""
      };
    }
  },
  { immediate: true }
);

function close(payload: any) {
  emit("close", payload);
  emit("update:open", false);

  // 重置表单
  form.value = {
    id: "",
    type: props.ruleType,
    name: "",
    value: "",
    status: "enabled",
    description: ""
  };
  errors.value = {};
}
</script>