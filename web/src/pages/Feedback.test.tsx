import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../api/client'
import Feedback from './Feedback'

// The API client is replaced, so no gateway is needed. ApiError stays real.
vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>()
  return { ...actual, api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), upload: vi.fn(), blob: vi.fn() } }
})

const get = vi.mocked(api.get)
const post = vi.mocked(api.post)

const ID = 'r1'
// As the gateway sends it (feedback service, SCRUM-147).
const saved = {
  feedback: {
    id: 'f1',
    reportId: ID,
    studentUserId: 'student-1',
    rating: 4,
    comment: 'Fixed the same day, thank you.',
    createdAt: '2026-11-12T08:00:00Z',
  },
}
const noFeedback = new ApiError(404, 'this report has no feedback')

beforeEach(() => {
  get.mockReset()
  post.mockReset()
})

async function rate(stars: string, comment = '') {
  const user = userEvent.setup()
  render(<Feedback reportId={ID} />)
  const group = await screen.findByRole('radiogroup', { name: /how was the fix/i })
  if (stars) await user.click(within(group).getByRole('radio', { name: stars }))
  if (comment) await user.type(screen.getByLabelText(/comment/i), comment)
  await user.click(screen.getByRole('button', { name: 'Send feedback' }))
  return user
}

describe('SCRUM-145 feedback stars', () => {
  it('offers five stars and a comment when the report has no feedback yet', async () => {
    get.mockRejectedValue(noFeedback)
    render(<Feedback reportId={ID} />)

    const group = await screen.findByRole('radiogroup', { name: /how was the fix/i })
    expect(within(group).getAllByRole('radio')).toHaveLength(5)
    expect(within(group).getByRole('radio', { name: '1 star' })).toBeInTheDocument()
    expect(within(group).getByRole('radio', { name: '5 stars' })).toBeInTheDocument()
    expect(screen.getByLabelText(/comment/i)).toBeInTheDocument()
    expect(get).toHaveBeenCalledWith(`/reports/${ID}/feedback`)
  })

  it('sends the rating and comment, then shows what was saved', async () => {
    get.mockRejectedValue(noFeedback)
    post.mockResolvedValue(saved)

    await rate('4 stars', 'Fixed the same day, thank you.')

    expect(post).toHaveBeenCalledWith(`/reports/${ID}/feedback`, {
      rating: 4,
      comment: 'Fixed the same day, thank you.',
    })
    expect(await screen.findByText('You rated this 4 out of 5 stars.')).toBeInTheDocument()
    expect(screen.getByText('Fixed the same day, thank you.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Send feedback' })).not.toBeInTheDocument()
  })

  it('shows the saved rating, not the form, when the student already rated it', async () => {
    get.mockResolvedValue(saved)
    render(<Feedback reportId={ID} />)

    expect(await screen.findByText('You rated this 4 out of 5 stars.')).toBeInTheDocument()
    expect(screen.getByText('Fixed the same day, thank you.')).toBeInTheDocument()
    expect(screen.queryByRole('radiogroup')).not.toBeInTheDocument()
  })

  it('asks for stars before sending', async () => {
    get.mockRejectedValue(noFeedback)

    await rate('')

    expect(await screen.findByRole('alert')).toHaveTextContent('Choose from 1 to 5 stars.')
    expect(post).not.toHaveBeenCalled()
  })

  it('shows the saved rating when the report was already rated elsewhere', async () => {
    get.mockRejectedValueOnce(noFeedback).mockResolvedValueOnce(saved)
    post.mockRejectedValue(new ApiError(409, 'this report already has feedback'))

    await rate('2 stars')

    expect(await screen.findByText('You rated this 4 out of 5 stars.')).toBeInTheDocument()
    expect(screen.queryByText(/already has feedback/)).not.toBeInTheDocument()
  })

  it('says plainly when the report is no longer resolved', async () => {
    get.mockRejectedValue(noFeedback)
    post.mockRejectedValue(new ApiError(422, 'feedback is allowed only after the report is RESOLVED'))

    await rate('5 stars')

    expect(await screen.findByRole('alert')).toHaveTextContent('You can rate a report once it is resolved.')
  })

  it('keeps the form and the choice when the service is busy', async () => {
    get.mockRejectedValue(noFeedback)
    post.mockRejectedValue(new ApiError(504, 'timeout'))

    await rate('3 stars', 'ok')

    expect(await screen.findByRole('alert')).toHaveTextContent(/service is busy or unreachable/i)
    expect(screen.getByRole('radio', { name: '3 stars' })).toBeChecked()
    expect(screen.getByRole('button', { name: 'Send feedback' })).toBeEnabled()
  })
})
