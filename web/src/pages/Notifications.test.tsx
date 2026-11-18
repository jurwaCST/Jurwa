import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../api/client'
import { lastSeen } from '../notifications/seen'
import Notifications from './Notifications'

// The API client is replaced, so no gateway is needed. ApiError stays real.
vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>()
  return { ...actual, api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), upload: vi.fn(), blob: vi.fn() } }
})

const ME = 'student-1'
vi.mock('../auth/useAuth', () => ({
  useAuth: () => ({
    user: { id: ME, email: 'pema@rub.edu.bt', fullName: 'Pema Lhamo', roles: ['STUDENT'] },
    login: vi.fn(),
    logout: vi.fn(),
  }),
}))

const get = vi.mocked(api.get)

// As the gateway sends them (notification service, SCRUM-155), newest first.
const resolved = {
  id: 'n3',
  eventType: 'status.changed',
  reportId: 'r1',
  subject: 'Your report is resolved',
  body: 'HC, Floor 1, Left wing bathroom: it is marked as fixed. You can rate the fix in My reports.',
  createdAt: '2026-11-12T08:00:00Z',
}
const progress = {
  id: 'n2',
  eventType: 'status.changed',
  reportId: 'r1',
  subject: 'Your report is in progress',
  body: 'HC, Floor 1, Left wing bathroom: someone is working on it now.',
  createdAt: '2026-11-11T06:00:00Z',
}
const received = {
  id: 'n1',
  eventType: 'report.submitted',
  reportId: 'r1',
  subject: 'We received your report',
  body: 'HC, Floor 1, Left wing bathroom. You can follow it in My reports.',
  createdAt: '2026-11-10T06:00:00Z',
}

function renderPage() {
  return render(
    <MemoryRouter>
      <Notifications />
    </MemoryRouter>,
  )
}

beforeEach(() => {
  get.mockReset()
  localStorage.clear()
  // Only the date is fixed, so "Today" and "Yesterday" do not depend on the day the test runs.
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-11-12T10:00:00Z'))
})

afterEach(() => {
  vi.useRealTimers()
})

describe('SCRUM-144 notifications page', () => {
  it('lists the messages grouped by day, newest first', async () => {
    get.mockResolvedValue({ notifications: [resolved, progress, received], nextPageToken: '' })
    renderPage()

    expect(await screen.findByRole('heading', { level: 1, name: 'Notifications' })).toBeInTheDocument()
    const days = await screen.findAllByRole('heading', { level: 2 })
    expect(days.map((h) => h.textContent)).toEqual(['Today', 'Yesterday', '10 Nov 2026'])
    expect(get).toHaveBeenCalledWith('/me/notifications')

    const today = screen.getByRole('region', { name: 'Today' })
    expect(within(today).getByText(/it is marked as fixed/)).toBeInTheDocument()
  })

  it('links each message to its report', async () => {
    get.mockResolvedValue({ notifications: [resolved], nextPageToken: '' })
    renderPage()

    expect(await screen.findByRole('link', { name: 'Your report is resolved' })).toHaveAttribute('href', '/reports/r1')
  })

  it('says so when there are no messages yet', async () => {
    get.mockResolvedValue({ notifications: [], nextPageToken: '' })
    renderPage()

    expect(await screen.findByRole('heading', { name: 'No notifications yet' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Report an issue' })).toHaveAttribute('href', '/report/new')
  })

  it('shows a kind message and tries again', async () => {
    const user = userEvent.setup()
    get.mockRejectedValueOnce(new ApiError(504, 'timeout'))
    get.mockResolvedValueOnce({ notifications: [received], nextPageToken: '' })
    renderPage()

    expect(await screen.findByText(/service is busy or unreachable/i)).toBeInTheDocument()
    expect(screen.queryByText('timeout')).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(await screen.findByRole('link', { name: 'We received your report' })).toBeInTheDocument()
  })

  it('loads more when there are more', async () => {
    const user = userEvent.setup()
    get.mockResolvedValueOnce({ notifications: [resolved], nextPageToken: '1' })
    get.mockResolvedValueOnce({ notifications: [received], nextPageToken: '' })
    renderPage()

    await user.click(await screen.findByRole('button', { name: 'Load more' }))

    expect(await screen.findByRole('link', { name: 'We received your report' })).toBeInTheDocument()
    expect(get).toHaveBeenLastCalledWith('/me/notifications?pageToken=1')
    expect(screen.queryByRole('button', { name: 'Load more' })).not.toBeInTheDocument()
  })

  it('remembers the newest message as seen, so the bell dot goes away', async () => {
    get.mockResolvedValue({ notifications: [resolved, progress], nextPageToken: '' })
    renderPage()

    await screen.findByRole('link', { name: 'Your report is resolved' })
    expect(lastSeen(ME)).toBe(resolved.createdAt)
  })
})
