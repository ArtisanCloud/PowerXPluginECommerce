<template>
  <div class="p-4 sm:p-6 lg:p-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-3xl">
          媒体素材
        </h1>
        <p class="mt-2 text-sm text-gray-700 dark:text-gray-200">
          管理您的图片、视频和3D模型，并将它们轻松绑定到商品。
        </p>
      </div>
    </div>

    <div class="mt-8 grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Media Library Section -->
      <div class="lg:col-span-2">
        <MediaLibrary @selection-changed="updateSelection" />
      </div>

      <!-- Binding Helper Section -->
      <div class="lg:col-span-1">
        <BindingHelper :selected-media="selectedMedia" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import MediaLibrary from '~/components/product/media/MediaLibrary.vue';
import BindingHelper from '~/components/product/media/BindingHelper.vue';

definePageMeta({
  layout: 'default',
  key: 'product-media',
  title: '媒体素材',
  perm: 'ec.media.*',
});

const selectedMedia = ref([]);

function updateSelection(newSelection) {
  selectedMedia.value = newSelection;
}
</script>
