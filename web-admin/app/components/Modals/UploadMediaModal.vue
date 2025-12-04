<template>
  <UModal v-model:open="isOpen" :ui="{ content: 'w-full sm:max-w-2xl' }">
    <UCard>
      <template #header>
        <h2 class="font-semibold text-xl">上传新素材</h2>
      </template>

      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">选择文件</label>
          <div
            class="border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg p-8 text-center cursor-pointer hover:border-primary-500"
            @click="triggerFileInput"
            @dragover.prevent
            @drop.prevent="handleFileDrop"
          >
            <input
              ref="fileInput"
              type="file"
              multiple
              class="hidden"
              @change="handleFileSelect"
            >
            <UIcon name="i-heroicons-cloud-arrow-up" class="w-12 h-12 mx-auto text-gray-400" />
            <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">
              拖拽文件到此处，或
              <span class="font-medium text-primary-600">点击上传</span>
            </p>
            <p class="mt-1 text-xs text-gray-500">支持图片、视频、3D模型文件</p>
          </div>
        </div>

        <div v-if="selectedFiles.length > 0" class="space-y-2">
          <h3 class="text-sm font-medium">待上传文件:</h3>
          <div class="max-h-60 overflow-y-auto space-y-2 pr-2">
            <div v-for="(file, index) in selectedFiles" :key="index" class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800 rounded-md">
              <div class="flex items-center gap-3 min-w-0">
                <UIcon name="i-heroicons-document" class="w-5 h-5 flex-shrink-0" />
                <p class="text-sm truncate">{{ file.name }}</p>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-500">{{ (file.size / 1024).toFixed(1) }} KB</span>
                <UButton
                  icon="i-heroicons-x-mark"
                  size="2xs"
                  color="red"
                  variant="ghost"
                  @click="removeFile(index)"
                />
              </div>
            </div>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">所属分组</label>
          <USelectMenu v-model="targetGroup" :options="groups" />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <UButton variant="ghost" color="gray" @click="isOpen = false">取消</UButton>
          <UButton
            label="开始上传"
            :loading="isUploading"
            :disabled="selectedFiles.length === 0"
            @click="startUpload"
          />
        </div>
      </template>
    </UCard>
  </UModal>
</template>

<script setup lang="ts">

const props = defineProps({
  open: { type: Boolean, default: false }
});
const emit = defineEmits(['update:open', 'uploaded']);

const isOpen = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
});

const fileInput = ref<HTMLInputElement | null>(null);
const selectedFiles = ref<File[]>([]);
const isUploading = ref(false);

const groups = ['默认分组', '商品主图', '详情页素材'];
const targetGroup = ref(groups[0]);

function triggerFileInput() {
  fileInput.value?.click();
}

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  if (target.files) {
    addFiles(Array.from(target.files));
  }
}

function handleFileDrop(event: DragEvent) {
  if (event.dataTransfer?.files) {
    addFiles(Array.from(event.dataTransfer.files));
  }
}

function addFiles(files: File[]) {
  selectedFiles.value.push(...files);
}

function removeFile(index: number) {
  selectedFiles.value.splice(index, 1);
}

async function startUpload() {
  isUploading.value = true;
  // Simulate upload process
  await new Promise(resolve => setTimeout(resolve, 1500));

  // In a real app, you would upload files here and get back their details.
  // For now, we just emit the file objects.
  emit('uploaded', selectedFiles.value);

  isUploading.value = false;
  isOpen.value = false;
  selectedFiles.value = [];
}

watch(isOpen, (newVal) => {
  if (!newVal) {
    selectedFiles.value = [];
  }
});
</script>
