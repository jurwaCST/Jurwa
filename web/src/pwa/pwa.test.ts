// SCRUM-98: the app is installable and its shell opens offline (NFR-07).
// The manifest, index.html and the service worker are read as text, and the
// service worker runs here against a fake cache, so its rules are tested as
// they run in the browser.
import { describe, expect, it, vi } from 'vitest'
import indexHtml from '../../index.html?raw'
import manifestText from '../../public/manifest.webmanifest?raw'
import swSource from '../../public/sw.js?raw'
import { registerServiceWorker } from './register'

// Every file in web/public, by its path from this folder.
const publicFiles = Object.keys(import.meta.glob('../../public/**/*', { query: '?url', import: 'default' }))

type Icon = { src: string; sizes: string; type: string; purpose?: string }
const manifest = JSON.parse(manifestText) as Record<string, unknown> & { icons: Icon[] }

describe('SCRUM-98 manifest', () => {
  it('has what Chrome needs to install the app', () => {
    expect(manifest).toMatchObject({
      name: 'Jurwa',
      short_name: 'Jurwa',
      start_url: '/',
      scope: '/',
      display: 'standalone',
    })
    // The top bar's pine green (--pine-800 in index.css) and the page colour (--bg).
    expect(manifest.theme_color).toBe('#12372a')
    expect(manifest.background_color).toBe('#f7f6f1')
    const sizes = manifest.icons.map((i) => `${i.sizes} ${i.purpose ?? 'any'}`)
    expect(sizes).toEqual(expect.arrayContaining(['192x192 any', '512x512 any', '512x512 maskable']))
  })

  it('points only at icons that exist in public/', () => {
    for (const icon of manifest.icons) {
      expect(publicFiles).toContain(`../../public${icon.src}`)
      expect(icon.type).toBe('image/png')
    }
  })
})

describe('SCRUM-98 index.html', () => {
  it('links the manifest, the theme colour and the home-screen icon', () => {
    expect(indexHtml).toContain('<link rel="manifest" href="/manifest.webmanifest" />')
    expect(indexHtml).toContain(`<meta name="theme-color" content="${manifest.theme_color}" />`)
    expect(indexHtml).toContain('<link rel="apple-touch-icon" href="/icon-192.png" />')
  })
})

// ---- The service worker, run against a fake cache ----

type FakeResponse = { body: string; ok: boolean; clone: () => FakeResponse }
type FakeRequest = { url: string; mode?: string; method?: string }
type Listener = (event: Record<string, unknown>) => void

const ORIGIN = 'https://jurwa.test'
const response = (body: string, ok = true): FakeResponse => ({ body, ok, clone: () => response(body, ok) })
const keyOf = (r: string | FakeRequest) => (typeof r === 'string' ? r : new URL(r.url).pathname)

function loadWorker() {
  const listeners: Record<string, Listener> = {}
  const stores = new Map<string, Map<string, FakeResponse>>()
  const store = (name: string) => {
    if (!stores.has(name)) stores.set(name, new Map())
    return stores.get(name)!
  }
  const caches = {
    open: async (name: string) => {
      const s = store(name)
      return {
        addAll: async (urls: string[]) => urls.forEach((u) => s.set(u, response(`cached ${u}`))),
        put: async (req: string | FakeRequest, res: FakeResponse) => void s.set(keyOf(req), res),
        match: async (req: string | FakeRequest) => s.get(keyOf(req)),
      }
    },
    keys: async () => [...stores.keys()],
    delete: async (name: string) => stores.delete(name),
    match: async (req: string | FakeRequest) => {
      for (const s of stores.values()) {
        const hit = s.get(keyOf(req))
        if (hit) return hit
      }
      return undefined
    },
  }
  const fetch = vi.fn<(req: FakeRequest) => Promise<FakeResponse>>()
  const self = {
    addEventListener: (type: string, fn: Listener) => (listeners[type] = fn),
    skipWaiting: vi.fn(async () => undefined),
    clients: { claim: vi.fn(async () => undefined) },
    location: { origin: ORIGIN },
  }
  new Function('self', 'caches', 'fetch', swSource)(self, caches, fetch)

  // lifecycle runs install or activate and waits for it.
  async function lifecycle(type: 'install' | 'activate') {
    let done: Promise<unknown> = Promise.resolve()
    listeners[type]({ waitUntil: (p: Promise<unknown>) => (done = p) })
    await done
  }

  // request sends one fetch event; answer is undefined when the worker left
  // the request to the browser (did not call respondWith).
  async function request(path: string, init: { mode?: string; method?: string } = {}) {
    let answer: Promise<FakeResponse | undefined> | undefined
    listeners.fetch({
      request: { url: ORIGIN + path, mode: init.mode ?? 'cors', method: init.method ?? 'GET' },
      respondWith: (p: Promise<FakeResponse | undefined>) => (answer = p),
    })
    return answer ? await answer : undefined
  }

  return { lifecycle, request, fetch, stores }
}

describe('SCRUM-98 service worker', () => {
  it('keeps the app shell when it installs', async () => {
    const w = loadWorker()
    await w.lifecycle('install')
    const [cache] = [...w.stores.values()]
    expect([...cache.keys()]).toEqual(expect.arrayContaining(['/', '/index.html', '/icon-192.png', '/manifest.webmanifest']))
  })

  it('never answers or keeps an API call', async () => {
    const w = loadWorker()
    await w.lifecycle('install')
    expect(await w.request('/api/me/reports')).toBeUndefined()
    expect(await w.request('/api/reports', { method: 'POST' })).toBeUndefined()
    expect(await w.request('/api')).toBeUndefined()
    expect(w.fetch).not.toHaveBeenCalled()
    for (const cache of w.stores.values()) {
      expect([...cache.keys()].some((k) => k.startsWith('/api'))).toBe(false)
    }
  })

  // Gayley (SCRUM-98 review): reports, photos and notifications are private
  // to each student; on a shared phone or lab PC the next person must never
  // see them. Only the app's own files may be answered or kept.
  it('leaves private data alone, whatever its address', async () => {
    const w = loadWorker()
    await w.lifecycle('install')
    const privatePaths = [
      '/api/me/reports',
      '/api/reports/r1/timeline',
      '/api/photos/p1',
      '/me/reports',
      '/me/notifications',
      '/reports/r1',
      '/photos/p1',
    ]
    for (const path of privatePaths) {
      expect(await w.request(path)).toBeUndefined()
    }
    expect(w.fetch).not.toHaveBeenCalled()
    const kept = [...w.stores.values()].flatMap((cache) => [...cache.keys()])
    expect(kept.every((k) => k === '/' || k === '/index.html' || k.startsWith('/assets/') || /^\/[\w.-]+\.(png|svg|webmanifest)$/.test(k))).toBe(true)
  })

  it('opens a page from the network, and from the saved shell when offline', async () => {
    const w = loadWorker()
    await w.lifecycle('install')
    w.fetch.mockResolvedValueOnce(response('fresh page'))
    expect((await w.request('/staff', { mode: 'navigate' }))?.body).toBe('fresh page')

    w.fetch.mockRejectedValueOnce(new TypeError('offline'))
    expect((await w.request('/my-reports', { mode: 'navigate' }))?.body).toBe('cached /index.html')
  })

  it('keeps built files the first time and then serves them without the network', async () => {
    const w = loadWorker()
    w.fetch.mockResolvedValueOnce(response('app code'))
    expect((await w.request('/assets/index-abc123.js'))?.body).toBe('app code')
    expect((await w.request('/assets/index-abc123.js'))?.body).toBe('app code')
    expect(w.fetch).toHaveBeenCalledTimes(1)
  })

  it('does not keep a failed built file', async () => {
    const w = loadWorker()
    w.fetch.mockResolvedValueOnce(response('not found', false))
    await w.request('/assets/missing.js')
    w.fetch.mockResolvedValueOnce(response('app code'))
    expect((await w.request('/assets/missing.js'))?.body).toBe('app code')
  })

  it('removes the caches of older versions when it activates', async () => {
    const w = loadWorker()
    w.stores.set('jurwa-old', new Map([['/', response('old')]]))
    await w.lifecycle('install')
    await w.lifecycle('activate')
    expect([...w.stores.keys()]).toHaveLength(1)
    expect([...w.stores.keys()][0]).not.toBe('jurwa-old')
  })
})

describe('SCRUM-98 registration', () => {
  it('registers /sw.js in the built app only', async () => {
    const register = vi.fn(async () => ({}))
    const nav = { serviceWorker: { register } }
    expect(await registerServiceWorker(false, nav)).toBe(false)
    expect(register).not.toHaveBeenCalled()
    expect(await registerServiceWorker(true, nav)).toBe(true)
    expect(register).toHaveBeenCalledWith('/sw.js')
  })

  it('never breaks the app when the browser has no service workers or refuses', async () => {
    expect(await registerServiceWorker(true, {})).toBe(false)
    const refuse = { serviceWorker: { register: vi.fn(async () => Promise.reject(new Error('blocked'))) } }
    expect(await registerServiceWorker(true, refuse)).toBe(false)
  })
})
