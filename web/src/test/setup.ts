// Adds matchers such as toBeInTheDocument() to Vitest's expect.
import '@testing-library/jest-dom/vitest'
import { afterEach } from 'vitest'

afterEach(() => {
  try {
    sessionStorage.clear()
    localStorage.clear() // the sign-in lives here since SCRUM-159
  } catch {
    // storage can be blocked; tests then run without it
  }
})
