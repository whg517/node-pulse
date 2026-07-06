import { useState, type FormEvent, useEffect } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { useAuthStore } from '@/stores/authStore'
import { login as apiLogin, mfaLogin as apiMfaLogin } from '@/api/auth'
import { ACCESS_TOKEN_EXPIRY_MINUTES } from '@/config/constants'
import type { User } from '@/stores/types'

interface LocationState {
  from?: { pathname?: string }
}

export default function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const authLoading = useAuthStore((s) => s.isLoading)
  const setUser = useAuthStore((s) => s.setUser)
  const setAccessToken = useAuthStore((s) => s.setAccessToken)
  const setCsrfToken = useAuthStore((s) => s.setCsrfToken)

  const from = (location.state as LocationState | null)?.from?.pathname || '/dashboard'

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [apiError, setApiError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  // 2FA second factor: when login returns mfa_required we render the code
  // input instead of the password form and trade the ticket for tokens.
  const [mfaTicket, setMfaTicket] = useState<string | null>(null)
  const [mfaCode, setMfaCode] = useState('')

  useEffect(() => {
    if (isAuthenticated && !authLoading) {
      navigate(from, { replace: true })
    }
  }, [isAuthenticated, authLoading, navigate, from])

  // Apply a completed-login response (either step) to the auth store and navigate.
  const applyLoginResponse = (response: { data: { user_id: string; username: string; role: 'admin' | 'operator' | 'viewer'; access_token: string; csrf_token?: string } }) => {
    const user: User = {
      id: response.data.user_id,
      username: response.data.username,
      role: response.data.role,
    }
    setUser(user)
    setAccessToken(response.data.access_token, ACCESS_TOKEN_EXPIRY_MINUTES * 60 * 1000)
    if (response.data.csrf_token) setCsrfToken(response.data.csrf_token)
    navigate(from, { replace: true })
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setApiError(null)
    if (!username.trim() || !password) return

    setIsLoading(true)
    try {
      const response = await apiLogin({ username: username.trim(), password })
      if (response.data.mfa_required && response.data.mfa_ticket) {
        // First step succeeded; render the TOTP code input.
        setMfaTicket(response.data.mfa_ticket)
        setMfaCode('')
        return
      }
      applyLoginResponse(response)
    } catch (error) {
      const err = error as { code?: string }
      if (err.code === 'ERR_INVALID_CREDENTIALS') {
        setApiError('Invalid username or password')
      } else if (err.code === 'ERR_RATE_LIMITED' || err.code === 'ERR_RATE_LIMIT_EXCEEDED') {
        setApiError('Too many login attempts. Please try again later.')
      } else {
        setApiError('Connection failed. Please check your network connection.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  // 2FA second step: trade the pending ticket + TOTP code for tokens.
  const handleMfaSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setApiError(null)
    if (!mfaTicket || !mfaCode.trim()) return
    setIsLoading(true)
    try {
      const response = await apiMfaLogin(mfaTicket, mfaCode.trim())
      applyLoginResponse(response)
    } catch (error) {
      const err = error as { code?: string }
      if (err.code === 'ERR_MFA_INVALID') {
        setApiError('Invalid authentication code. Try again.')
      } else {
        setApiError('Verification failed. Please sign in again.')
        setMfaTicket(null) // ticket likely expired; force a fresh login
      }
    } finally {
      setIsLoading(false)
    }
  }

  if (authLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-muted">
        <div className="text-center">
          <div className="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="mt-4 text-sm text-muted-foreground">Checking session...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted px-4 py-12">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold">NodePulse</CardTitle>
          {mfaTicket && (
            <p className="mt-1 text-sm text-muted-foreground">Enter the 6-digit code from your authenticator app.</p>
          )}
        </CardHeader>
        <CardContent>
          {mfaTicket ? (
            <form onSubmit={handleMfaSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="mfa-code">Authentication code</Label>
                <Input
                  id="mfa-code"
                  type="text"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  pattern="[0-9]*"
                  maxLength={6}
                  value={mfaCode}
                  onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, ''))}
                  disabled={isLoading}
                  autoFocus
                  className="text-center text-lg tracking-widest"
                />
              </div>
              {apiError && (
                <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{apiError}</div>
              )}
              <Button type="submit" className="w-full" disabled={isLoading || mfaCode.length !== 6}>
                {isLoading ? 'Verifying...' : 'Verify'}
              </Button>
              <button
                type="button"
                onClick={() => { setMfaTicket(null); setMfaCode(''); setApiError(null) }}
                className="w-full text-sm text-muted-foreground hover:text-foreground"
              >
                Back to sign in
              </button>
            </form>
          ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
          <p className="text-sm text-muted-foreground">Sign in to your account</p>
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                type="text"
                autoComplete="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={isLoading}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  disabled={isLoading}
                  className="pr-10"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  tabIndex={-1}
                >
                  {showPassword ? 'Hide' : 'Show'}
                </button>
              </div>
            </div>

            {apiError && (
              <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
                {apiError}
              </div>
            )}

            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? 'Signing in...' : 'Sign in'}
            </Button>
          </form>
          )}
          <div className="mt-3 text-center text-sm">
            <Link to="/forgot-password" className="text-primary hover:underline">Forgot password?</Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
