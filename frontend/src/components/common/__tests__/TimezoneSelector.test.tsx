import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { TimezoneSelector } from '../TimezoneSelector'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

// Mock settingsStore
const mockSetTimezone = vi.fn()
const mockSetDisplayMode = vi.fn()

vi.mock('../../../stores/settingsStore', () => ({
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

describe('TimezoneSelector', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders timezone dropdown', () => {
    render(<TimezoneSelector />)
    const select = screen.getByLabelText('settings.timezone')
    expect(select).toBeInTheDocument()
  })

  it('renders timezone options', () => {
    render(<TimezoneSelector />)
    expect(screen.getByText('(+00:00) UTC')).toBeInTheDocument()
    expect(screen.getByText('(+08:00) Beijing, Shanghai')).toBeInTheDocument()
  })

  it('calls setTimezone when changed', () => {
    render(<TimezoneSelector />)
    const select = screen.getByLabelText('settings.timezone')
    fireEvent.change(select, { target: { value: 'Asia/Shanghai' } })
    expect(mockSetTimezone).toHaveBeenCalledWith('Asia/Shanghai')
  })

  it('does not show display mode selector by default', () => {
    render(<TimezoneSelector />)
    expect(screen.queryByLabelText('settings.displayMode')).not.toBeInTheDocument()
  })

  it('shows display mode selector when showDisplayMode is true', () => {
    render(<TimezoneSelector showDisplayMode={true} />)
    const displayModeSelect = screen.getByLabelText('settings.displayMode')
    expect(displayModeSelect).toBeInTheDocument()
  })

  it('calls setDisplayMode when display mode changes', () => {
    render(<TimezoneSelector showDisplayMode={true} />)
    const displayModeSelect = screen.getByLabelText('settings.displayMode')
    fireEvent.change(displayModeSelect, { target: { value: 'utc' } })
    expect(mockSetDisplayMode).toHaveBeenCalledWith('utc')
  })

  it('renders with sm size', () => {
    render(<TimezoneSelector size="sm" />)
    const select = screen.getByLabelText('settings.timezone')
    expect(select.className).toContain('text-xs')
  })

  it('renders with lg size', () => {
    render(<TimezoneSelector size="lg" />)
    const select = screen.getByLabelText('settings.timezone')
    expect(select.className).toContain('text-base')
  })

  it('applies custom className', () => {
    const { container } = render(<TimezoneSelector className="custom-tz" />)
    expect(container.firstChild).toBeDefined()
  })
})
