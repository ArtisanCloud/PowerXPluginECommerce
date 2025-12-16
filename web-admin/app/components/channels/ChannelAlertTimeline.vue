<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold">告警</h3>
        <UBadge color="warning" variant="subtle">{{ alerts?.length ?? 0 }} 个</UBadge>
      </div>
    </template>
    <div v-if="loading" class="py-6 text-center text-sm text-gray-500">加载中...</div>
    <ul v-else-if="alerts?.length" class="space-y-4">
      <li
        v-for="alert in alerts"
        :key="alert.id"
        class="rounded-lg border border-gray-100 p-4 dark:border-gray-800"
      >
        <div class="flex items-center justify-between">
          <div>
            <p class="font-semibold text-gray-900 dark:text-white">
              {{ alert.title }}
            </p>
            <p class="text-sm text-gray-500">{{ alert.description }}</p>
          </div>
          <UBadge :color="alertSeverity(alert.severity)" variant="subtle">{{ alert.status }}</UBadge>
        </div>
        <div class="mt-3 flex flex-wrap gap-2 text-sm text-gray-500">
          <span>触发时间：{{ formatDate(alert.triggeredAt) }}</span>
          <span v-if="alert.resolvedAt"> · 已解决：{{ formatDate(alert.resolvedAt) }}</span>
        </div>
        <div class="mt-3 flex flex-wrap gap-2">
          <UButton
            size="xs"
            variant="outline"
            @click="emit('update', { alertId: alert.id, status: 'acknowledged' })"
          >
            标记确认
          </UButton>
          <UButton
            size="xs"
            color="success"
            variant="outline"
            @click="emit('update', { alertId: alert.id, status: 'resolved' })"
          >
            标记解决
          </UButton>
        </div>
      </li>
    </ul>
    <div v-else class="py-6 text-center text-sm text-gray-500">暂无告警</div>
  </UCard>
</template>

<script setup lang="ts">
import type { ChannelAlert } from '~/types/channels'

const props = defineProps<{
  alerts?: ChannelAlert[]
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update', payload: { alertId: string; status: string }): void
}>()

const alertSeverity = (severity?: string) => {
  if (severity === 'critical') return 'error'
  if (severity === 'warning') return 'warning'
  return 'info'
}

const formatDate = (value?: string) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}
</script>
