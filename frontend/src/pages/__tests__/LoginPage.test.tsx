import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import '@testing-library/jest-dom'

// --- Mocks -------------------------------------------------------------

// Selector-aware mock of authStore, mirroring the pattern in
// src/hooks/useAuth.test.ts. The page calls useAuthStore((s) => s.X).
let mockAuthState: Record<string, unknown> = {}
vi.mock('@/stores/authStore', () => ({
  useAuthStore: vi.fn((selector?: (s: unknown) => unknown) => {
    return selector ? selector(mockAuthState) : mockAuthState
  }),
}))

// Mock the auth API module so no real HTTP is made.
const mockLogin = vi.fn()
vi.mock('@/api/auth', () => ({
  login: (...args: unknown[]) => mockLogin(...args),
}))

// Mock the constants module to avoid timer side effects in the test.
vi.mock('@/config/constants', () => ({
  ACCESS_TOKEN_EXPIRY_MINUTES: 15,
}))

// Capture navigation targets without exercising real routing.
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

// LoginPage is a default export; import after mocks are registered.
import LoginPage from '../LoginPage'

// Re-import the mocked store so tests can reset selector state.
import { useAuthStore } from '@/stores/authStore'

function renderPage(initialPath = '/login') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <LoginPage />
    </MemoryRouter>,
  )
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Default: logged-out, session check finished.
    mockAuthState = {
      isAuthenticated: false,
      isLoading: false,
      setUser: vi.fn(),
      setAccessToken: vi.fn(),
      setCsrfToken: vi.fn(),
    }
    vi.mocked(useAuthStore).mockImplementation(
      (selector?: (s: unknown) => unknown) =>
        selector ? selector(mockAuthState) : mockAuthState,
    )
  })

  it('renders the login form with username and password fields', () => {
    renderPage()
    expect(screen.getByLabelText(/Username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Sign in/i })).toBeInTheDocument()
  })

  it('does not submit when fields are empty', () => {
    renderPage()
    fireEvent.click(screen.getByRole('button', { name: /Sign in/i }))
    expect(mockLogin).not.toHaveBeenCalled()
  })

  it('submits credentials and navigates on success', async () => {
    mockLogin.mockResolvedValue({
      data: {
        user_id: 'u1',
        username: 'admin',
        role: 'admin',
        access_token: 'tok-123',
        csrf_token: 'csrf-456',
      },
    })
    renderPage()

    fireEvent.change(screen.getByLabelText(/Username/i), {
      target: { value: 'admin' },
    })
    fireEvent.change(screen.getByLabelText(/Password/i), {
      target: { value: 'secret123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /Sign in/i }))

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ username: 'admin', password: 'secret123' })
    })

    // store setters invoked with the response payload
    const setUser = mockAuthState.setUser as ReturnType<typeof vi.fn>
    expect(setUser).toHaveBeenCalledWith({ id: 'u1', username: 'admin', role: 'admin' })
    const setAccessToken = mockAuthState.setAccessToken as ReturnType<typeof vi.fn>
    expect(setAccessToken).toHaveBeenCalledWith('tok-123', expect.any(Number))
    const setCsrfToken = mockAuthState.setCsrfToken as ReturnType<typeof vi.fn>
    expect(setCsrfToken).toHaveBeenCalledWith('csrf-456')

    // navigates to the default target
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true })
  })

  it('shows an error banner on invalid credentials', async () => {
    mockLogin.mockRejectedValue({ code: 'ERR_INVALID_CREDENTIALS' })
    renderPage()

    fireEvent.change(screen.getByLabelText(/Username/i), {
      target: { value: 'admin' },
    })
    fireEvent.change(screen.getByLabelText(/Password/i), {
      target: { value: 'wrong' },
    })
    fireEvent.click(screen.getByRole('button', { name: /Sign in/i }))

    expect(await screen.findByText('Invalid username or password')).toBeInTheDocument()
    expect(mockNavigate).not.toHaveBeenCalled()
  })

  it('shows a rate-limit message when the API returns ERR_RATE_LIMITED', async () => {
    mockLogin.mockRejectedValue({ code: 'ERR_RATE_LIMITED' })
    renderPage()

    fireEvent.change(screen.getByLabelText(/Username/i), {
      target: { value: 'admin' },
    })
    fireEvent.change(screen.getByLabelText(/Password/i), {
      target: { value: 'x' },
    })
    fireEvent.click(screen.getByRole('button', { name: /Sign in/i }))

    expect(await screen.findByText(/Too many login attempts/i)).toBeInTheDocument()
  })

  it('shows a generic connection message for unexpected errors', async () => {
    mockLogin.mockRejectedValue(new Error('network down'))
    renderPage()

    fireEvent.change(screen.getByLabelText(/Username/i), {
      target: { value: 'admin' },
    })
    fireEvent.change(screen.getByLabelText(/Password/i), {
      target: { value: 'x' },
    })
    fireEvent.click(screen.getByRole('button', { name: /Sign in/i }))

    expect(await screen.findByText(/Connection failed/i)).toBeInTheDocument()
  })

  it('toggles password visibility', () => {
    renderPage()
    const passwordInput = screen.getByLabelText(/Password/i) as HTMLInputElement
    expect(passwordInput.type).toBe('password')

    fireEvent.click(screen.getByText('Show'))
    expect(passwordInput.type).toBe('text')

    fireEvent.click(screen.getByText('Hide'))
    expect(passwordInput.type).toBe('password')
  })

  it('redirects to the target path when already authenticated', () => {
    mockAuthState.isAuthenticated = true
    mockAuthState.isLoading = false
    vi.mocked(useAuthStore).mockImplementation(
      (selector?: (s: unknown) => unknown) =>
        selector ? selector(mockAuthState) : mockAuthState,
    )

    renderPage('/login')

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true })
  })
})
