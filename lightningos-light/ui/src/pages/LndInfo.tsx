import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getBitcoinSource, getLndConfig, getLndUpgradeStatus } from '../api'

type LndConfigResponse = {
  current?: {
    alias?: string
    color?: string
    min_channel_size_sat?: number | string
    max_channel_size_sat?: number | string
  }
  raw_user_conf?: string
}

type BitcoinSourceStatus = {
  source?: 'local' | 'remote'
}

export default function LndInfo() {
  const { t } = useTranslation()
  const [config, setConfig] = useState<LndConfigResponse | null>(null)
  const [bitcoinSource, setBitcoinSource] = useState<'remote' | 'local'>('remote')
  const [upgrade, setUpgrade] = useState<any>(null)
  const [showRaw, setShowRaw] = useState(false)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')

  useEffect(() => {
    let mounted = true

    const loadData = async (initial = false) => {
      const [configRes, sourceRes, upgradeRes] = await Promise.allSettled([
        getLndConfig(),
        getBitcoinSource(),
        getLndUpgradeStatus()
      ])

      if (!mounted) return

      if (configRes.status === 'fulfilled') {
        setConfig((configRes.value || null) as LndConfigResponse | null)
        setLoadError('')
      } else {
        setLoadError(t('lndInfo.loadFailed'))
      }

      if (sourceRes.status === 'fulfilled') {
        const sourceData = sourceRes.value as BitcoinSourceStatus
        if (sourceData?.source === 'local' || sourceData?.source === 'remote') {
          setBitcoinSource(sourceData.source)
        }
      }

      if (upgradeRes.status === 'fulfilled') {
        setUpgrade(upgradeRes.value || null)
      }

      if (initial) {
        setLoading(false)
      }
    }

    void loadData(true)
    const timer = window.setInterval(() => {
      void loadData(false)
    }, 30000)

    return () => {
      mounted = false
      window.clearInterval(timer)
    }
  }, [t])

  const formatVersion = (value?: string) => {
    if (!value) return t('common.na')
    const trimmed = value.startsWith('v') ? value.slice(1) : value
    return `v${trimmed}`
  }

  const formatCheckedAt = (value?: string) => {
    if (!value) return t('common.na')
    const parsed = new Date(value)
    if (Number.isNaN(parsed.getTime())) return t('common.na')
    return parsed.toLocaleString()
  }

  const formatSatValue = (value: unknown) => {
    const numeric = Number(value)
    if (!Number.isFinite(numeric) || numeric <= 0) return t('common.na')
    return `${Math.trunc(numeric).toLocaleString()} sat`
  }

  const aliasValue = config?.current?.alias?.trim() || t('common.na')
  const colorValue = config?.current?.color?.trim() || '#ff9900'
  const minChannelValue = formatSatValue(config?.current?.min_channel_size_sat)
  const maxChannelValue = formatSatValue(config?.current?.max_channel_size_sat)
  const bitcoinSourceValue = bitcoinSource === 'local' ? t('common.local') : t('common.remote')
  const rawConfig = config?.raw_user_conf || ''

  const upgradeStatusText = useMemo(() => {
    if (upgrade?.error) {
      return t('lndUpgrade.statusError', { error: upgrade.error })
    }
    if (upgrade?.running) {
      return t('lndUpgrade.inProgress')
    }
    if (upgrade?.update_available) {
      return t('lndUpgrade.updateAvailable')
    }
    return t('lndUpgrade.upToDate')
  }, [t, upgrade])

  return (
    <section className="space-y-6">
      <div className="section-card">
        <h2 className="text-2xl font-semibold">{t('lndInfo.title')}</h2>
        <p className="text-fog/60">{t('lndInfo.subtitle')}</p>
        <p className="text-fog/50 text-sm">{t('lndInfo.readOnlyHint')}</p>
        <p className="mt-4 inline-flex rounded-full border border-cyan-400/30 bg-cyan-500/10 px-3 py-1 text-xs text-cyan-200">
          {t('lndInfo.externalGuard')}
        </p>
        {loadError && <p className="text-sm text-brass mt-4">{loadError}</p>}
      </div>

      <div className="section-card space-y-4">
        <h3 className="text-lg font-semibold">{t('lndConfig.basicSettings')}</h3>
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3">
            <span className="text-fog/70">{t('lndConfig.alias')}</span>
            <span className="text-fog/90">{aliasValue}</span>
          </div>
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3">
            <span className="text-fog/70">{t('lndConfig.nodeColor')}</span>
            <span className="inline-flex items-center gap-2">
              <span
                className="h-4 w-4 rounded-full border border-white/30"
                style={{ backgroundColor: colorValue }}
                aria-hidden="true"
              />
              <span className="font-mono text-fog/90">{colorValue}</span>
            </span>
          </div>
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3">
            <span className="text-fog/70">{t('lndConfig.minChannelSize')}</span>
            <span className="font-mono text-fog/90">{minChannelValue}</span>
          </div>
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3">
            <span className="text-fog/70">{t('lndConfig.maxChannelSize')}</span>
            <span className="font-mono text-fog/90">{maxChannelValue}</span>
          </div>
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3 sm:col-span-2">
            <span className="text-fog/70">{t('lndConfig.bitcoinSource')}</span>
            <span className="text-fog/90">{bitcoinSourceValue}</span>
          </div>
        </div>
      </div>

      <div className="section-card space-y-4">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h3 className="text-lg font-semibold">{t('lndUpgrade.title')}</h3>
            <p className="text-fog/60">{t('lndInfo.upgradeReadOnly')}</p>
          </div>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 text-sm">
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3">
            <span className="text-fog/70">{t('lndUpgrade.current')}</span>
            <span className="font-mono text-fog">{formatVersion(upgrade?.current_version)}</span>
          </div>
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3">
            <span className="text-fog/70">{t('lndUpgrade.latest')}</span>
            <span className="font-mono text-fog">{formatVersion(upgrade?.latest_version)}</span>
          </div>
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3">
            <span className="text-fog/70">{t('lndUpgrade.channel')}</span>
            <span className="text-fog/90">
              {upgrade?.latest_channel
                ? t(`lndUpgrade.channels.${upgrade.latest_channel}`)
                : t('common.na')}
            </span>
          </div>
          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-ink/40 px-4 py-3">
            <span className="text-fog/70">{t('lndUpgrade.checkedAt')}</span>
            <span className="text-fog/90">{formatCheckedAt(upgrade?.checked_at)}</span>
          </div>
        </div>

        <p className={`text-sm ${upgrade?.error ? 'text-rose-200' : 'text-fog/70'}`}>
          {upgradeStatusText}
        </p>
        <p className="text-xs text-fog/50">{t('lndInfo.upgradeDisabled')}</p>
      </div>

      <div className="section-card space-y-4">
        <div className="flex items-center justify-between gap-4">
          <h3 className="text-lg font-semibold">{t('lndInfo.lndConfTitle')}</h3>
          <button className="btn-secondary" onClick={() => setShowRaw((current) => !current)} type="button">
            {showRaw ? t('lndInfo.hideLndConf') : t('lndInfo.viewLndConf')}
          </button>
        </div>
        {showRaw && (
          <>
            <textarea className="input-field min-h-[220px] font-mono text-xs" value={rawConfig} readOnly />
            {!rawConfig && <p className="text-xs text-fog/50">{t('lndInfo.emptyLndConf')}</p>}
          </>
        )}
      </div>

      {loading && !config && <p className="text-fog/60">{t('lndConfig.loadingConfig')}</p>}
    </section>
  )
}
