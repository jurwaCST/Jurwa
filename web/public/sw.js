// Jurwa service worker (SCRUM-98). Owner: Nyendrak.
// Makes the app installable and lets its shell open offline (NFR-07).
// Written by hand, no library. Only the app's own files are ever answered or
// kept here, never data: reports, photos, notes and notifications are private
// to each user, and on a shared phone or lab PC the next person must not see
// them (Gayley's review). Rules:
// - API calls (/api/...) and any other address that is not an app file are
//   left to the browser: always the network, never kept.
// - Pages (navigations): network first; offline, the saved index.html, which
//   is the empty app shell (the app's router then shows the page).
// - Built files in /assets/ (JS, CSS, fonts) have a hash in their name: kept
//   the first time they load, then served from the cache.
// - The shell files (icons, manifest): network first, the saved copy offline.
// Change CACHE_VERSION when the shell list changes; older caches are deleted.

const CACHE_VERSION = 'jurwa-shell-v1'
const SHELL = ['/', '/index.html', '/manifest.webmanifest', '/favicon.svg', '/icon-192.png', '/icon-512.png']

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches
      .open(CACHE_VERSION)
      .then((cache) => cache.addAll(SHELL))
      .then(() => self.skipWaiting()),
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((names) => Promise.all(names.filter((name) => name !== CACHE_VERSION).map((name) => caches.delete(name))))
      .then(() => self.clients.claim()),
  )
})

self.addEventListener('fetch', (event) => {
  const request = event.request
  if (request.method !== 'GET') return
  const url = new URL(request.url)
  if (url.origin !== self.location.origin) return
  if (url.pathname === '/api' || url.pathname.startsWith('/api/')) return

  if (request.mode === 'navigate') {
    event.respondWith(fetch(request).catch(() => caches.match('/index.html')))
    return
  }

  if (url.pathname.startsWith('/assets/')) {
    event.respondWith(
      caches.match(request).then(
        (saved) =>
          saved ||
          fetch(request).then((response) => {
            if (!response.ok) return response
            const copy = response.clone()
            return caches
              .open(CACHE_VERSION)
              .then((cache) => cache.put(request, copy))
              .then(() => response)
          }),
      ),
    )
    return
  }

  if (SHELL.includes(url.pathname)) {
    event.respondWith(fetch(request).catch(() => caches.match(request)))
  }
  // Anything else is not an app file: not answered here, never kept.
})
