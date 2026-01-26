// Authentication TypeScript types

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  data: {
    user_id: string
    username: string
    role: 'admin' | 'operator' | 'viewer'
  }
  message: string
  timestamp: string
}

export interface LoginErrorResponse {
  code: 'ERR_INVALID_CREDENTIALS' | 'ERR_ACCOUNT_LOCKED' | 'ERR_RATE_LIMITED'
  message: string
  details?: {
    failed_attempts?: number
    remaining_attempts?: number
    locked_until?: string
    lock_duration_minutes?: number
  }
}

export interface LogoutResponse {
  message: string
  timestamp: string
}

export interface ValidationError {
  field: 'username' | 'password'
  message: string
}
