/**
 * TimezoneSelector Component
 *
 * Dropdown for selecting timezone and display mode.
 * Persists preference via settings store.
 */

import { useTranslation } from 'react-i18next'
import {
  useSettingsStore,
  COMMON_TIMEZONES,
  type TimezoneDisplayMode,
} from '../../stores/settingsStore'

export interface TimezoneSelectorProps {
  /** Show display mode selector */
  showDisplayMode?: boolean
  /** Additional CSS classes */
  className?: string
  /** Size variant */
  size?: 'sm' | 'md' | 'lg'
}

const sizeClasses = {
  sm: 'text-xs px-2 py-1',
  md: 'text-sm px-3 py-1.5',
  lg: 'text-base px-4 py-2',
}

const displayModes: { value: TimezoneDisplayMode; labelKey: string }[] = [
  { value: 'utc', labelKey: 'settings.utc' },
  { value: 'local', labelKey: 'settings.local' },
  { value: 'nodeLocal', labelKey: 'settings.nodeLocal' },
  { value: 'multi', labelKey: 'settings.multi' },
]

/**
 * TimezoneSelector Component
 */
export function TimezoneSelector({
  showDisplayMode = false,
  className = '',
  size = 'md',
}: TimezoneSelectorProps) {
  const { t } = useTranslation()
  const timezone = useSettingsStore((state) => state.timezone)
  const displayMode = useSettingsStore((state) => state.timezoneDisplayMode)
  const setTimezone = useSettingsStore((state) => state.setTimezone)
  const setDisplayMode = useSettingsStore((state) => state.setTimezoneDisplayMode)

  return (
    <div className={`flex flex-col gap-2 ${className}`}>
      {/* Timezone Dropdown */}
      <div>
        <label
          htmlFor="timezone-select"
          className="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]"
        >
          {t('settings.timezone')}
        </label>
        <select
          id="timezone-select"
          value={timezone}
          onChange={(e) => setTimezone(e.target.value)}
          className={`w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)]
            text-[var(--color-text-primary)] focus:border-[var(--color-brand)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand)]
            ${sizeClasses[size]}`}
        >
          {COMMON_TIMEZONES.map((tz) => (
            <option key={tz.value} value={tz.value}>
              ({tz.offset}) {tz.label}
            </option>
          ))}
        </select>
      </div>

      {/* Display Mode Selector */}
      {showDisplayMode && (
        <div>
          <label
            htmlFor="display-mode-select"
            className="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]"
          >
            {t('settings.displayMode')}
          </label>
          <select
            id="display-mode-select"
            value={displayMode}
            onChange={(e) => setDisplayMode(e.target.value as TimezoneDisplayMode)}
            className={`w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)]
              text-[var(--color-text-primary)] focus:border-[var(--color-brand)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand)]
              ${sizeClasses[size]}`}
          >
            {displayModes.map((mode) => (
              <option key={mode.value} value={mode.value}>
                {t(mode.labelKey)}
              </option>
            ))}
          </select>
        </div>
      )}
    </div>
  )
}

export default TimezoneSelector
