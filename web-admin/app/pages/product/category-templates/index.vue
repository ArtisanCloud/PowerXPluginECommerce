<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">类目模板管理</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理商品类目模板定义</p>
      </div>
    </div>
    
    <CategoryTemplateList />
  </div>
</template>

<script setup lang="ts">
import CategoryTemplateList from '~/components/product/attributes/CategoryTemplateList.vue'

definePageMeta({
  layout: 'default'
})
</script>