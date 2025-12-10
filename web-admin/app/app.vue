<template>
  <UApp>
    <NuxtLoadingIndicator :throttle="150" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <ConfirmDialog />
    <ToastAlert
      v-model="globalToast.visible"
      :title="globalToast.title"
      :message="globalToast.message"
      :color="globalToast.color"
      :variant="globalToast.variant"
      :duration="globalToast.duration"
      :icon="globalToast.icon"
    />
  </UApp>
</template>

<script setup>
import ToastAlert from "~/components/ToastAlert.vue";
import { useToastAlertState } from "~/composables/useToastAlert";

const globalToast = useToastAlertState();

// PowerX Skeleton Admin Interface
useSeoMeta({
  title: "Powerx Plugin Ecommerce Plugin",
  description: "Template management starter for the PowerX platform",
});
</script>
