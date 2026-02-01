import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import TrendChart, { DataPoint, TimeRange, MetricType } from '../TrendChart'

describe('TrendChart', () => {
  const mockData: DataPoint[] = [
    { timestamp: '2024-01-01T00:00:00Z', value: 45 },
    { timestamp: '2024-01-01T01:00:00Z', value: 50 },
    { timestamp: '2024-01-01T02:00:00Z', value: 55 },
  ]

  it('renders chart container', () => {
    const { container } = render(
      <TrendChart data={mockData} metric="latency_ms" timeRange="24h" />
    )

    expect(container.querySelector('.trend-chart')).toBeInTheDocument()
    expect(screen.getByText('Latency Trend')).toBeInTheDocument()
  })

  it('renders time range selector', () => {
    render(<TrendChart data={mockData} metric="latency_ms" timeRange="24h" />)

    expect(screen.getByText('24 Hours')).toBeInTheDocument()
    expect(screen.getByText('7 Days')).toBeInTheDocument()
    expect(screen.getByText('30 Days')).toBeInTheDocument()
  })

  it('highlights active time range', () => {
    render(<TrendChart data={mockData} metric="latency_ms" timeRange="7d" />)

    const button7d = screen.getByText('7 Days')
    expect(button7d.closest('button')).toHaveClass('bg-blue-600', 'text-white')
  })

  it('calls onTimeRangeChange when time range button is clicked', async () => {
    const handleTimeRangeChange = vi.fn()

    render(
      <TrendChart
        data={mockData}
        metric="latency_ms"
        timeRange="24h"
        onTimeRangeChange={handleTimeRangeChange}
      />
    )

    const button30d = screen.getByText('30 Days')
    button30d.click()

    expect(handleTimeRangeChange).toHaveBeenCalledWith('30d')
  })

  it('renders empty state when no data', () => {
    render(<TrendChart data={[]} metric="latency_ms" timeRange="24h" />)

    expect(screen.getByText('No Data Available')).toBeInTheDocument()
    expect(
      screen.getByText('No trend data available for the selected time range.')
    ).toBeInTheDocument()
  })

  it('renders loading state', () => {
    render(
      <TrendChart data={mockData} metric="latency_ms" timeRange="24h" isLoading={true} />
    )

    expect(screen.getByText('Loading chart data...')).toBeInTheDocument()
  })

  it('displays baseline when showBaseline is true', () => {
    render(
      <TrendChart
        data={mockData}
        metric="latency_ms"
        timeRange="24h"
        showBaseline={true}
        baselineValue={50}
      />
    )

    expect(screen.getByText(/Baseline/)).toBeInTheDocument()
    expect(screen.getByText('50 ms')).toBeInTheDocument()
  })

  it('does not display baseline when showBaseline is false', () => {
    render(
      <TrendChart
        data={mockData}
        metric="latency_ms"
        timeRange="24h"
        showBaseline={false}
        baselineValue={50}
      />
    )

    expect(screen.queryByText(/Baseline/)).not.toBeInTheDocument()
  })

  it('renders different metric titles', () => {
    const { rerender } = render(
      <TrendChart data={mockData} metric="latency_ms" timeRange="24h" />
    )
    expect(screen.getByText('Latency Trend')).toBeInTheDocument()

    rerender(<TrendChart data={mockData} metric="packet_loss_rate" timeRange="24h" />)
    expect(screen.getByText('Packet Loss Rate Trend')).toBeInTheDocument()

    rerender(<TrendChart data={mockData} metric="jitter_ms" timeRange="24h" />)
    expect(screen.getByText('Jitter Trend')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <TrendChart
        data={mockData}
        metric="latency_ms"
        timeRange="24h"
        className="custom-class"
      />
    )

    const chart = container.querySelector('.trend-chart')
    expect(chart).toHaveClass('custom-class')
  })

  it('applies custom height', () => {
    const { container } = render(
      <TrendChart data={mockData} metric="latency_ms" timeRange="24h" height="500px" />
    )

    const chartContainer = container.querySelector('[style*="height"]')
    expect(chartContainer).toHaveStyle({ height: '500px' })
  })

  it('has proper ARIA attributes', () => {
    render(<TrendChart data={mockData} metric="latency_ms" timeRange="24h" />)

    const region = screen.getByRole('region', { name: 'Latency trend chart' })
    expect(region).toBeInTheDocument()

    const group = screen.getByRole('group', { name: 'Time range selector' })
    expect(group).toBeInTheDocument()
  })

  it('disables time range buttons when loading', () => {
    render(
      <TrendChart data={mockData} metric="latency_ms" timeRange="24h" isLoading={true} />
    )

    const buttons = screen.getAllByRole('button')
    buttons.forEach((button) => {
      expect(button).toBeDisabled()
    })
  })

  it('has aria-pressed on time range buttons', () => {
    render(<TrendChart data={mockData} metric="latency_ms" timeRange="24h" />)

    const button24h = screen.getByText('24 Hours')
    expect(button24h.closest('button')).toHaveAttribute('aria-pressed', 'true')

    const button7d = screen.getByText('7 Days')
    expect(button7d.closest('button')).toHaveAttribute('aria-pressed', 'false')
  })

  it('handles metric type changes correctly', () => {
    const metrics: MetricType[] = ['latency_ms', 'packet_loss_rate', 'jitter_ms']

    metrics.forEach((metric) => {
      const { unmount } = render(
        <TrendChart data={mockData} metric={metric} timeRange="24h" />
      )

      expect(screen.getByRole('region')).toBeInTheDocument()

      unmount()
    })
  })

  it('handles all time ranges', () => {
    const timeRanges: TimeRange[] = ['24h', '7d', '30d']

    timeRanges.forEach((timeRange) => {
      const { unmount } = render(
        <TrendChart data={mockData} metric="latency_ms" timeRange={timeRange} />
      )

      expect(screen.getByRole('region')).toBeInTheDocument()

      unmount()
    })
  })
})
