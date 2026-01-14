<template>
  <view class="min-h-screen w-full bg-background-light font-display text-gray-900" style="padding-bottom: 110px;">
    <view class="sticky top-0 z-50 bg-surface-95 px-4 pb-3 shadow-sm" :style="`padding-top:${topInset}px;`">
      <view class="flex items-center">
        <view class="-ml-2 flex h-10 w-10 items-center justify-center rounded-full" hover-class="opacity-80" @tap="goBack">
          <text class="text-lg font-black">←</text>
        </view>
        <view class="flex-1 pr-8 text-center text-base font-extrabold">管理地址</view>
      </view>
    </view>

    <view class="px-4 pt-4 pb-32 space-y-3">
      <view v-if="!list.length" class="rounded-2xl bg-white p-6 text-center shadow-sm">
        <view class="text-base font-extrabold">暂无地址</view>
        <view class="mt-2 text-sm text-muted">最多可保存 20 个配送地址</view>
      </view>

      <view
        v-for="a in list"
        :key="a.id"
        class="address-card rounded-xl bg-white shadow-sm"
        :style="a.id === selectedId ? selectedCardStyle : ''"
        hover-class="opacity-95"
        @tap="onTapCard(a.id)"
      >
        <view class="p-4">
          <view class="flex items-center justify-between mb-2">
            <view class="flex items-center" style="gap: 8px;">
              <text class="text-base font-extrabold text-gray-900">{{ a.shippingAddress.recipientName }}</text>
              <text class="text-sm text-gray-500 font-semibold">{{ maskPhone(a.shippingAddress.recipientPhone) }}</text>
            </view>
            <view v-if="a.isDefault" class="flex items-center rounded bg-primary-10 px-2" style="gap: 4px; padding-top: 2px; padding-bottom: 2px;">
              <text class="text-primary" style="font-size: 12px; font-weight: 900;">✓</text>
              <text class="text-primary font-extrabold" style="font-size: 10px;">默认地址</text>
            </view>
          </view>

          <text class="text-sm text-gray-600" style="line-height: 1.6;" :number-of-lines="2">{{ formatShippingAddress(a.shippingAddress) }}</text>

          <view class="mt-4 flex items-center justify-end border-t" style="gap: 16px; border-color: rgba(0,0,0,0.04); padding-top: 12px;">
            <view class="flex items-center" style="gap: 4px;" hover-class="opacity-80" @tap.stop="openEdit(a.id)">
              <text class="text-gray-500" style="font-size: 18px; font-weight: 900;">✎</text>
              <text class="text-sm text-gray-500 font-semibold">编辑</text>
            </view>
            <view class="flex items-center" style="gap: 4px;" hover-class="opacity-80" @tap.stop="remove(a.id)">
              <text class="text-red-500" style="font-size: 18px; font-weight: 900;">🗑</text>
              <text class="text-sm text-red-500 font-semibold">删除</text>
            </view>
          </view>
        </view>
      </view>

      <view v-if="list.length" class="py-8 flex flex-col items-center justify-center" style="opacity: 0.4;">
        <text class="mb-2" style="font-size: 32px;">📍</text>
        <text class="text-xs text-muted">最多可保存 20 个配送地址</text>
      </view>
    </view>

    <view class="fixed bottom-0 left-0 right-0 z-40 bg-white border-t safe-pb" style="border-color: rgba(0,0,0,0.06);">
      <view class="px-4 py-2">
        <view class="add-btn flex items-center justify-center rounded-full bg-primary shadow-sm" hover-class="opacity-90" @tap="openCreate">
          <text class="text-lg text-white" style="font-weight: 900;">＋</text>
          <text class="ml-2 text-base font-extrabold text-white">新增地址</text>
        </view>
      </view>
    </view>

    <view v-if="sheetOpen" class="fixed inset-0 z-50">
      <view class="absolute inset-0" style="background: rgba(17, 24, 39, 0.55);" @tap="closeSheet"></view>
      <view class="absolute bottom-0 left-0 right-0 bg-white rounded-t-2xl px-4 pt-4 pb-4 safe-pb" @tap.stop>
        <view class="text-base font-extrabold text-gray-900 mb-3">{{ editingId ? "编辑地址" : "新增地址" }}</view>

        <view class="space-y-3">
          <view class="rounded-xl bg-gray-50 px-3 py-3" style="border: 1px solid rgba(0,0,0,0.04);">
            <input v-model="form.name" class="text-sm" placeholder="收货人" />
          </view>
          <view class="rounded-xl bg-gray-50 px-3 py-3" style="border: 1px solid rgba(0,0,0,0.04);">
            <input v-model="form.phone" class="text-sm" placeholder="手机号" type="number" />
          </view>
          <view class="rounded-xl bg-gray-50 px-3 py-3" style="border: 1px solid rgba(0,0,0,0.04);">
            <input v-model="form.region" class="text-sm" placeholder="省市区（如：广东省 深圳市 南山区）" />
          </view>
          <view class="rounded-xl bg-gray-50 px-3 py-3" style="border: 1px solid rgba(0,0,0,0.04);">
            <input v-model="form.detail" class="text-sm" placeholder="详细地址（街道/门牌号）" />
          </view>
          <view class="flex items-center gap-2" hover-class="opacity-80" @tap="form.isDefault = !form.isDefault">
            <view class="h-5 w-5 rounded-full border flex items-center justify-center" :style="form.isDefault ? checkedStyle : uncheckedStyle">
              <view v-if="form.isDefault" class="h-3 w-3 rounded-full bg-primary"></view>
            </view>
            <text class="text-sm text-muted font-semibold">设为默认地址</text>
          </view>
        </view>

        <view class="mt-4 flex gap-3">
          <view class="flex-1 h-11 rounded-full border flex items-center justify-center" style="border-color: rgba(0,0,0,0.08);" hover-class="opacity-90" @tap="closeSheet">
            <text class="text-sm font-extrabold text-gray-700">取消</text>
          </view>
          <view class="flex-1 h-11 rounded-full bg-primary flex items-center justify-center shadow-sm" hover-class="opacity-90" @tap="save">
            <text class="text-sm font-extrabold text-white">保存</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from "@dcloudio/uni-app";
import { onMounted, reactive, ref } from "vue";
import { getSelectedAddressId, maskPhone, setSelectedAddressId } from "@/services/address";
import { createMyAddress, deleteMyAddress, formatShippingAddress, listMyAddresses, updateMyAddress, type CustomerAddressDTO } from "@/services/miniapp-address";

const topInset = ref(44);
const from = ref<"order" | "manage">("manage");

const list = ref<CustomerAddressDTO[]>([]);
const selectedId = ref("");
const loading = ref(false);
const busy = ref(false);

const checkedStyle = "border-color: rgba(56,102,87,0.65); background: rgba(56,102,87,0.10);";
const uncheckedStyle = "border-color: rgba(209, 213, 219, 1); background: #fff;";
const selectedCardStyle = "border: 1px solid rgba(56,102,87,0.40); box-shadow: 0 10px 24px rgba(0,0,0,0.05);";

const sheetOpen = ref(false);
const editingId = ref("");
const form = reactive({
  name: "",
  phone: "",
  region: "",
  detail: "",
  isDefault: true,
});

async function load() {
  if (loading.value) return;
  loading.value = true;
  try {
    const items = await listMyAddresses();
    list.value = (items || []).slice(0, 20);
    selectedId.value = getSelectedAddressId() || (list.value.find((x) => x.isDefault)?.id || list.value[0]?.id || "");
  } catch (e: any) {
    uni.showToast({ title: e?.message || "加载地址失败", icon: "none" });
    list.value = [];
    selectedId.value = "";
  } finally {
    loading.value = false;
  }
}

function goBack() {
  uni.navigateBack();
}

function onTapCard(id: string) {
  const next = String(id || "").trim();
  if (!next) return;
  setSelectedAddressId(next);
  selectedId.value = next;
  if (from.value === "order") {
    uni.navigateBack();
    return;
  }
  uni.showToast({ title: "已选择", icon: "none" });
}

function openCreate() {
  editingId.value = "";
  form.name = "";
  form.phone = "";
  form.region = "";
  form.detail = "";
  form.isDefault = list.value.length === 0;
  sheetOpen.value = true;
}

function openEdit(id: string) {
  const a = list.value.find((x) => x.id === id);
  if (!a) return;
  editingId.value = a.id;
  form.name = a.shippingAddress.recipientName;
  form.phone = a.shippingAddress.recipientPhone;
  form.region = [a.shippingAddress.province, a.shippingAddress.city, a.shippingAddress.district].filter(Boolean).join(" ");
  form.detail = [a.shippingAddress.address1, a.shippingAddress.address2].filter(Boolean).join(" ");
  form.isDefault = Boolean(a.isDefault);
  sheetOpen.value = true;
}

function closeSheet() {
  sheetOpen.value = false;
}

function parseRegion(s: string) {
  const text = String(s || "").trim();
  if (!text) return { province: "", city: "", district: "" };
  const parts = text.split(/\s+/).filter(Boolean);
  return { province: parts[0] || "", city: parts[1] || "", district: parts.slice(2).join(" ") || "" };
}

async function save() {
  if (busy.value) return;
  if (list.value.length >= 20 && !editingId.value) {
    uni.showToast({ title: "最多可保存 20 个地址", icon: "none" });
    return;
  }
  const name = String(form.name || "").trim();
  const phone = String(form.phone || "").trim();
  const detail = String(form.detail || "").trim();
  if (!name) return uni.showToast({ title: "请填写收货人", icon: "none" });
  if (!phone) return uni.showToast({ title: "请填写手机号", icon: "none" });
  if (!detail) return uni.showToast({ title: "请填写详细地址", icon: "none" });

  const { province, city, district } = parseRegion(form.region);
  const isDefault = Boolean(form.isDefault);

  busy.value = true;
  try {
    const req = {
      isDefault,
      shippingAddress: {
        recipientName: name,
        recipientPhone: phone,
        countryCode: "CN",
        province: province || undefined,
        city: city || undefined,
        district: district || undefined,
        address1: detail,
      },
    };
    const resp = editingId.value ? await updateMyAddress(editingId.value, req) : await createMyAddress(req);
    const id = String(resp?.id || "").trim();
    if (id) {
      setSelectedAddressId(id);
      selectedId.value = id;
    }
    sheetOpen.value = false;
    await load();
    uni.showToast({ title: "已保存", icon: "none" });
    if (from.value === "order") uni.navigateBack();
  } catch (e: any) {
    uni.showToast({ title: e?.message || "保存失败", icon: "none" });
  } finally {
    busy.value = false;
  }
}

async function remove(id: string) {
  if (busy.value) return;
  const target = String(id || "").trim();
  if (!target) return;
  const ok = await new Promise<boolean>((resolve) => {
    uni.showModal({
      title: "删除地址",
      content: "确认删除该地址？",
      confirmText: "删除",
      confirmColor: "#ef4444",
      success: (res) => resolve(Boolean(res?.confirm)),
      fail: () => resolve(false),
    });
  });
  if (!ok) return;
  busy.value = true;
  try {
    await deleteMyAddress(target);
    await load();
    const curSel = getSelectedAddressId();
    if (curSel === target) {
      const fallback = list.value.find((x) => x.isDefault)?.id || list.value[0]?.id || "";
      setSelectedAddressId(fallback);
      selectedId.value = fallback;
    }
    uni.showToast({ title: "已删除", icon: "none" });
  } catch (e: any) {
    uni.showToast({ title: e?.message || "删除失败", icon: "none" });
  } finally {
    busy.value = false;
  }
}

onLoad((q: any) => {
  from.value = String(q?.from || "").toLowerCase() === "order" ? "order" : "manage";
});

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
});

onShow(() => {
  void load();
});
</script>

<style scoped>
.safe-pb {
  padding-bottom: env(safe-area-inset-bottom);
}
.add-btn {
  height: 48px;
}
.address-card {
  overflow: hidden;
  border: 1px solid rgba(0, 0, 0, 0.04);
}
</style>
