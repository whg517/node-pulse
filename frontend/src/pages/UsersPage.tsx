import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import { PageHeader } from '@/components/layout/PageHeader'
import { fetchUsers, createUser, updateUser, deleteUser } from '@/api/users'
import { adminRevokeAllUserSessions } from '@/api/adminAuth'
import type { UserDTO, CreateUserRequest, UpdateUserRequest } from '@/api/users'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
  const [forceLogoutUser, setForceLogoutUser] = useState<{ id: string; name: string } | null>(null)
  const [forceLogoutLoading, setForceLogoutLoading] = useState(false)
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

  const handleForceLogout = (u: UserDTO) => {
    setForceLogoutUser({ id: u.user_id, name: u.username })
  }

  const confirmForceLogout = async () => {
    if (!forceLogoutUser) return
    setForceLogoutLoading(true)
    try {
      await adminRevokeAllUserSessions(forceLogoutUser.id)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setForceLogoutLoading(false)
      setForceLogoutUser(null)
    }
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
                          <Button variant="link" size="sm" onClick={() => handleForceLogout(u)} disabled={u.user_id === user?.id}>
                            {t('settings.forceLogout', 'Force logout')}
                          </Button>
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

      <Dialog open={dialogOpen} onOpenChange={(open) => { if (!open) setDialogOpen(false) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{dialogState.mode === 'create' ? t('settings.addUser') : t('settings.editUser')}</DialogTitle>
          </DialogHeader>
          {dialogError && <div className="rounded-md bg-destructive/10 px-4 py-2 text-sm text-destructive">{dialogError}</div>}
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="user-username">{t('settings.username')} *</Label>
              <Input id="user-username" value={formUsername} onChange={(e) => setFormUsername(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="user-email">{t('settings.email')}</Label>
              <Input id="user-email" type="email" value={formEmail} onChange={(e) => setFormEmail(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="user-password">{t('settings.password')}{dialogState.mode === 'create' && ' *'}</Label>
              <Input id="user-password" type="password" value={formPassword} onChange={(e) => setFormPassword(e.target.value)} placeholder={dialogState.mode === 'edit' ? t('settings.leaveBlankToKeep') : ''} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="user-role">{t('settings.role')} *</Label>
              <select id="user-role" value={formRole} onChange={(e) => setFormRole(e.target.value as 'admin' | 'operator' | 'viewer')} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                <option value="admin">Admin</option><option value="operator">Operator</option><option value="viewer">Viewer</option>
              </select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button onClick={() => void handleDialogSubmit()} disabled={dialogLoading}>{dialogLoading ? t('common.saving') : t('common.save')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('settings.confirmDeleteUser')}</AlertDialogTitle>
            <AlertDialogDescription>{t('settings.confirmDeleteUserMessage')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => { setDeleteConfirmOpen(false); setUserToDelete(undefined) }}>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete} variant="destructive">{t('common.delete')}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={!!forceLogoutUser} onOpenChange={(open) => { if (!open) setForceLogoutUser(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('settings.confirmForceLogoutTitle', 'Force sign out?')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('settings.confirmForceLogoutMessage', 'This will immediately end all sessions for {name}. They will need to sign in again.', { name: forceLogoutUser?.name })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={forceLogoutLoading}>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={() => void confirmForceLogout()} disabled={forceLogoutLoading}>
              {forceLogoutLoading ? t('common.saving', 'Working...') : t('settings.forceLogout', 'Force logout')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
