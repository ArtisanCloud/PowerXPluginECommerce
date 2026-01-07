import { createI18n } from "vue-i18n";
import en from "./locales/en.json";
import zhCN from "./locales/zh-CN.json";

type AppLocale = "en" | "zh-CN";

const STORAGE_KEY = "miniapp.locale";

function normalizeLocale(raw: string | undefined | null): AppLocale {
  const v = String(raw ?? "").trim();
  if (!v) return "zh-CN";
  if (v === "zh-CN" || v === "zh") return "zh-CN";
  if (v.toLowerCase().startsWith("zh")) return "zh-CN";
  return "en";
}

function detectLocale(): AppLocale {
  try {
    const stored = uni.getStorageSync(STORAGE_KEY);
    if (stored) return normalizeLocale(stored);
  } catch {}

  try {
    // H5 / App / 部分小程序可用
    const uniAny = uni as any;
    if (typeof uniAny.getLocale === "function") return normalizeLocale(uniAny.getLocale());
  } catch {}

  try {
    const sys = uni.getSystemInfoSync();
    return normalizeLocale((sys as any).language);
  } catch {}

  return "zh-CN";
}

export function setAppLocale(locale: AppLocale) {
  try {
    uni.setStorageSync(STORAGE_KEY, locale);
  } catch {}
}

export function createAppI18n() {
  return createI18n({
    legacy: false,
    globalInjection: true,
    locale: detectLocale(),
    fallbackLocale: "en",
    messages: {
      en,
      "zh-CN": zhCN,
    },
  });
}
