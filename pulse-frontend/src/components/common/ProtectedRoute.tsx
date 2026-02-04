import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'

interface ProtectedRouteProps {
  children: React.ReactNode
}

/**
 * ProtectedRoute Component
 *
 * Checks if user is authenticated before allowing access to protected routes.
 * Redirects to /login if not authenticated, storing the original location for post-login redirect.
 *
 * Authentication checks:
 * 1. User has valid authentication state in Zustand store
 * 2. Session has not expired
 * 3. Session restoration is complete (isLoading is false)
 *
 * Note: Actual session validation happens server-side via API calls.
 * Client-side checks are for UX optimization only.
 */
export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const location = useLocation()
  const { isAuthenticated, checkSession, isLoading } = useAuthStore()

  // Show loading indicator while session is being restored
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

  const isValid = checkSession()

  if (!isAuthenticated || !isValid) {
    // Store the original location for post-login redirect
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return <>{children}</>
}
