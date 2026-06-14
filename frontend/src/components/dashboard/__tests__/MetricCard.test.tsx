import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import MetricCard from '../MetricCard'

describe('MetricCard', () => {
  it('renders metric title, value, and unit', () => {
    render(<MetricCard title="Latency" value={45} unit="ms" />)

    expect(screen.getByText('Latency')).toBeInTheDocument()
    expect(screen.getByText('45')).toBeInTheDocument()
    expect(screen.getByText('ms')).toBeInTheDocument()
  })

  it('renders without unit when not provided', () => {
    render(<MetricCard title="Status" value="Online" />)

    expect(screen.getByText('Status')).toBeInTheDocument()
    expect(screen.getByText('Online')).toBeInTheDocument()
    expect(screen.queryByText('ms')).not.toBeInTheDocument()
  })

  it('applies correct status colors for good status', () => {
    const { container } = render(
      <MetricCard title="Latency" value={45} unit="ms" status="good" />
    )

    const card = container.querySelector('.metric-card')
    expect(card).toHaveClass('border-l-healthy', 'bg-card')
  })

  it('applies correct status colors for warning status', () => {
    const { container } = render(
      <MetricCard title="Latency" value={150} unit="ms" status="warning" />
    )

    const card = container.querySelector('.metric-card')
    expect(card).toHaveClass('border-l-warning', 'bg-card')
  })

  it('applies correct status colors for critical status', () => {
    const { container } = render(
      <MetricCard title="Latency" value={500} unit="ms" status="critical" />
    )

    const card = container.querySelector('.metric-card')
    expect(card).toHaveClass('border-l-destructive', 'bg-card')
  })

  it('renders trend when provided', () => {
    render(
      <MetricCard
        title="Latency"
        value={45}
        unit="ms"
        trend={{ value: 5, isPositive: false }}
      />
    )

    expect(screen.getByText(/5%/)).toBeInTheDocument()
    expect(screen.getByText(/vs. previous period/)).toBeInTheDocument()
  })

  it('renders icon when provided', () => {
    const icon = <svg data-testid="test-icon" />
    render(<MetricCard title="Latency" value={45} unit="ms" icon={icon} />)

    expect(screen.getByTestId('test-icon')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <MetricCard title="Latency" value={45} unit="ms" className="custom-class" />
    )

    const card = container.querySelector('.metric-card')
    expect(card).toHaveClass('custom-class')
  })

  it('has proper ARIA labels', () => {
    render(<MetricCard title="Latency" value={45} unit="ms" />)

    expect(screen.getByLabelText('Latency metric')).toBeInTheDocument()
    expect(screen.getByLabelText('Latency value')).toBeInTheDocument()
    expect(screen.getByLabelText('unit')).toBeInTheDocument()
  })

  it('shows trend with positive change', () => {
    render(
      <MetricCard
        title="Latency"
        value={45}
        unit="ms"
        trend={{ value: 10, isPositive: true }}
      />
    )

    expect(screen.getByText('↑ 10%')).toBeInTheDocument()
  })

  it('shows trend with negative change', () => {
    render(
      <MetricCard
        title="Latency"
        value={45}
        unit="ms"
        trend={{ value: 10, isPositive: false }}
      />
    )

    expect(screen.getByText('↓ 10%')).toBeInTheDocument()
  })
})
