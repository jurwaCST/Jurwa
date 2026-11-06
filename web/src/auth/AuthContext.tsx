import { useCallback, useMemo, useState, type ReactNode } from 'react'
import { api, setToken } from '../api/client'
import type { LoginResponse, Session, User } from './types'
import { AuthContext, type AuthValue } from './useAuth'

/** sessionStorage key. The session ends when the browser tab is closed. */
export const SESSION_KEY = 'jurwa.session'

// loadSession reads a saved, unexpired session. Storage can be blocked
// (private mode, strict settings), so every access is guarded.
function loadSession(): Session | null {
  try {
    const raw = sessionStorage.getItem(SESSION_KEY)
    if (!raw) return null
    const s = JSON.parse(raw) as Session
    if (!s.token || !s.user || Date.parse(s.expiresAt) <= Date.now()) {
      sessionStorage.removeItem(SESSION_KEY)
      return null
    }
    return s
  } catch {
    return null
  }
}

function saveSession(s: Session | null): void {
  try {
    if (s) sessionStorage.setItem(SESSION_KEY, JSON.stringify(s))
    else sessionStorage.removeItem(SESSION_KEY)
  } catch {
    // the session then lives in memory only
  }
}

/** AuthProvider keeps the signed-in user for the whole app. */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(() => {
    const s = loadSession()
    setToken(s?.token ?? null)
    return s?.user ?? null
  })

  const login = useCallback(async (email: string, password: string) => {
    const res = await api.post<LoginResponse>('/auth/login', { email, password })
    setToken(res.accessToken)
    saveSession({ token: res.accessToken, expiresAt: res.expiresAt, user: res.user })
    setUser(res.user)
    return res.user
  }, [])

  const logout = useCallback(() => {
    setToken(null)
    saveSession(null)
    setUser(null)
  }, [])

  const value = useMemo<AuthValue>(() => ({ user, login, logout }), [user, login, logout])
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
