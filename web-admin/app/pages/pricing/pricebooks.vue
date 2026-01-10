<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          {{ $t("pricing.pricebooks") }}
        </h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          管理价格手册（价目表），并为不同渠道/客户组配置基础价格。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-heroicons-arrow-path"
          :loading="loading"
          @click="refresh"
        >
          {{ $t("common.refresh") }}
        </UButton>
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          @click="openCreate"
        >
          新增{{ $t("pricing.pricebooks") }}
        </UButton>
      </div>
    </div>

    <UCard>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UFormField :label="$t('common.search')" class="w-full">
          <UInput
            v-model="filters.keyword"
            :placeholder="$t('common.search')"
            icon="i-heroicons-magnifying-glass"
            @keydown.enter="refresh"
          />
        </UFormField>
        <UFormField label="类型" class="w-full">
          <USelect v-model="filters.type" :items="typeItems" class="w-full" />
        </UFormField>
        <UFormField label="币种" class="w-full">
          <UInput
            v-model="filters.currency"
            placeholder="币种（如 CNY）"
            @keydown.enter="refresh"
          />
        </UFormField>
        <UFormField :label="$t('common.status')" class="w-full">
          <USelect v-model="filters.status" :items="statusItems" class="w-full" />
        </UFormField>
      </div>
    </UCard>

    <UCard>
      <UTable :data="rows" :columns="columns" :loading="loading">
        <template #code-cell="{ row }">
          <div class="flex items-center gap-2">
            <span class="font-medium">{{ row.original.code }}</span>
            <UBadge
              v-if="isBase(row.original)"
              color="neutral"
              variant="subtle"
            >
              系统
            </UBadge>
          </div>
        </template>

        <template #status-cell="{ row }">
          <UBadge :color="statusColor(row.original.status)" variant="subtle">
            {{ row.original.status }}
          </UBadge>
        </template>

        <template #currentVersion-cell="{ row }">
          <div class="flex items-center gap-2">
            <span class="text-sm text-gray-600 dark:text-gray-300">
              {{ row.original.current_version ? `v${row.original.current_version}` : "-" }}
              <span v-if="row.original.current_version_state" class="text-gray-400">
                · {{ row.original.current_version_state }}
              </span>
            </span>
            <UButton
              v-if="row.original.current_version_id"
              size="xs"
              color="neutral"
              variant="ghost"
              icon="i-heroicons-clipboard-document"
              :title="$t('common.copy')"
              @click="copyText(row.original.current_version_id)"
            />
          </div>
        </template>

        <template #updatedAt-cell="{ row }">
          <span class="text-sm text-gray-600 dark:text-gray-300">
            {{ formatTime(row.original.updated_at) }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              size="sm"
              color="neutral"
              variant="ghost"
              icon="i-heroicons-eye"
              @click="view(row.original)"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              size="sm"
              color="primary"
              variant="ghost"
              icon="i-heroicons-pencil-square"
              :disabled="isBase(row.original)"
              :title="isBase(row.original) ? '基础价目表为系统托管，不支持在此编辑' : ''"
              @click="openEdit(row.original)"
            >
              {{ $t("common.edit") }}
            </UButton>
            <UButton
              size="sm"
              color="error"
              variant="ghost"
              icon="i-heroicons-trash"
              :disabled="isBase(row.original)"
              :title="isBase(row.original) ? '基础价目表不能删除' : ''"
              @click="confirmRemove(row.original)"
            >
              {{ $t("common.delete") }}
            </UButton>
          </div>
        </template>
      </UTable>

      <div class="flex items-center justify-between mt-4">
        <p class="text-sm text-gray-500">
          {{ $t("common.total", { count: meta.total }) }}
        </p>
        <div class="flex items-center gap-2">
          <UButton
            size="sm"
            variant="outline"
            color="neutral"
            :disabled="filters.page <= 1 || loading"
            @click="prevPage"
          >
            {{ $t("common.previous") }}
          </UButton>
          <span class="text-sm text-gray-600 dark:text-gray-300">
            {{ filters.page }} / {{ totalPages }}
          </span>
          <UButton
            size="sm"
            variant="outline"
            color="neutral"
            :disabled="filters.page >= totalPages || loading"
            @click="nextPage"
          >
            {{ $t("common.next") }}
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- 详情 -->
    <UModal
      v-model:open="detailOpen"
      :title="active?.name || $t('pricing.pricebooks')"
      :description="active?.code ? `code: ${active.code}` : '查看价格手册详情'"
      :close="true"
      :prevent-close="true"
      :ui="{ content: 'max-w-3xl w-full' }"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <div v-if="active" class="space-y-4">
            <UTabs v-model="detailTab" :items="detailTabs" class="w-full" />

            <div v-if="detailTab === 'detail'" class="space-y-4 text-sm">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div class="flex justify-between gap-2">
                  <span class="text-gray-500">Code</span>
                  <span class="font-medium">{{ active.code }}</span>
                </div>
                <div class="flex justify-between gap-2">
                  <span class="text-gray-500">类型</span>
                  <span class="font-medium">{{ active.type }}</span>
                </div>
                <div class="flex justify-between gap-2">
                  <span class="text-gray-500">币种</span>
                  <span class="font-medium">{{ active.currency }}</span>
                </div>
                <div class="flex justify-between gap-2">
                  <span class="text-gray-500">状态</span>
                  <span class="font-medium">{{ active.status }}</span>
                </div>
                <div class="flex justify-between gap-2 md:col-span-2">
                  <span class="text-gray-500">当前版本</span>
                  <span class="font-medium">
                    {{ active.current_version ? `v${active.current_version}` : "-" }}
                    <span v-if="active.current_version_state" class="text-gray-400">
                      · {{ active.current_version_state }}
                    </span>
                  </span>
                </div>
              </div>

              <div>
                <p class="text-gray-500 mb-1">描述</p>
                <p class="text-gray-700 dark:text-gray-200 whitespace-pre-wrap">
                  {{ active.description || "-" }}
                </p>
              </div>
            </div>

            <div v-else-if="detailTab === 'versions'" class="space-y-4">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p class="text-sm text-gray-500">
                    版本用于承载草稿与发布（draft/publish）。只有版本为 active 才会对外生效。
                  </p>
                </div>
                <div class="flex gap-2">
                  <UButton
                    color="neutral"
                    variant="outline"
                    icon="i-heroicons-arrow-path"
                    :loading="versionsLoading"
                    @click="refreshVersions"
                  >
                    {{ $t("common.refresh") }}
                  </UButton>
                  <UButton
                    color="primary"
                    icon="i-heroicons-plus"
                    :disabled="!active?.id || isBase(active)"
                    :title="active && isBase(active) ? '基础价目表为系统托管，不支持在此创建版本' : ''"
                    @click="openCreateVersion"
                  >
                    新建草稿版本
                  </UButton>
                </div>
              </div>

              <UFormField label="选择版本">
                <USelect
                  v-model="selectedVersionId"
                  :items="versionSelectItems"
                  class="w-full"
                  :disabled="versionsLoading || versionSelectItems.length === 0"
                />
              </UFormField>

              <UCard>
                <div class="space-y-3">
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <div class="text-sm">
                      <span class="text-gray-500">状态</span>
                      <span class="ml-2 font-medium">{{ selectedVersion?.state || "-" }}</span>
                      <span class="mx-2 text-gray-300">·</span>
                      <span class="text-gray-500">生效</span>
                      <span class="ml-2 font-medium">{{ selectedVersion ? formatTime(selectedVersion.effective_at) : "-" }}</span>
                      <span class="mx-2 text-gray-300">·</span>
                      <span class="text-gray-500">失效</span>
                      <span class="ml-2 font-medium">{{ selectedVersion?.expires_at ? formatTime(selectedVersion.expires_at) : "-" }}</span>
                    </div>
                    <div class="flex flex-wrap gap-2">
                      <UButton
                        color="neutral"
                        variant="outline"
                        icon="i-heroicons-pencil-square"
                        :disabled="!selectedVersion"
                        :title="!selectedVersion ? '请先选择版本' : (!canEditItems ? '当前版本为只读（仅 draft 可编辑）' : '')"
                        @click="openItems"
                      >
                        {{ canEditItems ? "编辑条目" : "查看条目" }}
                      </UButton>
                      <UButton
                        color="primary"
                        icon="i-heroicons-rocket-launch"
                        :disabled="!canPublish"
                        :title="!canPublish ? '仅 draft 版本可发布' : ''"
                        @click="openPublish"
                      >
                        发布
                      </UButton>
                      <UButton
                        color="neutral"
                        variant="outline"
                        icon="i-heroicons-archive-box"
                        :disabled="!selectedVersion || selectedVersion.state === 'archived'"
                        @click="openArchive"
                      >
                        归档
                      </UButton>
                    </div>
                  </div>

                  <div class="text-sm text-gray-500">
                    条目数量（当前页）：{{ itemsMeta.total }}
                    <span v-if="selectedVersion?.state !== 'draft'" class="ml-2">
                     （提示：非 draft 版本不可改条目）
                    </span>
                  </div>
                </div>
	              </UCard>
	            </div>
	          </div>
	        </div>
      </template>

      <template #footer>
        <div class="flex w-full justify-end gap-2 p-4 sm:p-5">
          <UButton color="neutral" variant="subtle" type="button" @click="closeDetail">
            {{ $t("common.close") }}
          </UButton>
          <UButton
            color="primary"
            type="button"
            :disabled="!active || isBase(active)"
            :title="active && isBase(active) ? '基础价目表为系统托管，不支持在此编辑' : ''"
            @click="active && openEdit(active)"
          >
            {{ $t("common.edit") }}
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- 新建版本 -->
    <UModal
      v-model:open="createVersionOpen"
      title="新建草稿版本"
      description="将创建一个新的 draft 版本；可选从某个版本复制（当前仅创建空版本）。"
      :close="true"
      :prevent-close="true"
      :ui="{ content: 'max-w-3xl w-full' }"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <form id="pricebook-version-create-form" class="space-y-4" @submit.prevent="submitCreateVersion">
            <UFormField label="从版本复制（可选）">
              <USelect v-model="createVersionForm.copyFromVersionId" :items="versionCopyItems" class="w-full" />
            </UFormField>
          </form>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="creatingVersion" @click="closeCreateVersion">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton type="submit" form="pricebook-version-create-form" color="primary" :loading="creatingVersion">
            {{ $t("common.submit") }}
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- 发布版本 -->
    <UModal
      v-model:open="publishOpen"
      title="发布版本"
      description="发布后该版本将变为 active 并参与查价；可选设置生效/失效时间。"
      :close="true"
      :prevent-close="true"
      :ui="{ content: 'max-w-3xl w-full' }"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <form id="pricebook-version-publish-form" class="space-y-4" @submit.prevent="submitPublish">
            <UFormField label="说明（可选）">
              <UTextarea v-model="publishForm.note" class="w-full" :rows="3" placeholder="例如：发布 v1" />
            </UFormField>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="生效时间（可选）">
                <UInput v-model="publishForm.effectiveAt" type="datetime-local" class="w-full" />
              </UFormField>
              <UFormField label="失效时间（可选）">
                <UInput v-model="publishForm.expiresAt" type="datetime-local" class="w-full" />
              </UFormField>
            </div>
          </form>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="publishing" @click="closePublish">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton type="submit" form="pricebook-version-publish-form" color="primary" :loading="publishing">
            发布
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- 归档版本 -->
    <UModal
      v-model:open="archiveOpen"
      title="归档版本"
      description="归档后该版本不再参与查价。"
      :close="true"
      :prevent-close="true"
      :ui="{ content: 'max-w-3xl w-full' }"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <form id="pricebook-version-archive-form" class="space-y-4" @submit.prevent="submitArchive">
            <UFormField label="说明（可选）">
              <UTextarea v-model="archiveForm.note" class="w-full" :rows="3" placeholder="例如：下线该版本" />
            </UFormField>
          </form>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="archiving" @click="closeArchive">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton type="submit" form="pricebook-version-archive-form" color="primary" :loading="archiving">
            归档
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- 编辑条目 -->
    <UModal
      v-model:open="itemsOpen"
      :title="itemsEditable ? '编辑条目（draft）' : '查看条目'"
      :description="
        itemsEditable
          ? '保存会执行 upsert：同 SKU 会更新；未填写的行会忽略；不会删除旧条目。'
          : '当前版本为只读；只有 draft 版本允许编辑与保存。'
      "
      :close="true"
      :prevent-close="true"
      :ui="{ content: 'max-w-6xl w-[90vw] mx-auto' }"
    >
      <template #body>
        <div class="p-4 sm:p-5 space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <p class="text-sm text-gray-500">
              当前版本：{{ selectedVersion ? `v${selectedVersion.version} · ${selectedVersion.state}` : "-" }}
            </p>
            <div class="flex gap-2">
              <UButton
                v-if="itemsEditable"
                color="neutral"
                variant="outline"
                icon="i-heroicons-plus"
                @click="addItemRow"
              >
                新增一行
              </UButton>
              <UButton color="neutral" variant="outline" icon="i-heroicons-arrow-path" :loading="itemsLoading" @click="refreshItems">
                刷新
              </UButton>
            </div>
          </div>

          <div class="overflow-auto rounded-lg border border-gray-200 dark:border-gray-800">
            <table class="min-w-full text-sm">
              <thead class="bg-gray-50 dark:bg-gray-900/40">
                <tr>
                  <th class="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-300">SKU</th>
                  <th class="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Base</th>
                  <th class="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Sale</th>
                  <th class="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Tax</th>
                  <th class="px-3 py-2 text-right font-medium text-gray-600 dark:text-gray-300">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(it, idx) in itemsForm" :key="idx" class="border-t border-gray-200 dark:border-gray-800">
                  <td class="px-3 py-2">
                    <div class="space-y-1">
                      <UInput
                        v-model.trim="it.sku_code"
                        placeholder="SKU Code（例如 BALL-BASKET-001-STD）"
                        class="w-80"
                        :disabled="!itemsEditable"
                      />
                      <p class="text-xs text-gray-500">
                        <span v-if="it.spu_name" class="mr-2">{{ it.spu_name }}</span>
                        <span v-if="it.spec_display">spec: {{ it.spec_display }}</span>
                      </p>
                      <div v-if="it.sku_id" class="flex items-center gap-2 text-xs text-gray-500">
                        <UButton
                          size="xs"
                          color="neutral"
                          variant="ghost"
                          icon="i-heroicons-clipboard-document"
                          title="复制 SKU ID（内部）"
                          @click="copyText(it.sku_id)"
                        />
                      </div>
                    </div>
                  </td>
                  <td class="px-3 py-2">
                    <UInput
                      v-model.number="it.base_amount_minor"
                      type="number"
                      placeholder="例如 19900"
                      class="w-40"
                      :disabled="!itemsEditable"
                    />
                    <p class="mt-1 text-xs text-gray-500">
                      {{ formatMinor(it.base_amount_minor, active?.currency || 'CNY') }}
                    </p>
                  </td>
                  <td class="px-3 py-2">
                    <UInput
                      v-model.number="it.sale_amount_minor"
                      type="number"
                      placeholder="可选"
                      class="w-40"
                      :disabled="!itemsEditable"
                    />
                    <p class="mt-1 text-xs text-gray-500">
                      {{ formatMinor(it.sale_amount_minor, active?.currency || 'CNY') }}
                    </p>
                  </td>
                  <td class="px-3 py-2">
                    <USwitch v-model="it.tax_included" :disabled="!itemsEditable" />
                  </td>
                  <td class="px-3 py-2 text-right">
                    <UButton
                      color="error"
                      variant="ghost"
                      icon="i-heroicons-trash"
                      :disabled="!itemsEditable"
                      @click="removeItemRow(idx)"
                    >
                      删除
                    </UButton>
                  </td>
                </tr>
                <tr v-if="itemsForm.length === 0">
                  <td colspan="5" class="px-3 py-8 text-center text-gray-500">暂无条目，点“新增一行”开始。</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="savingItems" @click="closeItems">
            {{ itemsEditable ? $t("common.cancel") : $t("common.close") }}
          </UButton>
          <UButton
            v-if="itemsEditable"
            color="primary"
            type="button"
            :loading="savingItems"
            @click="submitItems"
          >
            {{ $t("common.save") }}
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- 新建 -->
    <UModal
      v-model:open="createOpen"
      title="新增价格手册"
      description="创建后会自动生成 v1 草稿版本（draft）。"
      :close="true"
      :prevent-close="true"
      :ui="{ content: 'max-w-3xl w-full' }"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <form id="pricebook-create-form" class="space-y-4" @submit.prevent="submitCreate">
            <UFormField label="Code（唯一）" required>
              <UInput v-model.trim="createForm.code" class="w-full" placeholder="例如 tmall_standard" />
            </UFormField>
            <UFormField label="名称" required>
              <UInput v-model.trim="createForm.name" class="w-full" placeholder="例如 天猫标准价目表" />
            </UFormField>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="类型" required>
                <USelect v-model="createForm.type" :items="createTypeItems" class="w-full" />
              </UFormField>
              <UFormField label="币种" required>
                <UInput
                  v-model.trim="createForm.currency"
                  class="w-full"
                  placeholder="CNY"
                  maxlength="3"
                />
              </UFormField>
            </div>
            <UFormField label="描述">
              <UTextarea
                v-model="createForm.description"
                class="w-full"
                :rows="3"
                placeholder="可选：说明适用范围/用途"
              />
            </UFormField>
          </form>
        </div>
      </template>

      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="creating" @click="closeCreate">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton type="submit" form="pricebook-create-form" color="primary" :loading="creating">
            {{ $t("common.submit") }}
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- 编辑 -->
    <UModal
      v-model:open="editOpen"
      title="编辑价格手册"
      description="当前仅支持修改名称/描述/状态（code、币种、类型不可编辑）。"
      :close="true"
      :prevent-close="true"
      :ui="{ content: 'max-w-3xl w-full' }"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <form id="pricebook-edit-form" class="space-y-4" @submit.prevent="submitEdit">
            <UFormField label="Code（不可编辑）">
              <UInput :model-value="editForm.code" class="w-full" disabled />
            </UFormField>
            <UFormField label="名称" required>
              <UInput v-model.trim="editForm.name" class="w-full" placeholder="请输入名称" />
            </UFormField>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="类型（不可编辑）">
                <UInput :model-value="editForm.type" class="w-full" disabled />
              </UFormField>
              <UFormField label="币种（不可编辑）">
                <UInput :model-value="editForm.currency" class="w-full" disabled />
              </UFormField>
            </div>
            <UFormField label="状态">
              <USelect v-model="editForm.status" :items="editStatusItems" class="w-full" />
            </UFormField>
            <UFormField label="描述">
              <UTextarea
                v-model="editForm.description"
                class="w-full"
                :rows="3"
                placeholder="可选：说明适用范围/用途"
              />
            </UFormField>
          </form>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="editing" @click="closeEdit">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton type="submit" form="pricebook-edit-form" color="primary" :loading="editing">
            {{ $t("common.save") }}
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- 删除确认 -->
    <UModal
      v-model:open="removeOpen"
      title="删除价格手册"
      description="该操作不可恢复。"
      :close="true"
      :prevent-close="true"
      :ui="{ content: 'max-w-lg w-full' }"
    >
      <template #body>
        <div class="p-4 sm:p-5 text-sm text-gray-600 dark:text-gray-300">
          确认删除 <span class="font-medium">{{ removeTarget?.name }}</span>（{{ removeTarget?.code }}）？
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="removing" @click="closeRemove">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton color="error" type="button" :loading="removing" @click="submitRemove">
            {{ $t("common.delete") }}
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useToastAlert } from "~/composables/useToastAlert";
import {
  usePricebookApi,
  type Pricebook,
  type PricebookItem,
  type PricebookListParams,
  type PricebookVersion,
} from "~/composables/api";

const toast = useToastAlert();
const route = useRoute();

const api = usePricebookApi();

const loading = ref(false);
const creating = ref(false);
const editing = ref(false);
const removing = ref(false);

const filters = reactive<Required<Pick<PricebookListParams, "page" | "page_size">> &
  Omit<PricebookListParams, "page" | "page_size">>({
  keyword: "",
  type: null,
  currency: "",
  status: null,
  page: 1,
  page_size: 20,
});

const rows = ref<Pricebook[]>([]);
const meta = reactive({ total: 0, page: 1, page_size: 20 });

const typeItems = [
  { label: "全部类型", value: null },
  { label: "销售（sales）", value: "sales" },
  { label: "采购（purchase）", value: "purchase" },
];
const createTypeItems = [
  { label: "销售（sales）", value: "sales" },
  { label: "采购（purchase）", value: "purchase" },
];

const statusItems = [
  { label: "全部状态", value: null },
  { label: "active", value: "active" },
  { label: "archived", value: "archived" },
];

const editStatusItems = [
  { label: "active（启用）", value: "active" },
  { label: "archived（归档）", value: "archived" },
];

const versionSelectItems = computed(() => {
  return versions.value.map((v) => ({
    label: `v${v.version} · ${v.state}${v.id === active.value?.current_version_id ? "（当前）" : ""}`,
    value: v.id,
  }));
});

const versionCopyItems = computed(() => {
  return [
    { label: "不复制（空版本）", value: null },
    ...versions.value.map((v) => ({ label: `v${v.version} · ${v.state}`, value: v.id })),
  ];
});

const canPublish = computed(() => {
  if (!active.value?.id || isBase(active.value)) return false;
  if (!selectedVersion.value) return false;
  return String(selectedVersion.value.state).toLowerCase() === "draft";
});

const canEditItems = computed(() => {
  if (!active.value?.id || isBase(active.value)) return false;
  if (!selectedVersion.value) return false;
  return String(selectedVersion.value.state).toLowerCase() === "draft";
});

const itemsEditable = computed(() => canEditItems.value);

const columns: TableColumn<Pricebook>[] = [
  { accessorKey: "name", header: "名称" },
  { accessorKey: "code", header: "Code" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "currency", header: "币种" },
  { accessorKey: "status", header: "状态" },
  { id: "currentVersion", header: "当前版本" },
  { id: "updatedAt", header: "更新时间" },
  { id: "actions", header: "操作" },
];

const totalPages = computed(() => {
  const size = meta.page_size || 20;
  const t = meta.total || 0;
  return Math.max(1, Math.ceil(t / size));
});

const detailOpen = ref(false);
const active = ref<Pricebook | null>(null);
const detailTab = ref<"detail" | "versions">("detail");
const detailTabs = [
  { label: "详情", value: "detail" },
  { label: "版本与发布", value: "versions" },
];

const versionsLoading = ref(false);
const versions = ref<PricebookVersion[]>([]);
const selectedVersionId = ref("");
const selectedVersion = computed(
  () => versions.value.find((v) => v.id === selectedVersionId.value) || null,
);

const itemsLoading = ref(false);
const itemsMeta = reactive({ total: 0, page: 1, page_size: 200 });
const itemsForm = ref<
  Array<{
    sku_id: string;
    sku_code?: string;
    spu_name?: string;
    spec_display?: string;
    base_amount_minor: number | null;
    sale_amount_minor: number | null;
    tax_included: boolean;
  }>
>([]);

const createVersionOpen = ref(false);
const creatingVersion = ref(false);
const createVersionForm = reactive<{ copyFromVersionId: string | null }>({ copyFromVersionId: null });

const publishOpen = ref(false);
const publishing = ref(false);
const publishForm = reactive<{ note: string; effectiveAt: string; expiresAt: string }>({
  note: "",
  effectiveAt: "",
  expiresAt: "",
});

const archiveOpen = ref(false);
const archiving = ref(false);
const archiveForm = reactive<{ note: string }>({ note: "" });

const itemsOpen = ref(false);
const savingItems = ref(false);

const createOpen = ref(false);
const createForm = reactive({
  code: "",
  name: "",
  type: "sales" as "sales" | "purchase",
  currency: "CNY",
  description: "",
});

const editOpen = ref(false);
const editTarget = ref<Pricebook | null>(null);
const editForm = reactive({
  code: "",
  name: "",
  type: "",
  currency: "",
  status: "active" as string,
  description: "",
});

const removeOpen = ref(false);
const removeTarget = ref<Pricebook | null>(null);

function isBase(pb: Pricebook) {
  return (pb?.code || "").toLowerCase() === "base";
}

function statusColor(status: string) {
  const s = (status || "").toLowerCase();
  if (s === "active") return "green";
  if (s === "archived") return "gray";
  return "neutral";
}

function formatTime(v?: string) {
  if (!v) return "-";
  const d = new Date(v);
  if (Number.isNaN(d.getTime())) return v;
  return d.toISOString().slice(0, 19).replace("T", " ");
}

function shortId(id?: string | null) {
  const s = String(id || "").trim();
  if (!s) return "-";
  if (s.length <= 12) return s;
  return `${s.slice(0, 8)}…${s.slice(-4)}`;
}

async function copyText(v?: string | null) {
  const s = String(v || "").trim();
  if (!s) return;
  try {
    await navigator.clipboard.writeText(s);
    toast.add({ title: "已复制", color: "primary" });
  } catch {
    toast.add({ title: "复制失败", color: "error" });
  }
}

function currencySymbol(currency: string) {
  const cur = String(currency || "").trim().toUpperCase();
  if (cur === "CNY" || cur === "RMB") return "¥";
  if (cur === "USD") return "$";
  if (cur === "EUR") return "€";
  if (cur === "HKD") return "HK$";
  return cur ? `${cur} ` : "";
}

function formatMinor(minor: number | null | undefined, currency: string) {
  if (minor === null || minor === undefined) return "-";
  const v = Number(minor);
  if (Number.isNaN(v)) return "-";
  const amount = (v / 100).toFixed(2);
  return `${currencySymbol(currency)}${amount}`;
}

async function refresh() {
  loading.value = true;
  try {
    const res = await api.list({
      keyword: filters.keyword || undefined,
      type: filters.type || undefined,
      currency: filters.currency || undefined,
      status: filters.status || undefined,
      page: filters.page,
      page_size: filters.page_size,
    });
    rows.value = res.items || [];
    meta.total = res.meta?.total ?? 0;
    meta.page = res.meta?.page ?? filters.page;
    meta.page_size = res.meta?.page_size ?? filters.page_size;
  } catch (err: any) {
    toast.add({
      title: "加载价格手册失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    loading.value = false;
  }
}

function prevPage() {
  if (filters.page <= 1) return;
  filters.page -= 1;
  refresh();
}

function nextPage() {
  if (filters.page >= totalPages.value) return;
  filters.page += 1;
  refresh();
}

function view(pb: Pricebook) {
  active.value = pb;
  detailTab.value = "detail";
  detailOpen.value = true;
  refreshVersions();
}

function openCreate() {
  createOpen.value = true;
}

function blurActiveElement() {
  if (typeof document === "undefined") return;
  (document.activeElement as HTMLElement | null)?.blur?.();
}

function closeDetail() {
  blurActiveElement();
  detailOpen.value = false;
}

function closeCreate() {
  blurActiveElement();
  createOpen.value = false;
}

function closeEdit() {
  blurActiveElement();
  editOpen.value = false;
}

function closeRemove() {
  blurActiveElement();
  removeOpen.value = false;
}

function openEdit(pb: Pricebook) {
  if (isBase(pb)) {
    toast.add({ title: "基础价目表为系统托管，不支持在此编辑", color: "warning" });
    return;
  }
  blurActiveElement();
  detailOpen.value = false;
  active.value = pb;
  editTarget.value = pb;
  editForm.code = pb.code || "";
  editForm.name = pb.name || "";
  editForm.type = pb.type || "";
  editForm.currency = pb.currency || "";
  editForm.status = pb.status || "active";
  editForm.description = pb.description || "";
  editOpen.value = true;
}

async function refreshVersions() {
  if (!active.value?.id) return;
  versionsLoading.value = true;
  try {
    const res = await api.listVersions(active.value.id);
    versions.value = res.items || [];
    const preferred = active.value.current_version_id || versions.value[0]?.id || "";
    if (preferred && preferred !== selectedVersionId.value) {
      selectedVersionId.value = preferred;
    }
    await refreshItems();
  } catch (err: any) {
    toast.add({
      title: "加载版本失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    versionsLoading.value = false;
  }
}

watch(selectedVersionId, async () => {
  await refreshItems();
});

async function refreshItems() {
  if (!active.value?.id || !selectedVersionId.value) {
    itemsMeta.total = 0;
    itemsMeta.page = 1;
    itemsForm.value = [];
    return;
  }
  itemsLoading.value = true;
  try {
    const res = await api.listItems(active.value.id, selectedVersionId.value, {
      page: 1,
      page_size: itemsMeta.page_size,
    });
    itemsMeta.total = res.meta?.total ?? 0;
    itemsMeta.page = res.meta?.page ?? 1;
    itemsMeta.page_size = res.meta?.page_size ?? itemsMeta.page_size;

    const list = (res.items || []) as PricebookItem[];
    itemsForm.value = list.map((it) => ({
      sku_id: it.sku_id,
      sku_code: it.sku_code || (it.meta as any)?.sku_code || "",
      spu_name: it.spu_name || "",
      spec_display: it.spec_display || (it.meta as any)?.spec_display || "",
      base_amount_minor: (it.base_amount_minor ?? null) as any,
      sale_amount_minor: (it.sale_amount_minor ?? null) as any,
      tax_included: !!it.tax_included,
    }));
  } catch (err: any) {
    toast.add({
      title: "加载条目失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    itemsLoading.value = false;
  }
}

function openCreateVersion() {
  if (!active.value?.id) return;
  if (active.value && isBase(active.value)) {
    toast.add({ title: "基础价目表为系统托管，不支持在此创建版本", color: "warning" });
    return;
  }
  blurActiveElement();
  createVersionForm.copyFromVersionId = null;
  createVersionOpen.value = true;
}

function closeCreateVersion() {
  blurActiveElement();
  createVersionOpen.value = false;
}

async function submitCreateVersion() {
  if (!active.value?.id) return;
  if (active.value && isBase(active.value)) {
    toast.add({ title: "基础价目表为系统托管，不支持在此创建版本", color: "warning" });
    return;
  }
  creatingVersion.value = true;
  try {
    const created = await api.createVersion(active.value.id, {
      copy_from_version_id: createVersionForm.copyFromVersionId || undefined,
    });
    toast.add({ title: "已创建草稿版本", color: "primary" });
    closeCreateVersion();
    await refreshVersions();
    selectedVersionId.value = created.id;
    detailTab.value = "versions";
  } catch (err: any) {
    toast.add({
      title: "创建版本失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    creatingVersion.value = false;
  }
}

function openPublish() {
  if (!canPublish.value) {
    toast.add({ title: "仅 draft 版本可发布", color: "warning" });
    return;
  }
  blurActiveElement();
  publishForm.note = "";
  publishForm.effectiveAt = "";
  publishForm.expiresAt = "";
  publishOpen.value = true;
}

function closePublish() {
  blurActiveElement();
  publishOpen.value = false;
}

function toRFC3339(v: string): string | undefined {
  const s = (v || "").trim();
  if (!s) return undefined;
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return undefined;
  return d.toISOString();
}

async function submitPublish() {
  if (!active.value?.id || !selectedVersionId.value) return;
  if (!canPublish.value) {
    toast.add({ title: "仅 draft 版本可发布", color: "warning" });
    return;
  }
  publishing.value = true;
  try {
    await api.publishVersion(active.value.id, selectedVersionId.value, {
      note: publishForm.note?.trim() || undefined,
      effective_at: toRFC3339(publishForm.effectiveAt),
      expires_at: toRFC3339(publishForm.expiresAt),
    });
    toast.add({ title: "已发布", color: "primary" });
    closePublish();
    await refresh();
    const updated = rows.value.find((r) => r.id === active.value?.id) || null;
    if (updated) active.value = updated;
    await refreshVersions();
  } catch (err: any) {
    toast.add({
      title: "发布失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    publishing.value = false;
  }
}

function openArchive() {
  if (!active.value?.id || !selectedVersionId.value) return;
  blurActiveElement();
  archiveForm.note = "";
  archiveOpen.value = true;
}

function closeArchive() {
  blurActiveElement();
  archiveOpen.value = false;
}

async function submitArchive() {
  if (!active.value?.id || !selectedVersionId.value) return;
  archiving.value = true;
  try {
    await api.archiveVersion(active.value.id, selectedVersionId.value, {
      note: archiveForm.note?.trim() || undefined,
    });
    toast.add({ title: "已归档", color: "primary" });
    closeArchive();
    await refresh();
    const updated = rows.value.find((r) => r.id === active.value?.id) || null;
    if (updated) active.value = updated;
    await refreshVersions();
  } catch (err: any) {
    toast.add({
      title: "归档失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    archiving.value = false;
  }
}

function openItems() {
  if (!selectedVersion.value) {
    toast.add({ title: "请先选择版本", color: "warning" });
    return;
  }
  blurActiveElement();
  itemsOpen.value = true;
}

function closeItems() {
  blurActiveElement();
  itemsOpen.value = false;
}

function addItemRow() {
  itemsForm.value = [
    ...itemsForm.value,
    {
      sku_id: "",
      sku_code: "",
      spu_name: "",
      spec_display: "",
      base_amount_minor: null,
      sale_amount_minor: null,
      tax_included: false,
    },
  ];
}

function removeItemRow(idx: number) {
  itemsForm.value = itemsForm.value.filter((_, i) => i !== idx);
}

async function submitItems() {
  if (!active.value?.id || !selectedVersionId.value) return;
  if (!canEditItems.value) {
    toast.add({ title: "仅 draft 版本可编辑条目", color: "warning" });
    return;
  }

  const pending = itemsForm.value.filter((it) => !String(it.sku_id || "").trim() && String(it.sku_code || "").trim());
  if (pending.length > 0) {
    toast.add({
      title: "暂不支持仅用 SKU Code 新增条目",
      description: "新增条目需要先解析到 sku_id（下一步可做 SKU 选择器/自动解析）。",
      color: "warning",
    });
    return;
  }

  const items = itemsForm.value
    .map((it) => ({
      sku_id: (it.sku_id || "").trim(),
      base_amount_minor: it.base_amount_minor ?? undefined,
      sale_amount_minor: it.sale_amount_minor ?? undefined,
      tax_included: !!it.tax_included,
    }))
    .filter((it) => it.sku_id);

  if (items.length === 0) {
    toast.add({ title: "请至少填写一条 SKU 条目", color: "warning" });
    return;
  }

  savingItems.value = true;
  try {
    await api.upsertItems(active.value.id, selectedVersionId.value, { items });
    toast.add({ title: "已保存条目", color: "primary" });
    await refreshItems();
    closeItems();
  } catch (err: any) {
    toast.add({
      title: "保存条目失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    savingItems.value = false;
  }
}

async function submitCreate() {
  const code = createForm.code.trim();
  const name = createForm.name.trim();
  const currency = createForm.currency.trim().toUpperCase();
  const typ = createForm.type;
  if (!code || !name || !currency) {
    toast.add({ title: "请填写 Code/名称/币种", color: "warning" });
    return;
  }

  creating.value = true;
  try {
    await api.create({
      code,
      name,
      type: typ,
      currency,
      description: createForm.description?.trim() || undefined,
    });
    toast.add({ title: "已创建价格手册", color: "primary" });
    closeCreate();
    createForm.code = "";
    createForm.name = "";
    createForm.type = "sales";
    createForm.currency = "CNY";
    createForm.description = "";
    filters.page = 1;
    await refresh();
  } catch (err: any) {
    toast.add({
      title: "创建失败",
      description: err?.message || "请检查输入或稍后重试",
      color: "error",
    });
  } finally {
    creating.value = false;
  }
}

function confirmRemove(pb: Pricebook) {
  removeTarget.value = pb;
  removeOpen.value = true;
}

async function submitEdit() {
  const id = editTarget.value?.id;
  if (!id) return;
  if (editTarget.value && isBase(editTarget.value)) {
    toast.add({ title: "基础价目表为系统托管，不支持在此编辑", color: "warning" });
    return;
  }

  const name = editForm.name.trim();
  if (!name) {
    toast.add({ title: "请填写名称", color: "warning" });
    return;
  }

  editing.value = true;
  try {
    const updated = await api.update(id, {
      name,
      description: editForm.description?.trim() || undefined,
      status: (editForm.status || "active") as any,
    });
    toast.add({ title: "已保存", color: "primary" });
    closeEdit();

    const idx = rows.value.findIndex((r) => r.id === id);
    if (idx >= 0) rows.value[idx] = updated;
    if (active.value?.id === id) active.value = updated;

    await refresh();
  } catch (err: any) {
    toast.add({
      title: "保存失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    editing.value = false;
  }
}

async function submitRemove() {
  if (!removeTarget.value?.id) return;
  removing.value = true;
  try {
    await api.remove(removeTarget.value.id);
    toast.add({ title: "已删除", color: "primary" });
    closeRemove();
    removeTarget.value = null;
    await refresh();
  } catch (err: any) {
    toast.add({
      title: "删除失败",
      description: err?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    removing.value = false;
  }
}

onMounted(async () => {
  await refresh();
  if (route.query?.create === "1") {
    openCreate();
  }
});
</script>
