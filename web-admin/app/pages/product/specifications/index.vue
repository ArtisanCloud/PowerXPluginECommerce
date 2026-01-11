<template>
  <div>
    <UAlert
      class="mb-6"
      color="neutral"
      variant="soft"
      title="提示"
      description="当前“规格管理”页面为属性体系（规划中/占位页面），不会写入 SPU 变体规格（product_spec_*）。变体规格请到 SPU 编辑页的“规格定义”维护，并用于生成 SKU。"
    />
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">规格管理</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理商品规格定义</p>
      </div>
    </div>
    
    <SpecificationList />
  </div>
</template>

<script setup lang="ts">
import SpecificationList from '~/components/product/attributes/SpecificationList.vue'

definePageMeta({
  layout: 'default'
})
</script>
