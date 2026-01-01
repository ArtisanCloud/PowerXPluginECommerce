import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.E2E_BASE_URL || 'http://127.0.0.1:3001'
const isCI = process.env.CI === 'true' || process.env.CI === '1'

export default defineConfig({
	testDir: './tests/e2e',
	timeout: isCI ? 90_000 : 60_000,
	expect: {
		timeout: isCI ? 30_000 : 10_000,
	},
	fullyParallel: !isCI,
	workers: isCI ? 1 : undefined,
	reporter: [['list']],
	use: {
		baseURL,
		trace: 'on-first-retry',
		screenshot: 'only-on-failure',
		video: 'retain-on-failure',
	},
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] },
		},
	],
	webServer: {
		command:
			'NUXT_PUBLIC_E2E_HARNESS=1 QUIET_START=1 NUXT_PUBLIC_API_BASE=http://127.0.0.1:3001 NUXT_PUBLIC_API_PREFIX=/api/v1 DISABLE_DEV_PROXY=1 npm run dev -- --port 3001 --host 127.0.0.1',
		port: 3001,
		reuseExistingServer: !isCI,
		timeout: 180 * 1000,
	},
})
