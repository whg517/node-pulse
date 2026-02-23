/**
 * Preferences Page
 *
 * User preferences for timezone, language, and theme settings.
 * Route: /settings/preferences
 */

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useTheme } from '../hooks/useTheme'

import { TimezoneSelector } from '../components/common/TimezoneSelector'
import { LanguageSwitcher } from '../components/common/LanguageSwitcher'
import { ThemeToggle } from '../components/common/ThemeToggle'

export default function PreferencesPage() {
  const { t } = useTranslation()
  const { isDark } = useTheme()
  const [saved, setSaved] = useState(false)

  const handleSave = () => {
    // Components manage their own state via stores
    // This button could trigger additional actions if needed
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          {t('settings.preferences')}
        </h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {t('settings.preferencesDescription')}
        </p>
      </div>

      <div className={`rounded-lg border shadow-sm ${isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}>
        {/* Timezone Setting */}
        <div className={`p-6 ${isDark ? 'border-gray-700' : 'border-gray-200'} border-b`}>
          <h3 className={`text-sm font-medium mb-2 ${isDark ? 'text-gray-200' : 'text-gray-700'}`}>
            {t('settings.timezone')}
          </h3>
          <p className={`text-sm mb-3 ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
            {t('settings.timezoneDescription')}
          </p>
          <TimezoneSelector showDisplayMode size="md" />
        </div>

        {/* Language Setting */}
        <div className={`p-6 ${isDark ? 'border-gray-700' : 'border-gray-200'} border-b`}>
          <h3 className={`text-sm font-medium mb-2 ${isDark ? 'text-gray-200' : 'text-gray-700'}`}>
            {t('settings.language')}
          </h3>
          <p className={`text-sm mb-3 ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
            {t('settings.languageDescription')}
          </p>
          <LanguageSwitcher variant="buttons" size="md" />
        </div>

        {/* Theme Setting */}
        <div className={`p-6 ${isDark ? 'border-gray-700' : 'border-gray-200'} border-b`}>
          <h3 className={`text-sm font-medium mb-2 ${isDark ? 'text-gray-200' : 'text-gray-700'}`}>
            {t('settings.theme')}
          </h3>
          <p className={`text-sm mb-3 ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
            {t('settings.themeDescription')}
          </p>
          <div className="flex items-center gap-4">
            <ThemeToggle />
          </div>
        </div>

        {/* Save Button */}
        <div className="p-6 flex items-center justify-end">
          {saved && (
            <span className="mr-4 text-sm text-green-600 dark:text-green-400">
              {t('settings.saved')}
            </span>
          )}
          <button
            type="button"
            onClick={handleSave}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors"
          >
            {t('settings.savePreferences')}
          </button>
        </div>
      </div>
    </div>
  )
}
