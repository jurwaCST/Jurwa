// The one place the PWA talks to the api-gateway (docs/design/api-routes.md).
// Every page uses `api`, so the token, the base URL and the error handling
// are the same everywhere.

/** ApiError is thrown for every failed call. status 0 means the server could not be reached. */
export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

let token: string | null = null

/** setToken sets the JWT sent as "Authorization: Bearer <token>". null signs out. */
export function setToken(next: string | null): void {
  token = next
}

// baseUrl is read on every call so tests (and a changed .env) take effect at once.
function baseUrl(): string {
  const configured = import.meta.env.VITE_API_URL || 'http://localhost:8080'
  return configured.replace(/\/+$/, '')
}

async function request<T>(method: string, path: string, body?: BodyInit, contentType?: string, raw = false): Promise<T> {
  const headers = new Headers({ Accept: raw ? '*/*' : 'application/json' })
  if (contentType) headers.set('Content-Type', contentType)
  if (token) headers.set('Authorization', `Bearer ${token}`)

  let res: Response
  try {
    res = await fetch(baseUrl() + path, { method, headers, body })
  } catch {
    throw new ApiError(0, 'Cannot reach the server. Check your connection and try again.')
  }

  if (!res.ok) {
    throw new ApiError(res.status, await errorMessage(res))
  }
  if (res.status === 204) {
    return undefined as T
  }
  if (raw) {
    return (await res.blob()) as T
  }
  return (await res.json()) as T
}

// errorMessage reads the gateway's {"error": "..."} body, or gives a generic text.
async function errorMessage(res: Response): Promise<string> {
  try {
    const data: unknown = await res.json()
    if (data && typeof data === 'object' && 'error' in data && typeof data.error === 'string') {
      return data.error
    }
  } catch {
    // not JSON (for example an HTML error page from a proxy)
  }
  return `Request failed (${res.status})`
}

export const api = {
  get<T>(path: string): Promise<T> {
    return request<T>('GET', path)
  },
  post<T>(path: string, body: unknown): Promise<T> {
    return request<T>('POST', path, JSON.stringify(body), 'application/json')
  },
  put<T>(path: string, body: unknown): Promise<T> {
    return request<T>('PUT', path, JSON.stringify(body), 'application/json')
  },
  /** blob gets a file the gateway sends as it is, such as GET /photos/{photoId}. */
  blob(path: string): Promise<Blob> {
    return request<Blob>('GET', path, undefined, undefined, true)
  },
  /** upload sends one file as multipart/form-data. The browser sets the boundary header. */
  upload<T>(path: string, file: File, field: string): Promise<T> {
    const form = new FormData()
    form.append(field, file)
    return request<T>('POST', path, form)
  },
}
