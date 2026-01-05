<template>
  <view class="min-h-screen w-full bg-background-light font-display text-text-dark" style="padding-bottom: 96px;">
    <view
      :style="`padding-top:${topInset}px;`"
      class="sticky top-0 z-30 w-full bg-background-light"
      style="background-color: rgba(248, 250, 249, 0.95);"
    >
      <view class="flex items-center justify-between px-4 pt-4 pb-2">
        <view class="flex items-center gap-3">
          <view class="flex h-10 w-10 items-center justify-center rounded-full" hover-class="opacity-90" @tap="onMenu">
            <text class="text-xl font-black">≡</text>
          </view>
          <view class="text-xl font-extrabold tracking-tight">Distribution Mall</view>
        </view>
        <view class="flex items-center gap-2">
          <view class="relative flex h-10 w-10 items-center justify-center rounded-full" hover-class="opacity-90" @tap="onNotifications">
            <text class="text-xl">◉</text>
            <view class="absolute" style="top: 10px; right: 10px; width: 8px; height: 8px; background: #ef4444; border-radius: 9999px;" />
          </view>
        </view>
      </view>
    </view>

    <view class="px-4 py-2 sticky z-20 bg-background-light" :style="`top:${topInset + 64}px;`">
      <view class="flex w-full items-stretch overflow-hidden rounded-xl shadow-sm border border-line-5 bg-white">
        <view class="flex items-center justify-center pl-4 pr-2 text-muted">
          <text class="text-lg">⌕</text>
        </view>
        <input
          v-model="keyword"
          class="h-12 flex-1 bg-white pr-3 text-base font-medium"
          style="border: none;"
          placeholder="搜索商品、品牌..."
          @confirm="onSearch"
        />
      </view>
    </view>

    <view class="mt-4 pl-4 overflow-hidden">
      <scroll-view scroll-x class="w-full" style="white-space: nowrap;">
        <view class="flex gap-4 pr-4 pb-4">
          <view
            v-for="banner in banners"
            :key="banner.title"
            class="relative shrink-0 overflow-hidden rounded-2xl"
            style="width: 85vw; height: 192px;"
            @tap="onBanner(banner)"
          >
            <image class="h-full w-full" mode="aspectFill" :src="banner.image" />
            <view class="absolute inset-0" style="background: linear-gradient(to top, rgba(0,0,0,.6), rgba(0,0,0,0));" />
            <view class="absolute" style="left: 20px; bottom: 18px; right: 20px;">
              <view class="inline-block rounded-md bg-primary px-2 py-1 text-xs font-extrabold text-white">
                {{ banner.badge }}
              </view>
              <view class="mt-2 text-2xl font-extrabold text-white">{{ banner.title }}</view>
              <view class="mt-1 text-sm text-white" style="opacity: .9;">{{ banner.subtitle }}</view>
            </view>
          </view>
        </view>
      </scroll-view>
    </view>

    <view class="px-4 mt-2">
      <view class="mb-3 flex items-center justify-between">
        <view class="text-lg font-extrabold">分类</view>
        <view class="text-sm font-semibold text-primary" @tap="goCatalog">查看全部</view>
      </view>
      <view class="grid" style="grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px 8px;">
        <view
          v-for="cat in categories"
          :key="cat.key"
          class="flex flex-col items-center gap-2"
          @tap="onCategory(cat)"
        >
          <view
            class="flex items-center justify-center rounded-full bg-white shadow-sm"
            style="width: 56px; height: 56px; border: 1px solid rgba(0,0,0,.06);"
          >
            <text class="text-primary text-xl">{{ cat.icon }}</text>
          </view>
          <view class="text-xs font-semibold">{{ cat.name }}</view>
        </view>
      </view>
    </view>

    <view class="px-4 mt-8">
      <view class="mb-3 text-lg font-extrabold">特惠</view>
      <view class="grid" style="grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px;">
        <view class="rounded-xl bg-white p-3 shadow-sm" style="border: 1px solid rgba(0,0,0,.06);" @tap="noop">
          <view class="flex items-start justify-between">
            <view>
              <view class="font-extrabold">限时抢购</view>
              <view class="text-xs text-muted">剩余 02:45</view>
            </view>
            <view class="rounded bg-secondary-accent px-2 py-1 text-10 font-extrabold text-primary">-40%</view>
          </view>
          <view class="mt-2 overflow-hidden rounded-lg" style="height: 96px;">
            <image class="h-full w-full" mode="aspectFill" :src="offers[0].image" />
          </view>
        </view>
        <view class="rounded-xl bg-white p-3 shadow-sm" style="border: 1px solid rgba(0,0,0,.06);" @tap="noop">
          <view class="flex items-start justify-between">
            <view>
              <view class="font-extrabold">新品上架</view>
              <view class="text-xs text-muted">每日更新</view>
            </view>
            <view class="rounded bg-secondary-accent px-2 py-1 text-10 font-extrabold text-primary">NEW</view>
          </view>
          <view class="mt-2 overflow-hidden rounded-lg" style="height: 96px;">
            <image class="h-full w-full" mode="aspectFill" :src="offers[1].image" />
          </view>
        </view>
      </view>
    </view>

    <view class="px-4 mt-8">
      <view class="mb-3 text-lg font-extrabold">热门推荐</view>
      <view class="grid" style="grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px;">
        <view
          v-for="item in trending"
          :key="item.title"
          class="overflow-hidden rounded-xl bg-white shadow-sm"
          style="border: 1px solid rgba(0,0,0,.06);"
          @tap="noop"
        >
          <view class="overflow-hidden" :style="`height:${item.imageHeight}px;`">
            <image class="h-full w-full" mode="aspectFill" :src="item.image" />
          </view>
          <view class="p-3">
            <view class="text-sm font-semibold" style="line-height: 1.3;">{{ item.title }}</view>
            <view class="mt-2 flex items-center justify-between">
              <view class="text-base font-extrabold">{{ item.price }}</view>
              <view class="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-white" hover-class="opacity-90">
                <text class="text-base font-black">+</text>
              </view>
            </view>
          </view>
        </view>
      </view>
    </view>

  </view>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { miniAppGetCategoryTreeItems, type MiniAppCategoryNode } from "@/services/miniapp-category";

type Banner = { badge: string; title: string; subtitle: string; image: string };
type CategoryItem = { key: string; id: string; path: string; name: string; icon: string };
type Card = { title: string; price: string; image: string; imageHeight: number };

const keyword = ref("");
const topInset = ref<number>(40);

const banners = ref<Banner[]>([
  {
    badge: "热卖",
    title: "夏日精选",
    subtitle: "分销佣金最高 50%",
    image:
      "https://lh3.googleusercontent.com/aida-public/AB6AXuBJwE-an--fL5cWq8eK2uAXpwTi546F9cgIA33Qxn2UGw2yGBYoTx6KtCSwE9-PC9WdLaGPPSv3hGvaLo9wkuVSrnmNBQkZaafanL7IU6ZoM0Mb9QjoxCdDwljtG5-WJECE2aY9DY3aNo-h5RRUVZu7yDosGX3jSfqLnBNcG_S6JcjkTGVCNJPpT8YrUn77dtWS6K1t9L7JkmoPaHE3qlMmn7-eljVoCOjqesn-o7NagTu9cRFNkxq0lPathnovJgBFQxGn3rQDuOOF",
  },
  {
    badge: "上新",
    title: "数码潮品",
    subtitle: "探索未来电子产品",
    image:
      "https://lh3.googleusercontent.com/aida-public/AB6AXuARHGnuoGF0rW142GdTtjq11epf9CU4Zxx8q5XBK6w5zpSCP_LJ5zRUrW4_k0sM-FF8HAjdhz9ctFVhFIUgyugXHX1-i0m0Xa4GwAlWUbcvW54Ck30Yvi4QDmjYFW2hq28ocOzL2wJfmWyHi8OqFfTyP1uRATth3NkUTlyESZjwI7MIgOlg9YR2ZbmMz-xTOX_x_-rCJV-O9C_PuPJm0sFtysoiE4q6j2FzV8orTKiG9Y1eUf2GguBdX0r6xBNNyVZeUBmK0q1eZ7kI",
  },
]);

const categories = ref<CategoryItem[]>([]);

const offers = ref<{ image: string }[]>([
  {
    image:
      "https://lh3.googleusercontent.com/aida-public/AB6AXuBhci_cdlU3XG3BfRZb1ftV4Y01AWmfeoarvxV2G6TP_vhUXVQDLqZNTwK2X4JjVw0eDEXSYMCC2qdCeLKa1SI1DMTvI5wzhBQgUQ5t6WOR3dqVjPbMZSy06IqMrVnTk_2BLoKLasIWSLdoMekoKcrtKKqo9TzhdKWuZChL4XxLq3v73wvbCa1ThjwWZX0dku9C-2rxQTgEvI6cx-H95nabL75WWZQ0qFgbHC3mVzJJg1ks04SxS33iWBgvgeeKZoVxRzOvjKL4Dyn6",
  },
  {
    image:
      "https://lh3.googleusercontent.com/aida-public/AB6AXuB0gHw9HBH1VRXiCdF80YvME2ayL0m9R5XAdVIgWUxMpmVBjpzCn26npK_BWwzF25IG_Qqs5FWiDGFUIe_56awXFYrDm9XeOwQT-BZXUfEA9Nwpciuz5irSO0Wzp9I8NEpxiGU5cX--A2HgcDi49uIY0vYlbqYU5qSGyO7QNRmrIm2FV4SdxOyTRepESWod-1E9SoFw8EduLArO39f0zkGEPLCZpzM77hQOTX1wnmo_MisNfBpmWHy2-5FQAI8Uqq9wZpZBXWavdlMa",
  },
]);

const trending = ref<Card[]>([
  {
    title: "经典风衣 - 秋季限定",
    price: "¥129",
    image:
      "https://lh3.googleusercontent.com/aida-public/AB6AXuCcU1WvUiNZZmSST226b1fEjFip8mnAnU2msicMLnHr1MheaGdQP4HDzsZuHT946816aGd7qYwJvX-VAB8jm2uRBfzDZnEv1S12QStG8SVfb5mvZu-CH4XQH27YxiqSegQORk00DUYrp256ra2DyNXrbYfIh07xjTkzL-nq8gxiDoIZZnsieOSj87SSYNhs-4Mvb_i95zmE11z1S9C5QyvuzmjeOa1I--RWIOUkeRxu_fZHxyi3vxKu98D3E6T8hDJ3cQt9joXQTZUA",
    imageHeight: 180,
  },
  {
    title: "降噪无线耳机",
    price: "¥249",
    image:
      "https://lh3.googleusercontent.com/aida-public/AB6AXuCTlMP5E-hq2vw7PdVs-BdkJHC7B4yVzxeCWzAcrYldyIYZOC4us3l2Kk92C3_6NXj4u56U3EZLGxVrTugBxavPRg43y5zIHowP4dTYMBXeThNBxriEpMbe9WiJ9iaubWbXUuBy4Hnrb5MYsEbSBilqJc2lUYCEt_TPgaZX013TQqj15LlWvzGYRQFibEj--nEXqWphdK-AVbvCXYpeuSZ0KutuF_2L3iCLXsGecit_wyZfxCpODHSa8LXzwTra2NzZmJ28rtqrvzAG",
    imageHeight: 140,
  },
  {
    title: "智能手表 Series 7",
    price: "¥399",
    image:
      "https://lh3.googleusercontent.com/aida-public/AB6AXuAHqrYVMPfWlaKXK82e8pYwSUe8pxCZdkysVMU7x6It2l1XT1bkjULJ3CdKoZ8IK42zwIxgSB96dPdUFxA87-U0TNpy89_bWdrPOpBnWVpwo2ndsgl2S60QjY9slMK4hVSPF9ZmWiVMvHhK0mpQtGw5SUQFOnMyR-x5gATBemPxIZU7CgGfpWw3H-Fe-oq1kSiUQgYy8-bGe7BNToEgDrtSBedXoSdQKFEO4-Kw77-CgGXpA_MuLUKSOFQZKSw-OSPeyWMeURN0pxn8",
    imageHeight: 140,
  },
  {
    title: "有机绿色轻食",
    price: "¥24.5",
    image:
      "https://lh3.googleusercontent.com/aida-public/AB6AXuAdDdcctc4dGB1MhHbc2ba0DAjg8XSBkjKq2iRlb_HHpA7pwfSfXzk5S0Yyh3E0OWDzSBBojX1MaFUI2YVsylNCykDB9nKqVqRkERBpwVFfdVH_xtQi7bl7fLxZda4FEhVHsgrjO27TMG_XE-KEP8s5IPQAX6fei4Ac0MVmVOgqIhgbKrizzP7F6ZG8FryKS-9mu3iiyI9AOSJ9tg9Asb2KNd6CKc16KDsaXX-R7uaPjizl8luhVH13zcClc-sQl46xuJLVBIKx-bAQ",
    imageHeight: 180,
  },
]);

function ensureTopInset() {
  try {
    const wxAny = (globalThis as any).wx;
    if (wxAny && typeof wxAny.getWindowInfo === "function") {
      const info = wxAny.getWindowInfo();
      const statusBar = Number(info?.statusBarHeight ?? 0);
      topInset.value = Math.max(44, statusBar + 12);
      return;
    }
    const sys = uni.getSystemInfoSync();
    const statusBar = Number((sys as any).statusBarHeight ?? 0);
    topInset.value = Math.max(44, statusBar + 12);
  } catch {
    topInset.value = 44;
  }
}

function onSearch() {
  uni.showToast({ title: "搜索功能待接入", icon: "none" });
}

function goCatalog() {
  uni.navigateTo({ url: "/pages/catalog/index" });
}

function noop() {
  uni.showToast({ title: "功能待接入", icon: "none" });
}

function onMenu() {
  noop();
}

function onNotifications() {
  noop();
}

function onBanner(_banner: Banner) {
  noop();
}

function iconForCategory(node: MiniAppCategoryNode) {
  const code = String(node?.code || "").toLowerCase();
  if (node?.isFeatured) return "★";
  if (code.includes("basket") || code.includes("basketball")) return "🏀";
  if (code.includes("soccer") || code.includes("football")) return "⚽";
  if (code.includes("shoe")) return "👟";
  if (code.includes("apparel") || code.includes("clothes")) return "👕";
  if (code.includes("subscription")) return "📰";
  return "▣";
}

async function loadCategories() {
  try {
    const items = await miniAppGetCategoryTreeItems();
    const roots = Array.isArray(items) ? items.slice() : [];
    roots.sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
    categories.value = roots.slice(0, 8).map((n, idx) => ({
      key: n.id || String(idx),
      id: n.id,
      path: n.path,
      name: n.displayName || n.code || "Category",
      icon: iconForCategory(n),
    }));
  } catch {
    categories.value = [];
  }
}

function onCategory(cat: CategoryItem) {
  uni.setStorageSync("miniapp.mall.category.id", String(cat.id));
  uni.setStorageSync("miniapp.mall.category.path", String(cat.path || ""));
  uni.switchTab({ url: "/pages/mall/index" });
}

onMounted(() => {
  ensureTopInset();
  loadCategories();
});

onShow(() => {
  // 首页游客态可浏览；登录后可扩展分销/佣金/会员权益等能力
});
</script>
