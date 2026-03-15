<template>
  <div
    :class="navbarClass"
  >
    <div class="flex h-16 items-center justify-between px-6">
      <!-- 左侧品牌 -->
      <div class="flex items-center">
        <img
          :src="logoSrc"
          alt="PowerX Plugin Logo"
          class="h-8 w-auto mr-3"
        />
        <h1 class="text-xl font-semibold text-slate-100">
          {{ $t("common.appName") }}
        </h1>
      </div>

      <!-- 中间运行模式提示 -->
      <div class="flex items-center space-x-3">
        <p class="text-xs text-gray-400">
          {{ iamModeDescription }}
        </p>
      </div>

      <!-- 右侧控制区 -->
      <div class="flex items-center space-x-4">
        <!-- 通知 -->
        <UButton
          variant="ghost"
          color="neutral"
          size="sm"
          square
          @click="toggleNotifications"
        >
          <UIcon name="i-heroicons-bell" class="w-5 h-5" />
        </UButton>

        <ThemeSelector />
        <LanguageSelector />

        <!-- 用户头像和下拉菜单 -->
        <UDropdownMenu
          :items="userMenuItems"
          :popper="{ placement: 'bottom-end' }"
        >
          <UAvatar
            src="https://avatars.githubusercontent.com/u/739984?v=4"
            alt="管理员"
            size="sm"
            class="cursor-pointer"
          />
        </UDropdownMenu>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useAuth } from "~/composables/useAuth";

const { t } = useI18n();
const runtimeConfig = useRuntimeConfig();
const auth = useAuth();
const colorMode = useColorMode();
const iamModeDescription = computed(() =>
  runtimeConfig.public.insidePowerX ? "通过宿主 PowerX 鉴权" : "当前使用本地目录与 STS"
);

const navbarClass = computed(() =>
  colorMode.value === "dark"
    ? "border-b border-gray-800 bg-slate-900/95 text-slate-100 shadow-lg shadow-slate-900/40 backdrop-blur"
    : "border-b border-gray-200 bg-white text-slate-900 shadow"
);

const logoSrc = computed(() => {
  const base = runtimeConfig.public.insidePowerX
    ? runtimeConfig.public.pluginAdminBase ?? "/"
    : "/";
  return `${base.replace(/\/$/, "")}/images/logo-s.png`;
});

// 用户菜单项
const handleLogout = async () => {
  await auth.logout();
};

const userMenuItems = [
  [
    {
      label: t("navigation.profile"),
      avatar: {
        src: "https://avatars.githubusercontent.com/u/739984?v=4",
      },
      click: () => navigateTo("/profile"),
    },
  ],
  [
    {
      label: t("navigation.help"),
      icon: "i-heroicons-question-mark-circle",
      click: () => navigateTo("/help"),
    },
  ],
  [
    {
      label: t("navigation.logout"),
      icon: "i-heroicons-arrow-right-on-rectangle",
      onSelect: () => handleLogout(),
    },
  ],
];

// 切换通知
const toggleNotifications = () => {
  // TODO: 实现通知面板切换
  console.log("切换通知");
};

// 退出登录
const logout = handleLogout;
</script>
