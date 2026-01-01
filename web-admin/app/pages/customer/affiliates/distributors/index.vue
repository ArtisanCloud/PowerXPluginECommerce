<template>
  <div class="p-6">
    <!-- 标题与操作 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">分销员管理</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">管理分销员申请、状态和业绩</p>
      </div>
      <div class="flex gap-3">
        <UButton color="primary" icon="i-heroicons-plus" @click="exportData">导出数据</UButton>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-user-group" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">总人数</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ totalDistributors.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-check-badge" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">已启用</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ activeDistributors.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
            <UIcon name="i-heroicons-clock" class="w-6 h-6 text-yellow-600 dark:text-yellow-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">待审核</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ pendingDistributors.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
            <UIcon name="i-heroicons-x-circle" class="w-6 h-6 text-red-600 dark:text-red-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">已禁用</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ inactiveDistributors.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- 筛选 -->
    <UCard class="mb-6">
      <div class="flex flex-wrap items-center gap-3">
        <UInput
          v-model="searchQuery"
          placeholder="搜索：姓名 / 手机号 / ID"
          icon="i-heroicons-magnifying-glass"
          size="sm"
        />
        <USelectMenu
          v-model="statusFilter"
          :options="statusOptions"
          size="sm"
          placeholder="状态筛选"
        />
        <UButton color="neutral" @click="resetFilters">重置</UButton>
      </div>
    </UCard>

    <!-- 表格 -->
    <UCard>
      <UTable :columns="columns" :data="distributorTableData">
        <!-- 分销员 -->
        <template #distributor-cell="{ row }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="getRowData(row).avatar || avatarPlaceholder"
              :alt="getRowData(row).name || '分销员'"
              size="md"
              :ui="{ rounded: 'rounded-full' }"
            />
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ getRowData(row).name || '—' }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                ID: {{ getRowData(row).id || '—' }}
              </div>
            </div>
          </div>
        </template>

        <!-- 状态 -->
        <template #status-cell="{ row, getValue }">
          <UBadge
            :color="getStatusColor(getValue(row))"
            variant="soft"
          >
            {{ getStatusText(getValue(row)) }}
          </UBadge>
        </template>

        <!-- 申请时间 -->
        <template #createdAt-cell="{ row, getValue }">
          <div class="text-gray-900 dark:text-white">
            {{ formatDate(getValue(row)) }}
          </div>
        </template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="viewDistributorDetails(getRowData(row))"
            >
              查看
            </UButton>
            <UButton
              v-if="getRowData(row).status === 'pending'"
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-check"
              @click="approveDistributor(getRowData(row))"
            >
              审核通过
            </UButton>
            <UButton
              v-else-if="getRowData(row).status !== 'inactive'"
              color="orange"
              variant="ghost"
              size="sm"
              icon="i-heroicons-x-mark"
              @click="disableDistributor(getRowData(row))"
            >
              禁用
            </UButton>
            <UButton
              v-else-if="getRowData(row).status === 'inactive'"
              color="green"
              variant="ghost"
              size="sm"
              icon="i-heroicons-check"
              @click="enableDistributor(getRowData(row))"
            >
              启用
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 分页 -->
    <div class="flex items-center justify-between mt-6">
      <div class="text-sm text-gray-500 dark:text-gray-400">
        显示第 {{ (currentPage - 1) * pageSize + 1 }}
        到 {{ Math.min(currentPage * pageSize, filteredDistributors.length) }} 条，共 {{ filteredDistributors.length }} 条
      </div>
      <UPagination
        v-model="currentPage"
        :page-count="pageCount"
        :total="filteredDistributors.length"
        :ui="{ rounded: 'rounded-full' }"
      />

      <div class="flex items-center gap-2">
        <span class="text-sm text-gray-500 dark:text-gray-400">每页</span>
        <USelect
          v-model="pageSize"
          :options="[5, 10, 20, 50]"
          size="sm"
          class="w-20"
        />
        <span class="text-sm text-gray-500 dark:text-gray-400">条</span>
      </div>
    </div>

    <!-- 详情模态框 -->
    <UModal
      v-model:open="showDistributorModal"
      title="分销员详情"
      description="查看分销员的完整信息"
      :close="{ onClick: () => closeDistributorModal() }"
      :ui="{
        content: 'w-full sm:max-w-3xl',
        body: 'p-0',
        footer: 'justify-end',
      }"
    >
      <!-- Body -->
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-6 p-4 sm:p-6">
            <!-- 分销员基本信息 -->
            <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                分销员信息
              </h3>

              <div class="flex items-center gap-4">
                <UAvatar
                  :src="currentDistributor?.avatar || ''"
                  :alt="currentDistributor?.name || ''"
                  size="lg"
                  :ui="{ rounded: 'rounded-full' }"
                />
                <div>
                  <div class="font-medium text-gray-900 dark:text-white text-lg">
                    {{ currentDistributor?.name || '—' }}
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">
                    ID: {{ currentDistributor?.id || '—' }}
                  </div>
                </div>
              </div>

              <div class="mt-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <span class="text-gray-500 dark:text-gray-400">手机号：</span>
                  <span class="text-gray-900 dark:text-white">{{ currentDistributor?.phone || '—' }}</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">状态：</span>
                  <UBadge
                    :color="getStatusColor(currentDistributor?.status)"
                    variant="soft"
                    size="sm"
                  >
                    {{ getStatusText(currentDistributor?.status) }}
                  </UBadge>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">邮箱：</span>
                  <span class="text-gray-900 dark:text-white">{{ currentDistributor?.email || '—' }}</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">微信号：</span>
                  <span class="text-gray-900 dark:text-white">{{ currentDistributor?.wechat || '—' }}</span>
                </div>
              </div>
            </div>

            <!-- 业绩信息 -->
            <div class="border-b border-gray-200 dark:border-gray-700 pb-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                业绩信息
              </h3>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg text-center">
                  <div class="text-2xl font-bold text-gray-900 dark:text-white">
                    ¥{{ currentDistributor?.totalSales?.toLocaleString() || 0 }}
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">总销售额</div>
                </div>
                <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg text-center">
                  <div class="text-2xl font-bold text-gray-900 dark:text-white">
                    ¥{{ currentDistributor?.totalCommission?.toLocaleString() || 0 }}
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">总佣金</div>
                </div>
                <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg text-center">
                  <div class="text-2xl font-bold text-gray-900 dark:text-white">
                    {{ currentDistributor?.invitedCount || 0 }}
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">邀请人数</div>
                </div>
              </div>
            </div>

            <!-- 推广信息 -->
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                推广信息
              </h3>

              <div class="space-y-4">
                <div>
                  <span class="text-gray-500 dark:text-gray-400">推广码：</span>
                  <span class="text-gray-900 dark:text-white font-mono">{{ currentDistributor?.promotionCode || '—' }}</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">推广链接：</span>
                  <span class="text-gray-900 dark:text-white break-all">{{ currentDistributor?.promotionLink || '—' }}</span>
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <!-- Footer -->
      <template #footer>
        <div class="flex gap-3">
          <UButton variant="ghost" @click="closeDistributorModal()">关闭</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
// Nuxt3 如已开启 components: true，可不手动 import 子组件，直接模板里写 <ViewDistributorModal />
// 若不是 Nuxt，请把 ~/ 改成 @/ ，或用相对路径。
// import ViewDistributorModal from "~/components/modals/ViewDistributorModal.vue";

/* 搜索 & 筛选 */
const searchQuery = ref("");
const statusFilter = ref<"all" | "pending" | "active" | "inactive">("all");

/* 分页 */
const currentPage = ref(1);
const pageSize = ref(10);

/* 模态框 */
const showDistributorModal = ref(false);
const currentDistributor = ref<any>(null);

// 统一控制打开（本地 & overlay）
const isOpen = computed({
  get: () => showDistributorModal.value,
  set: (v) => showDistributorModal.value = v,
});

const avatarPlaceholder = "https://api.dicebear.com/7.x/miniavs/svg?seed=placeholder";

/* 选项 */
const statusOptions = [
  { label: "全部", value: "all" },
  { label: "待审核", value: "pending" },
  { label: "已启用", value: "active" },
  { label: "已禁用", value: "inactive" }
];

/* 表格列定义 */
const columns = [
  { accessorKey: "distributor", header: "分销员" },
  { accessorKey: "phone", header: "手机号" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "申请时间" },
  { id: "actions", header: "操作" },
];

/* 模拟数据（与你提供的一致） */
const distributors = ref([
  { id:"dist_1", name:"张三", phone:"13800138001", avatar:"https://api.dicebear.com/7.x/miniavs/svg?seed=1", status:"pending", createdAt:"2023-09-15 14:30:25", email:"zhangsan@example.com", idCard:"110101199001011234", address:"北京市朝阳区某某街道", wechat:"zhangsan123", promotionCode:"PROMO20230915", promotionLink:"https://example.com/promo?code=PROMO20230915", totalSales:12500, totalCommission:1250, invitedCount:12 },
  { id:"dist_2", name:"李四", phone:"13800138002", avatar:"https://api.dicebear.com/7.x/miniavs/svg?seed=2", status:"active", createdAt:"2023-09-12 10:15:42", email:"lisi@example.com", idCard:"110101199001011235", address:"上海市浦东新区某某街道", wechat:"lisi456", promotionCode:"PROMO20230912", promotionLink:"https://example.com/promo?code=PROMO20230912", totalSales:8600, totalCommission:860, invitedCount:8 },
  { id:"dist_3", name:"王五", phone:"13800138003", avatar:"https://api.dicebear.com/7.x/miniavs/svg?seed=3", status:"active", createdAt:"2023-09-10 08:45:17", email:"wangwu@example.com", idCard:"110101199001011236", address:"广州市天河区某某街道", wechat:"wangwu789", promotionCode:"PROMO20230910", promotionLink:"https://example.com/promo?code=PROMO20230910", totalSales:15200, totalCommission:1520, invitedCount:15 },
  { id:"dist_4", name:"赵六", phone:"13800138004", avatar:"https://api.dicebear.com/7.x/miniavs/svg?seed=4", status:"inactive", createdAt:"2023-09-08 16:22:33", email:"zhaoliu@example.com", idCard:"110101199001011237", address:"深圳市南山区某某街道", wechat:"zhaoliu101", promotionCode:"PROMO20230908", promotionLink:"https://example.com/promo?code=PROMO20230908", totalSales:3200, totalCommission:320, invitedCount:3 },
  { id:"dist_5", name:"孙七", phone:"13800138005", avatar:"https://api.dicebear.com/7.x/miniavs/svg?seed=5", status:"pending", createdAt:"2023-09-05 11:38:56", email:"sunqi@example.com", idCard:"110101199001011238", address:"杭州市西湖区某某街道", wechat:"sunqi202", promotionCode:"PROMO20230905", promotionLink:"https://example.com/promo?code=PROMO20230905", totalSales:0, totalCommission:0, invitedCount:0 }
]);

/* 过滤 */
const filteredDistributors = computed(() => {
  let result = distributors.value;
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase();
    result = result.filter(d =>
      (d.name || "").toLowerCase().includes(q) ||
      (d.phone || "").includes(searchQuery.value) ||
      (d.id || "").toLowerCase().includes(q)
    );
  }
  if (statusFilter.value !== "all") {
    result = result.filter(d => d.status === statusFilter.value);
  }
  return result;
});

/* 分页数据 & 页数 */
const distributorTableData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return filteredDistributors.value.slice(start, end);
});
const pageCount = computed(() =>
  Math.max(1, Math.ceil(filteredDistributors.value.length / pageSize.value))
);

/* 统计 */
const totalDistributors = computed(() => distributors.value.length);
const activeDistributors = computed(() => distributors.value.filter(d => d.status === "active").length);
const pendingDistributors = computed(() => distributors.value.filter(d => d.status === "pending").length);
const inactiveDistributors = computed(() => distributors.value.filter(d => d.status === "inactive").length);

/* 详情模态框中的计算属性 */
const totalSales = computed(() => currentDistributor.value?.totalSales || 0);
const totalCommission = computed(() => currentDistributor.value?.totalCommission || 0);
const invitedCount = computed(() => currentDistributor.value?.invitedCount || 0);

/* 状态样式 & 文本 */
const getStatusColor = (status: string) => {
  switch (status) {
    case "active": return "success";
    case "pending": return "warning";
    case "inactive": return "error";
    default: return "neutral";
  }
};
const getStatusText = (status: string) => {
  switch (status) {
    case "active": return "已启用";
    case "pending": return "待审核";
    case "inactive": return "已禁用";
    default: return "未知状态";
  }
};

/* 表格辅助方法 */
const getRowData = (row: any) => {
  return row?.original ?? row;
};

/* 日期格式化 */
const formatDate = (dateString: string) => {
  if (!dateString) return "";
  const isoLike = dateString.replace(" ", "T");
  const d = new Date(isoLike);
  return isNaN(d.getTime()) ? "" : d.toLocaleDateString("zh-CN");
};

/* 交互 */
const resetFilters = () => { searchQuery.value = ""; statusFilter.value = "all"; currentPage.value = 1; };
const approveDistributor = (d: any) => { d.status = "active"; alert(`分销员 ${d.name} 已审核通过并启用`); };
const disableDistributor = (d: any) => { d.status = "inactive"; alert(`分销员 ${d.name} 已禁用`); };
const enableDistributor = (d: any) => { d.status = "active"; alert(`分销员 ${d.name} 已启用`); };
const viewDistributorDetails = (d: any) => {
  currentDistributor.value = d;
  showDistributorModal.value = true;
};

function closeDistributorModal() {
  showDistributorModal.value = false;
  currentDistributor.value = null;
}
const exportData = () => { alert("导出数据功能待实现"); };
</script>
