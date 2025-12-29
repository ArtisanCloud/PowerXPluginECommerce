import type { Page } from '@playwright/test'

const header = Buffer.from(JSON.stringify({ alg: 'none', typ: 'JWT' })).toString('base64url')
const tenantUuid = 'tenant-e2e'
const payload = Buffer.from(JSON.stringify({ tenant_uuid: tenantUuid, tid: tenantUuid, tenantId: tenantUuid })).toString(
	'base64url',
)
const token = `${header}.${payload}.stub`

const shouldLogRequests = process.env.E2E_LOG_REQUESTS === '1'

export async function bootstrapAuthenticatedSession(page: Page) {
	const expiresAt = Date.now() + 60 * 60 * 1000
	await page.addInitScript(
		({ token, refreshToken, expiresAt, tenantUuid }) => {
			window.localStorage.setItem('access_token', token)
			window.localStorage.setItem('refresh_token', refreshToken)
			window.localStorage.setItem('token_type', 'Bearer')
			window.localStorage.setItem('expires_in', '3600')
			window.localStorage.setItem('scope', 'admin')
			window.localStorage.setItem('expires_at', String(expiresAt))
			document.cookie = `token=${token}; path=/; SameSite=Lax`
			document.cookie = `tenant_uuid=${tenantUuid}; path=/; SameSite=Lax`
		},
		{ token, refreshToken: 'refresh-e2e', expiresAt, tenantUuid },
	)
	if (shouldLogRequests) {
		page.on('request', (request) => {
			if (request.url().includes('/admin/product')) {
				console.info('[e2e] request', request.method(), request.url())
			}
		})
		page.on('requestfailed', (request) => {
			if (request.url().includes('/admin/product')) {
				console.warn('[e2e] request failed', request.url(), request.failure()?.errorText)
			}
		})
	}
}
