import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import { PageHeader } from '@/components/layout/PageHeader'
import { fetchUsers, createUser, updateUser, deleteUser } from '@/api/users'
import type { UserDTO, CreateUserRequest, UpdateUserRequest } from '@/api/users'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

interface UserDialogState { mode: 'create' | 'edit'; user?: UserDTO }

export default function UsersPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const [users, setUsers] = useState<UserDTO[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [userToDelete, setUserToDelete] = useState<string>()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogState, setDialogState] = useState<UserDialogState>({ mode: 'create' })
  const [dialogLoading, setDialogLoading] = useState(false)
  const [dialogError, setDialogError] = useState<string | null>(null)
  const [formUsername, setFormUsername] = useState('')
  const [formEmail, setFormEmail] = useState('')
  const [formPassword, setFormPassword] = useState('')
  const [formRole, setFormRole] = useState<'admin' | 'operator' | 'viewer'>('viewer')

  const isAdmin = user?.role === 'admin'

  const loadUsers = useCallback(async () => {
    try { setIsLoading(true); setError(null); const res = await fetchUsers(); setUsers(res.data.users) }
    catch (err) { setError(err instanceof Error ? err.message : String(err)) }
    finally { setIsLoading(false) }
  }, [])

  useEffect(() => { if (isAdmin) void loadUsers() }, [isAdmin, loadUsers])

  if (!isAdmin) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('settings.users')} />
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-destructive font-semibold">{t('errors.accessDenied')}</p>
            <p className="text-sm text-muted-foreground mt-1">{t('errors.adminOnly')}</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  const openCreateDialog = () => {
    setDialogState({ mode: 'create' }); setFormUsername(''); setFormEmail(''); setFormPassword(''); setFormRole('viewer'); setDialogError(null); setDialogOpen(true)
  }

  const openEditDialog = (u: UserDTO) => {
    setDialogState({ mode: 'edit', user: u }); setFormUsername(u.username); setFormEmail(u.email ?? ''); setFormPassword(''); setFormRole(u.role); setDialogError(null); setDialogOpen(true)
  }

  const handleDialogSubmit = async () => {
    setDialogLoading(true); setDialogError(null)
    try {
      if (dialogState.mode === 'create') {
        if (!formUsername || !formPassword) { setDialogError(t('settings.usernamePasswordRequired')); setDialogLoading(false); return }
        const req: CreateUserRequest = { username: formUsername, password: formPassword, role: formRole }
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
      setDialogOpen(false); await loadUsers()
    } catch (err) { setDialogError(err instanceof Error ? err.message : String(err)) }
    finally { setDialogLoading(false) }
  }

  const handleRoleChange = async (userId: string, newRole: 'admin' | 'operator' | 'viewer') => {
    try { await updateUser(userId, { role: newRole }); await loadUsers() }
    catch (err) { setError(err instanceof Error ? err.message : String(err)) }
  }

  const handleDelete = (userId: string) => { setUserToDelete(userId); setDeleteConfirmOpen(true) }

  const confirmDelete = async () => {
    if (!userToDelete) return
    try { await deleteUser(userToDelete); await loadUsers() }
    catch (err) { setError(err instanceof Error ? err.message : String(err)) }
    finally { setDeleteConfirmOpen(false); setUserToDelete(undefined) }
  }

  return (
    <div className="space-y-6">
      <PageHeader title={t('settings.users')} subtitle={t('settings.usersDescription')} actions={<Button onClick={openCreateDialog}>{t('settings.addUser')}</Button>} />

      {error && <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

      {isLoading ? (
        <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      ) : (
        <Card>
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-border">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.username')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.email')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.role')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.status')}</th>
                    <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.actions')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {users.map((u) => {
                    const isLocked = !!u.locked_until
                    return (
                      <tr key={u.user_id} className="hover:bg-muted/50">
                        <td className="whitespace-nowrap px-6 py-4 text-sm font-medium">
                          {u.username} {u.mfa_enabled && <Badge variant="secondary" className="ml-1">MFA</Badge>}
                        </td>
                        <td className="whitespace-nowrap px-6 py-4 text-sm text-muted-foreground">{u.email ?? '—'}</td>
                        <td className="whitespace-nowrap px-6 py-4">
                          <select value={u.role} onChange={(e) => handleRoleChange(u.user_id, e.target.value as 'admin' | 'operator' | 'viewer')} disabled={u.user_id === user?.id}
                            className="rounded-md border border-input bg-background px-2 py-1 text-sm">
                            <option value="admin">Admin</option><option value="operator">Operator</option><option value="viewer">Viewer</option>
                          </select>
                        </td>
                        <td className="whitespace-nowrap px-6 py-4">
                          <Badge variant={isLocked ? 'destructive' : 'default'}>{isLocked ? t('settings.locked') : t('settings.active')}</Badge>
                        </td>
                        <td className="whitespace-nowrap px-6 py-4 text-right text-sm space-x-2">
                          <Button variant="link" size="sm" onClick={() => openEditDialog(u)}>{t('settings.edit')}</Button>
                          <Button variant="link" size="sm" className="text-destructive" onClick={() => handleDelete(u.user_id)} disabled={u.user_id === user?.id}>{t('settings.delete')}</Button>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
            {users.length === 0 && <div className="py-12 text-center text-sm text-muted-foreground">{t('settings.noUsers')}</div>}
          </CardContent>
        </Card>
      )}

      {dialogOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="fixed inset-0 bg-black/50" onClick={() => setDialogOpen(false)} />
            <Card className="relative w-full max-w-md">
              <CardContent className="p-6">
                <h3 className="mb-4 text-lg font-semibold">{dialogState.mode === 'create' ? t('settings.addUser') : t('settings.editUser')}</h3>
                {dialogError && <div className="mb-4 rounded-md bg-destructive/10 px-4 py-2 text-sm text-destructive">{dialogError}</div>}
                <div className="space-y-4">
                  <div className="space-y-2"><Label>{t('settings.username')} *</Label><Input value={formUsername} onChange={(e) => setFormUsername(e.target.value)} /></div>
                  <div className="space-y-2"><Label>{t('settings.email')}</Label><Input type="email" value={formEmail} onChange={(e) => setFormEmail(e.target.value)} /></div>
                  <div className="space-y-2">
                    <Label>{t('settings.password')}{dialogState.mode === 'create' && ' *'}</Label>
                    <Input type="password" value={formPassword} onChange={(e) => setFormPassword(e.target.value)} placeholder={dialogState.mode === 'edit' ? t('settings.leaveBlankToKeep') : ''} />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('settings.role')} *</Label>
                    <select value={formRole} onChange={(e) => setFormRole(e.target.value as 'admin' | 'operator' | 'viewer')} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                      <option value="admin">Admin</option><option value="operator">Operator</option><option value="viewer">Viewer</option>
                    </select>
                  </div>
                </div>
                <div className="mt-6 flex justify-end gap-3">
                  <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
                  <Button onClick={() => void handleDialogSubmit()} disabled={dialogLoading}>{dialogLoading ? t('common.saving') : t('common.save')}</Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      )}

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('settings.confirmDeleteUser')}</AlertDialogTitle>
            <AlertDialogDescription>{t('settings.confirmDeleteUserMessage')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => { setDeleteConfirmOpen(false); setUserToDelete(undefined) }}>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">{t('common.delete')}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
