<template>
	<div class="space-y-4">
		<UAlert
			color="neutral"
			variant="soft"
			:title="t('product.spu.specEditor.noticeTitle')"
			:description="t('product.spu.specEditor.noticeDescription')"
		/>

		<div class="flex flex-wrap items-center justify-between gap-3">
			<div>
				<h3 class="text-base font-semibold text-gray-900 dark:text-white">
					{{ t('product.spu.specEditor.title') }}
				</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400">
					{{ t('product.spu.specEditor.description') }}
				</p>
			</div>
			<div class="flex items-center gap-2">
				<UButton
					color="neutral"
					variant="soft"
					size="sm"
					:loading="syncingCategorySpecs"
					:disabled="loading || !editable || !categoryId"
					@click="syncFromCategory"
				>
					{{ t('product.spu.specEditor.syncFromCategory') }}
				</UButton>
				<UButton
					color="primary"
					variant="outline"
					size="sm"
					:disabled="loading || !editable"
					@click="addGroup"
				>
					{{ t('product.spu.specEditor.addGroup') }}
				</UButton>
				<UButton color="primary" size="sm" :loading="saving" :disabled="loading || !editable" @click="save">
					{{ t('product.spu.specEditor.save') }}
				</UButton>
			</div>
		</div>

		<UAlert
			v-if="!editable"
			color="warning"
			variant="soft"
			:title="t('product.spu.specEditor.readonlyTitle')"
			:description="t('product.spu.specEditor.readonlyDescription')"
		/>

		<div v-if="loading">
			<USkeleton class="h-24" />
		</div>

		<div v-else class="space-y-4">
			<UCard v-for="(g, gi) in groups" :key="g.__key" class="overflow-hidden">
				<template #header>
					<div class="flex flex-wrap items-center justify-between gap-3">
						<div class="flex items-center gap-2">
							<UBadge variant="soft" color="neutral">
								{{ t('product.spu.specEditor.groupBadge', { index: gi + 1 }) }}
							</UBadge>
							<UBadge v-if="g.required" variant="soft" color="primary">
								{{ t('product.spu.specEditor.required') }}
							</UBadge>
						</div>
						<UButton
							color="error"
							variant="soft"
							size="xs"
							:disabled="!editable"
							@click="removeGroup(gi)"
						>
							{{ t('product.spu.specEditor.removeGroup') }}
						</UButton>
					</div>
				</template>

				<div class="grid gap-4 md:grid-cols-2">
					<UFormField
						:label="t('product.spu.specEditor.code')"
						:help="t('product.spu.specEditor.codeHelp')"
					>
						<UInput v-model="g.code" :disabled="!editable" :placeholder="t('product.spu.specEditor.codePlaceholder')" />
					</UFormField>
					<UFormField
						:label="t('product.spu.specEditor.name')"
						:help="t('product.spu.specEditor.nameHelp')"
					>
						<UInput v-model="g.name" :disabled="!editable" :placeholder="t('product.spu.specEditor.namePlaceholder')" />
					</UFormField>
					<UFormField :label="t('product.spu.specEditor.sortOrder')" :help="t('product.spu.specEditor.sortOrderHelp')">
						<UInput v-model.number="g.sort_order" :disabled="!editable" type="number" />
					</UFormField>
					<UFormField :label="t('product.spu.specEditor.requiredField')">
						<UCheckbox v-model="g.required" :disabled="!editable" :label="t('product.spu.specEditor.required')" />
					</UFormField>
				</div>

				<div class="mt-4 space-y-3">
					<div class="flex flex-wrap items-center justify-between gap-2">
						<h4 class="text-sm font-semibold text-gray-900 dark:text-white">
							{{ t('product.spu.specEditor.options') }}
						</h4>
						<UButton
							color="primary"
							variant="outline"
							size="xs"
							:disabled="!editable"
							@click="addOption(gi)"
						>
							{{ t('product.spu.specEditor.addOption') }}
						</UButton>
					</div>

					<div v-if="!g.options.length" class="text-sm text-gray-500 dark:text-gray-400">
						{{ t('product.spu.specEditor.emptyOptions') }}
					</div>

					<div v-else class="space-y-3">
						<div
							v-for="(o, oi) in g.options"
							:key="o.__key"
							class="rounded-md border border-gray-200 dark:border-gray-700 p-3"
						>
							<div class="flex items-center justify-between gap-3 mb-2">
								<UBadge variant="soft" color="neutral">
									{{ t('product.spu.specEditor.optionBadge', { index: oi + 1 }) }}
								</UBadge>
								<UButton
									color="error"
									variant="soft"
									size="xs"
									:disabled="!editable"
									@click="removeOption(gi, oi)"
								>
									{{ t('product.spu.specEditor.removeOption') }}
								</UButton>
							</div>
							<div class="grid gap-3 md:grid-cols-3">
								<UFormField :label="t('product.spu.specEditor.optionCode')">
									<UInput v-model="o.code" :disabled="!editable" :placeholder="t('product.spu.specEditor.optionCodePlaceholder')" />
								</UFormField>
								<UFormField :label="t('product.spu.specEditor.optionName')">
									<UInput v-model="o.name" :disabled="!editable" :placeholder="t('product.spu.specEditor.optionNamePlaceholder')" />
								</UFormField>
								<UFormField :label="t('product.spu.specEditor.sortOrder')">
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
				:title="t('product.spu.specEditor.emptyTitle')"
				:description="t('product.spu.specEditor.emptyDescription')"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useToast, useI18n } from "#imports";
import { useCategoryApi } from "~/composables/api/useCategory";
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

const props = defineProps<{ spuId: string; categoryId?: string; editable?: boolean }>();
const emit = defineEmits<{ (e: "updated", groups: ProductSpecGroup[]): void }>();

const api = useProductSpecApi();
const categoryApi = useCategoryApi();
const toast = useToast();
const { t } = useI18n();

const loading = ref(false);
const saving = ref(false);
const syncingCategorySpecs = ref(false);
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

const fromCategorySaleSpecs = (input: any[]): EditableGroup[] =>
  (input || []).filter((g) => (g.status || "active") === "active").map((g, index) => ({
    __key: newKey(),
    code: String(g.code ?? "").trim(),
    name: String(g.name ?? "").trim(),
    sort_order: Number(g.sortOrder ?? g.sort_order ?? index),
    required: Boolean(g.required ?? true),
    status: g.status || "active",
    options: (g.options || []).filter((o: any) => (o.status || "active") === "active").map((o: any, optionIndex: number) => ({
      __key: newKey(),
      code: String(o.code ?? "").trim(),
      name: String(o.name ?? "").trim(),
      sort_order: Number(o.sortOrder ?? o.sort_order ?? optionIndex),
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
    toast.add({ title: t("product.spu.specEditor.loadFailed"), description: e?.message || String(e), color: "error" });
  } finally {
    loading.value = false;
  }
};

const syncFromCategory = async () => {
  if (!props.categoryId) {
    toast.add({ title: t("product.spu.specEditor.categoryMissing"), color: "error" });
    return;
  }
  syncingCategorySpecs.value = true;
  try {
    const data = await categoryApi.saleSpecs(props.categoryId);
    const nextGroups = fromCategorySaleSpecs(data.groups || []);
    if (!nextGroups.length) {
      toast.add({ title: t("product.spu.specEditor.categorySpecEmpty"), color: "warning" });
      return;
    }
    groups.value = nextGroups;
    toast.add({ title: t("product.spu.specEditor.categorySpecSynced"), color: "primary" });
  } catch (e: any) {
    toast.add({ title: t("product.spu.specEditor.categorySpecSyncFailed"), description: e?.message || String(e), color: "error" });
  } finally {
    syncingCategorySpecs.value = false;
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
  const groupCodes = new Set<string>();
  for (const g of groups.value) {
    const groupCode = g.code.trim();
    if (!groupCode || !g.name.trim()) return t("product.spu.specEditor.groupRequired");
    const groupCodeKey = groupCode.toLowerCase();
    if (groupCodes.has(groupCodeKey)) return t("product.spu.specEditor.duplicateGroup", { code: groupCode });
    groupCodes.add(groupCodeKey);
    if (!g.options.length) return t("product.spu.specEditor.groupOptionRequired", { code: groupCode });
    const optionCodes = new Set<string>();
    for (const o of g.options) {
      const optionCode = o.code.trim();
      if (!optionCode || !o.name.trim()) return t("product.spu.specEditor.optionRequired", { code: groupCode });
      const optionCodeKey = optionCode.toLowerCase();
      if (optionCodes.has(optionCodeKey)) return t("product.spu.specEditor.duplicateOption", { code: groupCode, option: optionCode });
      optionCodes.add(optionCodeKey);
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
    toast.add({ title: t("product.spu.specEditor.validateFailed"), description: err, color: "warning" });
    return;
  }
  saving.value = true;
  try {
    const data = await api.replace(props.spuId, toPayload());
    groups.value = toEditable(data.groups || []);
    emit("updated", data.groups || []);
    toast.add({ title: t("product.spu.specEditor.saveSuccess"), color: "success" });
  } catch (e: any) {
    toast.add({ title: t("product.spu.specEditor.saveFailed"), description: e?.message || String(e), color: "error" });
  } finally {
    saving.value = false;
  }
};

onMounted(fetchGroups);
</script>
