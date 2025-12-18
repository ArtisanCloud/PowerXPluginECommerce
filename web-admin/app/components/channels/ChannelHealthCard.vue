<template>
  <UCard>
    <div class="flex items-center justify-between">
      <div>
        <p class="text-sm text-gray-500 dark:text-gray-400">健康度 / Health Score</p>
        <p class="text-3xl font-semibold text-gray-900 dark:text-white" data-testid="health-score">
          {{ displayScore }}
        </p>
      </div>
      <UBadge :color="badgeColor" size="lg">{{ badgeLabel }}</UBadge>
    </div>
    <div v-if="labels?.length" class="mt-4 flex flex-wrap gap-2">
      <UBadge
        v-for="label in labels"
        :key="label"
        variant="outline"
        color="warning"
        class="uppercase tracking-wide"
        data-testid="health-label"
      >
        {{ label }}
      </UBadge>
    </div>
    <p v-else class="mt-3 text-sm text-gray-500 dark:text-gray-400">
      暂无告警标签，渠道运行健康。
    </p>
  </UCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{
  score?: number
  labels?: string[]
}>()

const displayScore = computed(() => props.score ?? 0)

const badgeColor = computed(() => {
  if (displayScore.value >= 80) return 'success'
  if (displayScore.value >= 60) return 'info'
  if (displayScore.value >= 40) return 'warning'
  return 'error'
})

const badgeLabel = computed(() => {
  if (displayScore.value >= 80) return '优秀'
  if (displayScore.value >= 60) return '良好'
  if (displayScore.value >= 40) return '关注'
  return '告警'
})
</script>
