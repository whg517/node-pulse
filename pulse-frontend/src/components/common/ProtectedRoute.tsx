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
 *
 * Note: Actual session validation happens server-side via API calls.
 * Client-side checks are for UX optimization only.
 */
export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const location = useLocation()
  const { isAuthenticated, checkSession } = useAuthStore()
  const isValid = checkSession()

  if (!isAuthenticated || !isValid) {
    // Store the original location for post-login redirect
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return <>{children}</>
}
