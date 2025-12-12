<template>
	<div class="grid gap-4 lg:grid-cols-2">
		<UCard>
			<template #header>
				<div class="flex items-center justify-between">
					<div>
						<h3 class="text-lg font-semibold">版本时间线</h3>
						<p class="text-sm text-gray-500">查看历史版本并回滚。</p>
					</div>
					<UBadge color="gray" variant="soft">{{ versions.length }} 个版本</UBadge>
				</div>
			</template>
			<div v-if="!versions.length" class="text-sm text-gray-500">尚无历史版本。</div>
			<ul v-else class="space-y-3">
				<li
					v-for="version in versions"
					:key="version.id"
					:class="[
						'flex items-center justify-between rounded border px-4 py-3',
						isSelected(version.id) ? 'border-primary-500 bg-primary-50' : 'border-gray-200',
					]"
				>
					<div>
						<p class="font-medium">
							版本 #{{ version.versionNumber }}
							<UBadge class="ml-2" size="xs">{{ version.status }}</UBadge>
						</p>
						<p class="text-xs text-gray-500">
							提交人：{{ version.submittedBy || '—' }} · 提交时间：{{ formatTimestamp(version.submittedAt || version.createdAt) }}
						</p>
					</div>
					<div class="flex gap-2">
						<UButton size="xs" color="primary" variant="soft" @click="emit('select', version.id)"> 查看 </UButton>
						<UButton size="xs" color="gray" variant="ghost" @click="emit('rollback', version.id)"> 回滚 </UButton>
					</div>
				</li>
			</ul>
		</UCard>
		<UCard>
			<template #header>
				<div class="flex items-center justify-between">
					<div>
						<h3 class="text-lg font-semibold">字段差异</h3>
						<p class="text-sm text-gray-500">当前选中版本与已发布版本的差异。</p>
					</div>
					<UBadge v-if="detail" color="primary" variant="soft">#{{ detail.versionNumber }}</UBadge>
				</div>
			</template>
			<div v-if="loading" class="space-y-3">
				<USkeleton class="h-6 w-full" />
				<USkeleton class="h-6 w-11/12" />
				<USkeleton class="h-6 w-10/12" />
			</div>
			<div v-else-if="detail && detail.diff.length" class="overflow-x-auto">
				<table class="min-w-full text-sm">
					<thead>
						<tr class="text-left text-gray-500">
							<th class="pb-2 pr-4">字段</th>
							<th class="pb-2 pr-4">原值</th>
							<th class="pb-2">变更值</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="entry in detail.diff" :key="entry.field" class="border-t text-sm">
							<td class="py-2 pr-4 font-medium text-gray-700">{{ entry.field }}</td>
							<td class="py-2 pr-4 text-gray-500">{{ formatValue(entry.before) }}</td>
							<td class="py-2 text-gray-900">{{ formatValue(entry.after) }}</td>
						</tr>
					</tbody>
				</table>
			</div>
			<div v-else-if="detail && !detail.diff.length" class="text-sm text-gray-500">该版本与线上版本一致。</div>
			<div v-else class="text-sm text-gray-500">选择一个版本以查看差异。</div>
		</UCard>
	</div>
</template>

<script setup lang="ts">
import type { SpuVersionDetail, SpuVersionSummary } from '~/composables/api/useSpu'

const props = defineProps<{
	versions: SpuVersionSummary[]
	selectedId?: string
	detail: SpuVersionDetail | null
	loading?: boolean
}>()

const emit = defineEmits<{
	(e: 'select', id: string): void
	(e: 'rollback', id: string): void
}>()

const isSelected = (id: string) => props.selectedId === id

const formatTimestamp = (value?: string) => {
	if (!value) return '—'
	return new Date(value).toLocaleString()
}

const formatValue = (value: any) => {
	if (value === null || value === undefined || value === '') {
		return '—'
	}
	if (typeof value === 'object') {
		try {
			return JSON.stringify(value)
		} catch {
			return '[object]'
		}
	}
	return String(value)
}
</script>
