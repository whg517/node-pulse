import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { HealthStatusBadge } from '../HealthStatusBadge'

describe('HealthStatusBadge', () => {
  describe('rendering', () => {
    it('should render healthy status with green color', () => {
      render(<HealthStatusBadge status="healthy" />)
      const badge = screen.getByRole('status', { name: /health status: 健康/i })
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-green-100', 'text-green-800')
      expect(badge).toHaveTextContent('健康')
    })

    it('should render warning status with yellow color', () => {
      render(<HealthStatusBadge status="warning" />)
      const badge = screen.getByRole('status', { name: /health status: 预警/i })
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-yellow-100', 'text-yellow-800')
      expect(badge).toHaveTextContent('预警')
    })

    it('should render critical status with red color', () => {
      render(<HealthStatusBadge status="critical" />)
      const badge = screen.getByRole('status', { name: /health status: 异常/i })
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-red-100', 'text-red-800')
      expect(badge).toHaveTextContent('异常')
    })

    it('should render offline status with gray color', () => {
      render(<HealthStatusBadge status="offline" />)
      const badge = screen.getByRole('status', { name: /health status: 离线/i })
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-gray-100', 'text-gray-800')
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
      expect(screen.getByRole('status').querySelector('.bg-green-500')).toBeInTheDocument()

      rerender(<HealthStatusBadge status="warning" />)
      expect(screen.getByRole('status').querySelector('.bg-yellow-500')).toBeInTheDocument()

      rerender(<HealthStatusBadge status="critical" />)
      expect(screen.getByRole('status').querySelector('.bg-red-500')).toBeInTheDocument()

      rerender(<HealthStatusBadge status="offline" />)
      expect(screen.getByRole('status').querySelector('.bg-gray-500')).toBeInTheDocument()
    })
  })
})
