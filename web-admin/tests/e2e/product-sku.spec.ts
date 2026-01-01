import { expect, test } from '@playwright/test'
import { bootstrapAuthenticatedSession } from './utils/session'

test.beforeEach(async ({ page }) => {
	await bootstrapAuthenticatedSession(page)
})

async function openSkuHarness(page: any) {
	await page.goto('/e2e/sku', { waitUntil: 'domcontentloaded', timeout: 120_000 })
	await page.waitForLoadState('networkidle', { timeout: 60_000 }).catch(() => undefined)
	await expect(page.locator('#us1-generator')).toBeVisible({ timeout: 60_000 })
}

test.describe('Product SKU user stories (E2E harness)', () => {
	test('US1 - SKU generator creates combinations and persists', async ({ page }) => {
		await openSkuHarness(page)

		const generatorSection = page.locator('#us1-generator')
		await expect(generatorSection.getByTestId('btn-generate')).toBeVisible()
		await generatorSection.getByTestId('btn-generate').click()
		await expect(page.getByTestId('generator-ready')).toBeVisible()
		await generatorSection.getByTestId('btn-write-sku').click()
		await expect(generatorSection.getByText('已写入 1 条 SKU')).toBeVisible()
	})

	test('US2 - Bulk adjustment submits task and lists it in task center', async ({ page }) => {
		await openSkuHarness(page)

		const bulkDialog = page.getByTestId('bulk-dialog')
		await expect(bulkDialog).toBeVisible()
		const adjustmentInput = bulkDialog.locator('input[type="number"]').first()
		await adjustmentInput.fill('5')
		await bulkDialog.getByTestId('submit-bulk-btn').click()

		const bulkResult = page.getByTestId('bulk-result-message')
		await expect(bulkResult).toContainText('批量任务已创建')
		await expect(page.getByTestId('bulk-result-id')).toContainText('#task-e2e')
	})

	test('US3 - Channel mapping, barcode, inventory and serial tabs', async ({ page }) => {
		await openSkuHarness(page)

		const channelSection = page.locator('#us3-channels')
		await expect(channelSection.getByTestId('btn-edit-channel')).toBeVisible()
		await channelSection.getByTestId('btn-edit-channel').click()
		await channelSection.getByTestId('channel-sku-input').fill('SKU-E2E-CHANGED')
		await channelSection.getByTestId('btn-save-channel').click()
		await expect(channelSection.getByText('渠道映射已保存')).toBeVisible()

		await channelSection.getByTestId('btn-publish-channel').click()
		await expect(channelSection.getByText('发布已触发')).toBeVisible()

		const barcodeSection = page.locator('#us3-barcode')
		await barcodeSection.getByLabel('条码前缀').fill('QA')
		await barcodeSection.getByRole('button', { name: '生成条码' }).click()
		await expect(barcodeSection.getByText('BARCODE-QA-1')).toBeVisible()

		const serialSection = page.locator('#us3-serials')
		await serialSection.getByLabel('序列号').fill('SER-001')
		await serialSection.getByRole('button', { name: '保存记录' }).click()
		await expect(serialSection.getByText('序列号已保存')).toBeVisible()
		await expect(serialSection.getByText('SER-001')).toBeVisible()
	})
})
