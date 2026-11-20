import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../api/client'
import StaffView from './StaffView'

// The API client is replaced, so no gateway is needed. ApiError stays real.
vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>()
  return { ...actual, api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), upload: vi.fn() } }
})

// The signed-in user: the DSA, unless a test signs in someone else.
const ME = '00000000-0000-4000-8000-000000000001'
const DSA = { id: ME, email: 'dsa.demo@jurwa.example', fullName: 'Dorji DSA', roles: ['CATEGORY_HEAD'] }
const auth = vi.hoisted(() => ({ user: null as unknown }))
vi.mock('../auth/useAuth', () => ({
  useAuth: () => ({ user: auth.user, login: vi.fn(), logout: vi.fn() }),
}))

const get = vi.mocked(api.get)
const post = vi.mocked(api.post)

// Shapes as the gateway sends them (protojson of report-history, status-lifecycle
// and internal-notes; see services/api-gateway/README.md).
function summary(reportId: string, status: string, building: string, area: string) {
  return {
    reportId,
    reporterId: 'student-1',
    category: 'HOSTEL',
    location: { building, wing: 'Wing 1', area },
    status,
    currentAssigneeId: ME,
    createdAt: '2026-11-08T08:00:00Z',
    updatedAt: '2026-11-08T09:00:00Z',
  }
}

const assigned = {
  reports: [
    summary('r1', 'ASSIGNED', 'Hostel A', 'Bathroom 1'),
    summary('r2', 'IN_PROGRESS', 'Hostel B', 'Kitchen'),
    summary('r3', 'ESCALATED', 'Hostel A', 'Corridor'),
    summary('r4', 'RESOLVED', 'Hostel C', 'Room 12'),
  ],
  nextPageToken: '',
}

let statuses: Record<string, { status: string; version: number }>

// GET /positions (Zangmo): the people a report can be reassigned to.
const positions = {
  positions: [
    { code: 'DSA', title: 'Dean of Student Affairs', kind: 'CATEGORY_HEAD', holderUserId: ME },
    { code: 'DAA', title: 'Dean of Academic Affairs', kind: 'CATEGORY_HEAD', holderUserId: 'head-daa' },
    { code: 'SSO', title: 'Student Service Officer', kind: 'CATEGORY_HEAD', holderUserId: 'head-sso' },
    { code: 'BLOCK_COUNCILLOR_HA', title: 'Block Councillor (HA)', kind: 'COUNCILLOR', holderUserId: 'staff-8' },
    { code: 'GAMES_HEALTH_COUNCILLOR', title: 'Games and Health Councillor', kind: 'COUNCILLOR' },
    { code: 'MAINTENANCE_COUNCILLOR', title: 'Electrical Maintenance Councillor', kind: 'COUNCILLOR', holderUserId: 'staff-7' },
    { code: 'HOD_CT', title: 'Head of Department, Computing Technologies', kind: 'HOD', holderUserId: 'hod-ct' },
    { code: 'SSO_MALE', title: 'Student Service Officer (male)', kind: 'STAFF', holderUserId: 'staff-10' },
  ],
}

function stubGet(overrides: Record<string, unknown> = {}) {
  get.mockImplementation(async (path: string) => {
    if (path in overrides) {
      const v = overrides[path]
      if (v instanceof Error) throw v
      return v
    }
    if (path === '/me/assigned-reports') return assigned
    if (path === '/positions') return positions
    let m = path.match(/^\/reports\/(r\d)$/)
    if (m) {
      return {
        report: {
          id: m[1],
          category: 'HOSTEL',
          location: { building: 'Hostel A', wing: 'Wing 1', area: 'Bathroom 1' },
          description: 'The tap has been leaking since Monday.',
          createdAt: '2026-11-08T08:00:00Z',
        },
      }
    }
    m = path.match(/^\/reports\/(r\d)\/status$/)
    if (m) return { status: { reportId: m[1], ...statuses[m[1]] } }
    m = path.match(/^\/reports\/(r\d)\/timeline$/)
    if (m) {
      return {
        entries: [
          { id: 'e1', eventType: 'report.submitted', summary: 'Report submitted', occurredAt: '2026-11-08T08:00:00Z' },
          { id: 'e2', eventType: 'report.routed', summary: 'Assigned to DSA', occurredAt: '2026-11-08T08:00:01Z' },
          {
            id: 'e3',
            eventType: 'status.changed',
            fromStatus: 'OPEN',
            toStatus: 'ASSIGNED',
            summary: 'Routed to the Category Head',
            occurredAt: '2026-11-08T08:00:02Z',
          },
        ],
      }
    }
    m = path.match(/^\/reports\/(r\d)\/notes$/)
    if (m) {
      return {
        notes: [{ id: 'n1', reportId: m[1], authorId: 'staff-9', body: 'Plumber called.', createdAt: '2026-11-08T10:00:00Z' }],
      }
    }
    throw new ApiError(404, 'not found')
  })
}

function renderView() {
  return render(
    <MemoryRouter>
      <StaffView />
    </MemoryRouter>,
  )
}

type User = ReturnType<typeof userEvent.setup>

// openReport opens the card of a report by its place.
async function openReport(user: User, place: RegExp) {
  await user.click(await screen.findByRole('button', { name: place }))
  return screen.findByRole('heading', { level: 2, name: place })
}

beforeEach(() => {
  auth.user = DSA
  statuses = {
    r1: { status: 'ASSIGNED', version: 2 },
    r2: { status: 'IN_PROGRESS', version: 3 },
    r3: { status: 'ESCALATED', version: 3 },
    r4: { status: 'RESOLVED', version: 4 },
  }
  stubGet()
})

afterEach(() => {
  vi.clearAllMocks()
})

// SCRUM-158: the board is headed with who is signed in, not "Staff view".
describe('SCRUM-158 board heading', () => {
  it('shows the name of the signed-in person', async () => {
    auth.user = { id: 'staff-8', email: 'councillor.ha@jurwa.example', fullName: 'Block Councillor (HA)', roles: ['STAFF'] }
    renderView()
    expect(await screen.findByRole('heading', { level: 1, name: 'Block Councillor (HA)' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Staff view' })).not.toBeInTheDocument()
  })
})

describe('SCRUM-36 assigned reports', () => {
  it('lists the reports assigned to me with their place and status', async () => {
    renderView()
    expect(screen.getByRole('status', { name: 'Loading reports' })).toBeInTheDocument()
    const card = await screen.findByRole('button', { name: /Hostel A, Wing 1, Bathroom 1/ })
    expect(card).toHaveTextContent('Assigned')
    expect(screen.getAllByRole('listitem')).toHaveLength(4)
    expect(get).toHaveBeenCalledWith('/me/assigned-reports')
  })

  it('counts the reports per status on the chips and filters by one', async () => {
    const user = userEvent.setup()
    renderView()
    await screen.findByRole('button', { name: /Bathroom 1/ })
    expect(screen.getByRole('button', { name: 'All 4' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Escalated 1' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Escalated 1' }))
    expect(screen.getByRole('button', { name: 'Escalated 1' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getAllByRole('listitem')).toHaveLength(1)
    expect(screen.getByRole('button', { name: /Corridor/ })).toBeInTheDocument()
  })

  it('says so when nothing is assigned', async () => {
    stubGet({ '/me/assigned-reports': { reports: [], nextPageToken: '' } })
    renderView()
    expect(await screen.findByRole('heading', { name: 'Nothing assigned to you' })).toBeInTheDocument()
  })

  it('offers to try again when the list does not load', async () => {
    stubGet({ '/me/assigned-reports': new ApiError(503, 'report-history unavailable') })
    const user = userEvent.setup()
    renderView()
    expect(await screen.findByRole('alert')).toHaveTextContent('The service is busy or unreachable. Please try again.')
    stubGet()
    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(await screen.findByRole('button', { name: /Bathroom 1/ })).toBeInTheDocument()
  })

  it('loads the next page when there are more reports', async () => {
    stubGet({
      '/me/assigned-reports': { reports: [summary('r1', 'ASSIGNED', 'Hostel A', 'Bathroom 1')], nextPageToken: 't2' },
      '/me/assigned-reports?pageToken=t2': { reports: [summary('r2', 'IN_PROGRESS', 'Hostel B', 'Kitchen')], nextPageToken: '' },
    })
    const user = userEvent.setup()
    renderView()
    await user.click(await screen.findByRole('button', { name: 'Load more' }))
    expect(await screen.findByRole('button', { name: /Kitchen/ })).toBeInTheDocument()
    expect(screen.getAllByRole('listitem')).toHaveLength(2)
    expect(screen.queryByRole('button', { name: 'Load more' })).not.toBeInTheDocument()
  })
})

describe('SCRUM-36 one report: details, history and notes', () => {
  it('opens a report on the Details tab with its description', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Hostel A, Wing 1, Bathroom 1/)
    expect(screen.getByRole('tab', { name: 'Details' })).toHaveAttribute('aria-selected', 'true')
    expect(await screen.findByText('The tap has been leaking since Monday.')).toBeInTheDocument()
    expect(get).toHaveBeenCalledWith('/reports/r1')
  })

  it('moves between the tabs with the arrow keys', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    screen.getByRole('tab', { name: 'Details' }).focus()
    await user.keyboard('{ArrowRight}')
    expect(screen.getByRole('tab', { name: 'History' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('tab', { name: 'History' })).toHaveFocus()
  })

  it('shows the history of status changes, oldest first', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(screen.getByRole('tab', { name: 'History' }))
    const panel = screen.getByRole('tabpanel', { name: 'History' })
    await waitFor(() => expect(within(panel).getAllByRole('listitem')).toHaveLength(2))
    const items = within(panel).getAllByRole('listitem')
    expect(items[0]).toHaveTextContent('Open')
    expect(items[1]).toHaveTextContent('Assigned')
    expect(items[1]).toHaveTextContent('Routed to the Category Head')
  })

  it('lists the notes and adds one', async () => {
    post.mockResolvedValueOnce({
      note: { id: 'n2', reportId: 'r1', authorId: ME, body: 'Spare part ordered.', createdAt: '2026-11-08T11:00:00Z' },
    })
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(screen.getByRole('tab', { name: 'Notes' }))
    expect(await screen.findByText('Plumber called.')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Add note' }))
    expect(await screen.findByText('Write the note first.')).toBeInTheDocument()
    expect(post).not.toHaveBeenCalled()

    await user.type(screen.getByLabelText('New note'), 'Spare part ordered.')
    await user.click(screen.getByRole('button', { name: 'Add note' }))
    expect(await screen.findByText('Spare part ordered.')).toBeInTheDocument()
    expect(post).toHaveBeenCalledWith('/reports/r1/notes', { body: 'Spare part ordered.' })
    expect(screen.getByLabelText('New note')).toHaveValue('')
  })

  it('goes back to the list', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(screen.getByRole('button', { name: 'Back to the list' }))
    expect(screen.getByRole('button', { name: 'All 4' })).toBeInTheDocument()
  })
})

describe('SCRUM-36 update status', () => {
  it('offers only the next allowed status and sends it with the version', async () => {
    post.mockResolvedValueOnce({ status: { reportId: 'r1', status: 'IN_PROGRESS', version: 3 } })
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Update status' }))

    const sheet = screen.getByRole('dialog', { name: 'Update status' })
    expect(within(sheet).getAllByRole('radio').map((r) => r.closest('label')?.textContent)).toEqual(['In progress'])
    await user.click(within(sheet).getByRole('radio', { name: 'In progress' }))
    await user.type(within(sheet).getByLabelText('Note for the history (optional)'), 'Plumber on the way')
    await user.click(within(sheet).getByRole('button', { name: 'Save status' }))

    expect(post).toHaveBeenCalledWith('/reports/r1/status', {
      toStatus: 'IN_PROGRESS',
      expectedVersion: 2,
      reason: 'Plumber on the way',
    })
    expect(await screen.findByText('Status updated')).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: /Bathroom 1/ }).closest('section')).toHaveTextContent('In progress')
  })

  it('keeps every letter typed in the sheet', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Update status' }))
    const reason = screen.getByLabelText('Note for the history (optional)')
    await user.type(reason, 'abc')
    expect(reason).toHaveValue('abc')
    expect(reason).toHaveFocus()
  })

  it('says when someone else changed the report first, and shows the new status', async () => {
    post.mockRejectedValueOnce(new ApiError(409, 'version conflict'))
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Update status' }))
    statuses.r1 = { status: 'IN_PROGRESS', version: 3 }
    await user.click(screen.getByRole('radio', { name: 'In progress' }))
    await user.click(screen.getByRole('button', { name: 'Save status' }))
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Someone changed this report first. It now shows the latest status.',
    )
    await waitFor(() => expect(get).toHaveBeenCalledTimes(4)) // list, report, status, status again
  })

  it('needs a status to be chosen', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Update status' }))
    await user.click(screen.getByRole('button', { name: 'Save status' }))
    expect(await screen.findByText('Choose the new status.')).toBeInTheDocument()
    expect(post).not.toHaveBeenCalled()
  })

  it('has nothing to update on a resolved report', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Room 12/)
    expect(await screen.findByText('Resolved. Nothing more to do.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Update status' })).not.toBeInTheDocument()
  })
})

describe('SCRUM-36 reassign', () => {
  async function openSheet(user: User) {
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Reassign' }))
    return screen.getByRole('dialog', { name: 'Reassign' })
  }

  // SCRUM-158: the SSO (and the DSA after an escalation) hand reports to the
  // councillors who have someone in the post.
  it('offers the councillors that have someone in them', async () => {
    const user = userEvent.setup()
    renderView()
    const sheet = await openSheet(user)
    const to = await within(sheet).findByLabelText('Reassign to')
    await waitFor(() => expect(within(to).getAllByRole('option')).toHaveLength(3))
    expect(within(to).getAllByRole('option').map((o) => o.textContent)).toEqual([
      'Choose a person',
      'Block Councillor (HA)',
      'Electrical Maintenance Councillor',
    ])
    expect(get).toHaveBeenCalledWith('/positions')
  })

  // SCRUM-158: an escalated Academic Block report reaches the DAA, who hands it
  // back to an HoD.
  it('offers the DAA the HoDs', async () => {
    auth.user = { id: 'head-daa', email: 'daa.demo@jurwa.example', fullName: 'DAA (demo)', roles: ['CATEGORY_HEAD'] }
    const user = userEvent.setup()
    renderView()
    const sheet = await openSheet(user)
    const to = await within(sheet).findByLabelText('Reassign to')
    await waitFor(() => expect(within(to).getAllByRole('option')).toHaveLength(2))
    expect(within(to).getAllByRole('option').map((o) => o.textContent)).toEqual([
      'Choose a person',
      'Head of Department, Computing Technologies',
    ])
  })

  it('sends the chosen holder and the reason, then the report leaves my list', async () => {
    post.mockResolvedValueOnce({ assignment: { reportId: 'r1', assigneeUserId: 'staff-7', kind: 'REASSIGNED', current: true } })
    const user = userEvent.setup()
    renderView()
    const sheet = await openSheet(user)
    await user.selectOptions(await within(sheet).findByLabelText('Reassign to'), 'staff-7')
    await user.type(within(sheet).getByLabelText('Why (the new person sees this)'), 'Light in the corridor')
    await user.click(within(sheet).getByRole('button', { name: 'Reassign report' }))

    expect(post).toHaveBeenCalledWith('/reports/r1/reassign', { toUserId: 'staff-7', reason: 'Light in the corridor' })
    expect(await screen.findByText('Reassigned to Electrical Maintenance Councillor')).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Bathroom 1/ })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'All 3' })).toBeInTheDocument()
  })

  it('needs a person and a reason', async () => {
    const user = userEvent.setup()
    renderView()
    const sheet = await openSheet(user)
    await within(sheet).findByRole('option', { name: 'Block Councillor (HA)' })
    await user.click(within(sheet).getByRole('button', { name: 'Reassign report' }))
    expect(await within(sheet).findByText('Choose who to reassign it to.')).toBeInTheDocument()
    await user.selectOptions(within(sheet).getByLabelText('Reassign to'), 'staff-8')
    await user.type(within(sheet).getByLabelText('Why (the new person sees this)'), '   ')
    await user.click(within(sheet).getByRole('button', { name: 'Reassign report' }))
    expect(await within(sheet).findByText('Write why you are reassigning it.')).toBeInTheDocument()
    expect(post).not.toHaveBeenCalled()
  })

  it('explains a 403 and a 409', async () => {
    post.mockRejectedValueOnce(new ApiError(403, 'not allowed'))
    post.mockRejectedValueOnce(new ApiError(409, 'already reassigned'))
    const user = userEvent.setup()
    renderView()
    const sheet = await openSheet(user)
    await user.selectOptions(await within(sheet).findByLabelText('Reassign to'), 'staff-8')
    await user.type(within(sheet).getByLabelText('Why (the new person sees this)'), 'Hostel A')
    await user.click(within(sheet).getByRole('button', { name: 'Reassign report' }))
    expect(await within(sheet).findByRole('alert')).toHaveTextContent('Only a Category Head can reassign a report.')
    await user.click(within(sheet).getByRole('button', { name: 'Reassign report' }))
    expect(await within(sheet).findByRole('alert')).toHaveTextContent(
      'Someone reassigned this report first. Go back to the list to see it.',
    )
  })

  // SCRUM-158: HoDs and councillors fix what they get; they do not pass it on.
  it('gives staff no reassign, only the status', async () => {
    auth.user = { id: 'staff-8', email: 'councillor.ha@jurwa.example', fullName: 'Block Councillor (HA)', roles: ['STAFF'] }
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    expect(await screen.findByRole('button', { name: 'Update status' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Reassign' })).not.toBeInTheDocument()
  })

  it('has no reassign on a resolved report', async () => {
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Room 12/)
    await screen.findByText('Resolved. Nothing more to do.')
    expect(screen.queryByRole('button', { name: 'Reassign' })).not.toBeInTheDocument()
  })
})

// SCRUM-164: what the SSO writes when passing a report on reaches the
// councillor as a note, and the student's stars show on a resolved report.
describe('SCRUM-164 the reassign note and the stars', () => {
  it('saves the reason as a note the councillor can read, then reassigns', async () => {
    const calls: string[] = []
    post.mockImplementation(async (path: string) => {
      calls.push(path)
      if (path.endsWith('/notes')) return { note: { id: 'n9', reportId: 'r1', authorId: ME, body: 'x', createdAt: '2026-11-21T09:00:00Z' } }
      return { assignment: { reportId: 'r1', assigneeUserId: 'staff-7', kind: 'REASSIGNED', current: true } }
    })
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Bathroom 1/)
    await user.click(await screen.findByRole('button', { name: 'Reassign' }))
    const sheet = screen.getByRole('dialog', { name: 'Reassign' })
    await user.selectOptions(await within(sheet).findByLabelText('Reassign to'), 'staff-7')
    await user.type(within(sheet).getByLabelText('Why (the new person sees this)'), 'Light in the corridor')
    await user.click(within(sheet).getByRole('button', { name: 'Reassign report' }))

    expect(await screen.findByText('Reassigned to Electrical Maintenance Councillor')).toBeInTheDocument()
    expect(post).toHaveBeenCalledWith('/reports/r1/notes', {
      body: 'Passed to Electrical Maintenance Councillor: Light in the corridor',
    })
    expect(calls).toEqual(['/reports/r1/notes', '/reports/r1/reassign'])
  })

  it('shows the stars the student gave on a resolved report', async () => {
    stubGet({ '/reports/r4/feedback': { feedback: { reportId: 'r4', rating: 4, comment: 'Quick fix' } } })
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Room 12/)
    expect(await screen.findByText('Rated 4 out of 5 by the student')).toBeInTheDocument()
    expect(screen.getByText('Quick fix')).toBeInTheDocument()
  })

  it('says when a resolved report is not rated yet', async () => {
    stubGet({ '/reports/r4/feedback': new ApiError(404, 'no feedback') })
    const user = userEvent.setup()
    renderView()
    await openReport(user, /Room 12/)
    expect(await screen.findByText('Not rated by the student yet.')).toBeInTheDocument()
  })
})
