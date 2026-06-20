import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import { ReportGenerator } from './ReportGenerator'
import { fetchHistory, fetchLatestMTR, fetchMetrics } from '@/api/data'
import type { NodeDTO } from '@/api/types'

vi.mock('@/api/data', () => ({
  fetchHistory: vi.fn(() => Promise.resolve({ data: [] })),
  fetchLatestMTR: vi.fn(() => Promise.resolve(null)),
  fetchMetrics: vi.fn(() => Promise.resolve({ data: [] })),
}))

vi.mock('./HealthReportPDF', () => ({
  HealthReportPDF: ({
    metrics,
    mtrPath,
    timeline,
  }: {
    metrics: { latency: { current: number }; packetLoss: { current: number }; jitter: { current: number } }
    mtrPath?: Array<{ hop: number; ip: string; location?: string; avgLatency: number; lossRate: number }>
    timeline?: Array<{ event: string; severity: string }>
  }) => (
    <div data-testid="pdf-preview">
      <div>{metrics.latency.current} ms latency</div>
      <div>{metrics.packetLoss.current}% packet loss</div>
      <div>{metrics.jitter.current} ms jitter</div>
      {(timeline || []).map((event) => (
        <div key={event.event}>
          <span>{event.event}</span>
          <span>{event.severity}</span>
        </div>
      ))}
      {(mtrPath || []).map((hop) => (
        <div key={hop.hop}>
          <span>{hop.ip}</span>
          <span>{hop.location}</span>
          <span>{hop.avgLatency.toFixed(1)} ms</span>
          <span>{hop.lossRate.toFixed(1)}% loss</span>
        </div>
      ))}
    </div>
  ),
}))

const mockFetchHistory = fetchHistory as ReturnType<typeof vi.fn>
const mockFetchLatestMTR = fetchLatestMTR as ReturnType<typeof vi.fn>
const mockFetchMetrics = fetchMetrics as ReturnType<typeof vi.fn>

const nodes: NodeDTO[] = [
  {
    id: 'node-1',
    name: 'Singapore Edge',
    ip: '192.0.2.10',
    region: 'ap-southeast',
    tags: ['edge'],
    status: 'online',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
]

describe('ReportGenerator', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetchHistory.mockResolvedValue({ data: [] })
    mockFetchLatestMTR.mockResolvedValue(null)
    mockFetchMetrics.mockResolvedValue({ data: [] })
  })

  it('uses live metrics and latest MTR data in PDF preview', async () => {
    mockFetchMetrics.mockResolvedValue({
      data: [
        {
          node_id: 'node-1',
          latency_ms: 88,
          packet_loss_rate: 2.5,
          jitter_ms: 14,
          timestamp: '2024-01-01T12:00:00Z',
        },
      ],
    })
    mockFetchHistory.mockImplementation((query: { metrics: string[] }) => {
      const metric = query.metrics[0]
      return Promise.resolve({
        data: [
          {
            node_id: 'node-1',
            metric,
            data_points: [
              { timestamp: '2024-01-01T11:00:00Z', value: metric === 'latency' ? 42 : metric === 'packet_loss_rate' ? 0.2 : 7 },
              { timestamp: '2024-01-01T12:00:00Z', value: metric === 'latency' ? 520 : metric === 'packet_loss_rate' ? 6 : 120 },
            ],
          },
        ],
      })
    })
    mockFetchLatestMTR.mockResolvedValue({
      target: 'example.com',
      totalHops: 2,
      completedAt: '2024-01-01T12:00:00Z',
      success: true,
      hops: [
        {
          hopNumber: 1,
          ip: '192.0.2.1',
          hostname: 'gateway.local',
          sent: 10,
          received: 10,
          lossRate: 0,
          lastRTTMs: 1.2,
          avgRTTMs: 1.4,
          bestRTTMs: 1.1,
          worstRTTMs: 1.8,
          stdDevMs: 0.2,
        },
        {
          hopNumber: 2,
          ip: '198.51.100.1',
          location: 'Regional Hub',
          sent: 10,
          received: 9,
          lossRate: 10,
          lastRTTMs: 32.4,
          avgRTTMs: 35.1,
          bestRTTMs: 31.8,
          worstRTTMs: 42.7,
          stdDevMs: 3.3,
        },
      ],
    })

    render(
      <ReportGenerator
        nodes={nodes}
        onSubmit={vi.fn()}
        defaultNodeIds={['node-1']}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: 'PDF' }))
    fireEvent.click(screen.getByRole('button', { name: 'Generate Report' }))

    await waitFor(() => {
      expect(mockFetchMetrics).toHaveBeenCalledWith(['node-1'])
      expect(mockFetchLatestMTR).toHaveBeenCalledWith('node-1')
    })
    expect(mockFetchHistory).toHaveBeenCalledTimes(3)

    expect(await screen.findByTestId('pdf-preview')).toBeInTheDocument()
    expect(screen.getByText('88 ms latency')).toBeInTheDocument()
    expect(screen.getByText('2.5% packet loss')).toBeInTheDocument()
    expect(screen.getByText('14 ms jitter')).toBeInTheDocument()
    expect(screen.getByText('Latency threshold exceeded (520.0ms)')).toBeInTheDocument()
    expect(screen.getByText('Packet loss threshold exceeded (6.0%)')).toBeInTheDocument()
    expect(screen.getByText('Jitter threshold exceeded (120.0ms)')).toBeInTheDocument()
    expect(screen.getAllByText('critical').length).toBeGreaterThan(0)
    expect(screen.getByText('192.0.2.1')).toBeInTheDocument()
    expect(screen.getByText('gateway.local')).toBeInTheDocument()
    expect(screen.getByText('198.51.100.1')).toBeInTheDocument()
    expect(screen.getByText('Regional Hub')).toBeInTheDocument()
    expect(screen.queryByText('203.0.113.1')).not.toBeInTheDocument()
  })
})
