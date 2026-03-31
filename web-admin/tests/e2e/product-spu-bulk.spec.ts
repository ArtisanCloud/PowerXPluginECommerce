import { expect, test } from '@playwright/test'
import { bootstrapAuthenticatedSession } from './utils/session'
import { successResponse } from './utils/response'

test.beforeEach(async ({ page }) => {
	await bootstrapAuthenticatedSession(page)
})

test.describe('SPU channel visibility and bulk import', () => {
	test('configures channel visibility from detail modal', async ({ page }) => {
		const tracker = { channelSaved: false }
		await mockSpuChannelApis(page, tracker)

		await page.goto('/product/spus/edit/spu-e2e')
		await page.getByRole('button', { name: '渠道可见性' }).click()
		await page.getByLabel('上架状态').selectOption('published')
		await page.getByLabel('渠道标题').fill('双十一主会场')
		await page.getByRole('button', { name: '保存渠道' }).click()

		await expect.poll(() => tracker.channelSaved).toBeTruthy()
		await expect(page.getByText('已配置渠道')).toBeVisible()
		await expect(page.getByText('双十一主会场')).toBeVisible()
	})

	test('submits bulk import task from list page', async ({ page }) => {
		const tracker = { importSubmitted: false }
		await mockSpuListBulkApis(page, tracker)

		await page.goto('/product/spus')
		await page.getByRole('button', { name: '导入' }).click()
		await expect(page.getByText('批量导入 SPU')).toBeVisible()

		await page.setInputFiles('input[type="file"]', {
			name: 'spu-import.csv',
			mimeType: 'text/csv',
			buffer: Buffer.from('code,name,type\nSPU-BULK-1,批量商品,one_time\n'),
		})
		await page.getByRole('button', { name: '开始导入' }).click()

		await expect.poll(() => tracker.importSubmitted).toBeTruthy()
		await expect(page.getByText('批量导入 SPU')).not.toBeVisible()
	})
})

async function mockSpuChannelApis(page, tracker: { channelSaved: boolean }) {
	const detail = {
		id: 'spu-e2e',
		code: 'SPU-E2E',
		name: '自动化商品',
		type: 'one_time',
		status: 'published',
		categoryId: 'cat-e2e',
		categoryPath: 'root/cat-e2e',
		defaultLocale: 'zh-CN',
		currentVersionId: 'ver-e2e',
		createdAt: new Date().toISOString(),
		updatedAt: new Date().toISOString(),
	}
	const channels = [
		{
			id: 'ch-official',
			channel: 'official',
			availability: 'published',
			contentOverride: { title: '双十一主会场', description: '活动期可见' },
			updatedAt: new Date().toISOString(),
		},
	]

	await page.route('**/admin/product/spus/spu-e2e/channels*', async (route) => {
		if (route.request().method() === 'POST') {
			tracker.channelSaved = true
			await route.fulfill(successResponse(channels[0]))
			return
		}
		await route.fulfill(successResponse({ items: channels }))
	})
	await page.route('**/admin/product/spus/spu-e2e/versions/ver-e2e*', async (route) => {
		await route.fulfill(
			successResponse({
				id: 'ver-e2e',
				spuId: 'spu-e2e',
				versionNumber: 1,
				status: 'published',
				diff: [],
				approvals: [],
				payload: {},
			}),
		)
	})
	await page.route('**/admin/product/spus/spu-e2e/versions*', async (route) => {
		await route.fulfill(
			successResponse({
				items: [{ id: 'ver-e2e', versionNumber: 1, status: 'published', createdAt: new Date().toISOString() }],
			}),
		)
	})
	await page.route('**/admin/product/spus/spu-e2e/skus*', async (route) => {
		await route.fulfill(successResponse({ items: [] }))
	})
	await page.route('**/admin/product/spus/spu-e2e*', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill(successResponse(detail))
			return
		}
		await route.fulfill(successResponse(detail))
	})
}

async function mockSpuListBulkApis(page, tracker: { importSubmitted: boolean }) {
	await page.route('**/admin/product/spus/import*', async (route) => {
		tracker.importSubmitted = true
		await route.fulfill(successResponse({ taskId: 'task-bulk-import-e2e' }))
	})
	await page.route('**/admin/product/spus?**', async (route) => {
		await route.fulfill(successResponse({ items: [], meta: { total: 0, page: 1, pageSize: 10 } }))
	})
	await page.route('**/admin/product/spus*', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill(successResponse({ items: [], meta: { total: 0, page: 1, pageSize: 10 } }))
			return
		}
		await route.continue()
	})
}

