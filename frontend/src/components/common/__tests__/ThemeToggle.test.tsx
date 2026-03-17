import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ThemeToggle } from '../ThemeToggle'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

// Default mock for useTheme (light mode)
const mockToggleTheme = vi.fn()
const mockSetTheme = vi.fn()

vi.mock('../../../hooks/useTheme', () => ({
  useTheme: vi.fn(() => ({
    theme: 'light',
    isDark: false,
    toggleTheme: mockToggleTheme,
    setTheme: mockSetTheme,
  })),
}))

describe('ThemeToggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders toggle button by default', () => {
    render(<ThemeToggle />)
    const button = screen.getByRole('button')
    expect(button).toBeInTheDocument()
  })

  it('shows moon icon when in light mode', () => {
    render(<ThemeToggle />)
    const button = screen.getByRole('button')
    // In light mode, show dark mode toggle (moon)
    expect(button).toHaveAttribute('aria-label', 'settings.darkMode')
  })

  it('calls toggleTheme when button is clicked', () => {
    render(<ThemeToggle />)
    fireEvent.click(screen.getByRole('button'))
    expect(mockToggleTheme).toHaveBeenCalledTimes(1)
  })

  it('shows sun icon when in dark mode', async () => {
    const { useTheme } = await import('../../../hooks/useTheme')
    vi.mocked(useTheme).mockReturnValueOnce({
      theme: 'dark',
      effectiveTheme: 'dark',
      isDark: true,
      isLight: false,
      isSystem: false,
      toggleTheme: mockToggleTheme,
      setTheme: mockSetTheme,
    })
    render(<ThemeToggle />)
    const button = screen.getByRole('button')
    expect(button).toHaveAttribute('aria-label', 'settings.lightMode')
  })

  it('renders dropdown when showDropdown is true', () => {
    render(<ThemeToggle showDropdown={true} />)
    const buttons = screen.getAllByRole('button')
    // Should have 3 buttons: light, dark, system
    expect(buttons.length).toBe(3)
  })

  it('dropdown buttons call setTheme', () => {
    render(<ThemeToggle showDropdown={true} />)
    const buttons = screen.getAllByRole('button')
    fireEvent.click(buttons[1]) // dark button
    expect(mockSetTheme).toHaveBeenCalledWith('dark')
  })

  it('applies custom className', () => {
    const { container } = render(<ThemeToggle className="my-class" />)
    const button = container.querySelector('button')
    expect(button?.className).toContain('my-class')
  })

  it('renders with sm size', () => {
    render(<ThemeToggle size="sm" />)
    const button = screen.getByRole('button')
    expect(button.className).toContain('p-1')
  })

  it('renders with lg size', () => {
    render(<ThemeToggle size="lg" />)
    const button = screen.getByRole('button')
    expect(button.className).toContain('p-2.5')
  })
})
