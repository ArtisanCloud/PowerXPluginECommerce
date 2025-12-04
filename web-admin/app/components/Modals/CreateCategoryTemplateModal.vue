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
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">所属类目</label>
                <USelect v-model="formState.categoryId" :options="categoryOptions" option-attribute="label" value-attribute="value" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">状态</label>
                <USelectMenu v-model="formState.status" :options="statusOptions" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">显示顺序</label>
                <UInput v-model.number="formState.sortOrder" type="number" />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">可选属性组</label>
              <USelectMenu v-model="formState.attributeGroupIds" :options="attributeGroupOptions" multiple placeholder="选择属性组" searchable />
            </div>

            <div class="space-y-2">
              <h4 class="font-medium">已选属性组</h4>
              <UTable :columns="selectedAttributeGroupColumns" :rows="selectedAttributeGroups">
                <template #name-data="{ row }">
                  <div class="font-medium">{{ row.name }}</div>
                  <div class="text-xs text-gray-500">{{ row.code }}</div>
                </template>
                <template #sort-data="{ row }">
                  <UInput v-model.number="row.sort" type="number" class="w-20" />
                </template>
                <template #actions-data="{ index }">
                  <UButton color="red" variant="ghost" icon="i-heroicons-trash" @click="removeSelectedAttributeGroup(index)" />
                </template>
              </UTable>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">描述</label>
              <UTextarea v-model="formState.description" placeholder="输入类目模板描述" />
            </div>
          </UForm>
        </div>
      </UCard>
    </template>

    <template #footer>
      <div class="flex gap-3">
        <UButton variant="ghost" @click="close(false)">取消</UButton>
        <UButton color="primary" :loading="loading" @click="submit">创建模板</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'

const props = defineProps<{ open?: boolean, templateToCopy?: any, availableAttributeGroups: any[], categoryOptions: any[] }>();
const emit = defineEmits<{
  "update:open": [boolean];
  created: [template: any];
  close: [payload: any];
}>();

const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

const loading = ref(false);

const title = computed(() => props.templateToCopy ? '复制模板' : '新建模板');
const description = computed(() => props.templateToCopy ? '基于现有模板创建一个新模板' : '创建一个新的商品类目模板');

const statusOptions = ['enabled', 'disabled', 'deprecated'];

const formState = reactive({
  name: '',
  code: '',
  categoryId: '',
  status: 'enabled',
  sortOrder: 1,
  description: '',
  attributeGroupIds: [] as number[],
  attributeGroups: [] as any[]
});

const attributeGroupOptions = computed(() => {
  return props.availableAttributeGroups.map(group => ({
    label: `${group.name} (${group.code})`,
    value: group.id
  }))
});

const selectedAttributeGroupColumns = [
  { key: 'name', label: '属性组', id: 'name' },
  { key: 'sort', label: '排序', id: 'sort' },
  { key: 'actions', label: '操作', id: 'actions' }
];

const selectedAttributeGroups = computed(() => {
  return formState.attributeGroups.map(group => ({
    ...group,
    sort: group.sort || 0
  }))
});

watch(() => formState.attributeGroupIds, (newIds) => {
  formState.attributeGroups = newIds.map(id => {
    const existing = formState.attributeGroups.find(g => g.id === id);
    if (existing) return existing;
    const group = props.availableAttributeGroups.find(g => g.id === id);
    return { ...group, sort: 0 };
  });
});

const removeSelectedAttributeGroup = (index: number) => {
  const removedId = formState.attributeGroups[index].id;
  formState.attributeGroups.splice(index, 1);
  formState.attributeGroupIds = formState.attributeGroupIds.filter(id => id !== removedId);
};

function resetForm() {
  formState.name = '';
  formState.code = '';
  formState.categoryId = '';
  formState.status = 'enabled';
  formState.sortOrder = 1;
  formState.description = '';
  formState.attributeGroupIds = [];
  formState.attributeGroups = [];
}

watch(() => props.open, (newVal) => {
  if (newVal) {
    if (props.templateToCopy) {
      formState.name = props.templateToCopy.name + ' (副本)';
      formState.code = props.templateToCopy.code + '_copy';
      formState.categoryId = props.templateToCopy.categoryId || props.templateToCopy.category;
      formState.status = props.templateToCopy.status;
      formState.sortOrder = props.templateToCopy.sortOrder;
      formState.description = props.templateToCopy.description;
      formState.attributeGroups = JSON.parse(JSON.stringify(props.templateToCopy.attributeGroups || []));
      formState.attributeGroupIds = (props.templateToCopy.attributeGroups || []).map((g:any) => g.id);
    } else {
      resetForm();
    }
  }
});

function close(payload: any) {
  emit("close", payload);
  emit("update:open", false);
}

async function submit() {
  loading.value = true;
  try {
    await new Promise(resolve => setTimeout(resolve, 500));
    const newTemplate = {
      ...formState,
      id: Date.now(),
      attributeGroupCount: formState.attributeGroups.length,
      updatedAt: new Date().toISOString().split('T')[0]
    };
    emit('created', newTemplate);
    close({ action: 'create', template: newTemplate });
  } catch (error) {
    console.error("Failed to create category template:", error);
  } finally {
    loading.value = false;
  }
}
</script>
