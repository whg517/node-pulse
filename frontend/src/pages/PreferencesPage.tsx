import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { useSettingsStore, COMMON_TIMEZONES, type AlertLevel } from '@/stores/settingsStore'
import { useTheme } from '@/hooks/useTheme'
import { supportedLanguages, type LanguageCode } from '@/i18n-config'
import { PageHeader } from '@/components/layout/PageHeader'
import { changePassword, mfaStatus, mfaSetup, mfaVerify, mfaDisable, getNotificationPrefs, updateNotificationPrefs, type NotificationPrefsDTO } from '@/api/auth'
import * as NotificationService from '@/services/NotificationService'

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

  // 2FA state. Enrollment is a 3-step modal flow: setup → scan QR → verify.
  const [mfaEnabled, setMfaEnabled] = useState(false)
  const [mfaSetupData, setMfaSetupData] = useState<{ secret: string; otpauth_uri: string; ticket: string } | null>(null)
  const [mfaCode, setMfaCode] = useState('')
  const [mfaBusy, setMfaBusy] = useState(false)
  const [mfaMessage, setMfaMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [disablePassword, setDisablePassword] = useState('')

  useEffect(() => {
    mfaStatus().then((r) => setMfaEnabled(r.data.enabled)).catch(() => { /* best-effort */ })
  }, [])

  // Server-side notification preferences (F4 Phase 2) — email-notification floor.
  const [serverPrefs, setServerPrefs] = useState<NotificationPrefsDTO | null>(null)
  const [serverPrefsBusy, setServerPrefsBusy] = useState(false)
  const [serverPrefsMsg, setServerPrefsMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    getNotificationPrefs().then((r) => setServerPrefs(r.data)).catch(() => { /* best-effort */ })
  }, [])

  const saveServerPrefs = async (patch: { email_enabled?: boolean; min_alert_level?: 'P0' | 'P1' | 'P2' }) => {
    if (!serverPrefs) return
    setServerPrefsBusy(true)
    setServerPrefsMsg(null)
    try {
      const r = await updateNotificationPrefs(patch)
      setServerPrefs(r.data)
      setServerPrefsMsg({ type: 'success', text: t('settings.serverPrefsSaved', 'Email notification preferences saved.') })
    } catch (err) {
      setServerPrefsMsg({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' })
    } finally {
      setServerPrefsBusy(false)
    }
  }

  // Notification preferences (F4). Persisted via the settings store; the
  // filter itself is applied in NotificationService via useGlobalRealtime.
  const notificationPrefs = useSettingsStore((s) => s.notificationPrefs)
  const setNotificationPrefs = useSettingsStore((s) => s.setNotificationPrefs)
  const [notifPerm, setNotifPerm] = useState(NotificationService.getPermissionState())

  const levelOptions: { value: AlertLevel; label: string }[] = [
    { value: 'P0', label: t('settings.notifLevelP0', 'P0 and above (critical only)') },
    { value: 'P1', label: t('settings.notifLevelP1', 'P1 and above (warnings + critical)') },
    { value: 'P2', label: t('settings.notifLevelP2', 'All levels') },
  ]

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

  const startMfaSetup = async () => {
    setMfaMessage(null)
    setMfaBusy(true)
    try {
      const r = await mfaSetup()
      setMfaSetupData(r.data)
      setMfaCode('')
    } catch (err) {
      setMfaMessage({ type: 'error', text: err instanceof Error ? err.message : 'Setup failed' })
    } finally {
      setMfaBusy(false)
    }
  }

  const confirmMfaSetup = async () => {
    if (!mfaSetupData || mfaCode.length !== 6) return
    setMfaMessage(null)
    setMfaBusy(true)
    try {
      await mfaVerify(mfaSetupData.ticket, mfaCode)
      setMfaEnabled(true)
      setMfaSetupData(null)
      setMfaCode('')
      setMfaMessage({ type: 'success', text: 'Two-factor authentication enabled.' })
    } catch (err) {
      setMfaMessage({ type: 'error', text: err instanceof Error ? err.message : 'Invalid code' })
    } finally {
      setMfaBusy(false)
    }
  }

  const disableMfa = async () => {
    setMfaMessage(null)
    setMfaBusy(true)
    try {
      await mfaDisable(disablePassword)
      setMfaEnabled(false)
      setDisablePassword('')
      setMfaMessage({ type: 'success', text: 'Two-factor authentication disabled.' })
    } catch (err) {
      setMfaMessage({ type: 'error', text: err instanceof Error ? err.message : 'Disable failed' })
    } finally {
      setMfaBusy(false)
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

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('settings.twoFactor', 'Two-Factor Authentication (2FA)')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {mfaMessage && (
            <div className={`rounded-md px-3 py-2 text-sm ${mfaMessage.type === 'success' ? 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300' : 'bg-destructive/10 text-destructive'}`}>
              {mfaMessage.text}
            </div>
          )}

          {!mfaEnabled && !mfaSetupData && (
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">{t('settings.twoFactorHint', 'Add a second factor (TOTP authenticator app) to your account.')}</p>
              <Button onClick={startMfaSetup} disabled={mfaBusy}>{t('settings.enable2fa', 'Enable 2FA')}</Button>
            </div>
          )}

          {mfaSetupData && (
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">{t('settings.twoFactorScanHint', 'Scan this secret in your authenticator app (or enter it manually), then enter the 6-digit code it shows.')}</p>
              <div className="break-all rounded-md border bg-muted/50 p-3 font-mono text-xs">{mfaSetupData.secret}</div>
              <details className="text-xs text-muted-foreground">
                <summary className="cursor-pointer">otpauth URI</summary>
                <code className="break-all">{mfaSetupData.otpauth_uri}</code>
              </details>
              <Input
                type="text"
                inputMode="numeric"
                maxLength={6}
                placeholder="123456"
                value={mfaCode}
                onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, ''))}
                disabled={mfaBusy}
                className="text-center text-lg tracking-widest"
              />
              <div className="flex gap-2">
                <Button onClick={confirmMfaSetup} disabled={mfaBusy || mfaCode.length !== 6}>{t('settings.verify', 'Verify & enable')}</Button>
                <Button variant="outline" onClick={() => { setMfaSetupData(null); setMfaCode('') }} disabled={mfaBusy}>{t('common.cancel')}</Button>
              </div>
            </div>
          )}

          {mfaEnabled && (
            <div className="space-y-3">
              <p className="text-sm text-green-700 dark:text-green-300">{t('settings.twoFactorOn', '2FA is enabled. You will be asked for a code at every sign-in.')}</p>
              <Input
                type="password"
                placeholder={t('settings.currentPassword', 'Current password')}
                value={disablePassword}
                onChange={(e) => setDisablePassword(e.target.value)}
                disabled={mfaBusy}
              />
              <Button variant="destructive" onClick={disableMfa} disabled={mfaBusy || !disablePassword}>{t('settings.disable2fa', 'Disable 2FA')}</Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('settings.notifications', 'Notifications')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <Label className="text-sm">{t('settings.browserNotifications', 'Browser notifications')}</Label>
              <p className="text-xs text-muted-foreground">{t('settings.browserNotificationsHint', 'Show desktop alerts for matching events in real time.')}</p>
            </div>
            <Switch
              checked={notificationPrefs.enabled}
              onCheckedChange={(v) => {
                setNotificationPrefs({ enabled: v })
                if (v && !NotificationService.getPermissionState().granted) {
                  void NotificationService.requestPermission().then(() => setNotifPerm(NotificationService.getPermissionState()))
                }
              }}
            />
          </div>

          {notifPerm.denied && (
            <div className="rounded-md bg-destructive/10 px-3 py-2 text-xs text-destructive">
              {t('settings.notifPermissionDenied', 'Browser notification permission is denied. Update your browser site settings to re-enable.')}
            </div>
          )}

          <div className="space-y-2">
            <Label className="text-sm">{t('settings.minimumAlertLevel', 'Minimum alert level')}</Label>
            <p className="text-xs text-muted-foreground">{t('settings.minimumAlertLevelHint', 'Only notify for alerts at or above this severity.')}</p>
            <select
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
              value={notificationPrefs.minLevel}
              onChange={(e) => setNotificationPrefs({ minLevel: e.target.value as AlertLevel })}
              disabled={!notificationPrefs.enabled}
            >
              {levelOptions.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>

          <div className="flex items-center justify-between">
            <div>
              <Label className="text-sm">{t('settings.nodeOnlineOfflineNotif', 'Node online/offline events')}</Label>
              <p className="text-xs text-muted-foreground">{t('settings.nodeOnlineOfflineNotifHint', 'Also notify when nodes come and go (can be noisy).')}</p>
            </div>
            <Switch
              checked={notificationPrefs.nodeOnlineOffline}
              onCheckedChange={(v) => setNotificationPrefs({ nodeOnlineOffline: v })}
              disabled={!notificationPrefs.enabled}
            />
          </div>
        </CardContent>
      </Card>

      {/* Server-side email notification preferences (F4 Phase 2). Controls
          which alert severities trigger an email to the user. Composes with
          the client-side browser-notification floor above. */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('settings.emailNotifications', 'Email notifications')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {serverPrefsMsg && (
            <div className={`rounded-md px-3 py-2 text-sm ${serverPrefsMsg.type === 'success' ? 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300' : 'bg-destructive/10 text-destructive'}`}>
              {serverPrefsMsg.text}
            </div>
          )}
          <div className="flex items-center justify-between">
            <div>
              <Label className="text-sm">{t('settings.emailNotifEnabled', 'Send email on alerts')}</Label>
              <p className="text-xs text-muted-foreground">{t('settings.emailNotifEnabledHint', 'When enabled, alerts at or above the minimum level are emailed to your account address.')}</p>
            </div>
            <Switch
              checked={serverPrefs?.email_enabled ?? true}
              onCheckedChange={(v) => void saveServerPrefs({ email_enabled: v })}
              disabled={serverPrefsBusy || !serverPrefs}
            />
          </div>
          <div className="space-y-2">
            <Label className="text-sm">{t('settings.emailMinLevel', 'Minimum alert level for email')}</Label>
            <p className="text-xs text-muted-foreground">{t('settings.emailMinLevelHint', 'P0 = critical only; P1 = warnings + critical; P2 = all.')}</p>
            <select
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
              value={serverPrefs?.min_alert_level ?? 'P1'}
              onChange={(e) => void saveServerPrefs({ min_alert_level: e.target.value as 'P0' | 'P1' | 'P2' })}
              disabled={serverPrefsBusy || !serverPrefs || !serverPrefs.email_enabled}
            >
              <option value="P0">{t('settings.notifLevelP0', 'P0 and above (critical only)')}</option>
              <option value="P1">{t('settings.notifLevelP1', 'P1 and above (warnings + critical)')}</option>
              <option value="P2">{t('settings.notifLevelP2', 'All levels')}</option>
            </select>
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
