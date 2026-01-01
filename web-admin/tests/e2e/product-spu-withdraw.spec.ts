import { expect, test } from '@playwright/test'
import { bootstrapAuthenticatedSession } from './utils/session'
import { successResponse } from './utils/response'

test.beforeEach(async ({ page }) => {
	await bootstrapAuthenticatedSession(page)
})

test.describe('SPU withdraw flow', () => {
	test('submits withdraw request with reason', async ({ page }) => {
		await mockWithdrawApis(page)
		await page.goto('/product/spus/spu-e2e')
		await page.getByRole('button', { name: '下架' }).click()
		await page.getByPlaceholder('必填，下架原因').fill('停售下架')
		await page.getByRole('button', { name: '确认下架' }).click()
		await expect(page.getByText('当前状态：')).toContainText('offboarded')
	})
})

async function mockWithdrawApis(page) {
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
	await page.route('**/admin/product/spus/spu-e2e*', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill(successResponse(detail))
			return
		}
		await route.continue()
	})
	await page.route('**/admin/product/spus/spu-e2e/channels*', async (route) => {
		await route.fulfill(
			successResponse({
				items: [
					{ id: 'ch1', channel: 'official', availability: 'published', updatedAt: new Date().toISOString() },
				],
			}),
		)
	})
	await page.route('**/admin/product/spus/spu-e2e/skus*', async (route) => {
		await route.fulfill(successResponse({ items: [] }))
	})
	await page.route('**/admin/product/spus/spu-e2e/withdraw*', async (route) => {
		await route.fulfill(successResponse({ ...detail, status: 'offboarded', updatedAt: new Date().toISOString() }))
	})
	await page.route('**/admin/product/spus/spu-e2e/versions*', async (route) => {
		await route.fulfill(
			successResponse({
				items: [{ id: 'ver-e2e', versionNumber: 1, status: 'published', createdAt: new Date().toISOString() }],
			}),
		)
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
	await page.route('**/admin/product/spus/spu-e2e/submit*', async (route) => {
		await route.fulfill(successResponse(detail))
	})
	await page.route('**/admin/product/spus/spu-e2e/publish*', async (route) => {
		await route.fulfill(successResponse(detail))
	})
}
