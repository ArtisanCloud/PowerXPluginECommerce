<template>
  <div class="grid grid-cols-12 gap-6 auto-rows-fr">
    <UFormField label="SPU 编码" :description="errors.code" class="col-span-12 md:col-span-6">
      <template #default="{ id }">
        <UInput :id="id" v-model.trim="localValue.code" placeholder="SPU-0001" data-testid="spu-code-input" />
      </template>
    </UFormField>
    <UFormField label="名称" :description="errors.name" class="col-span-12 md:col-span-6">
      <template #default="{ id }">
        <UInput :id="id" v-model.trim="localValue.name" placeholder="示例商品" data-testid="spu-name-input" />
      </template>
    </UFormField>
    <UFormField label="类型" class="col-span-12 md:col-span-6">
      <template #default="{ id }">
        <USelect :id="id" v-model="localValue.type" :items="types" data-testid="spu-type-select" />
      </template>
    </UFormField>
    <UFormField label="负责人" class="col-span-12 md:col-span-6">
      <template #default="{ id }">
        <UInput :id="id" v-model.trim="localValue.responsibleUser" placeholder="ops-01" />
      </template>
    </UFormField>
    <UFormField label="类目 ID" :description="errors.categoryId" class="col-span-12 md:col-span-6">
      <template #default="{ id }">
        <UInput :id="id" v-model.trim="localValue.categoryId" placeholder="cat-001" data-testid="spu-category-id" />
      </template>
    </UFormField>
    <UFormField label="类目路径" :description="errors.categoryPath" class="col-span-12 md:col-span-6">
      <template #default="{ id }">
        <UInput :id="id" v-model.trim="localValue.categoryPath" placeholder="root/cat-001" data-testid="spu-category-path" />
      </template>
    </UFormField>
    <UFormField class="col-span-12" label="标签 (逗号分隔)">
      <template #default="{ id }">
        <UInput :id="id" v-model="tagsInput" placeholder="hot,new" />
      </template>
    </UFormField>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: Record<string, any>
}>()
const emit = defineEmits<{
  'update:modelValue': [Record<string, any>]
}>()

const localValue = reactive({
  code: '',
  name: '',
  type: 'one_time',
  categoryId: '',
  categoryPath: '',
  responsibleUser: '',
  tags: [] as string[],
})

const types = [
  { label: '一次性', value: 'one_time' },
  { label: '订阅', value: 'subscription' },
  { label: '组合', value: 'bundle' },
]

const errors = reactive<Record<string, string>>({})
const tagsInput = ref('')
watch(
  () => props.modelValue,
  (val) => {
    Object.assign(localValue, val || {})
    tagsInput.value = (val?.tags || []).join(',')
  },
  { immediate: true }
)

watch(localValue, () => {
  emit('update:modelValue', {
    ...localValue,
    tags: localValue.tags,
  })
})

watch(tagsInput, (val) => {
  localValue.tags = val
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean)
})
</script>
