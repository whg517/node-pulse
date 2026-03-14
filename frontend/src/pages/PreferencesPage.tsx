/**
 * Preferences Page
 *
 * User preferences for timezone, language, and theme settings.
 * Route: /settings/preferences
 */

import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { TimezoneSelector } from '../components/common/TimezoneSelector'
import { LanguageSwitcher } from '../components/common/LanguageSwitcher'
import { ThemeToggle } from '../components/common/ThemeToggle'
import { PageContainer, ActionButton } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'

export default function PreferencesPage() {
  const { t } = useTranslation()
  const [saved, setSaved] = useState(false)

  const handleSave = () => {
    // Components manage their own state via stores
    // This button could trigger additional actions if needed
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('settings.preferences')}
        subtitle={t('settings.preferencesDescription')}
        showBreadcrumb
      />

      <div>
        <div className="rounded-lg border shadow-sm bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
          {/* Timezone Setting */}
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <h3 className="text-sm font-medium mb-2 text-gray-700 dark:text-gray-200">
              {t('settings.timezone')}
            </h3>
            <p className="text-sm mb-3 text-gray-500 dark:text-gray-400">
              {t('settings.timezoneDescription')}
            </p>
            <TimezoneSelector showDisplayMode size="md" />
          </div>

          {/* Language Setting */}
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <h3 className="text-sm font-medium mb-2 text-gray-700 dark:text-gray-200">
              {t('settings.language')}
            </h3>
            <p className="text-sm mb-3 text-gray-500 dark:text-gray-400">
              {t('settings.languageDescription')}
            </p>
            <LanguageSwitcher variant="buttons" size="md" />
          </div>

          {/* Theme Setting */}
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <h3 className="text-sm font-medium mb-2 text-gray-700 dark:text-gray-200">
              {t('settings.theme')}
            </h3>
            <p className="text-sm mb-3 text-gray-500 dark:text-gray-400">
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
            <ActionButton onClick={handleSave}>
              {t('settings.savePreferences')}
            </ActionButton>
          </div>
        </div>
      </div>
    </PageContainer>
  )
}
