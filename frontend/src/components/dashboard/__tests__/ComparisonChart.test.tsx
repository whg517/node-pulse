import { render, screen, fireEvent, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import ComparisonChart, {
  ComparisonDataPoint,
  NodeComparisonData,
  type TimeRange,
  type MetricType,
  type GroupByType,
} from '../ComparisonChart'

// Mock ECharts (ComparisonChart imports from ../../../lib/echarts-core, not echarts directly)
vi.mock('../../../lib/echarts-core', () => ({
  default: {
    init: vi.fn(() => ({
      setOption: vi.fn(),
      dispose: vi.fn(),
      resize: vi.fn(),
      on: vi.fn(),
      off: vi.fn(),
    })),
    getMap: vi.fn(() => null),
    registerMap: vi.fn(),
  },
  graphic: {
    LinearGradient: vi.fn(),
  },
}))

describe('ComparisonChart', () => {
  const mockData: NodeComparisonData[] = [
    {
      node_id: 'node-1',
      node_name: 'Node 1',
      region: 'US-East',
      isp: 'AWS',
      data: [
        { timestamp: '2024-01-01T00:00:00Z', value: 100 },
        { timestamp: '2024-01-01T01:00:00Z', value: 110 },
        { timestamp: '2024-01-01T02:00:00Z', value: 105 },
      ],
    },
    {
      node_id: 'node-2',
      node_name: 'Node 2',
      region: 'US-West',
      isp: 'GCP',
      data: [
        { timestamp: '2024-01-01T00:00:00Z', value: 95 },
        { timestamp: '2024-01-01T01:00:00Z', value: 105 },
        { timestamp: '2024-01-01T02:00:00Z', value: 100 },
      ],
    },
  ]

  const defaultProps = {
    nodes: mockData,
    metric: 'latency_ms' as MetricType,
    timeRange: '24h' as TimeRange,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('renders chart with multiple nodes', () => {
      render(<ComparisonChart {...defaultProps} />)
      expect(screen.getByText('Latency Comparison')).toBeInTheDocument()
      expect(screen.getByText('Node 1')).toBeInTheDocument()
      expect(screen.getByText('Node 2')).toBeInTheDocument()
    })

    it('renders with loading state', () => {
      render(<ComparisonChart {...defaultProps} isLoading={true} />)
      expect(screen.getByText('Loading chart data...')).toBeInTheDocument()
    })

    it('renders empty state when no data', () => {
      render(<ComparisonChart {...defaultProps} nodes={[]} />)
      // When nodes array is empty, the component renders but shows empty state in chart area
      expect(screen.getByText('Latency Comparison')).toBeInTheDocument()
    })

    it('renders time range buttons', () => {
      render(<ComparisonChart {...defaultProps} />)
      expect(screen.getByText('24 Hours')).toBeInTheDocument()
      expect(screen.getByText('7 Days')).toBeInTheDocument()
      expect(screen.getByText('30 Days')).toBeInTheDocument()
    })

    it('renders statistics panel when enabled', () => {
      render(<ComparisonChart {...defaultProps} showStatistics={true} />)
      expect(screen.getByText('Average')).toBeInTheDocument()
      expect(screen.getByText('Maximum')).toBeInTheDocument()
      expect(screen.getByText('Minimum')).toBeInTheDocument()
      expect(screen.getByText('Difference')).toBeInTheDocument()
    })

    it('does not render statistics panel when disabled', () => {
      render(<ComparisonChart {...defaultProps} showStatistics={false} />)
      expect(screen.queryByText('Average')).not.toBeInTheDocument()
      expect(screen.queryByText('Maximum')).not.toBeInTheDocument()
    })
  })

  describe('validation', () => {
    it('validates max 5 nodes constraint', () => {
      const tooManyNodes = Array.from({ length: 6 }, (_, i) => ({
        node_id: `node-${i}`,
        node_name: `Node ${i}`,
        data: [{ timestamp: '2024-01-01T00:00:00Z', value: 100 }],
      }))

      const consoleWarn = vi.spyOn(console, 'warn').mockImplementation(() => {})
      render(<ComparisonChart {...defaultProps} nodes={tooManyNodes} />)
      expect(consoleWarn).toHaveBeenCalledWith('ComparisonChart requires 2-5 nodes')
      consoleWarn.mockRestore()
    })

    it('validates min 2 nodes requirement', () => {
      const consoleWarn = vi.spyOn(console, 'warn').mockImplementation(() => {})
      render(
        <ComparisonChart
          {...defaultProps}
          nodes={[
            {
              node_id: 'node-1',
              node_name: 'Node 1',
              data: [{ timestamp: '2024-01-01T00:00:00Z', value: 100 }],
            },
          ]}
        />
      )
      expect(consoleWarn).toHaveBeenCalledWith('ComparisonChart requires 2-5 nodes')
      consoleWarn.mockRestore()
    })

    it('accepts exactly 2 nodes', () => {
      const { container } = render(<ComparisonChart {...defaultProps} nodes={mockData.slice(0, 2)} />)
      expect(container.querySelector('.comparison-chart')).toBeInTheDocument()
    })

    it('accepts exactly 5 nodes', () => {
      const fiveNodes = Array.from({ length: 5 }, (_, i) => ({
        node_id: `node-${i}`,
        node_name: `Node ${i}`,
        data: [{ timestamp: '2024-01-01T00:00:00Z', value: 100 }],
      }))
      const { container } = render(<ComparisonChart {...defaultProps} nodes={fiveNodes} />)
      expect(container.querySelector('.comparison-chart')).toBeInTheDocument()
    })
  })

  describe('statistics calculation', () => {
    it('calculates average correctly', () => {
      render(<ComparisonChart {...defaultProps} showStatistics={true} />)
      // Average of [100, 110, 105] and [95, 105, 100] = 105 and 100
      // Overall average = (100 + 110 + 105 + 95 + 105 + 100) / 6 = 102.5
      expect(screen.getByText(/102\.5/)).toBeInTheDocument()
    })

    it('calculates maximum correctly', () => {
      render(<ComparisonChart {...defaultProps} showStatistics={true} />)
      expect(screen.getByText(/110\.00/)).toBeInTheDocument()
    })

    it('calculates minimum correctly', () => {
      render(<ComparisonChart {...defaultProps} showStatistics={true} />)
      expect(screen.getByText(/95\.00/)).toBeInTheDocument()
    })

    it('calculates difference correctly', () => {
      render(<ComparisonChart {...defaultProps} showStatistics={true} />)
      // Max = 110, Min = 95, Diff = 15
      expect(screen.getByText(/15\.00/)).toBeInTheDocument()
    })
  })

  describe('grouping', () => {
    it('groups nodes by region', () => {
      render(<ComparisonChart {...defaultProps} groupBy="region" />)
      expect(screen.getByText('Node 1 (US-East)')).toBeInTheDocument()
      expect(screen.getByText('Node 2 (US-West)')).toBeInTheDocument()
    })

    it('groups nodes by ISP', () => {
      render(<ComparisonChart {...defaultProps} groupBy="isp" />)
      expect(screen.getByText('Node 1 (AWS)')).toBeInTheDocument()
      expect(screen.getByText('Node 2 (GCP)')).toBeInTheDocument()
    })

    it('does not group when groupBy is none', () => {
      render(<ComparisonChart {...defaultProps} groupBy="none" />)
      expect(screen.getByText('Node 1')).toBeInTheDocument()
      expect(screen.getByText('Node 2')).toBeInTheDocument()
      expect(screen.queryByText('Node 1 (US-East)')).not.toBeInTheDocument()
    })

    it('handles missing region/ISP tags gracefully', () => {
      const dataWithoutTags: NodeComparisonData[] = [
        {
          node_id: 'node-1',
          node_name: 'Node 1',
          data: [{ timestamp: '2024-01-01T00:00:00Z', value: 100 }],
        },
        {
          node_id: 'node-2',
          node_name: 'Node 2',
          data: [{ timestamp: '2024-01-01T00:00:00Z', value: 100 }],
        },
      ]
      render(<ComparisonChart {...defaultProps} nodes={dataWithoutTags} groupBy="region" />)
      expect(screen.getByText('Node 1')).toBeInTheDocument()
      expect(screen.getByText('Node 2')).toBeInTheDocument()
    })
  })

  describe('metrics', () => {
    it('renders latency metric correctly', () => {
      render(<ComparisonChart {...defaultProps} metric="latency_ms" />)
      expect(screen.getByText('Latency Comparison')).toBeInTheDocument()
    })

    it('renders packet loss metric correctly', () => {
      render(<ComparisonChart {...defaultProps} metric="packet_loss_rate" />)
      expect(screen.getByText('Packet Loss Rate Comparison')).toBeInTheDocument()
    })

    it('renders jitter metric correctly', () => {
      render(<ComparisonChart {...defaultProps} metric="jitter_ms" />)
      expect(screen.getByText('Jitter Comparison')).toBeInTheDocument()
    })
  })

  describe('interactions', () => {
    it('handles time range change', () => {
      const handleChange = vi.fn()
      render(<ComparisonChart {...defaultProps} onTimeRangeChange={handleChange} />)

      const button7d = screen.getByText('7 Days')

      act(() => {
        fireEvent.click(button7d)
      })

      // Verify the callback was triggered
      expect(handleChange).toHaveBeenCalledWith('7d')
    })

    it('disables time range buttons when loading', () => {
      render(<ComparisonChart {...defaultProps} isLoading={true} />)

      const buttons = screen.getAllByRole('button')
      buttons.forEach((button) => {
        expect(button).toBeDisabled()
      })
    })
  })

  describe('accessibility', () => {
    it('has proper ARIA labels', () => {
      render(<ComparisonChart {...defaultProps} />)
      const chart = screen.getByRole('region', { name: 'Latency comparison chart' })
      expect(chart).toBeInTheDocument()
    })

    it('has proper aria-pressed on time range buttons', () => {
      render(<ComparisonChart {...defaultProps} />)

      const button24h = screen.getByText('24 Hours').closest('button')
      expect(button24h).toHaveAttribute('aria-pressed', 'true')
    })

    it('has proper role for loading spinner', () => {
      render(<ComparisonChart {...defaultProps} isLoading={true} />)
      const spinner = screen.getByRole('status', { name: 'Loading chart data' })
      expect(spinner).toBeInTheDocument()
    })
  })

  describe('difference highlighting', () => {
    it('renders with highlightDifferences enabled', () => {
      render(<ComparisonChart {...defaultProps} highlightDifferences={true} />)
      expect(screen.getByText('Latency Comparison')).toBeInTheDocument()
    })

    it('renders with highlightDifferences disabled', () => {
      render(<ComparisonChart {...defaultProps} highlightDifferences={false} />)
      expect(screen.getByText('Latency Comparison')).toBeInTheDocument()
    })

    it('calculates statistics with node identification', () => {
      render(<ComparisonChart {...defaultProps} showStatistics={true} />)

      // Should show statistics panel
      expect(screen.getByText('Average')).toBeInTheDocument()
      expect(screen.getByText('Maximum')).toBeInTheDocument()
      expect(screen.getByText('Minimum')).toBeInTheDocument()
    })
  })

  describe('edge cases', () => {
    it('handles nodes with different time ranges', () => {
      const differentTimeRanges: NodeComparisonData[] = [
        {
          node_id: 'node-1',
          node_name: 'Node 1',
          data: [
            { timestamp: '2024-01-01T00:00:00Z', value: 100 },
            { timestamp: '2024-01-01T01:00:00Z', value: 110 },
          ],
        },
        {
          node_id: 'node-2',
          node_name: 'Node 2',
          data: [
            { timestamp: '2024-01-01T01:00:00Z', value: 105 },
            { timestamp: '2024-01-01T02:00:00Z', value: 100 },
          ],
        },
      ]
      const { container } = render(
        <ComparisonChart {...defaultProps} nodes={differentTimeRanges} />
      )
      expect(container.querySelector('.comparison-chart')).toBeInTheDocument()
    })

    it('handles nodes with missing data points', () => {
      const missingData: NodeComparisonData[] = [
        {
          node_id: 'node-1',
          node_name: 'Node 1',
          data: [
            { timestamp: '2024-01-01T00:00:00Z', value: 100 },
            { timestamp: '2024-01-01T01:00:00Z', value: 110 },
          ],
        },
        {
          node_id: 'node-2',
          node_name: 'Node 2',
          data: [
            { timestamp: '2024-01-01T00:00:00Z', value: 95 },
            // Missing second data point
          ],
        },
      ]
      const { container } = render(<ComparisonChart {...defaultProps} nodes={missingData} />)
      expect(container.querySelector('.comparison-chart')).toBeInTheDocument()
    })

    it('handles custom height', () => {
      const { container } = render(<ComparisonChart {...defaultProps} height="600px" />)
      const chartContainer = container.querySelector('.comparison-chart .relative')
      expect(chartContainer).toHaveStyle({ height: '600px' })
    })

    it('handles custom className', () => {
      const { container } = render(
        <ComparisonChart {...defaultProps} className="custom-class" />
      )
      const chart = container.querySelector('.comparison-chart')
      expect(chart).toHaveClass('custom-class')
    })
  })
})
