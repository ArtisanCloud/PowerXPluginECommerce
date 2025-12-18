<template>
  <UForm :state="localForm" class="space-y-6" @submit.prevent="handleSubmit">
    <div class="grid gap-4 md:grid-cols-2">
      <UFormField label="Pricebook ID">
        <UInput v-model="localForm.strategy.pricebookId" placeholder="pb_default_channel" />
      </UFormField>
      <UFormField label="库存策略 ID">
        <UInput v-model="localForm.strategy.inventoryStrategyId" placeholder="inventory_strategy_main" />
      </UFormField>
      <UFormField label="物流策略 ID">
        <UInput v-model="localForm.strategy.logisticsStrategyId" placeholder="logistics_standard" />
      </UFormField>
      <UFormField label="客服 SLA ID">
        <UInput v-model="localForm.strategy.csSlaId" placeholder="cs_sla_priority" />
      </UFormField>
      <UFormField label="手续费 (%)">
        <UInput v-model.number="localForm.strategy.feeRate" type="number" min="0" max="100" step="0.1" />
      </UFormField>
      <UFormField label="结算周期">
        <USelectMenu
          v-model="localForm.strategy.settlementCycle"
          :options="settlementOptions"
          value-attribute="value"
          option-attribute="label"
          placeholder="选择结算周期"
        />
      </UFormField>
      <UFormField label="付款条款">
        <UInput v-model="localForm.strategy.paymentTerms" placeholder="预付 30%，货到付清" />
      </UFormField>
      <UFormField label="生效时间">
        <UInput v-model="effectiveInput" type="datetime-local" />
      </UFormField>
    </div>

    <UFormField label="策略备注">
      <UTextarea v-model="localForm.strategy.notes" :rows="3" placeholder="同步渠道合同或特殊条款" />
    </UFormField>

    <div class="space-y-4 rounded-lg border border-gray-200 p-4 dark:border-gray-800">
      <div class="flex items-center justify-between">
        <div>
          <h4 class="text-base font-semibold">团队配置</h4>
          <p class="text-sm text-gray-500">负责人、审批人与运维成员</p>
        </div>
      </div>
      <div class="grid gap-4 md:grid-cols-2">
        <UFormField label="负责人 / Owner" :description="ownerDescription" required>
          <USelectMenu
            v-model="localForm.team.ownerUuid"
            :items="ownerSelectOptions"
            value-key="value"
            label-key="label"
            searchable
            v-model:search-term="ownerSearchTerm"
            :loading="ownerLoading"
            :ignore-filter="true"
            class="w-full"
            placeholder="选择负责人"
            @keydown.enter.prevent.stop="confirmOwnerSearch"
          >
            <template #option="{ option }">
              <div class="flex flex-col">
                <span class="font-medium text-sm text-gray-900 dark:text-white">{{ option.label }}</span>
                <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
              </div>
            </template>
            <template #empty>
              <div class="px-3 py-2 text-sm text-gray-500">
                <p>未找到匹配负责人，可输入关键词继续搜索。</p>
                <p v-if="ownerDescription" class="mt-1">{{ ownerDescription }}</p>
              </div>
            </template>
          </USelectMenu>
        </UFormField>
        <UFormField label="审批人 / Approver" :description="approverDescription">
          <USelectMenu
            v-model="localForm.team.approverUuid"
            :items="ownerSelectOptions"
            value-key="value"
            label-key="label"
            searchable
            v-model:search-term="approverSearchTerm"
            :loading="ownerLoading"
            :ignore-filter="true"
            class="w-full"
            clearable
            placeholder="可选，选择审批人"
            @keydown.enter.prevent.stop="confirmApproverSearch"
          >
            <template #option="{ option }">
              <div class="flex flex-col">
                <span class="font-medium text-sm text-gray-900 dark:text-white">{{ option.label }}</span>
                <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
              </div>
            </template>
          </USelectMenu>
        </UFormField>
        <UFormField label="运维成员 (逗号分隔)">
          <UInput
            v-model="operatorsInput"
            placeholder="ops-a,ops-b"
            help="用于限定渠道操作权限，可填写 IAM 用户 ID 或外部账号"
          />
        </UFormField>
      </div>
    </div>

    <div class="flex items-center justify-end gap-3">
      <UButton variant="ghost" color="neutral" @click="emit('cancel')">取消</UButton>
      <UButton color="primary" type="submit" :loading="loading">保存策略</UButton>
    </div>
  </UForm>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useVModel } from '@vueuse/core'
import type { ChannelStrategyUpdatePayload, ChannelOwnerOption } from '~/types/channels'
import { createEmptyStrategyPayload } from '~/types/channels'

const props = defineProps<{
  modelValue: ChannelStrategyUpdatePayload
  loading?: boolean
  ownerOptions?: ChannelOwnerOption[]
  ownerLoading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ChannelStrategyUpdatePayload): void
  (e: 'submit', value: ChannelStrategyUpdatePayload): void
  (e: 'cancel'): void
  (e: 'search-owner', keyword: string, force?: boolean): void
}>()

const settlementOptions = [
  { label: '日结 (DD)', value: 'dd' },
  { label: 'Net 7', value: 'net7' },
  { label: 'Net 14', value: 'net14' },
  { label: 'Net 30', value: 'net30' },
  { label: 'Net 45', value: 'net45' },
  { label: 'Net 60', value: 'net60' },
]

const localForm = useVModel(props, 'modelValue', emit, {
  passive: true,
  defaultValue: createEmptyStrategyPayload(),
  deep: true,
})

const effectiveInput = computed({
  get: () => {
    const raw = localForm.value.strategy.effectiveAt
    return raw ? raw.slice(0, 16) : ''
  },
  set: (val: string) => {
    localForm.value.strategy.effectiveAt = val ? new Date(val).toISOString() : ''
  },
})

const operatorsInput = computed({
  get: () => (localForm.value.team.operators ?? []).join(', '),
  set: (val: string) => {
    localForm.value.team.operators = val
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)
  },
})
const ownerLoading = computed(() => props.ownerLoading ?? false)

type SelectOption = { label: string; value: string; description?: string }

const mapOwnerOption = (option: ChannelOwnerOption): SelectOption => {
  const label = option.displayName || option.username || `user:${option.id}`
  const value = option.username || option.displayName || `user:${option.id}`
  return {
    label,
    value,
    description: option.email,
  }
}

const fallbackOwnerOptions: SelectOption[] = [
  { label: 'admin', value: 'admin', description: 'admin@example.com' },
  { label: 'ops-oncall', value: 'ops-oncall', description: 'ops@example.com' },
]

const ownerOptionsSource = computed<SelectOption[]>(() => {
  if (props.ownerOptions?.length) {
    return props.ownerOptions.map(mapOwnerOption)
  }
  return fallbackOwnerOptions
})

const ownerSelectOptions = computed<SelectOption[]>(() => {
  const base = [...ownerOptionsSource.value]
  const ensureValue = (value?: string) => {
    if (value && !base.some((option) => option.value === value)) {
      base.push({ label: value, value })
    }
  }
  ensureValue(localForm.value.team.ownerUuid)
  ensureValue(localForm.value.team.approverUuid)
  return base
})

const ownerDescription = computed(() => {
  if (!ownerSelectOptions.value.length) {
    return '暂无负责人目录，请联系管理员同步 IAM 成员'
  }
  return `示例：${ownerSelectOptions.value
    .slice(0, 3)
    .map((item) => item.label)
    .join(' / ')}`
})

const approverDescription = computed(() => '可选：可输入姓名、用户名或邮箱前缀搜索')

const ownerSearchTerm = ref('')
const approverSearchTerm = ref('')
let ownerSearchTimer: ReturnType<typeof setTimeout> | null = null
let lastOwnerQuery = ''

const emitOwnerSearch = (term: string, force = false) => {
  const normalized = term.trim()
  if (!force && normalized === lastOwnerQuery) {
    return
  }
  lastOwnerQuery = normalized
  emit('search-owner', normalized, force)
}

const scheduleOwnerSearch = (term: string) => {
  if (ownerSearchTimer) {
    clearTimeout(ownerSearchTimer)
  }
  ownerSearchTimer = setTimeout(() => emitOwnerSearch(term), 400)
}

watch(
  ownerSearchTerm,
  (term) => scheduleOwnerSearch(term),
  { immediate: false },
)

watch(
  approverSearchTerm,
  (term) => scheduleOwnerSearch(term),
  { immediate: false },
)

const confirmOwnerSearch = () => {
  emitOwnerSearch(ownerSearchTerm.value, true)
}

const confirmApproverSearch = () => {
  emitOwnerSearch(approverSearchTerm.value, true)
}

onBeforeUnmount(() => {
  if (ownerSearchTimer) {
    clearTimeout(ownerSearchTimer)
  }
})

watch(
  () => localForm.value.team.ownerUuid as unknown,
  (val) => {
    if (val && typeof val === 'object') {
      const option = val as SelectOption
      localForm.value.team.ownerUuid = option.value
    }
  },
)

watch(
  () => localForm.value.team.approverUuid as unknown,
  (val) => {
    if (val && typeof val === 'object') {
      const option = val as SelectOption
      localForm.value.team.approverUuid = option.value
    }
  },
)


const clonePayload = (): ChannelStrategyUpdatePayload => ({
  strategy: { ...localForm.value.strategy },
  team: {
    ...localForm.value.team,
    operators: [...(localForm.value.team.operators ?? [])],
  },
})

const handleSubmit = () => {
  emit('submit', clonePayload())
}
</script>
