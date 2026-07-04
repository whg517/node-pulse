import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { requestPasswordReset } from '@/api/auth'

export default function ForgotPasswordPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [done, setDone] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      // Anti-enumeration: response is identical whether or not the email exists.
      await requestPasswordReset(email.trim())
      setDone(true)
    } catch {
      // Even on error, show the generic message to avoid leaking account state.
      setDone(true)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-center">{t('auth.forgotPassword', 'Reset your password')}</CardTitle>
        </CardHeader>
        <CardContent>
          {done ? (
            <div className="space-y-4 text-center">
              <p className="text-sm text-muted-foreground">
                {t('auth.resetLinkSent', 'If the email exists, a password reset link has been sent.')}
              </p>
              <Button variant="outline" className="w-full" onClick={() => navigate('/login')}>
                {t('auth.backToLogin', 'Back to login')}
              </Button>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">{t('auth.email', 'Email')}</Label>
                <Input
                  id="email"
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@example.com"
                />
              </div>
              <Button type="submit" className="w-full" disabled={submitting || !email.trim()}>
                {submitting ? t('common.saving') : t('auth.sendResetLink', 'Send reset link')}
              </Button>
              <div className="text-center text-sm">
                <Link to="/login" className="text-primary hover:underline">{t('auth.backToLogin', 'Back to login')}</Link>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
