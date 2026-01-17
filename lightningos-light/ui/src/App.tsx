import { useEffect, useMemo, useState } from 'react'
import Sidebar from './components/Sidebar'
import Topbar from './components/Topbar'
import Dashboard from './pages/Dashboard'
import Reports from './pages/Reports'
import Wizard from './pages/Wizard'
import Wallet from './pages/Wallet'
import LightningOps from './pages/LightningOps'
import Disks from './pages/Disks'
import Logs from './pages/Logs'
import BitcoinRemote from './pages/BitcoinRemote'
import BitcoinLocal from './pages/BitcoinLocal'
import Notifications from './pages/Notifications'
import LndConfig from './pages/LndConfig'
import AppStore from './pages/AppStore'
import Terminal from './pages/Terminal'
import { getWizardStatus } from './api'

const routes = [
  { key: 'dashboard', label: 'Dashboard', element: <Dashboard /> },
  { key: 'reports', label: 'Reports', element: <Reports /> },
  { key: 'wizard', label: 'Wizard', element: <Wizard /> },
  { key: 'wallet', label: 'Wallet', element: <Wallet /> },
  { key: 'lightning-ops', label: 'Lightning Ops', element: <LightningOps /> },
  { key: 'bitcoin', label: 'Bitcoin Remote', element: <BitcoinRemote /> },
  { key: 'lnd', label: 'LND Config', element: <LndConfig /> },
  { key: 'disks', label: 'Disks', element: <Disks /> },
  { key: 'logs', label: 'Logs', element: <Logs /> },
  { key: 'apps', label: 'Apps', element: <AppStore /> },
  { key: 'bitcoin-local', label: 'Bitcoin Local', element: <BitcoinLocal /> },
  { key: 'notifications', label: 'Notifications', element: <Notifications /> },
  { key: 'terminal', label: 'Terminal', element: <Terminal /> }
]

function useHashRoute() {
  const [hash, setHash] = useState(window.location.hash.replace('#', ''))

  useEffect(() => {
    const handler = () => setHash(window.location.hash.replace('#', ''))
    window.addEventListener('hashchange', handler)
    return () => window.removeEventListener('hashchange', handler)
  }, [])

  return hash
}

export default function App() {
  const route = useHashRoute()
  const [wizardRequired, setWizardRequired] = useState(true)
  const [menuOpen, setMenuOpen] = useState(false)

  useEffect(() => {
    let active = true
    getWizardStatus()
      .then((data: any) => {
        if (!active) return
        setWizardRequired(!data?.wallet_exists)
      })
      .catch(() => {
        if (!active) return
        setWizardRequired(true)
      })
    return () => {
      active = false
    }
  }, [])

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
  }, [route, wizardRequired])

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
          <Topbar onMenuToggle={() => setMenuOpen((prev) => !prev)} menuOpen={menuOpen} />
          <main className="px-6 pb-16 pt-6 lg:px-12">
            {current.element}
          </main>
        </div>
      </div>
    </>
  )
}
