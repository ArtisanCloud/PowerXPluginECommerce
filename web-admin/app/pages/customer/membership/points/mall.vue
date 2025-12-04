<template>
  <div class="p-6">
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">积分商城</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">使用积分兑换礼品和优惠券</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-plus" @click="exportData">导出数据</UButton>
      </div>
    </div>

    <!-- 积分统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-gift" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">我的积分</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ userPoints.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-ticket" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">累计兑换</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ totalExchanged.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon name="i-heroicons-arrow-trending-up" class="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">本月新增</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ monthlyPoints.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 标签页 -->
    <div class="mb-6">
      <UTabs v-model="activeTab" :items="tabs" />
    </div>

    <!-- 兑换礼品 -->
    <UCard v-if="activeTab === 'gifts'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">兑换礼品</h2>
          <UInput v-model="giftSearchQuery" placeholder="搜索礼品..." icon="i-heroicons-magnifying-glass" size="sm" />
        </div>
      </template>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        <UCard 
          v-for="gift in giftList" 
          :key="gift.id" 
          :ui="{ body: { padding: 'p-4' }, ring: '' }"
          class="hover:shadow-lg transition-shadow"
        >
          <div class="flex flex-col h-full">
            <div class="aspect-square rounded-lg overflow-hidden mb-4 bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
              <img 
                :src="gift.image || 'https://placehold.co/300x300?text=Gift'" 
                :alt="gift.name" 
                class="w-full h-full object-cover"
              />
            </div>
            <div class="flex-grow">
              <h3 class="font-medium text-gray-900 dark:text-white mb-1">{{ gift.name }}</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400 mb-3 line-clamp-2">{{ gift.description }}</p>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <div class="flex items-center">
                <UIcon name="i-heroicons-star" class="w-5 h-5 text-yellow-500" />
                <span class="ml-1 font-medium text-gray-900 dark:text-white">{{ gift.points }}</span>
              </div>
              <UButton 
                color="primary" 
                size="sm" 
                @click="exchangeGift(gift)"
                :disabled="userPoints < gift.points"
              >
                兑换
              </UButton>
            </div>
          </div>
        </UCard>
      </div>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (giftCurrentPage - 1) * giftPageSize + 1 }}
            到 {{ Math.min(giftCurrentPage * giftPageSize, totalGifts) }} 条，共 {{ totalGifts }} 条
          </div>
          <UPagination 
            v-model="giftCurrentPage" 
            :page-count="giftPageCount" 
            :total="totalGifts" 
            :ui="{ rounded: 'rounded-full' }" 
          />
        </div>
      </template>
    </UCard>

    <!-- 积分换券 -->
    <UCard v-if="activeTab === 'coupons'" class="mb-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">积分换券</h2>
          <UInput v-model="couponSearchQuery" placeholder="搜索优惠券..." icon="i-heroicons-magnifying-glass" size="sm" />
        </div>
      </template>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <UCard 
          v-for="coupon in couponList" 
          :key="coupon.id" 
          :ui="{ body: { padding: 'p-4' }, ring: '' }"
          class="hover:shadow-lg transition-shadow"
        >
          <div class="flex">
            <div class="flex-shrink-0 w-16 h-16 rounded-lg bg-red-100 dark:bg-red-900 flex items-center justify-center">
              <UIcon name="i-heroicons-ticket" class="w-8 h-8 text-red-600 dark:text-red-400" />
            </div>
            <div class="ml-4 flex-grow">
              <h3 class="font-medium text-gray-900 dark:text-white">{{ coupon.name }}</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ coupon.description }}</p>
              <div class="mt-2 flex items-center justify-between">
                <div class="flex items-center">
                  <UIcon name="i-heroicons-star" class="w-4 h-4 text-yellow-500" />
                  <span class="ml-1 text-sm font-medium text-gray-900 dark:text-white">{{ coupon.points }}</span>
                </div>
                <UButton 
                  color="primary" 
                  size="sm" 
                  @click="exchangeCoupon(coupon)"
                  :disabled="userPoints < coupon.points"
                >
                  兑换
                </UButton>
              </div>
            </div>
          </div>
        </UCard>
      </div>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (couponCurrentPage - 1) * couponPageSize + 1 }}
            到 {{ Math.min(couponCurrentPage * couponPageSize, totalCoupons) }} 条，共 {{ totalCoupons }} 条
          </div>
          <UPagination 
            v-model="couponCurrentPage" 
            :page-count="couponPageCount" 
            :total="totalCoupons" 
            :ui="{ rounded: 'rounded-full' }" 
          />
        </div>
      </template>
    </UCard>

    <!-- 兑换记录 -->
    <UCard v-if="activeTab === 'history'">
      <template #header>
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">兑换记录</h2>
          <UInput v-model="historySearchQuery" placeholder="搜索兑换记录..." icon="i-heroicons-magnifying-glass" size="sm" />
        </div>
      </template>

      <UTable :columns="historyColumns" :data="historyTableData">
        <!-- 礼品/优惠券 -->
        <template #item-cell="{ row, getValue }">
          <div class="flex items-center gap-3">
            <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
              <UIcon 
                :name="getValue(row)?.type === 'gift' ? 'i-heroicons-gift' : 'i-heroicons-ticket'" 
                class="w-5 h-5 text-gray-600 dark:text-gray-400" 
              />
            </div>
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ getValue(row)?.name || '—' }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ getValue(row)?.type === 'gift' ? '礼品' : '优惠券' }}
              </div>
            </div>
          </div>
        </template>

        <!-- 积分消耗 -->
        <template #points-cell="{ row, getValue }">
          <div class="flex items-center">
            <UIcon name="i-heroicons-star" class="w-4 h-4 text-yellow-500 mr-1" />
            <span class="font-medium text-gray-900 dark:text-white">{{ getValue(row) }}</span>
          </div>
        </template>

        <!-- 兑换时间 -->
        <template #createdAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ formatDate(getValue(row)) }}
          </div>
        </template>

        <!-- 状态 -->
        <template #status-cell="{ row, getValue }">
          <UBadge 
            :color="getValue(row) === 'completed' ? 'success' : getValue(row) === 'pending' ? 'warning' : 'error'"
            variant="soft"
          >
            {{ getValue(row) === 'completed' ? '已完成' : getValue(row) === 'pending' ? '处理中' : '已取消' }}
          </UBadge>
        </template>
      </UTable>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            显示第 {{ (historyCurrentPage - 1) * historyPageSize + 1 }}
            到 {{ Math.min(historyCurrentPage * historyPageSize, totalHistory) }} 条，共 {{ totalHistory }} 条
          </div>
          <UPagination 
            v-model="historyCurrentPage" 
            :page-count="historyPageCount" 
            :total="totalHistory" 
            :ui="{ rounded: 'rounded-full' }" 
          />
        </div>
      </template>
    </UCard>

    <!-- 兑换礼品确认弹窗 -->
    <UModal
      v-model:open="showGiftModal"
      title="兑换礼品确认"
      description="确认兑换以下礼品"
      :close="{ onClick: () => showGiftModal = false }"
      :ui="{
        content: 'w-full sm:max-w-md',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-4 p-4 sm:p-6">
            <div class="flex items-center gap-4">
              <div class="aspect-square w-16 rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
                <img 
                  :src="selectedGift?.image || 'https://placehold.co/300x300?text=Gift'" 
                  :alt="selectedGift?.name" 
                  class="w-full h-full object-cover"
                />
              </div>
              <div>
                <h3 class="font-medium text-gray-900 dark:text-white">{{ selectedGift?.name }}</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ selectedGift?.description }}</p>
              </div>
            </div>
            
            <div class="border-t border-gray-200 dark:border-gray-700 pt-4">
              <div class="flex justify-between items-center mb-2">
                <span class="text-gray-600 dark:text-gray-400">所需积分</span>
                <div class="flex items-center">
                  <UIcon name="i-heroicons-star" class="w-4 h-4 text-yellow-500 mr-1" />
                  <span class="font-medium text-gray-900 dark:text-white">{{ selectedGift?.points }}</span>
                </div>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-gray-600 dark:text-gray-400">剩余积分</span>
                <div class="flex items-center">
                  <UIcon name="i-heroicons-star" class="w-4 h-4 text-yellow-500 mr-1" />
                  <span class="font-medium text-gray-900 dark:text-white">{{ userPoints }}</span>
                </div>
              </div>
            </div>
            
            <div class="bg-blue-50 dark:bg-blue-950 rounded-lg p-3">
              <div class="flex items-start gap-2">
                <UIcon name="i-heroicons-information-circle" class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5" />
                <div class="text-sm text-blue-700 dark:text-blue-300">
                  <strong>兑换说明：</strong>
                  兑换成功后，礼品将发送到您的账户，请注意查收。
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>
      
      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="showGiftModal = false">取消</UButton>
          <UButton color="primary" @click="confirmExchangeGift">确认兑换</UButton>
        </div>
      </template>
    </UModal>

    <!-- 兑换优惠券确认弹窗 -->
    <UModal
      v-model:open="showCouponModal"
      title="兑换优惠券确认"
      description="确认兑换以下优惠券"
      :close="{ onClick: () => showCouponModal = false }"
      :ui="{
        content: 'w-full sm:max-w-md',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-4 p-4 sm:p-6">
            <div class="flex">
              <div class="flex-shrink-0 w-12 h-12 rounded-lg bg-red-100 dark:bg-red-900 flex items-center justify-center">
                <UIcon name="i-heroicons-ticket" class="w-6 h-6 text-red-600 dark:text-red-400" />
              </div>
              <div class="ml-4">
                <h3 class="font-medium text-gray-900 dark:text-white">{{ selectedCoupon?.name }}</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ selectedCoupon?.description }}</p>
              </div>
            </div>
            
            <div class="border-t border-gray-200 dark:border-gray-700 pt-4">
              <div class="flex justify-between items-center mb-2">
                <span class="text-gray-600 dark:text-gray-400">所需积分</span>
                <div class="flex items-center">
                  <UIcon name="i-heroicons-star" class="w-4 h-4 text-yellow-500 mr-1" />
                  <span class="font-medium text-gray-900 dark:text-white">{{ selectedCoupon?.points }}</span>
                </div>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-gray-600 dark:text-gray-400">剩余积分</span>
                <div class="flex items-center">
                  <UIcon name="i-heroicons-star" class="w-4 h-4 text-yellow-500 mr-1" />
                  <span class="font-medium text-gray-900 dark:text-white">{{ userPoints }}</span>
                </div>
              </div>
            </div>
            
            <div class="bg-blue-50 dark:bg-blue-950 rounded-lg p-3">
              <div class="flex items-start gap-2">
                <UIcon name="i-heroicons-information-circle" class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5" />
                <div class="text-sm text-blue-700 dark:text-blue-300">
                  <strong>兑换说明：</strong>
                  兑换成功后，优惠券将发送到您的账户，请在有效期内使用。
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>
      
      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="showCouponModal = false">取消</UButton>
          <UButton color="primary" @click="confirmExchangeCoupon">确认兑换</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";

// 当前激活的标签页
const activeTab = ref("gifts");

// 搜索查询
const giftSearchQuery = ref("");
const couponSearchQuery = ref("");
const historySearchQuery = ref("");

// 分页
const giftCurrentPage = ref(1);
const giftPageSize = ref(8);
const totalGifts = ref(24);
const giftPageCount = computed(() => Math.ceil(totalGifts.value / giftPageSize.value));

const couponCurrentPage = ref(1);
const couponPageSize = ref(6);
const totalCoupons = ref(12);
const couponPageCount = computed(() => Math.ceil(totalCoupons.value / couponPageSize.value));

const historyCurrentPage = ref(1);
const historyPageSize = ref(10);
const totalHistory = ref(45);
const historyPageCount = computed(() => Math.ceil(totalHistory.value / historyPageSize.value));

// 模态框
const showGiftModal = ref(false);
const showCouponModal = ref(false);
const selectedGift = ref<any>(null);
const selectedCoupon = ref<any>(null);

// Tabs
const tabs = [
  { label: "兑换礼品", value: "gifts" },
  { label: "积分换券", value: "coupons" },
  { label: "兑换记录", value: "history" },
];

// 用户积分数据
const userPoints = ref(2560);
const totalExchanged = ref(1200);
const monthlyPoints = ref(320);

// 礼品列表数据
const gifts = ref([
  { id: "gift_1", name: "蓝牙耳机", description: "高品质无线蓝牙耳机，音质清晰", points: 800, image: "https://placehold.co/300x300?text=蓝牙耳机" },
  { id: "gift_2", name: "智能手环", description: "健康监测智能手环，运动计步", points: 1200, image: "https://placehold.co/300x300?text=智能手环" },
  { id: "gift_3", name: "咖啡杯", description: "保温咖啡杯，350ml容量", points: 200, image: "https://placehold.co/300x300?text=咖啡杯" },
  { id: "gift_4", name: "移动电源", description: "10000mAh大容量移动电源", points: 600, image: "https://placehold.co/300x300?text=移动电源" },
  { id: "gift_5", name: "无线鼠标", description: "人体工学无线鼠标", points: 400, image: "https://placehold.co/300x300?text=无线鼠标" },
  { id: "gift_6", name: "定制T恤", description: "纯棉定制T恤，多种颜色可选", points: 300, image: "https://placehold.co/300x300?text=定制T恤" },
  { id: "gift_7", name: "保温饭盒", description: "不锈钢保温饭盒，便携设计", points: 500, image: "https://placehold.co/300x300?text=保温饭盒" },
  { id: "gift_8", name: "蓝牙音箱", description: "便携式蓝牙音箱，音质出色", points: 1000, image: "https://placehold.co/300x300?text=蓝牙音箱" },
]);

// 优惠券列表数据
const coupons = ref([
  { id: "coupon_1", name: "满100减20优惠券", description: "无门槛使用，全场通用", points: 200 },
  { id: "coupon_2", name: "满200减50优惠券", description: "限定商品使用", points: 500 },
  { id: "coupon_3", name: "满500减150优惠券", description: "高端商品专享", points: 1500 },
  { id: "coupon_4", name: "9折优惠券", description: "全场商品9折优惠", points: 800 },
  { id: "coupon_5", name: "包邮券", description: "无门槛包邮券", points: 300 },
  { id: "coupon_6", name: "满1000减300优惠券", description: "大额优惠券", points: 3000 },
]);

// 兑换历史数据
const exchangeHistory = ref([
  { id: "hist_1", name: "蓝牙耳机", type: "gift", points: 800, createdAt: "2023-09-15 14:30:25", status: "completed" },
  { id: "hist_2", name: "满200减50优惠券", type: "coupon", points: 500, createdAt: "2023-09-12 10:15:42", status: "completed" },
  { id: "hist_3", name: "咖啡杯", type: "gift", points: 200, createdAt: "2023-09-10 08:45:17", status: "completed" },
  { id: "hist_4", name: "满100减20优惠券", type: "coupon", points: 200, createdAt: "2023-09-08 16:22:33", status: "pending" },
  { id: "hist_5", name: "9折优惠券", type: "coupon", points: 800, createdAt: "2023-09-05 11:38:56", status: "completed" },
  { id: "hist_6", name: "智能手环", type: "gift", points: 1200, createdAt: "2023-09-01 09:25:14", status: "cancelled" },
]);

// 过滤礼品列表
const giftList = computed(() => {
  if (!giftSearchQuery.value) return gifts.value;
  const q = giftSearchQuery.value.toLowerCase();
  return gifts.value.filter(gift => 
    gift.name.toLowerCase().includes(q) || 
    gift.description.toLowerCase().includes(q)
  );
});

// 过滤优惠券列表
const couponList = computed(() => {
  if (!couponSearchQuery.value) return coupons.value;
  const q = couponSearchQuery.value.toLowerCase();
  return coupons.value.filter(coupon => 
    coupon.name.toLowerCase().includes(q) || 
    coupon.description.toLowerCase().includes(q)
  );
});

// 过滤历史记录
const filteredHistory = computed(() => {
  if (!historySearchQuery.value) return exchangeHistory.value;
  const q = historySearchQuery.value.toLowerCase();
  return exchangeHistory.value.filter(record => 
    record.name.toLowerCase().includes(q)
  );
});

// 历史记录表格数据
const historyTableData = computed(() => {
  const start = (historyCurrentPage.value - 1) * historyPageSize.value;
  const end = start + historyPageSize.value;
  return filteredHistory.value.slice(start, end);
});

// 历史记录表格列
const historyColumns = [
  { accessorKey: "item", header: "礼品/优惠券" },
  { accessorKey: "points", header: "积分消耗" },
  { accessorKey: "createdAt", header: "兑换时间" },
  { accessorKey: "status", header: "状态" },
];

// 日期格式化
const formatDate = (dateString: string) => {
  if (!dateString) return "";
  const isoLike = dateString.replace(" ", "T");
  const d = new Date(isoLike);
  return isNaN(d.getTime()) ? "" : d.toLocaleDateString("zh-CN");
};

// 兑换礼品
const exchangeGift = (gift: any) => {
  selectedGift.value = gift;
  showGiftModal.value = true;
};

// 确认兑换礼品
const confirmExchangeGift = () => {
  if (userPoints.value >= selectedGift.value.points) {
    userPoints.value -= selectedGift.value.points;
    totalExchanged.value += selectedGift.value.points;
    
    // 添加到兑换历史
    exchangeHistory.value.unshift({
      id: `hist_${Date.now()}`,
      name: selectedGift.value.name,
      type: "gift",
      points: selectedGift.value.points,
      createdAt: new Date().toISOString().replace("T", " ").substring(0, 19),
      status: "pending"
    });
    
    showGiftModal.value = false;
    alert(`兑换成功！${selectedGift.value.name}已发送到您的账户。`);
  } else {
    alert("积分不足，无法兑换该礼品。");
  }
};

// 兑换优惠券
const exchangeCoupon = (coupon: any) => {
  selectedCoupon.value = coupon;
  showCouponModal.value = true;
};

// 确认兑换优惠券
const confirmExchangeCoupon = () => {
  if (userPoints.value >= selectedCoupon.value.points) {
    userPoints.value -= selectedCoupon.value.points;
    totalExchanged.value += selectedCoupon.value.points;
    
    // 添加到兑换历史
    exchangeHistory.value.unshift({
      id: `hist_${Date.now()}`,
      name: selectedCoupon.value.name,
      type: "coupon",
      points: selectedCoupon.value.points,
      createdAt: new Date().toISOString().replace("T", " ").substring(0, 19),
      status: "pending"
    });
    
    showCouponModal.value = false;
    alert(`兑换成功！${selectedCoupon.value.name}已发送到您的账户。`);
  } else {
    alert("积分不足，无法兑换该优惠券。");
  }
};

// 工具函数
const exportData = () => alert("导出数据功能待实现");
</script>