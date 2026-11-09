// Words people see for roles and statuses, shared by every page.
import type { Role } from '../auth/types'
import type { IconName } from './Icon'

/** How each role is named to people, not to the system. */
export const roleNames: Record<Role, string> = {
  STUDENT: 'Student',
  CATEGORY_HEAD: 'Category Head',
  STAFF: 'Staff',
  SYSTEM_ADMIN: 'System Administrator',
  PRESIDENT: 'College President',
}

export type Status = 'OPEN' | 'ASSIGNED' | 'IN_PROGRESS' | 'RESOLVED' | 'ESCALATED'

/** Every status has a word, an icon and a colour key, so colour is never the only cue. */
export const statusLooks: Record<Status, { label: string; icon: IconName; tone: string }> = {
  OPEN: { label: 'Open', icon: 'open', tone: 'open' },
  ASSIGNED: { label: 'Assigned', icon: 'assigned', tone: 'assigned' },
  IN_PROGRESS: { label: 'In progress', icon: 'progress', tone: 'progress' },
  RESOLVED: { label: 'Resolved', icon: 'resolved', tone: 'resolved' },
  ESCALATED: { label: 'Escalated', icon: 'escalated', tone: 'escalated' },
}
