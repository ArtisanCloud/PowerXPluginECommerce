<template>
  <UCard>
    <template #header>
      <h2 class="font-semibold text-xl text-gray-800 dark:text-white">
        绑定助手
      </h2>
    </template>

    <div v-if="selectedMedia.length > 0" class="space-y-6">
      <div>
        <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-2">
          已选素材 ({{ selectedMedia.length }})
        </h3>
        <div class="grid grid-cols-3 gap-2">
          <div v-for="media in selectedMedia" :key="media.id" class="relative aspect-square border rounded-md overflow-hidden">
            <img :src="media.url" :alt="media.name" class="w-full h-full object-cover">
          </div>
        </div>
      </div>

      <USeparator />

      <UForm :state="{}" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">绑定目标类型</label>
          <USelect v-model="targetType" :options="targetTypes" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">搜索SPU/SKU</label>
          <UInput v-model="targetSearch" placeholder="输入名称或编码..." />
          <!-- In a real app, this would be a searchable select -->
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">绑定位置</label>
          <USelect v-model="bindPosition" :options="bindPositions" />
        </div>
      </UForm>

      <UButton
        label="执行绑定"
        block
        size="lg"
        :loading="isBinding"
        @click="executeBinding"
      />
    </div>

    <div v-else class="text-center py-12">
      <UIcon name="i-heroicons-paper-clip" class="w-12 h-12 mx-auto text-gray-400" />
      <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">未选择素材</h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">从左侧媒体库中选择素材以开始绑定。</p>
    </div>
  </UCard>
</template>

<script setup lang="ts">

const props = defineProps({
  selectedMedia: {
    type: Array,
    default: () => []
  }
});

const targetTypes = ['SPU', 'SKU'];
const targetType = ref('SPU');
const targetSearch = ref('');

const bindPositions = ['主图', '轮播图', '详情图'];
const bindPosition = ref('轮播图');

const isBinding = ref(false);

function executeBinding() {
  if (!props.selectedMedia.length || !targetSearch.value) {
    alert('请选择素材并指定一个绑定目标。');
    return;
  }
  isBinding.value = true;
  console.log('Binding:', {
    mediaIds: props.selectedMedia.map(m => m.id),
    targetType: targetType.value,
    target: targetSearch.value,
    position: bindPosition.value,
  });
  setTimeout(() => {
    isBinding.value = false;
    alert(`已将 ${props.selectedMedia.length} 个素材绑定到 ${targetType.value} "${targetSearch.value}" 的 ${bindPosition.value}。`);
  }, 1000);
}
</script>
