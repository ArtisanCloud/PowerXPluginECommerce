<template>
  <UModal v-model:open="isOpen" :ui="{ content: 'w-full sm:max-w-lg' }">
    <UCard>
      <template #header>
        <h2 class="font-semibold text-xl">编辑素材信息</h2>
      </template>

      <div class="space-y-4">
        <div class="aspect-video border rounded-lg overflow-hidden">
          <img :src="form.url" :alt="form.name" class="w-full h-full object-contain">
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">名称</label>
          <UInput v-model="form.name" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">所属分组</label>
          <USelectMenu v-model="form.group" :options="groups" editable />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">标签</label>
          <UInput
            v-model="newTag"
            placeholder="添加标签后按回车"
            @keydown.enter.prevent="addTag"
          />
          <div class="flex flex-wrap gap-2 mt-2">
            <UBadge
              v-for="(tag, index) in form.tags"
              :key="index"
              variant="soft"
            >
              {{ tag }}
              <UButton
                icon="i-heroicons-x-mark"
                size="2xs"
                color="gray"
                variant="link"
                class="-mr-1"
                @click="removeTag(index)"
              />
            </UBadge>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-between">
          <UButton
            label="删除"
            color="red"
            variant="soft"
            :loading="isDeleting"
            @click="deleteMedia"
          />
          <div class="flex gap-3">
            <UButton variant="ghost" color="gray" @click="isOpen = false">取消</UButton>
            <UButton
              label="保存"
              :loading="isSaving"
              @click="saveChanges"
            />
          </div>
        </div>
      </template>
    </UCard>
  </UModal>
</template>

<script setup lang="ts">

const props = defineProps({
  open: { type: Boolean, default: false },
  media: { type: Object, required: true }
});
const emit = defineEmits(['update:open', 'updated', 'deleted']);

const isOpen = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
});

const form = ref({
  id: '',
  name: '',
  url: '',
  group: '',
  tags: []
});

const groups = ['默认分组', '商品主图', '详情页素材'];
const newTag = ref('');

const isSaving = ref(false);
const isDeleting = ref(false);

watch(() => props.media, (newMedia) => {
  if (newMedia) {
    form.value = { ...newMedia, tags: [...(newMedia.tags || [])] };
  }
}, { immediate: true, deep: true });

function addTag() {
  if (newTag.value && !form.value.tags.includes(newTag.value)) {
    form.value.tags.push(newTag.value);
  }
  newTag.value = '';
}

function removeTag(index: number) {
  form.value.tags.splice(index, 1);
}

async function saveChanges() {
  isSaving.value = true;
  await new Promise(resolve => setTimeout(resolve, 500));
  emit('updated', { ...form.value });
  isSaving.value = false;
  isOpen.value = false;
}

async function deleteMedia() {
  if (!confirm(`确定要删除素材 "${props.media.name}" 吗？`)) return;

  isDeleting.value = true;
  await new Promise(resolve => setTimeout(resolve, 500));
  emit('deleted', props.media.id);
  isDeleting.value = false;
  isOpen.value = false;
}
</script>
