import { useState, type FormEvent } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { login, logout } from '../api/auth'
import { useAuthStore } from '../stores/authStore'
import type { ValidationError } from '../types/auth'
import { ToastNotification, type ToastProps } from '../components/ToastNotification'

export default function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const storeLogin = useAuthStore((state) => state.login)
  const storeLogout = useAuthStore((state) => state.logout)
  const { isAuthenticated } = useAuthStore()

  // Get the location from router state (where user was trying to go)
  const from = (location.state as any)?.from?.pathname || '/dashboard'

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [errors, setErrors] = useState<ValidationError[]>([])
  const [apiError, setApiError] = useState<string | null>(null)
  const [accountLockedMessage, setAccountLockedMessage] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [toasts, setToasts] = useState<ToastProps[]>([])

  const showToast = (type: ToastProps['type'], title: string, message?: string) => {
    const id = Date.now().toString()
    setToasts((prev) => [...prev, { id, type, title, message, onClose: handleToastClose }])
  }

  const handleToastClose = (id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id))
  }

  const validateForm = (): boolean => {
    const newErrors: ValidationError[] = []

    if (!username.trim()) {
      newErrors.push({ field: 'username', message: 'Username is required' })
    } else if (username.length < 3) {
      newErrors.push({ field: 'username', message: 'Username must be at least 3 characters' })
    } else if (username.length > 50) {
      newErrors.push({ field: 'username', message: 'Username must be less than 50 characters' })
    }

    if (!password) {
      newErrors.push({ field: 'password', message: 'Password is required' })
    } else if (password.length < 8) {
      newErrors.push({ field: 'password', message: 'Password must be at least 8 characters' })
    } else if (password.length > 32) {
      newErrors.push({ field: 'password', message: 'Password must be less than 32 characters' })
    } else {
      // Password complexity validation
      const hasUpperCase = /[A-Z]/.test(password)
      const hasLowerCase = /[a-z]/.test(password)
      const hasDigit = /[0-9]/.test(password)
      if (!hasUpperCase || !hasLowerCase || !hasDigit) {
        newErrors.push({
          field: 'password',
          message: 'Password must contain at least one uppercase letter, one lowercase letter, and one digit'
        })
      }
    }

    setErrors(newErrors)
    return newErrors.length === 0
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setApiError(null)
    setAccountLockedMessage(null)

    if (!validateForm()) {
      return
    }

    setIsLoading(true)

    try {
      await login({ username: username.trim(), password })
      await storeLogin(username.trim(), password)

      showToast('success', 'Login successful', 'Redirecting to dashboard...')

      setTimeout(() => {
        navigate(from, { replace: true })
      }, 500)
    } catch (error) {
      const err = error as { code?: string; details?: { locked_until?: string } }

      if (err.code === 'ERR_ACCOUNT_LOCKED') {
        const lockedUntil = err.details?.locked_until
          ? new Date(err.details.locked_until).toLocaleString()
          : 'unknown time'
        setAccountLockedMessage(`Account locked. Please try again after ${lockedUntil}`)
        showToast('warning', 'Account locked', `Locked until ${lockedUntil}`)
      } else if (err.code === 'ERR_INVALID_CREDENTIALS') {
        setApiError('Invalid username or password')
        setPassword('')
      } else if (err.code === 'ERR_RATE_LIMIT_EXCEEDED') {
        setApiError('Too many login attempts. Please try again later.')
        showToast('error', 'Too many attempts', 'Please try again later')
      } else {
        setApiError('Connection failed. Please check your network connection.')
        showToast('error', 'Connection failed', 'Please check your network connection')
      }
    } finally {
      setIsLoading(false)
    }
  }

  const getFieldError = (field: 'username' | 'password') => {
    return errors.find((e) => e.field === field)?.message
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">Node Pulse</h2>
          <p className="mt-2 text-center text-sm text-gray-600">Sign in to your account</p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-gray-700">
                Username
              </label>
              <input
                id="username"
                name="username"
                type="text"
                autoComplete="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 ${
                  getFieldError('username')
                    ? 'border-red-300 focus:border-red-500'
                    : 'border-gray-300 focus:border-blue-500'
                }`}
                disabled={isLoading}
              />
              {getFieldError('username') && (
                <p className="mt-1 text-sm text-red-600">{getFieldError('username')}</p>
              )}
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                Password
              </label>
              <div className="mt-1 relative">
                <input
                  id="password"
                  name="password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className={`block w-full px-3 py-2 pr-12 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 placeholder-gray-400 ${
                    getFieldError('password')
                      ? 'border-red-300 focus:border-red-500'
                      : 'border-gray-300 focus:border-blue-500'
                  }`}
                  disabled={isLoading}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600 focus:outline-none"
                  tabIndex={-1}
                >
                  {showPassword ? (
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                  ) : (
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                  )}
                </button>
              </div>
              {getFieldError('password') && (
                <p className="mt-1 text-sm text-red-600">{getFieldError('password')}</p>
              )}
            </div>
          </div>

          {apiError && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
              <p className="text-sm">{apiError}</p>
            </div>
          )}

          {accountLockedMessage && (
            <div className="bg-yellow-50 border border-yellow-200 text-yellow-700 px-4 py-3 rounded-md">
              <p className="text-sm">{accountLockedMessage}</p>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className={`w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-lg text-base font-semibold text-white focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-all duration-200 ${
                isLoading
                  ? 'bg-gray-400 cursor-not-allowed'
                  : 'bg-blue-600 hover:bg-blue-700 active:bg-blue-800'
              }`}
            >
              {isLoading ? (
                <>
                  <svg
                    className="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  Signing in...
                </>
              ) : (
                'Sign in'
              )}
            </button>
          </div>
        </form>

        {isAuthenticated && (
          <div className="mt-6">
            <button
              type="button"
              onClick={async () => {
                try {
                  await logout()
                  await storeLogout()
                  showToast('success', 'Logout successful', 'You have been logged out')
                  navigate('/login')
                } catch {
                  showToast('error', 'Logout failed', 'Please try again later')
                }
              }}
              className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-gray-600 hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500"
            >
              Sign Out
            </button>
          </div>
        )}

        {toasts.map((toast) => (
          <ToastNotification
            key={toast.id}
            {...toast}
            onClose={handleToastClose}
          />
        ))}
      </div>
    </div>
  )
}
