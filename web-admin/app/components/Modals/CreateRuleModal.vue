<template>
  <UModal
    v-model:open="isOpen"
    title="创建规则"
    description="创建新的会员等级规则"
    :close="{ onClick: () => close(false) }"
    :ui="{
      content: 'w-full sm:max-w-4xl',
      body: 'p-0',
      footer: 'justify-end',
    }"
  >
    <!-- Body -->
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="space-y-6 p-4 sm:p-6">
          <!-- 基本信息 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              基本信息
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
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
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  规则类型 <span class="text-red-500">*</span>
                </label>
                <USelect
                  v-model="form.type"
                  :options="ruleTypeOptions"
                  option-attribute="label"
                  value-attribute="value"
                  :error="!!errors.type"
                />
                <p v-if="errors.type" class="text-red-500 text-xs mt-1">
                  {{ errors.type }}
                </p>
              </div>

              <div class="sm:col-span-2">
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  规则描述
                </label>
                <UTextarea
                  v-model="form.description"
                  placeholder="请输入规则描述"
                  rows="3"
                />
              </div>

              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  优先级 <span class="text-red-500">*</span>
                </label>
                <UInput
                  v-model.number="form.priority"
                  type="number"
                  placeholder="数字越小优先级越高"
                  min="1"
                  :error="!!errors.priority"
                />
                <p v-if="errors.priority" class="text-red-500 text-xs mt-1">
                  {{ errors.priority }}
                </p>
              </div>

              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  状态
                </label>
                <USelect
                  v-model="form.status"
                  :options="statusOptions"
                  option-attribute="label"
                  value-attribute="value"
                />
              </div>
            </div>
          </div>

          <!-- 触发条件 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                触发条件
              </h3>
              <UButton
                color="primary"
                variant="outline"
                size="sm"
                icon="i-heroicons-plus"
                @click="addCondition"
              >
                添加条件
              </UButton>
            </div>

            <div class="space-y-4">
              <div
                v-for="(condition, index) in form.conditions"
                :key="index"
                class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div class="flex items-start justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">
                    条件 {{ index + 1 }}
                  </h4>
                  <UButton
                    color="error"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="removeCondition(index)"
                  >
                    删除
                  </UButton>
                </div>

                <div
                  class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4"
                >
                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      条件类型
                    </label>
                    <USelect
                      v-model="condition.type"
                      :options="conditionTypeOptions"
                      option-attribute="label"
                      value-attribute="value"
                    />
                  </div>

                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      操作符
                    </label>
                    <USelect
                      v-model="condition.operator"
                      :options="operatorOptions"
                      option-attribute="label"
                      value-attribute="value"
                    />
                  </div>

                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      目标值
                    </label>
                    <UInput
                      v-model.number="condition.value"
                      type="number"
                      placeholder="请输入目标值"
                      min="0"
                    />
                  </div>

                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      时间周期（天）
                    </label>
                    <UInput
                      v-model.number="condition.period"
                      type="number"
                      placeholder="可选，留空表示累计"
                      min="1"
                    />
                  </div>
                </div>

                <div class="mt-4">
                  <label
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                  >
                    条件描述
                  </label>
                  <UInput
                    v-model="condition.description"
                    placeholder="请输入条件描述"
                  />
                </div>
              </div>

              <div
                v-if="form.conditions.length === 0"
                class="text-center py-8 text-gray-500 dark:text-gray-400"
              >
                <UIcon
                  name="i-heroicons-cog-6-tooth"
                  class="w-12 h-12 mx-auto mb-2"
                />
                <p>暂无条件，点击上方按钮添加触发条件</p>
              </div>
            </div>

            <div class="mt-4 p-3 bg-blue-50 dark:bg-blue-950 rounded-lg">
              <div class="flex items-start gap-2">
                <UIcon
                  name="i-heroicons-information-circle"
                  class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5"
                />
                <div class="text-sm text-blue-700 dark:text-blue-300">
                  <strong>执行逻辑：</strong>
                  {{
                    form.type === "promotion"
                      ? "满足所有条件时触发晋升"
                      : "满足任一条件时触发降级"
                  }}
                </div>
              </div>
            </div>
          </div>

          <!-- 有效期设置 -->
          <div>
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              有效期设置
            </h3>

            <div class="space-y-4">
              <div class="flex items-center gap-3">
                <UCheckbox v-model="form.hasExpiry" label="设置有效期" />
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  勾选后可设置规则的有效期限
                </p>
              </div>

              <div
                v-if="form.hasExpiry"
                class="grid grid-cols-1 sm:grid-cols-2 gap-6"
              >
                <div>
                  <label
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                  >
                    有效期类型
                  </label>
                  <USelect
                    v-model="form.expiryType"
                    :options="expiryTypeOptions"
                    option-attribute="label"
                    value-attribute="value"
                  />
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {{
                      form.expiryType === "absolute"
                        ? "从规则创建时开始计算"
                        : "从满足条件时开始计算"
                    }}
                  </p>
                </div>

                <div>
                  <label
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                  >
                    有效天数
                  </label>
                  <UInput
                    v-model.number="form.expiryDays"
                    type="number"
                    placeholder="请输入有效天数"
                    min="1"
                  />
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    约 {{ Math.round((form.expiryDays || 0) / 30) }} 个月
                  </p>
                </div>
              </div>

              <div
                v-if="form.hasExpiry"
                class="p-3 bg-yellow-50 dark:bg-yellow-950 rounded-lg"
              >
                <div class="flex items-start gap-2">
                  <UIcon
                    name="i-heroicons-exclamation-triangle"
                    class="w-5 h-5 text-yellow-600 dark:text-yellow-400 mt-0.5"
                  />
                  <div class="text-sm text-yellow-700 dark:text-yellow-300">
                    <strong>到期处理：</strong>
                    {{
                      form.type === "promotion"
                        ? "等级将在到期后自动降级到上一级别"
                        : "降级规则到期后将不再生效"
                    }}
                  </div>
                </div>
              </div>
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
          @click="createRule"
        >
          创建规则
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
interface MembershipRule {
  id: string;
  name: string;
  description: string;
  type: "promotion" | "demotion";
  conditions: Array<{
    type: "spend" | "growth" | "frequency";
    operator: "gte" | "lte" | "eq";
    value: number;
    period?: number;
    description: string;
  }>;
  hasExpiry: boolean;
  expiryType?: "absolute" | "relative";
  expiryDays?: number;
  status: "active" | "inactive";
  priority: number;
  createdAt: string;
  updatedAt: string;
}

const props = defineProps<{ open?: boolean }>();
const emit = defineEmits<{
  "update:open": [boolean];
  created: [rule: MembershipRule];
  close: [payload: any];
}>();

// 统一控制打开（本地 & overlay）
const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

const loading = ref(false);

// 表单数据
const form = ref({
  name: "",
  description: "",
  type: "promotion" as "promotion" | "demotion",
  priority: 1,
  status: "active" as "active" | "inactive",
  conditions: [] as Array<{
    type: "spend" | "growth" | "frequency";
    operator: "gte" | "lte" | "eq";
    value: number;
    period?: number;
    description: string;
  }>,
  hasExpiry: false,
  expiryType: "relative" as "absolute" | "relative",
  expiryDays: 365,
});

// 表单验证错误
const errors = ref<Record<string, string>>({});

// 规则类型选项
const ruleTypeOptions = [
  { label: "晋升规则", value: "promotion" },
  { label: "降级规则", value: "demotion" },
];

// 状态选项
const statusOptions = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

// 条件类型选项
const conditionTypeOptions = [
  { label: "消费金额", value: "spend" },
  { label: "成长值", value: "growth" },
  { label: "消费次数", value: "frequency" },
];

// 操作符选项
const operatorOptions = [
  { label: "大于等于 (≥)", value: "gte" },
  { label: "小于等于 (≤)", value: "lte" },
  { label: "等于 (=)", value: "eq" },
];

// 有效期类型选项
const expiryTypeOptions = [
  { label: "相对期限", value: "relative" },
  { label: "绝对期限", value: "absolute" },
];

// 表单验证
const isFormValid = computed(() => {
  return (
    form.value.name.trim() !== "" &&
    form.value.type !== "" &&
    form.value.priority > 0 &&
    form.value.conditions.length > 0 &&
    Object.keys(errors.value).length === 0
  );
});

// 验证表单
const validateForm = () => {
  errors.value = {};

  if (!form.value.name.trim()) {
    errors.value.name = "规则名称不能为空";
  }

  if (!form.value.type) {
    errors.value.type = "请选择规则类型";
  }

  if (!form.value.priority || form.value.priority <= 0) {
    errors.value.priority = "优先级必须大于0";
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

// 添加条件
const addCondition = () => {
  form.value.conditions.push({
    type: "spend",
    operator: "gte",
    value: 0,
    description: "",
  });
};

// 删除条件
const removeCondition = (index: number) => {
  form.value.conditions.splice(index, 1);
};

// 创建规则
const createRule = async () => {
  validateForm();
  if (!isFormValid.value) return;

  loading.value = true;

  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    const newRule: MembershipRule = {
      id: `rule_${Date.now()}`,
      name: form.value.name,
      description: form.value.description,
      type: form.value.type,
      priority: form.value.priority,
      status: form.value.status,
      conditions: form.value.conditions.filter(
        (c) => c.type && c.value !== undefined && c.description
      ),
      hasExpiry: form.value.hasExpiry,
      expiryType: form.value.hasExpiry ? form.value.expiryType : undefined,
      expiryDays: form.value.hasExpiry ? form.value.expiryDays : undefined,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    emit("created", newRule);
    close({ action: "created", rule: newRule });
  } catch (error) {
    console.error("创建规则失败:", error);
  } finally {
    loading.value = false;
  }
};

function close(payload: any) {
  emit("close", payload);
  emit("update:open", false);

  // 重置表单
  form.value = {
    name: "",
    description: "",
    type: "promotion",
    priority: 1,
    status: "active",
    conditions: [],
    hasExpiry: false,
    expiryType: "relative",
    expiryDays: 365,
  };
  errors.value = {};
}
</script>
