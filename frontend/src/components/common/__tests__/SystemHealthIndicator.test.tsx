import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SystemHealthIndicator } from '../SystemHealthIndicator'

describe('SystemHealthIndicator', () => {
  it('renders healthy status', () => {
    render(<SystemHealthIndicator health="healthy" />)
    expect(screen.getByText('健康')).toBeInTheDocument()
  })

  it('renders unhealthy status', () => {
    render(<SystemHealthIndicator health="unhealthy" />)
    expect(screen.getByText('异常')).toBeInTheDocument()
  })

  it('applies green border for healthy', () => {
    const { container } = render(<SystemHealthIndicator health="healthy" />)
    const indicator = container.querySelector('.rounded-full.border-4')
    expect(indicator?.className).toContain('border-[var(--color-healthy)]')
  })

  it('applies red border for unhealthy', () => {
    const { container } = render(<SystemHealthIndicator health="unhealthy" />)
    const indicator = container.querySelector('.rounded-full.border-4')
    expect(indicator?.className).toContain('border-[var(--color-critical)]')
  })

  it('applies custom className', () => {
    const { container } = render(
      <SystemHealthIndicator health="healthy" className="mt-4" />
    )
    const wrapper = container.firstChild as HTMLElement
    expect(wrapper.className).toContain('mt-4')
  })

  it('renders inner pulse circle for healthy', () => {
    const { container } = render(<SystemHealthIndicator health="healthy" />)
    const inner = container.querySelector('.animate-pulse')
    expect(inner?.className).toContain('bg-[var(--color-healthy)]')
  })

  it('renders inner pulse circle for unhealthy', () => {
    const { container } = render(<SystemHealthIndicator health="unhealthy" />)
    const inner = container.querySelector('.animate-pulse')
    expect(inner?.className).toContain('bg-[var(--color-critical)]')
  })
})
