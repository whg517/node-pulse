import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useTimezone } from '../useTimezone'

// Mock settingsStore
const mockSetTimezone = vi.fn()
const mockSetDisplayMode = vi.fn()

vi.mock('../../stores/settingsStore', () => ({
  useSettingsStore: vi.fn((selector: (s: unknown) => unknown) =>
    selector({
      timezone: 'UTC',
      timezoneDisplayMode: 'local',
      setTimezone: mockSetTimezone,
      setTimezoneDisplayMode: mockSetDisplayMode,
    })
  ),
  COMMON_TIMEZONES: [
    { value: 'UTC', label: 'UTC', offset: '+00:00' },
    { value: 'Asia/Shanghai', label: 'Beijing, Shanghai', offset: '+08:00' },
    { value: 'America/New_York', label: 'New York', offset: '-05:00' },
  ],
}))

describe('useTimezone', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns current timezone', () => {
    const { result } = renderHook(() => useTimezone())
    expect(result.current.timezone).toBe('UTC')
  })

  it('returns current display mode', () => {
    const { result } = renderHook(() => useTimezone())
    expect(result.current.displayMode).toBe('local')
  })

  it('returns timezones list', () => {
    const { result } = renderHook(() => useTimezone())
    expect(result.current.timezones.length).toBe(3)
  })

  it('finds current timezone option', () => {
    const { result } = renderHook(() => useTimezone())
    expect(result.current.currentTimezone).toEqual({
      value: 'UTC',
      label: 'UTC',
      offset: '+00:00',
    })
  })

  it('calls setTimezone', () => {
    const { result } = renderHook(() => useTimezone())
    act(() => {
      result.current.setTimezone('Asia/Shanghai')
    })
    expect(mockSetTimezone).toHaveBeenCalledWith('Asia/Shanghai')
  })

  it('calls setDisplayMode', () => {
    const { result } = renderHook(() => useTimezone())
    act(() => {
      result.current.setDisplayMode('utc')
    })
    expect(mockSetDisplayMode).toHaveBeenCalledWith('utc')
  })

  describe('formatTime', () => {
    it('formats a date string', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTime('2024-01-15T10:30:00Z')
      expect(typeof formatted).toBe('string')
      expect(formatted).not.toBe('Invalid Date')
    })

    it('formats a Date object', () => {
      const { result } = renderHook(() => useTimezone())
      const date = new Date('2024-01-15T10:30:00Z')
      const formatted = result.current.formatTime(date)
      expect(typeof formatted).toBe('string')
    })

    it('formats a numeric timestamp', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTime(1705312200000)
      expect(typeof formatted).toBe('string')
    })

    it('returns Invalid Date for bad input', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTime('not-a-date')
      expect(formatted).toBe('Invalid Date')
    })

    it('formats with showSeconds option', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTime('2024-01-15T10:30:00Z', { showSeconds: true })
      expect(typeof formatted).toBe('string')
    })

    it('formats with showDate only', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTime('2024-01-15T10:30:00Z', {
        showDate: true,
        showTime: false,
      })
      expect(typeof formatted).toBe('string')
    })

    it('formats with short format', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTime('2024-01-15T10:30:00Z', { format: 'short' })
      expect(typeof formatted).toBe('string')
    })

    it('formats with long format', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTime('2024-01-15T10:30:00Z', { format: 'long' })
      expect(typeof formatted).toBe('string')
    })

    it('formats with showTimezone option', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTime('2024-01-15T10:30:00Z', { showTimezone: true })
      expect(typeof formatted).toBe('string')
    })
  })

  describe('formatTimeWithTimezone', () => {
    it('formats time with specific timezone', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTimeWithTimezone(
        '2024-01-15T10:30:00Z',
        'Asia/Tokyo'
      )
      expect(typeof formatted).toBe('string')
      expect(formatted).not.toBe('Invalid Date')
    })

    it('returns Invalid Date for bad input', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTimeWithTimezone('bad-date', 'UTC')
      expect(formatted).toBe('Invalid Date')
    })

    it('supports showSeconds option', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTimeWithTimezone(
        '2024-01-15T10:30:00Z',
        'UTC',
        { showSeconds: true }
      )
      expect(typeof formatted).toBe('string')
    })

    it('supports showTimezone option', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTimeWithTimezone(
        '2024-01-15T10:30:00Z',
        'UTC',
        { showTimezone: true }
      )
      expect(typeof formatted).toBe('string')
    })

    it('supports showDate false', () => {
      const { result } = renderHook(() => useTimezone())
      const formatted = result.current.formatTimeWithTimezone(
        '2024-01-15T10:30:00Z',
        'UTC',
        { showDate: false, showTime: true }
      )
      expect(typeof formatted).toBe('string')
    })
  })

  describe('convertToTimezone', () => {
    it('converts date to another timezone', () => {
      const { result } = renderHook(() => useTimezone())
      const converted = result.current.convertToTimezone('2024-01-15T10:30:00Z', 'Asia/Tokyo')
      expect(converted instanceof Date).toBe(true)
    })

    it('handles Date object input', () => {
      const { result } = renderHook(() => useTimezone())
      const date = new Date('2024-01-15T10:30:00Z')
      const converted = result.current.convertToTimezone(date, 'UTC')
      expect(converted instanceof Date).toBe(true)
    })

    it('handles numeric timestamp input', () => {
      const { result } = renderHook(() => useTimezone())
      const converted = result.current.convertToTimezone(1705312200000, 'America/New_York')
      expect(converted instanceof Date).toBe(true)
    })
  })

  describe('getOffsetString', () => {
    it('returns offset string for UTC', () => {
      const { result } = renderHook(() => useTimezone())
      const offset = result.current.getOffsetString('UTC')
      expect(typeof offset).toBe('string')
    })

    it('returns offset string for non-UTC timezone', () => {
      const { result } = renderHook(() => useTimezone())
      const offset = result.current.getOffsetString('America/New_York')
      expect(typeof offset).toBe('string')
    })
  })
})
