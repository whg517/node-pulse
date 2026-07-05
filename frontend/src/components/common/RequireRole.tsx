import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../../stores/authStore'
import { Card, CardContent } from '../ui/card'

interface RequireRoleProps {
  /** Roles allowed to view this route's content. Others see an "insufficient
   * permissions" panel instead. */
  roles: string[]
  children: ReactNode
}

/**
 * Route-level RBAC guard.
 *
 * The app has no central role-based router: ProtectedRoute only checks
 * authentication, not authorization, so every page historically had to
 * remember to gate its own write buttons (and the Reports schedule button
 * famously forgot — the F1 bug). This component is the centralized guard:
 * wrap admin-only pages (Users, API Keys, Audit Logs, System Config,
 * Webhooks) so a Viewer who lands on the URL gets a clear "insufficient
 * permissions" panel instead of a page full of buttons that 403 on click.
 *
 * Read-only pages (Dashboard, Nodes, Alerts, Reports view) stay open to
 * all roles per the documented RBAC matrix (user-journey.md §1.3): a
 * Viewer is allowed to see the data; only mutations are restricted.
 */
export default function RequireRole({ roles, children }: RequireRoleProps) {
  const { t } = useTranslation()
  const { role } = useAuthStore()

  if (role && !roles.includes(role)) {
    return (
      <div className="flex items-center justify-center py-12">
        <Card className="max-w-md">
          <CardContent className="p-6 text-center">
            <p className="text-sm font-medium text-foreground">
              {t('common.insufficientPermissions', 'Insufficient permissions')}
            </p>
            <p className="mt-2 text-sm text-muted-foreground">
              {t(
                'common.insufficientPermissionsHint',
                'You do not have a role permitted to view this page.'
              )}
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return <>{children}</>
}

export { RequireRole }
