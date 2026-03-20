/**
 * Users Page
 *
 * User management page for administrators.
 * Route: /settings/users (Admin only)
 */

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../stores/authStore'
import { PageContainer, ConfirmDialog } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'

interface User {
  id: string
  username: string
  email: string
  role: 'admin' | 'operator' | 'viewer'
  status: 'active' | 'disabled'
  createdAt: string
  lastLogin?: string
}

// TODO: Replace mock data with real API calls to /api/v1/users
// Mock users for demonstration (placeholder until user management API is implemented)
const mockUsers: User[] = [
  {
    id: '1',
    username: 'admin',
    email: 'admin@nodepulse.io',
    role: 'admin',
    status: 'active',
    createdAt: '2026-01-01T00:00:00Z',
    lastLogin: '2026-02-22T10:00:00Z',
  },
  {
    id: '2',
    username: 'operator1',
    email: 'operator@nodepulse.io',
    role: 'operator',
    status: 'active',
    createdAt: '2026-01-15T00:00:00Z',
    lastLogin: '2026-02-21T15:30:00Z',
  },
  {
    id: '3',
    username: 'viewer1',
    email: 'viewer@nodepulse.io',
    role: 'viewer',
    status: 'active',
    createdAt: '2026-02-01T00:00:00Z',
    lastLogin: '2026-02-20T09:00:00Z',
  },
]

export default function UsersPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const [users, setUsers] = useState<User[]>(mockUsers)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [userToDelete, setUserToDelete] = useState<string | undefined>(undefined)

  // Check if current user is admin
  const isAdmin = user?.role === 'admin'

  if (!isAdmin) {
    return (
      <PageContainer>
        <PageHeader title={t('settings.users')} />
        <div className="rounded-lg border p-8 text-center bg-[var(--color-bg-surface)] border-[var(--color-border)]">
          <svg className="mx-auto h-12 w-12 text-[var(--color-critical)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <h2 className="mt-4 text-xl font-semibold text-[var(--color-text-primary)]">
            {t('errors.accessDenied')}
          </h2>
          <p className="mt-2 text-[var(--color-text-secondary)]">
            {t('errors.adminOnly')}
          </p>
        </div>
      </PageContainer>
    )
  }

  const handleRoleChange = (userId: string, newRole: User['role']) => {
    setUsers((prev) =>
      prev.map((u) => (u.id === userId ? { ...u, role: newRole } : u))
    )
  }

  const handleToggleStatus = (userId: string) => {
    setUsers((prev) =>
      prev.map((u) =>
        u.id === userId
          ? { ...u, status: u.status === 'active' ? 'disabled' : 'active' }
          : u
      )
    )
  }

  const handleDelete = (userId: string) => {
    setUserToDelete(userId)
    setDeleteConfirmOpen(true)
  }

  const confirmDelete = () => {
    if (userToDelete) {
      setUsers((prev) => prev.filter((u) => u.id !== userToDelete))
    }
    setDeleteConfirmOpen(false)
    setUserToDelete(undefined)
  }

  const getStatusStyles = (status: User['status']): string => {
    return status === 'active'
      ? 'text-[var(--color-healthy)]'
      : 'text-[var(--color-critical)]'
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('settings.users')}
        subtitle={t('settings.usersDescription')}
        actions={
          <button
            type="button"
            className="px-4 py-2 bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] text-white text-sm font-medium rounded-lg transition-colors"
          >
            {t('settings.addUser')}
          </button>
        }
      />

      <div className="rounded-lg border shadow-sm overflow-hidden bg-[var(--color-bg-surface)] border-[var(--color-border)]">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-[var(--color-border)]">
            <thead className="bg-[var(--color-bg-muted)]">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  {t('settings.username')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  {t('settings.email')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  {t('settings.role')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  {t('settings.status')}
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  {t('settings.actions')}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--color-border)]">
              {users.map((u) => (
                <tr key={u.id} className="hover:bg-[var(--color-hover-overlay)]">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-[var(--color-text-primary)]">
                    {u.username}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-[var(--color-text-secondary)]">
                    {u.email}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <select
                      value={u.role}
                      onChange={(e) => handleRoleChange(u.id, e.target.value as User['role'])}
                      className="text-sm rounded-md border px-2 py-1 bg-[var(--color-input-bg)] border-[var(--color-input-border)] text-[var(--color-text-primary)]"
                      disabled={u.id === '1'}
                    >
                      <option value="admin">Admin</option>
                      <option value="operator">Operator</option>
                      <option value="viewer">Viewer</option>
                    </select>
                  </td>
                  <td className={`px-6 py-4 whitespace-nowrap text-sm font-medium ${getStatusStyles(u.status)}`}>
                    {u.status === 'active' ? t('settings.active') : t('settings.disabled')}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                    <button
                      type="button"
                      onClick={() => handleToggleStatus(u.id)}
                      className={`${
                        u.status === 'active'
                          ? 'text-[var(--color-warning)] hover:text-[var(--color-warning)] hover:opacity-80'
                          : 'text-[var(--color-healthy)] hover:text-[var(--color-healthy)] hover:opacity-80'
                      }`}
                      disabled={u.id === '1'}
                    >
                      {u.status === 'active' ? t('settings.disable') : t('settings.enable')}
                    </button>
                    <button
                      type="button"
                      onClick={() => handleDelete(u.id)}
                      className="text-[var(--color-critical)] hover:text-[var(--color-critical)] hover:opacity-80"
                      disabled={u.id === '1'}
                    >
                      {t('settings.delete')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <ConfirmDialog
        open={deleteConfirmOpen}
        title={t('settings.confirmDeleteUser')}
        message={t('settings.confirmDeleteUserMessage')}
        confirmText={t('common.delete')}
        onConfirm={confirmDelete}
        onCancel={() => {
          setDeleteConfirmOpen(false)
          setUserToDelete(undefined)
        }}
        loading={false}
        variant="danger"
      />
    </PageContainer>
  )
}
