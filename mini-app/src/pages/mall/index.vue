<template>
  <view class="min-h-screen w-full bg-background-light font-display text-text-dark" style="padding-bottom: 96px;">
    <view
      class="sticky top-0 z-30 w-full bg-white"
      :style="`padding-top:${topInset}px; box-shadow: 0 1px 3px rgba(0,0,0,0.05);`"
    >
      <view class="flex items-center gap-3 px-4 py-3">
        <view class="flex-1">
          <view class="relative overflow-hidden rounded-lg bg-background-light">
            <view class="absolute" style="left: 12px; top: 50%; transform: translateY(-50%); opacity: .7;">
              <text style="font-size: 18px;">⌕</text>
            </view>
            <input
              v-model="keyword"
              class="h-10 w-full bg-background-light pr-4"
              style="padding-left: 36px; border: none; font-size: 14px;"
              placeholder="Search products..."
              @confirm="onSearch"
            />
          </view>
        </view>
        <view class="relative flex h-10 w-10 items-center justify-center rounded-full" hover-class="opacity-90" @tap="noop">
          <text style="font-size: 18px;">🔔</text>
          <view
            class="absolute"
            style="top: 10px; right: 10px; width: 8px; height: 8px; background: #f87171; border-radius: 9999px; border: 2px solid #fff;"
          />
        </view>
      </view>

      <scroll-view scroll-x class="w-full" style="white-space: nowrap;">
        <view class="flex gap-2 px-4 pb-2">
          <view
            v-for="chip in chips"
            :key="chip.key"
            class="shrink-0 rounded-full px-4 py-2"
            :style="chip.key === activeChip ? activeChipStyle : idleChipStyle"
            @tap="activeChip = chip.key"
          >
            <text style="font-size: 12px; font-weight: 700;">{{ chip.label }}</text>
          </view>
        </view>
      </scroll-view>

      <view style="height: 12px; background: #fff; border-bottom: 1px solid rgba(0,0,0,0.06);" />
    </view>

    <view class="flex" style="min-height: calc(100vh - 160px);">
      <scroll-view scroll-y class="shrink-0" style="width: 85px; background: #eceff1; border-right: 1px solid rgba(0,0,0,0.06);">
        <view class="pb-24">
          <view
            v-for="item in sidebar"
            :key="item.key"
            class="w-full px-1"
            :style="item.key === activeSidebar ? 'background:#fff;border-left:3px solid #455a64;' : ''"
            @tap="activeSidebar = item.key"
          >
            <view class="flex flex-col items-center justify-center" style="padding: 14px 0; gap: 6px;">
              <view
                v-if="item.key === activeSidebar"
                class="flex items-center justify-center rounded-full"
                style="width: 28px; height: 28px; background: rgba(69,90,100,0.10);"
              >
                <text style="font-size: 16px; color: #455a64;">{{ item.icon }}</text>
              </view>
              <text v-else style="font-size: 16px; color: rgba(55,71,79,.55);">{{ item.icon }}</text>
              <text
                :style="`font-size:10px;font-weight:${item.key===activeSidebar?800:600};color:${item.key===activeSidebar?'#455a64':'#90a4ae'};line-height:1;`"
              >
                {{ item.label }}
              </text>
            </view>
          </view>
        </view>
      </scroll-view>

      <view class="flex-1 bg-background-light">
        <view class="px-3 pt-3">
          <view v-if="subcategories.length" class="grid" style="grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 8px;">
            <view
              v-for="c in subcategories.slice(0, 4)"
              :key="c.id"
              class="flex flex-col items-center"
              style="gap: 6px;"
              @tap="selectSubcategory(c)"
            >
              <view
                class="flex items-center justify-center rounded-xl bg-white shadow-sm"
                :style="`width:44px;height:44px;border:1px solid ${c.id===activeSubcategoryId?'rgba(69,90,100,0.35)':'rgba(0,0,0,0.06)'};`"
              >
                <text :style="`font-size:16px;color:${c.id===activeSubcategoryId?'#455a64':'rgba(55,71,79,.55)'};`">
                  {{ c.icon }}
                </text>
              </view>
              <text style="font-size: 10px; font-weight: 600; color: #37474f; line-height: 1;" :number-of-lines="1">
                {{ c.label }}
              </text>
            </view>
          </view>

          <view class="mt-4 flex items-center justify-between" style="border-bottom: 1px solid rgba(0,0,0,0.06); padding-bottom: 8px;">
            <view class="flex items-center" style="gap: 14px;">
              <text
                :style="`font-size:11px;font-weight:800;${sort==='comprehensive'?'color:#455a64;border-bottom:2px solid #455a64;':''}padding-bottom:2px;`"
                @tap="setSort('comprehensive')"
              >
                Comprehensive
              </text>
              <text
                :style="`font-size:11px;font-weight:700;color:${sort==='sales'?'#455a64':'#90a4ae'};padding-bottom:2px;`"
                @tap="setSort('sales')"
              >
                Sales
              </text>
              <view class="flex items-center" style="gap: 4px;" @tap="togglePriceSort">
                <text :style="`font-size:11px;font-weight:700;color:${sort==='price'?'#455a64':'#90a4ae'};padding-bottom:2px;`">
                  Price
                </text>
                <text style="font-size: 12px; color: rgba(144,164,174,.9);">⇅</text>
              </view>
            </view>
            <view
              class="flex items-center rounded-md bg-white"
              style="gap: 4px; padding: 6px 10px; border: 1px solid rgba(0,0,0,0.08);"
              @tap="noop"
            >
              <text style="font-size: 10px; color: #90a4ae; font-weight: 600;">Filter</text>
              <text style="font-size: 12px; color: #90a4ae;">≡</text>
            </view>
          </view>
        </view>

        <scroll-view
          scroll-y
          class="w-full"
          style="height: calc(100vh - 220px);"
          :lower-threshold="200"
          @scrolltolower="onReachBottom"
        >
          <view class="px-3 pb-28">
            <view v-if="loading" class="py-6 text-center" style="color:#90a4ae;font-size:12px;">
              加载中...
            </view>
            <view v-else-if="errorMsg" class="py-6 text-center" style="color:#ef4444;font-size:12px;">
              {{ errorMsg }}
            </view>
            <view v-else-if="products.length === 0" class="py-6 text-center" style="color:#90a4ae;font-size:12px;">
              暂无商品
            </view>
            <view class="grid" style="grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px;">
              <view
                v-for="p in products"
                :key="p.id"
                class="overflow-hidden rounded-lg bg-white"
                style="border: 1px solid rgba(0,0,0,0.06); box-shadow: 0 4px 20px -2px rgba(55, 71, 79, 0.05);"
                @tap="openProduct(p)"
              >
                <view class="relative overflow-hidden" style="aspect-ratio: 4 / 5; background: #f3f4f6;">
                  <image class="absolute inset-0 h-full w-full" mode="aspectFill" :src="p.image" @error="onProductImageError(p.id)" />
                  <view
                    v-if="p.badge"
                    class="absolute rounded-sm px-2 py-1"
                    :style="p.badgeColor === 'dark' ? 'top:6px;left:6px;background:#37474f;' : 'top:6px;left:6px;background:#455a64;'"
                  >
                    <text style="font-size: 9px; font-weight: 800; color: #fff; line-height: 1;">{{ p.badge }}</text>
                  </view>
                </view>
                <view class="flex flex-col" style="padding: 10px; gap: 8px;">
                  <text
                    style="font-size: 11px; font-weight: 700; color: #37474f; line-height: 1.3; height: 32px;"
                    :number-of-lines="2"
                  >
                    {{ p.title }}
                  </text>
                  <text style="font-size: 9px; color: #90a4ae; line-height: 1;">
                    {{ p.code }}
                  </text>
                  <view class="mt-auto flex items-end justify-between">
                    <view class="flex flex-col">
                      <text style="font-size: 13px; font-weight: 800; color: #455a64;">{{ p.price }}</text>
                      <text style="font-size: 9px; color: #90a4ae;">{{ p.meta }}</text>
                    </view>
                    <view
                      class="flex items-center justify-center rounded"
                      style="width: 24px; height: 24px; background: rgba(69,90,100,0.06);"
                      hover-class="opacity-90"
                      @tap.stop="addToCart(p)"
                    >
                      <text style="font-size: 16px; color: #455a64; font-weight: 900; line-height: 1;">+</text>
                    </view>
                  </view>
                </view>
              </view>
            </view>
            <view style="height: 40px;" />
            <view class="flex items-center justify-center" style="padding: 14px 0; color: #90a4ae; font-size: 11px;">
              <text v-if="loading">加载中...</text>
              <text v-else-if="noMore">没有更多了</text>
              <text v-else>上拉加载更多</text>
            </view>
          </view>
        </scroll-view>
      </view>
    </view>

    <!-- 商城页不需要居中浮动按钮（设计稿的 FAB 属于底部导航中间按钮形态） -->
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { miniAppGetCategoryTreeItems, type MiniAppCategoryNode } from "@/services/miniapp-category";
import { miniAppListProducts, miniAppListProductTags, type MiniAppProductTagItem } from "@/services/miniapp-product";
import { syncTabBarSelected } from "@/utils/tabbar";
import { pickPlaceholderImage } from "@/utils/product-images";

	type Chip = { key: string; label: string };
	type Sidebar = { key: string; label: string; icon: string; pathPrefix?: string; id?: string };
	type Subcategory = { id: string; label: string; icon: string; pathPrefix: string };
	type Product = {
	  id: string;
	  code: string;
	  title: string;
	  price: string;
  meta: string;
  image: string;
  badge?: string;
  badgeColor?: "dark" | "primary";
  disabled?: boolean;
  disabledReason?: string;
};

const keyword = ref("");
const topInset = ref(44);

const fallbackChips: Chip[] = [
  { key: "all", label: "All Items" },
  { key: "bestsellers", label: "Bestsellers" },
  { key: "new_arrival", label: "New Arrival" },
  { key: "flash_sale", label: "Flash Sale" },
  { key: "imported", label: "Imported" },
];
const chips = ref<Chip[]>(fallbackChips);
const tagChipsLoading = ref(false);
const suppressChipWatch = ref(false);

	const sidebar = ref<Sidebar[]>([]);
	const categoryTree = ref<MiniAppCategoryNode[]>([]);
	const subcategories = ref<Subcategory[]>([]);

	const products = ref<Product[]>([]);

const activeChip = ref("all");
	const activeSidebar = ref("");
	const activeSubcategoryId = ref("");
	const sort = ref<"comprehensive" | "sales" | "price">("comprehensive");
	const priceAsc = ref(true);

const activeChipStyle = "background:#455a64;color:#fff;box-shadow: 0 1px 3px rgba(0,0,0,0.08);";
const idleChipStyle = "background:#fff;color:#90a4ae;border:1px solid #e2e8f0;";

const loading = ref(false);
const errorMsg = ref("");
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const noMore = computed(() => total.value > 0 && products.value.length >= total.value);

const sellabilityChannel = ref(String(uni.getStorageSync("miniapp.channel") || "official").trim() || "official");
const sellabilityLocale = ref(String(uni.getStorageSync("miniapp.locale") || "zh-CN").trim() || "zh-CN");

function sellabilityReasonToText(code: string) {
  switch (String(code || "").toUpperCase()) {
    case "NO_PUBLIC_PRICE":
      return "暂无价格";
    case "OUT_OF_STOCK":
      return "缺货";
    case "CHANNEL_DISABLED":
    case "NOT_IN_AVAILABILITY_WINDOW":
    case "CHANNEL_STATUS_BLOCKED":
      return "暂不可售";
    case "SKU_NOT_ONLINE":
      return "未上架";
    default:
      return "暂不可售";
  }
}

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

function chipLabelForTag(tag: string) {
  const normalized = String(tag || "").trim().toLowerCase();
  const mapping: Record<string, string> = {
    bestsellers: "Bestsellers",
    new_arrival: "New Arrival",
    flash_sale: "Flash Sale",
    imported: "Imported",
  };
  return mapping[normalized] || tag;
}

function toChipsFromTags(items: MiniAppProductTagItem[]) {
  const cleaned = (Array.isArray(items) ? items : [])
    .map((it) => String(it?.tag || "").trim())
    .filter(Boolean);
  const unique: string[] = [];
  const seen = new Set<string>();
  cleaned.forEach((t) => {
    const k = t.toLowerCase();
    if (seen.has(k)) return;
    seen.add(k);
    unique.push(t);
  });
  const next: Chip[] = [{ key: "all", label: "All Items" }];
  unique.slice(0, 12).forEach((tag) => next.push({ key: tag.toLowerCase(), label: chipLabelForTag(tag) }));
  return next;
}

async function loadTagChips() {
  if (tagChipsLoading.value) return;
  tagChipsLoading.value = true;
  try {
    const resp = await miniAppListProductTags({
      categoryPathPrefix: currentCategoryPathPrefix() || undefined,
      limit: 50,
    });
    const next = toChipsFromTags(resp?.items || []);
    if (next.length > 1) {
      chips.value = next;
      if (!next.some((c) => c.key === activeChip.value)) {
        suppressChipWatch.value = true;
        activeChip.value = "all";
        suppressChipWatch.value = false;
      }
      return;
    }
  } catch {
    chips.value = fallbackChips;
    if (!chips.value.some((c) => c.key === activeChip.value)) {
      suppressChipWatch.value = true;
      activeChip.value = "all";
      suppressChipWatch.value = false;
    }
  } finally {
    tagChipsLoading.value = false;
  }
}

function flattenSidebar(nodes: MiniAppCategoryNode[]) {
  // mini-app/tree 返回根数组；通常为顶层类目
  const items = Array.isArray(nodes) ? nodes.slice() : [];
  items.sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
  const mapped = items.map((n, idx) => ({
    key: n.id || String(idx),
    id: n.id,
    label: n.displayName || n.code || "Category",
    icon: n.isFeatured ? "★" : "•",
    pathPrefix: n.path,
  }));
  return [{ key: "all", label: "全部商品", icon: "◎", pathPrefix: "" }, ...mapped];
}

	async function loadCategories() {
	  const items = await miniAppGetCategoryTreeItems();
	  categoryTree.value = items;
	  sidebar.value = flattenSidebar(items);
	  const storedKey = String(uni.getStorageSync("miniapp.mall.category.key") || "").trim();
	  const storedID = String(uni.getStorageSync("miniapp.mall.category.id") || "").trim();
	  if (storedKey) {
	    const hit = sidebar.value.find((s) => String(s.key || "").trim() === storedKey);
	    if (hit) activeSidebar.value = hit.key;
	  } else if (storedID) {
	    const hit = sidebar.value.find((s) => String(s.id || "").trim() === storedID);
	    if (hit) activeSidebar.value = hit.key;
	  }
	  if (!activeSidebar.value) activeSidebar.value = "all";
	  if (sidebar.value.length && !sidebar.value.find((s) => s.key === activeSidebar.value)) {
	    activeSidebar.value = "all";
	  }
	  refreshSubcategories();
	}

	function currentCategoryPathPrefix() {
	  if (activeSubcategoryId.value) {
	    const sub = subcategories.value.find((c) => c.id === activeSubcategoryId.value);
	    if (sub?.pathPrefix) return sub.pathPrefix;
	  }
	  const item = sidebar.value.find((s) => s.key === activeSidebar.value);
	  return item?.pathPrefix || "";
	}

	function refreshSubcategories() {
	  if (activeSidebar.value === "all") {
	    subcategories.value = [];
	    activeSubcategoryId.value = "";
	    return;
	  }
	  const top = categoryTree.value.find((n) => n.id === activeSidebar.value);
	  const children = Array.isArray(top?.children) ? top!.children.slice() : [];
	  children.sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
	  subcategories.value = children.map((c, idx) => ({
	    id: c.id,
	    label: c.displayName || c.code || `Category ${idx + 1}`,
	    icon: c.isFeatured ? "★" : "•",
	    pathPrefix: c.path,
	  }));
	  activeSubcategoryId.value = "";
	}

	function selectSubcategory(c: Subcategory) {
	  activeSubcategoryId.value = c.id;
	  loadTagChips().finally(() => loadProducts(true));
	}

async function loadProducts(reset: boolean) {
  if (loading.value) return;
  loading.value = true;
  errorMsg.value = "";
  try {
    const nextPage = reset ? 1 : page.value + 1;
    const resp = await miniAppListProducts({
      categoryPathPrefix: currentCategoryPathPrefix() || undefined,
      tag: activeChip.value !== "all" ? activeChip.value : undefined,
      keyword: keyword.value || undefined,
      sort: sort.value,
      order: sort.value === "price" ? (priceAsc.value ? "asc" : "desc") : "desc",
      channel: sellabilityChannel.value,
      locale: sellabilityLocale.value,
      includeSellability: 1,
      page: reset ? 1 : nextPage,
      pageSize: pageSize.value,
    });
    total.value = Number(resp?.total || 0);
    page.value = Number(resp?.page || (reset ? 1 : nextPage));
    const items = Array.isArray(resp?.items) ? resp.items : [];
    const visibleItems = items.filter((it) => String(it?.type || "").toLowerCase() !== "subscription");
    const mapped: Product[] = visibleItems.map((it) => {
      const sellability = it.sellability;
      const disabled = sellability && sellability.sellable === false;
      const reasonCode = (sellability?.reasons || [])[0] || "";
      return {
        id: it.id,
        code: it.code,
        title: it.name || it.code,
        price: String(it.priceLabel || "--"),
        meta: typeof it.skuCount === "number" && it.skuCount > 0 ? `${it.skuCount} 个规格` : "",
        image: it.coverUrl || pickPlaceholderImage(it.id),
        badge: disabled ? "暂不可售" : undefined,
        badgeColor: disabled ? "dark" : undefined,
        disabled,
        disabledReason: disabled ? sellabilityReasonToText(reasonCode) : "",
      };
    });
    products.value = reset ? mapped : products.value.concat(mapped);
  } catch (e: any) {
    errorMsg.value = e?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  loadProducts(true);
}

function setSort(next: "comprehensive" | "sales" | "price") {
  if (next !== "price") {
    if (next === "sales") {
      uni.showToast({ title: "按销量暂未支持", icon: "none" });
      return;
    }
    uni.showToast({ title: "综合排序", icon: "none" });
  }
  sort.value = next;
  loadProducts(true);
}

function togglePriceSort() {
  sort.value = "price";
  priceAsc.value = !priceAsc.value;
  uni.showToast({ title: priceAsc.value ? "价格升序" : "价格降序", icon: "none" });
  loadProducts(true);
}

function noop() {
  uni.showToast({ title: "功能待接入", icon: "none" });
}

function openProduct(p: Product) {
  if (!p?.id) return;
  uni.navigateTo({ url: `/pages/product/detail?id=${encodeURIComponent(String(p.id))}` });
}

function addToCart(p: Product) {
  if (p?.disabled) {
    uni.showToast({ title: p.disabledReason || "暂不可售", icon: "none" });
    return;
  }
  uni.showToast({ title: "请进入商品详情选择规格加入购物车", icon: "none" });
  openProduct(p);
}

function onProductImageError(productId: string) {
  const id = String(productId || "").trim();
  if (!id) return;
  // 微信小程序对外网图片域名有白名单限制：失败时回退到本地占位图，避免“空白卡片”
  products.value = products.value.map((p) => (p.id === id ? { ...p, image: "/static/icons/image-placeholder.png" } : p));
}

function onReachBottom() {
  if (loading.value) return;
  if (noMore.value) return;
  loadProducts(false);
}

onMounted(() => {
  ensureTopInset();
  loadCategories()
    .then(() => loadTagChips())
    .then(() => loadProducts(true))
    .catch((e: any) => {
      errorMsg.value = e?.message || "加载失败";
    });
});

watch(activeSidebar, () => {
  if (!activeSidebar.value) return;
  try {
    const item = sidebar.value.find((s) => s.key === activeSidebar.value);
    uni.setStorageSync("miniapp.mall.category.key", String(activeSidebar.value));
    if (item?.id) uni.setStorageSync("miniapp.mall.category.id", String(item.id));
  } catch {}
  refreshSubcategories();
  loadTagChips().finally(() => loadProducts(true));
});

watch(activeChip, () => {
  if (suppressChipWatch.value) return;
  loadProducts(true);
});

onShow(() => {
  syncTabBarSelected("pages/mall/index");
});
</script>
