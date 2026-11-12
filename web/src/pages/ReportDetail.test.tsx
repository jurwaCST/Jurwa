import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../api/client'
import ReportDetail from './ReportDetail'

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
const blob = vi.mocked(api.blob)

const ID = 'r1'

// Shapes as the gateway sends them (report-submission, status-lifecycle, report-history).
const report = {
  report: {
    id: ID,
    reporterId: ME,
    category: 'HOSTEL',
    location: { building: 'HC', wing: 'Floor 1', area: 'Left wing bathroom' },
    description: 'The drain is clogged.',
    photoId: 'p1',
    createdAt: '2026-11-12T04:20:00Z',
  },
}
const timeline = {
  entries: [
    { eventType: 'report.submitted', summary: 'Report submitted', occurredAt: '2026-11-12T04:20:00Z', actorId: ME },
    { eventType: 'report.routed', summary: 'Routed to DSA', occurredAt: '2026-11-12T04:20:01Z' },
    {
      eventType: 'status.changed',
      fromStatus: 'OPEN',
      toStatus: 'ASSIGNED',
      summary: 'Assigned to the DSA',
      occurredAt: '2026-11-12T04:20:02Z',
    },
    {
      eventType: 'status.changed',
      fromStatus: 'ASSIGNED',
      toStatus: 'IN_PROGRESS',
      summary: 'Plumber on the way',
      occurredAt: '2026-11-12T05:00:00Z',
      actorId: 'dsa-1',
    },
  ],
}

function stubGet(overrides: Record<string, unknown> = {}) {
  get.mockImplementation(async (path: string) => {
    if (path in overrides) {
      const v = overrides[path]
      if (v instanceof Error) throw v
      return v
    }
    if (path === `/reports/${ID}`) return report
    if (path === `/reports/${ID}/status`) return { status: { reportId: ID, status: 'IN_PROGRESS', version: 2 } }
    if (path === `/reports/${ID}/timeline`) return timeline
    throw new Error(`unexpected ${path}`)
  })
}

function renderAt(id = ID) {
  return render(
    <MemoryRouter initialEntries={[`/reports/${id}`]}>
      <Routes>
        <Route path="/reports/:reportId" element={<ReportDetail />} />
      </Routes>
    </MemoryRouter>,
  )
}

beforeEach(() => {
  get.mockReset()
  blob.mockReset()
  blob.mockResolvedValue(new Blob(['x'], { type: 'image/jpeg' }))
  // jsdom has no object URLs.
  URL.createObjectURL = vi.fn(() => 'blob:photo-1')
  URL.revokeObjectURL = vi.fn()
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('SCRUM-143 report detail', () => {
  it('shows the place, the current status and what was reported', async () => {
    stubGet()
    renderAt()

    expect(await screen.findByRole('heading', { level: 1, name: 'HC, Floor 1, Left wing bathroom' })).toBeInTheDocument()
    expect(screen.getAllByText('In progress').length).toBeGreaterThan(0)
    expect(screen.getByText('The drain is clogged.')).toBeInTheDocument()
    expect(screen.getByText('#R1')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /My reports/ })).toHaveAttribute('href', '/reports')
  })

  it('shows the timeline in order, oldest first', async () => {
    stubGet()
    renderAt()

    const history = await screen.findByRole('region', { name: 'What happened' })
    const items = await within(history).findAllByRole('listitem')
    expect(items.map((li) => li.textContent)).toEqual([
      expect.stringContaining('Open'),
      expect.stringContaining('Assigned'),
      expect.stringContaining('In progress'),
    ])
    expect(within(items[2]).getByText(/Plumber on the way/)).toBeInTheDocument()
  })

  it('shows the photo when there is one', async () => {
    stubGet()
    renderAt()

    const img = await screen.findByRole('img', { name: 'The photo you added' })
    expect(img).toHaveAttribute('src', 'blob:photo-1')
    expect(blob).toHaveBeenCalledWith('/photos/p1')
  })

  it('says there is no photo when none was added', async () => {
    stubGet({ [`/reports/${ID}`]: { report: { ...report.report, photoId: undefined } } })
    renderAt()

    expect(await screen.findByText('No photo added.')).toBeInTheDocument()
    expect(blob).not.toHaveBeenCalled()
  })

  it('shows a kind message, not the raw error, for a report that is not yours', async () => {
    stubGet({ [`/reports/${ID}`]: new ApiError(403, 'not allowed') })
    renderAt()

    expect(await screen.findByRole('heading', { name: 'Report not found' })).toBeInTheDocument()
    expect(screen.queryByText('not allowed')).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Back to My reports' })).toHaveAttribute('href', '/reports')
  })

  it('shows the same kind message for a report that does not exist', async () => {
    stubGet({ [`/reports/${ID}`]: new ApiError(404, 'report not found: r1') })
    renderAt()

    expect(await screen.findByRole('heading', { name: 'Report not found' })).toBeInTheDocument()
    expect(screen.queryByText(/report not found: r1/)).not.toBeInTheDocument()
  })

  it('still shows the report when only its status does not load', async () => {
    stubGet({ [`/reports/${ID}/status`]: new ApiError(404, 'unknown report') })
    renderAt()

    expect(await screen.findByText('The drain is clogged.')).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Report not found' })).not.toBeInTheDocument()
  })

  it('offers to try again when the service is busy', async () => {
    const user = userEvent.setup()
    let calls = 0
    stubGet()
    const base = get.getMockImplementation()!
    get.mockImplementation(async (path: string) => {
      if (path === `/reports/${ID}` && calls++ === 0) throw new ApiError(504, 'timeout')
      return base(path)
    })
    renderAt()

    expect(await screen.findByText(/service is busy or unreachable/i)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(await screen.findByText('The drain is clogged.')).toBeInTheDocument()
  })
})
