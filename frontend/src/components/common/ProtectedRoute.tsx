import { Component, type ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'

interface ProtectedRouteProps {
  children: React.ReactNode
}

// Track redirect count for loop protection
const REDIRECT_COUNT_KEY = 'auth:redirect_count'
const MAX_REDIRECTS = 3

/**
 * Check if redirect should be allowed
 * Returns true if redirect count is below threshold
 */
function checkRedirectCount(): boolean {
  const redirectCount = parseInt(sessionStorage.getItem(REDIRECT_COUNT_KEY) || '0', 10)
  return redirectCount < MAX_REDIRECTS
}

/**
 * Increment redirect count in sessionStorage
 */
function incrementRedirectCount(): void {
  const redirectCount = parseInt(sessionStorage.getItem(REDIRECT_COUNT_KEY) || '0', 10)
  sessionStorage.setItem(REDIRECT_COUNT_KEY, String(redirectCount + 1))
}

/**
 * Clear redirect count
 */
function clearRedirectCount(): void {
  sessionStorage.removeItem(REDIRECT_COUNT_KEY)
}

/**
 * Error Boundary for catching crashes in protected routes
 * Shows error UI on component crash - does NOT logout user
 * Component errors should not affect authentication state
 */
class AuthErrorBoundary extends Component<
  { children: React.ReactNode },
  { hasError: boolean }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError() {
    return { hasError: true }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('[ProtectedRoute] Component error caught by boundary:', error, errorInfo)
    // DO NOT clear auth state - component errors are not auth errors
    // Let the error UI handle recovery
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
          <div className="text-center max-w-md px-4">
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md mb-4">
              <p className="text-sm font-medium">Something went wrong</p>
              <p className="text-sm mt-1">The page encountered an error. You can try refreshing.</p>
            </div>
            <button
              onClick={() => window.location.reload()}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            >
              Refresh Page
            </button>
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
  const { isAuthenticated, isLoading, clearAuth } = useAuthStore()

  // Show loading indicator while session is being restored
  // Prevents content flash and premature redirects
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

  // Only check authentication after loading is complete
  if (!isAuthenticated) {
    // Check if we've exceeded redirect count (what was happening before)
    if (!checkRedirectCount()) {
      clearRedirectCount()
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

    incrementRedirectCount()

    return <Navigate to="/login" state={{ from: location }} replace />
  }

  // User is authenticated, allow access
  // Token validation happens server-side via 401 interceptor

  // Clear redirect counter on successful auth
  clearRedirectCount()

  return (
    <AuthErrorBoundary>
      {children as ReactNode}
    </AuthErrorBoundary>
  )
}

export { ProtectedRoute }
