import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getBitcoinSource, getLndConfig, setBitcoinSource, updateLndConfig, updateLndRawConfig } from '../api'

export default function LndConfig() {
  const { t } = useTranslation()
  const [config, setConfig] = useState<any>(null)
  const [alias, setAlias] = useState('')
  const [color, setColor] = useState('#ff9900')
  const [colorInput, setColorInput] = useState('#ff9900')
  const [minChan, setMinChan] = useState('')
  const [maxChan, setMaxChan] = useState('')
  const [raw, setRaw] = useState('')
  const [advanced, setAdvanced] = useState(false)
  const [status, setStatus] = useState('')
  const [bitcoinSource, setBitcoinSourceState] = useState<'remote' | 'local'>('remote')
  const [sourceBusy, setSourceBusy] = useState(false)

  useEffect(() => {
    getLndConfig().then((data: any) => {
      setConfig(data)
      setAlias(data.current.alias || '')
      const nextColor = data.current.color || '#ff9900'
      setColor(nextColor)
      setColorInput(nextColor)
      const minVal = Number(data.current.min_channel_size_sat || 0)
      const maxVal = Number(data.current.max_channel_size_sat || 0)
      setMinChan(minVal > 0 ? minVal.toString() : '')
      setMaxChan(maxVal > 0 ? maxVal.toString() : '')
      setRaw(data.raw_user_conf || '')
    }).catch(() => null)
    getBitcoinSource().then((data: any) => {
      if (data?.source === 'local' || data?.source === 'remote') {
        setBitcoinSourceState(data.source)
      }
    }).catch(() => null)
  }, [])

  const isHexColor = (value: string) => /^#[0-9a-fA-F]{6}$/.test(value.trim())

  const handleSave = async () => {
    if (!isHexColor(color)) {
      setStatus(t('lndConfig.colorInvalid'))
      return
    }
    setStatus(t('common.saving'))
    try {
      await updateLndConfig({
        alias,
        color,
        min_channel_size_sat: Number(minChan || 0),
        max_channel_size_sat: Number(maxChan || 0),
        apply_now: true
      })
      setStatus(t('lndConfig.savedApplied'))
    } catch {
      setStatus(t('lndConfig.saveFailed'))
    }
  }

  const handleSaveRaw = async () => {
    setStatus(t('lndConfig.savingAdvanced'))
    try {
      const result = await updateLndRawConfig({ raw_user_conf: raw, apply_now: true })
      if (result?.warning) {
        setStatus(t('lndConfig.advancedAppliedWarning', { warning: result.warning }))
      } else {
        setStatus(t('lndConfig.advancedApplied'))
      }
    } catch (err) {
      if (err instanceof Error && err.message) {
        setStatus(err.message)
      } else {
        setStatus(t('lndConfig.advancedFailed'))
      }
    }
  }

  const handleToggleSource = async () => {
    if (sourceBusy) return
    const next = bitcoinSource === 'remote' ? 'local' : 'remote'
    const targetLabel = next === 'local' ? t('common.local') : t('common.remote')
    setSourceBusy(true)
    setStatus(t('lndConfig.switchingBitcoin', { target: targetLabel }))
    try {
      await setBitcoinSource({ source: next })
      setBitcoinSourceState(next)
      setStatus(t('lndConfig.bitcoinSourceSet', { target: targetLabel }))
    } catch (err) {
      setStatus(err instanceof Error ? err.message : t('lndConfig.switchFailed'))
    } finally {
      setSourceBusy(false)
    }
  }

  return (
    <section className="space-y-6">
      <div className="section-card">
        <div className="flex items-start justify-between gap-6">
          <div>
            <h2 className="text-2xl font-semibold">{t('lndConfig.title')}</h2>
            <p className="text-fog/60">{t('lndConfig.subtitle')}</p>
            <p className="text-fog/50 text-sm">{t('lndConfig.advancedHint')}</p>
          </div>
          <div className="flex flex-col items-end gap-2">
            <span className="text-xs text-fog/60">{t('lndConfig.bitcoinSource')}</span>
            <button
              className={`relative flex h-9 w-32 items-center rounded-full border border-white/10 bg-ink/60 px-2 transition ${sourceBusy ? 'opacity-70' : 'hover:border-white/30'}`}
              onClick={handleToggleSource}
              type="button"
              disabled={sourceBusy}
              aria-label={t('lndConfig.toggleBitcoinSource')}
            >
              <span
                className={`absolute top-1 h-7 w-14 rounded-full bg-glow shadow transition-all ${bitcoinSource === 'local' ? 'left-[68px]' : 'left-[6px]'}`}
              />
              <span className={`relative z-10 flex-1 text-center text-xs ${bitcoinSource === 'remote' ? 'text-ink' : 'text-fog/60'}`}>{t('common.remote')}</span>
              <span className={`relative z-10 flex-1 text-center text-xs ${bitcoinSource === 'local' ? 'text-ink' : 'text-fog/60'}`}>{t('common.local')}</span>
            </button>
          </div>
        </div>
        {status && <p className="text-sm text-brass mt-4">{status}</p>}
      </div>

      <div className="section-card space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">{t('lndConfig.basicSettings')}</h3>
          <button className="btn-secondary" onClick={() => setAdvanced((v) => !v)}>
            {advanced ? t('lndConfig.hideAdvanced') : t('lndConfig.showAdvanced')}
          </button>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          <div className="space-y-2">
            <label className="text-sm text-fog/70">{t('lndConfig.alias')}</label>
            <input className="input-field" placeholder={t('lndConfig.nodeName')} value={alias} onChange={(e) => setAlias(e.target.value)} />
            <p className="text-xs text-fog/50">{t('lndConfig.aliasHint')}</p>
          </div>
          <div className="space-y-2">
            <label className="text-sm text-fog/70">{t('lndConfig.nodeColor')}</label>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <input
                className="h-12 w-16 rounded-xl border border-white/10 bg-ink/60 p-1"
                type="color"
                value={color}
                onChange={(e) => {
                  setColor(e.target.value)
                  setColorInput(e.target.value)
                }}
              />
              <input
                className="input-field flex-1"
                placeholder={t('lndConfig.colorPlaceholder')}
                value={colorInput}
                onChange={(e) => {
                  const next = e.target.value
                  setColorInput(next)
                  if (isHexColor(next)) {
                    setColor(next)
                  }
                }}
              />
            </div>
            <p className="text-xs text-fog/50">{t('lndConfig.colorHint')}</p>
          </div>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          <div className="space-y-2">
            <label className="text-sm text-fog/70">{t('lndConfig.minChannelSize')}</label>
            <input
              className="input-field"
              placeholder={t('lndConfig.minChannelPlaceholder')}
              type="number"
              min={0}
              value={minChan}
              onChange={(e) => setMinChan(e.target.value)}
            />
            <p className="text-xs text-fog/50">{t('lndConfig.minChannelHint')}</p>
          </div>
          <div className="space-y-2">
            <label className="text-sm text-fog/70">{t('lndConfig.maxChannelSize')}</label>
            <input
              className="input-field"
              placeholder={t('common.optional')}
              type="number"
              min={0}
              value={maxChan}
              onChange={(e) => setMaxChan(e.target.value)}
            />
            <p className="text-xs text-fog/50">{t('lndConfig.maxChannelHint')}</p>
          </div>
        </div>
        <button className="btn-primary" onClick={handleSave}>{t('lndConfig.saveRestart')}</button>
        <p className="text-xs text-fog/50">{t('lndConfig.restartHint')}</p>
      </div>

      {advanced && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">{t('lndConfig.advancedEditor')}</h3>
          <textarea className="input-field min-h-[180px]" value={raw} onChange={(e) => setRaw(e.target.value)} />
          <button className="btn-secondary" onClick={handleSaveRaw}>{t('lndConfig.applyAdvanced')}</button>
        </div>
      )}

      {!config && <p className="text-fog/60">{t('lndConfig.loadingConfig')}</p>}
    </section>
  )
}
