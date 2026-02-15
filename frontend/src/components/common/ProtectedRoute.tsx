import { Component, type ReactNode, useMemo } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'

interface ProtectedRouteProps {
  children: React.ReactNode
}

// Track redirect count for loop protection
const REDIRECT_COUNT_KEY = 'auth:redirect_count'
const MAX_REDIRECTS = 3

/**
 * Error Boundary for catching crashes in protected routes
 */
class AuthErrorBoundary extends Component<
  { children: React.ReactNode; onLogout: () => void },
  { hasError: boolean }
> {
  constructor(props: { children: React.ReactNode; onLogout: () => void }) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError() {
    return { hasError: true }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('[ProtectedRoute] Error caught by boundary:', error, errorInfo)
    // Clear auth state - this is the proper place for side effects
    this.props.onLogout()
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
          <div className="text-center">
            <p className="text-gray-600">Something went wrong. Redirecting to login...</p>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

/**
 * ProtectedRoute Component
 *
 * Checks if user is authenticated before allowing access to protected routes.
 * Redirects to /login if not authenticated, storing the original location for post-login redirect.
 *
 * Features:
 * - Loading state while session is being restored
 * - Error boundary for crash recovery
 * - Redirect loop protection (force logout after 3+ redirects)
 * - Prevents content flash by showing loading until auth state is known
 *
 * Note: Actual token validation happens server-side via API calls.
 * Client-side checks are for UX optimization only.
 */
export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const location = useLocation()
  const { isAuthenticated, tokenExpiresAt, isLoading, clearAuth } = useAuthStore()

  // Show loading indicator while session is being restored
  // Prevents content flash
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-solid border-current border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite] text-blue-600">
            <span className="!absolute !-m-px !h-px !w-px !overflow-hidden !whitespace-nowrap !border-0 !p-0 ![clip:rect(0,0,0,0)]">
              Loading...
            </span>
          </div>
          <p className="mt-4 text-gray-600">Restoring session...</p>
        </div>
      </div>
    )
  }

  // Check if token exists and is not expired
  const isValid = useMemo(() => {
    return tokenExpiresAt !== null && tokenExpiresAt > Date.now()
  }, [tokenExpiresAt])

  if (!isAuthenticated || !isValid) {
    // Redirect loop protection
    const redirectCount = parseInt(sessionStorage.getItem(REDIRECT_COUNT_KEY) || '0', 10)

    if (redirectCount >= MAX_REDIRECTS) {
      // Clear the counter and force logout
      sessionStorage.removeItem(REDIRECT_COUNT_KEY)
      clearAuth()

      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
          <div className="text-center max-w-md">
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
              <p className="text-sm">Authentication error detected. Please log in again.</p>
            </div>
            <button
              onClick={() => {
                clearAuth()
                window.location.href = '/login'
              }}
              className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              Go to Login
            </button>
          </div>
        </div>
      )
    }

    // Increment redirect counter
    sessionStorage.setItem(REDIRECT_COUNT_KEY, String(redirectCount + 1))

    // Store the original location for post-login redirect
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  // Clear redirect counter on successful auth
  sessionStorage.removeItem(REDIRECT_COUNT_KEY)

  return (
    <AuthErrorBoundary onLogout={clearAuth}>
      {children as ReactNode}
    </AuthErrorBoundary>
  )
}
