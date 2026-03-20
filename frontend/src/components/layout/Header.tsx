/**
 * Header Component
 *
 * Top header bar for authenticated pages.
 * Contains: Hamburger (mobile), Logo, TimezoneSelector, LanguageSwitcher, ThemeToggle, User menu
 */

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link, useNavigate } from 'react-router-dom'
import { ThemeToggle } from '../common/ThemeToggle'
import { LanguageSwitcher } from '../common/LanguageSwitcher'
import { TimezoneSelector } from '../common/TimezoneSelector'
import { useAuthStore } from '../../stores/authStore'

export interface HeaderProps {
  /** Callback to toggle sidebar on mobile */
  onMenuToggle: () => void
}

// Icon components
function MenuIcon({ className = '' }: { className?: string }) {
  return (
    <svg className={`h-6 w-6 ${className}`} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
    </svg>
  )
}

function UserIcon({ className = '' }: { className?: string }) {
  return (
    <svg className={`h-5 w-5 ${className}`} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
    </svg>
  )
}

function CogIcon({ className = '' }: { className?: string }) {
  return (
    <svg className={`h-5 w-5 ${className}`} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  )
}

function LogoutIcon({ className = '' }: { className?: string }) {
  return (
    <svg className={`h-5 w-5 ${className}`} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0l3-3m0 0l-3-3m3 3H9" />
    </svg>
  )
}

function ChevronDownIcon({ className = '' }: { className?: string }) {
  return (
    <svg className={`h-4 w-4 ${className}`} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
    </svg>
  )
}

/**
 * Header Component
 */
export function Header({ onMenuToggle }: HeaderProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const user = useAuthStore((state) => state.user)
  const logout = useAuthStore((state) => state.logout)
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const [isTimezoneOpen, setIsTimezoneOpen] = useState(false)

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between shadow-xs bg-[var(--color-bg-surface)] px-4">
      {/* Left section: Menu button + Logo */}
      <div className="flex items-center gap-4">
        {/* Mobile menu button */}
        <button
          type="button"
          onClick={onMenuToggle}
          className="rounded-lg p-2 text-[var(--color-text-muted)] hover:bg-[var(--color-hover-overlay)] hover:text-[var(--color-text-secondary)] md:hidden"
          aria-label={t('nav.toggleMenu')}
        >
          <MenuIcon />
        </button>

        {/* Logo (visible on mobile when sidebar is hidden) */}
        <Link to="/dashboard" className="flex items-center gap-2 md:hidden">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-[var(--color-brand)]">
            <svg className="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
            </svg>
          </div>
          <span className="text-lg font-bold text-[var(--color-text-primary)]">NodePulse</span>
        </Link>
      </div>

      {/* Right section: Controls */}
      <div className="flex items-center gap-2">
        {/* Timezone Selector (hidden on small screens) */}
        <div className="hidden lg:block">
          <button
            type="button"
            onClick={() => setIsTimezoneOpen(!isTimezoneOpen)}
            className="rounded-lg p-2 text-[var(--color-text-muted)] hover:bg-[var(--color-hover-overlay)] hover:text-[var(--color-text-secondary)]"
            title={t('settings.timezone')}
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
          {isTimezoneOpen && (
            <div className="absolute right-24 top-14 z-50 w-72 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-elevated)] p-4 shadow-lg">
              <TimezoneSelector size="sm" />
            </div>
          )}
        </div>

        {/* Language Switcher */}
        <LanguageSwitcher variant="dropdown" size="sm" className="w-24" />

        {/* Theme Toggle */}
        <ThemeToggle size="md" />

        {/* User Menu */}
        <div className="relative ml-2">
          <button
            type="button"
            data-testid="user-menu-button"
            onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
            className="flex items-center gap-2 rounded-lg p-2 text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]"
          >
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--color-brand-muted)]">
              <UserIcon className="text-[var(--color-brand)]" />
            </div>
            <span className="hidden font-medium sm:block">{user?.username}</span>
            <ChevronDownIcon className={`transition-transform ${isUserMenuOpen ? 'rotate-180' : ''}`} />
          </button>

          {/* Dropdown menu */}
          {isUserMenuOpen && (
            <div className="absolute right-0 top-full z-50 mt-1 w-48 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-elevated)] py-1 shadow-lg">
              <div className="border-b border-[var(--color-border)] px-4 py-2">
                <p className="text-sm font-medium text-[var(--color-text-primary)]">{user?.username}</p>
                <p className="text-xs text-[var(--color-text-muted)]">{user?.role}</p>
              </div>
              <Link
                to="/settings/sessions"
                onClick={() => setIsUserMenuOpen(false)}
                className="flex items-center gap-2 px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]"
              >
                <UserIcon />
                {t('nav.profile')}
              </Link>
              <Link
                to="/settings/preferences"
                onClick={() => setIsUserMenuOpen(false)}
                className="flex items-center gap-2 px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]"
              >
                <CogIcon />
                {t('nav.settings')}
              </Link>
              <hr className="my-1 border-[var(--color-border)]" />
              <button
                type="button"
                onClick={handleLogout}
                className="flex w-full items-center gap-2 px-4 py-2 text-sm text-[var(--color-critical)] hover:bg-[var(--color-hover-overlay)]"
              >
                <LogoutIcon />
                {t('nav.logout')}
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}

export default Header
