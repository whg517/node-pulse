/**
 * Page Object Models
 *
 * All page objects for E2E testing:
 * - Common: Base classes for building page objects
 * - Auth: Login, Sessions
 * - Dashboard: Dashboard, Performance
 * - Nodes: Nodes, NodeDetail, NodeComparison
 * - Alerts: AlertRules, AlertRecords, AlertHistory
 * - Webhooks: Webhooks
 * - Export: DataExport
 * - Settings: Reports, Users, Preferences, SystemHealth
 */

// Common base classes
export { BasePage, type PageSelectors, DEFAULT_SELECTORS } from './common/BasePage'
export { TablePage, type TableSelectors, DEFAULT_TABLE_SELECTORS } from './common/TablePage'
export { ModalPage, type ModalSelectors, DEFAULT_MODAL_SELECTORS } from './common/ModalPage'
export { FormPage, type FormSelectors, type FormField, DEFAULT_FORM_SELECTORS } from './common/FormPage'

// Auth pages
export { LoginPage, type LoginSelectors, DEFAULT_LOGIN_SELECTORS } from './LoginPage'
export { SessionsPage, type SessionsSelectors, DEFAULT_SESSIONS_SELECTORS } from './SessionsPage'

// Dashboard pages
export { DashboardPage } from './DashboardPage'
export { PerformancePage } from './PerformancePage'

// Nodes pages
export { NodesPage, type NodesSelectors, DEFAULT_NODES_SELECTORS } from './NodesPage'
export { NodeDetailPage } from './NodeDetailPage'
export { NodeComparisonPage } from './NodeComparisonPage'

// Alerts pages - use original implementations
export {
  AlertRulesPage,
  AlertRecordsPage,
  AlertHistoryPage,
} from './AlertsPage'

// Webhooks pages
export { WebhooksPage, type WebhooksSelectors, DEFAULT_WEBHOOKS_SELECTORS } from './WebhooksPage'

// Export pages
export { DataExportPage, type DataExportSelectors, DEFAULT_DATA_EXPORT_SELECTORS } from './DataExportPage'

// Settings pages
export { ReportsPage, type ReportsSelectors, DEFAULT_REPORTS_SELECTORS } from './ReportsPage'
export { UsersPage, type UsersSelectors, DEFAULT_USERS_SELECTORS } from './UsersPage'
export { PreferencesPage, type PreferencesSelectors, DEFAULT_PREFERENCES_SELECTORS } from './PreferencesPage'
export { SystemHealthPage, type SystemHealthSelectors, DEFAULT_SYSTEM_HEALTH_SELECTORS } from './SystemHealthPage'
