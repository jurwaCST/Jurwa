import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import Forbidden from '../pages/Forbidden'
import type { Role } from './types'
import { useAuth } from './useAuth'

interface Props {
  /** Roles that may open the page. */
  roles: Role[]
  children: ReactNode
}

/**
 * RequireRole sends a visitor who is not signed in to /login (and back here
 * after login), and shows a 403 page to a user without one of the roles.
 * The gateway checks roles again on every call; this only hides pages.
 */
export default function RequireRole({ roles, children }: Props) {
  const { user } = useAuth()
  const location = useLocation()

  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }
  if (!user.roles.some((r) => roles.includes(r))) {
    return <Forbidden />
  }
  return <>{children}</>
}
