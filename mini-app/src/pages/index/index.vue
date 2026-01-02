<template>
  <view
    class="relative flex min-h-screen w-full flex-col justify-between bg-background-light font-display text-text-dark"
  >
    <welcome-header
      :brand="t('brand')"
      :lang-switch-label="t('langSwitch')"
      @toggle-locale="toggleLocale"
    />

    <view class="flex-1 flex flex-col" style="flex: 1;">
      <view class="flex w-full flex-1 flex-col items-center justify-center px-2" style="flex: 1;">
        <view class="w-full">
          <welcome-hero :image-src="heroImage" :badge="t('badge')" />
        </view>

        <view class="w-full max-w-md px-6 pt-6 text-center">
          <view class="text-3xl font-extrabold leading-tight tracking-tight text-text-dark">
            <text>{{ t("titlePrefix") }}</text>
            <text class="text-primary"> {{ t("titleEmphasis") }}</text>
          </view>
          <view class="pt-3 text-base font-medium leading-relaxed text-muted">
            {{ t("subtitle") }}
          </view>
        </view>

        <view class="flex w-full items-center justify-center gap-2 py-8">
          <view class="h-2 w-8 rounded-full bg-primary" />
          <view class="h-2 w-2 rounded-full bg-gray-300" />
          <view class="h-2 w-2 rounded-full bg-gray-300" />
        </view>
      </view>
    </view>

    <view class="w-full">
      <welcome-actions
        :cta-start="t('ctaStart')"
        :cta-login-prefix="t('ctaLoginPrefix')"
        :cta-login-action="t('ctaLoginAction')"
        @start="onStart"
        @login="onLogin"
      />
    </view>
  </view>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import WelcomeActions from "@/components/welcome/welcome-actions.vue";
import WelcomeHeader from "@/components/welcome/welcome-header.vue";
import WelcomeHero from "@/components/welcome/welcome-hero.vue";
import { setAppLocale } from "@/i18n";

const { t, locale } = useI18n();

const heroImage =
  "https://lh3.googleusercontent.com/aida-public/AB6AXuDXhvQlSIAMyQqkrlRL25jpTRhnE6xR_9YNQFtevbLQlz_6d_o_A1Mic7jEz-Zu3-Bj_bbK6y7auL3TWvFIn6CY_AudULOEs2ZoZg00hm6TqFOmLe8wWvjgqphhmAB8Q-4WdIAYApOv_GSj2IX9LIeEmMe6rX3mNVl8d7aM4wYdVmyJu74Za8qqWkGZDvBqWDxWFOla9AuIf926uUftgGtMUY2GjtiQlDbaE5ZlFi6iwE-rVVfiKrcAFg9qGe0yVqXmunUBRAn6hQHL";

function toggleLocale() {
  const next = locale.value === "zh-CN" ? "en" : "zh-CN";
  locale.value = next;
  setAppLocale(next);
}

function onStart() {
  uni.showToast({ title: t("ctaStart"), icon: "none" });
}

function onLogin() {
  uni.showToast({ title: t("ctaLoginAction"), icon: "none" });
}

onMounted(() => {
  // noop: 先对齐设计稿的浅色风格；后续需要再补暗色模式
});
</script>
