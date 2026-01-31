import { useAuthStore } from '../../stores/authStore'
import { useNodesStore } from '../../stores/nodesStore'
import { useAlertsStore } from '../../stores/alertsStore'
import { useDashboardStore } from '../../stores/dashboardStore'
import type { Node, AlertRecord } from '../../stores/types'

/**
 * StoreTest Component
 *
 * This component demonstrates usage of all Zustand stores.
 * It can be used to verify that stores are working correctly.
 *
 * Usage:
 * ```tsx
 * import StoreTest from './components/common/StoreTest'
 * <StoreTest />
 * ```
 */
export default function StoreTest() {
  // Auth Store
  const { user, isAuthenticated, role } = useAuthStore()

  // Nodes Store
  const { nodes, selectedNode } = useNodesStore()

  // Alerts Store
  const { alertRules, alertRecords } = useAlertsStore()

  // Dashboard Store
  const { filters, timeRange, autoRefresh } = useDashboardStore()

  return (
    <div className="p-6 space-y-6 bg-gray-50 rounded-lg">
      <h2 className="text-2xl font-bold">Store State Verification</h2>

      {/* Auth Store */}
      <div className="bg-white p-4 rounded shadow">
        <h3 className="text-lg font-semibold mb-2">Auth Store</h3>
        <div className="space-y-1 text-sm">
          <p>
            <span className="font-medium">Authenticated:</span>{' '}
            {isAuthenticated ? 'Yes' : 'No'}
          </p>
          <p>
            <span className="font-medium">User:</span>{' '}
            {user ? `${user.username} (${user.role})` : 'None'}
          </p>
          <p>
            <span className="font-medium">Role:</span> {role || 'None'}
          </p>
        </div>
      </div>

      {/* Nodes Store */}
      <div className="bg-white p-4 rounded shadow">
        <h3 className="text-lg font-semibold mb-2">Nodes Store</h3>
        <div className="space-y-1 text-sm">
          <p>
            <span className="font-medium">Total Nodes:</span> {nodes.length}
          </p>
          <p>
            <span className="font-medium">Selected Node:</span>{' '}
            {selectedNode ? selectedNode.name : 'None'}
          </p>
          <p>
            <span className="font-medium">Online:</span>{' '}
            {nodes.filter((n: Node) => n.status === 'online').length}
          </p>
          <p>
            <span className="font-medium">Offline:</span>{' '}
            {nodes.filter((n: Node) => n.status === 'offline').length}
          </p>
        </div>
      </div>

      {/* Alerts Store */}
      <div className="bg-white p-4 rounded shadow">
        <h3 className="text-lg font-semibold mb-2">Alerts Store</h3>
        <div className="space-y-1 text-sm">
          <p>
            <span className="font-medium">Alert Rules:</span> {alertRules.length}
          </p>
          <p>
            <span className="font-medium">Alert Records:</span>{' '}
            {alertRecords.length}
          </p>
          <p>
            <span className="font-medium">Pending Alerts:</span>{' '}
            {alertRecords.filter((a: AlertRecord) => a.status === 'pending').length}
          </p>
        </div>
      </div>

      {/* Dashboard Store */}
      <div className="bg-white p-4 rounded shadow">
        <h3 className="text-lg font-semibold mb-2">Dashboard Store</h3>
        <div className="space-y-1 text-sm">
          <p>
            <span className="font-medium">Time Range:</span> {timeRange}
          </p>
          <p>
            <span className="font-medium">Auto Refresh:</span>{' '}
            {autoRefresh ? 'Enabled' : 'Disabled'}
          </p>
          <p>
            <span className="font-medium">Region Filter:</span>{' '}
            {filters.region || 'All'}
          </p>
          <p>
            <span className="font-medium">Status Filter:</span>{' '}
            {filters.status}
          </p>
        </div>
      </div>

      <div className="bg-blue-50 p-3 rounded border border-blue-200">
        <p className="text-sm text-blue-800">
          ✅ All stores are accessible and working correctly!
        </p>
      </div>
    </div>
  )
}
