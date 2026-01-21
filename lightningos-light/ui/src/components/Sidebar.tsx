import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import clsx from '../utils/clsx'

type RouteItem = {
  key: string
  label: string
  element: JSX.Element
}

type SidebarProps = {
  routes: RouteItem[]
  current: string
  open: boolean
  onClose: () => void
}

export default function Sidebar({ routes, current, open, onClose }: SidebarProps) {
  const { t } = useTranslation()
  const [version, setVersion] = useState('')

  useEffect(() => {
    let active = true
    fetch('/version.txt', { cache: 'no-store' })
      .then((res) => (res.ok ? res.text() : ''))
      .then((text) => {
        if (!active) return
        setVersion(text.trim())
      })
      .catch(() => {
        if (!active) return
        setVersion('')
      })
    return () => {
      active = false
    }
  }, [])

  return (
    <aside
      id="app-sidebar"
      className={clsx(
        'fixed inset-y-0 left-0 z-40 w-72 max-w-[85vw] h-full bg-ink/80 border-b lg:border-b-0 lg:border-r border-white/10 px-6 py-8 flex flex-col overflow-y-auto transition-transform duration-300 ease-out',
        open ? 'translate-x-0' : '-translate-x-full',
        'lg:translate-x-0 lg:sticky lg:top-0 lg:h-screen lg:w-72'
      )}
    >
      <div className="flex items-center justify-between lg:justify-start gap-3">
        <div className="h-10 w-10 rounded-2xl bg-glow/20 border border-glow/30 grid place-items-center text-glow font-semibold">
          Lo
        </div>
        <div className="flex-1">
          <p className="text-lg font-semibold">{t('topbar.productName')}</p>
          <p className="text-xs text-fog/60">{t('topbar.mainnetOnly')}</p>
        </div>
        <button
          type="button"
          className="lg:hidden inline-flex items-center justify-center rounded-full border border-white/15 bg-ink/60 h-9 w-9 text-fog/70 hover:text-white hover:border-white/40 transition"
          onClick={onClose}
          aria-label={t('sidebar.closeMenu')}
        >
          <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.8">
            <path d="M6 6l12 12M18 6l-12 12" />
          </svg>
        </button>
      </div>
      <nav className="mt-8 space-y-2 flex-1">
        {routes.map((route) => (
          <a
            key={route.key}
            href={`#${route.key}`}
            className={clsx(
              'block px-4 py-3 rounded-2xl text-sm transition',
              current === route.key
                ? 'bg-white/10 text-white shadow-panel'
                : 'text-fog/70 hover:text-white hover:bg-white/5'
            )}
            onClick={onClose}
          >
            {route.label}
          </a>
        ))}
      </nav>
      <div className="mt-6 border-t border-white/10 pt-4 text-xs text-fog/60">
        <p>
          {t('sidebar.versionLabel')}{' '}
          <span className="text-fog">{version || t('common.unavailable')}</span>
        </p>
        <a
          className="mt-2 inline-flex text-fog/70 hover:text-white transition"
          href="https://br-ln.com"
          target="_blank"
          rel="noreferrer"
        >
          {t('sidebar.poweredBy')}
        </a>
      </div>
    </aside>
  )
}
