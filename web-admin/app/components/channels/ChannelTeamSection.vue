<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-lg font-semibold">策略与团队</h3>
          <p class="text-sm text-gray-500">负责人、审批人及当前策略摘要</p>
        </div>
        <UButton
          v-if="canEdit"
          size="xs"
          icon="i-heroicons-pencil-square"
          @click="emit('edit')"
        >
          编辑
        </UButton>
      </div>
    </template>
    <div class="grid gap-6 md:grid-cols-2">
      <section class="space-y-3">
        <h4 class="text-sm font-semibold text-gray-600">团队</h4>
        <dl class="space-y-2">
          <div>
            <dt class="text-xs text-gray-500">负责人</dt>
            <dd class="font-medium text-gray-900 dark:text-white">
              {{ team?.ownerUuid ?? '未设置' }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500">审批人</dt>
            <dd class="font-medium text-gray-900 dark:text-white">
              {{ team?.approverUuid ?? '未设置' }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500">运维成员</dt>
            <dd class="flex flex-wrap gap-2">
              <UBadge
                v-for="member in teamOperators"
                :key="member"
                variant="subtle"
                color="primary"
              >
                {{ member }}
              </UBadge>
              <span v-if="!teamOperators.length" class="text-gray-400">未配置</span>
            </dd>
          </div>
        </dl>
      </section>

      <section class="space-y-3">
        <h4 class="text-sm font-semibold text-gray-600">策略摘要</h4>
        <dl class="space-y-2">
          <div v-for="field in strategySummary" :key="field.label">
            <dt class="text-xs text-gray-500">{{ field.label }}</dt>
            <dd class="font-medium text-gray-900 dark:text-white">
              {{ field.value }}
            </dd>
          </div>
        </dl>
      </section>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ChannelStrategyConfig, ChannelTeamConfig } from '~/types/channels'

const props = defineProps<{
  team?: ChannelTeamConfig
  strategy?: ChannelStrategyConfig
  canEdit?: boolean
}>()

const emit = defineEmits<{
  (e: 'edit'): void
}>()

const teamOperators = computed(() => props.team?.operators ?? [])

const formatPercent = (value?: number) => {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return '--'
  }
  return `${value.toFixed(1)}%`
}

const formatDate = (value?: string) => {
  if (!value) {
    return '--'
  }
  try {
    return new Date(value).toLocaleString('zh-CN')
  } catch {
    return value
  }
}

const strategySummary = computed(() => {
  const strategy = props.strategy ?? {}
  return [
    { label: 'Pricebook ID', value: strategy.pricebookId || '--' },
    { label: '库存策略 ID', value: strategy.inventoryStrategyId || '--' },
    { label: '物流策略 ID', value: strategy.logisticsStrategyId || '--' },
    { label: '客服 SLA ID', value: strategy.csSlaId || '--' },
    { label: '手续费', value: formatPercent(strategy.feeRate) },
    { label: '结算周期', value: strategy.settlementCycle || '--' },
    { label: '付款条款', value: strategy.paymentTerms || '--' },
    { label: '生效时间', value: formatDate(strategy.effectiveAt) },
  ]
})
</script>
