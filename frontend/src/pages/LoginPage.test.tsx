import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import LoginPage from './LoginPage'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useLocation: () => ({
      state: null,
    }),
  }
})

vi.mock('../api/auth', () => ({
  login: vi.fn(),
  logout: vi.fn(),
}))

vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn((selector) => {
    const state = {
      isAuthenticated: false,
      userId: null,
      username: null,
      role: null,
      sessionExpiry: null,
      setSession: vi.fn(),
      clearSession: vi.fn(),
      checkSession: vi.fn(() => false),
    }
    return selector ? selector(state) : state
  }),
}))

import { login } from '../api/auth'

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders login form with username and password fields', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Sign in' })).toBeInTheDocument()
  })

  it('shows validation errors for empty fields', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const submitButton = screen.getByRole('button', { name: 'Sign in' })
    await userEvent.click(submitButton)

    expect(screen.getByText('Username is required')).toBeInTheDocument()
    expect(screen.getByText('Password is required')).toBeInTheDocument()
  })

  it('shows validation error for short username', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'ab')
    await userEvent.type(passwordInput, 'password123')
    await userEvent.click(submitButton)

    expect(screen.getByText('Username must be at least 3 characters')).toBeInTheDocument()
  })

  it('shows validation error for short password', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'pass')
    await userEvent.click(submitButton)

    expect(screen.getByText('Password must be at least 8 characters')).toBeInTheDocument()
  })

  it('shows validation error for long password', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'a'.repeat(33))
    await userEvent.click(submitButton)

    expect(screen.getByText('Password must be less than 32 characters')).toBeInTheDocument()
  })

  it('calls login API and navigates to dashboard on success', async () => {
    ;(login as any).mockResolvedValue({
      data: {
        user_id: '123',
        username: 'admin',
        role: 'admin',
      },
      message: 'Login successful',
      timestamp: '2026-01-26T10:00:00Z',
    })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'Password123')
    await userEvent.click(submitButton)

    expect(login).toHaveBeenCalledWith({ username: 'admin', password: 'Password123' })
  })

  it('displays error message on invalid credentials', async () => {
    ;(login as any).mockRejectedValue({
      code: 'ERR_INVALID_CREDENTIALS',
      message: 'Invalid username or password',
    })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'WrongPassword123')
    await userEvent.click(submitButton)

    expect(screen.getByText('Invalid username or password')).toBeInTheDocument()
    expect(passwordInput).toHaveValue('')
  })

  it('displays account locked message when account is locked', async () => {
    ;(login as any).mockRejectedValue({
      code: 'ERR_ACCOUNT_LOCKED',
      message: 'Account locked due to too many failed login attempts',
      details: {
        locked_until: '2026-01-26T11:00:00Z',
        lock_duration_minutes: 10,
      },
    })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'WrongPassword123')
    await userEvent.click(submitButton)

    // Wait for async state updates and error rendering
    await waitFor(() => {
      const accountLockedTexts = screen.queryAllByText(/Account locked/)
      expect(accountLockedTexts.length).toBeGreaterThan(0)
    })
  })

  it('disables submit button while loading', async () => {
    let resolveLogin: (value: any) => void
    const mockLoginPromise = new Promise((resolve) => {
      resolveLogin = resolve
    })
    ;(login as any).mockReturnValue(mockLoginPromise)

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    // Use valid password (uppercase + lowercase + digit, min 8 chars)
    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'Password123')

    // Click the button
    await userEvent.click(submitButton)

    // Wait for loading state to be set
    await waitFor(
      () => {
        expect(submitButton).toBeDisabled()
      },
      { timeout: 3000 }
    )

    // Verify loading text is shown
    expect(screen.getByText('Signing in...')).toBeInTheDocument()

    // Resolve the login promise to clean up
    resolveLogin!({
      data: { user_id: '123', username: 'admin', role: 'admin' },
      message: 'Login successful',
      timestamp: '2026-01-26T10:00:00Z'
    })
  })
})
