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
      />

      <div>
        <div className="rounded-lg border shadow-sm bg-[var(--color-bg-surface)] border-[var(--color-border)]">
          {/* Timezone Setting */}
          <div className="p-6 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-medium mb-2 text-[var(--color-text-primary)]">
              {t('settings.timezone')}
            </h3>
            <p className="text-sm mb-3 text-[var(--color-text-muted)]">
              {t('settings.timezoneDescription')}
            </p>
            <TimezoneSelector showDisplayMode size="md" />
          </div>

          {/* Language Setting */}
          <div className="p-6 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-medium mb-2 text-[var(--color-text-primary)]">
              {t('settings.language')}
            </h3>
            <p className="text-sm mb-3 text-[var(--color-text-muted)]">
              {t('settings.languageDescription')}
            </p>
            <LanguageSwitcher variant="buttons" size="md" />
          </div>

          {/* Theme Setting */}
          <div className="p-6 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-medium mb-2 text-[var(--color-text-primary)]">
              {t('settings.theme')}
            </h3>
            <p className="text-sm mb-3 text-[var(--color-text-muted)]">
              {t('settings.themeDescription')}
            </p>
            <div className="flex items-center gap-4">
              <ThemeToggle />
            </div>
          </div>

          {/* Save Button */}
          <div className="p-6 flex items-center justify-end">
            {saved && (
              <span className="mr-4 text-sm text-[var(--color-healthy)]">
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
