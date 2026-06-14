import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { useSettingsStore, COMMON_TIMEZONES } from '@/stores/settingsStore'
import { useTheme } from '@/hooks/useTheme'
import { supportedLanguages, type LanguageCode } from '@/i18n-config'
import { PageHeader } from '@/components/layout/PageHeader'

export default function PreferencesPage() {
  const { t, i18n } = useTranslation()
  const [saved, setSaved] = useState(false)
  const language = useSettingsStore((s) => s.language)
  const setLanguage = useSettingsStore((s) => s.setLanguage)
  const timezone = useSettingsStore((s) => s.timezone)
  const setTimezone = useSettingsStore((s) => s.setTimezone)
  const { setTheme, isDark } = useTheme()

  const handleLanguageChange = (code: LanguageCode) => {
    setLanguage(code)
    i18n.changeLanguage(code)
  }

  const handleSave = () => {
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('settings.preferences')}
        subtitle={t('settings.preferencesDescription')}
      />

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('settings.timezone')}</CardTitle>
          <p className="text-xs text-muted-foreground">{t('settings.timezoneDescription')}</p>
        </CardHeader>
        <CardContent>
          <select
            value={timezone}
            onChange={(e) => setTimezone(e.target.value)}
            className="w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm"
          >
            {COMMON_TIMEZONES.map((tz: { value: string; offset: string; label: string }) => (
              <option key={tz.value} value={tz.value}>
                ({tz.offset}) {tz.label}
              </option>
            ))}
          </select>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('settings.language')}</CardTitle>
          <p className="text-xs text-muted-foreground">{t('settings.languageDescription')}</p>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            {supportedLanguages.map((lang) => (
              <button
                key={lang.code}
                type="button"
                onClick={() => handleLanguageChange(lang.code)}
                className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                  language === lang.code
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted text-muted-foreground hover:bg-accent'
                }`}
              >
                {lang.nativeName}
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('settings.theme')}</CardTitle>
          <p className="text-xs text-muted-foreground">{t('settings.themeDescription')}</p>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3">
            <Switch
              id="dark-mode"
              checked={isDark}
              onCheckedChange={(checked) => setTheme(checked ? 'dark' : 'light')}
            />
            <Label htmlFor="dark-mode">{isDark ? t('settings.darkMode') : t('settings.lightMode')}</Label>
          </div>
        </CardContent>
      </Card>

      <div className="flex items-center justify-end gap-4">
        {saved && <span className="text-sm text-green-600">{t('settings.saved')}</span>}
        <Button onClick={handleSave}>{t('settings.savePreferences')}</Button>
      </div>
    </div>
  )
}
