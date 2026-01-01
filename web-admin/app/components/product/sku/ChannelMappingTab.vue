<template>
  <section class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-primary-500">{{ $t('product.sku.channel.kicker') }}</p>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.channel.title') }}</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('product.sku.channel.subtitle') }}</p>
      </div>
      <div class="flex gap-2">
        <UButton :loading="loading" icon="i-heroicons-arrow-path" variant="ghost" @click="fetchChannels">
          {{ $t('common.refresh') }}
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h4 class="text-lg font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.channel.listTitle') }}</h4>
          <span class="text-sm text-gray-500 dark:text-gray-400">{{ channels.length }} {{ $t('product.sku.channel.countUnit') }}</span>
        </div>
      </template>
      <div v-if="!channels.length && !loading" class="text-sm text-gray-500 dark:text-gray-400">
        {{ $t('product.sku.channel.empty') }}
      </div>
      <UTable
        v-else
        :rows="channels"
        :columns="tableColumns"
        :loading="loading"
        class="w-full"
      >
        <template #status-data="{ row }">
          <UBadge :color="statusColor(row.status)">{{ statusLabel(row.status) }}</UBadge>
        </template>
        <template #actions-data="{ row }">
          <div class="flex flex-wrap gap-2">
            <UButton
              size="xs"
              variant="ghost"
              icon="i-heroicons-pencil"
              @click="prefillForm(row)"
            >
              {{ $t('common.edit') }}
            </UButton>
            <UButton
              size="xs"
              color="primary"
              variant="soft"
              icon="i-heroicons-paper-airplane"
              :loading="publishingId === row.channelCode"
              @click="handlePublish(row)"
            >
              {{ $t('product.sku.channel.publishAction') }}
            </UButton>
          </div>
        </template>
        <template #publishTime-data="{ row }">
          <span class="text-sm text-gray-500">{{ formatDateTime(row.publishTime) }}</span>
        </template>
        <template #lastError-data="{ row }">
          <span class="text-xs text-red-500" v-if="row.lastError">{{ row.lastError }}</span>
          <span v-else class="text-xs text-gray-400">—</span>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <h4 class="text-lg font-semibold text-gray-900 dark:text-white">{{ $t('product.sku.channel.formTitle') }}</h4>
      </template>
      <UForm class="grid gap-4 md:grid-cols-2" @submit.prevent="handleSave">
        <UFormGroup :label="$t('product.sku.channel.fields.channelCode')" required>
          <UInput v-model="form.channelCode" placeholder="ec_shop" />
        </UFormGroup>
        <UFormGroup :label="$t('product.sku.channel.fields.channelSkuId')" required>
          <UInput v-model="form.channelSkuId" placeholder="SKU-001" />
        </UFormGroup>
        <UFormGroup :label="$t('product.sku.channel.fields.syncMode')">
          <USelect v-model="form.syncMode" :options="syncModeOptions" />
        </UFormGroup>
        <UFormGroup :label="$t('product.sku.channel.fields.status')">
          <USelect v-model="form.status" :options="statusOptions" />
        </UFormGroup>
        <div class="md:col-span-2 flex justify-end gap-2">
          <UButton variant="ghost" @click="resetForm">{{ $t('common.reset') }}</UButton>
          <UButton type="submit" color="primary" :loading="saving">
            {{ $t('product.sku.channel.saveAction') }}
          </UButton>
        </div>
      </UForm>
    </UCard>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n, useToast } from '#imports'
import type { TableColumn } from '@nuxt/ui'
import type { SkuChannelMapping, SkuChannelMappingPayload, SkuChannelStatus } from '~/types/product/sku'
import { useProductSkuStore } from '~/stores/productSku'

const props = defineProps<{ skuId: string }>()

const store = useProductSkuStore()
const toast = useToast()
const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const publishingId = ref<string>('')
const form = ref<SkuChannelMappingPayload>({
  channelCode: '',
  channelSkuId: '',
  syncMode: 'push',
  status: 'pending',
})

const channels = computed<SkuChannelMapping[]>(() => store.channelMappings[props.skuId] ?? [])

const syncModeOptions = [
  { label: 'Push', value: 'push' },
  { label: 'Pull', value: 'pull' },
  { label: 'Hybrid', value: 'hybrid' },
]

const statusOptions = [
  { label: t('product.sku.channel.status.pending'), value: 'pending' },
  { label: t('product.sku.channel.status.published'), value: 'published' },
  { label: t('product.sku.channel.status.failed'), value: 'failed' },
  { label: t('product.sku.channel.status.offline'), value: 'offline' },
]

const tableColumns = computed<TableColumn<SkuChannelMapping>[]>(() => [
  { accessorKey: 'channelCode', header: t('product.sku.channel.table.channel') },
  { accessorKey: 'channelSkuId', header: t('product.sku.channel.table.skuId') },
  { accessorKey: 'status', header: t('product.sku.channel.table.status') },
  { accessorKey: 'publishTime', header: t('product.sku.channel.table.publishTime') },
  { accessorKey: 'lastError', header: t('product.sku.channel.table.lastError') },
  { id: 'actions', header: t('common.actions') },
])

const fetchChannels = async () => {
  if (!props.skuId) return
  loading.value = true
  try {
    await store.fetchChannelMappings(props.skuId)
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  if (!props.skuId || !form.value.channelCode || !form.value.channelSkuId) {
    return
  }
  saving.value = true
  try {
    await store.saveChannelMapping(props.skuId, { ...form.value })
    toast.add({ title: t('product.sku.channel.toastSaved') })
    resetForm()
  } catch (error: any) {
    toast.add({ title: t('common.error'), description: error?.message, color: 'red' })
  } finally {
    saving.value = false
  }
}

const handlePublish = async (mapping: SkuChannelMapping) => {
  if (!props.skuId) return
  publishingId.value = mapping.channelCode
  try {
    await store.publishChannelMapping(props.skuId, { channelCode: mapping.channelCode })
    toast.add({ title: t('product.sku.channel.toastPublished', { channel: mapping.channelCode }) })
  } catch (error: any) {
    toast.add({ title: t('common.error'), description: error?.message, color: 'red' })
  } finally {
    publishingId.value = ''
  }
}

const prefillForm = (mapping: SkuChannelMapping) => {
  form.value = {
    channelCode: mapping.channelCode,
    channelSkuId: mapping.channelSkuId,
    syncMode: mapping.syncMode,
    status: mapping.status,
  }
}

const resetForm = () => {
  form.value = {
    channelCode: '',
    channelSkuId: '',
    syncMode: 'push',
    status: 'pending',
  }
}

const statusColor = (status?: SkuChannelStatus) => {
  switch (status) {
    case 'published':
      return 'success'
    case 'failed':
      return 'error'
    case 'offline':
      return 'gray'
    default:
      return 'warning'
  }
}

const statusLabel = (status?: SkuChannelStatus) => {
  switch (status) {
    case 'published':
      return t('product.sku.channel.status.published')
    case 'failed':
      return t('product.sku.channel.status.failed')
    case 'offline':
      return t('product.sku.channel.status.offline')
    default:
      return t('product.sku.channel.status.pending')
  }
}

const formatDateTime = (value?: string) => {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}

watch(
  () => props.skuId,
  () => {
    fetchChannels()
  },
  { immediate: true },
)
</script>
