<template>
	<UCard>
		<template #header>
			<div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white">审计时间轴</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400">记录版本、审批与系统事件。</p>
			</div>
		</template>
		<div v-if="!events.length" class="text-sm text-gray-500 dark:text-gray-400">暂时没有审计事件。</div>
		<ul v-else class="space-y-4">
			<li v-for="event in events" :key="event.id" class="flex gap-3">
				<div class="mt-1 h-2 w-2 rounded-full bg-primary-500 dark:bg-primary-400"></div>
				<div>
					<p class="font-medium text-gray-900 dark:text-gray-100">{{ event.title }}</p>
					<p v-if="event.description" class="text-sm text-gray-500 dark:text-gray-400">{{ event.description }}</p>
					<p class="text-xs text-gray-400 dark:text-gray-500">{{ formatTimestamp(event.timestamp) }}</p>
				</div>
			</li>
		</ul>
	</UCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface AuditEvent {
	id: string
	title: string
	description?: string
	timestamp?: string
}

const props = defineProps<{
	events: AuditEvent[]
}>()

const events = computed(() => props.events ?? [])

const formatTimestamp = (value?: string) => {
	if (!value) return '—'
	return new Date(value).toLocaleString()
}
</script>
