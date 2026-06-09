import { apiClient } from './client'

export interface UserDTO {
  user_id: string
  username: string
  email?: string
  role: 'admin' | 'operator' | 'viewer'
  failed_login_attempts: number
  locked_until?: string
  mfa_enabled: boolean
  created_at: string
  updated_at: string
}

export interface ListUsersResponse {
  users: UserDTO[]
  total: number
  limit: number
  offset: number
  pagination: { has_next: boolean; has_prev: boolean }
}

export interface CreateUserRequest {
  username: string
  email?: string
  password: string
  role: 'admin' | 'operator' | 'viewer'
}

export interface UpdateUserRequest {
  username?: string
  email?: string
  password?: string
  role?: 'admin' | 'operator' | 'viewer'
}

export async function fetchUsers(
  limit = 100,
  offset = 0
): Promise<{ data: ListUsersResponse; message: string; timestamp: string }> {
  return apiClient(`/api/v1/admin/users?limit=${limit}&offset=${offset}`)
}

export async function fetchUser(
  id: string
): Promise<{ data: { user: UserDTO }; message: string; timestamp: string }> {
  return apiClient(`/api/v1/admin/users/${id}`)
}

export async function createUser(
  request: CreateUserRequest
): Promise<{ data: { user: UserDTO }; message: string; timestamp: string }> {
  return apiClient('/api/v1/admin/users', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

export async function updateUser(
  id: string,
  request: UpdateUserRequest
): Promise<{ data: { user: UserDTO }; message: string; timestamp: string }> {
  return apiClient(`/api/v1/admin/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  })
}

export async function deleteUser(
  id: string
): Promise<{ message: string; timestamp: string }> {
  return apiClient(`/api/v1/admin/users/${id}`, {
    method: 'DELETE',
  })
}
