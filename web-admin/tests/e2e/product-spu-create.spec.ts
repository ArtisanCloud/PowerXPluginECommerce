import { expect, test } from '@playwright/test'

test.describe('SPU creation wizard', () => {
	test('fills wizard and submits draft when APIs succeed', async ({ page }) => {
		await mockSpuApis(page)
		const code = `SPU-E2E-${Date.now()}`
		await page.goto('/product/spus/create')
		await page.getByTestId('spu-code-input').fill(code)
		await page.getByTestId('spu-name-input').fill('自动化商品')
		await page.getByTestId('spu-category-id').fill('cat-e2e')
		await page.getByTestId('spu-category-path').fill('root/cat-e2e')
		await page.getByTestId('wizard-next-button').click()
		await page.getByTestId('locale-title-zh-CN').fill('测试标题')
		await page.getByTestId('wizard-next-button').click()
		await page.getByTestId('wizard-submit-button').click()
		await expect(page).toHaveURL(/\/product\/spus\/spu-e2e$/)
		await expect(page.getByRole('heading', { level: 1 })).toContainText('自动化商品')
	})
})

async function mockSpuApis(page) {
	const detail = {
		id: 'spu-e2e',
		code: 'SPU-E2E',
		name: '自动化商品',
		type: 'one_time',
		status: 'draft',
		categoryId: 'cat-e2e',
		categoryPath: 'root/cat-e2e',
		defaultLocale: 'zh-CN',
		currentVersionId: 'ver-e2e',
		createdAt: new Date().toISOString(),
		updatedAt: new Date().toISOString(),
	}
	await page.route('**/admin/product/spus?**', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({ json: { items: [], meta: { total: 0, page: 1, pageSize: 20 } } })
			return
		}
		await route.continue()
	})
	await page.route('**/admin/product/spus', async (route) => {
		if (route.request().method() === 'POST') {
			await route.fulfill({ json: detail })
			return
		}
		await route.fulfill({ json: { items: [], meta: { total: 0, page: 1, pageSize: 20 } } })
	})
	await page.route('**/admin/product/spus/spu-e2e', async (route) => {
		await route.fulfill({ json: detail })
	})
	await page.route('**/admin/product/spus/spu-e2e/skus', async (route) => {
		await route.fulfill({ json: { items: [] } })
	})
}
