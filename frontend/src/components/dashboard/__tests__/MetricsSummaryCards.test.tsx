import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { MetricsSummaryCards } from '../MetricsSummaryCards'
import type { MetricsDTO } from '../../../api/types'

describe('MetricsSummaryCards', () => {
  const mockMetrics: MetricsDTO[] = [
    {
      node_id: 'node-1',
      latency_ms: 50,
      packet_loss_rate: 0.1,
      jitter_ms: 5,
      timestamp: '2026-01-26T10:00:00Z',
    },
    {
      node_id: 'node-2',
      latency_ms: 150,
      packet_loss_rate: 2.5,
      jitter_ms: 30,
      timestamp: '2026-01-26T10:00:00Z',
    },
  ]

  describe('rendering', () => {
    it('should render three metric cards', () => {
      render(<MetricsSummaryCards metrics={mockMetrics} />)

      expect(screen.getByText('Avg Latency')).toBeInTheDocument()
      expect(screen.getByText('Avg Packet Loss')).toBeInTheDocument()
      expect(screen.getByText('Avg Jitter')).toBeInTheDocument()
    })

    it('should display calculated average values', () => {
      render(<MetricsSummaryCards metrics={mockMetrics} />)

      // Average latency: (50 + 150) / 2 = 100
      expect(screen.getByText('100.0')).toBeInTheDocument()
      const msLabels = screen.getAllByText('ms')
      expect(msLabels.length).toBeGreaterThan(0)

      // Average packet loss: (0.1 + 2.5) / 2 = 1.3
      expect(screen.getByText('1.30')).toBeInTheDocument()

      // Average jitter: (5 + 30) / 2 = 17.5
      expect(screen.getByText('17.5')).toBeInTheDocument()
    })

    it('should display node count', () => {
      render(<MetricsSummaryCards metrics={mockMetrics} />)

      const nodeCounts = screen.getAllByText('Across 2 nodes')
      expect(nodeCounts.length).toBe(3) // One for each card
    })

    it('should use singular "node" when only one node', () => {
      const singleNodeMetrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 100,
          packet_loss_rate: 1,
          jitter_ms: 20,
          timestamp: '2026-01-26T10:00:00Z',
        },
      ]

      render(<MetricsSummaryCards metrics={singleNodeMetrics} />)

      const nodeCounts = screen.getAllByText('Across 1 node')
      expect(nodeCounts.length).toBe(3)
    })
  })

  describe('color coding', () => {
    it('should show green color for good metrics', () => {
      const goodMetrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 50, // Good
          packet_loss_rate: 0.1, // Good
          jitter_ms: 5, // Good
          timestamp: '2026-01-26T10:00:00Z',
        },
      ]

      const { container } = render(
        <MetricsSummaryCards metrics={goodMetrics} />
      )

      // Check for healthy background class
      expect(container.querySelector('[class*="bg-[var(--color-healthy-bg)"]')).toBeInTheDocument()
    })

    it('should show warning color for warning metrics (80-100% threshold)', () => {
      const warningMetrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 170, // 85% of 200ms threshold
          packet_loss_rate: 4.2, // 84% of 5% threshold
          jitter_ms: 42, // 84% of 50ms threshold
          timestamp: '2026-01-26T10:00:00Z',
        },
      ]

      const { container } = render(
        <MetricsSummaryCards metrics={warningMetrics} />
      )

      // Check for warning background
      expect(container.querySelector('[class*="bg-[var(--color-warning-bg)"]')).toBeInTheDocument()
    })

    it('should show critical color for critical metrics (exceeds threshold)', () => {
      const criticalMetrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 250, // Exceeds 200ms
          packet_loss_rate: 10, // Exceeds 5%
          jitter_ms: 80, // Exceeds 50ms
          timestamp: '2026-01-26T10:00:00Z',
        },
      ]

      const { container } = render(
        <MetricsSummaryCards metrics={criticalMetrics} />
      )

      // Check for critical background
      expect(container.querySelector('[class*="bg-[var(--color-critical-bg)"]')).toBeInTheDocument()
    })
  })

  describe('loading state', () => {
    it('should render loading skeletons when isLoading is true', () => {
      render(<MetricsSummaryCards metrics={[]} isLoading={true} />)

      const skeletons = document.querySelectorAll('.animate-pulse')
      expect(skeletons.length).toBeGreaterThan(0)
    })

    it('should render three skeleton cards', () => {
      render(<MetricsSummaryCards metrics={[]} isLoading={true} />)

      const skeletons = document.querySelectorAll('.animate-pulse')
      expect(skeletons.length).toBe(3)
    })
  })

  describe('empty state', () => {
    it('should display zero values when no metrics', () => {
      render(<MetricsSummaryCards metrics={[]} />)

      const zeroValues = screen.getAllByText('0.0')
      expect(zeroValues.length).toBeGreaterThan(0)

      const zeroDecimalValues = screen.getAllByText('0.00')
      expect(zeroDecimalValues.length).toBeGreaterThan(0)

      const nodeCounts = screen.getAllByText('Across 0 nodes')
      expect(nodeCounts.length).toBe(3)
    })

    it('should handle single data point correctly', () => {
      const singleMetric: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 123,
          packet_loss_rate: 3.5,
          jitter_ms: 27,
          timestamp: '2026-01-26T10:00:00Z',
        },
      ]

      render(<MetricsSummaryCards metrics={singleMetric} />)

      expect(screen.getByText('123.0')).toBeInTheDocument()
      expect(screen.getByText('3.50')).toBeInTheDocument()
      expect(screen.getByText('27.0')).toBeInTheDocument()
    })
  })

  describe('calculations', () => {
    it('should calculate average latency correctly', () => {
      const metrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 100,
          packet_loss_rate: 0,
          jitter_ms: 0,
          timestamp: '2026-01-26T10:00:00Z',
        },
        {
          node_id: 'node-2',
          latency_ms: 200,
          packet_loss_rate: 0,
          jitter_ms: 0,
          timestamp: '2026-01-26T10:00:00Z',
        },
      ]

      render(<MetricsSummaryCards metrics={metrics} />)

      expect(screen.getByText('150.0')).toBeInTheDocument() // (100 + 200) / 2
    })

    it('should calculate average packet loss correctly', () => {
      const metrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 0,
          packet_loss_rate: 1.0,
          jitter_ms: 0,
          timestamp: '2026-01-26T10:00:00Z',
        },
        {
          node_id: 'node-2',
          latency_ms: 0,
          packet_loss_rate: 3.0,
          jitter_ms: 0,
          timestamp: '2026-01-26T10:00:00Z',
        },
      ]

      render(<MetricsSummaryCards metrics={metrics} />)

      expect(screen.getByText('2.00')).toBeInTheDocument() // (1.0 + 3.0) / 2
    })

    it('should calculate average jitter correctly', () => {
      const metrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 0,
          packet_loss_rate: 0,
          jitter_ms: 10,
          timestamp: '2026-01-26T10:00:00Z',
        },
        {
          node_id: 'node-2',
          latency_ms: 0,
          packet_loss_rate: 0,
          jitter_ms: 20,
          timestamp: '2026-01-26T10:00:00Z',
        },
      ]

      render(<MetricsSummaryCards metrics={metrics} />)

      expect(screen.getByText('15.0')).toBeInTheDocument() // (10 + 20) / 2
    })
  })

  describe('responsive design', () => {
    it('should render with responsive grid classes', () => {
      const { container } = render(
        <MetricsSummaryCards metrics={mockMetrics} />
      )

      const grid = container.querySelector('.grid')
      expect(grid).toHaveClass('grid-cols-1')
      expect(grid).toHaveClass('sm:grid-cols-2')
      expect(grid).toHaveClass('lg:grid-cols-3')
    })
  })

  describe('accessibility', () => {
    it('should have proper headings for each card', () => {
      render(<MetricsSummaryCards metrics={mockMetrics} />)

      const headings = screen.getAllByText(/Avg (Latency|Packet Loss|Jitter)/)
      expect(headings.length).toBe(3)
    })

    it('should have proper SVG icons with aria-hidden', () => {
      render(<MetricsSummaryCards metrics={mockMetrics} />)

      const svgs = document.querySelectorAll('svg[aria-hidden="true"]')
      expect(svgs.length).toBeGreaterThan(0)
    })
  })
})
