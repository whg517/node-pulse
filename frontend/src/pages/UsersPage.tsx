import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../stores/authStore'
import { PageContainer, ConfirmDialog, ErrorBanner, LoadingSpinner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { fetchUsers, createUser, updateUser, deleteUser } from '../api/users'
import type { UserDTO, CreateUserRequest, UpdateUserRequest } from '../api/users'

interface UserDialogState {
  mode: 'create' | 'edit'
  user?: UserDTO
}

export default function UsersPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const [users, setUsers] = useState<UserDTO[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [userToDelete, setUserToDelete] = useState<string | undefined>(undefined)
  const [deleteLoading, setDeleteLoading] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogState, setDialogState] = useState<UserDialogState>({ mode: 'create' })
  const [dialogLoading, setDialogLoading] = useState(false)
  const [dialogError, setDialogError] = useState<string | null>(null)

  // Form state
  const [formUsername, setFormUsername] = useState('')
  const [formEmail, setFormEmail] = useState('')
  const [formPassword, setFormPassword] = useState('')
  const [formRole, setFormRole] = useState<'admin' | 'operator' | 'viewer'>('viewer')

  const isAdmin = user?.role === 'admin'

  const loadUsers = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const res = await fetchUsers()
      setUsers(res.data.users)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (isAdmin) void loadUsers()
  }, [isAdmin, loadUsers])

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

  const openCreateDialog = () => {
    setDialogState({ mode: 'create' })
    setFormUsername('')
    setFormEmail('')
    setFormPassword('')
    setFormRole('viewer')
    setDialogError(null)
    setDialogOpen(true)
  }

  const openEditDialog = (u: UserDTO) => {
    setDialogState({ mode: 'edit', user: u })
    setFormUsername(u.username)
    setFormEmail(u.email ?? '')
    setFormPassword('')
    setFormRole(u.role)
    setDialogError(null)
    setDialogOpen(true)
  }

  const handleDialogSubmit = async () => {
    setDialogLoading(true)
    setDialogError(null)
    try {
      if (dialogState.mode === 'create') {
        if (!formUsername || !formPassword) {
          setDialogError(t('settings.usernamePasswordRequired'))
          setDialogLoading(false)
          return
        }
        const req: CreateUserRequest = {
          username: formUsername,
          password: formPassword,
          role: formRole,
        }
        if (formEmail) req.email = formEmail
        await createUser(req)
      } else if (dialogState.user) {
        const req: UpdateUserRequest = {}
        if (formUsername !== dialogState.user.username) req.username = formUsername
        if (formEmail !== (dialogState.user.email ?? '')) req.email = formEmail
        if (formPassword) req.password = formPassword
        if (formRole !== dialogState.user.role) req.role = formRole
        await updateUser(dialogState.user.user_id, req)
      }
      setDialogOpen(false)
      await loadUsers()
    } catch (err) {
      setDialogError(err instanceof Error ? err.message : String(err))
    } finally {
      setDialogLoading(false)
    }
  }

  const handleRoleChange = async (userId: string, newRole: 'admin' | 'operator' | 'viewer') => {
    try {
      await updateUser(userId, { role: newRole })
      await loadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleDelete = (userId: string) => {
    setUserToDelete(userId)
    setDeleteConfirmOpen(true)
  }

  const confirmDelete = async () => {
    if (!userToDelete) return
    setDeleteLoading(true)
    try {
      await deleteUser(userToDelete)
      await loadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setDeleteLoading(false)
      setDeleteConfirmOpen(false)
      setUserToDelete(undefined)
    }
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('settings.users')}
        subtitle={t('settings.usersDescription')}
        actions={
          <button
            type="button"
            onClick={openCreateDialog}
            className="px-4 py-2 bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] text-white text-sm font-medium rounded-lg transition-colors"
          >
            {t('settings.addUser')}
          </button>
        }
      />

      {error && <ErrorBanner error={new Error(error)} onRetry={loadUsers} className="mb-4" />}

      {isLoading ? (
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      ) : (
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
                {users.map((u) => {
                  const isLocked = !!u.locked_until
                  return (
                    <tr key={u.user_id} className="hover:bg-[var(--color-hover-overlay)]">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-[var(--color-text-primary)]">
                        {u.username}
                        {u.mfa_enabled && (
                          <span className="ml-2 text-xs text-[var(--color-brand)]">MFA</span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-[var(--color-text-secondary)]">
                        {u.email ?? '—'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <select
                          value={u.role}
                          onChange={(e) => handleRoleChange(u.user_id, e.target.value as 'admin' | 'operator' | 'viewer')}
                          className="text-sm rounded-md border px-2 py-1 bg-[var(--color-input-bg)] border-[var(--color-input-border)] text-[var(--color-text-primary)]"
                          disabled={u.user_id === user?.id}
                        >
                          <option value="admin">Admin</option>
                          <option value="operator">Operator</option>
                          <option value="viewer">Viewer</option>
                        </select>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium ${isLocked ? 'bg-[var(--color-critical-bg)] text-[var(--color-critical)]' : 'bg-[var(--color-healthy-bg)] text-[var(--color-healthy)]'}`}>
                          <span className={`w-1.5 h-1.5 rounded-full ${isLocked ? 'bg-[var(--color-critical)]' : 'bg-[var(--color-healthy)]'}`} />
                          {isLocked ? t('settings.locked') : t('settings.active')}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                        <button
                          type="button"
                          onClick={() => openEditDialog(u)}
                          className="text-[var(--color-brand)] hover:opacity-80"
                        >
                          {t('settings.edit')}
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(u.user_id)}
                          className="text-[var(--color-critical)] hover:opacity-80"
                          disabled={u.user_id === user?.id}
                        >
                          {t('settings.delete')}
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
          {users.length === 0 && (
            <div className="py-12 text-center text-sm text-[var(--color-text-secondary)]">
              {t('settings.noUsers')}
            </div>
          )}
        </div>
      )}

      {/* Create/Edit Dialog */}
      {dialogOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="fixed inset-0 bg-black/50" onClick={() => setDialogOpen(false)} />
            <div className="relative w-full max-w-md rounded-lg bg-[var(--color-bg-surface)] border border-[var(--color-border)] shadow-xl p-6">
              <h3 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
                {dialogState.mode === 'create' ? t('settings.addUser') : t('settings.editUser')}
              </h3>

              {dialogError && (
                <div className="mb-4 rounded-lg bg-[var(--color-critical-bg)] text-[var(--color-critical)] px-4 py-2 text-sm">
                  {dialogError}
                </div>
              )}

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                    {t('settings.username')} <span className="text-[var(--color-critical)]">*</span>
                  </label>
                  <input
                    type="text"
                    value={formUsername}
                    onChange={(e) => setFormUsername(e.target.value)}
                    className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:ring-2 focus:ring-[var(--color-brand)]"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                    {t('settings.email')}
                  </label>
                  <input
                    type="email"
                    value={formEmail}
                    onChange={(e) => setFormEmail(e.target.value)}
                    className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:ring-2 focus:ring-[var(--color-brand)]"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                    {t('settings.password')}
                    {dialogState.mode === 'create' && <span className="text-[var(--color-critical)]">*</span>}
                  </label>
                  <input
                    type="password"
                    value={formPassword}
                    onChange={(e) => setFormPassword(e.target.value)}
                    placeholder={dialogState.mode === 'edit' ? t('settings.leaveBlankToKeep') : ''}
                    className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:ring-2 focus:ring-[var(--color-brand)]"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                    {t('settings.role')} <span className="text-[var(--color-critical)]">*</span>
                  </label>
                  <select
                    value={formRole}
                    onChange={(e) => setFormRole(e.target.value as 'admin' | 'operator' | 'viewer')}
                    className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
                  >
                    <option value="admin">Admin</option>
                    <option value="operator">Operator</option>
                    <option value="viewer">Viewer</option>
                  </select>
                </div>
              </div>

              <div className="mt-6 flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setDialogOpen(false)}
                  className="px-4 py-2 text-sm font-medium rounded-lg border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="button"
                  onClick={() => void handleDialogSubmit()}
                  disabled={dialogLoading}
                  className="px-4 py-2 text-sm font-medium rounded-lg bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] text-white disabled:opacity-50"
                >
                  {dialogLoading ? t('common.saving') : t('common.save')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deleteConfirmOpen}
        title={t('settings.confirmDeleteUser')}
        message={t('settings.confirmDeleteUserMessage')}
        confirmText={t('common.delete')}
        onConfirm={() => void confirmDelete()}
        onCancel={() => {
          setDeleteConfirmOpen(false)
          setUserToDelete(undefined)
        }}
        loading={deleteLoading}
        variant="danger"
      />
    </PageContainer>
  )
}
