import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import { PageHeader } from '@/components/layout/PageHeader'
import { getSystemConfig, validateSystemConfig, type SystemConfigDTO, type ValidateConfigResult } from '@/api/config'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const SECTIONS: { key: keyof SystemConfigDTO; label: string }[] = [
  { key: 'server', label: 'Server' },
  { key: 'database', label: 'Database' },
  { key: 'cleanup', label: 'Cleanup' },
  { key: 'log', label: 'Log' },
  { key: 'cors', label: 'CORS' },
  { key: 'admin', label: 'Admin' },
  { key: 'session', label: 'Session' },
  { key: 'jwt', label: 'JWT' },
]

export default function SystemConfigPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const isAdmin = user?.role === 'admin'

  const [config, setConfig] = useState<SystemConfigDTO | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [validation, setValidation] = useState<ValidateConfigResult | null>(null)
  const [isValidating, setIsValidating] = useState(false)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await getSystemConfig()
      setConfig(res.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (isAdmin) void load()
  }, [isAdmin, load])

  const handleValidate = async () => {
    setIsValidating(true)
    setError(null)
    try {
      const res = await validateSystemConfig()
      setValidation(res.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsValidating(false)
    }
  }

  if (!isAdmin) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('settings.systemConfig')} subtitle={t('settings.systemConfigDescription')} />
        <Card>
          <CardContent className="py-12 text-center text-sm text-muted-foreground">
            {t('settings.adminOnly')}
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('settings.systemConfig')}
        subtitle={t('settings.systemConfigDescription')}
        actions={<Button variant="outline" size="sm" onClick={() => void handleValidate()} disabled={isValidating}>{t('settings.revalidate')}</Button>}
      />

      {error && <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

      {validation && (
        <div className={`rounded-md px-4 py-3 text-sm ${validation.valid ? 'bg-healthy-bg text-healthy-text' : 'bg-destructive/10 text-destructive'}`}>
          <p className="font-medium">
            {validation.valid ? t('settings.configValid') : validation.error || 'Invalid'}
          </p>
          {validation.warnings && validation.warnings.length > 0 && (
            <ul className="mt-1 list-disc pl-5">
              {validation.warnings.map((w, i) => <li key={i}>{w}</li>)}
            </ul>
          )}
        </div>
      )}

      {isLoading ? (
        <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      ) : config ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {SECTIONS.map((section) => {
            const value = config[section.key]
            return (
              <Card key={section.key}>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-sm">
                    {section.label}
                    {value == null && <Badge variant="outline">—</Badge>}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {value ? (
                    <pre className="overflow-x-auto whitespace-pre-wrap break-all rounded bg-muted/40 p-2 text-xs">
                      {JSON.stringify(value, null, 2)}
                    </pre>
                  ) : (
                    <p className="text-xs text-muted-foreground">—</p>
                  )}
                </CardContent>
              </Card>
            )
          })}
        </div>
      ) : null}
    </div>
  )
}
