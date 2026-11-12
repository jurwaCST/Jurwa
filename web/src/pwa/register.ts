// Registers the service worker in public/sw.js (SCRUM-98). Owner: Nyendrak.
// Only in the built app: in `npm run dev` a service worker would keep old
// files and hide changes. It never stops the app from starting: a browser
// without service workers, or one that refuses, just runs the app online.

type Nav = { serviceWorker?: { register: (url: string) => Promise<unknown> } }

/** registerServiceWorker returns true when /sw.js was registered. */
export async function registerServiceWorker(
  enabled: boolean = import.meta.env.PROD,
  nav: Nav = navigator as unknown as Nav,
): Promise<boolean> {
  if (!enabled || !nav.serviceWorker) return false
  try {
    await nav.serviceWorker.register('/sw.js')
    return true
  } catch {
    return false
  }
}
