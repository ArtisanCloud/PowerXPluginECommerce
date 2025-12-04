<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">客户管理</h1>
      <UButton color="primary" icon="i-heroicons-plus" @click="openCreateModal">
        添加客户
      </UButton>
    </div>

    <!-- 客户表格 -->
    <UCard>
      <UTable
        :data="customers"
        :columns="columns"
        class="w-full"
        :ui="{
          root: 'relative overflow-auto',
          base: 'min-w-full table-fixed',
          thead: 'bg-gray-50',
          tbody: 'bg-white divide-y',
          tr: 'hover:bg-gray-50',
          th: 'text-left rtl:text-right px-4 py-3.5 text-sm font-semibold text-gray-900',
          td: 'whitespace-nowrap px-4 py-4 text-sm text-gray-500',
        }"
      >
        <!-- v3: 用 -cell，而不是 -data -->
        <template #status-cell="{ getValue }">
          <UBadge
            :color="getValue() === 'active' ? 'success' : 'neutral'"
            variant="subtle"
          >
            {{ getValue() === "active" ? "活跃" : "非活跃" }}
          </UBadge>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              @click="openViewModal(row)"
            >
              查看
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
              @click="openEditModal(row)"
            >
              编辑
            </UButton>
            <UButton
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              @click="deleteCustomer(row)"
            >
              删除
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 添加客户模态框 -->
    <CreateCustomerModal
      v-model:open="showCreateModal"
      @created="handleCustomerCreated"
    />

    <!-- 查看客户模态框 -->
    <ViewCustomerModal
      v-if="viewingCustomer"
      v-model:open="showViewModal"
      :customer="viewingCustomer"
      @edit="handleViewToEdit"
    />

    <!-- 编辑客户模态框 -->
    <EditCustomerModal
      v-if="editingCustomer"
      v-model:open="showEditModal"
      :customer="editingCustomer"
      @updated="handleCustomerUpdated"
      @deleted="handleCustomerDeleted"
    />
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import CreateCustomerModal from "~/components/Modals/CreateCustomerModal.vue";
import EditCustomerModal from "~/components/Modals/EditCustomerModal.vue";
import ViewCustomerModal from "~/components/Modals/ViewCustomerModal.vue";

type Customer = {
  id: string;
  name: string;
  email: string;
  phone: string;
  customerType: string;
  gender?: string;
  birthDate?: string;
  address?: string;
  membershipTier: string;
  source: string;
  tags: string[];
  notes?: string;
  registrationDate: string;
  status: "active" | "inactive";
  totalOrders?: number;
  totalSpent?: number;
};

// 模态框状态
const showCreateModal = ref(false);
const showViewModal = ref(false);
const showEditModal = ref(false);
const viewingCustomer = ref<Customer | null>(null);
const editingCustomer = ref<Customer | null>(null);

// 列定义
const columns = computed<TableColumn<Customer>[]>(() => [
  { accessorKey: "id", header: "客户ID" },
  { accessorKey: "name", header: "客户姓名" },
  { accessorKey: "email", header: "邮箱" },
  { accessorKey: "phone", header: "手机号" },
  { accessorKey: "registrationDate", header: "注册日期" },
  { accessorKey: "status", header: "状态" },
  // 自定义操作列
  { id: "actions", header: "操作" },
]);

// 客户数据（使用响应式数据）
const customers = ref<Customer[]>([
  {
    id: "C001",
    name: "张三",
    email: "zhangsan@example.com",
    phone: "138****1234",
    customerType: "individual",
    membershipTier: "gold",
    source: "website",
    tags: ["VIP", "老客户"],
    registrationDate: "2024-01-10",
    status: "active",
    totalOrders: 15,
    totalSpent: 12500,
  },
  {
    id: "C002",
    name: "李四",
    email: "lisi@example.com",
    phone: "139****5678",
    customerType: "enterprise",
    membershipTier: "platinum",
    source: "referral",
    tags: ["企业客户"],
    registrationDate: "2024-01-08",
    status: "active",
    totalOrders: 8,
    totalSpent: 25600,
  },
  {
    id: "C003",
    name: "王五",
    email: "wangwu@example.com",
    phone: "137****9012",
    customerType: "individual",
    membershipTier: "silver",
    source: "advertisement",
    tags: [],
    registrationDate: "2024-01-05",
    status: "inactive",
    totalOrders: 3,
    totalSpent: 890,
  },
]);

// 打开创建模态框
const openCreateModal = () => {
  showCreateModal.value = true;
};

// 打开查看模态框
const openViewModal = (customer: Customer) => {
  viewingCustomer.value = customer;
  showViewModal.value = true;
};

// 打开编辑模态框
const openEditModal = (customer: Customer) => {
  editingCustomer.value = customer;
  showEditModal.value = true;
};

// 从查看模态框切换到编辑模态框
const handleViewToEdit = (customer: Customer) => {
  showViewModal.value = false;
  viewingCustomer.value = null;
  editingCustomer.value = customer;
  showEditModal.value = true;
};

// 处理客户创建
const handleCustomerCreated = (newCustomer: Customer) => {
  customers.value.unshift(newCustomer);
  showCreateModal.value = false;
};

// 处理客户更新
const handleCustomerUpdated = (updatedCustomer: Customer) => {
  const index = customers.value.findIndex((c) => c.id === updatedCustomer.id);
  if (index !== -1) {
    customers.value[index] = updatedCustomer;
  }
  showEditModal.value = false;
  editingCustomer.value = null;
};

// 处理客户删除
const handleCustomerDeleted = (customerId: string) => {
  const index = customers.value.findIndex((c) => c.id === customerId);
  if (index !== -1) {
    customers.value.splice(index, 1);
  }
  showEditModal.value = false;
  editingCustomer.value = null;
};

// 直接删除客户
const deleteCustomer = async (customer: Customer) => {
  if (!confirm(`确定要删除客户 "${customer.name}" 吗？此操作不可撤销。`)) {
    return;
  }

  // 模拟删除操作
  const index = customers.value.findIndex((c) => c.id === customer.id);
  if (index !== -1) {
    customers.value.splice(index, 1);
  }
};
</script>
