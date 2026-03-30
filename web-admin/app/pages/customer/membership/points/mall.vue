<template>
  <div class="space-y-6 p-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">积分商城</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">基于真实会籍数据展示可兑换权益（无伪造商品数据）</p>
      </div>
      <UButton color="neutral" variant="outline" icon="i-heroicons-arrow-down-tray" @click="exportData">
        导出数据
      </UButton>
    </div>

    <UAlert
      color="success"
      variant="soft"
      title="兑换链路已打通"
      description="可直接下单兑换权益；下方展示真实积分交易记录并支持筛选。"
    />

    <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
            <UIcon name="i-heroicons-star" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">积分总余额</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ totalPoints.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
            <UIcon name="i-heroicons-user-group" class="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">有积分客户</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ membersWithPoints.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
            <UIcon name="i-heroicons-gift" class="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <div class="ml-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">可兑换权益数</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ benefitItems.length.toLocaleString() }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <div ref="txPanelRef">
      <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">可兑换权益清单</h2>
          <UInput
            v-model="keyword"
            placeholder="搜索权益名称/类型"
            icon="i-heroicons-magnifying-glass"
            size="sm"
            @update:model-value="syncAllQuery"
          />
        </div>
      </template>

      <UTable :columns="columns" :data="filteredBenefits" :loading="loading">
        <template #name-cell="{ row }">
          <div>
            <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400">ID: {{ row.id }}</div>
          </div>
        </template>

        <template #type-cell="{ row }">
          <UBadge variant="soft" :color="row.type === 'bundle' ? 'warning' : 'success'">
            {{ row.type || 'single' }}
          </UBadge>
        </template>

        <template #status-cell="{ row }">
          <UBadge variant="soft" :color="row.status === 'active' ? 'success' : 'neutral'">
            {{ row.status || 'inactive' }}
          </UBadge>
        </template>

        <template #items-cell="{ row }">
          <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatItems(row.items) }}</span>
        </template>

        <template #actions-cell="{ row }">
          <UButton
            size="sm"
            color="primary"
            variant="soft"
            :disabled="row.status !== 'active' || !memberOptions.length"
            @click="openRedeemModal(row)"
          >
            立即兑换
          </UButton>
        </template>
      </UTable>

      <UAlert v-if="!filteredBenefits.length && !loading" color="gray" class="mt-4">
        暂无可兑换权益
      </UAlert>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">积分交易记录</h2>
          <div class="flex flex-wrap gap-2">
            <USelectMenu
              v-model="txCustomerId"
              :options="txCustomerOptions"
              value-attribute="value"
              option-attribute="label"
              searchable
              size="sm"
              class="min-w-56"
              placeholder="客户"
            />
            <USelectMenu
              v-model="txSourceTypes"
              :options="txSourceTypeOptions"
              value-attribute="value"
              option-attribute="label"
              multiple
              searchable
              size="sm"
              class="min-w-44"
              placeholder="来源（多选）"
            />
            <UInput
              v-model="txSourceId"
              size="sm"
              class="min-w-52"
              placeholder="来源ID（支持模糊）"
              @keyup.enter="applyTxFilters(true)"
            />
            <UButton
              size="sm"
              color="neutral"
              variant="outline"
              :disabled="!txSourceId.trim()"
              @click="clearSourceIdFilter"
            >
              清空来源ID
            </UButton>
            <UInput
              v-model="txCreatedFrom"
              type="datetime-local"
              size="sm"
              class="min-w-52"
              placeholder="开始时间"
            />
            <UInput
              v-model="txCreatedTo"
              type="datetime-local"
              size="sm"
              class="min-w-52"
              placeholder="结束时间"
            />
            <UButton
              size="sm"
              color="neutral"
              variant="soft"
              :disabled="!!txDateRangeError"
              @click="applyTxFilters(true)"
            >
              查询
            </UButton>
            <UButton
              size="sm"
              color="neutral"
              variant="soft"
              :disabled="transactionsLoading || !transactions.length"
              @click="exportTransactionsCsv(false)"
            >
              导出交易CSV
            </UButton>
            <USelectMenu
              v-model="exportLimitKey"
              :options="exportLimitOptions"
              value-attribute="value"
              option-attribute="label"
              size="sm"
              class="min-w-36"
              :disabled="exportingAllTransactions"
              placeholder="导出上限"
            />
            <UButton
              size="sm"
              color="neutral"
              variant="soft"
              :loading="exportingAllTransactions"
              :disabled="transactionsLoading || exportingAllTransactions"
              @click="openExportAllConfirm"
            >
              导出全部匹配CSV
            </UButton>
            <span v-if="exportingAllTransactions" class="text-xs text-gray-500 dark:text-gray-400">
              正在导出：{{ exportAllFetched }}/{{ exportAllTotalEstimate > 0 ? exportAllTotalEstimate : '?' }}
            </span>
            <UButton
              v-if="exportingAllTransactions"
              size="sm"
              color="warning"
              variant="outline"
              @click="cancelExportAll"
            >
              取消导出
            </UButton>
            <UButton
              size="sm"
              color="neutral"
              variant="outline"
              @click="resetAllFilters"
            >
              重置
            </UButton>
          </div>
          <div v-if="txDateRangeError" class="w-full text-xs text-red-500 dark:text-red-400">
            {{ txDateRangeError }}
          </div>
          <div v-if="activeTxFilterChips.length" class="flex w-full flex-wrap gap-2">
            <UButton
              v-for="chip in activeTxFilterChips"
              :key="chip.id"
              size="xs"
              color="neutral"
              variant="soft"
              @click="removeTxFilter(chip.key, chip.value)"
            >
              {{ chip.label }} ×
            </UButton>
            <UButton
              size="xs"
              color="warning"
              variant="soft"
              @click="resetAllFilters"
            >
              清空筛选
            </UButton>
          </div>
          <div class="w-full text-xs text-gray-500 dark:text-gray-400">
            说明：结束时间按分钟输入时，查询会自动补齐到该分钟的 59 秒。
          </div>
          <div v-if="exportRecords.length" class="w-full rounded-lg border border-gray-200 p-3 dark:border-gray-800">
            <div class="mb-2 flex items-center justify-between">
              <div class="text-xs font-medium text-gray-700 dark:text-gray-200">最近导出记录</div>
              <div class="flex items-center gap-2">
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="!filteredExportRecords.length"
                  @click="toggleSelectAllFilteredExportRecords"
                >
                  {{ allFilteredExportRecordsSelected ? '取消全选' : '全选当前列表' }}
                </UButton>
                <UButton
                  size="xs"
                  color="warning"
                  variant="ghost"
                  :disabled="!selectedExportRecordIds.length"
                  @click="removeSelectedExportRecords"
                >
                  批量删除({{ selectedExportRecordIds.length }})
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="!exportRecords.length"
                  @click="keepOnlySuccessExportRecords"
                >
                  仅保留成功
                </UButton>
                <USelectMenu
                  v-model="exportRecordPresetKey"
                  :options="exportRecordPresetOptions"
                  value-attribute="value"
                  option-attribute="label"
                  size="xs"
                  class="min-w-36"
                  @update:model-value="onExportRecordPresetChange"
                />
                <USelectMenu
                  v-model="selectedSavedExportRecordPresetId"
                  :options="savedExportRecordPresetOptions"
                  value-attribute="value"
                  option-attribute="label"
                  size="xs"
                  class="min-w-40"
                  placeholder="个人快捷预设"
                />
                <UButton
                  size="xs"
                  color="primary"
                  variant="ghost"
                  :disabled="!selectedSavedExportRecordPresetId"
                  @click="applySavedExportRecordPreset"
                >
                  应用收藏
                </UButton>
                <UButton size="xs" color="neutral" variant="ghost" @click="saveCurrentExportRecordPreset">
                  保存当前
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="!selectedSavedExportRecordPresetId"
                  @click="renameSavedExportRecordPreset"
                >
                  重命名
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="!selectedSavedExportRecordPresetId"
                  @click="cloneSavedExportRecordPreset"
                >
                  克隆
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="selectedSavedExportRecordPresetIndex <= 0"
                  @click="moveSavedExportRecordPreset(-1)"
                >
                  上移
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="selectedSavedExportRecordPresetIndex < 0 || selectedSavedExportRecordPresetIndex >= savedExportRecordPresets.length - 1"
                  @click="moveSavedExportRecordPreset(1)"
                >
                  下移
                </UButton>
                <UButton
                  size="xs"
                  color="error"
                  variant="ghost"
                  :disabled="!selectedSavedExportRecordPresetId"
                  @click="removeSavedExportRecordPreset"
                >
                  删除收藏
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="!savedExportRecordPresets.length"
                  @click="showSavedExportRecordPresetCards = !showSavedExportRecordPresetCards"
                >
                  {{ showSavedExportRecordPresetCards ? '收起卡片' : '展开卡片' }}
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="!savedExportRecordPresets.length"
                  @click="exportSavedExportRecordPresets"
                >
                  导出收藏
                </UButton>
                <UButton size="xs" color="neutral" variant="ghost" @click="openSavedExportRecordPresetImport">
                  导入收藏
                </UButton>
                <input
                  ref="savedExportRecordPresetImportRef"
                  type="file"
                  class="hidden"
                  accept="application/json,.json"
                  @change="handleSavedExportRecordPresetImport"
                >
                <UInput
                  v-model="exportRecordKeyword"
                  size="xs"
                  class="min-w-28"
                  placeholder="搜索记录"
                />
                <USelectMenu
                  v-model="exportRecordSortOrder"
                  :options="exportRecordSortOptions"
                  value-attribute="value"
                  option-attribute="label"
                  size="xs"
                  class="min-w-24"
                />
                <USelectMenu
                  v-model="exportRecordStatusFilter"
                  :options="exportRecordStatusOptions"
                  value-attribute="value"
                  option-attribute="label"
                  size="xs"
                  class="min-w-24"
                />
                <USelectMenu
                  v-model="exportRecordModeFilter"
                  :options="exportRecordModeOptions"
                  value-attribute="value"
                  option-attribute="label"
                  size="xs"
                  class="min-w-24"
                />
                <USelectMenu
                  v-model="exportRecordTimeRangeFilter"
                  :options="exportRecordTimeRangeOptions"
                  value-attribute="value"
                  option-attribute="label"
                  size="xs"
                  class="min-w-24"
                />
                <USelectMenu
                  v-model="exportRecordRowsRangeFilter"
                  :options="exportRecordRowsRangeOptions"
                  value-attribute="value"
                  option-attribute="label"
                  size="xs"
                  class="min-w-28"
                />
                <UButton size="xs" color="neutral" variant="ghost" @click="resetExportRecordFilters">
                  重置记录筛选
                </UButton>
                <UButton size="xs" color="neutral" variant="ghost" @click="clearExportRecords">清空</UButton>
              </div>
            </div>
            <div v-if="selectedSavedExportRecordPresetSummary" class="mb-2 text-xs text-gray-500 dark:text-gray-400">
              收藏预览：{{ selectedSavedExportRecordPresetSummary }}
            </div>
            <div
              v-if="showSavedExportRecordPresetCards && savedExportRecordPresets.length"
              class="mb-3 grid grid-cols-1 gap-2 md:grid-cols-2"
            >
              <div class="col-span-full mb-1 flex flex-wrap items-center gap-2">
                <UInput
                  v-model="savedPresetCardKeyword"
                  size="xs"
                  class="min-w-28"
                  placeholder="搜索收藏名称"
                />
                <USelectMenu
                  v-model="savedPresetCardSort"
                  :options="savedPresetCardSortOptions"
                  value-attribute="value"
                  option-attribute="label"
                  size="xs"
                  class="min-w-32"
                />
              </div>
              <div
                v-for="preset in filteredSavedExportRecordPresetCards"
                :key="preset.id"
                class="rounded border border-gray-200 p-2 text-xs dark:border-gray-800"
              >
                <div class="mb-1 flex items-center justify-between gap-2">
                  <button
                    type="button"
                    class="truncate font-medium text-gray-800 hover:text-primary-600 dark:text-gray-100 dark:hover:text-primary-400"
                    @click="selectSavedExportRecordPreset(preset.id)"
                  >
                    {{ preset.name }}
                  </button>
                  <label class="flex items-center gap-1 text-[11px] text-gray-500 dark:text-gray-400">
                    <input
                      :checked="compareSavedExportRecordPresetIds.includes(preset.id)"
                      type="checkbox"
                      class="h-3.5 w-3.5 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                      @change="toggleCompareSavedExportRecordPreset(preset.id, ($event.target as HTMLInputElement).checked)"
                    >
                    对比
                  </label>
                </div>
                <div class="truncate text-gray-600 dark:text-gray-300">{{ formatSavedExportRecordPresetSummary(preset) }}</div>
                <div class="mt-1 text-[11px] text-gray-500 dark:text-gray-400">更新于：{{ formatDate(preset.updatedAt) }}</div>
              </div>
            </div>
            <div
              v-if="showSavedExportRecordPresetCards && comparedSavedExportRecordPresets.length >= 2"
              class="mb-3 rounded border border-gray-200 p-2 text-xs dark:border-gray-800"
            >
              <div class="mb-1 font-medium text-gray-700 dark:text-gray-200">预设对比</div>
              <div class="space-y-1 text-gray-600 dark:text-gray-300">
                <div v-for="row in comparedSavedExportRecordPresetRows" :key="row.key">
                  {{ row.label }}：{{ row.values.join(' ｜ ') }}
                </div>
              </div>
            </div>
            <div class="space-y-2">
              <div
                v-for="item in filteredExportRecords"
                :key="item.id"
                class="flex flex-wrap items-center gap-2 rounded px-2 py-1 text-xs text-gray-600 dark:text-gray-300"
                :class="highlightedExportRecordId === item.id ? 'bg-green-50 ring-1 ring-green-300 dark:bg-green-900/20 dark:ring-green-700' : ''"
                :data-export-record-id="item.id"
              >
                <input
                  :checked="selectedExportRecordIds.includes(item.id)"
                  type="checkbox"
                  class="h-3.5 w-3.5 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  @change="toggleExportRecordSelection(item.id, ($event.target as HTMLInputElement).checked)"
                >
                <span class="font-mono text-gray-500 dark:text-gray-400">{{ item.time }}</span>
                <UBadge :color="item.status === 'success' ? 'success' : (item.status === 'cancelled' ? 'warning' : 'error')" variant="soft" size="xs">
                  {{ item.status === 'success' ? '成功' : (item.status === 'cancelled' ? '已取消' : '失败') }}
                </UBadge>
                <UBadge color="neutral" variant="soft" size="xs">{{ item.mode === 'all' ? '全部匹配' : '当前页' }}</UBadge>
                <UBadge v-if="highlightedExportRecordId === item.id" color="success" variant="soft" size="xs">刚重导</UBadge>
                <span>条数 {{ item.rows }}</span>
                <span v-if="item.truncated" class="text-amber-600 dark:text-amber-400">已截断</span>
                <span class="truncate">筛选：{{ item.summary }}</span>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="!item.filters"
                  :title="item.filters ? '' : '旧记录缺少筛选快照，无法重导'"
                  @click="reExportFromRecord(item, 'current')"
                >
                  重导当前页
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :disabled="!item.filters"
                  :title="item.filters ? '' : '旧记录缺少筛选快照，无法重导'"
                  @click="reExportFromRecord(item, 'all')"
                >
                  重导全部
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  :loading="downloadingRecordId === item.id"
                  :disabled="!item.filters || !!downloadingRecordId || exportingAllTransactions"
                  :title="item.filters ? '' : '旧记录缺少筛选快照，无法下载'"
                  @click="downloadLatestFromRecord(item)"
                >
                  下载最新
                </UButton>
                <UButton
                  size="xs"
                  color="primary"
                  variant="ghost"
                  :disabled="!item.filters"
                  @click="reuseExportRecordFilters(item)"
                >
                  复用筛选
                </UButton>
                <UButton
                  size="xs"
                  color="neutral"
                  variant="ghost"
                  @click="copyExportRecordSummary(item)"
                >
                  {{ copiedExportRecordId === item.id ? '已复制' : '复制摘要' }}
                </UButton>
                <UButton
                  size="xs"
                  color="error"
                  variant="ghost"
                  @click="removeExportRecord(item.id)"
                >
                  删除
                </UButton>
              </div>
              <div v-if="!filteredExportRecords.length" class="text-xs text-gray-500 dark:text-gray-400">
                当前筛选下暂无导出记录
              </div>
            </div>
          </div>
        </div>
      </template>
      <UTable :columns="transactionColumns" :data="transactions" :loading="transactionsLoading">
        <template #customerId-cell="{ row }">
          <div>
            <div class="font-medium text-gray-900 dark:text-white">{{ resolveCustomerName(row.customerId) }}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400">ID: {{ row.customerId }}</div>
          </div>
        </template>

        <template #delta-cell="{ row }">
          <span :class="row.delta >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'">
            {{ row.delta >= 0 ? `+${row.delta}` : row.delta }}
          </span>
        </template>

        <template #createdAt-cell="{ row }">
          <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDate(row.createdAt) }}</span>
        </template>
        <template #sourceId-cell="{ row }">
          <div
            class="flex items-center gap-2"
            :data-highlighted-tx="highlightedTxId && row.id === highlightedTxId ? '1' : '0'"
          >
            <button
              type="button"
              class="font-mono text-xs text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300"
              :disabled="!row.sourceId"
              @click="copySourceId(row.sourceId)"
            >
              {{ copiedSourceId === row.sourceId ? '已复制' : (row.sourceId || '-') }}
            </button>
            <UButton size="xs" color="neutral" variant="soft" :disabled="!row.sourceId" @click.stop="quickFilterBySourceId(row.sourceId)">
              筛选
            </UButton>
            <UBadge
              v-if="highlightedTxId && row.id === highlightedTxId"
              color="success"
              variant="soft"
              size="xs"
              class="cursor-pointer"
              @click.stop="toggleHighlightPinned(row.id)"
            >
              刚兑换
            </UBadge>
          </div>
        </template>
      </UTable>
      <UAlert v-if="!transactionsLoading && !transactions.length" color="gray" class="mt-4">
        暂无积分交易记录
      </UAlert>

      <template #footer>
        <div class="flex items-center justify-between">
          <div class="space-y-1 text-sm text-gray-500 dark:text-gray-400">
            <div>共 {{ transactionsTotal }} 条记录</div>
            <div v-if="txResultSummary">{{ txResultSummary }}</div>
          </div>
          <UPagination
            v-model="txPage"
            :total="transactionsTotal"
            :page-count="txPageCount"
            @update:model-value="reloadTransactions"
          />
        </div>
      </template>
      </UCard>
    </div>

    <UModal v-model:open="redeemModalOpen">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">兑换权益</h3>
          </template>
          <form class="space-y-4" @submit.prevent="submitRedeem">
            <UFormField label="权益">
              <UInput :model-value="selectedBenefit?.name || '-'" disabled />
            </UFormField>

            <UFormField label="兑换客户" required>
              <USelect
                v-model="redeemForm.customerId"
                :items="memberOptions"
                option-attribute="label"
                value-attribute="value"
                placeholder="选择客户"
              />
            </UFormField>

            <UFormField label="消耗积分" required>
              <UInput v-model.number="redeemForm.pointsCost" type="number" min="1" />
            </UFormField>

            <UFormField label="备注">
              <UTextarea v-model="redeemForm.reason" :rows="2" placeholder="可选：填写兑换原因" />
            </UFormField>

            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" type="button" @click="redeemModalOpen = false">取消</UButton>
              <UButton color="primary" type="submit" :loading="redeemSubmitting" :disabled="!canSubmitRedeem">
                确认兑换
              </UButton>
            </div>
          </form>
        </UCard>
      </template>
    </UModal>

    <UModal v-model:open="exportAllConfirmOpen">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">确认导出全部匹配交易</h3>
          </template>
          <div class="space-y-2 text-sm text-gray-600 dark:text-gray-300">
            <p>当前筛选：{{ txResultSummary }}</p>
            <p>
              预计条数：
              <span v-if="exportAllEstimateLoading">计算中...</span>
              <span v-else>{{ exportAllEstimateTotal }}</span>
            </p>
            <p>导出上限：{{ selectedExportLimit > 0 ? selectedExportLimit : '不限' }}</p>
            <p v-if="exportAllWillTruncate" class="text-amber-600 dark:text-amber-400">
              预计会截断：仅导出前 {{ selectedExportLimit }} 条
            </p>
          </div>
          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" @click="cancelExportAllConfirm">取消</UButton>
              <UButton
                color="primary"
                :loading="exportAllEstimateLoading || exportingAllTransactions"
                :disabled="exportAllEstimateLoading || exportingAllTransactions"
                @click="confirmExportAll"
              >
                开始导出
              </UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCustomerApi } from '~/composables/api/useCustomer'
import { useMembershipAdminApi } from '~/composables/api/useMembership'
import { useManagedQuerySync } from '~/composables/useManagedQuerySync'
import { collectManagedQueryValues, createManagedQueryKeySet, warnUnknownManagedQueryKeys } from '~/utils/query-sync-debug'
import type { MembershipBenefit, MembershipTokenTransaction, RedeemPointsBenefitResult } from '~/types/membership'
import type { MembershipInsight } from '~/types/customer'
import { useToast } from '#imports'

const route = useRoute()
const router = useRouter()
const customerApi = useCustomerApi()
const membershipApi = useMembershipAdminApi()
const toast = useToast()

const loading = ref(false)
const keyword = ref('')
const members = ref<MembershipInsight[]>([])
const benefitItems = ref<MembershipBenefit[]>([])
const transactions = ref<MembershipTokenTransaction[]>([])
const transactionsTotal = ref(0)
const transactionsLoading = ref(false)
const txPage = ref(1)
const txPageSize = ref(20)
const txCustomerId = ref('all')
const txSourceTypes = ref<string[]>([])
const txSourceId = ref('')
const txCreatedFrom = ref('')
const txCreatedTo = ref('')
const highlightedTxId = ref('')
const highlightedTxPinned = ref(false)
const txPanelRef = ref<HTMLElement | null>(null)
const copiedSourceId = ref('')
const exportingAllTransactions = ref(false)
const exportAllFetched = ref(0)
const exportAllTotalEstimate = ref(0)
const exportLimitKey = ref('5000')
const exportAllConfirmOpen = ref(false)
const exportAllEstimateLoading = ref(false)
const exportAllEstimateTotal = ref(0)
const exportRecords = ref<Array<{
  id: string;
  ts: number;
  time: string;
  mode: 'current' | 'all';
  rows: number;
  truncated: boolean;
  summary: string;
  status: 'success' | 'cancelled' | 'failed';
  filters?: {
    customerId: string;
    sourceTypes: string[];
    sourceId: string;
    createdFrom: string;
    createdTo: string;
    keyword: string;
  };
}>>([])
const EXPORT_RECORDS_STORAGE_KEY = 'membership.points.mall.export.records'
const EXPORT_RECORD_SAVED_PRESETS_STORAGE_KEY = 'membership.points.mall.export.saved-presets'
const copiedExportRecordId = ref('')
const exportRecordStatusFilter = ref<'all' | 'success' | 'cancelled' | 'failed'>('all')
const exportRecordModeFilter = ref<'all' | 'current' | 'all_mode'>('all')
const exportRecordSortOrder = ref<'desc' | 'asc'>('desc')
const exportRecordTimeRangeFilter = ref<'all' | '1d' | '7d' | '30d'>('all')
const exportRecordRowsRangeFilter = ref<'all' | 'zero' | '1_100' | '101_1000' | '1000_plus'>('all')
const exportRecordKeyword = ref('')
const exportRecordPresetKey = ref<'custom' | 'recent_success' | 'recent_failed' | 'all_mode_recent' | 'high_volume'>('custom')
const selectedSavedExportRecordPresetId = ref('')
const savedExportRecordPresetImportRef = ref<HTMLInputElement | null>(null)
const showSavedExportRecordPresetCards = ref(false)
const compareSavedExportRecordPresetIds = ref<string[]>([])
const savedPresetCardKeyword = ref('')
const savedPresetCardSort = ref<'updated_desc' | 'name_asc'>('updated_desc')
const savedExportRecordPresets = ref<Array<{
  id: string;
  name: string;
  updatedAt: string;
  filters: {
    status: 'all' | 'success' | 'cancelled' | 'failed';
    mode: 'all' | 'current' | 'all_mode';
    sort: 'desc' | 'asc';
    timeRange: 'all' | '1d' | '7d' | '30d';
    rowsRange: 'all' | 'zero' | '1_100' | '101_1000' | '1000_plus';
    keyword: string;
  };
}>>([])
const highlightedExportRecordId = ref('')
const downloadingRecordId = ref('')
const selectedExportRecordIds = ref<string[]>([])
let exportAbortController: AbortController | null = null
let highlightedTxTimer: ReturnType<typeof setTimeout> | null = null
let copiedSourceIdTimer: ReturnType<typeof setTimeout> | null = null
let copiedExportRecordTimer: ReturnType<typeof setTimeout> | null = null
let highlightedExportRecordTimer: ReturnType<typeof setTimeout> | null = null
let shouldHighlightNextExportRecord = false
const redeemModalOpen = ref(false)
const redeemSubmitting = ref(false)
const selectedBenefit = ref<MembershipBenefit | null>(null)
const txFiltersInitialized = ref(false)
const querySyncReady = ref(false)
const redeemForm = ref({
  customerId: '',
  pointsCost: 1,
  reason: '',
  sourceId: '',
})

const columns = [
  { accessorKey: 'name', header: '权益名称' },
  { accessorKey: 'type', header: '类型' },
  { accessorKey: 'items', header: '内容' },
  { accessorKey: 'status', header: '状态' },
  { id: 'actions', header: '操作' },
]

const transactionColumns = [
  { accessorKey: 'customerId', header: '客户' },
  { accessorKey: 'delta', header: '积分变动' },
  { accessorKey: 'sourceType', header: '来源' },
  { accessorKey: 'sourceId', header: '来源ID' },
  { accessorKey: 'createdAt', header: '时间' },
]
const txSourceTypeOptions = [
  { label: '商城兑换', value: 'points_mall_redeem' },
  { label: '手工调整', value: 'manual' },
]
const txPageCount = computed(() => Math.max(1, Math.ceil(transactionsTotal.value / txPageSize.value)))
const txDateRangeError = computed(() => {
  const from = txCreatedFrom.value.trim()
  const to = txCreatedTo.value.trim()
  if (!from || !to) return ''
  const fromTime = new Date(from).getTime()
  const toTime = new Date(to).getTime()
  if (Number.isNaN(fromTime) || Number.isNaN(toTime)) return ''
  return fromTime > toTime ? '开始时间不能晚于结束时间' : ''
})

const totalPoints = computed(() => members.value.reduce((sum, item) => sum + (Number(item.snapshot?.points ?? item.customer?.points ?? 0) || 0), 0))
const membersWithPoints = computed(() => members.value.filter((item) => (Number(item.snapshot?.points ?? item.customer?.points ?? 0) || 0) > 0).length)
const memberOptions = computed(() =>
  members.value
    .map((item) => ({
      label: `${item.customer?.name || '未知客户'} (${item.customer?.id || '-'})`,
      value: item.customer?.id || '',
    }))
    .filter((item) => item.value !== ''),
)
const txCustomerOptions = computed(() => [
  { label: '全部客户', value: 'all' },
  ...memberOptions.value,
])
const txSourceTypeMap = computed(() =>
  Object.fromEntries(txSourceTypeOptions.map((item) => [item.value, item.label] as const)),
)
const canSubmitRedeem = computed(() =>
  !!selectedBenefit.value &&
  redeemForm.value.customerId.trim() !== '' &&
  Number(redeemForm.value.pointsCost) > 0,
)

type TxFilterChip = {
  id: string;
  key: 'customer' | 'sourceType' | 'sourceId' | 'from' | 'to';
  value?: string;
  label: string;
}

const activeTxFilterChips = computed<TxFilterChip[]>(() => {
  const chips: TxFilterChip[] = []
  if (txCustomerId.value !== 'all') {
    const customerLabel = txCustomerOptions.value.find((item) => item.value === txCustomerId.value)?.label || txCustomerId.value
    chips.push({ id: `customer:${txCustomerId.value}`, key: 'customer', label: `客户: ${customerLabel}` })
  }
  txSourceTypes.value.forEach((sourceType) => {
    chips.push({
      id: `sourceType:${sourceType}`,
      key: 'sourceType',
      value: sourceType,
      label: `来源: ${txSourceTypeMap.value[sourceType] || sourceType}`,
    })
  })
  if (txSourceId.value.trim()) {
    chips.push({ id: 'sourceId', key: 'sourceId', label: `来源ID: ${txSourceId.value.trim()}` })
  }
  if (txCreatedFrom.value.trim()) {
    chips.push({ id: 'from', key: 'from', label: `开始: ${txCreatedFrom.value.trim()}` })
  }
  if (txCreatedTo.value.trim()) {
    chips.push({ id: 'to', key: 'to', label: `结束: ${txCreatedTo.value.trim()}` })
  }
  return chips
})

const txResultSummary = computed(() => {
  if (!activeTxFilterChips.value.length) return '当前为全部数据'
  return `当前筛选：${activeTxFilterChips.value.map((chip) => chip.label).join('，')}`
})

const filteredBenefits = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  const list = benefitItems.value
  if (!q) return list
  return list.filter((item) => `${item.name} ${item.type}`.toLowerCase().includes(q))
})

const formatItems = (items?: any) => {
  if (!items) return '-'
  if (Array.isArray(items)) return `${items.length} 项`
  if (typeof items === 'object') return '对象配置'
  return String(items)
}

const formatDate = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

const resolveCustomerName = (customerId: string) => {
  const match = members.value.find((item) => item.customer?.id === customerId)
  return match?.customer?.name || '未知客户'
}

const toRfc3339 = (value: string, endOfRange = false) => {
  const raw = value.trim()
  if (!raw) return undefined
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return undefined
  if (endOfRange && raw.length <= 16) {
    date.setSeconds(59, 999)
  }
  return date.toISOString()
}

const withTxFilterBatch = async (fn: () => Promise<void> | void) => {
  txFiltersInitialized.value = false
  if (txFilterTimer) {
    clearTimeout(txFilterTimer)
    txFilterTimer = null
  }
  try {
    await fn()
  } finally {
    txFiltersInitialized.value = true
  }
}

const applyRedeemFocusedFilters = async (customerId: string) => {
  await withTxFilterBatch(() => {
    txCustomerId.value = customerId
    txSourceTypes.value = ['points_mall_redeem']
    txPage.value = 1
  })
  await syncAllQuery()
}

const scrollToHighlightedTransaction = async () => {
  await nextTick()
  const panel = txPanelRef.value
  if (!panel) return
  panel.scrollIntoView({ behavior: 'smooth', block: 'start' })
  await nextTick()
  const highlighted = panel.querySelector('[data-highlighted-tx="1"]') as HTMLElement | null
  if (highlighted) {
    highlighted.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
}

const scheduleClearHighlightedTx = () => {
  if (highlightedTxPinned.value) return
  if (highlightedTxTimer) {
    clearTimeout(highlightedTxTimer)
    highlightedTxTimer = null
  }
  if (!highlightedTxId.value) return
  highlightedTxTimer = setTimeout(() => {
    highlightedTxId.value = ''
    highlightedTxTimer = null
  }, 8000)
}

const toggleHighlightPinned = (rowId: string) => {
  if (!highlightedTxId.value || highlightedTxId.value !== rowId) return
  highlightedTxPinned.value = !highlightedTxPinned.value
  if (highlightedTxPinned.value) {
    if (highlightedTxTimer) {
      clearTimeout(highlightedTxTimer)
      highlightedTxTimer = null
    }
    toast.add({ title: '已固定高亮', color: 'info' })
    return
  }
  toast.add({ title: '已取消固定', color: 'info' })
  scheduleClearHighlightedTx()
}

const copySourceId = async (value?: string) => {
  const text = (value || '').trim()
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    copiedSourceId.value = text
    if (copiedSourceIdTimer) {
      clearTimeout(copiedSourceIdTimer)
      copiedSourceIdTimer = null
    }
    copiedSourceIdTimer = setTimeout(() => {
      copiedSourceId.value = ''
      copiedSourceIdTimer = null
    }, 2000)
    toast.add({ title: '已复制来源ID', color: 'success' })
  } catch {
    toast.add({ title: '复制失败', description: '请手动复制', color: 'warning' })
  }
}

const quickFilterBySourceId = async (value?: string) => {
  const sourceId = (value || '').trim()
  if (!sourceId) return
  await withTxFilterBatch(() => {
    txSourceId.value = sourceId
    txPage.value = 1
  })
  await syncAllQuery()
  await loadTransactions()
  toast.add({ title: '已按来源ID筛选', description: sourceId, color: 'info' })
}

const clearSourceIdFilter = async () => {
  if (!txSourceId.value.trim()) return
  await withTxFilterBatch(() => {
    txSourceId.value = ''
    txPage.value = 1
  })
  await syncAllQuery()
  await loadTransactions()
}

const removeTxFilter = async (key: TxFilterChip['key'], value?: string) => {
  await withTxFilterBatch(() => {
    if (key === 'customer') txCustomerId.value = 'all'
    if (key === 'sourceType') {
      txSourceTypes.value = txSourceTypes.value.filter((item) => item !== value)
    }
    if (key === 'sourceId') txSourceId.value = ''
    if (key === 'from') txCreatedFrom.value = ''
    if (key === 'to') txCreatedTo.value = ''
    txPage.value = 1
  })
  await syncAllQuery()
  await loadTransactions()
}

const escapeCsv = (value: unknown) => {
  const text = String(value ?? '')
  if (/[",\n]/.test(text)) {
    return `"${text.replace(/"/g, '""')}"`
  }
  return text
}

const formatFileStamp = (date: Date) => {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}${pad(date.getMonth() + 1)}${pad(date.getDate())}-${pad(date.getHours())}${pad(date.getMinutes())}${pad(date.getSeconds())}`
}

const sanitizeFileToken = (value: string) => value.replace(/[^a-zA-Z0-9_-]/g, '_')
const exportLimitOptions = [
  { label: '上限 5000', value: '5000' },
  { label: '上限 10000', value: '10000' },
  { label: '不限', value: '0' },
]

const selectedExportLimit = computed(() => {
  const raw = Number(exportLimitKey.value)
  if (!Number.isFinite(raw) || raw < 0) return 5000
  return Math.floor(raw)
})
const exportAllWillTruncate = computed(() =>
  selectedExportLimit.value > 0 && exportAllEstimateTotal.value > selectedExportLimit.value,
)
const exportRecordStatusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '成功', value: 'success' },
  { label: '已取消', value: 'cancelled' },
  { label: '失败', value: 'failed' },
]
const exportRecordSortOptions = [
  { label: '最新在前', value: 'desc' },
  { label: '最早在前', value: 'asc' },
]
const exportRecordModeOptions = [
  { label: '全部范围', value: 'all' },
  { label: '当前页', value: 'current' },
  { label: '全部匹配', value: 'all_mode' },
]
const exportRecordTimeRangeOptions = [
  { label: '全部时间', value: 'all' },
  { label: '近1天', value: '1d' },
  { label: '近7天', value: '7d' },
  { label: '近30天', value: '30d' },
]
const exportRecordRowsRangeOptions = [
  { label: '全部条数', value: 'all' },
  { label: '0条', value: 'zero' },
  { label: '1-100', value: '1_100' },
  { label: '101-1000', value: '101_1000' },
  { label: '1000+', value: '1000_plus' },
]
const exportRecordPresetOptions = [
  { label: '自定义筛选', value: 'custom' },
  { label: '最近成功', value: 'recent_success' },
  { label: '最近失败', value: 'recent_failed' },
  { label: '全部匹配(近7天)', value: 'all_mode_recent' },
  { label: '大批量(1000+)', value: 'high_volume' },
]
const savedPresetCardSortOptions = [
  { label: '最近修改', value: 'updated_desc' },
  { label: '名称 A-Z', value: 'name_asc' },
]
const savedExportRecordPresetOptions = computed(() =>
  savedExportRecordPresets.value.map((item) => ({
    label: item.name,
    value: item.id,
  })),
)
const filteredSavedExportRecordPresetCards = computed(() => {
  const keyword = savedPresetCardKeyword.value.trim().toLowerCase()
  let list = savedExportRecordPresets.value
  if (keyword) {
    list = list.filter((item) => item.name.toLowerCase().includes(keyword))
  }
  const next = [...list]
  if (savedPresetCardSort.value === 'name_asc') {
    next.sort((a, b) => a.name.localeCompare(b.name, 'zh-Hans-CN'))
  } else {
    next.sort((a, b) => {
      const bTime = new Date(b.updatedAt).getTime()
      const aTime = new Date(a.updatedAt).getTime()
      return bTime - aTime
    })
  }
  return next
})
const selectedSavedExportRecordPresetIndex = computed(() =>
  savedExportRecordPresets.value.findIndex((item) => item.id === selectedSavedExportRecordPresetId.value),
)
const selectedSavedExportRecordPresetSummary = computed(() => {
  const id = selectedSavedExportRecordPresetId.value
  if (!id) return ''
  const preset = savedExportRecordPresets.value.find((item) => item.id === id)
  if (!preset) return ''
  const statusLabel = exportRecordStatusOptions.find((item) => item.value === preset.filters.status)?.label || preset.filters.status
  const modeLabel = exportRecordModeOptions.find((item) => item.value === preset.filters.mode)?.label || preset.filters.mode
  const timeLabel = exportRecordTimeRangeOptions.find((item) => item.value === preset.filters.timeRange)?.label || preset.filters.timeRange
  const rowsLabel = exportRecordRowsRangeOptions.find((item) => item.value === preset.filters.rowsRange)?.label || preset.filters.rowsRange
  const keyword = preset.filters.keyword.trim()
  return [
    `名称: ${preset.name}`,
    `状态: ${statusLabel}`,
    `范围: ${modeLabel}`,
    `时间: ${timeLabel}`,
    `条数: ${rowsLabel}`,
    keyword ? `关键词: ${keyword}` : '关键词: （空）',
  ].join(' | ')
})
const comparedSavedExportRecordPresets = computed(() =>
  savedExportRecordPresets.value.filter((item) => compareSavedExportRecordPresetIds.value.includes(item.id)).slice(0, 3),
)
const comparedSavedExportRecordPresetRows = computed(() => {
  const presets = comparedSavedExportRecordPresets.value
  if (presets.length < 2) return []
  return [
    { key: 'name', label: '名称', values: presets.map((item) => item.name) },
    {
      key: 'status',
      label: '状态',
      values: presets.map((item) => exportRecordStatusOptions.find((x) => x.value === item.filters.status)?.label || item.filters.status),
    },
    {
      key: 'mode',
      label: '范围',
      values: presets.map((item) => exportRecordModeOptions.find((x) => x.value === item.filters.mode)?.label || item.filters.mode),
    },
    {
      key: 'time',
      label: '时间',
      values: presets.map((item) => exportRecordTimeRangeOptions.find((x) => x.value === item.filters.timeRange)?.label || item.filters.timeRange),
    },
    {
      key: 'rows',
      label: '条数',
      values: presets.map((item) => exportRecordRowsRangeOptions.find((x) => x.value === item.filters.rowsRange)?.label || item.filters.rowsRange),
    },
    {
      key: 'keyword',
      label: '关键词',
      values: presets.map((item) => item.filters.keyword.trim() || '（空）'),
    },
  ]
})

const applyExportRecordPreset = (key: 'custom' | 'recent_success' | 'recent_failed' | 'all_mode_recent' | 'high_volume') => {
  if (key === 'custom') return
  if (key === 'recent_success') {
    exportRecordStatusFilter.value = 'success'
    exportRecordModeFilter.value = 'all'
    exportRecordSortOrder.value = 'desc'
    exportRecordTimeRangeFilter.value = '7d'
    exportRecordRowsRangeFilter.value = 'all'
    exportRecordKeyword.value = ''
    return
  }
  if (key === 'recent_failed') {
    exportRecordStatusFilter.value = 'failed'
    exportRecordModeFilter.value = 'all'
    exportRecordSortOrder.value = 'desc'
    exportRecordTimeRangeFilter.value = '30d'
    exportRecordRowsRangeFilter.value = 'zero'
    exportRecordKeyword.value = ''
    return
  }
  if (key === 'all_mode_recent') {
    exportRecordStatusFilter.value = 'all'
    exportRecordModeFilter.value = 'all_mode'
    exportRecordSortOrder.value = 'desc'
    exportRecordTimeRangeFilter.value = '7d'
    exportRecordRowsRangeFilter.value = 'all'
    exportRecordKeyword.value = ''
    return
  }
  exportRecordStatusFilter.value = 'all'
  exportRecordModeFilter.value = 'all'
  exportRecordSortOrder.value = 'desc'
  exportRecordTimeRangeFilter.value = '30d'
  exportRecordRowsRangeFilter.value = '1000_plus'
  exportRecordKeyword.value = ''
}

const inferExportRecordPresetKey = () => {
  const keyword = exportRecordKeyword.value.trim()
  if (
    keyword === ''
    && exportRecordStatusFilter.value === 'success'
    && exportRecordModeFilter.value === 'all'
    && exportRecordSortOrder.value === 'desc'
    && exportRecordTimeRangeFilter.value === '7d'
    && exportRecordRowsRangeFilter.value === 'all'
  ) {
    return 'recent_success' as const
  }
  if (
    keyword === ''
    && exportRecordStatusFilter.value === 'failed'
    && exportRecordModeFilter.value === 'all'
    && exportRecordSortOrder.value === 'desc'
    && exportRecordTimeRangeFilter.value === '30d'
    && exportRecordRowsRangeFilter.value === 'zero'
  ) {
    return 'recent_failed' as const
  }
  if (
    keyword === ''
    && exportRecordStatusFilter.value === 'all'
    && exportRecordModeFilter.value === 'all_mode'
    && exportRecordSortOrder.value === 'desc'
    && exportRecordTimeRangeFilter.value === '7d'
    && exportRecordRowsRangeFilter.value === 'all'
  ) {
    return 'all_mode_recent' as const
  }
  if (
    keyword === ''
    && exportRecordStatusFilter.value === 'all'
    && exportRecordModeFilter.value === 'all'
    && exportRecordSortOrder.value === 'desc'
    && exportRecordTimeRangeFilter.value === '30d'
    && exportRecordRowsRangeFilter.value === '1000_plus'
  ) {
    return 'high_volume' as const
  }
  return 'custom' as const
}

const onExportRecordPresetChange = (value: string) => {
  const key = value === 'recent_success' || value === 'recent_failed' || value === 'all_mode_recent' || value === 'high_volume'
    ? value
    : 'custom'
  exportRecordPresetKey.value = key
  applyExportRecordPreset(key)
}

const filteredExportRecords = computed(() => {
  const now = Date.now()
  const rangeMs = exportRecordTimeRangeFilter.value === '1d'
    ? 24 * 60 * 60 * 1000
    : exportRecordTimeRangeFilter.value === '7d'
      ? 7 * 24 * 60 * 60 * 1000
      : exportRecordTimeRangeFilter.value === '30d'
        ? 30 * 24 * 60 * 60 * 1000
        : 0
  const records = exportRecords.value.filter((item) => {
    const statusMatch = exportRecordStatusFilter.value === 'all' || item.status === exportRecordStatusFilter.value
    const modeMatch = exportRecordModeFilter.value === 'all'
      || (exportRecordModeFilter.value === 'current' && item.mode === 'current')
      || (exportRecordModeFilter.value === 'all_mode' && item.mode === 'all')
    const timeMatch = rangeMs <= 0 || (Number(item.ts || 0) > 0 && now - Number(item.ts || 0) <= rangeMs)
    const rows = Number(item.rows || 0)
    const rowsMatch = exportRecordRowsRangeFilter.value === 'all'
      || (exportRecordRowsRangeFilter.value === 'zero' && rows <= 0)
      || (exportRecordRowsRangeFilter.value === '1_100' && rows >= 1 && rows <= 100)
      || (exportRecordRowsRangeFilter.value === '101_1000' && rows >= 101 && rows <= 1000)
      || (exportRecordRowsRangeFilter.value === '1000_plus' && rows > 1000)
    const keyword = exportRecordKeyword.value.trim().toLowerCase()
    const statusText = item.status === 'success' ? '成功' : (item.status === 'cancelled' ? '已取消' : '失败')
    const modeText = item.mode === 'all' ? '全部匹配' : '当前页'
    const keywordTarget = `${item.time} ${statusText} ${modeText} ${item.summary}`.toLowerCase()
    const keywordMatch = !keyword || keywordTarget.includes(keyword)
    return statusMatch && modeMatch && timeMatch && rowsMatch && keywordMatch
  })
  records.sort((a, b) => {
    const at = Number(a.ts || 0)
    const bt = Number(b.ts || 0)
    return exportRecordSortOrder.value === 'asc' ? at - bt : bt - at
  })
  return records
})
const allFilteredExportRecordsSelected = computed(() => {
  const ids = filteredExportRecords.value.map((item) => item.id)
  if (!ids.length) return false
  return ids.every((id) => selectedExportRecordIds.value.includes(id))
})

const pushExportRecord = (record: {
  mode: 'current' | 'all';
  rows: number;
  truncated?: boolean;
  summary: string;
  status: 'success' | 'cancelled' | 'failed';
  filters?: {
    customerId: string;
    sourceTypes: string[];
    sourceId: string;
    createdFrom: string;
    createdTo: string;
    keyword: string;
  };
}) => {
  const recordId = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(16).slice(2)}`
  exportRecords.value.unshift({
    id: recordId,
    ts: Date.now(),
    time: formatDate(new Date().toISOString()),
    mode: record.mode,
    rows: record.rows,
    truncated: Boolean(record.truncated),
    summary: record.summary,
    status: record.status,
    filters: record.filters,
  })
  if (exportRecords.value.length > 10) {
    exportRecords.value = exportRecords.value.slice(0, 10)
  }
  return recordId
}

const highlightExportRecord = (id: string) => {
  highlightedExportRecordId.value = id
  if (highlightedExportRecordTimer) {
    clearTimeout(highlightedExportRecordTimer)
    highlightedExportRecordTimer = null
  }
  highlightedExportRecordTimer = setTimeout(() => {
    highlightedExportRecordId.value = ''
    highlightedExportRecordTimer = null
  }, 8000)
}

const focusHighlightedExportRecord = async (id: string) => {
  exportRecordStatusFilter.value = 'success'
  exportRecordSortOrder.value = 'desc'
  await nextTick()
  const target = document.querySelector(`[data-export-record-id="${id}"]`) as HTMLElement | null
  if (target) {
    target.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }
}

const clearExportRecords = () => {
  exportRecords.value = []
  selectedExportRecordIds.value = []
}

const removeExportRecord = (id: string) => {
  exportRecords.value = exportRecords.value.filter((item) => item.id !== id)
  selectedExportRecordIds.value = selectedExportRecordIds.value.filter((itemId) => itemId !== id)
}

const toggleExportRecordSelection = (id: string, checked: boolean) => {
  if (checked) {
    if (!selectedExportRecordIds.value.includes(id)) {
      selectedExportRecordIds.value = [...selectedExportRecordIds.value, id]
    }
    return
  }
  selectedExportRecordIds.value = selectedExportRecordIds.value.filter((itemId) => itemId !== id)
}

const toggleSelectAllFilteredExportRecords = () => {
  const ids = filteredExportRecords.value.map((item) => item.id)
  if (!ids.length) return
  if (allFilteredExportRecordsSelected.value) {
    selectedExportRecordIds.value = selectedExportRecordIds.value.filter((id) => !ids.includes(id))
    return
  }
  const merged = new Set([...selectedExportRecordIds.value, ...ids])
  selectedExportRecordIds.value = Array.from(merged)
}

const removeSelectedExportRecords = () => {
  const ids = new Set(selectedExportRecordIds.value)
  if (!ids.size) return
  const before = exportRecords.value.length
  exportRecords.value = exportRecords.value.filter((item) => !ids.has(item.id))
  selectedExportRecordIds.value = []
  const removed = before - exportRecords.value.length
  toast.add({
    title: '已批量删除导出记录',
    description: `共删除 ${removed} 条`,
    color: 'success',
  })
}

const keepOnlySuccessExportRecords = () => {
  const before = exportRecords.value.length
  exportRecords.value = exportRecords.value.filter((item) => item.status === 'success')
  selectedExportRecordIds.value = selectedExportRecordIds.value.filter((id) =>
    exportRecords.value.some((item) => item.id === id),
  )
  const removed = before - exportRecords.value.length
  toast.add({
    title: '已仅保留成功记录',
    description: `共清理 ${removed} 条`,
    color: 'success',
  })
}

const resetExportRecordFilters = () => {
  exportRecordStatusFilter.value = 'all'
  exportRecordModeFilter.value = 'all'
  exportRecordSortOrder.value = 'desc'
  exportRecordTimeRangeFilter.value = 'all'
  exportRecordRowsRangeFilter.value = 'all'
  exportRecordKeyword.value = ''
  exportRecordPresetKey.value = 'custom'
}

const captureExportRecordFilters = () => ({
  status: exportRecordStatusFilter.value,
  mode: exportRecordModeFilter.value,
  sort: exportRecordSortOrder.value,
  timeRange: exportRecordTimeRangeFilter.value,
  rowsRange: exportRecordRowsRangeFilter.value,
  keyword: exportRecordKeyword.value,
})

const applyCapturedExportRecordFilters = (filters: {
  status: 'all' | 'success' | 'cancelled' | 'failed';
  mode: 'all' | 'current' | 'all_mode';
  sort: 'desc' | 'asc';
  timeRange: 'all' | '1d' | '7d' | '30d';
  rowsRange: 'all' | 'zero' | '1_100' | '101_1000' | '1000_plus';
  keyword: string;
}) => {
  exportRecordStatusFilter.value = filters.status
  exportRecordModeFilter.value = filters.mode
  exportRecordSortOrder.value = filters.sort
  exportRecordTimeRangeFilter.value = filters.timeRange
  exportRecordRowsRangeFilter.value = filters.rowsRange
  exportRecordKeyword.value = filters.keyword
}

const saveCurrentExportRecordPreset = () => {
  if (!process.client) return
  const defaultName = `预设 ${formatDate(new Date().toISOString())}`
  const input = window.prompt('输入收藏名称', defaultName)
  const name = (input || '').trim()
  if (!name) return
  const nameKey = normalizeSavedExportRecordPresetName(name)
  const existingIndex = savedExportRecordPresets.value.findIndex((item) =>
    normalizeSavedExportRecordPresetName(item.name) === nameKey,
  )
  if (existingIndex >= 0) {
    const shouldOverwrite = window.confirm(`已存在同名收藏「${name}」，确定覆盖吗？`)
    if (!shouldOverwrite) return
    const existing = savedExportRecordPresets.value[existingIndex]
    savedExportRecordPresets.value[existingIndex] = {
      ...existing,
      name,
      updatedAt: new Date().toISOString(),
      filters: captureExportRecordFilters(),
    }
    savedExportRecordPresets.value = [...savedExportRecordPresets.value]
    selectedSavedExportRecordPresetId.value = existing.id
    toast.add({ title: '已覆盖同名个人预设', color: 'success' })
    return
  }
  const id = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(16).slice(2)}`
  savedExportRecordPresets.value.unshift({
    id,
    name,
    updatedAt: new Date().toISOString(),
    filters: captureExportRecordFilters(),
  })
  if (savedExportRecordPresets.value.length > 12) {
    savedExportRecordPresets.value = savedExportRecordPresets.value.slice(0, 12)
  }
  selectedSavedExportRecordPresetId.value = id
  toast.add({ title: '已保存个人预设', color: 'success' })
}

const applySavedExportRecordPreset = () => {
  const id = selectedSavedExportRecordPresetId.value
  if (!id) return
  const target = savedExportRecordPresets.value.find((item) => item.id === id)
  if (!target) return
  applyCapturedExportRecordFilters(target.filters)
  exportRecordPresetKey.value = inferExportRecordPresetKey()
  toast.add({ title: '已应用个人预设', description: target.name, color: 'success' })
}

const removeSavedExportRecordPreset = () => {
  const id = selectedSavedExportRecordPresetId.value
  if (!id) return
  const before = savedExportRecordPresets.value.length
  savedExportRecordPresets.value = savedExportRecordPresets.value.filter((item) => item.id !== id)
  if (savedExportRecordPresets.value.length === before) return
  selectedSavedExportRecordPresetId.value = ''
  toast.add({ title: '已删除个人预设', color: 'success' })
}

const exportSavedExportRecordPresets = () => {
  if (!savedExportRecordPresets.value.length) return
  const payload = {
    version: 1,
    exportedAt: new Date().toISOString(),
    presets: savedExportRecordPresets.value,
  }
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `points-mall-export-presets-${formatFileStamp(new Date())}.json`
  a.click()
  URL.revokeObjectURL(url)
  toast.add({ title: '已导出收藏预设', color: 'success' })
}

const openSavedExportRecordPresetImport = () => {
  savedExportRecordPresetImportRef.value?.click()
}

const normalizeSavedExportRecordPresetName = (name: string) => name.trim().toLowerCase()
const resolveSavedExportRecordPresetUpdatedAt = (value: unknown) => {
  const raw = typeof value === 'string' ? value.trim() : ''
  if (!raw) return new Date().toISOString()
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return new Date().toISOString()
  return date.toISOString()
}

const formatSavedExportRecordPresetSummary = (preset: {
  filters: {
    status: 'all' | 'success' | 'cancelled' | 'failed';
    mode: 'all' | 'current' | 'all_mode';
    timeRange: 'all' | '1d' | '7d' | '30d';
    rowsRange: 'all' | 'zero' | '1_100' | '101_1000' | '1000_plus';
    keyword: string;
  };
}) => {
  const statusLabel = exportRecordStatusOptions.find((item) => item.value === preset.filters.status)?.label || preset.filters.status
  const modeLabel = exportRecordModeOptions.find((item) => item.value === preset.filters.mode)?.label || preset.filters.mode
  const timeLabel = exportRecordTimeRangeOptions.find((item) => item.value === preset.filters.timeRange)?.label || preset.filters.timeRange
  const rowsLabel = exportRecordRowsRangeOptions.find((item) => item.value === preset.filters.rowsRange)?.label || preset.filters.rowsRange
  const keyword = preset.filters.keyword.trim()
  return [statusLabel, modeLabel, timeLabel, rowsLabel, keyword ? `关键词:${keyword}` : '关键词:（空）'].join(' / ')
}

const selectSavedExportRecordPreset = (id: string) => {
  selectedSavedExportRecordPresetId.value = id
}

const toggleCompareSavedExportRecordPreset = (id: string, checked: boolean) => {
  if (checked) {
    if (compareSavedExportRecordPresetIds.value.includes(id)) return
    compareSavedExportRecordPresetIds.value = [...compareSavedExportRecordPresetIds.value, id].slice(-3)
    return
  }
  compareSavedExportRecordPresetIds.value = compareSavedExportRecordPresetIds.value.filter((item) => item !== id)
}

const handleSavedExportRecordPresetImport = async (event: Event) => {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (!file) return
  try {
    const text = await file.text()
    const parsed = JSON.parse(text)
    const source = Array.isArray(parsed) ? parsed : (Array.isArray(parsed?.presets) ? parsed.presets : [])
    if (!Array.isArray(source)) {
      throw new Error('导入文件格式不正确')
    }
    const parsedImported = source
      .map((item: any) => ({
        id: typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
          ? crypto.randomUUID()
          : `${Date.now()}-${Math.random().toString(16).slice(2)}`,
        name: String(item?.name || '').trim(),
        updatedAt: new Date().toISOString(),
        filters: {
          status: item?.filters?.status === 'success' || item?.filters?.status === 'cancelled' || item?.filters?.status === 'failed'
            ? item.filters.status
            : 'all',
          mode: item?.filters?.mode === 'current' || item?.filters?.mode === 'all_mode' ? item.filters.mode : 'all',
          sort: item?.filters?.sort === 'asc' ? 'asc' : 'desc',
          timeRange: item?.filters?.timeRange === '1d' || item?.filters?.timeRange === '7d' || item?.filters?.timeRange === '30d'
            ? item.filters.timeRange
            : 'all',
          rowsRange: item?.filters?.rowsRange === 'zero' || item?.filters?.rowsRange === '1_100' || item?.filters?.rowsRange === '101_1000' || item?.filters?.rowsRange === '1000_plus'
            ? item.filters.rowsRange
            : 'all',
          keyword: String(item?.filters?.keyword || ''),
        },
      }))
      .filter((item: any) => item.name)
    if (!parsedImported.length) {
      throw new Error('未发现可导入的预设')
    }

    const dedupedImported: typeof parsedImported = []
    const importedNameSet = new Set<string>()
    let droppedInFile = 0
    parsedImported.forEach((item) => {
      const key = normalizeSavedExportRecordPresetName(item.name)
      if (!key) return
      if (importedNameSet.has(key)) {
        droppedInFile += 1
        return
      }
      importedNameSet.add(key)
      dedupedImported.push(item)
    })
    if (!dedupedImported.length) {
      throw new Error('导入文件中没有有效预设')
    }

    const existingByName = new Map<string, number>()
    savedExportRecordPresets.value.forEach((item, index) => {
      existingByName.set(normalizeSavedExportRecordPresetName(item.name), index)
    })
    const duplicatedWithExisting = dedupedImported.filter((item) =>
      existingByName.has(normalizeSavedExportRecordPresetName(item.name)),
    )
    const shouldOverwrite = duplicatedWithExisting.length > 0
      ? window.confirm(`检测到 ${duplicatedWithExisting.length} 个同名收藏，确定覆盖同名项吗？\n点击“取消”将跳过同名项。`)
      : false

    const next = [...savedExportRecordPresets.value]
    const newItems: typeof dedupedImported = []
    let overwritten = 0
    let skipped = 0
    let selectedId = ''
    dedupedImported.forEach((item) => {
      const key = normalizeSavedExportRecordPresetName(item.name)
      const existingIndex = existingByName.get(key)
      if (existingIndex == null) {
        newItems.push(item)
        if (!selectedId) selectedId = item.id
        return
      }
      if (!shouldOverwrite) {
        skipped += 1
        return
      }
      const current = next[existingIndex]
      next[existingIndex] = {
        ...current,
        name: item.name,
        updatedAt: new Date().toISOString(),
        filters: item.filters,
      }
      overwritten += 1
      if (!selectedId) selectedId = current.id
    })

    savedExportRecordPresets.value = [...newItems, ...next].slice(0, 12)
    selectedSavedExportRecordPresetId.value = selectedId
    toast.add({
      title: '已导入收藏预设',
      description: `新增 ${newItems.length} 条，覆盖 ${overwritten} 条，跳过 ${skipped + droppedInFile} 条`,
      color: 'success',
    })
  } catch (error: any) {
    toast.add({
      title: '导入收藏预设失败',
      description: error?.message || '请检查文件内容',
      color: 'error',
    })
  } finally {
    if (input) input.value = ''
  }
}

const renameSavedExportRecordPreset = () => {
  const id = selectedSavedExportRecordPresetId.value
  if (!id || !process.client) return
  const target = savedExportRecordPresets.value.find((item) => item.id === id)
  if (!target) return
  const input = window.prompt('输入新名称', target.name)
  const name = (input || '').trim()
  if (!name || name === target.name) return
  target.name = name
  target.updatedAt = new Date().toISOString()
  savedExportRecordPresets.value = [...savedExportRecordPresets.value]
  toast.add({ title: '已重命名个人预设', color: 'success' })
}

const cloneSavedExportRecordPreset = () => {
  const id = selectedSavedExportRecordPresetId.value
  if (!id || !process.client) return
  const target = savedExportRecordPresets.value.find((item) => item.id === id)
  if (!target) return
  const defaultName = `${target.name} - 副本`
  const input = window.prompt('输入克隆名称', defaultName)
  const name = (input || '').trim()
  if (!name) return
  const nameKey = normalizeSavedExportRecordPresetName(name)
  const exists = savedExportRecordPresets.value.some((item) =>
    normalizeSavedExportRecordPresetName(item.name) === nameKey,
  )
  if (exists) {
    const shouldOverwrite = window.confirm(`已存在同名收藏「${name}」，确定覆盖吗？`)
    if (shouldOverwrite) {
      const existingIndex = savedExportRecordPresets.value.findIndex((item) =>
        normalizeSavedExportRecordPresetName(item.name) === nameKey,
      )
      if (existingIndex >= 0) {
        const existing = savedExportRecordPresets.value[existingIndex]
        savedExportRecordPresets.value[existingIndex] = {
          ...existing,
          name,
          updatedAt: new Date().toISOString(),
          filters: { ...target.filters },
        }
        savedExportRecordPresets.value = [...savedExportRecordPresets.value]
        selectedSavedExportRecordPresetId.value = existing.id
        toast.add({ title: '已覆盖同名收藏', color: 'success' })
      }
    }
    return
  }
  const newId = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(16).slice(2)}`
  savedExportRecordPresets.value.unshift({
    id: newId,
    name,
    updatedAt: new Date().toISOString(),
    filters: { ...target.filters },
  })
  if (savedExportRecordPresets.value.length > 12) {
    savedExportRecordPresets.value = savedExportRecordPresets.value.slice(0, 12)
  }
  selectedSavedExportRecordPresetId.value = newId
  toast.add({ title: '已克隆个人预设', color: 'success' })
}

const moveSavedExportRecordPreset = (direction: -1 | 1) => {
  const fromIndex = selectedSavedExportRecordPresetIndex.value
  if (fromIndex < 0) return
  const toIndex = fromIndex + direction
  if (toIndex < 0 || toIndex >= savedExportRecordPresets.value.length) return
  const next = [...savedExportRecordPresets.value]
  const [moved] = next.splice(fromIndex, 1)
  next.splice(toIndex, 0, moved)
  savedExportRecordPresets.value = next
}

const captureCurrentExportFilters = () => ({
  customerId: txCustomerId.value,
  sourceTypes: [...txSourceTypes.value],
  sourceId: txSourceId.value,
  createdFrom: txCreatedFrom.value,
  createdTo: txCreatedTo.value,
  keyword: keyword.value,
})

const buildTransactionQueryFromFilters = (
  filters: {
    customerId: string;
    sourceTypes: string[];
    sourceId: string;
    createdFrom: string;
    createdTo: string;
    keyword: string;
  },
  page = 1,
  pageSize = txPageSize.value,
) => ({
  tokenCode: 'points',
  customerId: filters.customerId === 'all' ? undefined : filters.customerId,
  sourceType: filters.sourceTypes.length ? filters.sourceTypes.join(',') : undefined,
  sourceId: filters.sourceId.trim() || undefined,
  createdFrom: toRfc3339(filters.createdFrom),
  createdTo: toRfc3339(filters.createdTo, true),
  page,
  pageSize,
})

const buildFilterSummary = (filters: {
  customerId: string;
  sourceTypes: string[];
  sourceId: string;
  createdFrom: string;
  createdTo: string;
  keyword: string;
}) => {
  const chips: string[] = []
  if (filters.customerId !== 'all') {
    const customerLabel = txCustomerOptions.value.find((item) => item.value === filters.customerId)?.label || filters.customerId
    chips.push(`客户: ${customerLabel}`)
  }
  filters.sourceTypes.forEach((sourceType) => {
    chips.push(`来源: ${txSourceTypeMap.value[sourceType] || sourceType}`)
  })
  if (filters.sourceId.trim()) chips.push(`来源ID: ${filters.sourceId.trim()}`)
  if (filters.createdFrom.trim()) chips.push(`开始: ${filters.createdFrom.trim()}`)
  if (filters.createdTo.trim()) chips.push(`结束: ${filters.createdTo.trim()}`)
  return chips.length ? `当前筛选：${chips.join('，')}` : '当前为全部数据'
}

const hydrateExportRecords = () => {
  if (!process.client) return
  try {
    const raw = localStorage.getItem(EXPORT_RECORDS_STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return
    const records = parsed
      .map((item: any) => ({
        id: String(item?.id || ''),
        ts: Number(item?.ts || 0),
        time: String(item?.time || ''),
        mode: item?.mode === 'all' ? 'all' : 'current',
        rows: Number(item?.rows || 0),
        truncated: Boolean(item?.truncated),
        summary: String(item?.summary || '当前为全部数据'),
        status: item?.status === 'success' || item?.status === 'cancelled' || item?.status === 'failed'
          ? item.status
          : 'success',
        filters: item?.filters && typeof item.filters === 'object'
          ? {
              customerId: String(item.filters.customerId || 'all'),
              sourceTypes: Array.isArray(item.filters.sourceTypes)
                ? item.filters.sourceTypes.map((x: any) => String(x)).filter((x: string) => x.trim() !== '')
                : [],
              sourceId: String(item.filters.sourceId || ''),
              createdFrom: String(item.filters.createdFrom || ''),
              createdTo: String(item.filters.createdTo || ''),
              keyword: String(item.filters.keyword || ''),
            }
          : undefined,
      }))
      .filter((item: any) => item.id && item.time)
      .slice(0, 10)
    exportRecords.value = records
  } catch {}
}

const persistExportRecords = () => {
  if (!process.client) return
  try {
    localStorage.setItem(EXPORT_RECORDS_STORAGE_KEY, JSON.stringify(exportRecords.value.slice(0, 10)))
  } catch {}
}

const hydrateSavedExportRecordPresets = () => {
  if (!process.client) return
  try {
    const raw = localStorage.getItem(EXPORT_RECORD_SAVED_PRESETS_STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return
    const presets = parsed
      .map((item: any) => ({
        id: String(item?.id || ''),
        name: String(item?.name || '').trim(),
        updatedAt: resolveSavedExportRecordPresetUpdatedAt(item?.updatedAt),
        filters: {
          status: item?.filters?.status === 'success' || item?.filters?.status === 'cancelled' || item?.filters?.status === 'failed'
            ? item.filters.status
            : 'all',
          mode: item?.filters?.mode === 'current' || item?.filters?.mode === 'all_mode' ? item.filters.mode : 'all',
          sort: item?.filters?.sort === 'asc' ? 'asc' : 'desc',
          timeRange: item?.filters?.timeRange === '1d' || item?.filters?.timeRange === '7d' || item?.filters?.timeRange === '30d'
            ? item.filters.timeRange
            : 'all',
          rowsRange: item?.filters?.rowsRange === 'zero' || item?.filters?.rowsRange === '1_100' || item?.filters?.rowsRange === '101_1000' || item?.filters?.rowsRange === '1000_plus'
            ? item.filters.rowsRange
            : 'all',
          keyword: String(item?.filters?.keyword || ''),
        },
      }))
      .filter((item: any) => item.id && item.name)
      .slice(0, 12)
    savedExportRecordPresets.value = presets
  } catch {}
}

const persistSavedExportRecordPresets = () => {
  if (!process.client) return
  try {
    localStorage.setItem(
      EXPORT_RECORD_SAVED_PRESETS_STORAGE_KEY,
      JSON.stringify(savedExportRecordPresets.value.slice(0, 12)),
    )
  } catch {}
}

const copyExportRecordSummary = async (item: {
  id: string;
  time: string;
  mode: 'current' | 'all';
  rows: number;
  truncated: boolean;
  summary: string;
  status: 'success' | 'cancelled' | 'failed';
}) => {
  const text = [
    `时间: ${item.time}`,
    `状态: ${item.status === 'success' ? '成功' : (item.status === 'cancelled' ? '已取消' : '失败')}`,
    `范围: ${item.mode === 'all' ? '全部匹配' : '当前页'}`,
    `条数: ${item.rows}`,
    item.truncated ? '截断: 是' : '截断: 否',
    `筛选: ${item.summary}`,
  ].join(' | ')
  try {
    await navigator.clipboard.writeText(text)
    copiedExportRecordId.value = item.id
    if (copiedExportRecordTimer) {
      clearTimeout(copiedExportRecordTimer)
      copiedExportRecordTimer = null
    }
    copiedExportRecordTimer = setTimeout(() => {
      copiedExportRecordId.value = ''
      copiedExportRecordTimer = null
    }, 2000)
    toast.add({ title: '已复制导出摘要', color: 'success' })
  } catch {
    toast.add({ title: '复制失败', description: '请手动复制', color: 'warning' })
  }
}

const reuseExportRecordFilters = async (item: {
  filters?: {
    customerId: string;
    sourceTypes: string[];
    sourceId: string;
    createdFrom: string;
    createdTo: string;
    keyword: string;
  };
}) => {
  if (!item.filters) return
  await withTxFilterBatch(() => {
    txCustomerId.value = item.filters?.customerId || 'all'
    txSourceTypes.value = Array.isArray(item.filters?.sourceTypes) ? [...item.filters.sourceTypes] : []
    txSourceId.value = item.filters?.sourceId || ''
    txCreatedFrom.value = item.filters?.createdFrom || ''
    txCreatedTo.value = item.filters?.createdTo || ''
    keyword.value = item.filters?.keyword || ''
    txPage.value = 1
  })
  await syncAllQuery()
  await loadTransactions()
  toast.add({ title: '已复用筛选', color: 'success' })
}

const applyExportFiltersSnapshot = async (filters: {
  customerId: string;
  sourceTypes: string[];
  sourceId: string;
  createdFrom: string;
  createdTo: string;
  keyword: string;
}) => {
  await withTxFilterBatch(() => {
    txCustomerId.value = filters.customerId || 'all'
    txSourceTypes.value = Array.isArray(filters.sourceTypes) ? [...filters.sourceTypes] : []
    txSourceId.value = filters.sourceId || ''
    txCreatedFrom.value = filters.createdFrom || ''
    txCreatedTo.value = filters.createdTo || ''
    keyword.value = filters.keyword || ''
    txPage.value = 1
  })
  await syncAllQuery()
  await loadTransactions()
}

const reExportFromRecord = async (item: {
  filters?: {
    customerId: string;
    sourceTypes: string[];
    sourceId: string;
    createdFrom: string;
    createdTo: string;
    keyword: string;
  };
}, mode: 'current' | 'all') => {
  if (!item.filters) {
    toast.add({
      title: '无法重导',
      description: '该记录来自旧版本，缺少筛选快照',
      color: 'warning',
    })
    return
  }
  await applyExportFiltersSnapshot(item.filters)
  shouldHighlightNextExportRecord = true
  if (mode === 'all') {
    await openExportAllConfirm()
    return
  }
  await exportTransactionsCsv(false)
}

const downloadLatestFromRecord = async (item: {
  id: string;
  mode: 'current' | 'all';
  filters?: {
    customerId: string;
    sourceTypes: string[];
    sourceId: string;
    createdFrom: string;
    createdTo: string;
    keyword: string;
  };
}) => {
  if (!item.filters) {
    toast.add({
      title: '无法下载',
      description: '该记录来自旧版本，缺少筛选快照',
      color: 'warning',
    })
    return
  }
  if (downloadingRecordId.value || exportingAllTransactions.value) return

  downloadingRecordId.value = item.id
  try {
    const allMatched = item.mode === 'all'
    if (allMatched) {
      exportingAllTransactions.value = true
      exportAllFetched.value = 0
      exportAllTotalEstimate.value = 0
      exportAbortController = new AbortController()
    }

    const summary = buildFilterSummary(item.filters)
    const allResult = allMatched
      ? await listAllMatchedTransactionsForExport(selectedExportLimit.value, item.filters)
      : null
    const rows = allMatched
      ? (allResult?.items || [])
      : ((await membershipApi.listTokenTransactions(buildTransactionQueryFromFilters(item.filters, 1, txPageSize.value)))?.items || [])

    const lines: string[] = [
      `# ${summary}`,
      'id,customer_id,customer_name,token_code,delta,source_type,source_id,created_at',
    ]
    rows.forEach((tx) => {
      lines.push([
        escapeCsv(tx.id),
        escapeCsv(tx.customerId),
        escapeCsv(resolveCustomerName(tx.customerId)),
        escapeCsv(tx.tokenCode),
        escapeCsv(tx.delta),
        escapeCsv(tx.sourceType),
        escapeCsv(tx.sourceId),
        escapeCsv(tx.createdAt),
      ].join(','))
    })

    const blob = new Blob(["\uFEFF" + lines.join('\n')], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const parts = ['points-mall-transactions', 'record-latest', formatFileStamp(new Date())]
    if (allMatched) parts.push('all')
    if (item.filters.customerId !== 'all') parts.push(`customer-${sanitizeFileToken(item.filters.customerId)}`)
    if (item.filters.sourceTypes.length) parts.push(`source-${sanitizeFileToken(item.filters.sourceTypes.join('-'))}`)
    if (item.filters.sourceId.trim()) parts.push(`sid-${sanitizeFileToken(item.filters.sourceId.trim()).slice(0, 24)}`)
    a.href = url
    a.download = `${parts.join('_')}.csv`
    a.click()
    URL.revokeObjectURL(url)

    toast.add({
      title: '已下载同条件最新CSV',
      description: `共 ${rows.length} 条`,
      color: 'success',
    })
    const newRecordId = pushExportRecord({
      mode: allMatched ? 'all' : 'current',
      rows: rows.length,
      truncated: Boolean(allResult?.truncated),
      summary,
      status: 'success',
      filters: { ...item.filters, sourceTypes: [...item.filters.sourceTypes] },
    })
    highlightExportRecord(newRecordId)
    await focusHighlightedExportRecord(newRecordId)
    if (allMatched && allResult?.truncated) {
      toast.add({
        title: '导出已截断',
        description: `匹配总量 ${allResult.total} 条，已按上限导出前 ${selectedExportLimit.value} 条`,
        color: 'warning',
      })
    }
  } catch (error: any) {
    const allMatched = item.mode === 'all'
    if (allMatched && (error?.name === 'AbortError' || String(error?.message || '').toLowerCase().includes('aborted'))) {
      toast.add({ title: '已取消下载', color: 'info' })
      pushExportRecord({
        mode: 'all',
        rows: exportAllFetched.value,
        summary: buildFilterSummary(item.filters),
        status: 'cancelled',
        filters: { ...item.filters, sourceTypes: [...item.filters.sourceTypes] },
      })
      return
    }
    toast.add({
      title: '下载最新CSV失败',
      description: error?.message || '请稍后重试',
      color: 'error',
    })
    pushExportRecord({
      mode: item.mode === 'all' ? 'all' : 'current',
      rows: 0,
      summary: buildFilterSummary(item.filters),
      status: 'failed',
      filters: { ...item.filters, sourceTypes: [...item.filters.sourceTypes] },
    })
  } finally {
    downloadingRecordId.value = ''
    if (item.mode === 'all') {
      exportingAllTransactions.value = false
      exportAllFetched.value = 0
      exportAllTotalEstimate.value = 0
      exportAbortController = null
    }
  }
}

const buildTransactionQuery = (page = txPage.value, pageSize = txPageSize.value) => ({
  tokenCode: 'points',
  customerId: txCustomerId.value === 'all' ? undefined : txCustomerId.value,
  sourceType: txSourceTypes.value.length ? txSourceTypes.value.join(',') : undefined,
  sourceId: txSourceId.value.trim() || undefined,
  createdFrom: toRfc3339(txCreatedFrom.value),
  createdTo: toRfc3339(txCreatedTo.value, true),
  page,
  pageSize,
})

const listAllMatchedTransactionsForExport = async (
  exportLimit: number,
  filters?: {
    customerId: string;
    sourceTypes: string[];
    sourceId: string;
    createdFrom: string;
    createdTo: string;
    keyword: string;
  },
) => {
  const pageSize = 200
  const maxPages = 1000
  let page = 1
  let total = 0
  const allItems: MembershipTokenTransaction[] = []
  let truncated = false

  while (page <= maxPages) {
    const query = filters
      ? buildTransactionQueryFromFilters(filters, page, pageSize)
      : buildTransactionQuery(page, pageSize)
    const resp = await membershipApi.listTokenTransactions(
      query,
      exportAbortController ? { signal: exportAbortController.signal } : undefined,
    )
    const items = resp?.items || []
    total = Number(resp?.total || 0)
    exportAllTotalEstimate.value = total
    allItems.push(...items)
    exportAllFetched.value = allItems.length
    if (exportLimit > 0 && allItems.length >= exportLimit) {
      allItems.splice(exportLimit)
      truncated = total > exportLimit
      break
    }
    if (!items.length || allItems.length >= total) break
    page += 1
  }

  return {
    items: allItems,
    truncated,
    total,
  }
}

const exportTransactionsCsv = async (allMatched = false) => {
  try {
    if (allMatched) {
      exportingAllTransactions.value = true
      exportAllFetched.value = 0
      exportAllTotalEstimate.value = 0
      exportAbortController = new AbortController()
    }
    const summary = txResultSummary.value || '当前为全部数据'
    const allResult = allMatched ? await listAllMatchedTransactionsForExport(selectedExportLimit.value) : null
    const rows = allMatched ? (allResult?.items || []) : transactions.value
    const lines: string[] = [
      `# ${summary}`,
      'id,customer_id,customer_name,token_code,delta,source_type,source_id,created_at',
    ]
    rows.forEach((item) => {
      lines.push([
        escapeCsv(item.id),
        escapeCsv(item.customerId),
        escapeCsv(resolveCustomerName(item.customerId)),
        escapeCsv(item.tokenCode),
        escapeCsv(item.delta),
        escapeCsv(item.sourceType),
        escapeCsv(item.sourceId),
        escapeCsv(item.createdAt),
      ].join(','))
    })
    const blob = new Blob(["\uFEFF" + lines.join('\n')], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const parts = ['points-mall-transactions', formatFileStamp(new Date())]
    if (allMatched) parts.push('all')
    if (txCustomerId.value !== 'all') parts.push(`customer-${sanitizeFileToken(txCustomerId.value)}`)
    if (txSourceTypes.value.length) parts.push(`source-${sanitizeFileToken(txSourceTypes.value.join('-'))}`)
    if (txSourceId.value.trim()) parts.push(`sid-${sanitizeFileToken(txSourceId.value.trim()).slice(0, 24)}`)
    a.href = url
    a.download = `${parts.join('_')}.csv`
    a.click()
    URL.revokeObjectURL(url)
    toast.add({
      title: allMatched ? '已导出全部匹配交易' : '已导出当前页交易',
      description: `共 ${rows.length} 条`,
      color: 'success',
    })
    const recordId = pushExportRecord({
      mode: allMatched ? 'all' : 'current',
      rows: rows.length,
      truncated: Boolean(allResult?.truncated),
      summary,
      status: 'success',
      filters: captureCurrentExportFilters(),
    })
    if (shouldHighlightNextExportRecord) {
      highlightExportRecord(recordId)
      await focusHighlightedExportRecord(recordId)
      shouldHighlightNextExportRecord = false
    }
    if (allMatched && allResult?.truncated) {
      toast.add({
        title: '导出已截断',
        description: `匹配总量 ${allResult.total} 条，已按上限导出前 ${selectedExportLimit.value} 条`,
        color: 'warning',
      })
    }
  } catch (error: any) {
    if (allMatched && (error?.name === 'AbortError' || String(error?.message || '').toLowerCase().includes('aborted'))) {
      toast.add({
        title: '已取消导出',
        color: 'info',
      })
      pushExportRecord({
        mode: 'all',
        rows: exportAllFetched.value,
        summary: txResultSummary.value || '当前为全部数据',
        status: 'cancelled',
        filters: captureCurrentExportFilters(),
      })
      shouldHighlightNextExportRecord = false
      return
    }
    toast.add({
      title: allMatched ? '导出全部匹配交易失败' : '导出交易失败',
      description: error?.message || '请稍后重试',
      color: 'error',
    })
    pushExportRecord({
      mode: allMatched ? 'all' : 'current',
      rows: 0,
      summary: txResultSummary.value || '当前为全部数据',
      status: 'failed',
      filters: captureCurrentExportFilters(),
    })
    shouldHighlightNextExportRecord = false
  } finally {
    if (allMatched) {
      exportingAllTransactions.value = false
      exportAllFetched.value = 0
      exportAllTotalEstimate.value = 0
      exportAbortController = null
    }
  }
}

const cancelExportAll = () => {
  if (!exportingAllTransactions.value) return
  exportAbortController?.abort()
}

const estimateExportAllTotal = async () => {
  exportAllEstimateLoading.value = true
  try {
    const resp = await membershipApi.listTokenTransactions(buildTransactionQuery(1, 1))
    exportAllEstimateTotal.value = Number(resp?.total || 0)
  } catch (error: any) {
    exportAllEstimateTotal.value = Number(transactionsTotal.value || 0)
    toast.add({
      title: '估算条数失败',
      description: error?.message || '已使用当前列表统计值',
      color: 'warning',
    })
  } finally {
    exportAllEstimateLoading.value = false
  }
}

const openExportAllConfirm = async () => {
  exportAllConfirmOpen.value = true
  await estimateExportAllTotal()
}

const cancelExportAllConfirm = () => {
  exportAllConfirmOpen.value = false
  shouldHighlightNextExportRecord = false
}

const confirmExportAll = async () => {
  exportAllConfirmOpen.value = false
  await exportTransactionsCsv(true)
}

const openRedeemModal = (benefit: MembershipBenefit) => {
  selectedBenefit.value = benefit
  redeemForm.value = {
    customerId: memberOptions.value[0]?.value || '',
    pointsCost: 1,
    reason: '',
    sourceId: typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID()
      : `${Date.now()}-${Math.random().toString(16).slice(2)}`,
  }
  redeemModalOpen.value = true
}

const submitRedeem = async () => {
  if (!selectedBenefit.value || !canSubmitRedeem.value) return
  try {
    redeemSubmitting.value = true
    const result = await membershipApi.redeemPointsBenefit({
      customerId: redeemForm.value.customerId,
      benefitId: selectedBenefit.value.id,
      pointsCost: Number(redeemForm.value.pointsCost),
      reason: redeemForm.value.reason.trim() || undefined,
      sourceId: redeemForm.value.sourceId,
    }) as RedeemPointsBenefitResult
    highlightedTxId.value = result?.transactionId || ''
    highlightedTxPinned.value = false
    scheduleClearHighlightedTx()
    redeemModalOpen.value = false
    await applyRedeemFocusedFilters(redeemForm.value.customerId)
    toast.add({
      title: '兑换成功',
      description: `已自动定位交易记录，当前余额 ${Number(result?.balance ?? 0)}`,
      color: 'success',
    })
    await loadData()
    await scrollToHighlightedTransaction()
  } catch (error: any) {
    toast.add({ title: '兑换失败', description: error?.message || '请稍后重试', color: 'error' })
  } finally {
    redeemSubmitting.value = false
  }
}

const loadData = async () => {
  try {
    loading.value = true
    const [memberResp, benefitResp] = await Promise.all([
      customerApi.listMembers({ page: 1, pageSize: 1000 }),
      membershipApi.listBenefits(),
    ])
    members.value = memberResp?.data || []
    benefitItems.value = benefitResp?.items || []
    await loadTransactions()
  } catch (error: any) {
    toast.add({ title: '加载积分商城数据失败', description: error?.message || '请稍后重试', color: 'error' })
  } finally {
    loading.value = false
  }
}

const loadTransactions = async () => {
  transactionsLoading.value = true
  try {
    const resp = await membershipApi.listTokenTransactions(buildTransactionQuery())
    transactions.value = resp?.items || []
    transactionsTotal.value = Number(resp?.total || 0)
  } catch (error: any) {
    toast.add({ title: '加载交易记录失败', description: error?.message || '请稍后重试', color: 'error' })
  } finally {
    transactionsLoading.value = false
  }
}

const reloadTransactions = async () => {
  if (txDateRangeError.value) return
  await syncAllQuery()
  await loadTransactions()
}

const applyTxFilters = async (showToast = false) => {
  if (txDateRangeError.value) {
    if (showToast) {
      toast.add({ title: '时间范围无效', description: txDateRangeError.value, color: 'warning' })
    }
    return
  }
  txPage.value = 1
  await syncAllQuery()
  await loadTransactions()
}

const toLocalInput = (value?: string | null) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const MANAGED_QUERY_KEYS = [
  'txCustomer',
  'txSource',
  'txSourceId',
  'txFrom',
  'txTo',
  'txPage',
  'keyword',
  'erPreset',
  'erStatus',
  'erMode',
  'erSort',
  'erTime',
  'erRows',
  'erKeyword',
  'erCards',
  'erCardKeyword',
  'erCardSort',
  'erSavedPreset',
  'erSavedPresetName',
] as const
const MANAGED_QUERY_KEY_SET = createManagedQueryKeySet(MANAGED_QUERY_KEYS)

const hydrateTxFiltersFromQuery = () => {
  warnUnknownManagedQueryKeys({
    scope: 'points.mall',
    rawQuery: route.query as Record<string, unknown>,
    managedKeys: MANAGED_QUERY_KEYS,
    managedKeySet: MANAGED_QUERY_KEY_SET,
    emitConsole: false,
    debugStats: { page: 'points.mall', logEvery: 10 },
  })
  const managed = collectManagedQueryValues({
    rawQuery: route.query as Record<string, unknown>,
    managedKeys: MANAGED_QUERY_KEYS,
  })
  const keywordQuery = managed.keyword || ''
  const customer = managed.txCustomer || ''
  const source = managed.txSource || ''
  const sourceId = managed.txSourceId || ''
  const from = managed.txFrom || ''
  const to = managed.txTo || ''
  const page = Number(managed.txPage || '')
  const erStatus = managed.erStatus || ''
  const erMode = managed.erMode || ''
  const erSort = managed.erSort || ''
  const erTime = managed.erTime || ''
  const erRows = managed.erRows || ''
  const erKeyword = managed.erKeyword || ''
  const erPreset = managed.erPreset || ''
  const erCardKeyword = managed.erCardKeyword || ''
  const erCardSort = managed.erCardSort || ''
  const erCards = managed.erCards || ''
  const erSavedPreset = managed.erSavedPreset || ''
  const erSavedPresetName = managed.erSavedPresetName || ''

  if (keywordQuery) keyword.value = keywordQuery
  if (customer) txCustomerId.value = customer
  if (source) {
    txSourceTypes.value = source.split(',').map((item) => item.trim()).filter((item) => item)
  }
  if (sourceId) txSourceId.value = sourceId
  if (from) txCreatedFrom.value = toLocalInput(from)
  if (to) txCreatedTo.value = toLocalInput(to)
  if (Number.isFinite(page) && page > 0) txPage.value = Math.floor(page)
  if (erStatus === 'all' || erStatus === 'success' || erStatus === 'cancelled' || erStatus === 'failed') {
    exportRecordStatusFilter.value = erStatus
  }
  if (erMode === 'all' || erMode === 'current' || erMode === 'all_mode') {
    exportRecordModeFilter.value = erMode
  }
  if (erSort === 'desc' || erSort === 'asc') {
    exportRecordSortOrder.value = erSort
  }
  if (erTime === 'all' || erTime === '1d' || erTime === '7d' || erTime === '30d') {
    exportRecordTimeRangeFilter.value = erTime
  }
  if (erRows === 'all' || erRows === 'zero' || erRows === '1_100' || erRows === '101_1000' || erRows === '1000_plus') {
    exportRecordRowsRangeFilter.value = erRows
  }
  if (erKeyword) exportRecordKeyword.value = erKeyword
  if (erPreset === 'recent_success' || erPreset === 'recent_failed' || erPreset === 'all_mode_recent' || erPreset === 'high_volume') {
    exportRecordPresetKey.value = erPreset
    applyExportRecordPreset(erPreset)
  }
  if (erCardKeyword) savedPresetCardKeyword.value = erCardKeyword
  if (erCardSort === 'updated_desc' || erCardSort === 'name_asc') {
    savedPresetCardSort.value = erCardSort
  }
  if (erCards === '1') showSavedExportRecordPresetCards.value = true
  if (erSavedPreset && savedExportRecordPresets.value.some((item) => item.id === erSavedPreset)) {
    selectedSavedExportRecordPresetId.value = erSavedPreset
  } else if (erSavedPresetName) {
    const nameKey = normalizeSavedExportRecordPresetName(erSavedPresetName)
    const matched = savedExportRecordPresets.value.find((item) => normalizeSavedExportRecordPresetName(item.name) === nameKey)
    if (matched) selectedSavedExportRecordPresetId.value = matched.id
  }
  exportRecordPresetKey.value = inferExportRecordPresetKey()
}

const buildQueryFromState = () => {
  const currentQuery: Record<string, string> = {}
  for (const [key, value] of Object.entries(route.query)) {
    if (value != null) {
      currentQuery[key] = Array.isArray(value) ? String(value[0]) : String(value)
    }
  }
  const preservedQueryEntries = Object.entries(currentQuery)
    .filter(([key]) => !MANAGED_QUERY_KEY_SET.has(key))
    .sort(([a], [b]) => a.localeCompare(b))
  const managedQuery: Record<string, string> = {}

  if (txCustomerId.value !== 'all') managedQuery.txCustomer = txCustomerId.value
  if (txSourceTypes.value.length) managedQuery.txSource = txSourceTypes.value.join(',')
  if (txSourceId.value.trim()) managedQuery.txSourceId = txSourceId.value.trim()

  const from = toRfc3339(txCreatedFrom.value)
  const to = toRfc3339(txCreatedTo.value, true)
  if (from) managedQuery.txFrom = from
  if (to) managedQuery.txTo = to

  if (txPage.value > 1) managedQuery.txPage = String(txPage.value)

  const kw = keyword.value.trim()
  if (kw) managedQuery.keyword = kw

  const erKeyword = exportRecordKeyword.value.trim()
  if (exportRecordPresetKey.value !== 'custom') {
    managedQuery.erPreset = exportRecordPresetKey.value
    if (erKeyword) managedQuery.erKeyword = erKeyword
  } else {
    if (exportRecordStatusFilter.value !== 'all') managedQuery.erStatus = exportRecordStatusFilter.value
    if (exportRecordModeFilter.value !== 'all') managedQuery.erMode = exportRecordModeFilter.value
    if (exportRecordSortOrder.value !== 'desc') managedQuery.erSort = exportRecordSortOrder.value
    if (exportRecordTimeRangeFilter.value !== 'all') managedQuery.erTime = exportRecordTimeRangeFilter.value
    if (exportRecordRowsRangeFilter.value !== 'all') managedQuery.erRows = exportRecordRowsRangeFilter.value
    if (erKeyword) managedQuery.erKeyword = erKeyword
  }
  if (showSavedExportRecordPresetCards.value) {
    managedQuery.erCards = '1'
    const erCardKeyword = savedPresetCardKeyword.value.trim()
    if (erCardKeyword) managedQuery.erCardKeyword = erCardKeyword
    if (savedPresetCardSort.value !== 'updated_desc') managedQuery.erCardSort = savedPresetCardSort.value
  }
  if (selectedSavedExportRecordPresetId.value) {
    managedQuery.erSavedPreset = selectedSavedExportRecordPresetId.value
    const selectedPreset = savedExportRecordPresets.value.find((item) => item.id === selectedSavedExportRecordPresetId.value)
    if (selectedPreset?.name.trim()) managedQuery.erSavedPresetName = selectedPreset.name.trim().slice(0, 64)
  }
  const nextQuery: Record<string, string> = {}
  preservedQueryEntries.forEach(([key, value]) => {
    nextQuery[key] = value
  })
  MANAGED_QUERY_KEYS.forEach((key) => {
    const value = managedQuery[key]
    if (value != null) nextQuery[key] = value
  })
  return nextQuery
}

const { queueManagedQuerySync } = useManagedQuerySync({
  route,
  router,
  buildNextQuery: buildQueryFromState,
  isReady: () => querySyncReady.value,
  debugStats: { page: 'points.mall', logEvery: 10 },
})

const syncAllQuery = async () => {
  await queueManagedQuerySync()
}

const resetAllFilters = async () => {
  if (highlightedTxTimer) {
    clearTimeout(highlightedTxTimer)
    highlightedTxTimer = null
  }
  highlightedTxId.value = ''
  highlightedTxPinned.value = false
  await withTxFilterBatch(() => {
    keyword.value = ''
    txCustomerId.value = 'all'
    txSourceTypes.value = []
    txSourceId.value = ''
    txCreatedFrom.value = ''
    txCreatedTo.value = ''
    txPage.value = 1
  })
  await syncAllQuery()
  await loadTransactions()
}

let txFilterTimer: ReturnType<typeof setTimeout> | null = null
let exportRecordFilterTimer: ReturnType<typeof setTimeout> | null = null
const scheduleAutoTransactionQuery = () => {
  if (!txFiltersInitialized.value) return
  if (txFilterTimer) clearTimeout(txFilterTimer)
  txFilterTimer = setTimeout(() => {
    void applyTxFilters(false)
  }, 250)
}

const scheduleExportRecordQuerySync = () => {
  if (!txFiltersInitialized.value) return
  if (exportRecordFilterTimer) clearTimeout(exportRecordFilterTimer)
  exportRecordFilterTimer = setTimeout(() => {
    void syncAllQuery()
  }, 250)
}

watch([txCustomerId, txSourceTypes, txSourceId, txCreatedFrom, txCreatedTo], scheduleAutoTransactionQuery)
watch(
  [exportRecordStatusFilter, exportRecordModeFilter, exportRecordSortOrder, exportRecordTimeRangeFilter, exportRecordRowsRangeFilter, exportRecordKeyword],
  () => {
    const inferred = inferExportRecordPresetKey()
    if (exportRecordPresetKey.value !== inferred) {
      exportRecordPresetKey.value = inferred
    }
  },
)
watch(
  [exportRecordStatusFilter, exportRecordModeFilter, exportRecordSortOrder, exportRecordTimeRangeFilter, exportRecordRowsRangeFilter, exportRecordKeyword],
  scheduleExportRecordQuerySync,
)
watch([savedPresetCardKeyword, savedPresetCardSort, showSavedExportRecordPresetCards], scheduleExportRecordQuerySync)
watch(selectedSavedExportRecordPresetId, scheduleExportRecordQuerySync)
watch(exportRecords, persistExportRecords, { deep: true })
watch(savedExportRecordPresets, persistSavedExportRecordPresets, { deep: true })
watch(exportRecords, () => {
  const validIds = new Set(exportRecords.value.map((item) => item.id))
  selectedExportRecordIds.value = selectedExportRecordIds.value.filter((id) => validIds.has(id))
}, { deep: true })
watch(savedExportRecordPresets, () => {
  const validIds = new Set(savedExportRecordPresets.value.map((item) => item.id))
  if (!validIds.has(selectedSavedExportRecordPresetId.value)) {
    selectedSavedExportRecordPresetId.value = ''
  }
  compareSavedExportRecordPresetIds.value = compareSavedExportRecordPresetIds.value.filter((id) => validIds.has(id))
}, { deep: true })

const exportData = () => {
  const lines = ['benefit_id,benefit_name,type,status']
  filteredBenefits.value.forEach((item) => {
    lines.push(`${item.id},${item.name},${item.type},${item.status}`)
  })
  const blob = new Blob(["\uFEFF" + lines.join('\n')], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'points-mall-benefits.csv'
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(() => {
  hydrateExportRecords()
  hydrateSavedExportRecordPresets()
  hydrateTxFiltersFromQuery()
  querySyncReady.value = true
  txFiltersInitialized.value = true
  loadData()
})

onBeforeUnmount(() => {
  if (copiedSourceIdTimer) {
    clearTimeout(copiedSourceIdTimer)
    copiedSourceIdTimer = null
  }
  if (highlightedTxTimer) {
    clearTimeout(highlightedTxTimer)
    highlightedTxTimer = null
  }
  if (txFilterTimer) {
    clearTimeout(txFilterTimer)
    txFilterTimer = null
  }
  if (exportRecordFilterTimer) {
    clearTimeout(exportRecordFilterTimer)
    exportRecordFilterTimer = null
  }
  if (copiedExportRecordTimer) {
    clearTimeout(copiedExportRecordTimer)
    copiedExportRecordTimer = null
  }
  if (highlightedExportRecordTimer) {
    clearTimeout(highlightedExportRecordTimer)
    highlightedExportRecordTimer = null
  }
})
</script>
