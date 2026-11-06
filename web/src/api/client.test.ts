import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError, setToken } from './client'

// jsonResponse builds a fake fetch Response with a JSON body.
function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

let fetchMock: ReturnType<typeof vi.fn>

beforeEach(() => {
  fetchMock = vi.fn()
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  setToken(null)
  vi.unstubAllGlobals()
  vi.unstubAllEnvs()
})

// call returns the URL and options of fetch call number n.
function call(n = 0): [string, RequestInit] {
  return fetchMock.mock.calls[n] as [string, RequestInit]
}

describe('base URL', () => {
  it('defaults to the gateway on localhost:8080', async () => {
    vi.stubEnv('VITE_API_URL', undefined)
    fetchMock.mockResolvedValue(jsonResponse(200, { status: 'ok' }))
    await api.get('/health')
    expect(call()[0]).toBe('http://localhost:8080/health')
  })

  it('uses VITE_API_URL when it is set, without a double slash', async () => {
    vi.stubEnv('VITE_API_URL', 'https://jurwa.local/api/')
    fetchMock.mockResolvedValue(jsonResponse(200, {}))
    await api.get('/categories')
    expect(call()[0]).toBe('https://jurwa.local/api/categories')
  })
})

describe('token', () => {
  it('sends no Authorization header before setToken', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))
    await api.get('/health')
    expect(new Headers(call()[1].headers).has('Authorization')).toBe(false)
  })

  it('sends Authorization: Bearer <token> after setToken', async () => {
    setToken('abc.def.ghi')
    fetchMock.mockResolvedValue(jsonResponse(200, {}))
    await api.get('/me')
    expect(new Headers(call()[1].headers).get('Authorization')).toBe('Bearer abc.def.ghi')
  })

  it('stops sending it after setToken(null)', async () => {
    setToken('abc')
    setToken(null)
    fetchMock.mockResolvedValue(jsonResponse(200, {}))
    await api.get('/me')
    expect(new Headers(call()[1].headers).has('Authorization')).toBe(false)
  })
})

describe('methods', () => {
  it('get returns the parsed JSON body', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { categories: [{ name: 'Hostel' }] }))
    const res = await api.get<{ categories: { name: string }[] }>('/categories')
    expect(res.categories[0].name).toBe('Hostel')
    expect(call()[1].method).toBe('GET')
  })

  it('post sends the body as JSON', async () => {
    fetchMock.mockResolvedValue(jsonResponse(201, { id: 'r1' }))
    const res = await api.post<{ id: string }>('/reports', { description: 'Leaking tap' })
    expect(res.id).toBe('r1')
    const [, init] = call()
    expect(init.method).toBe('POST')
    expect(new Headers(init.headers).get('Content-Type')).toBe('application/json')
    expect(init.body).toBe(JSON.stringify({ description: 'Leaking tap' }))
  })

  it('put sends the body as JSON', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))
    await api.put('/admin/sla-policies/HOSTEL', { hours: 48 })
    const [, init] = call()
    expect(init.method).toBe('PUT')
    expect(init.body).toBe(JSON.stringify({ hours: 48 }))
  })

  it('upload sends the file as multipart form data under the given field', async () => {
    fetchMock.mockResolvedValue(jsonResponse(201, { photoId: 'p1' }))
    const file = new File(['png-bytes'], 'tap.png', { type: 'image/png' })
    const res = await api.upload<{ photoId: string }>('/photos', file, 'photo')
    expect(res.photoId).toBe('p1')
    const [, init] = call()
    expect(init.method).toBe('POST')
    expect(init.body).toBeInstanceOf(FormData)
    expect((init.body as FormData).get('photo')).toBeInstanceOf(File)
    // The browser sets the multipart boundary itself, so we must not set Content-Type.
    expect(new Headers(init.headers).has('Content-Type')).toBe(false)
  })

  it('returns undefined for 204 No Content', async () => {
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }))
    await expect(api.post('/reports/r1/notes', { text: 'ok' })).resolves.toBeUndefined()
  })
})

describe('errors', () => {
  it('turns a 401 into an ApiError with status 401', async () => {
    fetchMock.mockResolvedValue(jsonResponse(401, { error: 'missing token' }))
    const err = await api.get('/me').catch((e: unknown) => e)
    expect(err).toBeInstanceOf(ApiError)
    expect((err as ApiError).status).toBe(401)
  })

  it('uses the "error" field of a JSON error body as the message', async () => {
    fetchMock.mockResolvedValue(jsonResponse(400, { error: 'unknown category' }))
    await expect(api.get('/categories/NOPE/locations')).rejects.toThrow('unknown category')
  })

  it('falls back to a generic message when the body is not JSON', async () => {
    fetchMock.mockResolvedValue(new Response('<html>bad gateway</html>', { status: 502 }))
    const err = (await api.get('/health').catch((e: unknown) => e)) as ApiError
    expect(err.status).toBe(502)
    expect(err.message).toBe('Request failed (502)')
  })

  it('turns a network failure into an ApiError with status 0', async () => {
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))
    const err = (await api.get('/health').catch((e: unknown) => e)) as ApiError
    expect(err).toBeInstanceOf(ApiError)
    expect(err.status).toBe(0)
    expect(err.message).toBe('Cannot reach the server. Check your connection and try again.')
  })
})
