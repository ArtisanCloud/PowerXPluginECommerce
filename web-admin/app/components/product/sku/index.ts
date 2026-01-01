import { defineAsyncComponent } from 'vue'

export const SkuWorkspaceBanner = defineAsyncComponent(() => import('./SkuWorkspaceBanner.vue'))
export const SkuEmptyState = defineAsyncComponent(() => import('./SkuEmptyState.vue'))
export const SkuGeneratorPanel = defineAsyncComponent(() => import('./GeneratorPanel.vue'))
export const SkuDefaultValueForm = defineAsyncComponent(() => import('./DefaultValueForm.vue'))
export const SkuBulkImportUploader = defineAsyncComponent(() => import('./BulkImportUploader.vue'))
export const SkuBulkExportPanel = defineAsyncComponent(() => import('./BulkExportPanel.vue'))
export const SkuBulkTaskStatusList = defineAsyncComponent(() => import('./BulkTaskStatusList.vue'))
export const SkuMatrixGrid = defineAsyncComponent(() => import('./MatrixGrid.vue'))
export const SkuChannelMappingTab = defineAsyncComponent(() => import('./ChannelMappingTab.vue'))
export const SkuInventoryPanel = defineAsyncComponent(() => import('./InventorySnapshotCard.vue'))
export const SkuBarcodePanel = defineAsyncComponent(() => import('./BarcodePanel.vue'))
export const SkuSerialBatchSection = defineAsyncComponent(() => import('./SerialBatchSection.vue'))

export const skuComponentRegistry = {
	banner: SkuWorkspaceBanner,
	emptyState: SkuEmptyState,
	generator: SkuGeneratorPanel,
	defaultForm: SkuDefaultValueForm,
	importUploader: SkuBulkImportUploader,
	exportPanel: SkuBulkExportPanel,
	tasks: SkuBulkTaskStatusList,
	matrix: SkuMatrixGrid,
	channels: SkuChannelMappingTab,
	inventory: SkuInventoryPanel,
	barcodes: SkuBarcodePanel,
	serials: SkuSerialBatchSection,
}
