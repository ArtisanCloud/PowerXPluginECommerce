<template>
  <view class="w-full px-6 pb-10 pt-4">
    <view class="mx-auto flex max-w-md flex-col gap-4">
      <view
        class="flex h-14 w-full items-center justify-center rounded-xl bg-primary px-5 text-lg font-extrabold text-white shadow-cta"
        hover-class="opacity-95"
        @tap="$emit('start')"
      >
        <text class="truncate">{{ ctaStart }}</text>
        <text class="ml-2 text-lg font-black">→</text>
      </view>

      <view class="flex h-12 w-full items-center justify-center" @tap="$emit('login')">
        <text class="text-base font-semibold text-text-dark">{{ ctaLoginPrefix }}</text>
        <text class="ml-1 text-base font-semibold text-primary">{{ ctaLoginAction }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
defineOptions({
  options: {
    addGlobalClass: true,
    styleIsolation: "apply-shared",
  },
});

defineProps<{
  ctaStart: string;
  ctaLoginPrefix: string;
  ctaLoginAction: string;
}>();

defineEmits<{
  (e: "start"): void;
  (e: "login"): void;
}>();
</script>
