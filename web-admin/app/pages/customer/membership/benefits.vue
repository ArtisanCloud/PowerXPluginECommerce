<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">等级权益</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">
          管理不同会员等级的专属权益
        </p>
      </div>
      <div class="flex items-center gap-3">
        <UButton
          icon="i-heroicons-plus"
          color="primary"
          @click="openCreateModal"
        >
          添加权益
        </UButton>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <UCard
        v-for="benefit in benefits"
        :key="benefit.id"
        class="hover:shadow-lg transition-shadow duration-200"
      >
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="font-semibold text-lg">{{ benefit.name }}</h3>
            <UBadge :color="benefit.status === 'active' ? 'green' : 'gray'">
              {{ benefit.status === 'active' ? '启用' : '禁用' }}
            </UBadge>
          </div>
        </template>

        <div class="space-y-3">
          <p class="text-gray-600 dark:text-gray-300 text-sm">
            {{ benefit.description }}
          </p>
          
          <div class="flex items-center text-sm text-gray-500 dark:text-gray-400">
            <UIcon name="i-heroicons-tag" class="w-4 h-4 mr-1.5" />
            <span>{{ benefit.tiers?.map(t => t.name).join(', ') || '未分配等级' }}</span>
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton
              variant="ghost"
              color="gray"
              size="sm"
              @click="editBenefit(benefit)"
            >
              编辑
            </UButton>
            <UButton
              variant="ghost"
              color="red"
              size="sm"
              @click="deleteBenefit(benefit.id)"
            >
              删除
            </UButton>
          </div>
        </template>
      </UCard>
    </div>

    <!-- 添加/编辑权益模态框 -->
    <UModal
      v-model:open="isModalOpen"
      :title="editingBenefit ? '编辑权益' : '添加权益'"
      description="设置会员等级权益信息"
      :ui="{ body: 'p-0', footer: 'justify-end' }"
    >
      <!-- 表单主体 -->
      <template #body>
        <UCard class="rounded-xl border border-gray-200 dark:border-gray-800">
          <div class="space-y-4 p-4 sm:p-6">
            <UFormField label="权益名称" required>
              <UInput v-model="form.name" placeholder="输入权益名称" />
            </UFormField>

            <UFormField label="权益描述">
              <UTextarea v-model="form.description" placeholder="输入权益描述" />
            </UFormField>

            <UFormField label="适用等级">
              <USelectMenu
                v-model="form.tiers"
                :options="membershipTiers"
                multiple
                value-attribute="id"
                option-attribute="name"
                placeholder="选择适用等级"
              />
            </UFormField>

            <UFormField label="状态">
              <USwitch
                v-model="form.status"
                on-value="active"
                off-value="inactive"
                :labels="{ on: '启用', off: '禁用' }"
              />
            </UFormField>
          </div>
        </UCard>
      </template>

      <!-- 操作按钮 -->
      <template #footer>
        <UButton variant="ghost" @click="isModalOpen = false">取消</UButton>
        <UButton @click="saveBenefit">保存</UButton>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { Ref } from 'vue'

/** 允许本地 v-model:open，也允许 useOverlay 打开后默认即为 true */
const props = defineProps<{ open?: boolean }>();
const emit = defineEmits<{
  "update:open": [boolean];
  created: [benefit: Benefit];
  close: [payload: any]; // useOverlay 通过它拿 result
}>();

// 统一的打开状态
const isOpen = computed({
  get: () => props.open ?? true,
  set: (v) => emit("update:open", v),
});

function close(payload: any) {
  emit("close", payload); // 给 overlay：resolve(instance.result)
  emit("update:open", false); // 给本地：关闭
}

interface MembershipTier {
  id: string
  name: string
  level: number
}

interface Benefit {
  id: string
  name: string
  description: string
  status: 'active' | 'inactive'
  tiers: MembershipTier[]
}

const benefits: Ref<Benefit[]> = ref([
  {
    id: '1',
    name: '免费配送',
    description: '享受全场免费配送服务',
    status: 'active',
    tiers: [
      { id: '2', name: '黄金会员', level: 2 },
      { id: '3', name: '钻石会员', level: 3 }
    ]
  },
  {
    id: '2',
    name: '专属客服',
    description: '享受7x24小时专属客服服务',
    status: 'active',
    tiers: [
      { id: '3', name: '钻石会员', level: 3 }
    ]
  },
  {
    id: '3',
    name: '生日礼包',
    description: '生日当月可领取专属礼包',
    status: 'inactive',
    tiers: [
      { id: '1', name: '普通会员', level: 1 },
      { id: '2', name: '黄金会员', level: 2 },
      { id: '3', name: '钻石会员', level: 3 }
    ]
  }
])

const membershipTiers: Ref<MembershipTier[]> = ref([
  { id: '1', name: '普通会员', level: 1 },
  { id: '2', name: '黄金会员', level: 2 },
  { id: '3', name: '钻石会员', level: 3 }
])

// 表单相关
const isModalOpen = ref(false)
const editingBenefit = ref<Benefit | null>(null)
const form = reactive({
  name: '',
  description: '',
  tiers: [] as string[],
  status: 'active' as 'active' | 'inactive'
})

// 表单验证
const valid = computed(
  () => form.name.trim()
);

// 重置表单
function reset() {
  form.name = ''
  form.description = ''
  form.tiers = []
  form.status = 'active'
}

// 打开创建模态框
const openCreateModal = () => {
  editingBenefit.value = null
  form.name = ''
  form.description = ''
  form.tiers = []
  form.status = 'active'
  isModalOpen.value = true
}

// 编辑权益
const editBenefit = (benefit: Benefit) => {
  editingBenefit.value = benefit
  form.name = benefit.name
  form.description = benefit.description
  form.tiers = benefit.tiers.map(t => t.id)
  form.status = benefit.status
  isModalOpen.value = true
}

// 取消操作
function handleCancel() {
  close(false); // 给 overlay 一个 falsy 结果
}

// 保存权益
const saveBenefit = () => {
  if (editingBenefit.value) {
    // 编辑现有权益
    const index = benefits.value.findIndex(b => b.id === editingBenefit.value!.id)
    if (index !== -1) {
      const tierObjects = form.tiers.map(tierId => {
        const tier = membershipTiers.value.find(t => t.id === tierId)
        return tier || { id: tierId, name: '未知等级', level: 0 }
      })
      
      benefits.value[index] = {
        ...editingBenefit.value,
        name: form.name,
        description: form.description,
        tiers: tierObjects,
        status: form.status
      }
    }
  } else {
    // 添加新权益
    const tierObjects = form.tiers.map(tierId => {
      const tier = membershipTiers.value.find(t => t.id === tierId)
      return tier || { id: tierId, name: '未知等级', level: 0 }
    })
    
    const newBenefit: Benefit = {
      id: String(benefits.value.length + 1),
      name: form.name,
      description: form.description,
      tiers: tierObjects,
      status: form.status
    };
    
    benefits.value.push(newBenefit);
    emit("created", newBenefit); // 本地使用
    close({ action: "create", benefit: newBenefit }); // overlay 拿 result 的关键
  }
  
  isModalOpen.value = false
}

// 删除权益
const deleteBenefit = (id: string) => {
  if (confirm('确定要删除这个权益吗？')) {
    const index = benefits.value.findIndex(b => b.id === id)
    if (index !== -1) {
      benefits.value.splice(index, 1)
    }
  }
}
</script>