import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { createShortcut, deleteShortcut, getShortcuts } from '../api'

type ShortcutItem = {
  id: number
  url: string
  name: string
  description: string
  icon_type: 'emoji' | 'image'
  icon_value: string
  sort_order: number
}

export default function Shortcuts() {
  const { t } = useTranslation()
  const defaultEmoji = '\u26A1'
  const [items, setItems] = useState<ShortcutItem[]>([])
  const [loading, setLoading] = useState(true)
  const [status, setStatus] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formURL, setFormURL] = useState('')
  const [formEmoji, setFormEmoji] = useState(defaultEmoji)
  const [saving, setSaving] = useState(false)
  const [deletingID, setDeletingID] = useState<number | null>(null)

  const load = async () => {
    setLoading(true)
    setStatus('')
    try {
      const res: any = await getShortcuts()
      const loaded = Array.isArray(res?.items) ? res.items : []
      setItems(loaded)
    } catch (err: any) {
      setStatus(err?.message || t('shortcuts.loadFailed'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const handleCreate = async () => {
    if (!formURL.trim()) {
      setStatus(t('shortcuts.urlRequired'))
      return
    }
    if (!formEmoji.trim()) {
      setStatus(t('shortcuts.emojiRequired'))
      return
    }
    setSaving(true)
    setStatus('')
    try {
      await createShortcut({ url: formURL, emoji: formEmoji })
      setFormURL('')
      setFormEmoji(defaultEmoji)
      setShowForm(false)
      await load()
    } catch (err: any) {
      setStatus(err?.message || t('shortcuts.createFailed'))
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: number) => {
    if (!window.confirm(t('shortcuts.removeConfirm'))) {
      return
    }
    setDeletingID(id)
    setStatus('')
    try {
      await deleteShortcut(id)
      await load()
    } catch (err: any) {
      setStatus(err?.message || t('shortcuts.removeFailed'))
    } finally {
      setDeletingID(null)
    }
  }

  return (
    <section className="space-y-6">
      <div className="section-card">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="text-2xl font-semibold">{t('shortcuts.title')}</h2>
            <p className="text-fog/60">{t('shortcuts.subtitle')}</p>
          </div>
          <button
            type="button"
            className="inline-flex h-11 w-11 items-center justify-center rounded-full border border-white/20 bg-white/5 text-xl text-fog transition hover:border-glow/70 hover:text-white"
            aria-label={t('shortcuts.add')}
            title={t('shortcuts.add')}
            onClick={() => setShowForm((current) => !current)}
          >
            +
          </button>
        </div>
        {status && <p className="mt-4 text-sm text-brass">{status}</p>}
      </div>

      {showForm && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">{t('shortcuts.newShortcut')}</h3>
          <div className="grid gap-3 lg:grid-cols-[1fr_160px_auto] lg:items-end">
            <label className="space-y-2">
              <span className="text-sm text-fog/70">{t('shortcuts.urlLabel')}</span>
              <input
                className="input-field"
                placeholder="https://services.br-ln.com"
                value={formURL}
                onChange={(event) => setFormURL(event.target.value)}
              />
            </label>
            <label className="space-y-2">
              <span className="text-sm text-fog/70">{t('shortcuts.emojiLabel')}</span>
              <input
                className="input-field text-center text-xl"
                placeholder={defaultEmoji}
                value={formEmoji}
                onChange={(event) => setFormEmoji(event.target.value)}
                maxLength={8}
              />
            </label>
            <button className="btn-primary h-12" onClick={handleCreate} disabled={saving}>
              {saving ? t('shortcuts.adding') : t('shortcuts.saveShortcut')}
            </button>
          </div>
          <p className="text-xs text-fog/60">{t('shortcuts.formHint')}</p>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {items.map((item) => (
          <div
            key={item.id}
            className="relative rounded-3xl border border-white/10 bg-gradient-to-br from-white/10 to-white/[0.03] p-5 shadow-panel transition hover:border-glow/40 hover:shadow-[0_18px_40px_rgba(2,6,23,0.45)]"
          >
            <button
              type="button"
              className="absolute right-3 top-3 inline-flex h-8 w-8 items-center justify-center rounded-full border border-white/15 text-fog/60 transition hover:border-ember/70 hover:text-ember"
              title={t('shortcuts.remove')}
              aria-label={t('shortcuts.remove')}
              disabled={deletingID === item.id}
              onClick={() => handleDelete(item.id)}
            >
              x
            </button>
            <a href={item.url} target="_blank" rel="noreferrer" className="block pr-10">
              <div className="mb-4 h-14 w-14 overflow-hidden rounded-2xl border border-white/15 bg-ink/40">
                {item.icon_type === 'image' ? (
                  <img src={item.icon_value} alt={item.name} className="h-full w-full object-cover" />
                ) : (
                  <div className="grid h-full w-full place-items-center text-2xl">{item.icon_value}</div>
                )}
              </div>
              <p className="text-base font-semibold text-fog">{item.name}</p>
              <p className="mt-1 text-sm text-fog/60 break-all">{item.description || item.url}</p>
              <p className="mt-4 text-xs uppercase tracking-[0.14em] text-glow/80">{t('shortcuts.openExternal')}</p>
            </a>
          </div>
        ))}
      </div>

      {loading && <p className="text-fog/60">{t('shortcuts.loading')}</p>}
      {!loading && items.length === 0 && <p className="text-fog/60">{t('shortcuts.empty')}</p>}
    </section>
  )
}
