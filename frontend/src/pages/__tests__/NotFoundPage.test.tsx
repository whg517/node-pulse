import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import NotFoundPage from '../NotFoundPage'

vi.mock('../../hooks/useTheme', () => ({
  useTheme: () => ({ isDark: false }),
}))

describe('NotFoundPage', () => {
  it('renders 404 text', () => {
    render(<MemoryRouter><NotFoundPage /></MemoryRouter>)
    expect(screen.getByText('404')).toBeInTheDocument()
  })

  it('renders page not found heading', () => {
    render(<MemoryRouter><NotFoundPage /></MemoryRouter>)
    expect(screen.getByText('Page Not Found')).toBeInTheDocument()
  })

  it('renders description text', () => {
    render(<MemoryRouter><NotFoundPage /></MemoryRouter>)
    expect(screen.getByText("The page you're looking for doesn't exist or has been moved.")).toBeInTheDocument()
  })

  it('renders back to dashboard link', () => {
    render(<MemoryRouter><NotFoundPage /></MemoryRouter>)
    const link = screen.getByRole('link', { name: /Back to Dashboard/i })
    expect(link).toBeInTheDocument()
    expect(link).toHaveAttribute('href', '/dashboard')
  })
})
