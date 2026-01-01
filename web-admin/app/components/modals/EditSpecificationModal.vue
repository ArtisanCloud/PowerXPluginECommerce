<template>
  <UModal
    v-model:open="isOpen"
    title="编辑规格"
    description="修改商品规格的详细信息"
    :ui="{
      content: 'w-full sm:max-w-2xl',
      body: 'p-0',
      footer: 'justify-end',
    }"
  >
    <template #body>
      <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
        <div class="p-4 sm:p-6">
          <UForm :state="formState" class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">名称</label>
                <UInput v-model="formState.name" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">编码</label>
                <UInput v-model="formState.code" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">作用域</label>
                <URadioGroup v-model="formState.scope" :options="scopeOptions" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">状态</label>
                <USelectMenu v-model="formState.status" :options="statusOptions" />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">类型</label>
              <USelectMenu v-model="formState.dataType" :options="dataTypeOptions" />
            </div>

            <!-- Conditional validation rules -->
            <div v-if="formState.dataType === 'text'" class="p-3 bg-gray-50 dark:bg-gray-800 rounded-md">
              <div class="space-y-2">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">正则表达式</label>
                <UInput v-model="formState.validation.pattern" placeholder="e.g., ^[a-z]+$" />
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">最小长度</label>
                  <UInput v-model.number="formState.validation.minLength" type="number" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">最大长度</label>
                  <UInput v-model.number="formState.validation.maxLength" type="number" />
                </div>
              </div>
            </div>

            <div v-if="formState.dataType === 'enum' || formState.dataType === 'multi_enum'" class="p-3 bg-gray-50 dark:bg-gray-800 rounded-md">
              <h4 class="font-medium mb-2">枚举选项</h4>
              <UTable :columns="enumColumns" :rows="formState.validation.enumOptions">
                <template #value-data="{ row }">
                  <UInput v-model="row.value" placeholder="值" />
                </template>
                <template #alias-data="{ row }">
                  <UInput v-model="row.alias" placeholder="别名,逗号分隔" />
                </template>
                <template #sort-data="{ row }">
                  <UInput v-model.number="row.sort" type="number" class="w-20" />
                </template>
                <template #actions-data="{ index }">
                  <UButton color="red" variant="ghost" icon="i-heroicons-trash" @click="formState.validation.enumOptions.splice(index, 1)" />
                </template>
              </UTable>
              <div class="mt-2">
                <UButton icon="i-heroicons-plus-circle" @click="addNewEnumOption">添加选项</UButton>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">可检索</label>
                <UCheckbox v-model="formState.isSearchable" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">可聚合</label>
                <UCheckbox v-model="formState.isAggregatable" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">可排序</label>
                <UCheckbox v-model="formState.isSortable" />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">搜索权重</label>
              <UInput v-model.number="formState.searchWeight" type="number" :min="0" :max="10" />
            </div>
          </UForm>
        </div>
      </UCard>
    </template>

    <template #footer>
      <div class="flex gap-3">
        <UButton variant="ghost" @click="close(false)">取消</UButton>
        <UButton color="primary" :loading="loading" @click="submit">保存修改</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
const props = defineProps<{ open?: boolean, specification: any }>();
const emit = defineEmits<{
  "update:open": [boolean];
  updated: [specification: any];
  close: [payload: any];
}>();

const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

const loading = ref(false);

const scopeOptions = ['SPU', 'SKU', 'Both'];
const statusOptions = ['enabled', 'disabled', 'deprecated'];
const dataTypeOptions = ['text', 'number', 'boolean', 'enum', 'multi_enum', 'range', 'date', 'richtext', 'json'];

const formState = reactive({
  id: '',
  name: '',
  code: '',
  scope: 'SKU',
  status: 'enabled',
  dataType: 'text',
  validation: {
    pattern: '',
    minLength: null,
    maxLength: null,
    enumOptions: []
  },
  isSearchable: true,
  isAggregatable: false,
  isSortable: false,
  searchWeight: 5
});

const enumColumns = [
  { key: 'value', label: '值', id: 'value' },
  { key: 'alias', label: '别名/同义词', id: 'alias' },
  { key: 'sort', label: '排序', id: 'sort' },
  { key: 'actions', label: '操作', id: 'actions' }
];

watch(() => props.specification, (newSpec) => {
  if (newSpec) {
    formState.id = newSpec.id;
    formState.name = newSpec.name;
    formState.code = newSpec.code;
    formState.scope = newSpec.scope;
    formState.status = newSpec.status;
    formState.dataType = newSpec.dataType;
    formState.isSearchable = newSpec.isSearchable;
    formState.isAggregatable = newSpec.isAggregatable;
    formState.isSortable = newSpec.isSortable;
    formState.searchWeight = newSpec.searchWeight;
    formState.validation = {
      pattern: newSpec.validation?.pattern || '',
      minLength: newSpec.validation?.minLength || null,
      maxLength: newSpec.validation?.maxLength || null,
      enumOptions: newSpec.validation?.enumOptions ? JSON.parse(JSON.stringify(newSpec.validation.enumOptions)) : []
    };
  }
}, { immediate: true, deep: true });

const addNewEnumOption = () => {
  const newSort = formState.validation.enumOptions.length > 0 ? Math.max(...formState.validation.enumOptions.map(o => o.sort || 0)) + 1 : 1;
  formState.validation.enumOptions.push({ value: '', alias: '', sort: newSort });
};

function close(payload: any) {
  emit("close", payload);
  emit("update:open", false);
}

async function submit() {
  loading.value = true;
  try {
    await new Promise(resolve => setTimeout(resolve, 500));
    const updatedSpecification = { ...formState };
    emit('updated', updatedSpecification);
    close({ action: 'update', specification: updatedSpecification });
  } catch (error) {
    console.error("Failed to update specification:", error);
  } finally {
    loading.value = false;
  }
}
</script>
