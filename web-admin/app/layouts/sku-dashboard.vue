<template>
  <div class="min-h-screen bg-slate-50">
    <AppNavbar />
    <div class="mx-auto flex w-full max-w-6xl flex-col gap-4 px-6 py-6">
      <slot name="banner" />
      <main class="rounded-2xl bg-white p-6 shadow-sm">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppNavbar from '~/components/AppNavbar.vue'
</script>
