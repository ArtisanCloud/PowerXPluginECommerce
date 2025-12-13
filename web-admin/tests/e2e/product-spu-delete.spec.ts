import { expect, test } from '@playwright/test'

test.describe('SPU delete flow', () => {
	test('deletes draft spu with reason', async ({ page }) => {
		await mockDeleteApis(page)
		await page.goto('/product/spus/spu-delete')
		await page.getByRole('button', { name: '删除' }).click()
		await page.getByPlaceholder('必填，说明删除原因').fill('不再使用')
		await page.getByRole('button', { name: '确认删除' }).click()
		await expect(page).toHaveURL(/\/product\/spus$/)
	})
})

async function mockDeleteApis(page) {
	const detail = {
		id: 'spu-delete',
		code: 'SPU-DEL',
		name: 'Delete Me',
		type: 'one_time',
		status: 'draft',
		categoryId: 'cat',
		categoryPath: 'root/cat',
		defaultLocale: 'zh-CN',
		currentVersionId: 'ver-1',
		createdAt: new Date().toISOString(),
		updatedAt: new Date().toISOString(),
	}
	await page.route('**/admin/product/spus/spu-delete', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({ json: detail })
			return
		}
		await route.fulfill({ json: detail })
	})
	await page.route('**/admin/product/spus/spu-delete/skus', async (route) => {
		await route.fulfill({ json: { items: [] } })
	})
	await page.route('**/admin/product/spus/spu-delete/channels', async (route) => {
		await route.fulfill({ json: { items: [] } })
	})
	await page.route('**/admin/product/spus/spu-delete/versions', async (route) => {
		await route.fulfill({ json: { items: [] } })
	})
	await page.route('**/admin/product/spus/spu-delete/delete', async (route) => {
		await route.fulfill({ json: detail })
	})
	await page.route('**/admin/product/spus?**', async (route) => {
		await route.fulfill({ json: { items: [], meta: { total: 0, page: 1, pageSize: 20 } } })
	})
}
