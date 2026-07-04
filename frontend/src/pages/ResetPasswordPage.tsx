import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { confirmPasswordReset } from '@/api/auth'

export default function ResetPasswordPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const token = params.get('token') || ''

  const [newPassword, setNewPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [done, setDone] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    if (newPassword !== confirm) {
      setError(t('settings.passwordMismatch', 'Passwords do not match'))
      return
    }
    if (newPassword.length < 8) {
      setError(t('settings.passwordTooShort', 'Password must be at least 8 characters'))
      return
    }
    setSubmitting(true)
    try {
      await confirmPasswordReset(token, newPassword)
      setDone(true)
      // Backend revokes all sessions on reset; redirect to login after a beat.
      setTimeout(() => navigate('/login', { replace: true }), 2000)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('auth.resetFailed', 'Reset failed'))
    } finally {
      setSubmitting(false)
    }
  }

  if (!token) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4">
        <Card className="w-full max-w-md">
          <CardContent className="py-8 text-center text-sm text-destructive">
            {t('auth.resetMissingToken', 'No reset token provided. Use the link from your email.')}
            <div className="mt-4">
              <Link to="/login" className="text-primary hover:underline">{t('auth.backToLogin', 'Back to login')}</Link>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-center">{t('auth.resetPassword', 'Choose a new password')}</CardTitle>
        </CardHeader>
        <CardContent>
          {done ? (
            <p className="text-center text-sm text-muted-foreground">
              {t('auth.resetSuccess', 'Password reset successfully. Redirecting to login...')}
            </p>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</div>}
              <div className="space-y-2">
                <Label htmlFor="new-password">{t('settings.newPassword', 'New password')}</Label>
                <Input id="new-password" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} autoComplete="new-password" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirm-password">{t('settings.confirmPassword', 'Confirm new password')}</Label>
                <Input id="confirm-password" type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} autoComplete="new-password" />
              </div>
              <Button type="submit" className="w-full" disabled={submitting}>
                {submitting ? t('common.saving') : t('auth.resetPassword', 'Reset password')}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
