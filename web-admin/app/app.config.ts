const defaultPluginId = 'com.powerx.plugin.ecommerce'
const resolvePluginId = () => {
  const candidates = [
    process.env.NUXT_PUBLIC_POWERX_PLUGIN_ID,
    process.env.POWERX_PLUGIN_ID,
    defaultPluginId
  ]
  for (const candidate of candidates) {
    const trimmed = candidate?.trim()
    if (trimmed) {
      return trimmed
    }
  }
  return defaultPluginId
}

const pluginId = resolvePluginId()
const productSkuApiBase = `/_p/${pluginId}/api/v1/products/skus`

export default defineAppConfig({
  productSku: {
    menuRootPath: '/product',
    skuListPath: '/product/skus',
    inventoryPath: '/product/inventory',
    defaultApiEndpoints: {
      base: productSkuApiBase,
      bulkTasks: `${productSkuApiBase}/bulk-tasks`,
      import: `${productSkuApiBase}/import`,
      export: `${productSkuApiBase}/export`
    }
  }
})
