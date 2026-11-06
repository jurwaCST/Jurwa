// Shapes from proto/identity/v1 as the gateway sends them (lowerCamelCase JSON,
// enums as their names, see docs/design/api-routes.md rule 3).

/** Role names from common.v1.Role. */
export type Role = 'STUDENT' | 'CATEGORY_HEAD' | 'STAFF' | 'SYSTEM_ADMIN' | 'PRESIDENT'

export interface User {
  id: string
  email: string
  fullName: string
  roles: Role[]
}

/** Body of POST /auth/login. */
export interface LoginResponse {
  accessToken: string
  expiresAt: string
  user: User
}

/** What the app keeps after login. Never the password. */
export interface Session {
  token: string
  expiresAt: string
  user: User
}
