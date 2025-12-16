<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold">任务关联</h3>
        <UButton size="xs" color="primary" variant="soft" @click="showForm = !showForm">
          {{ showForm ? '收起' : '新增任务' }}
        </UButton>
      </div>
    </template>

    <div v-if="showForm" class="mb-4 rounded-lg border border-dashed border-gray-200 p-4">
      <UForm :state="form" class="space-y-3" @submit.prevent="handleSubmit">
        <UFormGroup label="任务 ID" required>
          <UInput v-model="form.taskId" placeholder="TASK-1001" />
        </UFormGroup>
        <UFormGroup label="来源">
          <USelectMenu v-model="form.taskSource" :options="sources" placeholder="task_center" />
        </UFormGroup>
        <UFormGroup label="备注">
          <UTextarea v-model="form.note" placeholder="补充说明" />
        </UFormGroup>
        <div class="flex justify-end gap-2">
          <UButton variant="ghost" color="neutral" @click="resetForm">取消</UButton>
          <UButton color="primary" type="submit" :loading="loading">关联任务</UButton>
        </div>
      </UForm>
    </div>

    <div v-if="loading" class="py-4 text-center text-sm text-gray-500">加载中...</div>
    <ul v-else-if="tasks?.length" class="space-y-3">
      <li
        v-for="task in tasks"
        :key="task.id"
        class="rounded-lg border border-gray-100 p-4 dark:border-gray-800"
      >
        <div class="flex items-center justify-between">
          <div>
            <p class="font-semibold text-gray-900 dark:text-white">{{ task.taskId }}</p>
            <p class="text-xs uppercase text-gray-400">{{ task.taskSource }}</p>
            <p class="mt-1 text-sm text-gray-500">{{ task.note || '—' }}</p>
          </div>
          <div class="flex items-center gap-2">
            <UBadge :color="task.status === 'done' ? 'success' : 'info'" variant="subtle">
              {{ statusLabel(task.status) }}
            </UBadge>
            <UButton
              size="xs"
              variant="ghost"
              color="success"
              @click="emit('update', { id: task.id, status: 'done' })"
            >
              完成
            </UButton>
            <UButton
              size="xs"
              variant="ghost"
              color="neutral"
              @click="emit('remove', task.id)"
            >
              解除
            </UButton>
          </div>
        </div>
        <p class="mt-2 text-xs text-gray-400">关联人：{{ task.linkedBy }} · {{ formatDate(task.linkedAt) }}</p>
      </li>
    </ul>
    <div v-else class="py-4 text-center text-sm text-gray-500">尚未关联任务</div>
  </UCard>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ChannelTaskLink } from '~/types/channels'

const props = defineProps<{
  tasks?: ChannelTaskLink[]
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'link', payload: { taskId: string; taskSource?: string; note?: string }): void
  (e: 'update', payload: { id: string; status: string; note?: string }): void
  (e: 'remove', id: string): void
}>()

const showForm = ref(false)
const form = ref<{ taskId: string; taskSource?: string; note?: string }>({
  taskId: '',
  taskSource: 'task_center',
  note: '',
})

const sources = [
  { label: '任务中心', value: 'task_center' },
  { label: '手动', value: 'manual' },
]

const handleSubmit = () => {
  if (!form.value.taskId) return
  emit('link', { ...form.value })
  resetForm()
}

const resetForm = () => {
  form.value = { taskId: '', taskSource: 'task_center', note: '' }
  showForm.value = false
}

const statusLabel = (status: string) => {
  switch (status) {
    case 'done':
      return '已完成'
    case 'in_progress':
      return '处理中'
    default:
      return '未开始'
  }
}

const formatDate = (value?: string) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}
</script>
