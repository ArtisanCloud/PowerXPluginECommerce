import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.E2E_BASE_URL || 'http://127.0.0.1:3001'

export default defineConfig({
	testDir: './tests/e2e',
	timeout: 60_000,
	expect: {
		timeout: 10_000,
	},
	fullyParallel: true,
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
			'NUXT_PUBLIC_API_BASE=http://127.0.0.1:3001/api/v1 DISABLE_DEV_PROXY=1 npm run dev -- --port 3001',
		port: 3001,
		reuseExistingServer: !process.env.CI,
		timeout: 120 * 1000,
	},
})
