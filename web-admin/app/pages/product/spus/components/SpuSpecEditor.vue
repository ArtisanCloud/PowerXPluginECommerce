<template>
	<div class="space-y-4">
		<UAlert
			color="neutral"
			variant="soft"
			title="说明"
			description="这里维护的是“变体规格”（颜色/尺码等），用于生成 SKU 组合。属性和规格菜单中的“属性/规格”目前未对接此数据。"
		/>

		<div class="flex flex-wrap items-center justify-between gap-3">
			<div>
				<h3 class="text-base font-semibold text-gray-900 dark:text-white">规格定义</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400">每个规格组包含若干规格值，SKU 由规格值组合生成。</p>
			</div>
			<div class="flex items-center gap-2">
				<UButton
					color="primary"
					variant="outline"
					size="sm"
					:disabled="loading || !editable"
					@click="addGroup"
				>
					新增规格组
				</UButton>
				<UButton color="primary" size="sm" :loading="saving" :disabled="loading || !editable" @click="save">
					保存
				</UButton>
			</div>
		</div>

		<UAlert
			v-if="!editable"
			color="warning"
			variant="soft"
			title="只读"
			description="当前 SPU 不是草稿状态，规格定义为只读；如需修改请先创建草稿版本。"
		/>

		<div v-if="loading">
			<USkeleton class="h-24" />
		</div>

		<div v-else class="space-y-4">
			<UCard v-for="(g, gi) in groups" :key="g.__key" class="overflow-hidden">
				<template #header>
					<div class="flex flex-wrap items-center justify-between gap-3">
						<div class="flex items-center gap-2">
							<UBadge variant="soft" color="neutral">规格组 {{ gi + 1 }}</UBadge>
							<UBadge v-if="g.required" variant="soft" color="primary">必选</UBadge>
						</div>
						<UButton
							color="error"
							variant="soft"
							size="xs"
							:disabled="!editable"
							@click="removeGroup(gi)"
						>
							删除规格组
						</UButton>
					</div>
				</template>

				<div class="grid gap-4 md:grid-cols-2">
					<UFormField label="编码（Code）" help="建议全小写下划线，例如 color/size">
						<UInput v-model="g.code" :disabled="!editable" placeholder="例如：color" />
					</UFormField>
					<UFormField label="名称（Name）" help="展示给用户的名称，例如 颜色/尺码">
						<UInput v-model="g.name" :disabled="!editable" placeholder="例如：颜色" />
					</UFormField>
					<UFormField label="排序" help="数字越小越靠前">
						<UInput v-model.number="g.sort_order" :disabled="!editable" type="number" />
					</UFormField>
					<UFormField label="是否必选">
						<UCheckbox v-model="g.required" :disabled="!editable" label="必选" />
					</UFormField>
				</div>

				<div class="mt-4 space-y-3">
					<div class="flex flex-wrap items-center justify-between gap-2">
						<h4 class="text-sm font-semibold text-gray-900 dark:text-white">规格值（Options）</h4>
						<UButton
							color="primary"
							variant="outline"
							size="xs"
							:disabled="!editable"
							@click="addOption(gi)"
						>
							新增规格值
						</UButton>
					</div>

					<div v-if="!g.options.length" class="text-sm text-gray-500 dark:text-gray-400">
						暂无规格值，请先添加至少 1 个。
					</div>

					<div v-else class="space-y-3">
						<div
							v-for="(o, oi) in g.options"
							:key="o.__key"
							class="rounded-md border border-gray-200 dark:border-gray-700 p-3"
						>
							<div class="flex items-center justify-between gap-3 mb-2">
								<UBadge variant="soft" color="neutral">值 {{ oi + 1 }}</UBadge>
								<UButton
									color="error"
									variant="soft"
									size="xs"
									:disabled="!editable"
									@click="removeOption(gi, oi)"
								>
									删除
								</UButton>
							</div>
							<div class="grid gap-3 md:grid-cols-3">
								<UFormField label="编码（Code）">
									<UInput v-model="o.code" :disabled="!editable" placeholder="例如：red" />
								</UFormField>
								<UFormField label="名称（Name）">
									<UInput v-model="o.name" :disabled="!editable" placeholder="例如：红色" />
								</UFormField>
								<UFormField label="排序">
									<UInput v-model.number="o.sort_order" :disabled="!editable" type="number" />
								</UFormField>
							</div>
						</div>
					</div>
				</div>
			</UCard>

			<UAlert
				v-if="!groups.length"
				color="neutral"
				variant="soft"
				title="暂无规格定义"
				description="你可以点击“新增规格组”开始配置，例如：颜色/尺码。配置后即可到“关联 SKU / 批量生成”生成 SKU 组合。"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useToast } from "#imports";
import { useProductSpecApi, type ProductSpecGroup, type ReplaceProductSpecRequest } from "~/composables/api";

type EditableOption = {
  __key: string;
  id?: string;
  code: string;
  name: string;
  sort_order: number;
  status?: string;
  meta?: Record<string, any> | null;
};

type EditableGroup = {
  __key: string;
  id?: string;
  code: string;
  name: string;
  sort_order: number;
  required: boolean;
  status?: string;
  options: EditableOption[];
};

const props = defineProps<{ spuId: string; editable?: boolean }>();
const emit = defineEmits<{ (e: "updated", groups: ProductSpecGroup[]): void }>();

const api = useProductSpecApi();
const toast = useToast();

const loading = ref(false);
const saving = ref(false);
const groups = ref<EditableGroup[]>([]);

const newKey = () => `${Date.now()}-${Math.random().toString(16).slice(2)}`;

const toEditable = (input: ProductSpecGroup[]): EditableGroup[] =>
  (input || []).map((g) => ({
    __key: newKey(),
    id: g.id,
    code: g.code,
    name: g.name,
    sort_order: Number(g.sort_order || 0),
    required: Boolean(g.required),
    status: g.status || "active",
    options: (g.options || []).map((o) => ({
      __key: newKey(),
      id: o.id,
      code: o.code,
      name: o.name,
      sort_order: Number(o.sort_order || 0),
      status: o.status || "active",
      meta: o.meta || null,
    })),
  }));

const fetchGroups = async () => {
  if (!props.spuId) return;
  loading.value = true;
  try {
    const data = await api.list(props.spuId);
    groups.value = toEditable(data.groups || []);
  } catch (e: any) {
    toast.add({ title: "加载规格失败", description: e?.message || String(e), color: "error" });
  } finally {
    loading.value = false;
  }
};

const addGroup = () => {
  groups.value.push({
    __key: newKey(),
    code: "",
    name: "",
    sort_order: groups.value.length,
    required: true,
    status: "active",
    options: [],
  });
};

const removeGroup = (index: number) => {
  groups.value.splice(index, 1);
};

const addOption = (groupIndex: number) => {
  const g = groups.value[groupIndex];
  if (!g) return;
  g.options.push({
    __key: newKey(),
    code: "",
    name: "",
    sort_order: g.options.length,
    status: "active",
    meta: null,
  });
};

const removeOption = (groupIndex: number, optionIndex: number) => {
  const g = groups.value[groupIndex];
  if (!g) return;
  g.options.splice(optionIndex, 1);
};

const validate = (): string | null => {
  for (const g of groups.value) {
    if (!g.code.trim() || !g.name.trim()) return "规格组的编码/名称不能为空";
    if (!g.options.length) return `规格组 ${g.code || g.name} 至少需要 1 个规格值`;
    for (const o of g.options) {
      if (!o.code.trim() || !o.name.trim()) return `规格组 ${g.code || g.name} 的规格值编码/名称不能为空`;
    }
  }
  return null;
};

const toPayload = (): ReplaceProductSpecRequest => ({
  groups: groups.value.map((g) => ({
    id: g.id,
    code: g.code.trim(),
    name: g.name.trim(),
    sort_order: Number(g.sort_order || 0),
    required: Boolean(g.required),
    status: g.status || "active",
    options: g.options.map((o) => ({
      id: o.id,
      code: o.code.trim(),
      name: o.name.trim(),
      sort_order: Number(o.sort_order || 0),
      status: o.status || "active",
      meta: o.meta || null,
    })),
  })),
});

const save = async () => {
  if (!props.editable) return;
  const err = validate();
  if (err) {
    toast.add({ title: "校验失败", description: err, color: "warning" });
    return;
  }
  saving.value = true;
  try {
    const data = await api.replace(props.spuId, toPayload());
    groups.value = toEditable(data.groups || []);
    emit("updated", data.groups || []);
    toast.add({ title: "已保存规格定义", color: "success" });
  } catch (e: any) {
    toast.add({ title: "保存失败", description: e?.message || String(e), color: "error" });
  } finally {
    saving.value = false;
  }
};

onMounted(fetchGroups);
</script>

