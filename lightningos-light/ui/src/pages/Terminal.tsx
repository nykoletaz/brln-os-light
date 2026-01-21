import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getTerminalStatus } from '../api'

type TerminalStatus = {
  enabled: boolean
  credential?: string
  allow_write?: boolean
  port?: number
  operator_user?: string
  operator_password?: string
}

export default function Terminal() {
  const { t } = useTranslation()
  const [status, setStatus] = useState<TerminalStatus | null>(null)
  const [statusMessage, setStatusMessage] = useState('')

  const copyToClipboard = async (value: string) => {
    if (!value) return
    try {
      await navigator.clipboard.writeText(value)
    } catch {
      // ignore copy failures
    }
  }

  useEffect(() => {
    let mounted = true
    getTerminalStatus()
      .then((data: TerminalStatus) => {
        if (!mounted) return
        setStatus(data)
        setStatusMessage('')
      })
      .catch((err: any) => {
        if (!mounted) return
        setStatus(null)
        setStatusMessage(err?.message || t('terminal.statusUnavailable'))
      })
    return () => {
      mounted = false
    }
  }, [])

  return (
    <div className="space-y-6">
      <div className="rounded-3xl border border-white/10 bg-ink/60 p-6 shadow-panel">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p className="text-sm uppercase tracking-[0.2em] text-fog/60">{t('terminal.kicker')}</p>
            <h2 className="text-2xl font-semibold text-white">{t('terminal.title')}</h2>
            {statusMessage && (
              <p className="mt-2 text-sm text-brass">{statusMessage}</p>
            )}
            {status && (
              <div className="mt-4 space-y-2 text-sm text-fog/70">
                <div className="flex items-center gap-2">
                  <span className="text-fog/50">{t('common.status')}</span>
                  <span>{status.enabled ? t('common.enabled') : t('common.disabled')}</span>
                </div>
                {!status.enabled && (
                  <p className="text-brass">
                    {t('terminal.disabledMessage')}
                  </p>
                )}
                {status?.operator_password && (
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-fog/50">{t('terminal.operator')}</span>
                    <span className="font-mono text-fog/80">{status.operator_user || 'losop'}</span>
                    <span className="text-fog/40">/</span>
                    <span className="font-mono text-fog/80">{status.operator_password}</span>
                    <button
                      className="text-fog/50 hover:text-fog"
                      onClick={() => copyToClipboard(status.operator_password || '')}
                      title={t('terminal.copyOperatorPassword')}
                      aria-label={t('terminal.copyOperatorPassword')}
                    >
                      <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.6">
                        <rect x="9" y="9" width="11" height="11" rx="2" />
                        <rect x="4" y="4" width="11" height="11" rx="2" />
                      </svg>
                    </button>
                  </div>
                )}
                <p className="text-xs text-fog/50">{t('terminal.pasteHint')}</p>
              </div>
            )}
          </div>
          <a className="btn-secondary" href="/terminal/" target="_blank" rel="noreferrer">
            {t('terminal.openNewTab')}
          </a>
        </div>
      </div>

      <div className="rounded-3xl border border-white/10 bg-ink/70 shadow-panel overflow-hidden">
        <iframe
          title={t('terminal.title')}
          src="/terminal/"
          className="w-full h-[70vh] bg-black"
        />
      </div>
    </div>
  )
}
