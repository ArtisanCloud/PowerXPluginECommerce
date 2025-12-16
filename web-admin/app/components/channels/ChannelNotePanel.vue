<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold">运营备注</h3>
        <UBadge variant="subtle" color="neutral">{{ notes?.length ?? 0 }}</UBadge>
      </div>
    </template>

    <UForm class="space-y-3" @submit.prevent="handleSubmit">
      <UFormGroup label="新增备注">
        <UTextarea v-model="body" rows="3" placeholder="记录运营决策、授权背景信息等" />
      </UFormGroup>
      <div class="flex items-center justify-between">
        <USelectMenu v-model="visibility" :options="visibilityOptions" size="xs" />
        <UButton type="submit" size="xs" color="primary" :loading="loading">保存备注</UButton>
      </div>
    </UForm>

    <ul v-if="notes?.length" class="mt-5 space-y-3">
      <li
        v-for="note in notes"
        :key="note.id"
        class="rounded-lg border border-gray-100 p-3 text-sm dark:border-gray-800"
      >
        <div class="flex items-center justify-between text-xs text-gray-400">
          <span>{{ note.authorUuid }}</span>
          <span>{{ formatDate(note.createdAt) }}</span>
        </div>
        <p class="mt-2 text-gray-800 dark:text-gray-100">{{ note.body }}</p>
      </li>
    </ul>
    <div v-else class="mt-4 text-center text-sm text-gray-500">暂无备注</div>
  </UCard>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ChannelNote } from '~/types/channels'

const props = defineProps<{
  notes?: ChannelNote[]
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'create', payload: { body: string; visibility: string }): void
}>()

const body = ref('')
const visibility = ref('team')
const visibilityOptions = [
  { label: '团队可见', value: 'team' },
  { label: '管理员', value: 'admin' },
]

const handleSubmit = () => {
  if (!body.value.trim()) return
  emit('create', { body: body.value, visibility: visibility.value })
  body.value = ''
}

const formatDate = (value: string) => new Date(value).toLocaleString('zh-CN')
</script>
