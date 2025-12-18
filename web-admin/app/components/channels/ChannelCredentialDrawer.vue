<template>
  <UForm :state="localForm" class="space-y-6" data-testid="credential-form" @submit.prevent="handleSubmit">
    <div class="grid gap-4 md:grid-cols-2">
      <UFormField label="凭证类型 / Type" :error="errors.type" required>
        <template #default="{ id }">
          <USelectMenu
            :id="id"
            v-model="localForm.type"
            :options="typeOptions"
            value-attribute="value"
            option-attribute="label"
            placeholder="选择授权模式"
            data-testid="credential-type"
          />
        </template>
      </UFormField>
      <UFormField label="作用范围 / Scope (逗号分隔)">
        <template #default="{ id }">
          <UInput
            :id="id"
            v-model="scopeInput"
            placeholder="orders.read, orders.write"
            data-testid="credential-scope"
          />
        </template>
      </UFormField>
      <UFormField label="到期时间 / Expiration">
        <template #default="{ id }">
          <UInput
            :id="id"
            v-model="expiresInput"
            type="datetime-local"
            data-testid="credential-expiration"
          />
        </template>
      </UFormField>
      <UFormField label="附加元数据 / Metadata JSON">
        <template #default="{ id }">
          <UTextarea
            :id="id"
            v-model="metadataInput"
            placeholder='{"note":"线下合同号"}'
            data-testid="credential-metadata"
          />
        </template>
      </UFormField>
    </div>

    <UFormField label="凭证载荷 / Payload JSON" :error="errors.payload" required>
      <template #default="{ id }">
        <UTextarea
          :id="id"
          v-model="payloadInput"
          :rows="8"
          placeholder='{ "token": "xxx", "refresh_token": "yyy" }'
          data-testid="credential-payload"
        />
      </template>
    </UFormField>

    <UFormField
      v-if="localForm.type === 'offline'"
      label="线下凭证附件 URL"
      :error="errors.attachment"
      help="请先上传至任务中心或存储，再粘贴可访问链接"
    >
      <template #default="{ id }">
        <UInput
          :id="id"
          v-model="attachmentInput"
          placeholder="https://files.example.com/contracts/offline.pdf"
          data-testid="credential-attachment"
        />
      </template>
    </UFormField>

    <div class="flex items-center justify-end gap-3">
      <UButton variant="ghost" color="neutral" @click="emit('cancel')">取消</UButton>
      <UButton color="primary" :loading="loading" type="submit">
        保存凭证
      </UButton>
    </div>
  </UForm>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import type { ChannelCredentialUpsertPayload } from '~/types/channels'

const props = defineProps<{
  modelValue: ChannelCredentialUpsertPayload
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ChannelCredentialUpsertPayload): void
  (e: 'submit', value: ChannelCredentialUpsertPayload): void
  (e: 'cancel'): void
}>()

const localForm = reactive<ChannelCredentialUpsertPayload>({
  type: 'oauth',
  payload: {},
  scope: [],
  metadata: {},
})

const errors = reactive<Record<string, string>>({})
const payloadInput = ref('{}')
const metadataInput = ref('{}')
const scopeInput = ref('')
const expiresInput = ref<string | undefined>()
const attachmentInput = ref('')

const typeOptions = [
  { label: 'OAuth', value: 'oauth' },
  { label: 'API Key', value: 'api_key' },
  { label: 'Offline', value: 'offline' },
]

const syncLocal = (val?: ChannelCredentialUpsertPayload) => {
  const next: ChannelCredentialUpsertPayload = {
    type: val?.type ?? 'oauth',
    payload: val?.payload ?? {},
    scope: val?.scope ?? [],
    metadata: val?.metadata ?? {},
    expiresAt: val?.expiresAt,
    attachmentUrl: val?.attachmentUrl,
  }
  Object.assign(localForm, next)
  payloadInput.value = JSON.stringify(next.payload ?? {}, null, 2)
  metadataInput.value = JSON.stringify(next.metadata ?? {}, null, 2)
  scopeInput.value = (next.scope ?? []).join(', ')
  expiresInput.value = next.expiresAt
    ? next.expiresAt.replace('Z', '').slice(0, 16)
    : undefined
  attachmentInput.value = next.attachmentUrl ?? ''
}

watch(
  () => props.modelValue,
  (val) => syncLocal(val),
  { immediate: true, deep: true },
)

watch(
  () => localForm,
  () => {
    emit('update:modelValue', clonePayload())
  },
  { deep: true },
)

watch(scopeInput, (val) => {
  localForm.scope = val
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
})

watch(expiresInput, (val) => {
  localForm.expiresAt = val ? new Date(val).toISOString() : undefined
})

watch(attachmentInput, (val) => {
  localForm.attachmentUrl = val?.trim() || undefined
})

const clonePayload = (): ChannelCredentialUpsertPayload => ({
  type: localForm.type,
  payload: localForm.payload,
  scope: localForm.scope,
  expiresAt: localForm.expiresAt,
  metadata: localForm.metadata,
  attachmentUrl: localForm.attachmentUrl,
})

const parseJSON = (value: string, field: string): Record<string, any> | undefined => {
  if (!value || !value.trim()) {
    return {}
  }
  try {
    return JSON.parse(value)
  } catch (error) {
    errors[field] = 'JSON 格式错误'
    return undefined
  }
}

const clearErrors = () => {
  Object.keys(errors).forEach((key) => delete errors[key])
}

const handleSubmit = () => {
  clearErrors()
  if (!localForm.type) {
    errors.type = '请选择类型'
    return
  }
  if (localForm.type === 'offline' && !attachmentInput.value.trim()) {
    errors.attachment = '请提供线下凭证附件链接'
    return
  }
  const payload = parseJSON(payloadInput.value, 'payload')
  if (!payload) {
    return
  }
  const metadata = parseJSON(metadataInput.value, 'metadata')
  if (metadata === undefined && metadataInput.value.trim().length) {
    return
  }
  localForm.payload = payload
  localForm.metadata = metadata ?? {}
  emit('submit', clonePayload())
}
</script>
