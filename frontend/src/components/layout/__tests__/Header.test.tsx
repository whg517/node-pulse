import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Header } from '../Header'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
  initReactI18next: { type: '3rdParty', init: vi.fn() },
}))

// Mock React Router hooks
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

// Mock ThemeToggle, LanguageSwitcher, TimezoneSelector
vi.mock('../../common/ThemeToggle', () => ({
  ThemeToggle: () => <button data-testid="theme-toggle">Theme</button>,
}))
vi.mock('../../common/LanguageSwitcher', () => ({
  LanguageSwitcher: () => <select data-testid="language-switcher" />,
}))
vi.mock('../../common/TimezoneSelector', () => ({
  TimezoneSelector: () => <div data-testid="timezone-selector" />,
}))

// Mock authStore
const mockLogout = vi.fn()
vi.mock('../../../stores/authStore', () => ({
  useAuthStore: vi.fn((selector: (s: unknown) => unknown) =>
    selector({
      user: { username: 'testuser', role: 'admin' },
      logout: mockLogout,
    })
  ),
}))

describe('Header', () => {
  const defaultProps = {
    onMenuToggle: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders header element', () => {
    render(
      <MemoryRouter>
        <Header {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByRole('banner')).toBeInTheDocument()
  })

  it('calls onMenuToggle when menu button is clicked', () => {
    render(
      <MemoryRouter>
        <Header {...defaultProps} />
      </MemoryRouter>
    )
    const menuButton = screen.getByLabelText('nav.toggleMenu')
    fireEvent.click(menuButton)
    expect(defaultProps.onMenuToggle).toHaveBeenCalledTimes(1)
  })

  it('displays username in header', () => {
    render(
      <MemoryRouter>
        <Header {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByText('testuser')).toBeInTheDocument()
  })

  it('opens user menu on button click', () => {
    render(
      <MemoryRouter>
        <Header {...defaultProps} />
      </MemoryRouter>
    )
    const userBtn = screen.getAllByRole('button').find(b => b.className.includes('rounded-full') || b.getAttribute('aria-label')?.includes('user'))
    // Find button containing username
    const userMenuButton = screen.getByText('testuser').closest('button') ??
      screen.getAllByRole('button')[screen.getAllByRole('button').length - 1]
    fireEvent.click(userMenuButton)
    expect(screen.getByText('nav.logout')).toBeInTheDocument()
  })

  it('calls logout and navigates on logout click', async () => {
    mockLogout.mockResolvedValueOnce(undefined)
    render(
      <MemoryRouter>
        <Header {...defaultProps} />
      </MemoryRouter>
    )
    // Open user menu first
    const buttons = screen.getAllByRole('button')
    const userMenuButton = buttons[buttons.length - 1]
    fireEvent.click(userMenuButton)

    const logoutButton = screen.getByText('nav.logout')
    fireEvent.click(logoutButton)

    await waitFor(() => {
      expect(mockLogout).toHaveBeenCalledTimes(1)
    })
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login')
    })
  })

  it('renders ThemeToggle', () => {
    render(
      <MemoryRouter>
        <Header {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByTestId('theme-toggle')).toBeInTheDocument()
  })

  it('renders LanguageSwitcher', () => {
    render(
      <MemoryRouter>
        <Header {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByTestId('language-switcher')).toBeInTheDocument()
  })
})
