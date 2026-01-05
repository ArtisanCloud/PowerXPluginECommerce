<template>
  <view class="relative w-full px-6 pb-4" :style="rootStyle">
    <view class="flex w-full items-center justify-center gap-3">
      <view class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-white shadow-md">
        <text class="text-xl font-extrabold leading-none">S</text>
      </view>
      <text class="text-xl font-extrabold tracking-tight text-text-dark">{{ brand }}</text>
    </view>

    <view
      class="absolute right-6 rounded-full bg-surface-90 px-3 py-2 text-xs font-extrabold text-text-dark shadow-sm border border-line-5"
      :style="switchStyle"
      @tap="$emit('toggle-locale')"
    >
      {{ langSwitchLabel }}
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from "vue";

defineOptions({
  options: {
    addGlobalClass: true,
    styleIsolation: "apply-shared",
  },
});

const props = defineProps<{
  brand: string;
  langSwitchLabel: string;
  topInset?: number;
}>();

defineEmits<{
  (e: "toggle-locale"): void;
}>();

const rootStyle = computed(() => {
  const top = Number(props.topInset ?? 0);
  return top > 0 ? `padding-top:${top}px;` : "padding-top:40px;";
});

const switchStyle = computed(() => {
  const top = Number(props.topInset ?? 0);
  return top > 0 ? `top:${top}px;` : "top:40px;";
});
</script>
