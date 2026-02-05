import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import {
  determineHealthStatus,
  isNodeOffline,
  DEFAULT_THRESHOLDS,
  type HealthStatus,
  type HealthThresholds,
  type NodeMetrics,
} from '../healthStatus'

describe('healthStatus utilities', () => {
  describe('isNodeOffline', () => {
    beforeEach(() => {
      // Mock current time to 2026-01-26 10:02:00 UTC
      vi.setSystemTime(new Date('2026-01-26T10:02:00Z'))
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('should return true for node with heartbeat >120 seconds ago', () => {
      const oldHeartbeat = '2026-01-26T09:59:00Z' // 3 minutes ago
      expect(isNodeOffline(oldHeartbeat)).toBe(true)
    })

    it('should return true for node with heartbeat exactly 120 seconds ago', () => {
      const thresholdHeartbeat = '2026-01-26T10:00:00Z' // exactly 120 seconds ago
      expect(isNodeOffline(thresholdHeartbeat)).toBe(true)
    })

    it('should return false for node with heartbeat <120 seconds ago', () => {
      const recentHeartbeat = '2026-01-26T10:01:00Z' // 60 seconds ago
      expect(isNodeOffline(recentHeartbeat)).toBe(false)
    })

    it('should return false for node with very recent heartbeat', () => {
      const veryRecentHeartbeat = '2026-01-26T10:01:55Z' // 5 seconds ago
      expect(isNodeOffline(veryRecentHeartbeat)).toBe(false)
    })

    it('should return true for missing or empty heartbeat', () => {
      expect(isNodeOffline('')).toBe(true)
      expect(isNodeOffline(undefined as any)).toBe(true)
    })
  })

  describe('determineHealthStatus', () => {
    beforeEach(() => {
      // Mock current time to 2026-01-26 10:02:00 UTC
      vi.setSystemTime(new Date('2026-01-26T10:02:00Z'))
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    describe('offline status', () => {
      it('should return offline for node with no heartbeat >120 seconds', () => {
        const metrics: NodeMetrics = {
          latency_ms: 50,
          packet_loss_rate: 0,
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T09:59:00Z', // >120 seconds ago
        }

        expect(determineHealthStatus(metrics)).toBe('offline')
      })

      it('should return offline for null or undefined metrics', () => {
        expect(determineHealthStatus(null as any)).toBe('offline')
        expect(determineHealthStatus(undefined as any)).toBe('offline')
      })

      it('should prioritize offline status over metric thresholds', () => {
        const metrics: NodeMetrics = {
          latency_ms: 50, // Good metrics
          packet_loss_rate: 0,
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T09:50:00Z', // Old heartbeat
        }

        expect(determineHealthStatus(metrics)).toBe('offline')
      })
    })

    describe('critical status', () => {
      it('should return critical when latency exceeds threshold', () => {
        const metrics: NodeMetrics = {
          latency_ms: 250, // Exceeds 200ms threshold
          packet_loss_rate: 0,
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T10:01:00Z', // Recent heartbeat
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('critical')
      })

      it('should return critical when packet loss exceeds threshold', () => {
        const metrics: NodeMetrics = {
          latency_ms: 50,
          packet_loss_rate: 10, // Exceeds 5% threshold
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('critical')
      })

      it('should return critical when jitter exceeds threshold', () => {
        const metrics: NodeMetrics = {
          latency_ms: 50,
          packet_loss_rate: 0,
          jitter_ms: 100, // Exceeds 50ms threshold
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('critical')
      })

      it('should return critical when multiple metrics exceed threshold', () => {
        const metrics: NodeMetrics = {
          latency_ms: 300,
          packet_loss_rate: 15,
          jitter_ms: 80,
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('critical')
      })
    })

    describe('warning status', () => {
      it('should return warning when latency is 80-100% of threshold', () => {
        const metrics: NodeMetrics = {
          latency_ms: 170, // 85% of 200ms threshold
          packet_loss_rate: 0,
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('warning')
      })

      it('should return warning when packet loss is 80-100% of threshold', () => {
        const metrics: NodeMetrics = {
          latency_ms: 50,
          packet_loss_rate: 4.2, // 84% of 5% threshold
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('warning')
      })

      it('should return warning when jitter is 80-100% of threshold', () => {
        const metrics: NodeMetrics = {
          latency_ms: 50,
          packet_loss_rate: 0,
          jitter_ms: 42, // 84% of 50ms threshold
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('warning')
      })

      it('should return warning when heartbeat is 60-120 seconds old', () => {
        const metrics: NodeMetrics = {
          latency_ms: 50, // Good metrics
          packet_loss_rate: 0,
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T10:00:30Z', // 90 seconds ago
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('warning')
      })
    })

    describe('healthy status', () => {
      it('should return healthy for all good metrics', () => {
        const metrics: NodeMetrics = {
          latency_ms: 50,
          packet_loss_rate: 0,
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T10:01:30Z', // Recent heartbeat
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('healthy')
      })

      it('should return healthy when metrics are below 80% of thresholds', () => {
        const metrics: NodeMetrics = {
          latency_ms: 150, // 75% of 200ms threshold
          packet_loss_rate: 3, // 60% of 5% threshold
          jitter_ms: 30, // 60% of 50ms threshold
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('healthy')
      })

      it('should return healthy for optimal metrics', () => {
        const metrics: NodeMetrics = {
          latency_ms: 20,
          packet_loss_rate: 0.1,
          jitter_ms: 5,
          last_heartbeat: '2026-01-26T10:01:55Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('healthy')
      })
    })

    describe('edge cases', () => {
      it('should handle missing metric values with defaults', () => {
        const metrics: NodeMetrics = {
          latency_ms: 0,
          packet_loss_rate: 0,
          jitter_ms: 0,
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('healthy')
      })

      it('should handle undefined metric values', () => {
        const metrics: NodeMetrics = {
          latency_ms: undefined as any,
          packet_loss_rate: undefined as any,
          jitter_ms: undefined as any,
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, DEFAULT_THRESHOLDS)).toBe('healthy')
      })

      it('should use default thresholds when not provided', () => {
        const metrics: NodeMetrics = {
          latency_ms: 250,
          packet_loss_rate: 0,
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        // Should use DEFAULT_THRESHOLDS
        expect(determineHealthStatus(metrics)).toBe('critical')
      })

      it('should accept custom thresholds', () => {
        const customThresholds: HealthThresholds = {
          latency: 100,
          packetLoss: 2,
          jitter: 30,
        }

        const metrics: NodeMetrics = {
          latency_ms: 90, // Would be healthy with default, critical with custom
          packet_loss_rate: 0,
          jitter_ms: 10,
          last_heartbeat: '2026-01-26T10:01:00Z',
        }

        expect(determineHealthStatus(metrics, customThresholds)).toBe('warning')
      })
    })
  })

  describe('DEFAULT_THRESHOLDS', () => {
    it('should have correct default values', () => {
      expect(DEFAULT_THRESHOLDS.latency).toBe(200)
      expect(DEFAULT_THRESHOLDS.packetLoss).toBe(5)
      expect(DEFAULT_THRESHOLDS.jitter).toBe(50)
    })
  })
})
