<template>
	<div class="space-y-6">
		<div class="flex flex-wrap items-end gap-4">
			<UFormField label="默认语言" :description="errors.defaultLocale" class="w-full md:w-1/3">
				<template #default="{ id }">
					<USelect :id="id" v-model="localValue.defaultLocale" :items="localeOptions" />
				</template>
			</UFormField>
			<div class="flex flex-wrap items-center gap-3">
				<USelect v-model="pendingLocale" :items="availableLocaleOptions" class="w-44" />
				<UButton color="primary" variant="ghost" :disabled="!pendingLocale" @click="append">添加语言</UButton>
				<UButton color="gray" variant="ghost" :disabled="localValue.locales.length <= 1" @click="copyDefaultToAll">
					复制默认语言内容到其他
				</UButton>
			</div>
		</div>
		<div class="space-y-4">
			<div
				v-for="(entry, idx) in localValue.locales"
				:key="entry.locale"
				class="space-y-3 rounded border p-4"
			>
				<div class="flex flex-wrap items-center justify-between gap-2">
					<div class="flex items-center gap-2">
						<strong>{{ entry.locale }}</strong>
						<span v-if="entry.locale === localValue.defaultLocale" class="text-xs text-primary">默认</span>
					</div>
					<div class="flex gap-2">
						<UButton
							size="xs"
							variant="ghost"
							color="primary"
							:disabled="entry.locale === localValue.defaultLocale"
							@click="setAsDefault(entry.locale)"
						>
							设为默认
						</UButton>
						<UButton size="xs" variant="ghost" color="gray" @click="copyFromDefault(idx)">复制默认</UButton>
						<UButton
							size="xs"
							variant="ghost"
							icon="i-heroicons-trash"
							color="gray"
							@click="remove(idx)"
							:disabled="localValue.locales.length === 1"
						>
							移除
						</UButton>
					</div>
				</div>
				<UFormField label="标题" :description="errors.entries[idx]?.title">
					<template #default="{ id }">
						<UInput :id="id" v-model.trim="entry.title" :data-testid="`locale-title-${entry.locale}`" placeholder="请输入标题" />
					</template>
				</UFormField>
                <UFormField label="描述" :description="errors.entries[idx]?.description">
                    <template #default="{ id }">
                        <UTextarea :id="id" v-model="entry.description" :rows="3" placeholder="请输入描述" />
                    </template>
                </UFormField>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useToast } from '#imports'
import { computed, reactive, ref, watch } from 'vue'
import { useVModel } from '@vueuse/core'

type LocaleEntry = {
	locale: string
	title: string
	description?: string
}

interface LocaleModel {
	defaultLocale: string
	locales: LocaleEntry[]
}

const defaultLocales = [
	{ label: '简体中文 (zh-CN)', value: 'zh-CN' },
	{ label: '繁体中文 (zh-TW)', value: 'zh-TW' },
	{ label: '英文 (en-US)', value: 'en-US' },
	{ label: '日语 (ja-JP)', value: 'ja-JP' },
]

const props = defineProps<{ modelValue: LocaleModel }>()
const emit = defineEmits<{
	'update:modelValue': [LocaleModel]
	validation: [boolean]
}>()

const toast = useToast()
const localeOptions = defaultLocales
const localValue = useVModel(props, 'modelValue', emit, {
	passive: true,
	deep: true,
	defaultValue: {
		defaultLocale: 'zh-CN',
		locales: [{ locale: 'zh-CN', title: '', description: '' }],
	},
})

const pendingLocale = ref<string>('')
const errors = reactive<{ defaultLocale: string; entries: Record<number, { title?: string; description?: string }> }>({
	defaultLocale: '',
	entries: {},
})

const availableLocaleOptions = computed(() =>
	localeOptions.filter((option) => !localValue.value.locales?.some((entry) => entry.locale === option.value))
)

watch(
	() => [localValue.value.defaultLocale, localValue.value.locales?.length],
	() => {
		if (!localValue.value.defaultLocale) {
			localValue.value.defaultLocale = 'zh-CN'
		}
		if (!Array.isArray(localValue.value.locales) || !localValue.value.locales.length) {
			localValue.value.locales = [{ locale: localValue.value.defaultLocale, title: '', description: '' }]
		}
		const exists = localValue.value.locales.some((entry) => entry.locale === localValue.value.defaultLocale)
		if (!exists) {
			localValue.value.locales.unshift({
				locale: localValue.value.defaultLocale,
				title: '',
				description: '',
			})
		}
		pendingLocale.value = firstAvailableLocale()
		validate()
	},
	{ immediate: true }
)

watch(
	() => localValue.value.locales,
	() => {
		validate()
	},
	{ deep: true }
)

function append() {
	const locale = pendingLocale.value || firstAvailableLocale()
	if (!locale || localValue.value.locales.some((entry) => entry.locale === locale)) {
		return
	}
	localValue.value.locales.push({ locale, title: '', description: '' })
	pendingLocale.value = firstAvailableLocale()
}

function remove(index: number) {
	if (localValue.value.locales.length === 1) {
		return
	}
	const removed = localValue.value.locales[index]
	localValue.value.locales.splice(index, 1)
	if (removed?.locale === localValue.value.defaultLocale) {
		localValue.value.defaultLocale = localValue.value.locales[0]?.locale || 'zh-CN'
	}
	pendingLocale.value = firstAvailableLocale()
}

function setAsDefault(locale: string) {
	localValue.value.defaultLocale = locale
}

function firstAvailableLocale() {
	return availableLocaleOptions.value[0]?.value || ''
}

function copyFromDefault(targetIndex: number) {
	const defaultEntry = localValue.value.locales.find((entry) => entry.locale === localValue.value.defaultLocale)
	if (!defaultEntry) {
		toast.add({ title: '未设置默认语言', color: 'red' })
		return
	}
	const target = localValue.value.locales[targetIndex]
	if (!target || target.locale === defaultEntry.locale) {
		return
	}
	target.title = defaultEntry.title
	target.description = defaultEntry.description
	toast.add({ title: `已复制 ${defaultEntry.locale} 内容` })
}

function copyDefaultToAll() {
	const defaultEntry = localValue.value.locales.find((entry) => entry.locale === localValue.value.defaultLocale)
	if (!defaultEntry) {
		toast.add({ title: '未设置默认语言', color: 'red' })
		return
	}
	localValue.value.locales.forEach((entry) => {
		if (entry.locale === defaultEntry.locale) return
		entry.title = defaultEntry.title
		entry.description = defaultEntry.description
	})
	toast.add({ title: '已复制默认语言内容到全部语言' })
}

function validate() {
	errors.defaultLocale = ''
	errors.entries = {}
	let valid = true
	if (!localValue.value.defaultLocale) {
		errors.defaultLocale = '请选择默认语言'
		valid = false
	}
	const defaultEntry = localValue.value.locales.find((entry) => entry.locale === localValue.value.defaultLocale)
	if (!defaultEntry) {
		errors.defaultLocale = '请添加默认语言内容'
		valid = false
	} else if (!defaultEntry.title) {
		errors.defaultLocale = '默认语言标题必填'
		errors.entries[getIndexByLocale(defaultEntry.locale)] = { title: '必填' }
		valid = false
	}
	localValue.value.locales.forEach((entry, idx) => {
		if (entry.locale === localValue.value.defaultLocale && !entry.title) {
			errors.entries[idx] = { title: '默认语言标题必填' }
		}
	})
	emit('validation', valid)
	return valid
}

function getIndexByLocale(locale: string) {
	return localValue.value.locales.findIndex((entry) => entry.locale === locale)
}

defineExpose({ validate })
</script>
