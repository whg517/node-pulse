/**
 * ThemeToggle Component
 *
 * Toggle button for switching between light and dark themes.
 * Persists preference via settings store.
 */

import { useTranslation } from 'react-i18next'
import { useTheme } from '../../hooks/useTheme'
import type { ThemeMode } from '../../stores/settingsStore'

export interface ThemeToggleProps {
  /** Show dropdown with all options (light/dark/system) instead of simple toggle */
  showDropdown?: boolean
  /** Additional CSS classes */
  className?: string
  /** Size variant */
  size?: 'sm' | 'md' | 'lg'
}

/**
 * Simple icon components for themes
 */
function SunIcon({ className = '' }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={1.5}
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z"
      />
    </svg>
  )
}

function MoonIcon({ className = '' }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={1.5}
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z"
      />
    </svg>
  )
}

function ComputerIcon({ className = '' }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={1.5}
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25"
      />
    </svg>
  )
}

const sizeClasses = {
  sm: 'h-4 w-4',
  md: 'h-5 w-5',
  lg: 'h-6 w-6',
}

const buttonSizeClasses = {
  sm: 'p-1',
  md: 'p-2',
  lg: 'p-2.5',
}

/**
 * ThemeToggle Component
 */
export function ThemeToggle({
  showDropdown = false,
  className = '',
  size = 'md',
}: ThemeToggleProps) {
  const { t } = useTranslation()
  const { theme, setTheme, toggleTheme, isDark } = useTheme()

  if (!showDropdown) {
    return (
      <button
        type="button"
        onClick={toggleTheme}
        className={`rounded-lg bg-gray-100 text-gray-600 hover:bg-gray-200
          dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700
          transition-colors duration-200 ${buttonSizeClasses[size]} ${className}`}
        title={isDark ? t('settings.lightMode') : t('settings.darkMode')}
        aria-label={isDark ? t('settings.lightMode') : t('settings.darkMode')}
      >
        {isDark ? (
          <SunIcon className={sizeClasses[size]} />
        ) : (
          <MoonIcon className={sizeClasses[size]} />
        )}
      </button>
    )
  }

  const themeOptions: { value: ThemeMode; label: string; icon: typeof SunIcon }[] = [
    { value: 'light', label: t('settings.lightMode'), icon: SunIcon },
    { value: 'dark', label: t('settings.darkMode'), icon: MoonIcon },
    { value: 'system', label: t('settings.systemDefault'), icon: ComputerIcon },
  ]

  return (
    <div className={`relative ${className}`}>
      <div className="flex items-center gap-1 rounded-lg bg-gray-100 p-1 dark:bg-gray-800">
        {themeOptions.map((option) => (
          <button
            key={option.value}
            type="button"
            onClick={() => setTheme(option.value)}
            className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors
              ${theme === option.value
                ? 'bg-white text-gray-900 shadow dark:bg-gray-700 dark:text-white'
                : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
              }`}
            title={option.label}
            aria-label={option.label}
            aria-pressed={theme === option.value}
          >
            <option.icon className={sizeClasses[size]} />
          </button>
        ))}
      </div>
    </div>
  )
}

export default ThemeToggle
