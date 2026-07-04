import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import { PageHeader } from '@/components/layout/PageHeader'
import {
  getApiKeys,
  createApiKey,
  rotateApiKey,
  revokeApiKey,
  type ApiKeyDTO,
  type CreateApiKeyRequest,
} from '@/api/apiKeys'
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

export default function ApiKeysPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const isAdmin = user?.role === 'admin'

  const [keys, setKeys] = useState<ApiKeyDTO[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Create dialog state
  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)

  // Newly created / rotated key reveal (shown once)
  const [revealedKey, setRevealedKey] = useState<{ fullKey: string; kind: 'created' | 'rotated' } | null>(null)
  const [copied, setCopied] = useState(false)

  // Revoke state
  const [revokeTarget, setRevokeTarget] = useState<{ id: number; name: string } | null>(null)
  const [revoking, setRevoking] = useState(false)

  const loadKeys = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await getApiKeys()
      setKeys(res.data?.keys || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (isAdmin) void loadKeys()
  }, [isAdmin, loadKeys])

  const handleCreate = async () => {
    setCreating(true)
    setCreateError(null)
    try {
      const req: CreateApiKeyRequest = { name: name.trim() }
      const res = await createApiKey(req)
      setRevealedKey({ fullKey: res.data.full_key, kind: 'created' })
      setCreateOpen(false)
      setName('')
      await loadKeys()
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : String(err))
    } finally {
      setCreating(false)
    }
  }

  const handleRotate = async (id: number) => {
    try {
      const res = await rotateApiKey(id)
      setRevealedKey({ fullKey: res.data.full_key, kind: 'rotated' })
      await loadKeys()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const confirmRevoke = async () => {
    if (!revokeTarget) return
    setRevoking(true)
    try {
      await revokeApiKey(revokeTarget.id)
      await loadKeys()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setRevoking(false)
      setRevokeTarget(null)
    }
  }

  const copyKey = async () => {
    if (!revealedKey) return
    try {
      await navigator.clipboard.writeText(revealedKey.fullKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Clipboard may be unavailable; the user can still select the text manually.
    }
  }

  if (!isAdmin) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('settings.apiKeys', 'API Keys')} subtitle={t('settings.apiKeysDescription', 'Manage API keys for Beacon authentication.')} />
        <Card>
          <CardContent className="py-12 text-center text-sm text-muted-foreground">
            {t('settings.adminOnly', 'Access denied — administrator role required.')}
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('settings.apiKeys', 'API Keys')}
        subtitle={t('settings.apiKeysDescription', 'Manage API keys for Beacon authentication.')}
        actions={<Button onClick={() => { setCreateError(null); setCreateOpen(true) }}>{t('settings.createApiKey', 'Create API Key')}</Button>}
      />

      {error && <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
          ) : keys.length === 0 ? (
            <div className="py-12 text-center text-sm text-muted-foreground">{t('settings.noApiKeys', 'No API keys yet. Create one to authenticate a Beacon.')}</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-border">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.keyName', 'Name')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.keyPrefix', 'Prefix')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('common.status')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.lastUsed', 'Last used')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.created', 'Created')}</th>
                    <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.actions')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {keys.map((k) => (
                    <tr key={k.id} className="hover:bg-muted/50">
                      <td className="whitespace-nowrap px-6 py-4 text-sm font-medium">{k.name}</td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm font-mono text-muted-foreground">{k.key_prefix}…</td>
                      <td className="whitespace-nowrap px-6 py-4">
                        <Badge variant={k.is_active ? 'default' : 'destructive'}>
                          {k.is_active ? t('settings.active') : t('settings.revoked')}
                        </Badge>
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm text-muted-foreground">
                        {k.last_used_at ? new Date(k.last_used_at).toLocaleString() : '—'}
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm text-muted-foreground">
                        {k.created_at ? new Date(k.created_at).toLocaleString() : '—'}
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-right text-sm space-x-2">
                        <Button variant="link" size="sm" onClick={() => void handleRotate(k.id)} disabled={!k.is_active}>
                          {t('settings.rotate', 'Rotate')}
                        </Button>
                        <Button variant="link" size="sm" className="text-destructive" onClick={() => setRevokeTarget({ id: k.id, name: k.name })} disabled={!k.is_active}>
                          {t('settings.revoke', 'Revoke')}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={(open) => { if (!open) setCreateOpen(false) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('settings.createApiKey', 'Create API Key')}</DialogTitle>
          </DialogHeader>
          {createError && <div className="rounded-md bg-destructive/10 px-4 py-2 text-sm text-destructive">{createError}</div>}
          <div className="space-y-2">
            <Label htmlFor="api-key-name">{t('settings.keyName', 'Name')} *</Label>
            <Input id="api-key-name" value={name} onChange={(e) => setName(e.target.value)} placeholder={t('settings.keyNamePlaceholder', 'e.g. beacon-prod-01')} />
            <p className="text-xs text-muted-foreground">{t('settings.apiKeyHelp', 'The full key is shown only once after creation. Store it securely.')}</p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)} disabled={creating}>{t('common.cancel')}</Button>
            <Button onClick={() => void handleCreate()} disabled={creating || !name.trim()}>{creating ? t('common.saving') : t('common.create')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Reveal-once dialog (created or rotated) */}
      <Dialog open={!!revealedKey} onOpenChange={(open) => { if (!open) setRevealedKey(null) }}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {revealedKey?.kind === 'rotated' ? t('settings.rotatedKeyTitle', 'API key rotated') : t('settings.createdKeyTitle', 'API key created')}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-2">
            <p className="text-sm text-destructive">
              {t('settings.apiKeyOnceWarning', 'Copy this key now. For security, it will not be shown again.')}
            </p>
            <div className="flex items-center gap-2">
              <Input readOnly value={revealedKey?.fullKey || ''} className="font-mono text-xs" onFocus={(e) => e.target.select()} />
              <Button size="sm" onClick={() => void copyKey()}>{copied ? t('common.copied', 'Copied') : t('common.copy', 'Copy')}</Button>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setRevealedKey(null)}>{t('common.done', 'Done')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Revoke confirmation */}
      <AlertDialog open={!!revokeTarget} onOpenChange={(open) => { if (!open) setRevokeTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('settings.confirmRevokeKeyTitle', 'Revoke API key?')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('settings.confirmRevokeKeyMessage', 'Revoke "{name}"? Beacons using this key will immediately fail to authenticate.', { name: revokeTarget?.name })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={revoking}>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={() => void confirmRevoke()} disabled={revoking}>
              {revoking ? t('common.saving') : t('settings.revoke', 'Revoke')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
