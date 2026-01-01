type SuccessInit = {
	status?: number
	headers?: Record<string, string>
	wrap?: boolean
}

export function successResponse<T>(data: T, init: SuccessInit = {}) {
	const body = init.wrap === false ? data : { success: true, data }
	return {
		status: init.status ?? 200,
		contentType: 'application/json',
		headers: init.headers,
		body: JSON.stringify(body),
	}
}
