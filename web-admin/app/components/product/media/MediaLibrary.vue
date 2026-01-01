<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <h2 class="font-semibold text-xl text-gray-800 dark:text-white">
          媒体库
        </h2>
        <div class="flex items-center gap-2">
          <UButton
            icon="i-heroicons-arrow-up-tray"
            size="sm"
            color="primary"
            variant="solid"
            label="上传素材"
            @click="isUploadModalOpen = true"
          />
          <UButton
            icon="i-heroicons-squares-plus"
            size="sm"
            color="white"
            variant="solid"
            label="新建分组"
            @click="createGroup"
          />
        </div>
      </div>
    </template>

    <!-- Toolbar -->
    <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">搜索</label>
          <UInput
            v-model="searchQuery"
            icon="i-heroicons-magnifying-glass"
            placeholder="搜索名称、标签..."
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">分组</label>
          <USelectMenu
            v-model="selectedGroup"
            :options="groups"
            placeholder="所有分组"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">标签</label>
          <USelectMenu
            v-model="selectedTag"
            :options="tags"
            placeholder="所有标签"
          />
        </div>
      </div>
    </div>

    <!-- Tabs and Media Grid -->
    <UTabs v-model="selectedTab" :items="tabs" class="w-full pt-4">
      <template #item="{ item }">
        <div v-if="filteredMedia.length > 0">
          <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
            <div
              v-for="media in filteredMedia"
              :key="media.id"
              class="relative group aspect-square border rounded-lg overflow-hidden cursor-pointer"
              :class="isSelected(media) ? 'ring-2 ring-primary-500' : 'border-gray-200 dark:border-gray-700'"
              @click="toggleSelection(media)"
            >
              <img :src="media.url" :alt="media.name" class="w-full h-full object-cover">
              <div class="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-50 transition-all duration-300 flex items-center justify-center">
                <div class="absolute top-2 right-2">
                  <UCheckbox :model-value="isSelected(media)" @click.stop="toggleSelection(media)" />
                </div>
                <div class="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-black/80 to-transparent">
                  <p class="text-white text-sm font-medium truncate">{{ media.name }}</p>
                </div>
                <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <UButton icon="i-heroicons-pencil-square" size="sm" color="white" @click.stop="openEditModal(media)" />
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="text-center py-12">
          <UIcon name="i-heroicons-photo" class="w-12 h-12 mx-auto text-gray-400" />
          <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">没有素材</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">上传一些新的媒体素材吧。</p>
        </div>
      </template>
    </UTabs>

    <!-- Modals -->
    <UploadMediaModal v-model:open="isUploadModalOpen" @uploaded="handleUploaded" />
    <EditMediaModal v-if="editingMedia" v-model:open="isEditModalOpen" :media="editingMedia" @updated="handleUpdated" />
  </UCard>
</template>

<script setup lang="ts">
import UploadMediaModal from '~/components/modals/UploadMediaModal.vue';
import EditMediaModal from '~/components/modals/EditMediaModal.vue';

const emit = defineEmits(['selection-changed']);

const tabs = [
  { key: 'image', label: '图片' },
  { key: 'video', label: '视频' },
  { key: '3d', label: '3D模型' },
];
const selectedTab = ref(0);

const searchQuery = ref('');
const groups = ['默认分组', '商品主图', '详情页素材'];
const selectedGroup = ref(groups[0]);
const tags = ['夏季新品', '促销活动', '背景图'];
const selectedTag = ref('');

const allMedia = ref([
  { id: 1, name: 'Product A Main', url: 'https://picsum.photos/seed/1/400/400', type: 'image', group: '商品主图', tags: ['夏季新品'] },
  { id: 2, name: 'Product B Detail', url: 'https://picsum.photos/seed/2/400/400', type: 'image', group: '详情页素材', tags: ['促销活动'] },
  { id: 3, name: 'Summer Background', url: 'https://picsum.photos/seed/3/400/400', type: 'image', group: '默认分组', tags: ['背景图', '夏季新品'] },
  { id: 4, name: 'Promo Video', url: 'https://picsum.photos/seed/4/400/400', type: 'video', group: '促销活动', tags: ['视频'] },
  { id: 5, name: '3D Model of Chair', url: 'https://picsum.photos/seed/5/400/400', type: '3d', group: '默认分组', tags: ['家具'] },
]);

const filteredMedia = computed(() => {
  const currentType = tabs[selectedTab.value].key;
  return allMedia.value.filter(media => {
    const typeMatch = media.type === currentType;
    const groupMatch = !selectedGroup.value || media.group === selectedGroup.value;
    const tagMatch = !selectedTag.value || media.tags.includes(selectedTag.value);
    const searchMatch = !searchQuery.value ||
      media.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      media.tags.some(t => t.toLowerCase().includes(searchQuery.value.toLowerCase()));
    return typeMatch && groupMatch && tagMatch && searchMatch;
  });
});

const selectedMedia = ref([]);

function isSelected(media) {
  return selectedMedia.value.some(m => m.id === media.id);
}

function toggleSelection(media) {
  if (isSelected(media)) {
    selectedMedia.value = selectedMedia.value.filter(m => m.id !== media.id);
  } else {
    selectedMedia.value.push(media);
  }
  emit('selection-changed', selectedMedia.value);
}

// Modals
const isUploadModalOpen = ref(false);
const isEditModalOpen = ref(false);
const editingMedia = ref(null);

function createGroup() {
  const newGroup = prompt('请输入新的分组名称:');
  if (newGroup && !groups.includes(newGroup)) {
    groups.push(newGroup);
    selectedGroup.value = newGroup;
  }
}

function openEditModal(media) {
  editingMedia.value = media;
  isEditModalOpen.value = true;
}

function handleUploaded(newFiles) {
  // Mock adding files
  newFiles.forEach(file => {
    allMedia.value.push({
      id: Date.now() + Math.random(),
      name: file.name,
      url: URL.createObjectURL(file),
      type: file.type.startsWith('image') ? 'image' : 'video',
      group: '默认分组',
      tags: []
    });
  });
}

function handleUpdated(updatedMedia) {
  const index = allMedia.value.findIndex(m => m.id === updatedMedia.id);
  if (index !== -1) {
    allMedia.value[index] = updatedMedia;
  }
}
</script>
