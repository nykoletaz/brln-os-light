import { useEffect, useState } from 'react'
import { getLndConfig, updateLndConfig, updateLndRawConfig } from '../api'

export default function LndConfig() {
  const [config, setConfig] = useState<any>(null)
  const [alias, setAlias] = useState('')
  const [color, setColor] = useState('#ff9900')
  const [colorInput, setColorInput] = useState('#ff9900')
  const [minChan, setMinChan] = useState('')
  const [maxChan, setMaxChan] = useState('')
  const [raw, setRaw] = useState('')
  const [advanced, setAdvanced] = useState(false)
  const [status, setStatus] = useState('')

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
  }, [])

  const isHexColor = (value: string) => /^#[0-9a-fA-F]{6}$/.test(value.trim())

  const handleSave = async () => {
    if (!isHexColor(color)) {
      setStatus('Color must be hex (#RRGGBB).')
      return
    }
    setStatus('Saving...')
    try {
      await updateLndConfig({
        alias,
        color,
        min_channel_size_sat: Number(minChan || 0),
        max_channel_size_sat: Number(maxChan || 0),
        apply_now: true
      })
      setStatus('Saved and applied.')
    } catch {
      setStatus('Save failed.')
    }
  }

  const handleSaveRaw = async () => {
    setStatus('Saving advanced config...')
    try {
      await updateLndRawConfig({ raw_user_conf: raw, apply_now: true })
      setStatus('Advanced config applied.')
    } catch {
      setStatus('Advanced config failed.')
    }
  }

  return (
    <section className="space-y-6">
      <div className="section-card">
        <h2 className="text-2xl font-semibold">LND configuration</h2>
        <p className="text-fog/60">Edit supported values in /data/lnd/lnd.conf.</p>
        <p className="text-fog/50 text-sm">Advanced editor rewrites the full lnd.conf.</p>
        {status && <p className="text-sm text-brass mt-4">{status}</p>}
      </div>

      <div className="section-card space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Basic settings</h3>
          <button className="btn-secondary" onClick={() => setAdvanced((v) => !v)}>
            {advanced ? 'Hide advanced' : 'Show advanced'}
          </button>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          <div className="space-y-2">
            <label className="text-sm text-fog/70">Alias</label>
            <input className="input-field" placeholder="Node name" value={alias} onChange={(e) => setAlias(e.target.value)} />
            <p className="text-xs text-fog/50">Public node name shown in the graph.</p>
          </div>
          <div className="space-y-2">
            <label className="text-sm text-fog/70">Node color</label>
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
                placeholder="#ff9900"
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
            <p className="text-xs text-fog/50">HEX color for your node in the public graph.</p>
          </div>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          <div className="space-y-2">
            <label className="text-sm text-fog/70">Min channel size (sat)</label>
            <input
              className="input-field"
              placeholder="20000"
              type="number"
              min={0}
              value={minChan}
              onChange={(e) => setMinChan(e.target.value)}
            />
            <p className="text-xs text-fog/50">LND enforces a minimum of 20000 sat. Leave blank to use default.</p>
          </div>
          <div className="space-y-2">
            <label className="text-sm text-fog/70">Max channel size (sat)</label>
            <input
              className="input-field"
              placeholder="Optional"
              type="number"
              min={0}
              value={maxChan}
              onChange={(e) => setMaxChan(e.target.value)}
            />
            <p className="text-xs text-fog/50">Leave blank to use the LND default. Must be higher than min size.</p>
          </div>
        </div>
        <button className="btn-primary" onClick={handleSave}>Save and restart LND</button>
        <p className="text-xs text-fog/50">Changes require an LND restart and are applied automatically.</p>
      </div>

      {advanced && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">Advanced editor</h3>
          <textarea className="input-field min-h-[180px]" value={raw} onChange={(e) => setRaw(e.target.value)} />
          <button className="btn-secondary" onClick={handleSaveRaw}>Apply advanced config</button>
        </div>
      )}

      {!config && <p className="text-fog/60">Loading config...</p>}
    </section>
  )
}
