import { expect, test } from '@playwright/test'
import { bootstrapAuthenticatedSession } from './utils/session'

test.beforeEach(async ({ page }) => {
	await bootstrapAuthenticatedSession(page)
})

test.describe('Product SKU user stories (E2E harness)', () => {
	test('US1 - SKU generator creates combinations and persists', async ({ page }) => {
		await page.goto('/e2e/sku')

		const generatorSection = page.locator('#us1-generator')
		await generatorSection.getByRole('button', { name: '生成预览' }).click()
		await expect(page.getByTestId('generator-ready')).toBeVisible()
		await generatorSection.getByRole('button', { name: '写入 SKU' }).click()
		await expect(generatorSection.getByText('已写入 1 条 SKU')).toBeVisible()
	})

	test('US2 - Bulk adjustment submits task and lists it in task center', async ({ page }) => {
		await page.goto('/e2e/sku')

		const bulkSection = page.locator('#us2-bulk')
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
		await page.goto('/e2e/sku')

		const channelSection = page.locator('#us3-channels')
		await channelSection.getByRole('button', { name: '编辑' }).first().click()
		await channelSection.getByTestId('channel-sku-input').fill('SKU-E2E-CHANGED')
		await channelSection.getByRole('button', { name: '保存映射' }).click()
		await expect(channelSection.getByText('渠道映射已保存')).toBeVisible()

		await channelSection.getByRole('button', { name: '推送发布' }).first().click()
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
