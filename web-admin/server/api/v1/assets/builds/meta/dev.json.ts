import { defineEventHandler } from 'h3'

export default defineEventHandler(() => {
	return {
		builds: [],
		generatedAt: new Date().toISOString(),
	}
})
