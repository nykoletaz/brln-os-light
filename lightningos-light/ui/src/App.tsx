import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import Sidebar from './components/Sidebar'
import Topbar from './components/Topbar'
import Dashboard from './pages/Dashboard'
import Reports from './pages/Reports'
import Wizard from './pages/Wizard'
import Wallet from './pages/Wallet'
import LightningOps from './pages/LightningOps'
import Chat from './pages/Chat'
import Disks from './pages/Disks'
import Logs from './pages/Logs'
import BitcoinRemote from './pages/BitcoinRemote'
import BitcoinLocal from './pages/BitcoinLocal'
import Elements from './pages/Elements'
import Notifications from './pages/Notifications'
import LndConfig from './pages/LndConfig'
import AppStore from './pages/AppStore'
import Terminal from './pages/Terminal'
import { getLndStatus, getWizardStatus } from './api'
import { defaultPalette, paletteOrder, resolvePalette, resolveTheme, type PaletteKey, type ThemeMode } from './theme'

function useHashRoute() {
  const [hash, setHash] = useState(window.location.hash.replace('#', ''))

  useEffect(() => {
    const handler = () => setHash(window.location.hash.replace('#', ''))
    window.addEventListener('hashchange', handler)
    return () => window.removeEventListener('hashchange', handler)
  }, [])

  return hash
}

type RouteItem = {
  key: string
  label: string
  element: JSX.Element
}

export default function App() {
  const { t, i18n } = useTranslation()
  const route = useHashRoute()
  const [theme, setTheme] = useState<ThemeMode>(() => resolveTheme(window.localStorage.getItem('los-theme')))
  const [palette, setPalette] = useState<PaletteKey>(() => resolvePalette(window.localStorage.getItem('los-palette')))
  const [walletUnlocked, setWalletUnlocked] = useState<boolean | null>(null)
  const [walletExists, setWalletExists] = useState<boolean | null>(null)
  const [menuOpen, setMenuOpen] = useState(false)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    window.localStorage.setItem('los-theme', theme)
  }, [theme])

  useEffect(() => {
    document.documentElement.setAttribute('data-palette', palette)
    window.localStorage.setItem('los-palette', palette)
  }, [palette])

  useEffect(() => {
    let active = true
    const load = async () => {
      try {
        const data: any = await getWizardStatus()
        if (!active) return
        setWalletExists(Boolean(data?.wallet_exists))
      } catch {
        if (!active) return
      }
      try {
        const status: any = await getLndStatus()
        if (!active) return
        if (typeof status?.wallet_state === 'string') {
          setWalletUnlocked(status.wallet_state === 'unlocked')
        }
      } catch {
        if (!active) return
      }
    }
    load()
    const timer = window.setInterval(load, 30000)
    return () => {
      active = false
      window.clearInterval(timer)
    }
  }, [])

  const wizardHidden = walletUnlocked === true
  const wizardRequired = walletExists === false && !wizardHidden

  const routes = useMemo(() => {
    const list: RouteItem[] = []
    if (!wizardHidden) {
      list.push({ key: 'wizard', label: t('nav.wizard'), element: <Wizard /> })
    }
    list.push(
      { key: 'dashboard', label: t('nav.dashboard'), element: <Dashboard /> },
      { key: 'reports', label: t('nav.reports'), element: <Reports /> },
      { key: 'wallet', label: t('nav.wallet'), element: <Wallet /> },
      { key: 'lightning-ops', label: t('nav.lightningOps'), element: <LightningOps /> },
      { key: 'chat', label: t('nav.chat'), element: <Chat /> },
      { key: 'lnd', label: t('nav.lndConfig'), element: <LndConfig /> },
      { key: 'apps', label: t('nav.apps'), element: <AppStore /> },
      { key: 'bitcoin', label: t('nav.bitcoinRemote'), element: <BitcoinRemote /> },
      { key: 'bitcoin-local', label: t('nav.bitcoinLocal'), element: <BitcoinLocal /> },
      { key: 'elements', label: t('nav.elements'), element: <Elements /> },
      { key: 'notifications', label: t('nav.notifications'), element: <Notifications /> },
      { key: 'disks', label: t('nav.disks'), element: <Disks /> },
      { key: 'terminal', label: t('nav.terminal'), element: <Terminal /> },
      { key: 'logs', label: t('nav.logs'), element: <Logs /> }
    )
    return list
  }, [i18n.language, t, wizardHidden])

  useEffect(() => {
    setMenuOpen(false)
  }, [route])

  useEffect(() => {
    document.body.style.overflow = menuOpen ? 'hidden' : ''
    if (!menuOpen) {
      return
    }
    const handleKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setMenuOpen(false)
      }
    }
    window.addEventListener('keydown', handleKey)
    return () => {
      window.removeEventListener('keydown', handleKey)
      document.body.style.overflow = ''
    }
  }, [menuOpen])

  const current = useMemo(() => {
    const matched = routes.find((item) => item.key === route)
    if (wizardRequired) {
      return routes.find((item) => item.key === 'wizard') || matched || routes[0]
    }
    if (matched) {
      return matched
    }
    return routes.find((item) => item.key === 'dashboard') || routes[0]
  }, [route, routes, wizardRequired])

  const handlePaletteToggle = () => {
    setPalette((current) => {
      const index = paletteOrder.indexOf(current)
      if (index === -1) {
        return defaultPalette
      }
      return paletteOrder[(index + 1) % paletteOrder.length]
    })
  }

  return (
    <>
      <div
        className={`fixed inset-0 z-30 bg-black/60 backdrop-blur-sm transition-opacity lg:hidden ${
          menuOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'
        }`}
        onClick={() => setMenuOpen(false)}
        aria-hidden="true"
      />
      <div className="min-h-screen flex flex-col lg:flex-row text-fog">
        <Sidebar routes={routes} current={current.key} open={menuOpen} onClose={() => setMenuOpen(false)} />
        <div className="flex-1 flex flex-col">
          <Topbar
            onMenuToggle={() => setMenuOpen((prev) => !prev)}
            menuOpen={menuOpen}
            theme={theme}
            palette={palette}
            onThemeToggle={() => setTheme((prev) => (prev === 'dark' ? 'light' : 'dark'))}
            onPaletteToggle={handlePaletteToggle}
          />
          <main className="px-6 pb-16 pt-6 lg:px-12">
            {current.element}
          </main>
        </div>
      </div>
    </>
  )
}
