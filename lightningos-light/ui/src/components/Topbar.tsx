import { useEffect, useState } from 'react'
import { getHealth, getLndConfig, getLndStatus } from '../api'

const statusColors: Record<string, string> = {
  OK: 'bg-glow/20 text-glow border-glow/40',
  WARN: 'bg-brass/20 text-brass border-brass/40',
  ERR: 'bg-ember/20 text-ember border-ember/40'
}

type TopbarProps = {
  onMenuToggle?: () => void
  menuOpen?: boolean
}

export default function Topbar({ onMenuToggle, menuOpen }: TopbarProps) {
  const [status, setStatus] = useState('...')
  const [issues, setIssues] = useState<Array<{ component?: string; level?: string; message?: string }>>([])
  const [nodeAlias, setNodeAlias] = useState('')
  const [nodePubkey, setNodePubkey] = useState('')

  useEffect(() => {
    let mounted = true
    const load = async () => {
      try {
        const data = await getHealth()
        if (!mounted) return
        setStatus(data.status)
        setIssues(Array.isArray(data.issues) ? data.issues : [])
      } catch {
        if (!mounted) return
        setStatus('ERR')
        setIssues([{ component: 'system', level: 'ERR', message: 'Health check failed' }])
      }
    }

    load()
    const timer = setInterval(load, 30000)
    return () => {
      mounted = false
      clearInterval(timer)
    }
  }, [])

  useEffect(() => {
    let mounted = true
    const load = async () => {
      const [statusRes, configRes] = await Promise.allSettled([getLndStatus(), getLndConfig()])
      if (!mounted) return
      if (statusRes.status === 'fulfilled') {
        const pubkey = typeof statusRes.value?.pubkey === 'string' ? statusRes.value.pubkey.trim() : ''
        setNodePubkey(pubkey)
      }
      if (configRes.status === 'fulfilled') {
        const alias = typeof configRes.value?.current?.alias === 'string' ? configRes.value.current.alias.trim() : ''
        setNodeAlias(alias)
      }
    }

    load()
    const timer = setInterval(load, 30000)
    return () => {
      mounted = false
      clearInterval(timer)
    }
  }, [])

  const resolvedNodeLabel = nodeAlias || nodePubkey
  const compactPubkey = nodePubkey.length > 20
    ? `${nodePubkey.slice(0, 12)}...${nodePubkey.slice(-6)}`
    : nodePubkey
  const displayNodeLabel = nodeAlias || compactPubkey

  return (
    <header className="px-6 lg:px-12 pt-8">
      {onMenuToggle && (
        <div className="mb-6 flex items-center justify-between lg:hidden">
          <button
            type="button"
            className="inline-flex items-center gap-2 rounded-full border border-white/15 bg-ink/60 px-3 py-2 text-xs uppercase tracking-wide text-fog/70 hover:text-white hover:border-white/40 transition"
            onClick={onMenuToggle}
            aria-label={menuOpen ? 'Close menu' : 'Open menu'}
            aria-expanded={menuOpen ? true : false}
            aria-controls="app-sidebar"
          >
            {menuOpen ? (
              <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.8">
                <path d="M6 6l12 12M18 6l-12 12" />
              </svg>
            ) : (
              <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.8">
                <path d="M4 7h16M4 12h16M4 17h10" />
              </svg>
            )}
            <span>{menuOpen ? 'Close' : 'Menu'}</span>
          </button>
          <div className="text-right text-xs text-fog/60">
            <p className="text-fog font-semibold">LightningOS Light</p>
            <p>Mainnet only</p>
          </div>
        </div>
      )}
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
        <div>
          <p className="text-sm uppercase tracking-[0.3em] text-fog/50">Status overview</p>
          <h1 className="text-3xl lg:text-4xl font-semibold">LightningOS Control Center</h1>
          {displayNodeLabel && (
            <p className="mt-2 text-sm text-fog/60" title={resolvedNodeLabel}>
              Node: {displayNodeLabel}
            </p>
          )}
        </div>
        <div className="flex items-center gap-4">
          <div className={`px-4 py-2 rounded-full border text-sm ${statusColors[status] || 'bg-white/10 border-white/20'}`}>
            {status}
          </div>
          <div className="text-xs text-fog/60 max-w-xs">
            {issues.length
              ? issues
                .map((issue) => {
                  const label = issue.component ? issue.component.toUpperCase() : 'SYSTEM'
                  const message = issue.message || 'Issue detected'
                  return `${label}: ${message}`
                })
                .join(' • ')
              : status === '...'
                ? 'Checking system status...'
                : status === 'OK'
                  ? 'All systems green'
                  : 'Status unavailable'}
          </div>
        </div>
      </div>
      <div className="glow-divider mt-6" />
    </header>
  )
}
