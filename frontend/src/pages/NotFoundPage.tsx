/**
 * Not Found Page (404)
 *
 * Displayed when a route is not found.
 */

import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
export default function NotFoundPage() {
  const { t } = useTranslation()

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg-page)]">
      <div className="text-center px-4">
        <div className="mb-8">
          <span className="text-9xl font-bold text-blue-500">404</span>
        </div>
        <h1 className="text-3xl font-bold mb-4 text-[var(--color-text-primary)]">
          {t('errors.pageNotFound')}
        </h1>
        <p className="text-lg mb-8 text-[var(--color-text-secondary)]">
          {t('errors.pageNotFoundDescription')}
        </p>
        <Link
          to="/dashboard"
          className="inline-flex items-center px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
        >
          <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          {t('errors.backToDashboard')}
        </Link>
      </div>
    </div>
  )
}
