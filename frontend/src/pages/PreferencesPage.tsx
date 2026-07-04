import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { useSettingsStore, COMMON_TIMEZONES } from '@/stores/settingsStore'
import { useTheme } from '@/hooks/useTheme'
import { supportedLanguages, type LanguageCode } from '@/i18n-config'
import { PageHeader } from '@/components/layout/PageHeader'
import { changePassword } from '@/api/auth'

export default function PreferencesPage() {
  const { t, i18n } = useTranslation()
  const [saved, setSaved] = useState(false)
  const language = useSettingsStore((s) => s.language)
  const setLanguage = useSettingsStore((s) => s.setLanguage)
  const timezone = useSettingsStore((s) => s.timezone)
  const setTimezone = useSettingsStore((s) => s.setTimezone)
  const { setTheme, isDark } = useTheme()

  // Change-password form state
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [isChangingPassword, setIsChangingPassword] = useState(false)
  const [passwordMessage, setPasswordMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const handleLanguageChange = (code: LanguageCode) => {
    setLanguage(code)
    i18n.changeLanguage(code)
  }

  const handleSave = () => {
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const handleChangePassword = async () => {
    setPasswordMessage(null)
    if (newPassword !== confirmPassword) {
      setPasswordMessage({ type: 'error', text: t('settings.passwordMismatch', 'New passwords do not match') })
      return
    }
    if (newPassword.length < 8) {
      setPasswordMessage({ type: 'error', text: t('settings.passwordTooShort', 'Password must be at least 8 characters') })
      return
    }
    setIsChangingPassword(true)
    try {
      const res = await changePassword(currentPassword, newPassword)
      setPasswordMessage({
        type: 'success',
        text: res.sessions_revoked
          ? t('settings.passwordChangedSessionsRevoked', 'Password changed. Other sessions were signed out.')
          : t('settings.passwordChanged', 'Password changed successfully'),
      })
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err) {
      setPasswordMessage({
        type: 'error',
        text: err instanceof Error ? err.message : t('settings.passwordChangeFailed', 'Failed to change password'),
      })
    } finally {
      setIsChangingPassword(false)
    }
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

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('settings.security', 'Security')}</CardTitle>
          <p className="text-xs text-muted-foreground">{t('settings.securityDescription', 'Change your account password. Other sessions will be signed out.')}</p>
        </CardHeader>
        <CardContent className="space-y-3">
          {passwordMessage && (
            <div className={`rounded-md px-3 py-2 text-sm ${passwordMessage.type === 'error' ? 'bg-destructive/10 text-destructive' : 'bg-healthy-bg text-healthy-text'}`}>
              {passwordMessage.text}
            </div>
          )}
          <div className="space-y-1.5">
            <Label htmlFor="current-password">{t('settings.currentPassword', 'Current password')}</Label>
            <Input
              id="current-password"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              autoComplete="current-password"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="new-password">{t('settings.newPassword', 'New password')}</Label>
            <Input
              id="new-password"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="confirm-password">{t('settings.confirmPassword', 'Confirm new password')}</Label>
            <Input
              id="confirm-password"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => void handleChangePassword()}
            disabled={isChangingPassword || !currentPassword || !newPassword || !confirmPassword}
          >
            {isChangingPassword ? t('common.saving', 'Saving...') : t('settings.changePassword', 'Change Password')}
          </Button>
        </CardContent>
      </Card>

      <div className="flex items-center justify-end gap-4">
        {saved && <span className="text-sm text-green-600">{t('settings.saved')}</span>}
        <Button onClick={handleSave}>{t('settings.savePreferences')}</Button>
      </div>
    </div>
  )
}
