// Adds matchers such as toBeInTheDocument() to Vitest's expect.
import '@testing-library/jest-dom/vitest'
import { afterEach } from 'vitest'

afterEach(() => {
  try {
    sessionStorage.clear()
  } catch {
    // storage can be blocked; tests then run without it
  }
})
