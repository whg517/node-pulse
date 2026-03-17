import { describe, it, expect, beforeEach, vi } from 'vitest'
import { act, renderHook } from '@testing-library/react'

// Mock i18n
vi.mock('../../i18n', () => ({
  default: { changeLanguage: vi.fn() },
}))

describe('settingsStore', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    // Re-import store fresh for each test block
  })

  it('has default language based on browser locale', async () => {
    const { useSettingsStore } = await import('../settingsStore')
    const { result } = renderHook(() => useSettingsStore())
    expect(typeof result.current.language).toBe('string')
  })

  it('setLanguage updates language and calls i18n', async () => {
    const { useSettingsStore } = await import('../settingsStore')
    const i18n = await import('../../i18n')
    const { result } = renderHook(() => useSettingsStore())

    act(() => {
      result.current.setLanguage('zh-CN')
    })

    expect(result.current.language).toBe('zh-CN')
    expect(i18n.default.changeLanguage).toHaveBeenCalledWith('zh-CN')
  })

  it('setTheme updates theme', async () => {
    const { useSettingsStore } = await import('../settingsStore')
    const { result } = renderHook(() => useSettingsStore())

    act(() => {
      result.current.setTheme('dark')
    })

    expect(result.current.theme).toBe('dark')
  })

  it('setTheme applies theme to document', async () => {
    const { useSettingsStore } = await import('../settingsStore')
    const { result } = renderHook(() => useSettingsStore())

    act(() => {
      result.current.setTheme('dark')
    })

    // applyTheme should toggle dark class
    expect(document.documentElement.classList.contains('dark')).toBe(true)

    act(() => {
      result.current.setTheme('light')
    })
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('setTimezone updates timezone', async () => {
    const { useSettingsStore } = await import('../settingsStore')
    const { result } = renderHook(() => useSettingsStore())

    act(() => {
      result.current.setTimezone('Asia/Tokyo')
    })

    expect(result.current.timezone).toBe('Asia/Tokyo')
  })

  it('setTimezoneDisplayMode updates display mode', async () => {
    const { useSettingsStore } = await import('../settingsStore')
    const { result } = renderHook(() => useSettingsStore())

    act(() => {
      result.current.setTimezoneDisplayMode('utc')
    })

    expect(result.current.timezoneDisplayMode).toBe('utc')
  })

  it('resetSettings restores defaults', async () => {
    const { useSettingsStore } = await import('../settingsStore')
    const { result } = renderHook(() => useSettingsStore())

    act(() => {
      result.current.setLanguage('zh-CN')
      result.current.setTheme('dark')
      result.current.setTimezone('Asia/Tokyo')
    })
    act(() => {
      result.current.resetSettings()
    })

    expect(result.current.theme).toBe('system')
  })
})

describe('settingsStore standalone functions', () => {
  it('applyTheme sets dark class for dark theme', async () => {
    const { applyTheme } = await import('../settingsStore')
    applyTheme('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('applyTheme removes dark class for light theme', async () => {
    const { applyTheme } = await import('../settingsStore')
    applyTheme('light')
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('applyTheme uses matchMedia for system theme', async () => {
    const { applyTheme } = await import('../settingsStore')
    applyTheme('system')
    // matchMedia is mocked to return matches: false
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('getEffectiveTheme returns light for light theme', async () => {
    const { getEffectiveTheme } = await import('../settingsStore')
    expect(getEffectiveTheme('light')).toBe('light')
  })

  it('getEffectiveTheme returns dark for dark theme', async () => {
    const { getEffectiveTheme } = await import('../settingsStore')
    expect(getEffectiveTheme('dark')).toBe('dark')
  })

  it('getEffectiveTheme resolves system to light when matchMedia returns false', async () => {
    const { getEffectiveTheme } = await import('../settingsStore')
    expect(getEffectiveTheme('system')).toBe('light')
  })

  it('initializeTheme applies current theme', async () => {
    const { initializeTheme } = await import('../settingsStore')
    // Should not throw
    expect(() => initializeTheme()).not.toThrow()
  })

  it('initializeLanguage sets language in localStorage', async () => {
    const { initializeLanguage } = await import('../settingsStore')
    initializeLanguage()
    expect(localStorage.setItem).toHaveBeenCalledWith('settings:language', expect.any(String))
  })

  it('COMMON_TIMEZONES has at least 20 entries', async () => {
    const { COMMON_TIMEZONES } = await import('../settingsStore')
    expect(COMMON_TIMEZONES.length).toBeGreaterThanOrEqual(20)
  })
})
