import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from '../App'

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn())
})

afterEach(() => {
  vi.unstubAllGlobals()
})

// SCRUM-150: the home page describes Others with the same text as category-taxonomy.
describe('home page categories', () => {
  it('describes Others as campus places, without Wi-Fi', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByText('Campus places such as the library, mess, gym and sports grounds.')).toBeInTheDocument()
    expect(screen.queryByText(/Wi-Fi, roads and grounds/)).not.toBeInTheDocument()
  })
})
