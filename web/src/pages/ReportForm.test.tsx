import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../api/client'
import ReportForm, { MAX_DESCRIPTION, MAX_PHOTO_BYTES } from './ReportForm'

// The API client is replaced, so no gateway is needed. ApiError stays real.
vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>()
  return { ...actual, api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), upload: vi.fn() } }
})

const get = vi.mocked(api.get)
const post = vi.mocked(api.post)
const upload = vi.mocked(api.upload)

// Shapes as the gateway sends them (docs/design/api-routes.md, Sonam Choki's samples).
const categories = {
  categories: [
    { category: 'OTHERS', name: 'Others', sortOrder: 3, active: true },
    { category: 'HOSTEL', name: 'Hostel', sortOrder: 1, active: true },
    { category: 'ACADEMIC_BLOCK', name: 'Academic Block', sortOrder: 2, active: true },
    { category: 'OLD', name: 'Retired category', sortOrder: 4, active: false },
  ],
}

const locations: Record<string, unknown> = {
  HOSTEL: {
    locations: [
      { id: 'h1', category: 'HOSTEL', location: { building: 'Hostel A', wing: 'Wing 1', area: 'Bathroom 1' } },
      { id: 'h2', category: 'HOSTEL', location: { building: 'Hostel A', wing: 'Wing 1', area: 'Corridor' } },
      { id: 'h3', category: 'HOSTEL', location: { building: 'Hostel A', wing: 'Wing 2', area: 'Bathroom 2' } },
      { id: 'h4', category: 'HOSTEL', location: { building: 'Hostel B', wing: 'Wing 1', area: 'Kitchen' } },
    ],
  },
  OTHERS: {
    // OTHERS has no wing: building is "Campus" and the area is picked directly.
    locations: [
      { id: 'o1', category: 'OTHERS', location: { building: 'Campus', area: 'Wi-Fi' } },
      { id: 'o2', category: 'OTHERS', location: { building: 'Campus', area: 'Street lights' } },
    ],
  },
}

function stubGet() {
  get.mockImplementation(async (path: string) => {
    if (path === '/categories') return categories
    const m = path.match(/^\/categories\/([A-Z_]+)\/locations$/)
    if (m && locations[m[1]]) return locations[m[1]]
    throw new ApiError(404, 'not found')
  })
}

function renderForm() {
  return render(
    <MemoryRouter>
      <ReportForm />
    </MemoryRouter>,
  )
}

// fillValid picks Hostel A, Wing 1, Bathroom 1 and types a description.
async function fillValid(user: ReturnType<typeof userEvent.setup>) {
  await user.selectOptions(await screen.findByLabelText('Category'), 'HOSTEL')
  await user.selectOptions(await screen.findByLabelText('Building'), 'Hostel A')
  await user.selectOptions(screen.getByLabelText('Wing'), 'Wing 1')
  await user.selectOptions(screen.getByLabelText('Area'), 'Bathroom 1')
  await user.type(screen.getByLabelText('Description'), 'Tap is leaking')
}

function pngFile(name = 'tap.png', size = 1024, type = 'image/png') {
  const f = new File(['x'], name, { type })
  Object.defineProperty(f, 'size', { value: size })
  return f
}

beforeEach(() => {
  stubGet()
  // jsdom has no object URLs; the preview only needs a string.
  URL.createObjectURL = vi.fn(() => 'blob:preview')
  URL.revokeObjectURL = vi.fn()
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('SCRUM-65 category and location dropdowns', () => {
  it('shows the active categories from the API in their order', async () => {
    renderForm()
    const select = await screen.findByLabelText('Category')
    await waitFor(() => expect(within(select).getAllByRole('option')).toHaveLength(4))
    const labels = within(select).getAllByRole('option').map((o) => o.textContent)
    expect(labels).toEqual(['Choose a category', 'Hostel', 'Academic Block', 'Others'])
    expect(get).toHaveBeenCalledWith('/categories')
  })

  it('loads the locations of the chosen category and narrows building, wing and area', async () => {
    const user = userEvent.setup()
    renderForm()
    await user.selectOptions(await screen.findByLabelText('Category'), 'HOSTEL')
    expect(get).toHaveBeenCalledWith('/categories/HOSTEL/locations')

    const building = await screen.findByLabelText('Building')
    expect(within(building).getAllByRole('option').map((o) => o.textContent)).toEqual([
      'Choose a building',
      'Hostel A',
      'Hostel B',
    ])
    await user.selectOptions(building, 'Hostel A')
    const wing = screen.getByLabelText('Wing')
    expect(within(wing).getAllByRole('option').map((o) => o.textContent)).toEqual(['Choose a wing', 'Wing 1', 'Wing 2'])
    await user.selectOptions(wing, 'Wing 1')
    const area = screen.getByLabelText('Area')
    expect(within(area).getAllByRole('option').map((o) => o.textContent)).toEqual([
      'Choose an area',
      'Bathroom 1',
      'Corridor',
    ])
  })

  it('has no wing step when the locations have no wing (Others)', async () => {
    const user = userEvent.setup()
    renderForm()
    await user.selectOptions(await screen.findByLabelText('Category'), 'OTHERS')
    await user.selectOptions(await screen.findByLabelText('Building'), 'Campus')
    expect(screen.queryByLabelText('Wing')).not.toBeInTheDocument()
    expect(within(screen.getByLabelText('Area')).getAllByRole('option')).toHaveLength(3)
  })

  it('needs a category and a full location before sending', async () => {
    const user = userEvent.setup()
    renderForm()
    await screen.findByLabelText('Category')
    await user.type(screen.getByLabelText('Description'), 'Broken window')
    await user.click(screen.getByRole('button', { name: 'Send report' }))
    expect(await screen.findByText('Choose a category.')).toBeInTheDocument()
    expect(post).not.toHaveBeenCalled()

    await user.selectOptions(screen.getByLabelText('Category'), 'HOSTEL')
    await user.selectOptions(await screen.findByLabelText('Building'), 'Hostel A')
    await user.click(screen.getByRole('button', { name: 'Send report' }))
    expect(await screen.findByText('Choose the full location: building, wing and area.')).toBeInTheDocument()
    expect(post).not.toHaveBeenCalled()
  })

  it('shows a message when the categories cannot be loaded', async () => {
    get.mockRejectedValueOnce(new ApiError(503, 'category-taxonomy unavailable'))
    renderForm()
    expect(await screen.findByRole('alert')).toHaveTextContent('Could not load the categories. Please try again.')
  })
})

describe('SCRUM-66 description and photo field', () => {
  it('shows a live counter and stops at 500 characters', async () => {
    const user = userEvent.setup()
    renderForm()
    const description = await screen.findByLabelText('Description')
    expect(screen.getByText(`0 / ${MAX_DESCRIPTION}`)).toBeInTheDocument()
    await user.type(description, 'Leak')
    expect(screen.getByText('4 / 500')).toBeInTheDocument()

    fireEvent.change(description, { target: { value: 'a'.repeat(500) } })
    await user.type(description, 'b') // the 501st character is refused
    expect(description).toHaveValue('a'.repeat(500))
    expect(screen.getByText('500 / 500')).toBeInTheDocument()
  })

  it('needs a description that is not only spaces', async () => {
    const user = userEvent.setup()
    renderForm()
    await fillValid(user)
    await user.clear(screen.getByLabelText('Description'))
    await user.type(screen.getByLabelText('Description'), '   ')
    await user.click(screen.getByRole('button', { name: 'Send report' }))
    expect(await screen.findByText('Describe the problem.')).toBeInTheDocument()
    expect(post).not.toHaveBeenCalled()
  })

  it('refuses a photo that is not JPG or PNG', async () => {
    const user = userEvent.setup({ applyAccept: false })
    renderForm()
    await user.upload(await screen.findByLabelText('Photo (optional)'), pngFile('notes.pdf', 1024, 'application/pdf'))
    expect(screen.getByText('The photo must be a JPG or PNG image.')).toBeInTheDocument()
    expect(screen.queryByAltText('Photo preview')).not.toBeInTheDocument()
  })

  it('refuses a photo over 5 MB and accepts exactly 5 MB', async () => {
    const user = userEvent.setup()
    renderForm()
    const input = await screen.findByLabelText('Photo (optional)')
    await user.upload(input, pngFile('big.png', MAX_PHOTO_BYTES + 1))
    expect(screen.getByText('The photo must be 5 MB or smaller.')).toBeInTheDocument()

    await user.upload(input, pngFile('ok.jpg', MAX_PHOTO_BYTES, 'image/jpeg'))
    expect(screen.queryByText('The photo must be 5 MB or smaller.')).not.toBeInTheDocument()
    expect(screen.getByAltText('Photo preview')).toBeInTheDocument()
  })

  it('shows a preview and can remove the photo', async () => {
    const user = userEvent.setup()
    renderForm()
    await user.upload(await screen.findByLabelText('Photo (optional)'), pngFile())
    expect(screen.getByAltText('Photo preview')).toHaveAttribute('src', 'blob:preview')
    await user.click(screen.getByRole('button', { name: 'Remove photo' }))
    expect(screen.queryByAltText('Photo preview')).not.toBeInTheDocument()
  })
})

describe('SCRUM-67 submit to report-submission', () => {
  it('sends the report and shows the report id', async () => {
    post.mockResolvedValueOnce({ report: { id: 'r-123', status: 'OPEN' } })
    const user = userEvent.setup()
    renderForm()
    await fillValid(user)
    await user.click(screen.getByRole('button', { name: 'Send report' }))

    expect(await screen.findByRole('status')).toHaveTextContent('Report sent. Your report number is r-123.')
    expect(upload).not.toHaveBeenCalled()
    expect(post).toHaveBeenCalledWith('/reports', {
      category: 'HOSTEL',
      location: { building: 'Hostel A', wing: 'Wing 1', area: 'Bathroom 1' },
      description: 'Tap is leaking',
    })
  })

  it('uploads the photo first and sends its photoId with the report', async () => {
    upload.mockResolvedValueOnce({ photoId: 'p-9' })
    post.mockResolvedValueOnce({ report: { id: 'r-124' } })
    const user = userEvent.setup()
    renderForm()
    await fillValid(user)
    const photo = pngFile()
    await user.upload(screen.getByLabelText('Photo (optional)'), photo)
    await user.click(screen.getByRole('button', { name: 'Send report' }))

    await screen.findByRole('status')
    expect(upload).toHaveBeenCalledWith('/photos', photo, 'photo')
    expect(post).toHaveBeenCalledWith('/reports', expect.objectContaining({ photoId: 'p-9' }))
    expect(upload.mock.invocationCallOrder[0]).toBeLessThan(post.mock.invocationCallOrder[0])
  })

  it('shows the reason for a 400 and keeps the form', async () => {
    post.mockRejectedValueOnce(new ApiError(400, 'description must be at most 500 characters'))
    const user = userEvent.setup()
    renderForm()
    await fillValid(user)
    await user.click(screen.getByRole('button', { name: 'Send report' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('description must be at most 500 characters')
    expect(screen.getByLabelText('Description')).toHaveValue('Tap is leaking')
  })

  it('asks to try again when a service is unavailable or the server cannot be reached', async () => {
    post.mockRejectedValueOnce(new ApiError(503, 'report-submission unavailable'))
    post.mockRejectedValueOnce(new ApiError(0, 'Cannot reach the server.'))
    const user = userEvent.setup()
    renderForm()
    await fillValid(user)
    await user.click(screen.getByRole('button', { name: 'Send report' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Please try again.')
    await user.click(screen.getByRole('button', { name: 'Send report' }))
    await waitFor(() => expect(post).toHaveBeenCalledTimes(2))
    expect(screen.getByRole('alert')).toHaveTextContent('Please try again.')
  })

  it('does not send the report when the photo upload fails', async () => {
    upload.mockRejectedValueOnce(new ApiError(400, 'invalid photo: file bytes do not match the content type'))
    const user = userEvent.setup()
    renderForm()
    await fillValid(user)
    await user.upload(screen.getByLabelText('Photo (optional)'), pngFile())
    await user.click(screen.getByRole('button', { name: 'Send report' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('invalid photo: file bytes do not match the content type')
    expect(post).not.toHaveBeenCalled()
  })

  it('disables the button while sending, so a report is sent only once', async () => {
    let finish: (v: unknown) => void = () => {}
    post.mockReturnValueOnce(new Promise((resolve) => (finish = resolve)))
    const user = userEvent.setup()
    renderForm()
    await fillValid(user)
    await user.click(screen.getByRole('button', { name: 'Send report' }))

    const busy = screen.getByRole('button', { name: 'Sending...' })
    expect(busy).toBeDisabled()
    await user.click(busy)
    expect(post).toHaveBeenCalledTimes(1)

    await act(async () => finish({ report: { id: 'r-1' } }))
    expect(await screen.findByRole('status')).toHaveTextContent('r-1')
  })
})
