import { defineAsyncComponent } from 'vue'

export const SkuWorkspaceBanner = defineAsyncComponent(() => import('./SkuWorkspaceBanner.vue'))
export const SkuEmptyState = defineAsyncComponent(() => import('./SkuEmptyState.vue'))

export const skuComponentRegistry = {
	banner: SkuWorkspaceBanner,
	emptyState: SkuEmptyState,
}
