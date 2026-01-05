<template>
  <view class="min-h-screen w-full bg-background-light font-display text-text-dark" style="padding-bottom: 96px;">
    <view :style="`padding-top:${topInset}px;`" class="sticky top-0 z-10 w-full bg-background-light">
      <view class="px-4 pt-4 pb-2">
        <view class="text-xl font-extrabold">分类</view>
      </view>
    </view>
    <view class="px-4 pb-6">
      <view class="mt-2 flex w-full items-stretch overflow-hidden rounded-xl shadow-sm border border-line-5 bg-white">
        <view class="flex items-center justify-center pl-4 pr-2 text-muted">
          <text class="text-lg">⌕</text>
        </view>
        <input
          v-model="keyword"
          class="h-12 flex-1 bg-white pr-3 text-base font-medium"
          style="border: none;"
          placeholder="搜索类目..."
          @confirm="onSearch"
        />
      </view>

      <view class="mt-4">
        <view v-if="loading" class="py-6 text-center" style="color:#90a4ae;font-size:12px;">加载中...</view>
        <view v-else-if="errorMsg" class="py-6 text-center" style="color:#ef4444;font-size:12px;">{{ errorMsg }}</view>
        <view v-else-if="filtered.length === 0" class="py-6 text-center" style="color:#90a4ae;font-size:12px;">暂无类目</view>

        <view v-else class="grid" style="grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px;">
          <view
            v-for="cat in filtered"
            :key="cat.id"
            class="rounded-2xl bg-white p-4 shadow-sm"
            style="border: 1px solid rgba(0,0,0,.06);"
            hover-class="opacity-95"
            @tap="openCategory(cat)"
          >
            <view class="flex items-center justify-between">
              <view class="text-base font-extrabold">{{ cat.displayName || cat.code || "—" }}</view>
              <view v-if="cat.isFeatured" class="rounded bg-secondary-accent px-2 py-1 text-10 font-extrabold text-primary">推荐</view>
            </view>
            <view class="mt-2 text-xs text-muted">
              {{ cat.childrenCount > 0 ? `${cat.childrenCount} 个子类目` : "—" }}
            </view>
            <view class="mt-3 flex items-center justify-between">
              <view class="text-xs font-semibold text-primary">去挑选</view>
              <text class="text-primary text-xl">›</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { miniAppGetCategoryTreeItems, type MiniAppCategoryNode } from "@/services/miniapp-category";

const topInset = ref<number>(40);
const keyword = ref("");
const loading = ref(false);
const errorMsg = ref("");
const categories = ref<MiniAppCategoryNode[]>([]);
const filtered = ref<Array<MiniAppCategoryNode & { childrenCount: number }>>([]);

onMounted(() => {
  try {
    const wxAny = (globalThis as any).wx;
    if (wxAny && typeof wxAny.getWindowInfo === "function") {
      const info = wxAny.getWindowInfo();
      const statusBar = Number(info?.statusBarHeight ?? 0);
      topInset.value = Math.max(40, statusBar + 12);
      return;
    }
    const sys = uni.getSystemInfoSync();
    const statusBar = Number((sys as any).statusBarHeight ?? 0);
    topInset.value = Math.max(40, statusBar + 12);
  } catch {
    topInset.value = 40;
  }

  loadCategories();
});

function flattenTopCategories(items: MiniAppCategoryNode[]) {
  const out: Array<MiniAppCategoryNode & { childrenCount: number }> = [];
  const roots = Array.isArray(items) ? items.slice() : [];
  roots.sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
  roots.forEach((n) => {
    const childrenCount = Array.isArray(n.children) ? n.children.length : 0;
    out.push({ ...n, childrenCount });
  });
  return out;
}

function applyFilter() {
  const q = keyword.value.trim().toLowerCase();
  const all = flattenTopCategories(categories.value);
  if (!q) {
    filtered.value = all;
    return;
  }
  filtered.value = all.filter((c) => {
    const name = String(c.displayName || "").toLowerCase();
    const code = String(c.code || "").toLowerCase();
    const slug = String(c.aliasSlug || "").toLowerCase();
    return name.includes(q) || code.includes(q) || slug.includes(q);
  });
}

async function loadCategories() {
  if (loading.value) return;
  loading.value = true;
  errorMsg.value = "";
  try {
    const items = await miniAppGetCategoryTreeItems();
    categories.value = Array.isArray(items) ? items : [];
    applyFilter();
  } catch (e: any) {
    errorMsg.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  applyFilter();
}

function openCategory(cat: MiniAppCategoryNode) {
  if (!cat?.id) return;
  uni.setStorageSync("miniapp.mall.category.id", String(cat.id));
  uni.setStorageSync("miniapp.mall.category.path", String(cat.path || ""));
  uni.switchTab({ url: "/pages/mall/index" });
}
</script>
