// Which notifications a student has seen (SCRUM-144). Owner: Gayley.
// The gateway has no "mark as read" route, so the browser remembers the time
// of the newest message the student has seen, per student. The bell shows a
// dot for anything newer. It is per browser: another device starts unseen.
// Storage can be blocked (private windows), so every call is guarded.

/** Fired on window when the seen time changes, so the bell updates at once. */
export const SEEN_EVENT = 'jurwa:notifications-seen'

const key = (userId: string) => `jurwa.notifications.seen.${userId}`

/** The time (ISO) of the newest message this student has seen, or null. */
export function lastSeen(userId: string): string | null {
  try {
    return localStorage.getItem(key(userId))
  } catch {
    return null
  }
}

/** Remembers iso as seen, unless a later time is already remembered. */
export function markSeen(userId: string, iso: string): void {
  if (!isNewer(iso, lastSeen(userId))) return
  try {
    localStorage.setItem(key(userId), iso)
  } catch {
    // Not remembered; the dot just stays until the next visit.
  }
  window.dispatchEvent(new Event(SEEN_EVENT))
}

/** True when iso is later than seen (or nothing was seen yet). */
export function isNewer(iso: string, seen: string | null): boolean {
  if (!seen) return true
  return Date.parse(iso) > Date.parse(seen)
}
