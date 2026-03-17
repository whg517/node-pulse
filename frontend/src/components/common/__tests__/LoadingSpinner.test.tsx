import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { LoadingSpinner } from '../LoadingSpinner'

describe('LoadingSpinner', () => {
  it('renders with default props', () => {
    render(<LoadingSpinner />)
    const spinner = screen.getByRole('status')
    expect(spinner).toBeInTheDocument()
  })

  it('renders with custom label', () => {
    render(<LoadingSpinner label="Please wait..." />)
    expect(screen.getByLabelText('Please wait...')).toBeInTheDocument()
    expect(screen.getByText('Please wait...')).toBeInTheDocument()
  })

  it('renders small size', () => {
    render(<LoadingSpinner size="sm" />)
    const spinner = screen.getByRole('status')
    expect(spinner).toBeInTheDocument()
    const inner = spinner.querySelector('div')
    expect(inner?.className).toContain('h-8')
    expect(inner?.className).toContain('w-8')
  })

  it('renders medium size by default', () => {
    render(<LoadingSpinner />)
    const spinner = screen.getByRole('status')
    const inner = spinner.querySelector('div')
    expect(inner?.className).toContain('h-12')
    expect(inner?.className).toContain('w-12')
  })

  it('renders large size', () => {
    render(<LoadingSpinner size="lg" />)
    const spinner = screen.getByRole('status')
    const inner = spinner.querySelector('div')
    expect(inner?.className).toContain('h-16')
    expect(inner?.className).toContain('w-16')
  })

  it('applies custom className', () => {
    render(<LoadingSpinner className="mt-4 custom-class" />)
    const spinner = screen.getByRole('status')
    expect(spinner.className).toContain('mt-4')
    expect(spinner.className).toContain('custom-class')
  })

  it('renders accessibility label in sr-only span', () => {
    render(<LoadingSpinner label="Loading data" />)
    const srOnly = screen.getByText('Loading data')
    expect(srOnly.className).toContain('sr-only')
  })
})
