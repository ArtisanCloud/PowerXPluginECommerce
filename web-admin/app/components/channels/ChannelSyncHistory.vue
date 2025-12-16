<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-lg font-semibold">同步历史</h3>
          <p class="text-sm text-gray-500">最近 {{ history?.length ?? 0 }} 条记录</p>
        </div>
        <UButton
          size="xs"
          color="primary"
          :loading="loading"
          @click="emit('trigger')"
        >
          手动同步
        </UButton>
      </div>
    </template>
    <div v-if="history?.length">
      <UTable :columns="columns" :data="tableData" />
    </div>
    <div v-else class="py-4 text-center text-sm text-gray-500">暂无同步记录</div>
  </UCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import type { ChannelSyncHistoryItem } from '~/types/channels'

const props = defineProps<{
  history?: ChannelSyncHistoryItem[]
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'trigger'): void
}>()

const columns: TableColumn[] = [
  { accessorKey: 'createdAt', header: '触发时间' },
  { accessorKey: 'triggerType', header: '方式' },
  { accessorKey: 'triggeredBy', header: '触发人' },
  { accessorKey: 'durationMs', header: '耗时(ms)' },
  { accessorKey: 'result', header: '结果' },
]

const tableData = computed(() =>
  (props.history ?? []).map((item) => ({
    createdAt: formatDate(item.createdAt),
    triggerType: item.triggerType,
    triggeredBy: item.triggeredBy || '系统',
    durationMs: item.durationMs ?? 0,
    result: item.result,
  })),
)

const formatDate = (value: string) => new Date(value).toLocaleString('zh-CN')
</script>
