import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { HealthStatusBadge } from '../HealthStatusBadge'

describe('HealthStatusBadge', () => {
  describe('rendering', () => {
    it('should render healthy status with green color', () => {
      render(<HealthStatusBadge status="healthy" />)
      const badge = screen.getByRole('status', { name: /health status: 健康/i })
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-healthy-bg', 'text-healthy-text')
      expect(badge).toHaveTextContent('健康')
    })

    it('should render warning status with yellow color', () => {
      render(<HealthStatusBadge status="warning" />)
      const badge = screen.getByRole('status', { name: /health status: 预警/i })
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-warning-bg', 'text-warning-text')
      expect(badge).toHaveTextContent('预警')
    })

    it('should render critical status with red color', () => {
      render(<HealthStatusBadge status="critical" />)
      const badge = screen.getByRole('status', { name: /health status: 异常/i })
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-destructive/10', 'text-destructive')
      expect(badge).toHaveTextContent('异常')
    })

    it('should render offline status with gray color', () => {
      render(<HealthStatusBadge status="offline" />)
      const badge = screen.getByRole('status', { name: /health status: 离线/i })
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-muted', 'text-muted-foreground')
      expect(badge).toHaveTextContent('离线')
    })
  })

  describe('accessibility', () => {
    it('should have proper role attribute', () => {
      render(<HealthStatusBadge status="healthy" />)
      const badge = screen.getByRole('status')
      expect(badge).toBeInTheDocument()
    })

    it('should have aria-label for screen readers', () => {
      render(<HealthStatusBadge status="critical" />)
      const badge = screen.getByRole('status', { name: /health status: 异常/i })
      expect(badge).toHaveAttribute('aria-label', 'Health status: 异常')
    })

    it('should include visual indicator dot with aria-hidden', () => {
      render(<HealthStatusBadge status="warning" />)
      const badge = screen.getByRole('status')
      const dot = badge.querySelector('span[aria-hidden="true"]')
      expect(dot).toBeInTheDocument()
      expect(dot).toHaveClass('w-2', 'h-2', 'rounded-full')
    })
  })

  describe('styling', () => {
    it('should apply rounded corners and padding', () => {
      render(<HealthStatusBadge status="healthy" />)
      const badge = screen.getByRole('status')
      expect(badge).toHaveClass('rounded-full', 'px-2.5', 'py-0.5')
    })

    it('should apply font styling', () => {
      render(<HealthStatusBadge status="critical" />)
      const badge = screen.getByRole('status')
      expect(badge).toHaveClass('text-xs', 'font-medium')
    })

    it('should use flex layout for proper alignment', () => {
      render(<HealthStatusBadge status="offline" />)
      const badge = screen.getByRole('status')
      expect(badge).toHaveClass('inline-flex', 'items-center', 'gap-1.5')
    })

    it('should display colored dot for all status types', () => {
      const { rerender } = render(<HealthStatusBadge status="healthy" />)
      expect(screen.getByRole('status').querySelector('.bg-healthy')).toBeInTheDocument()

      rerender(<HealthStatusBadge status="warning" />)
      expect(screen.getByRole('status').querySelector('.bg-warning')).toBeInTheDocument()

      rerender(<HealthStatusBadge status="critical" />)
      expect(screen.getByRole('status').querySelector('.bg-destructive')).toBeInTheDocument()

      rerender(<HealthStatusBadge status="offline" />)
      expect(screen.getByRole('status').querySelector('.bg-muted-foreground')).toBeInTheDocument()
    })
  })
})
