import { act, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import OfflineBanner from './OfflineBanner'

function setOnline(value: boolean) {
  Object.defineProperty(window.navigator, 'onLine', { configurable: true, get: () => value })
}

afterEach(() => setOnline(true))

describe('SCRUM-98 offline banner', () => {
  it('shows nothing while online', () => {
    setOnline(true)
    render(<OfflineBanner />)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('says so when the connection drops, and goes away when it is back', () => {
    setOnline(true)
    render(<OfflineBanner />)
    act(() => {
      setOnline(false)
      window.dispatchEvent(new Event('offline'))
    })
    expect(screen.getByRole('status')).toHaveTextContent('You are offline. Showing saved pages.')
    act(() => {
      setOnline(true)
      window.dispatchEvent(new Event('online'))
    })
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('shows at once when the app opens without a connection', () => {
    setOnline(false)
    render(<OfflineBanner />)
    expect(screen.getByRole('status')).toHaveTextContent('You are offline.')
  })
})
