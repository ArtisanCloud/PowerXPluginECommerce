<template>
  <UModal
    v-model:open="isOpen"
    title="编辑等级"
    description="修改会员等级信息"
    :close="{ onClick: () => close(false) }"
    :ui="{
      content: 'w-full sm:max-w-3xl',
      body: 'p-0',
      footer: 'justify-end',
    }"
  >
    <!-- Body -->
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="space-y-6 p-4 sm:p-6">
          <!-- 等级统计 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div
                class="text-center p-4 bg-blue-50 dark:bg-blue-950 rounded-lg"
              >
                <div
                  class="text-2xl font-bold text-blue-600 dark:text-blue-400"
                >
                  {{ tier.level }}
                </div>
                <div class="text-sm text-blue-600 dark:text-blue-400 mt-1">
                  等级级别
                </div>
              </div>

              <div
                class="text-center p-4 bg-green-50 dark:bg-green-950 rounded-lg"
              >
                <div
                  class="text-2xl font-bold text-green-600 dark:text-green-400"
                >
                  {{ (tier.memberCount || 0).toLocaleString() }}
                </div>
                <div class="text-sm text-green-600 dark:text-green-400 mt-1">
                  当前会员数
                </div>
              </div>

              <div
                class="text-center p-4 bg-purple-50 dark:bg-purple-950 rounded-lg"
              >
                <div
                  class="text-2xl font-bold text-purple-600 dark:text-purple-400"
                >
                  {{ formatDate(tier.createdAt) }}
                </div>
                <div class="text-sm text-purple-600 dark:text-purple-400 mt-1">
                  创建时间
                </div>
              </div>
            </div>
          </div>

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
                  等级名称 <span class="text-red-500">*</span>
                </label>
                <UInput
                  v-model="form.name"
                  placeholder="请输入等级名称"
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
                  等级级别 <span class="text-red-500">*</span>
                </label>
                <UInput
                  v-model.number="form.level"
                  type="number"
                  placeholder="请输入等级级别"
                  :error="!!errors.level"
                />
                <p v-if="errors.level" class="text-red-500 text-xs mt-1">
                  {{ errors.level }}
                </p>
              </div>

              <div class="sm:col-span-2">
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  等级描述
                </label>
                <UTextarea
                  v-model="form.description"
                  placeholder="请输入等级描述"
                  rows="3"
                />
              </div>

              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  等级颜色 <span class="text-red-500">*</span>
                </label>
                <div class="flex items-center gap-3">
                  <input
                    v-model="form.color"
                    type="color"
                    class="w-12 h-10 rounded border border-gray-300 dark:border-gray-600"
                  />
                  <UInput
                    v-model="form.color"
                    placeholder="#000000"
                    class="flex-1"
                    :error="!!errors.color"
                  />
                </div>
                <p v-if="errors.color" class="text-red-500 text-xs mt-1">
                  {{ errors.color }}
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

          <!-- 升级门槛 -->
          <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-4"
            >
              升级门槛
            </h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  消费门槛（元）
                </label>
                <UInput
                  v-model.number="form.spendThreshold"
                  type="number"
                  placeholder="0"
                  min="0"
                  step="0.01"
                />
                <p class="text-gray-500 dark:text-gray-400 text-xs mt-1">
                  累计消费金额达到此数值可升级
                </p>
              </div>

              <div>
                <label
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  积分门槛
                </label>
                <UInput
                  v-model.number="form.pointsThreshold"
                  type="number"
                  placeholder="0"
                  min="0"
                />
                <p class="text-gray-500 dark:text-gray-400 text-xs mt-1">
                  累计积分达到此数值可升级
                </p>
              </div>
            </div>

            <div class="mt-4 p-3 bg-blue-50 dark:bg-blue-950 rounded-lg">
              <div class="flex items-start gap-2">
                <UIcon
                  name="i-heroicons-information-circle"
                  class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5"
                />
                <div class="text-sm text-blue-700 dark:text-blue-300">
                  <strong>升级规则：</strong
                  >满足消费门槛或积分门槛任一条件即可升级到此等级。
                </div>
              </div>
            </div>
          </div>

          <!-- 会员权益 -->
          <div>
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                会员权益
              </h3>
              <UButton
                color="primary"
                variant="outline"
                size="sm"
                icon="i-heroicons-plus"
                @click="addBenefit"
              >
                添加权益
              </UButton>
            </div>

            <div class="space-y-4">
              <div
                v-for="(benefit, index) in form.benefits"
                :key="index"
                class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div class="flex items-start justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">
                    权益 {{ index + 1 }}
                  </h4>
                  <UButton
                    color="error"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="removeBenefit(index)"
                  >
                    删除
                  </UButton>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      权益类型
                    </label>
                    <USelect
                      v-model="benefit.type"
                      :options="benefitTypeOptions"
                      option-attribute="label"
                      value-attribute="value"
                    />
                  </div>

                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      权益值
                    </label>
                    <UInput
                      v-model="benefit.value"
                      :type="getBenefitValueType(benefit.type)"
                      placeholder="请输入权益值"
                    />
                  </div>

                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      权益描述
                    </label>
                    <UInput
                      v-model="benefit.description"
                      placeholder="请输入权益描述"
                    />
                  </div>
                </div>
              </div>

              <div
                v-if="form.benefits.length === 0"
                class="text-center py-8 text-gray-500 dark:text-gray-400"
              >
                <UIcon name="i-heroicons-gift" class="w-12 h-12 mx-auto mb-2" />
                <p>暂无权益，点击上方按钮添加权益</p>
              </div>
            </div>
          </div>

          <!-- 等级规则 -->
          <div>
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                等级规则
              </h3>
              <UButton
                color="primary"
                variant="outline"
                size="sm"
                icon="i-heroicons-plus"
                @click="addRule"
              >
                添加规则
              </UButton>
            </div>

            <div class="space-y-4">
              <div
                v-for="(rule, index) in form.rules"
                :key="index"
                class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div class="flex items-start justify-between mb-3">
                  <h4 class="font-medium text-gray-900 dark:text-white">
                    规则 {{ index + 1 }}
                  </h4>
                  <UButton
                    color="error"
                    variant="ghost"
                    size="sm"
                    icon="i-heroicons-trash"
                    @click="removeRule(index)"
                  >
                    删除
                  </UButton>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-4 gap-4">
                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      规则类型
                    </label>
                    <USelect
                      v-model="rule.type"
                      :options="ruleTypeOptions"
                      option-attribute="label"
                      value-attribute="value"
                    />
                  </div>

                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      条件
                    </label>
                    <UInput
                      v-model="rule.condition"
                      placeholder="例如: >, <, =, >=, <="
                    />
                  </div>

                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      规则值
                    </label>
                    <UInput
                      v-model="rule.value"
                      type="text"
                      placeholder="请输入规则值"
                    />
                  </div>

                  <div>
                    <label
                      class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                    >
                      规则描述
                    </label>
                    <UInput
                      v-model="rule.description"
                      placeholder="请输入规则描述"
                    />
                  </div>
                </div>
              </div>

              <div
                v-if="form.rules.length === 0"
                class="text-center py-8 text-gray-500 dark:text-gray-400"
              >
                <UIcon name="i-heroicons-document-text" class="w-12 h-12 mx-auto mb-2" />
                <p>暂无规则，点击上方按钮添加规则</p>
              </div>
            </div>
          </div>
        </div>
      </UCard>
    </template>

    <!-- Footer -->
    <template #footer>
      <div class="flex justify-between">
        <div class="flex gap-3">
          <UButton
            color="error"
            variant="outline"
            icon="i-heroicons-trash"
            @click="deleteTier"
          >
            删除等级
          </UButton>
          <UButton
            color="primary"
            variant="outline"
            icon="i-heroicons-document-duplicate"
            @click="duplicateTier"
          >
            复制等级
          </UButton>
        </div>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="close(false)">取消</UButton>
          <UButton
            color="primary"
            :loading="loading"
            :disabled="!isFormValid"
            @click="updateTier"
          >
            保存修改
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
interface MembershipTier {
  id: string;
  name: string;
  description: string;
  color: string;
  level: number;
  spendThreshold?: number;
  pointsThreshold?: number;
  benefits: Array<{
    type: string;
    value: number | string;
    description: string;
  }>;
  rules: Array<{
    type: string;
    condition: string;
    value: number | string;
    description: string;
  }>;
  memberCount?: number;
  status: "active" | "inactive";
  createdAt: string;
  updatedAt: string;
}

const props = defineProps<{ open?: boolean; tier: MembershipTier }>();
const emit = defineEmits<{
  "update:open": [boolean];
  updated: [tier: MembershipTier];
  deleted: [tierId: string];
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
  color: "#3B82F6",
  level: 1,
  spendThreshold: 0,
  pointsThreshold: 0,
  status: "active" as "active" | "inactive",
  benefits: [] as Array<{
    type: string;
    value: number | string;
    description: string;
  }>,
  rules: [] as Array<{
    type: string;
    condition: string;
    value: number | string;
    description: string;
  }>,
});

// 表单验证错误
const errors = ref<Record<string, string>>({});

// 状态选项
const statusOptions = [
  { label: "启用", value: "active" },
  { label: "停用", value: "inactive" },
];

// 权益类型选项
const benefitTypeOptions = [
  { label: "购物折扣", value: "discount" },
  { label: "积分倍率", value: "points" },
  { label: "免运费", value: "shipping" },
  { label: "优先服务", value: "priority" },
  { label: "专属活动", value: "exclusive" },
  { label: "专属客服", value: "personal" },
];

// 规则类型选项
const ruleTypeOptions = [
  { label: "消费限制", value: "spend_limit" },
  { label: "积分限制", value: "points_limit" },
  { label: "购买次数", value: "purchase_count" },
  { label: "特殊活动", value: "special_event" },
  { label: "地区限制", value: "region_limit" },
];

// 初始化表单数据
const initializeForm = () => {
  form.value = {
    name: props.tier.name,
    description: props.tier.description,
    color: props.tier.color,
    level: props.tier.level,
    spendThreshold: props.tier.spendThreshold || 0,
    pointsThreshold: props.tier.pointsThreshold || 0,
    status: props.tier.status,
    benefits: Array.isArray(props.tier.benefits)
      ? [...props.tier.benefits]
      : [],
    rules: Array.isArray(props.tier.rules)
      ? [...props.tier.rules]
      : [],
  };
};

// 监听 tier 变化
watch(
  () => props.tier,
  () => {
    if (props.tier) {
      initializeForm();
    }
  },
  { immediate: true }
);

// 表单验证
const isFormValid = computed(() => {
  return (
    form.value.name.trim() !== "" &&
    form.value.level > 0 &&
    form.value.color !== "" &&
    Object.keys(errors.value).length === 0
  );
});

// 验证表单
const validateForm = () => {
  errors.value = {};

  if (!form.value.name?.trim()) {
    errors.value.name = "等级名称不能为空";
  }

  if (!form.value.level || form.value.level <= 0) {
    errors.value.level = "等级级别必须大于0";
  }

  if (!form.value.color) {
    errors.value.color = "请选择等级颜色";
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

// 获取权益值类型
const getBenefitValueType = (type: string) => {
  switch (type) {
    case "discount":
    case "points":
      return "number";
    default:
      return "text";
  }
};

// 添加权益
const addBenefit = () => {
  form.value.benefits.push({
    type: "discount",
    value: "",
    description: "",
  });
};

// 删除权益
const removeBenefit = (index: number) => {
  form.value.benefits.splice(index, 1);
};

// 添加规则
const addRule = () => {
  form.value.rules.push({
    type: "spend_limit",
    condition: "",
    value: "",
    description: "",
  });
};

// 删除规则
const removeRule = (index: number) => {
  form.value.rules.splice(index, 1);
};

// 日期格式化
const formatDate = (iso: string) => new Date(iso).toLocaleDateString("zh-CN");

// 更新等级
const updateTier = async () => {
  validateForm();
  if (!isFormValid.value) return;

  loading.value = true;

  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    const updatedTier: MembershipTier = {
      ...props.tier,
      name: form.value.name,
      description: form.value.description,
      color: form.value.color,
      level: form.value.level,
      spendThreshold: form.value.spendThreshold || 0,
      pointsThreshold: form.value.pointsThreshold || 0,
      benefits: form.value.benefits.filter(
        (b) => b.type && b.value && b.description
      ),
      rules: form.value.rules.filter(
        (r) => r.type && r.condition && r.value && r.description
      ),
      status: form.value.status,
      updatedAt: new Date().toISOString(),
    };

    emit("updated", updatedTier);
    close({ action: "updated", tier: updatedTier });
  } catch (error) {
    console.error("更新等级失败:", error);
  } finally {
    loading.value = false;
  }
};

// 删除等级
const deleteTier = async () => {
  if (!confirm(`确定要删除等级 "${props.tier.name}" 吗？此操作不可撤销。`)) {
    return;
  }

  loading.value = true;

  try {
    // 模拟API调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    emit("deleted", props.tier.id);
    close({ action: "deleted", tierId: props.tier.id });
  } catch (error) {
    console.error("删除等级失败:", error);
  } finally {
    loading.value = false;
  }
};

// 复制等级
const duplicateTier = () => {
  const duplicatedTier: MembershipTier = {
    id: `tier_${Date.now()}`,
    name: `${props.tier.name} (副本)`,
    description: props.tier.description,
    color: props.tier.color,
    level: props.tier.level + 1,
    spendThreshold: props.tier.spendThreshold,
    pointsThreshold: props.tier.pointsThreshold,
    benefits: [...props.tier.benefits],
    memberCount: 0,
    status: "inactive",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  close({ action: "duplicated", tier: duplicatedTier });
};

function close(payload: any) {
  emit("close", payload);
  emit("update:open", false);
}
</script>
