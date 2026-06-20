import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import { ReportGenerator } from './ReportGenerator'
import { fetchLatestMTR } from '@/api/data'
import type { NodeDTO } from '@/api/types'

vi.mock('@/api/data', () => ({
  fetchLatestMTR: vi.fn(() => Promise.resolve(null)),
}))

vi.mock('./HealthReportPDF', () => ({
  HealthReportPDF: ({ mtrPath }: { mtrPath?: Array<{ hop: number; ip: string; location?: string; avgLatency: number; lossRate: number }> }) => (
    <div data-testid="pdf-preview">
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

const mockFetchLatestMTR = fetchLatestMTR as ReturnType<typeof vi.fn>

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
    mockFetchLatestMTR.mockResolvedValue(null)
  })

  it('uses latest MTR data in PDF preview', async () => {
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
      expect(mockFetchLatestMTR).toHaveBeenCalledWith('node-1')
    })

    expect(await screen.findByTestId('pdf-preview')).toBeInTheDocument()
    expect(screen.getByText('192.0.2.1')).toBeInTheDocument()
    expect(screen.getByText('gateway.local')).toBeInTheDocument()
    expect(screen.getByText('198.51.100.1')).toBeInTheDocument()
    expect(screen.getByText('Regional Hub')).toBeInTheDocument()
    expect(screen.queryByText('203.0.113.1')).not.toBeInTheDocument()
  })
})
