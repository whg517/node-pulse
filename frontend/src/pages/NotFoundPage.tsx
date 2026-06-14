import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { ArrowLeft } from '@phosphor-icons/react'

export default function NotFoundPage() {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="text-center px-4">
        <span className="text-8xl font-bold text-primary">404</span>
        <h1 className="mt-4 text-3xl font-bold">{t('errors.pageNotFound')}</h1>
        <p className="mt-2 text-lg text-muted-foreground">{t('errors.pageNotFoundDescription')}</p>
        <Button asChild className="mt-8">
          <Link to="/dashboard">
            <ArrowLeft className="mr-2 size-4" />
            {t('errors.backToDashboard')}
          </Link>
        </Button>
      </div>
    </div>
  )
}
