<template>
  <div class="max-w-6xl mx-auto py-6 space-y-6">
    <!-- 顶部栏 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-gray-100">新建 SPU</h1>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">请完善以下信息，带 * 为必填项。</p>
      </div>
      <div class="flex gap-3">
        <UButton variant="soft" color="gray" @click="goToList">
          返回列表
        </UButton>
        <UButton icon="i-heroicons-check" color="primary" @click="saveProduct">
          保存并返回
        </UButton>
      </div>
    </div>

    <!-- 表单主体 -->
    <template v-if="productForm">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- 左侧：主要信息 -->
        <div class="lg:col-span-2 space-y-6">
          <!-- 基础信息 -->
          <section class="rounded-xl border border-gray-200/60 dark:border-gray-800/60 bg-white/70 dark:bg-gray-900/60 shadow-sm">
            <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <h2 class="text-base font-medium">基础信息</h2>
            </div>
            <div class="p-5 grid grid-cols-1 md:grid-cols-2 gap-5">
              <UFormField :label="$t('form.name')" required>
                <UInput v-model="productForm.name" :placeholder="$t('form.name')" />
              </UFormField>

              <UFormField :label="$t('product.productNumber')" required>
                <UInput v-model="productForm.productNumber" :placeholder="$t('product.productNumber')" />
              </UFormField>

              <UFormField :label="$t('product.brand')">
                <UInput v-model="productForm.brand" :placeholder="$t('product.brand')" />
              </UFormField>

              <UFormField :label="$t('product.category')">
                <UInput v-model="productForm.category" :placeholder="$t('product.category')" />
              </UFormField>

              <UFormField :label="$t('product.productType')">
                <UInput v-model="productForm.productType" :placeholder="$t('product.productType')" />
              </UFormField>

              <UFormField :label="$t('product.barcode')">
                <UInput v-model="productForm.barcode" :placeholder="$t('product.barcode')" />
              </UFormField>

              <UFormField :label="$t('product.taxCategory')">
                <UInput v-model="productForm.taxCategory" :placeholder="$t('product.taxCategory')" />
              </UFormField>

              <UFormField :label="$t('product.regulatoryRequirements')">
                <UInput v-model="productForm.regulatoryRequirements" :placeholder="$t('product.regulatoryRequirements')" />
              </UFormField>

              <UFormField class="md:col-span-2" :label="$t('form.description')">
                <UTextarea v-model="productForm.description" :rows="5" :placeholder="$t('form.description')" />
              </UFormField>
            </div>
          </section>

          <!-- 媒体 -->
          <section class="rounded-xl border border-gray-200/60 dark:border-gray-800/60 bg-white/70 dark:bg-gray-900/60 shadow-sm">
            <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <h2 class="text-base font-medium">媒体</h2>
            </div>
            <div class="p-5 space-y-4">
              <div class="flex items-center gap-3">
                <UButton icon="i-heroicons-arrow-up-tray" variant="soft" @click="uploadImage">上传图片</UButton>
                <UInput v-model="productForm.video" :placeholder="$t('product.video')" class="flex-1" />
              </div>
              <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                <div v-for="(img, index) in productForm.images" :key="index" class="group relative rounded-lg overflow-hidden border border-gray-200 dark:border-gray-800">
                  <img :src="img" class="w-full h-28 object-cover" />
                  <button class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition" @click="removeImage(index)">
                    <UButton size="xs" color="red" variant="solid" icon="i-heroicons-x-mark" />
                  </button>
                </div>
              </div>
            </div>
          </section>

          <!-- SEO -->
          <section class="rounded-xl border border-gray-200/60 dark:border-gray-800/60 bg-white/70 dark:bg-gray-900/60 shadow-sm">
            <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <h2 class="text-base font-medium">SEO</h2>
            </div>
            <div class="p-5 grid grid-cols-1 md:grid-cols-2 gap-5">
              <UFormField :label="$t('product.seoTitle')">
                <UInput v-model="productForm.seoTitle" :placeholder="$t('product.seoTitle')" />
              </UFormField>
              <UFormField :label="$t('product.seoDescription')" class="md:col-span-2">
                <UTextarea v-model="productForm.seoDescription" :rows="3" :placeholder="$t('product.seoDescription')" />
              </UFormField>
            </div>
          </section>

          <!-- 自定义属性 -->
          <section class="rounded-xl border border-gray-200/60 dark:border-gray-800/60 bg-white/70 dark:bg-gray-900/60 shadow-sm">
            <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
              <h2 class="text-base font-medium">自定义属性</h2>
              <UButton size="sm" variant="soft" icon="i-heroicons-plus" @click="addCustomField">添加属性</UButton>
            </div>
            <div class="p-5 space-y-3">
              <div v-for="(field, index) in productForm.customFields" :key="index" class="grid grid-cols-1 md:grid-cols-6 gap-3 items-center">
                <UInput v-model="field.name" :placeholder="$t('form.name')" class="md:col-span-2" />
                <UInput v-model="field.value" :placeholder="$t('common.value')" class="md:col-span-3" />
                <div class="md:col-span-1 flex justify-end">
                  <UButton color="red" variant="soft" icon="i-heroicons-trash" @click="removeCustomField(index)">删除</UButton>
                </div>
              </div>
            </div>
          </section>

          <!-- 多语言 -->
          <section class="rounded-xl border border-gray-200/60 dark:border-gray-800/60 bg-white/70 dark:bg-gray-900/60 shadow-sm">
            <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
              <h2 class="text-base font-medium">多语言</h2>
              <UButton size="sm" variant="soft" icon="i-heroicons-plus" @click="addLanguage">添加语言</UButton>
            </div>
            <div class="p-5 space-y-3">
              <div v-for="(lang, index) in productForm.languages" :key="index" class="grid grid-cols-1 md:grid-cols-5 gap-3 items-center">
                <UInput v-model="lang.code" :placeholder="$t('language.language')" class="md:col-span-1" />
                <UInput v-model="lang.name" :placeholder="$t('form.name')" class="md:col-span-3" />
                <div class="md:col-span-1 flex justify-end">
                  <UButton color="red" variant="soft" icon="i-heroicons-trash" @click="removeLanguage(index)">删除</UButton>
                </div>
              </div>
            </div>
          </section>
        </div>

        <!-- 右侧：辅助信息 -->
        <div class="space-y-6">
          <!-- 上架渠道 -->
          <section class="rounded-xl border border-gray-200/60 dark:border-gray-800/60 bg-white/70 dark:bg-gray-900/60 shadow-sm">
            <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <h2 class="text-base font-medium">上架渠道</h2>
            </div>
            <div class="p-5 space-y-3">
              <UCheckbox v-model="productForm.channels.online" :label="$t('channel.online')" />
              <UCheckbox v-model="productForm.channels.mobile" :label="$t('channel.mobile')" />
              <UCheckbox v-model="productForm.channels.inStore" :label="$t('channel.inStore')" />
            </div>
          </section>

          <!-- 库存与物流 -->
          <section class="rounded-xl border border-gray-200/60 dark:border-gray-800/60 bg-white/70 dark:bg-gray-900/60 shadow-sm">
            <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <h2 class="text-base font-medium">库存与物流</h2>
            </div>
            <div class="p-5 space-y-3">
              <UCheckbox v-model="productForm.needInventoryManagement" :label="$t('product.needInventoryManagement')" />
              <UCheckbox v-model="productForm.needLogistics" :label="$t('product.needLogistics')" />
            </div>
          </section>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="max-w-6xl mx-auto py-10 text-center text-gray-500">加载中...</div>
    </template>

    <!-- 底部吸附操作条 -->
    <div class="sticky bottom-0 left-0 right-0 border-t border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-gray-900/80 backdrop-blur supports-[backdrop-filter]:bg-white/50 supports-[backdrop-filter]:dark:bg-gray-900/50">
      <div class="max-w-6xl mx-auto px-4 py-3 flex items-center justify-end gap-3">
        <UButton variant="soft" color="gray" @click="goToList">取消</UButton>
        <UButton icon="i-heroicons-check" color="primary" @click="saveProduct">保存</UButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
section { transition: border-color .2s ease, box-shadow .2s ease; }
section:hover { box-shadow: 0 4px 18px rgba(0,0,0,.06); border-color: rgba(0,0,0,.08); }
</style>

<script setup lang="ts">
const { t } = useI18n();
const router = useRouter();

// 表单数据
const productForm = ref({
  // 识别属性
  productNumber: "",
  category: "",
  brand: "",
  barcode: "",

  // 展示属性
  name: "",
  images: [] as string[],
  video: "",
  description: "",
  seoTitle: "",
  seoDescription: "",

  // 约束属性
  productType: "physical",
  needInventoryManagement: true,
  needLogistics: true,

  // 合规属性
  taxCategory: "",
  regulatoryRequirements: "",

  // 扩展属性
  languages: [] as { code: string; name: string }[],
  channels: {
    online: true,
    inStore: false,
    mobile: false
  },
  customFields: [] as { name: string; value: string }[]
});

// 选项数据
const categoryOptions = [
  { label: "手机", value: "手机" },
  { label: "电脑", value: "电脑" },
  { label: "配件", value: "配件" },
  { label: "家电", value: "家电" },
  { label: "服装", value: "服装" }
];

const brandOptions = [
  { label: "Apple", value: "Apple" },
  { label: "Samsung", value: "Samsung" },
  { label: "Huawei", value: "Huawei" },
  { label: "Xiaomi", value: "Xiaomi" },
  { label: "OPPO", value: "OPPO" }
];

const productTypeOptions = [
  { label: t("product.physicalProduct"), value: "physical" },
  { label: t("product.virtualProduct"), value: "virtual" }
];

const taxCategoryOptions = [
  { label: t("common.none"), value: "" },
  { label: "电子产品", value: "electronics" },
  { label: "服装", value: "clothing" },
  { label: "食品", value: "food" },
  { label: "图书", value: "books" }
];

const languageOptions = [
  { label: "中文", value: "zh" },
  { label: "English", value: "en" },
  { label: "日本語", value: "ja" },
  { label: "한국어", value: "ko" }
];

// 方法
const addLanguage = () => {
  productForm.value.languages.push({ code: "", name: "" });
};

const removeLanguage = (index: number) => {
  productForm.value.languages.splice(index, 1);
};

const addCustomField = () => {
  productForm.value.customFields.push({ name: "", value: "" });
};

const removeCustomField = (index: number) => {
  productForm.value.customFields.splice(index, 1);
};

const uploadImage = () => {
  // 模拟图片上传
  const placeholderImages = [
    "https://via.placeholder.com/300x300",
    "https://via.placeholder.com/300x300/ff0000/ffffff",
    "https://via.placeholder.com/300x300/00ff00/ffffff",
    "https://via.placeholder.com/300x300/0000ff/ffffff"
  ];

  const randomImage = placeholderImages[Math.floor(Math.random() * placeholderImages.length)];
  productForm.value.images.push(randomImage);
};

const removeImage = (index: number) => {
  productForm.value.images.splice(index, 1);
};

const saveProduct = () => {
  // 这里应该调用API保存商品
  alert(t("message.success.created"));
  goToList();
};

const goToList = () => {
  router.push("/product/spu");
};
</script>
