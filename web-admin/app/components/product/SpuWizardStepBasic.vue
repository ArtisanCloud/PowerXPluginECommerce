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
    <UFormField label="类目" :description="errors.categoryId" class="col-span-12 md:col-span-6">
      <template #default="{ id }">
        <div class="flex gap-2">
          <UInput
            :id="id"
            :model-value="selectedCategoryDisplay"
            placeholder="请选择类目"
            data-testid="spu-category-id"
            class="flex-1"
            readonly
          />
          <UButton variant="soft" icon="i-heroicons-squares-2x2" @click="openCategoryPicker">选择</UButton>
        </div>
      </template>
    </UFormField>
    <UFormField label="类目路径" :description="errors.categoryPath" class="col-span-12 md:col-span-6">
      <template #default="{ id }">
        <UInput
          :id="id"
          :model-value="selectedCategoryPathDisplay"
          placeholder="选择类目后自动生成"
          data-testid="spu-category-path"
          readonly
        />
      </template>
    </UFormField>
    <UFormField class="col-span-12" label="标签 (逗号分隔)">
      <template #default="{ id }">
        <UInput :id="id" v-model="tagsInput" placeholder="hot,new" />
      </template>
    </UFormField>

    <div v-if="templateFields.length" class="col-span-12">
      <div class="mb-2 flex items-center justify-between">
        <div class="text-sm font-semibold text-white/90">模板字段（attributes）</div>
        <div class="text-xs text-white/50">随类目生效模板变化</div>
      </div>
      <div class="grid grid-cols-12 gap-4 rounded-lg border border-white/10 bg-white/5 p-4">
        <div v-for="field in templateFields" :key="field.fieldKey" class="col-span-12 md:col-span-6">
          <UFormField :label="fieldLabel(field)" :description="field.required ? '必填' : ''">
            <template #default="{ id }">
              <UInput
                v-if="field.fieldType === 'string' || field.fieldType === 'date'"
                :id="id"
                v-model="localValue.attributes[field.fieldKey]"
                :placeholder="`输入 ${field.fieldKey}`"
              />
              <UInput
                v-else-if="field.fieldType === 'number'"
                :id="id"
                v-model.number="localValue.attributes[field.fieldKey]"
                type="number"
              />
              <USwitch
                v-else-if="field.fieldType === 'boolean'"
                v-model="localValue.attributes[field.fieldKey]"
              />
              <USelect
                v-else-if="field.fieldType === 'enum'"
                :id="id"
                v-model="localValue.attributes[field.fieldKey]"
                :items="enumItems(field)"
                class="w-full"
              />
              <UTextarea
                v-else
                :id="id"
                v-model="localValue.attributes[field.fieldKey]"
                :rows="2"
                placeholder="JSON / 任意内容"
              />
            </template>
          </UFormField>
        </div>
      </div>
    </div>
  </div>

  <UModal
    v-model:open="categoryPickerOpen"
    title="选择类目"
    description="请选择一个类目用于 SPU 归类。"
    :close="{ onClick: closeCategoryPicker }"
    :ui="{
      content: 'max-w-2xl w-[90vw] overflow-visible',
      body: 'space-y-4 p-4 sm:p-5 overflow-visible',
      footer: 'px-4 sm:px-5 pb-4 sm:pb-5 pt-0 flex justify-end gap-2',
    }"
  >
    <template #body>
      <UInput v-model="categoryKeyword" icon="i-heroicons-magnifying-glass-20-solid" placeholder="搜索类目" clearable />
      <USelectMenu
        v-model="selectedCategoryId"
        :items="filteredCategoryOptions"
        value-key="value"
        label-key="label"
        :portal="false"
        :popper="{ placement: 'top-start' }"
        placeholder="请选择类目"
        class="w-full"
      />
    </template>
    <template #footer>
      <UButton variant="ghost" @click="closeCategoryPicker">取消</UButton>
      <UButton color="primary" :disabled="!selectedCategoryId" @click="confirmCategory">确认</UButton>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useToastAlert } from '~/composables/useToastAlert'
import { useCategoryApi, type CategoryNode } from '~/composables/api/useCategory'
import { useCategoryTemplateApi, type CategoryTemplateField, type EffectiveTemplateResponse } from '~/composables/api/useCategoryTemplate'

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
  attributes: {} as Record<string, any>,
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
    if (!localValue.attributes) {
      localValue.attributes = {}
    }
  },
  { immediate: true }
)

watch(localValue, () => {
  emit('update:modelValue', {
    ...localValue,
    tags: localValue.tags,
    attributes: localValue.attributes,
  })
})

watch(tagsInput, (val) => {
  localValue.tags = val
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean)
})

const toast = useToastAlert()
const categoryApi = useCategoryApi()
const templateApi = useCategoryTemplateApi()

const categoryPickerOpen = ref(false)
const categoryKeyword = ref('')
const selectedCategoryId = ref<string | null>(null)
const categoryTree = ref<CategoryNode[]>([])
const categoryMap = ref<Record<string, CategoryNode>>({})
const categoryTreeLoading = ref(false)

const templateFields = ref<CategoryTemplateField[]>([])
const effectiveTemplate = ref<EffectiveTemplateResponse | null>(null)

const selectedCategoryNode = computed(() => {
	const id = String(localValue.categoryId || '').trim()
	return id ? categoryMap.value[id] ?? null : null
})

const selectedCategoryDisplay = computed(() => {
	const node = selectedCategoryNode.value
	if (node) {
		const alias = (node.aliasSlug || node.code || '').trim()
		return alias ? `${node.displayName}（${alias}）` : node.displayName
	}
	return String(localValue.categoryId || '').trim()
})

const selectedCategoryPathDisplay = computed(() => {
	const raw = String(localValue.categoryPath || '').trim()
	if (!raw) return ''
	const ids = parseCategoryPathIDs(raw)
	if (!ids.length) return raw
	const parts = ids
		.map((id) => categoryMap.value[id])
		.filter(Boolean)
		.map((node) => (node?.aliasSlug || node?.code || node?.displayName || node?.id || '').trim())
		.filter(Boolean)
	if (parts.length) {
		return parts.join(' / ')
	}
	return raw
})

const blurActiveElement = () => {
  if (typeof document === 'undefined') return
  const active = document.activeElement as HTMLElement | null
  active?.blur()
}

const ensureCategoryTreeLoaded = async () => {
	if (categoryTree.value.length || categoryTreeLoading.value) return
	try {
		categoryTreeLoading.value = true
		const resp = await categoryApi.tree()
		categoryTree.value = resp.items ?? []
		categoryMap.value = buildCategoryMap(categoryTree.value)
	} catch (error: any) {
		toast.add({ title: '加载类目失败', description: error?.message || '无法获取类目树', color: 'error' })
	} finally {
		categoryTreeLoading.value = false
	}
}

const closeCategoryPicker = () => {
  blurActiveElement()
  categoryPickerOpen.value = false
}

const openCategoryPicker = async () => {
	categoryPickerOpen.value = true
	selectedCategoryId.value = localValue.categoryId || null
	await ensureCategoryTreeLoaded()
}

const categoryOptions = computed(() => flattenTree(categoryTree.value))
const filteredCategoryOptions = computed(() => {
	const kw = categoryKeyword.value.trim().toLowerCase()
	if (!kw) return categoryOptions.value
	return categoryOptions.value.filter((opt) => String(opt.label).toLowerCase().includes(kw))
})

const confirmCategory = async () => {
	if (!selectedCategoryId.value) return
	const node = categoryMap.value[selectedCategoryId.value]
	if (!node) return
	localValue.categoryId = node.id
	localValue.categoryPath = node.path
	closeCategoryPicker()
	await loadEffectiveTemplate()
}

watch(
	() => localValue.categoryId,
	(id) => {
		const trimmed = String(id || '').trim()
		if (trimmed) {
			ensureCategoryTreeLoaded()
		}
		loadEffectiveTemplate()
	},
	{ immediate: true },
)

async function loadEffectiveTemplate() {
	const catId = String(localValue.categoryId || '').trim()
	if (!catId) {
		templateFields.value = []
		effectiveTemplate.value = null
		return
	}
	try {
		const eff = await templateApi.effective(catId).catch(() => null)
		effectiveTemplate.value = eff
		templateFields.value = (eff?.fields ?? []).map((f) => ({
			...f,
			fieldType: normalizeFieldType(String(f.fieldType || 'string')),
		}))
		for (const f of templateFields.value) {
			if (f.defaultValue !== undefined && localValue.attributes[f.fieldKey] === undefined) {
				localValue.attributes[f.fieldKey] = f.defaultValue
			}
			if (f.fieldType === 'boolean' && localValue.attributes[f.fieldKey] === undefined) {
				localValue.attributes[f.fieldKey] = false
			}
		}
	} catch (error: any) {
		templateFields.value = []
		effectiveTemplate.value = null
		toast.add({ title: '加载生效模板失败', description: error?.message || '请稍后重试', color: 'error' })
	}
}

const fieldLabel = (field: CategoryTemplateField) => {
	const suffix = field.required ? ' *' : ''
	return `${field.fieldKey}${suffix}`
}

function normalizeFieldType(raw: string) {
	const t = String(raw || '').trim().toLowerCase()
	switch (t) {
		case 'bool':
			return 'boolean'
		case 'int':
		case 'float':
			return 'number'
		default:
			return t || 'string'
	}
}

const enumItems = (field: CategoryTemplateField) => {
	const raw = field.validationRules?.options ?? field.validationRules?.enum ?? field.validationRules?.values
	const options = Array.isArray(raw) ? raw : []
	return options.map((v: any) => ({ label: String(v), value: String(v) }))
}

function buildCategoryMap(nodes: CategoryNode[]) {
	const out: Record<string, CategoryNode> = {}
	const walk = (items: CategoryNode[]) => {
		for (const node of items) {
			out[node.id] = node
			if (node.children?.length) walk(node.children)
		}
	}
	walk(nodes)
	return out
}

function flattenTree(nodes: CategoryNode[], level = 0): Array<{ label: string; value: string }> {
	const out: Array<{ label: string; value: string }> = []
	for (const node of nodes) {
		const prefix = level > 0 ? `${'—'.repeat(Math.min(level, 6))} ` : ''
		out.push({ label: `${prefix}${node.displayName}`, value: node.id })
		if (node.children?.length) {
			out.push(...flattenTree(node.children, level + 1))
		}
	}
	return out
}

function parseCategoryPathIDs(path: string) {
	return String(path || '')
		.split('/')
		.map((p) => p.trim())
		.filter(Boolean)
}
</script>
