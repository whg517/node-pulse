import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PerformanceMetricCard } from '../PerformanceMetricCard'
import type { PerformanceMetric } from '../../../api/performance'

describe('PerformanceMetricCard', () => {
  const healthyMetric: PerformanceMetric = {
    metric_name: 'latency',
    display_name: 'Latency',
    current_p99: 80,
    current_p95: 50,
    target_p99: 100,
    target_p95: 80,
    unit: 'ms',
    status: 'healthy',
  }

  const unhealthyMetric: PerformanceMetric = {
    metric_name: 'latency',
    display_name: 'Latency',
    current_p99: 200,
    current_p95: 150,
    target_p99: 100,
    target_p95: 80,
    unit: 'ms',
    status: 'unhealthy',
    anomaly: 'High latency detected',
  }

  it('renders metric display name', () => {
    render(<PerformanceMetricCard metric={healthyMetric} />)
    expect(screen.getByText('Latency')).toBeInTheDocument()
  })

  it('renders P99 current value', () => {
    render(<PerformanceMetricCard metric={healthyMetric} />)
    expect(screen.getByText('80 ms')).toBeInTheDocument()
  })

  it('renders P95 current value', () => {
    render(<PerformanceMetricCard metric={healthyMetric} />)
    expect(screen.getByText('50 ms')).toBeInTheDocument()
  })

  it('renders target values', () => {
    render(<PerformanceMetricCard metric={healthyMetric} />)
    expect(screen.getByText('≤ 100 ms')).toBeInTheDocument()
    expect(screen.getByText('≤ 80 ms')).toBeInTheDocument()
  })

  it('shows healthy status for healthy metric', () => {
    render(<PerformanceMetricCard metric={healthyMetric} />)
    expect(screen.getByText('健康')).toBeInTheDocument()
  })

  it('shows unhealthy status for unhealthy metric', () => {
    render(<PerformanceMetricCard metric={unhealthyMetric} />)
    expect(screen.getByText('异常')).toBeInTheDocument()
  })

  it('renders anomaly message when present', () => {
    render(<PerformanceMetricCard metric={unhealthyMetric} />)
    expect(screen.getByText('High latency detected')).toBeInTheDocument()
  })

  it('does not render anomaly section for healthy metric', () => {
    render(<PerformanceMetricCard metric={healthyMetric} />)
    expect(screen.queryByText(/detected/)).not.toBeInTheDocument()
  })

  it('applies green border for healthy metric', () => {
    const { container } = render(<PerformanceMetricCard metric={healthyMetric} />)
    const card = container.firstChild as HTMLElement
    expect(card.className).toContain('border-[var(--color-healthy)]')
  })

  it('applies red border for unhealthy metric', () => {
    const { container } = render(<PerformanceMetricCard metric={unhealthyMetric} />)
    const card = container.firstChild as HTMLElement
    expect(card.className).toContain('border-[var(--color-critical)]')
  })
})
