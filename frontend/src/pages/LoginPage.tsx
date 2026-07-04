import { useState, type FormEvent, useEffect } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { useAuthStore } from '@/stores/authStore'
import { login as apiLogin } from '@/api/auth'
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

  useEffect(() => {
    if (isAuthenticated && !authLoading) {
      navigate(from, { replace: true })
    }
  }, [isAuthenticated, authLoading, navigate, from])

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setApiError(null)
    if (!username.trim() || !password) return

    setIsLoading(true)
    try {
      const response = await apiLogin({ username: username.trim(), password })
      const user: User = {
        id: response.data.user_id,
        username: response.data.username,
        role: response.data.role,
      }
      setUser(user)
      setAccessToken(response.data.access_token, ACCESS_TOKEN_EXPIRY_MINUTES * 60 * 1000)
      if (response.data.csrf_token) setCsrfToken(response.data.csrf_token)
      navigate(from, { replace: true })
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
          <p className="text-sm text-muted-foreground">Sign in to your account</p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
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
          <div className="mt-3 text-center text-sm">
            <Link to="/forgot-password" className="text-primary hover:underline">Forgot password?</Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
