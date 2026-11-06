import { createContext, useContext } from 'react'
import type { User } from './types'

export interface AuthValue {
  /** The signed-in user, or null. */
  user: User | null
  /** login calls POST /auth/login and keeps the session. Throws ApiError on failure. */
  login: (email: string, password: string) => Promise<User>
  logout: () => void
}

export const AuthContext = createContext<AuthValue | null>(null)

/** useAuth gives a page the signed-in user, login and logout. */
export function useAuth(): AuthValue {
  const value = useContext(AuthContext)
  if (!value) {
    throw new Error('useAuth must be used inside <AuthProvider>')
  }
  return value
}
