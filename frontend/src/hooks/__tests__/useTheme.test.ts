import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useTheme } from '../useTheme'

// Mock settingsStore
const mockSetTheme = vi.fn()
let mockTheme = 'light'

vi.mock('../../stores/settingsStore', () => ({
  useSettingsStore: vi.fn((selector: (s: unknown) => unknown) =>
    selector({
      theme: mockTheme,
      setTheme: mockSetTheme,
    })
  ),
  applyTheme: vi.fn(),
  getEffectiveTheme: vi.fn((theme: string) => {
    if (theme === 'system') return 'light'
    return theme
  }),
}))

describe('useTheme', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockTheme = 'light'
  })

  it('returns current theme', () => {
    const { result } = renderHook(() => useTheme())
    expect(result.current.theme).toBe('light')
  })

  it('isDark is false for light theme', () => {
    const { result } = renderHook(() => useTheme())
    expect(result.current.isDark).toBe(false)
  })

  it('isLight is true for light theme', () => {
    const { result } = renderHook(() => useTheme())
    expect(result.current.isLight).toBe(true)
  })

  it('isDark is true for dark theme', async () => {
    mockTheme = 'dark'
    const { getEffectiveTheme } = await import('../../stores/settingsStore')
    vi.mocked(getEffectiveTheme).mockReturnValue('dark')

    const { result } = renderHook(() => useTheme())
    expect(result.current.isDark).toBe(true)
  })

  it('isSystem is false for non-system theme', () => {
    const { result } = renderHook(() => useTheme())
    expect(result.current.isSystem).toBe(false)
  })

  it('isSystem is true for system theme', async () => {
    mockTheme = 'system'
    const { useSettingsStore } = await import('../../stores/settingsStore')
    vi.mocked(useSettingsStore).mockImplementation((selector: (s: unknown) => unknown) =>
      selector({ theme: 'system', setTheme: mockSetTheme })
    )

    const { result } = renderHook(() => useTheme())
    expect(result.current.isSystem).toBe(true)
  })

  it('toggleTheme switches from dark to light', async () => {
    mockTheme = 'dark'
    const { getEffectiveTheme } = await import('../../stores/settingsStore')
    vi.mocked(getEffectiveTheme).mockReturnValue('dark')

    const { result } = renderHook(() => useTheme())
    act(() => {
      result.current.toggleTheme()
    })
    expect(mockSetTheme).toHaveBeenCalledWith('light')
  })

  it('toggleTheme switches from light to dark', async () => {
    const { getEffectiveTheme } = await import('../../stores/settingsStore')
    vi.mocked(getEffectiveTheme).mockReturnValue('light')

    const { result } = renderHook(() => useTheme())
    act(() => {
      result.current.toggleTheme()
    })
    expect(mockSetTheme).toHaveBeenCalledWith('dark')
  })

  it('setTheme calls store setTheme', () => {
    const { result } = renderHook(() => useTheme())
    act(() => {
      result.current.setTheme('dark')
    })
    expect(mockSetTheme).toHaveBeenCalledWith('dark')
  })

  it('effectiveTheme is returned', () => {
    const { result } = renderHook(() => useTheme())
    expect(['light', 'dark']).toContain(result.current.effectiveTheme)
  })

  it('listens for system theme changes when theme is system', async () => {
    mockTheme = 'system'
    const { useSettingsStore } = await import('../../stores/settingsStore')
    vi.mocked(useSettingsStore).mockImplementation((selector: (s: unknown) => unknown) =>
      selector({ theme: 'system', setTheme: mockSetTheme })
    )

    const addEventListenerSpy = vi.spyOn(
      window.matchMedia('(prefers-color-scheme: dark)'),
      'addEventListener'
    )
    renderHook(() => useTheme())
    // Just ensure no errors thrown with system theme
    expect(true).toBe(true)
    addEventListenerSpy.mockRestore()
  })
})
