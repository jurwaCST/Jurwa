// Tests for the shared components other pages build on (written before the
// components). They check what a user and a screen reader get, not markup.
import { act, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import BottomSheet from './BottomSheet'
import Button from './Button'
import ErrorState from './ErrorState'
import Field from './Field'
import Skeleton from './Skeleton'
import TextArea from './TextArea'
import Timeline from './Timeline'
import Toast from './Toast'

describe('Field', () => {
  it('labels its control and links the hint and the error to it', () => {
    render(
      <Field label="Building" hint="Pick from the list" error="Choose a building">
        <select>
          <option>Hostel A</option>
        </select>
      </Field>,
    )
    const select = screen.getByLabelText('Building')
    expect(select).toHaveAttribute('aria-invalid', 'true')
    expect(select).toHaveAccessibleDescription('Pick from the list Choose a building')
    expect(screen.getByRole('alert')).toHaveTextContent('Choose a building')
  })

  it('without an error the control is valid and nothing is announced', () => {
    render(
      <Field label="Wing">
        <input />
      </Field>,
    )
    expect(screen.getByLabelText('Wing')).not.toHaveAttribute('aria-invalid')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})

describe('TextArea', () => {
  it('counts characters against the limit', async () => {
    function Wrapper() {
      return <TextArea label="Description" max={500} value={'x'.repeat(12)} onChange={() => {}} />
    }
    render(<Wrapper />)
    expect(screen.getByLabelText('Description')).toBeInTheDocument()
    expect(screen.getByText('12 / 500')).toBeInTheDocument()
  })

  it('marks the count as over the limit and the field invalid', () => {
    render(<TextArea label="Description" max={500} value={'x'.repeat(501)} onChange={() => {}} />)
    expect(screen.getByText('501 / 500')).toHaveClass('over')
    expect(screen.getByLabelText('Description')).toHaveAttribute('aria-invalid', 'true')
  })

  it('counts Dzongkha letters as characters, not bytes', () => {
    render(<TextArea label="Description" max={500} value={'ཀ'.repeat(10)} onChange={() => {}} />)
    expect(screen.getByText('10 / 500')).toBeInTheDocument()
  })

  it('passes each change up', async () => {
    const onChange = vi.fn()
    render(<TextArea label="Description" max={500} value="" onChange={onChange} />)
    await userEvent.type(screen.getByLabelText('Description'), 'a')
    expect(onChange).toHaveBeenCalledWith('a')
  })
})

describe('Button', () => {
  it('while loading it is busy, disabled and says so', () => {
    render(
      <Button loading loadingText="Sending...">
        Send report
      </Button>,
    )
    const button = screen.getByRole('button', { name: 'Sending...' })
    expect(button).toBeDisabled()
    expect(button).toHaveAttribute('aria-busy', 'true')
  })

  it('applies the variant', () => {
    render(<Button variant="danger">Delete rule</Button>)
    expect(screen.getByRole('button', { name: 'Delete rule' })).toHaveClass('btn-danger')
  })
})

describe('Timeline', () => {
  it('lists the status changes in the order given, with the status word and time', () => {
    render(
      <Timeline
        entries={[
          { status: 'OPEN', at: '2026-11-08T08:00:00Z' },
          { status: 'ASSIGNED', at: '2026-11-08T08:00:05Z', by: 'DSA' },
          { status: 'RESOLVED', at: '2026-11-08T12:30:00Z', by: 'Karma Dorji', note: 'Tap replaced' },
        ]}
      />,
    )
    const items = screen.getAllByRole('listitem')
    expect(items).toHaveLength(3)
    expect(items[0]).toHaveTextContent('Open')
    expect(items[1]).toHaveTextContent('Assigned')
    expect(items[1]).toHaveTextContent('DSA')
    expect(items[2]).toHaveTextContent('Resolved')
    expect(items[2]).toHaveTextContent('Tap replaced')
    expect(items[2].querySelector('time')).toHaveAttribute('datetime', '2026-11-08T12:30:00Z')
  })

  it('says so when there is nothing yet', () => {
    render(<Timeline entries={[]} />)
    expect(screen.getByText('No changes yet.')).toBeInTheDocument()
  })
})

describe('BottomSheet', () => {
  it('is a labelled dialog that closes with Escape and with its close button', async () => {
    const onClose = vi.fn()
    render(
      <BottomSheet open title="Update status" onClose={onClose}>
        <p>Pick the new status.</p>
      </BottomSheet>,
    )
    expect(screen.getByRole('dialog', { name: 'Update status' })).toHaveAttribute('aria-modal', 'true')

    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(1)

    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(onClose).toHaveBeenCalledTimes(2)
  })

  it('renders nothing while closed', () => {
    render(
      <BottomSheet open={false} title="Update status" onClose={() => {}}>
        <p>Pick the new status.</p>
      </BottomSheet>,
    )
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('moves focus into the sheet when it opens', () => {
    render(
      <BottomSheet open title="Reassign" onClose={() => {}}>
        <button type="button">Choose a person</button>
      </BottomSheet>,
    )
    expect(screen.getByRole('button', { name: 'Close' })).toHaveFocus()
  })
})

describe('Toast', () => {
  it('announces the message politely and goes away by itself', () => {
    vi.useFakeTimers()
    const onDone = vi.fn()
    render(<Toast message="Status updated" onDone={onDone} />)
    expect(screen.getByRole('status')).toHaveTextContent('Status updated')
    act(() => {
      vi.advanceTimersByTime(4000)
    })
    expect(onDone).toHaveBeenCalledTimes(1)
    vi.useRealTimers()
  })
})

describe('ErrorState', () => {
  it('says what went wrong and retries on request', async () => {
    const onRetry = vi.fn()
    render(<ErrorState message="Cannot reach the server." onRetry={onRetry} />)
    expect(screen.getByRole('alert')).toHaveTextContent('Cannot reach the server.')
    await userEvent.click(screen.getByRole('button', { name: 'Try again' }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})

describe('Skeleton', () => {
  it('is hidden from screen readers and tells them the area is loading', () => {
    render(<Skeleton lines={3} label="Loading reports" />)
    const region = screen.getByRole('status', { name: 'Loading reports' })
    expect(region).toHaveAttribute('aria-busy', 'true')
    expect(region.querySelectorAll('.skeleton-line')).toHaveLength(3)
  })
})
