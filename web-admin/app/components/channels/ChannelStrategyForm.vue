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
      <UTextarea v-model="localForm.strategy.notes" rows="3" placeholder="同步渠道合同或特殊条款" />
    </UFormField>

    <div class="space-y-4 rounded-lg border border-gray-200 p-4 dark:border-gray-800">
      <div class="flex items-center justify-between">
        <div>
          <h4 class="text-base font-semibold">团队配置</h4>
          <p class="text-sm text-gray-500">负责人、审批人与运维成员</p>
        </div>
      </div>
      <div class="grid gap-4 md:grid-cols-2">
        <UFormField label="负责人 Owner UUID" required>
          <UInput v-model="localForm.team.ownerUuid" placeholder="user:owner" />
        </UFormField>
        <UFormField label="审批人 Approver UUID">
          <UInput v-model="localForm.team.approverUuid" placeholder="user:approver" />
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
import { reactive, ref, watch } from 'vue'
import type { ChannelStrategyUpdatePayload } from '~/types/channels'
import { createEmptyStrategyPayload } from '~/types/channels'

const props = defineProps<{
  modelValue: ChannelStrategyUpdatePayload
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ChannelStrategyUpdatePayload): void
  (e: 'submit', value: ChannelStrategyUpdatePayload): void
  (e: 'cancel'): void
}>()

const settlementOptions = [
  { label: '日结 (DD)', value: 'dd' },
  { label: 'Net 7', value: 'net7' },
  { label: 'Net 14', value: 'net14' },
  { label: 'Net 30', value: 'net30' },
  { label: 'Net 45', value: 'net45' },
  { label: 'Net 60', value: 'net60' },
]

const localForm = reactive<ChannelStrategyUpdatePayload>(createEmptyStrategyPayload())
const effectiveInput = ref('')
const operatorsInput = ref('')

const syncLocal = (payload?: ChannelStrategyUpdatePayload) => {
  const next = payload ?? createEmptyStrategyPayload()
  Object.assign(localForm.strategy, next.strategy ?? {})
  Object.assign(localForm.team, next.team ?? {})
  localForm.team.operators = [...(next.team.operators ?? [])]
  effectiveInput.value = next.strategy.effectiveAt
    ? next.strategy.effectiveAt.slice(0, 16)
    : ''
  operatorsInput.value = (next.team.operators ?? []).join(', ')
}

watch(
  () => props.modelValue,
  (val) => syncLocal(val),
  { immediate: true, deep: true },
)

watch(
  () => localForm,
  () => {
    emit('update:modelValue', clonePayload())
  },
  { deep: true },
)

watch(effectiveInput, (val) => {
  localForm.strategy.effectiveAt = val ? new Date(val).toISOString() : ''
})

watch(operatorsInput, (val) => {
  localForm.team.operators = val
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
})

const clonePayload = (): ChannelStrategyUpdatePayload => ({
  strategy: { ...localForm.strategy },
  team: {
    ...localForm.team,
    operators: [...(localForm.team.operators ?? [])],
  },
})

const handleSubmit = () => {
  emit('submit', clonePayload())
}
</script>
