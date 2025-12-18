<template>
<UForm :state="form" class="space-y-6" @submit.prevent="handleSubmit">
    <div class="grid gap-4 md:grid-cols-2">
      <UFormField label="渠道名称 / Channel Name" :error="errors.name" required>
        <template #default="{ id }">
          <UInput :id="id" v-model.trim="form.name" class="w-full" placeholder="天猫旗舰店"/>
        </template>
      </UFormField>
      <UFormField label="平台 / Platform" :error="errors.platform" required>
        <template #default="{ id }">
          <USelect
            :id="id"
            v-model="form.platform"
            :items="platformSelectOptions"
            value-attribute="value"
            option-attribute="label"
            class="w-full"
            placeholder="选择平台"
          />
        </template>
      </UFormField>
      <UFormField label="店铺 ID / Store ID" :error="errors.storeId" required>
        <template #default="{ id }">
          <UInput :id="id" v-model.trim="form.storeId" class="w-full" placeholder="tmall-001"/>
        </template>
      </UFormField>
      <UFormField label="渠道类型 / Channel Type" :error="errors.channelType" required>
        <template #default="{ id }">
          <USelect
            :id="id"
            v-model="form.channelType"
            :items="channelTypeSelectOptions"
            value-attribute="value"
            option-attribute="label"
            class="w-full"
            placeholder="选择类型"
          />
        </template>
      </UFormField>
      <UFormField label="国家 / Country" :error="errors.region" required>
        <template #default="{ id }">
          <USelect
            :id="id"
            v-model="selectedCountry"
            :items="countrySelectOptions"
            value-attribute="value"
            option-attribute="label"
            class="w-full"
            placeholder="选择国家"
          />
        </template>
      </UFormField>
      <UFormField label="城市 / City" :error="errors.region" required>
        <template #default="{ id }">
          <USelect
            :id="id"
            v-model="selectedCity"
            :items="citySelectOptions"
            value-attribute="value"
            option-attribute="label"
            class="w-full"
            placeholder="选择城市"
          />
        </template>
      </UFormField>
      <UFormField label="域名 / Domain">
        <template #default="{ id }">
          <UInput :id="id" v-model.trim="form.domain" class="w-full" placeholder="store.example.com"/>
        </template>
      </UFormField>
      <UFormField
        label="运营负责人 / Owner"
        :error="errors.ownerUuid"
        :description="ownerDescription"
        required
      >
        <template #default="{ id }">
          <USelectMenu
            :id="id"
            v-model="form.ownerUuid"
            :items="ownerSelectOptions"
            value-key="value"
            label-key="label"
            searchable
            v-model:search-term="ownerSearchTerm"
            :loading="ownerLoading"
            :ignore-filter="true"
            class="w-full"
            placeholder="选择负责人"
            @keydown.enter.prevent.stop="confirmOwnerSearch"
          >
            <template #option="{ option }">
              <div class="flex flex-col">
                <span class="font-medium text-sm text-gray-900 dark:text-white">{{ option.label }}</span>
                <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
              </div>
            </template>
            <template #empty>
              <div class="px-3 py-2 text-sm text-gray-500">
                <p>未找到匹配负责人。</p>
                <p v-if="ownerDescription" class="mt-1">{{ ownerDescription }}</p>
                <p class="mt-1">可尝试输入用户名、花名或邮箱前缀。</p>
              </div>
            </template>
          </USelectMenu>
        </template>
      </UFormField>
      <UFormField label="审批人 / Approver">
        <template #default="{ id }">
          <USelectMenu
            :id="id"
            v-model="form.approverUuid"
            :items="ownerSelectOptions"
            value-key="value"
            label-key="label"
            searchable
            clearable
            v-model:search-term="approverSearchTerm"
            :loading="ownerLoading"
            :ignore-filter="true"
            class="w-full"
            placeholder="可选，选择审批人"
            @keydown.enter.prevent.stop="confirmApproverSearch"
          >
            <template #option="{ option }">
              <div class="flex flex-col">
                <span class="font-medium text-sm text-gray-900 dark:text-white">{{ option.label }}</span>
                <span v-if="option.description" class="text-xs text-gray-500">{{ option.description }}</span>
              </div>
            </template>
            <template #empty>
              <div class="px-3 py-2 text-sm text-gray-500">
                <p>未找到匹配审批人，可输入关键词继续搜索。</p>
              </div>
            </template>
          </USelectMenu>
        </template>
      </UFormField>
      <UFormField class="md:col-span-2" label="标签 / Tags (逗号分隔)">
        <template #default="{ id }">
          <UInput :id="id" v-model="tagsInput" class="w-full" placeholder="data_gap, alerting"/>
        </template>
      </UFormField>
    </div>

    <div class="grid gap-4 md:grid-cols-2">
      <UFormField label="联系人姓名 / Contact Name" :error="errors.contactName" required>
        <template #default="{ id }">
          <UInput :id="id" v-model.trim="form.contact.name" class="w-full" placeholder="张敏"/>
        </template>
      </UFormField>
      <UFormField label="联系电话 / Phone" :error="errors.contactPhone" required>
        <template #default="{ id }">
          <UInput :id="id" v-model.trim="form.contact.phone" class="w-full" placeholder="+86 138****7788"/>
        </template>
      </UFormField>
      <UFormField class="md:col-span-2" label="联系邮箱 / Email" :error="errors.contactEmail" required>
        <template #default="{ id }">
          <UInput :id="id" v-model.trim="form.contact.email" class="w-full" placeholder="ops@example.com"/>
        </template>
      </UFormField>
    </div>

    <UFormField
      v-if="isOfflineChannel"
      label="线下凭证 (占位) / Offline Evidence"
      description="当前版本仅展示占位，稍后可在提交审批时上传线下附件。"
    >
      <div class="rounded-lg border border-dashed border-gray-300 p-4 text-sm text-gray-500 dark:border-gray-700 dark:text-gray-400">
        <p>将于 US2 阶段支持上传 PDF/图片。现在可通过备注记录凭证位置。</p>
      </div>
    </UFormField>

    <div class="flex items-center justify-end gap-3">
      <UButton variant="ghost" color="neutral" type="button" @click="emit('cancel')">
        取消
      </UButton>
      <UButton color="primary" type="submit" :loading="loading">
        {{ mode === 'edit' ? '保存修改' : '创建渠道' }}
      </UButton>
    </div>
  </UForm>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useVModel } from '@vueuse/core'
import type { ChannelDraftPayload } from '~/types/channels'
import { createEmptyChannelPayload } from '~/types/channels'

type SelectOption = { label: string; value: string }
type ExtendedSelectOption = SelectOption & { description?: string }

type CountryOption = {
  code: string
  label: string
  cities: { code: string; label: string }[]
}

const props = defineProps<{
  modelValue: ChannelDraftPayload
  loading?: boolean
  mode?: 'create' | 'edit'
  platformOptions?: SelectOption[]
  channelTypeOptions?: SelectOption[]
  countryOptions?: CountryOption[]
  ownerOptions?: ExtendedSelectOption[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ChannelDraftPayload): void
  (e: 'submit', value: ChannelDraftPayload): void
  (e: 'cancel'): void
  (e: 'search-owner', keyword: string): void
}>()

const form = useVModel(props, 'modelValue', emit, {
  passive: true,
  defaultValue: createEmptyChannelPayload(),
})
const errors = reactive<Record<string, string>>({})
const tagsInput = computed({
  get: () => (form.value.tags ?? []).join(', '),
  set: (val: string) => {
    form.value.tags = val
      .split(',')
      .map((tag) => tag.trim())
      .filter(Boolean)
  },
})

const fallbackPlatformOptions: SelectOption[] = [
  { label: '天猫 Tmall', value: 'tmall' },
  { label: '京东 JD', value: 'jd' },
  { label: '抖音 Douyin', value: 'douyin' },
  { label: '线下 Offline', value: 'offline' },
]

const fallbackChannelTypeOptions: SelectOption[] = [
  { label: '平台授权 / Platform OAuth', value: 'platform_oauth' },
  { label: '手动凭证 / Manual Credential', value: 'platform_manual' },
  { label: '线下渠道 / Offline', value: 'offline' },
]

const fallbackCountryOptions: CountryOption[] = [
  {
    code: 'CN',
    label: '中国 China',
    cities: [
      { code: 'cn-beijing', label: '北京 Beijing' },
      { code: 'cn-shanghai', label: '上海 Shanghai' },
      { code: 'cn-shenzhen', label: '深圳 Shenzhen' },
      { code: 'cn-guangzhou', label: '广州 Guangzhou' },
    ],
  },
  {
    code: 'SG',
    label: '新加坡 Singapore',
    cities: [{ code: 'sg-singapore', label: '新加坡 Singapore' }],
  },
  {
    code: 'US',
    label: '美国 USA',
    cities: [
      { code: 'us-nyc', label: '纽约 New York' },
      { code: 'us-la', label: '洛杉矶 Los Angeles' },
      { code: 'us-sf', label: '旧金山 San Francisco' },
    ],
  },
]

const fallbackOwnerOptions: ExtendedSelectOption[] = [
  { label: 'user:ops-01', value: 'user:ops-01', description: 'ops@example.com' },
  { label: 'user:ops-02', value: 'user:ops-02', description: 'ops02@example.com' },
]

const platformSelectOptions = computed<SelectOption[]>(() =>
  props.platformOptions?.length ? props.platformOptions : fallbackPlatformOptions,
)

const channelTypeSelectOptions = computed<SelectOption[]>(() =>
  props.channelTypeOptions?.length ? props.channelTypeOptions : fallbackChannelTypeOptions,
)

const countryOptions = computed<CountryOption[]>(() =>
  props.countryOptions?.length ? props.countryOptions : fallbackCountryOptions,
)

const countrySelectOptions = computed<SelectOption[]>(() =>
  countryOptions.value.map((country) => ({ label: country.label, value: country.code })),
)

const selectedCountry = ref('')
const selectedCity = ref('')

const citySelectOptions = computed<SelectOption[]>(() => {
  const country = countryOptions.value.find((item) => item.code === selectedCountry.value)
  if (!country) {
    return []
  }
  return country.cities.map((city) => ({ label: city.label, value: city.code }))
})

const ownerOptionsSource = computed<ExtendedSelectOption[]>(() =>
  props.ownerOptions?.length ? props.ownerOptions : fallbackOwnerOptions,
)

const ownerSearchTerm = ref('')
const approverSearchTerm = ref('')
const ownerLoading = computed(() => props.ownerLoading ?? false)

const ownerSelectOptions = computed<ExtendedSelectOption[]>(() => {
  const base = [...ownerOptionsSource.value]
  const ensureValue = (value?: string) => {
    if (value && !base.some((option) => option.value === value)) {
      base.push({ label: value, value })
    }
  }
  ensureValue(form.value.ownerUuid)
  ensureValue(form.value.approverUuid)
  return base
})

watch(
  () => form.value.ownerUuid as unknown,
  (val) => {
    if (val && typeof val === 'object') {
      const normalized =
        typeof (val as ExtendedSelectOption).value === 'string'
          ? (val as ExtendedSelectOption).value
          : String((val as ExtendedSelectOption).value ?? '')
      form.value.ownerUuid = normalized
    }
  },
)

watch(
  () => form.value.approverUuid as unknown,
  (val) => {
    if (val && typeof val === 'object') {
      const normalized =
        typeof (val as ExtendedSelectOption).value === 'string'
          ? (val as ExtendedSelectOption).value
          : String((val as ExtendedSelectOption).value ?? '')
      form.value.approverUuid = normalized
    }
  },
)

const ownerDescription = computed(() => {
  if (!ownerSelectOptions.value.length) {
    return '暂无 IAM 成员，请先在用户目录创建负责人'
  }
  const names = ownerSelectOptions.value
    .slice(0, 3)
    .map((option) => option.label || option.value)
    .filter(Boolean)
  if (!names.length) {
    return ''
  }
  return `示例：${names.join(' / ')}`
})

let ownerSearchTimer: ReturnType<typeof setTimeout> | null = null
let lastOwnerQuery = ''

const emitOwnerSearch = (term: string, force = false) => {
  const normalized = term.trim()
  if (!force && normalized === lastOwnerQuery) {
    return
  }
  lastOwnerQuery = normalized
  emit('search-owner', normalized)
}

const scheduleOwnerSearch = (term: string) => {
  if (ownerSearchTimer) {
    clearTimeout(ownerSearchTimer)
  }
  ownerSearchTimer = setTimeout(() => emitOwnerSearch(term), 400)
}

watch(
  ownerSearchTerm,
  (term) => scheduleOwnerSearch(term),
  { immediate: false },
)

watch(
  approverSearchTerm,
  (term) => scheduleOwnerSearch(term),
  { immediate: false },
)

const confirmOwnerSearch = () => {
  emitOwnerSearch(ownerSearchTerm.value, true)
}

const confirmApproverSearch = () => {
  emitOwnerSearch(approverSearchTerm.value, true)
}

onBeforeUnmount(() => {
  if (ownerSearchTimer) {
    clearTimeout(ownerSearchTimer)
  }
})

const isOfflineChannel = computed(() => form.value.channelType === 'offline')

watch(
  () => form.value.region,
  (region) => {
    selectedCity.value = region || ''
    if (!region) {
      return
    }
    const matchedCountry = countryOptions.value.find((country) =>
      country.cities.some((city) => city.code === region),
    )
    if (matchedCountry && matchedCountry.code !== selectedCountry.value) {
      selectedCountry.value = matchedCountry.code
    }
  },
  { immediate: true },
)

watch(selectedCountry, (countryCode) => {
  const options = citySelectOptions.value
  if (!countryCode || !options.length) {
    if (!form.value.region) {
      selectedCity.value = ''
    }
    return
  }
  const currentRegion = form.value.region
  const belongsToCountry = options.some((city) => city.value === currentRegion)
  if (!belongsToCountry || !currentRegion) {
    selectedCity.value = options[0]?.value ?? ''
  }
})

watch(selectedCity, (cityCode) => {
  form.value.region = cityCode
})

watch(
  countrySelectOptions,
  (options) => {
    if (!selectedCountry.value && !form.value.region && options.length) {
      selectedCountry.value = options[0].value
    }
  },
  { immediate: true },
)

const validate = () => {
  Object.keys(errors).forEach((key) => {
    delete errors[key]
  })
  const requiredMap: Record<string, string> = {
    name: '请填写渠道名称',
    platform: '请选择平台',
    storeId: '请填写店铺 ID',
    channelType: '请选择渠道类型',
    region: '请选择区域',
    ownerUuid: '请填写负责人',
    contactName: '请填写联系人',
    contactPhone: '请填写联系电话',
    contactEmail: '请填写联系邮箱',
  }
  if (!form.value.name) errors.name = requiredMap.name
  if (!form.value.platform) errors.platform = requiredMap.platform
  if (!form.value.storeId) errors.storeId = requiredMap.storeId
  if (!form.value.channelType) errors.channelType = requiredMap.channelType
  if (!form.value.region) errors.region = requiredMap.region
  if (!form.value.ownerUuid) errors.ownerUuid = requiredMap.ownerUuid
  if (!form.value.contact.name) errors.contactName = requiredMap.contactName
  if (!form.value.contact.phone) errors.contactPhone = requiredMap.contactPhone
  if (!form.value.contact.email) errors.contactEmail = requiredMap.contactEmail

  return Object.keys(errors).length === 0
}

const clonePayload = (): ChannelDraftPayload => ({
  ...form.value,
  tags: [...(form.value.tags ?? [])],
  contact: { ...form.value.contact },
})

const handleSubmit = () => {
  if (!validate()) {
    return
  }
  emit('submit', clonePayload())
}
</script>
