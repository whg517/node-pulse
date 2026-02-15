/**
 * Authentication TypeScript types
 *
 * Aligned with backend pulse/internal/auth/auth_handler.go response structures
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
  id: string
  user_id: string
  created_at: string
  last_used_at: string
  expires_at: string
  ip_address: string
  user_agent: string
  is_current: boolean
}

export interface SessionListResponse {
  data: {
    sessions: Session[]
    total: number
  }
  message: string
  timestamp: string
}

export interface SessionInfoResponse {
  data: {
    current_session_id: string
    active_sessions_count: number
  }
  message: string
  timestamp: string
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
