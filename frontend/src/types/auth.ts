/**
 * Authentication TypeScript types
 *
 * Aligned with backend pulse/internal/auth/auth_handler.go response structures
 *
 * JWT Token Details:
 * - Algorithm: RS256 (RSA asymmetric signing)
 * - Access tokens: Signed with RSA private key, verified with public key
 * - Key ID (kid): Included in token header for key rotation support
 * - Refresh tokens: Opaque tokens stored in HttpOnly cookies
 */

// ============== User Types ==============

export type UserRole = 'admin' | 'operator' | 'viewer'

export interface User {
  id: string
  username: string
  role: UserRole
}

// ============== Request/Response Types ==============

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  data: {
    user_id: string
    username: string
    role: UserRole
    access_token: string
    csrf_token: string
    // Present when the account has 2FA enabled — the client must then call
    // /auth/login/mfa with this ticket + a TOTP code to finish login.
    mfa_required?: boolean
    mfa_ticket?: string
  }
  message: string
  timestamp: string
}

export interface RefreshResponse {
  data: {
    access_token: string
    expires_in: number // seconds until expiry
  }
  message: string
  timestamp: string
}

export interface GetMeResponse {
  data: {
    user_id: string
    username: string
    role: UserRole
  }
  message: string
  timestamp: string
}

export interface LogoutResponse {
  message: string
  timestamp: string
}

// ============== Session Management Types ==============

export interface Session {
  session_id: string
  created_at: string
  expires_at: string
  max_valid_until: string
  ip_address: string
  user_agent: string
}

export type SessionListResponse = Session[]

export interface SessionInfoResponse {
  session_id: string
  expires_at: string
  max_valid_until: string
}

// ============== Error Types ==============

export interface LoginErrorResponse {
  code: 'ERR_INVALID_CREDENTIALS' | 'ERR_ACCOUNT_LOCKED' | 'ERR_RATE_LIMITED' | 'ERR_RATE_LIMIT_EXCEEDED'
  message: string
  details?: {
    failed_attempts?: number
    remaining_attempts?: number
    locked_until?: string
    lock_duration_minutes?: number
    retry_after?: number // seconds until rate limit resets
  }
}

export interface ValidationError {
  field: 'username' | 'password'
  message: string
}

// ============== Auth State Types ==============

export interface AuthState {
  user: User | null
  isAuthenticated: boolean
  role: UserRole | null
  accessToken: string | null
  tokenExpiresAt: number | null
  isLoading: boolean
  refreshFailureCount: number
}
