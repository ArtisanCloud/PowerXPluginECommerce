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
const cloneEntries = (entries: LocaleEntry[] = []) => entries.map((entry) => ({ ...entry }))

const localValue = reactive<LocaleModel>({
	defaultLocale: 'zh-CN',
	locales: [{ locale: 'zh-CN', title: '', description: '' }],
})

const pendingLocale = ref<string>('')
const errors = reactive<{ defaultLocale: string; entries: Record<number, { title?: string; description?: string }> }>({
	defaultLocale: '',
	entries: {},
})

const availableLocaleOptions = computed(() =>
	localeOptions.filter((option) => !localValue.locales.some((entry) => entry.locale === option.value))
)

watch(
	() => props.modelValue,
	(val) => {
		if (!val) return
		localValue.defaultLocale = val.defaultLocale || 'zh-CN'
		localValue.locales = val.locales?.length ? cloneEntries(val.locales) : [{ locale: 'zh-CN', title: '', description: '' }]
		if (!localValue.locales.find((entry) => entry.locale === localValue.defaultLocale)) {
			localValue.locales.unshift({ locale: localValue.defaultLocale, title: '', description: '' })
		}
		pendingLocale.value = firstAvailableLocale()
		validate()
	},
	{ immediate: true }
)

watch(
	localValue,
	() => {
		emit('update:modelValue', {
			defaultLocale: localValue.defaultLocale,
			locales: cloneEntries(localValue.locales),
		})
		validate()
	},
	{ deep: true }
)

function append() {
	const locale = pendingLocale.value || firstAvailableLocale()
	if (!locale || localValue.locales.some((entry) => entry.locale === locale)) {
		return
	}
	localValue.locales.push({ locale, title: '', description: '' })
	pendingLocale.value = firstAvailableLocale()
}

function remove(index: number) {
	if (localValue.locales.length === 1) {
		return
	}
	const removed = localValue.locales[index]
	localValue.locales.splice(index, 1)
	if (removed?.locale === localValue.defaultLocale) {
		localValue.defaultLocale = localValue.locales[0]?.locale || 'zh-CN'
	}
	pendingLocale.value = firstAvailableLocale()
}

function setAsDefault(locale: string) {
	localValue.defaultLocale = locale
}

function firstAvailableLocale() {
	return availableLocaleOptions.value[0]?.value || ''
}

function copyFromDefault(targetIndex: number) {
	const defaultEntry = localValue.locales.find((entry) => entry.locale === localValue.defaultLocale)
	if (!defaultEntry) {
		toast.add({ title: '未设置默认语言', color: 'red' })
		return
	}
	const target = localValue.locales[targetIndex]
	if (!target || target.locale === defaultEntry.locale) {
		return
	}
	target.title = defaultEntry.title
	target.description = defaultEntry.description
	toast.add({ title: `已复制 ${defaultEntry.locale} 内容` })
}

function copyDefaultToAll() {
	const defaultEntry = localValue.locales.find((entry) => entry.locale === localValue.defaultLocale)
	if (!defaultEntry) {
		toast.add({ title: '未设置默认语言', color: 'red' })
		return
	}
	localValue.locales.forEach((entry) => {
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
	if (!localValue.defaultLocale) {
		errors.defaultLocale = '请选择默认语言'
		valid = false
	}
	const defaultEntry = localValue.locales.find((entry) => entry.locale === localValue.defaultLocale)
	if (!defaultEntry) {
		errors.defaultLocale = '请添加默认语言内容'
		valid = false
	} else if (!defaultEntry.title) {
		errors.defaultLocale = '默认语言标题必填'
		errors.entries[getIndexByLocale(defaultEntry.locale)] = { title: '必填' }
		valid = false
	}
	localValue.locales.forEach((entry, idx) => {
		if (entry.locale === localValue.defaultLocale && !entry.title) {
			errors.entries[idx] = { title: '默认语言标题必填' }
		}
	})
	emit('validation', valid)
	return valid
}

function getIndexByLocale(locale: string) {
	return localValue.locales.findIndex((entry) => entry.locale === locale)
}

defineExpose({ validate })
</script>
