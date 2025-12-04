<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">属性组管理</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理商品属性组定义</p>
      </div>
    </div>
    
    <AttributeGroupList />
  </div>
</template>

<script setup lang="ts">
import AttributeGroupList from '~/components/product/attributes/AttributeGroupList.vue'

definePageMeta({
  layout: 'default'
})
</script>