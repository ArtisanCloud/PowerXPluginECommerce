<template>
  <UModal
    v-model:open="isOpen"
    title="编辑属性组"
    description="修改商品属性组的详细信息"
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
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">显示顺序</label>
                <UInput v-model.number="formState.sortOrder" type="number" />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">可选属性</label>
              <USelectMenu v-model="formState.attributeIds" :options="attributeOptions" multiple placeholder="选择属性" searchable />
            </div>

            <div class="space-y-2">
              <h4 class="font-medium">已选属性</h4>
              <UTable :columns="selectedAttributeColumns" :rows="selectedAttributes">
                <template #name-data="{ row }">
                  <div class="font-medium">{{ row.name }}</div>
                  <div class="text-xs text-gray-500">{{ row.code }}</div>
                </template>
                <template #required-data="{ row }">
                  <UCheckbox v-model="row.required" />
                </template>
                <template #sort-data="{ row }">
                  <UInput v-model.number="row.sort" type="number" class="w-20" />
                </template>
                <template #actions-data="{ index }">
                  <UButton color="red" variant="ghost" icon="i-heroicons-trash" @click="removeSelectedAttribute(index)" />
                </template>
              </UTable>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">描述</label>
              <UTextarea v-model="formState.description" placeholder="输入属性组描述" />
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
import { ref, reactive, computed, watch } from 'vue'

const props = defineProps<{ open?: boolean, group: any, availableAttributes: any[] }>();
const emit = defineEmits<{
  "update:open": [boolean];
  updated: [group: any];
  close: [payload: any];
}>();

const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

const loading = ref(false);

const statusOptions = ['enabled', 'disabled', 'deprecated'];
const scopeOptions = ['SPU', 'SKU', 'Both'];

const formState = reactive({
  id: '',
  name: '',
  code: '',
  scope: 'SPU',
  status: 'enabled',
  sortOrder: 1,
  description: '',
  attributeIds: [] as number[],
  attributes: [] as any[]
});

const attributeOptions = computed(() => {
  return props.availableAttributes.map(attr => ({
    label: `${attr.name} (${attr.code})`,
    value: attr.id
  }))
});

const selectedAttributeColumns = [
  { key: 'name', label: '属性', id: 'name' },
  { key: 'required', label: '必填', id: 'required' },
  { key: 'sort', label: '排序', id: 'sort' },
  { key: 'actions', label: '操作', id: 'actions' }
];

const selectedAttributes = computed(() => {
  return formState.attributes.map(attr => ({
    ...attr,
    required: attr.required || false,
    sort: attr.sort || 0
  }))
});

watch(() => formState.attributeIds, (newIds, oldIds) => {
  if (JSON.stringify(newIds) === JSON.stringify(oldIds)) return;

  const newAttributes = newIds.map(id => {
    const existing = formState.attributes.find(a => a.id === id);
    if (existing) return existing;
    const attr = props.availableAttributes.find(a => a.id === id);
    return { ...attr, required: false, sort: 0 };
  });
  formState.attributes = newAttributes;
});

const removeSelectedAttribute = (index: number) => {
  const removedId = formState.attributes[index].id;
  formState.attributes.splice(index, 1);
  formState.attributeIds = formState.attributeIds.filter(id => id !== removedId);
};

watch(() => props.group, (newGroup) => {
  if (newGroup) {
    formState.id = newGroup.id;
    formState.name = newGroup.name;
    formState.code = newGroup.code;
    formState.scope = newGroup.scope;
    formState.status = newGroup.status;
    formState.sortOrder = newGroup.sortOrder;
    formState.description = newGroup.description;
    formState.attributes = JSON.parse(JSON.stringify(newGroup.attributes || []));
    formState.attributeIds = (newGroup.attributes || []).map((a:any) => a.id);
  }
}, { immediate: true, deep: true });

function close(payload: any) {
  emit("close", payload);
  emit("update:open", false);
}

async function submit() {
  loading.value = true;
  try {
    await new Promise(resolve => setTimeout(resolve, 500));
    const updatedGroup = { 
      ...formState, 
      attributeCount: formState.attributes.length 
    };
    emit('updated', updatedGroup);
    close({ action: 'update', group: updatedGroup });
  } catch (error) {
    console.error("Failed to update attribute group:", error);
  } finally {
    loading.value = false;
  }
}
</script>
