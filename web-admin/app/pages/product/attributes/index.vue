<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">属性管理</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理商品属性定义</p>
      </div>
    </div>
    
    <AttributeList />
  </div>
</template>

<script setup lang="ts">
import AttributeList from '~/components/product/attributes/AttributeList.vue'

definePageMeta({
  layout: 'default'
})
</script>