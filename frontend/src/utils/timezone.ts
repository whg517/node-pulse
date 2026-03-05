/**
 * Timezone Conversion Utilities
 * 
 * Provides timezone conversion and formatting helpers for displaying
 * timestamps in UTC and local timezones with proper formatting.
 * 
 * Features:
 * - UTC time display
 * - Local timezone conversion
 * - Custom timezone conversion
 * - Relative time formatting
 * - ISO 8601 parsing
 * - Cross-team collaboration with shared UTC timeline
 */

// ============== Types ==============

export type TimezoneMode = 'utc' | 'local' | 'custom'

export interface TimezoneOptions {
  mode?: TimezoneMode
  customTimezone?: string
  showSeconds?: boolean
  showDate?: boolean
  showTime?: boolean
  showTimezone?: boolean
  format?: 'short' | 'medium' | 'long'
}

export interface FormattedTime {
  utc: string
  local: string
  custom: string
  relative: string
  timestamp: number
}

// ============== Constants ==============

const DEFAULT_OPTIONS: Required<TimezoneOptions> = {
  mode: 'utc',
  customTimezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  showSeconds: false,
  showDate: true,
  showTime: true,
  showTimezone: true,
  format: 'medium',
}

// Common timezones for selection
export const COMMON_TIMEZONES = [
  { value: 'UTC', label: 'UTC (Coordinated Universal Time)', offset: '+00:00' },
  { value: 'America/New_York', label: 'Eastern Time (ET)', offset: '-05:00' },
  { value: 'America/Chicago', label: 'Central Time (CT)', offset: '-06:00' },
  { value: 'America/Denver', label: 'Mountain Time (MT)', offset: '-07:00' },
  { value: 'America/Los_Angeles', label: 'Pacific Time (PT)', offset: '-08:00' },
  { value: 'Europe/London', label: 'Greenwich Mean Time (GMT)', offset: '+00:00' },
  { value: 'Europe/Paris', label: 'Central European Time (CET)', offset: '+01:00' },
  { value: 'Europe/Berlin', label: 'Central European Time (CET)', offset: '+01:00' },
  { value: 'Europe/Moscow', label: 'Moscow Standard Time (MSK)', offset: '+03:00' },
  { value: 'Asia/Dubai', label: 'Gulf Standard Time (GST)', offset: '+04:00' },
  { value: 'Asia/Kolkata', label: 'India Standard Time (IST)', offset: '+05:30' },
  { value: 'Asia/Bangkok', label: 'Indochina Time (ICT)', offset: '+07:00' },
  { value: 'Asia/Singapore', label: 'Singapore Time (SGT)', offset: '+08:00' },
  { value: 'Asia/Shanghai', label: 'China Standard Time (CST)', offset: '+08:00' },
  { value: 'Asia/Tokyo', label: 'Japan Standard Time (JST)', offset: '+09:00' },
  { value: 'Asia/Seoul', label: 'Korea Standard Time (KST)', offset: '+09:00' },
  { value: 'Australia/Sydney', label: 'Australian Eastern Time (AET)', offset: '+10:00' },
  { value: 'Pacific/Auckland', label: 'New Zealand Standard Time (NZST)', offset: '+12:00' },
]

// ============== Utility Functions ==============

/**
 * Parse ISO 8601 timestamp to Date object
 */
export function parseTimestamp(timestamp: string): Date {
  const date = new Date(timestamp)
  if (isNaN(date.getTime())) {
    throw new Error(`Invalid timestamp: ${timestamp}`)
  }
  return date
}

/**
 * Get timezone offset string (e.g., "+08:00")
 */
export function getTimezoneOffset(timezone: string): string {
  try {
    const now = new Date()
    const formatter = new Intl.DateTimeFormat('en-US', {
      timeZone: timezone,
      timeZoneName: 'shortOffset',
    })
    const parts = formatter.formatToParts(now)
    const offsetPart = parts.find((p) => p.type === 'timeZoneName')
    return offsetPart?.value || '+00:00'
  } catch {
    return '+00:00'
  }
}

/**
 * Format date to UTC string
 */
export function formatUTC(
  date: Date | string,
  options?: Partial<TimezoneOptions>
): string {
  const d = typeof date === 'string' ? parseTimestamp(date) : date
  const opts = { ...DEFAULT_OPTIONS, ...options }
  
  const formatOpts: Intl.DateTimeFormatOptions = {
    timeZone: 'UTC',
    hour12: false,
  }
  
  if (opts.showDate) {
    formatOpts.year = 'numeric'
    formatOpts.month = '2-digit'
    formatOpts.day = '2-digit'
  }
  
  if (opts.showDate || opts.showTime !== false) {
    formatOpts.hour = '2-digit'
    formatOpts.minute = '2-digit'
    if (opts.showSeconds) {
      formatOpts.second = '2-digit'
    }
  }
  
  if (opts.showTimezone) {
    formatOpts.timeZoneName = 'short'
  }
  
  return d.toLocaleString(undefined, formatOpts)
}

/**
 * Format date to local timezone string
 */
export function formatLocal(
  date: Date | string,
  timezone?: string,
  options?: Partial<TimezoneOptions>
): string {
  const d = typeof date === 'string' ? parseTimestamp(date) : date
  const opts = { ...DEFAULT_OPTIONS, ...options }
  const tz = timezone || opts.customTimezone
  
  const formatOpts: Intl.DateTimeFormatOptions = {
    timeZone: tz,
    hour12: false,
  }
  
  if (opts.showDate) {
    formatOpts.year = 'numeric'
    formatOpts.month = '2-digit'
    formatOpts.day = '2-digit'
  }
  
  if (opts.showDate || opts.showTime !== false) {
    formatOpts.hour = '2-digit'
    formatOpts.minute = '2-digit'
    if (opts.showSeconds) {
      formatOpts.second = '2-digit'
    }
  }
  
  if (opts.showTimezone) {
    formatOpts.timeZoneName = 'short'
  }
  
  return d.toLocaleString(undefined, formatOpts)
}

/**
 * Format date to specific timezone string
 */
export function formatTimezone(
  date: Date | string,
  timezone: string,
  options?: Partial<TimezoneOptions>
): string {
  return formatLocal(date, timezone, options)
}

/**
 * Format relative time (e.g., "2 minutes ago")
 */
export function formatRelative(date: Date | string): string {
  const d = typeof date === 'string' ? parseTimestamp(date) : date
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)
  const diffWeeks = Math.floor(diffDays / 7)
  const diffMonths = Math.floor(diffDays / 30)
  
  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  if (diffWeeks < 4) return `${diffWeeks}w ago`
  if (diffMonths < 12) return `${diffMonths}mo ago`
  return `${Math.floor(diffMonths / 12)}y ago`
}

/**
 * Get complete formatted time information
 */
export function getFormattedTime(
  date: Date | string,
  options?: Partial<TimezoneOptions>
): FormattedTime {
  const d = typeof date === 'string' ? parseTimestamp(date) : date
  
  return {
    utc: formatUTC(d, options),
    local: formatLocal(d, undefined, options),
    custom: formatTimezone(d, options?.customTimezone || DEFAULT_OPTIONS.customTimezone, options),
    relative: formatRelative(d),
    timestamp: d.getTime(),
  }
}

/**
 * Convert date to specific timezone (for display purposes)
 * Returns adjusted timestamp for display
 */
export function convertToTimezone(
  date: Date | string,
  targetTimezone: string
): Date {
  const d = typeof date === 'string' ? parseTimestamp(date) : date
  
  // Create a date string in the target timezone
  const targetString = d.toLocaleString('en-US', { timeZone: targetTimezone })
  return new Date(targetString)
}

/**
 * Get time difference between two timestamps
 */
export function getTimeDifference(
  from: Date | string,
  to: Date | string = new Date()
): {
  ms: number
  seconds: number
  minutes: number
  hours: number
  days: number
  humanReadable: string
} {
  const fromDate = typeof from === 'string' ? parseTimestamp(from) : from
  const toDate = typeof to === 'string' ? parseTimestamp(to) : to
  const diffMs = toDate.getTime() - fromDate.getTime()
  
  const seconds = Math.floor(diffMs / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  let humanReadable = ''
  if (days > 0) {
    humanReadable = `${days}d ${hours % 24}h`
  } else if (hours > 0) {
    humanReadable = `${hours}h ${minutes % 60}m`
  } else if (minutes > 0) {
    humanReadable = `${minutes}m ${seconds % 60}s`
  } else {
    humanReadable = `${seconds}s`
  }
  
  return {
    ms: diffMs,
    seconds,
    minutes,
    hours,
    days,
    humanReadable,
  }
}

/**
 * Check if date is today
 */
export function isToday(date: Date | string): boolean {
  const d = typeof date === 'string' ? parseTimestamp(date) : date
  const now = new Date()
  return (
    d.getDate() === now.getDate() &&
    d.getMonth() === now.getMonth() &&
    d.getFullYear() === now.getFullYear()
  )
}

/**
 * Format for cross-team collaboration (shows both UTC and local)
 */
export function formatForCollaboration(
  date: Date | string,
  localTimezone?: string
): string {
  const formatted = getFormattedTime(date)
  const localTz = localTimezone || Intl.DateTimeFormat().resolvedOptions().timeZone
  const localOffset = getTimezoneOffset(localTz)
  
  return `${formatted.utc} (${localOffset})`
}

// ============== React Hook ==============

import { useCallback, useMemo } from 'react'

export interface UseTimezoneUtilsReturn {
  formatUTC: (date: Date | string, options?: Partial<TimezoneOptions>) => string
  formatLocal: (date: Date | string, timezone?: string) => string
  formatRelative: (date: Date | string) => string
  formatTimezone: (date: Date | string, timezone: string) => string
  getFormattedTime: (date: Date | string, options?: Partial<TimezoneOptions>) => FormattedTime
  convertToTimezone: (date: Date | string, targetTimezone: string) => Date
  getTimeDifference: (from: Date | string, to?: Date | string) => ReturnType<typeof getTimeDifference>
  isToday: (date: Date | string) => boolean
  formatForCollaboration: (date: Date | string, localTimezone?: string) => string
  getTimezoneOffset: (timezone: string) => string
  timezones: typeof COMMON_TIMEZONES
}

/**
 * React hook for timezone utilities
 * Provides memoized timezone formatting functions
 */
export function useTimezoneUtils(): UseTimezoneUtilsReturn {
  const formatUTCMemo = useCallback(
    (date: Date | string, options?: Partial<TimezoneOptions>) => formatUTC(date, options),
    []
  )
  
  const formatLocalMemo = useCallback(
    (date: Date | string, timezone?: string) => formatLocal(date, timezone),
    []
  )
  
  const formatRelativeMemo = useCallback(
    (date: Date | string) => formatRelative(date),
    []
  )
  
  const formatTimezoneMemo = useCallback(
    (date: Date | string, timezone: string) => formatTimezone(date, timezone),
    []
  )
  
  const getFormattedTimeMemo = useCallback(
    (date: Date | string, options?: Partial<TimezoneOptions>) => getFormattedTime(date, options),
    []
  )
  
  const convertToTimezoneMemo = useCallback(
    (date: Date | string, targetTimezone: string) => convertToTimezone(date, targetTimezone),
    []
  )
  
  const getTimeDifferenceMemo = useCallback(
    (from: Date | string, to?: Date | string) => getTimeDifference(from, to),
    []
  )
  
  const isTodayMemo = useCallback(
    (date: Date | string) => isToday(date),
    []
  )
  
  const formatForCollaborationMemo = useCallback(
    (date: Date | string, localTimezone?: string) => formatForCollaboration(date, localTimezone),
    []
  )
  
  const getTimezoneOffsetMemo = useCallback(
    (timezone: string) => getTimezoneOffset(timezone),
    []
  )
  
  const timezones = useMemo(() => COMMON_TIMEZONES, [])
  
  return {
    formatUTC: formatUTCMemo,
    formatLocal: formatLocalMemo,
    formatRelative: formatRelativeMemo,
    formatTimezone: formatTimezoneMemo,
    getFormattedTime: getFormattedTimeMemo,
    convertToTimezone: convertToTimezoneMemo,
    getTimeDifference: getTimeDifferenceMemo,
    isToday: isTodayMemo,
    formatForCollaboration: formatForCollaborationMemo,
    getTimezoneOffset: getTimezoneOffsetMemo,
    timezones,
  }
}

// Export all utilities
export const TimezoneUtils = {
  parseTimestamp,
  formatUTC,
  formatLocal,
  formatTimezone,
  formatRelative,
  getFormattedTime,
  convertToTimezone,
  getTimeDifference,
  isToday,
  formatForCollaboration,
  getTimezoneOffset,
  useTimezoneUtils,
  COMMON_TIMEZONES,
}
