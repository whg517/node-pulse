import type { LoginRequest, LoginResponse, LogoutResponse } from '../types/auth'

export type { LoginRequest, LoginResponse, LogoutResponse }

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const SESSION_COOKIE_NAME = 'session_id'

export async function login(credentials: LoginRequest): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(credentials),
  })

  if (!response.ok) {
    const errorData = await response.json()
    throw new LoginError(errorData.message, errorData.code, errorData.details)
  }

  const data: LoginResponse = await response.json()
  return data
}

export async function logout(): Promise<LogoutResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error('Logout failed')
  }

  const data: LogoutResponse = await response.json()
  return data
}

export class LoginError extends Error {
  code: string
  details?: {
    failed_attempts?: number
    remaining_attempts?: number
    locked_until?: string
    lock_duration_minutes?: number
  }

  constructor(
    message: string,
    code: string,
    details?: {
      failed_attempts?: number
      remaining_attempts?: number
      locked_until?: string
      lock_duration_minutes?: number
    }
  ) {
    super(message)
    this.name = 'LoginError'
    this.code = code
    this.details = details
    Object.setPrototypeOf(this, LoginError.prototype)
  }
}

export { SESSION_COOKIE_NAME }
