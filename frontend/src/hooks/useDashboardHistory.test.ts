import { describe, it, expect } from 'vitest'
import { aggregateHistoryAverageByTimestamp } from './useDashboardHistory'
import type { HistoryDataDTO } from '../api/types'

describe('aggregateHistoryAverageByTimestamp', () => {
  it('averages values that share the same timestamp across nodes', () => {
    const series: HistoryDataDTO[] = [
      {
        node_id: 'a',
        metric: 'latency',
        data_points: [
          { timestamp: '2026-01-01T00:00:00Z', value: 10 },
          { timestamp: '2026-01-01T00:05:00Z', value: 20 },
        ],
      },
      {
        node_id: 'b',
        metric: 'latency',
        data_points: [
          { timestamp: '2026-01-01T00:00:00Z', value: 30 },
          { timestamp: '2026-01-01T00:05:00Z', value: 40 },
        ],
      },
    ]
    const out = aggregateHistoryAverageByTimestamp(series)
    expect(out).toEqual([
      { timestamp: '2026-01-01T00:00:00Z', value: 20 },
      { timestamp: '2026-01-01T00:05:00Z', value: 30 },
    ])
  })

  it('returns empty array for empty input', () => {
    expect(aggregateHistoryAverageByTimestamp([])).toEqual([])
  })
})
