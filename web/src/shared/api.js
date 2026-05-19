export async function api(path, options = {}) {
  const isForm = options.body instanceof FormData
  const headers = isForm
    ? { ...(options.headers || {}) }
    : { 'Content-Type': 'application/json', ...(options.headers || {}) }

  const response = await fetch(path, { ...options, headers })
  const contentType = response.headers.get('content-type') || ''
  const body = contentType.includes('application/json') ? await response.json() : await response.text()

  if (!response.ok) {
    const message = typeof body === 'object' ? body.error || body.message : body
    throw new Error(message || 'Request failed')
  }

  return typeof body === 'object' && body !== null && 'data' in body ? body.data : body
}

export function buildQuery(params) {
  const search = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      search.set(key, value)
    }
  })
  const query = search.toString()
  return query ? `?${query}` : ''
}
