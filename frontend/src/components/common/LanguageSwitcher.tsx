/**
 * LanguageSwitcher Component
 *
 * Dropdown for switching between supported languages.
 * Persists preference via settings store and updates i18n.
 */

import { useTranslation } from 'react-i18next'
import { useSettingsStore } from '../../stores/settingsStore'
import { supportedLanguages, type LanguageCode } from '../../i18n-config'

export interface LanguageSwitcherProps {
  /** Show as dropdown or inline buttons */
  variant?: 'dropdown' | 'buttons'
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

/**
 * LanguageSwitcher Component
 */
export function LanguageSwitcher({
  variant = 'dropdown',
  className = '',
  size = 'md',
}: LanguageSwitcherProps) {
  const { t } = useTranslation()
  const language = useSettingsStore((state) => state.language)
  const setLanguage = useSettingsStore((state) => state.setLanguage)

  const handleLanguageChange = (langCode: LanguageCode) => {
    setLanguage(langCode)
  }

  if (variant === 'buttons') {
    return (
      <div className={`flex items-center gap-1 ${className}`}>
        {supportedLanguages.map((lang) => (
          <button
            key={lang.code}
            type="button"
            onClick={() => handleLanguageChange(lang.code)}
            className={`rounded-md font-medium transition-colors ${sizeClasses[size]}
              ${language === lang.code
                ? 'bg-[var(--color-brand)] text-white'
                : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]'
              }`}
            title={lang.name}
            aria-label={lang.name}
            aria-pressed={language === lang.code}
          >
            {lang.nativeName}
          </button>
        ))}
      </div>
    )
  }

  // Dropdown variant
  return (
    <select
      value={language}
      onChange={(e) => handleLanguageChange(e.target.value as LanguageCode)}
      className={`rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)]
        text-[var(--color-text-primary)] focus:border-[var(--color-brand)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand)]
        ${sizeClasses[size]} ${className}`}
      aria-label={t('settings.language')}
    >
      {supportedLanguages.map((lang) => (
        <option key={lang.code} value={lang.code}>
          {lang.nativeName}
        </option>
      ))}
    </select>
  )
}

export default LanguageSwitcher
