// Words for a student's own reports, shared by My reports and the report page (SCRUM-143).
import { ApiError } from '../api/client'

export type Location = { building: string; wing?: string; area?: string }

export const categoryNames: Record<string, string> = {
  HOSTEL: 'Hostel',
  ACADEMIC_BLOCK: 'Academic Block',
  OTHERS: 'Others',
}

/** "HC, Floor 1, Room 101". Campus places have no wing. */
export function place(l: Location): string {
  return [l.building, l.wing, l.area].filter(Boolean).join(', ')
}

/** A short number to tell reports apart: "#89533B76". */
export function reportNumber(id: string): string {
  return `#${id.slice(0, 8).toUpperCase()}`
}

const relative = new Intl.RelativeTimeFormat('en-GB', { numeric: 'auto' })

/** "3 hours ago", "2 days ago". */
export function age(iso: string | undefined, now = Date.now()): string {
  if (!iso) return ''
  const minutes = Math.round((new Date(iso).getTime() - now) / 60000)
  if (Math.abs(minutes) < 60) return relative.format(minutes, 'minute')
  const hours = Math.round(minutes / 60)
  if (Math.abs(hours) < 24) return relative.format(hours, 'hour')
  return relative.format(Math.round(hours / 24), 'day')
}

/** A sentence for a failed call; never the raw error text. */
export function messageFor(err: unknown): string {
  if (err instanceof ApiError && err.status === 401) return 'Your sign-in has expired. Please sign in again.'
  return 'The service is busy or unreachable. Please try again.'
}
