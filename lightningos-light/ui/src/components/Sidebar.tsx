import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getChatInbox, getLnPeers } from '../api'
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

const lastReadKey = 'chat:lastRead'

const readLastReadMap = () => {
  try {
    const raw = localStorage.getItem(lastReadKey)
    if (!raw) return {}
    const parsed = JSON.parse(raw)
    if (parsed && typeof parsed === 'object') {
      return parsed as Record<string, number>
    }
  } catch {
    // ignore storage errors
  }
  return {}
}

export default function Sidebar({ routes, current, open, onClose }: SidebarProps) {
  const { t } = useTranslation()
  const [version, setVersion] = useState('')
  const [unreadChats, setUnreadChats] = useState(0)

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

  useEffect(() => {
    let mounted = true
    const loadUnread = async () => {
      try {
        const [inboxRes, peersRes] = await Promise.allSettled([getChatInbox(), getLnPeers()])
        if (!mounted) return
        const items = inboxRes.status === 'fulfilled' && Array.isArray(inboxRes.value?.items)
          ? inboxRes.value.items
          : []
        const peers = peersRes.status === 'fulfilled' && Array.isArray(peersRes.value?.peers)
          ? peersRes.value.peers
          : []
        const onlineSet = peersRes.status === 'fulfilled'
          ? new Set(peers.map((peer: any) => peer?.pub_key).filter(Boolean))
          : null
        const lastReadMap = readLastReadMap()
        const unread = new Set<string>()
        for (const item of items) {
          const peerKey = typeof item?.peer_pubkey === 'string' ? item.peer_pubkey : ''
          if (!peerKey) continue
          if (onlineSet && !onlineSet.has(peerKey)) {
            continue
          }
          const ts = new Date(item.last_inbound_at).getTime()
          if (!ts || Number.isNaN(ts)) continue
          const lastRead = lastReadMap[peerKey] || 0
          if (ts > lastRead) {
            unread.add(peerKey)
          }
        }
        setUnreadChats(unread.size)
      } catch {
        if (!mounted) return
        setUnreadChats(0)
      }
    }

    loadUnread()
    const timer = window.setInterval(loadUnread, 12000)
    return () => {
      mounted = false
      window.clearInterval(timer)
    }
  }, [])

  const unreadLabel = unreadChats === 1
    ? t('chat.unreadSingle')
    : t('chat.unreadMultiple', { count: unreadChats })

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
            {route.key === 'chat' ? (
              <span className="inline-flex items-center gap-2">
                <span>{route.label}</span>
                {unreadChats > 0 && (
                  <span className="rounded-full bg-ember px-2 py-0.5 text-[10px] font-semibold text-white" title={unreadLabel}>
                    {unreadChats}
                  </span>
                )}
              </span>
            ) : (
              route.label
            )}
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
