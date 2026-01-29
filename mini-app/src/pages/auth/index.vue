<template>
  <view class="min-h-screen w-full bg-background-light font-display text-text-dark">
    <view :style="`padding-top:${topInset}px;`" class="sticky top-0 z-10 w-full bg-background-light">
      <view class="flex items-center justify-between px-4 py-3">
        <view
          class="flex h-10 w-10 items-center justify-center rounded-full"
          hover-class="opacity-90"
          @tap="goBack"
        >
          <text class="text-xl font-black">←</text>
        </view>
        <view class="flex items-center justify-end">
          <view class="text-sm font-extrabold text-primary" @tap="onHelp">
            {{ t("auth.help") }}
          </view>
        </view>
      </view>
    </view>

    <view class="mx-auto w-full max-w-md px-4 pb-8">
      <view class="pt-4">
        <view class="relative w-full overflow-hidden rounded-xl shadow-sm">
          <image class="h-180 w-full" mode="aspectFill" :src="bannerImage" />
          <view
            class="absolute inset-0"
            style="background: linear-gradient(to top, rgba(0,0,0,0.65), rgba(0,0,0,0)); opacity: 0.6;"
          />
          <view class="absolute bottom-0 left-0 right-0 p-5">
            <view class="text-xs font-semibold uppercase tracking-wider text-white opacity-90">
              {{ t("auth.bannerKicker") }}
            </view>
            <view class="pt-1 text-xl font-extrabold text-white">
              {{ t("auth.bannerTitle") }}
            </view>
          </view>
        </view>
      </view>

      <view class="pt-6 pb-5">
        <view class="text-3xl font-extrabold leading-tight">
          {{ activeTab === "login" ? t("auth.welcomeBack") : t("auth.createAccount") }}
        </view>
        <view class="pt-2 text-base font-medium text-muted">
          {{ t("auth.subtitle") }}
        </view>
      </view>

      <view class="mb-6 flex gap-8 border-b border-line-5">
        <view
          class="px-2 pt-2 pb-3"
          :class="activeTab === 'login' ? 'border-b-4 border-primary text-text-dark' : 'border-b-4 border-transparent text-muted'"
          @tap="activeTab = 'login'"
        >
          <text class="text-base font-extrabold">{{ t("auth.tabLogin") }}</text>
        </view>
        <view
          class="px-2 pt-2 pb-3"
          :class="activeTab === 'signup' ? 'border-b-4 border-primary text-text-dark' : 'border-b-4 border-transparent text-muted'"
          @tap="activeTab = 'signup'"
        >
          <text class="text-base font-extrabold">{{ t("auth.tabSignup") }}</text>
        </view>
      </view>

	      <view class="flex flex-col gap-5">
        <view v-if="activeTab === 'signup'" class="flex flex-col gap-2">
          <text class="text-sm font-semibold">{{ t("auth.name") }}</text>
          <input
            v-model="form.name"
            class="h-14 w-full rounded-xl border border-line-5 bg-white px-4 text-base"
            :placeholder="t('auth.namePlaceholder')"
          />
        </view>

	        <view class="flex flex-col gap-2">
	          <text class="text-sm font-semibold">{{ t("auth.phone") }}</text>
	          <input
	            v-model="form.phone"
	            class="h-14 w-full rounded-xl border border-line-5 bg-white px-4 text-base"
	            :placeholder="t('auth.phonePlaceholder')"
	            type="number"
	          />
	        </view>

	        <view class="flex flex-col gap-2">
	          <text class="text-sm font-semibold">{{ t("auth.password") }}</text>
	          <input
	            v-model="form.password"
	            class="h-14 w-full rounded-xl border border-line-5 bg-white px-4 text-base"
	            :placeholder="t('auth.passwordPlaceholder')"
	            password
	          />
	        </view>

		        <view class="mt-1 flex items-start gap-3" @tap="toggleAgree">
		          <view
		            class="flex h-5 w-5 shrink-0 items-center justify-center rounded-md border border-line-5 bg-white"
		            :style="form.agree ? 'background:#4F8A7E;border-color:#4F8A7E;' : ''"
		          >
		            <view
		              v-if="form.agree"
		              class="checkmark"
			              style="width: 10px; height: 6px; border-left: 2px solid #fff; border-bottom: 2px solid #fff; transform: rotate(-45deg); margin-top: -1px;"
			            />
			          </view>
	          <view class="flex-1 text-sm text-muted">
	            {{ t("auth.agreePrefix") }}
	            <text class="text-primary" @tap.stop="onUserAgreement">{{ t("auth.userAgreement") }}</text>
	            {{ t("auth.and") }}
	            <text class="text-primary" @tap.stop="onPrivacy">{{ t("auth.privacyPolicy") }}</text>
	            {{ t("auth.dot") }}
	          </view>
	        </view>

        <view
          class="mt-4 flex h-14 w-full items-center justify-center rounded-xl bg-primary text-lg font-extrabold text-white shadow-cta"
          hover-class="opacity-95"
          @tap="submit"
        >
          {{ activeTab === "login" ? t("auth.submitLogin") : t("auth.submitSignup") }}
        </view>
      </view>

      <view class="py-8 flex items-center">
        <view class="flex-1 border-t border-line-5" />
        <view class="px-4 text-sm font-medium text-muted">{{ t("auth.orContinue") }}</view>
        <view class="flex-1 border-t border-line-5" />
      </view>

      <view class="mb-8 flex justify-center gap-6">
        <view class="flex h-14 w-14 items-center justify-center rounded-full border border-line-5 bg-white" @tap="noop">
          <text class="text-xl">#</text>
        </view>
        <view class="flex h-14 w-14 items-center justify-center rounded-full border border-line-5 bg-white" @tap="onWechatLogin">
          <text class="text-sm font-bold text-primary">{{ t("auth.wechatLogin") }}</text>
        </view>
        <view class="flex h-14 w-14 items-center justify-center rounded-full border border-line-5 bg-white" @tap="noop">
          <text class="text-xl">$</text>
        </view>
      </view>

      <view class="mt-2 flex justify-center gap-6 pb-4">
        <view class="text-sm font-semibold text-muted" @tap="onForgot">{{ t("auth.forgot") }}</view>
        <view class="h-4 w-px bg-line-5" />
        <view class="text-sm font-semibold text-muted" @tap="onSupport">{{ t("auth.support") }}</view>
      </view>
    </view>

    <view
      v-if="wechatConfirmOpen"
      class="fixed inset-0 z-50 flex items-center justify-center px-6"
      style="background-color: rgba(0,0,0,0.5);"
    >
      <view class="rounded-2xl bg-white p-6 shadow-xl" style="width: 620rpx; box-sizing: border-box;">
        <view class="text-lg font-extrabold text-text-dark">{{ t("auth.wechatConfirmTitle") }}</view>
        <view class="pt-4 flex items-center gap-4">
          <image
            v-if="wechatAvatarUrl"
            :src="wechatAvatarUrl"
            class="h-16 w-16 rounded-full bg-line-5"
            mode="aspectFill"
          />
          <view class="flex-1" style="min-width: 0;">
            <text class="text-sm font-semibold text-muted">{{ t("auth.wechatNickname") }}</text>
            <input
              v-model="wechatNickname"
              class="mt-2 h-12 rounded-xl border border-line-5 bg-white px-4 text-base"
              style="width: 100%; box-sizing: border-box;"
              :placeholder="t('auth.wechatNicknamePlaceholder')"
            />
          </view>
        </view>
        <view class="pt-6 flex gap-3">
          <view
            class="flex-1 h-12 rounded-xl border border-line-5 bg-white text-center text-sm font-semibold text-muted"
            style="line-height: 48px;"
            @tap="closeWechatConfirm"
          >
            {{ t("auth.cancel") }}
          </view>
          <view
            class="flex-1 h-12 rounded-xl bg-primary text-center text-sm font-extrabold text-white"
            style="line-height: 48px;"
            @tap="confirmWechatProfile"
          >
            {{ t("auth.confirm") }}
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
		import { reactive, ref } from "vue";
		import { onLoad } from "@dcloudio/uni-app";
	import { useI18n } from "vue-i18n";
	import { miniAppAuthLogin, miniAppAuthRegister, miniAppAuthWechatLogin } from "@/services/miniapp-auth";
	import { miniAppGetCategoryTree } from "@/services/miniapp-category";

type Tab = "login" | "signup";

		const { t } = useI18n();
		const activeTab = ref<Tab>("login");
		const topInset = ref<number>(40);

const bannerImage =
  "https://lh3.googleusercontent.com/aida-public/AB6AXuBjdUQHdpVa5nxh9kcfSdw01j9JQOzjPazG9mVP5DJ-rADq7EeTNV2uxq7Nv0D3cqoYIh7eXn8lAxJySYGFg-meAilBMA01hY7BQciiZcF4AztCYK2DkCtmWBy0MF83KpzkPsYh-aFLz7JdAxsKqpegxU6WBct3qoNQADLxdTcC5stSkvO6UuC1x_OURIUQovzD_l1pb77buafwbcCD8idP_KURyJcl-sUizeQV6zxn3BVGM6NK2CkA519EiJryOcZAcrgF6x1CWdN1";

		const form = reactive({
		  name: "",
		  phone: "",
		  password: "",
		  agree: false,
		});
    const wechatConfirmOpen = ref(false);
    const wechatCode = ref("");
    const wechatNickname = ref("");
    const wechatAvatarUrl = ref("");

function goBack() {
  try {
    uni.navigateBack();
  } catch {
    uni.switchTab?.({ url: "/pages/index/index" } as any);
  }
}

function onHelp() {
  uni.showToast({ title: t("auth.helpToast"), icon: "none" });
}

function onUserAgreement() {
  uni.showToast({ title: t("auth.userAgreement"), icon: "none" });
}

function onPrivacy() {
  uni.showToast({ title: t("auth.privacyPolicy"), icon: "none" });
}

function onForgot() {
  uni.showToast({ title: t("auth.forgot"), icon: "none" });
}

function onSupport() {
  uni.showToast({ title: t("auth.support"), icon: "none" });
}

	function noop() {
	  uni.showToast({ title: t("auth.todo"), icon: "none" });
	}

	function toggleAgree() {
	  form.agree = !form.agree;
	}

			function fullIdentifier() {
			  const phone = String(form.phone).trim();
			  return phone;
			}

			async function submit() {
		  if (!form.agree) {
		    uni.showToast({ title: t("auth.mustAgree"), icon: "none" });
		    return;
		  }
  if (!String(form.phone).trim() || !String(form.password).trim()) {
    uni.showToast({ title: t("auth.missingFields"), icon: "none" });
    return;
  }

			  try {
			    if (activeTab.value === "login") {
			      const auth = await miniAppAuthLogin({
			        identifier: fullIdentifier(),
			        password: String(form.password),
			      });
			      // 受保护接口联通校验：确认 token 已生效
			      await miniAppGetCategoryTree();
		      uni.showToast({ title: t("auth.loginSuccess"), icon: "none" });
		      redirectAfterLogin();
		      return;
		    }

    if (!String(form.name).trim()) {
      uni.showToast({ title: t("auth.missingFields"), icon: "none" });
      return;
    }

		    await miniAppAuthRegister({
		      name: String(form.name),
		      identifier: fullIdentifier(),
		      password: String(form.password),
		      phone: String(form.phone),
		    });
		    // 注册后同样做一次联通校验（注册接口也会返回 token）
		    await miniAppGetCategoryTree();
		    uni.showToast({ title: t("auth.signupSuccess"), icon: "none" });
		    if (redirectAfterLogin()) return;
		    activeTab.value = "login";
			  } catch (err: any) {
			    uni.showToast({ title: err?.message || t("auth.failed"), icon: "none" });
			  }
			}

      async function onWechatLogin() {
        try {
          const profile = await getWechatProfile();
          const code = await getWechatCode();
          const nickname = String(profile?.nickName || "").trim();
          const avatarUrl = String(profile?.avatarUrl || "").trim();
          const cached = getStoredWechatProfile();
          wechatCode.value = code;
          if (cached.nickname || cached.avatarUrl) {
            wechatNickname.value = cached.nickname || nickname;
            wechatAvatarUrl.value = cached.avatarUrl || avatarUrl;
          } else {
            wechatNickname.value = nickname;
            wechatAvatarUrl.value = avatarUrl;
          }
          wechatConfirmOpen.value = true;
        } catch (err: any) {
          uni.showToast({ title: err?.message || t("auth.wechatLoginFailed"), icon: "none" });
        }
      }

      function closeWechatConfirm() {
        wechatConfirmOpen.value = false;
      }

      function confirmWechatProfile() {
        if (!wechatCode.value) {
          uni.showToast({ title: t("auth.wechatLoginFailed"), icon: "none" });
          return;
        }
        const providerId = getWechatProviderId();
        if (!providerId) {
          uni.showToast({ title: t("auth.wechatLoginFailed"), icon: "none" });
          return;
        }
        wechatConfirmOpen.value = false;
        submitWechatLogin(wechatCode.value, wechatNickname.value, wechatAvatarUrl.value);
      }

      function redirectAfterLogin() {
        const redirect = String(uni.getStorageSync("miniapp.auth.redirect") || "").trim();
        if (redirect) {
          uni.removeStorageSync("miniapp.auth.redirect");
          const url = redirect.startsWith("/") ? redirect : `/${redirect}`;
          const tabPages = new Set(["/pages/index/index", "/pages/mall/index", "/pages/cart/index", "/pages/profile/index"]);
          if (tabPages.has(url)) uni.switchTab({ url });
          else uni.redirectTo({ url });
          return true;
        }
        uni.reLaunch({ url: "/pages/index/index" });
        return true;
      }

      function getWechatCode(): Promise<string> {
        return new Promise((resolve, reject) => {
          uni.login({
            provider: "weixin",
            success: (res: any) => {
              const code = String(res?.code || "").trim();
              console.log("[miniapp] wechat login code:", code);
              if (!code) return reject(new Error(t("auth.wechatLoginFailed")));
              resolve(code);
            },
            fail: (err: any) => reject(new Error(err?.errMsg || t("auth.wechatLoginFailed"))),
          });
        });
      }

      function getWechatProfile(): Promise<any | null> {
        return new Promise((resolve) => {
          const uniAny = uni as any;
          if (typeof uniAny.getUserProfile !== "function") return resolve(null);
          uniAny.getUserProfile({
            desc: t("auth.wechatProfileDesc"),
            success: (res: any) => resolve(res?.userInfo || null),
            fail: () => resolve(null),
          });
        });
      }

      function getWechatProviderId() {
        const stored = String(uni.getStorageSync("miniapp.wechat.providerId") || "").trim();
        if (stored) return stored;
        return String(import.meta.env.VITE_MINIAPP_WECHAT_PROVIDER_ID || "").trim();
      }

      function getStoredWechatProfile() {
        return {
          nickname: String(uni.getStorageSync("miniapp.wechat.profile.nickname") || "").trim(),
          avatarUrl: String(uni.getStorageSync("miniapp.wechat.profile.avatarUrl") || "").trim(),
        };
      }

      function setStoredWechatProfile(nickname: string, avatarUrl: string) {
        uni.setStorageSync("miniapp.wechat.profile.nickname", String(nickname || "").trim());
        uni.setStorageSync("miniapp.wechat.profile.avatarUrl", String(avatarUrl || "").trim());
      }

      async function submitWechatLogin(code: string, nickname: string, avatarUrl: string) {
        const providerId = getWechatProviderId();
        if (!providerId) {
          uni.showToast({ title: t("auth.wechatLoginFailed"), icon: "none" });
          return;
        }
        try {
          await miniAppAuthWechatLogin({
            providerId,
            code,
            nickname: String(nickname || "").trim(),
            avatarUrl: String(avatarUrl || "").trim(),
          });
          setStoredWechatProfile(nickname, avatarUrl);
          await miniAppGetCategoryTree();
          uni.showToast({ title: t("auth.loginSuccess"), icon: "none" });
          redirectAfterLogin();
        } catch (err: any) {
          uni.showToast({ title: err?.message || t("auth.wechatLoginFailed"), icon: "none" });
        }
      }

		onLoad((query) => {
		  try {
		    const wxAny = (globalThis as any).wx;
		    if (wxAny && typeof wxAny.getWindowInfo === "function") {
		      const info = wxAny.getWindowInfo();
      const statusBar = Number(info?.statusBarHeight ?? 0);
      topInset.value = Math.max(40, statusBar + 8);
    }
  } catch {}

	  const tab = String((query as any)?.tab || "").toLowerCase();
		  if (tab === "signup" || tab === "register") activeTab.value = "signup";
		  else activeTab.value = "login";
		});
	</script>
