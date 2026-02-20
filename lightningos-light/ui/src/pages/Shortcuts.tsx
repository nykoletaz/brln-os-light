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
  protected: boolean
}

export default function Shortcuts() {
  const { t } = useTranslation()
  const defaultEmoji = '\u26A1'
  const pickerEmoji = '\u{1F603}'
  const emojiOptions = [
    '\u26A1', '\u{1F680}', '\u{1F4B8}', '\u{1F4BC}', '\u{1F4CA}', '\u{1F4E1}', '\u{1F4BB}', '\u{1F310}',
    '\u{1F4F1}', '\u{1F4A1}', '\u{1F512}', '\u{1F527}', '\u{1F6E0}', '\u{1F9F0}', '\u{1F9E0}', '\u{1F525}',
    '\u{1F389}', '\u{1F4AF}', '\u{1F44D}', '\u{1F90C}', '\u2705', '\u{1F7E2}', '\u{1F7E1}', '\u{1F7E0}'
  ]
  const [items, setItems] = useState<ShortcutItem[]>([])
  const [loading, setLoading] = useState(true)
  const [status, setStatus] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formName, setFormName] = useState('')
  const [formURL, setFormURL] = useState('')
  const [formEmoji, setFormEmoji] = useState(defaultEmoji)
  const [showEmojiPicker, setShowEmojiPicker] = useState(false)
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
    if (!formName.trim()) {
      setStatus(t('shortcuts.nameRequired'))
      return
    }
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
      await createShortcut({ name: formName, url: formURL, emoji: formEmoji })
      setFormName('')
      setFormURL('')
      setFormEmoji(defaultEmoji)
      setShowEmojiPicker(false)
      setShowForm(false)
      await load()
    } catch (err: any) {
      setStatus(err?.message || t('shortcuts.createFailed'))
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (item: ShortcutItem) => {
    if (item.protected) {
      return
    }
    if (!window.confirm(t('shortcuts.removeConfirm'))) {
      return
    }
    setDeletingID(item.id)
    setStatus('')
    try {
      await deleteShortcut(item.id)
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
          <div className="grid gap-3 lg:grid-cols-4 lg:items-end">
            <label className="space-y-2">
              <span className="text-sm text-fog/70">{t('shortcuts.nameLabel')}</span>
              <input
                className="input-field"
                placeholder={t('shortcuts.namePlaceholder')}
                value={formName}
                onChange={(event) => setFormName(event.target.value)}
              />
            </label>
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
              <div className="flex items-center gap-2">
                <input
                  className="input-field text-center text-xl"
                  placeholder={defaultEmoji}
                  value={formEmoji}
                  onChange={(event) => setFormEmoji(event.target.value)}
                  maxLength={8}
                />
                <button
                  type="button"
                  className="inline-flex h-12 w-12 items-center justify-center rounded-2xl border border-white/20 bg-white/5 text-lg text-fog transition hover:border-glow/70 hover:text-white"
                  onClick={() => setShowEmojiPicker((current) => !current)}
                  aria-label={t('shortcuts.emojiPickerToggle')}
                  title={t('shortcuts.emojiPickerToggle')}
                >
                  {pickerEmoji}
                </button>
              </div>
            </label>
            <button className="btn-primary h-12" onClick={handleCreate} disabled={saving}>
              {saving ? t('shortcuts.adding') : t('shortcuts.saveShortcut')}
            </button>
          </div>
          {showEmojiPicker && (
            <div className="rounded-2xl border border-white/10 bg-ink/60 p-3">
              <p className="mb-2 text-xs text-fog/60">{t('shortcuts.emojiPickerHint')}</p>
              <div className="grid grid-cols-8 gap-2 sm:grid-cols-12">
                {emojiOptions.map((emoji) => (
                  <button
                    key={emoji}
                    type="button"
                    className="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-white/10 bg-white/5 text-lg transition hover:border-glow/70 hover:bg-white/10"
                    onClick={() => {
                      setFormEmoji(emoji)
                      setShowEmojiPicker(false)
                    }}
                    aria-label={t('shortcuts.selectEmoji')}
                    title={t('shortcuts.selectEmoji')}
                  >
                    {emoji}
                  </button>
                ))}
              </div>
            </div>
          )}
          <p className="text-xs text-fog/60">{t('shortcuts.formHint')}</p>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {items.map((item) => {
          const canDelete = !item.protected
          return (
            <div
              key={item.id}
              className="relative rounded-3xl border border-white/10 bg-gradient-to-br from-white/10 to-white/[0.03] p-5 shadow-panel transition hover:border-glow/40 hover:shadow-[0_18px_40px_rgba(2,6,23,0.45)]"
            >
              {canDelete ? (
                <button
                  type="button"
                  className="absolute right-3 top-3 inline-flex h-8 w-8 items-center justify-center rounded-full border border-white/15 text-fog/60 transition hover:border-ember/70 hover:text-ember"
                  title={t('shortcuts.remove')}
                  aria-label={t('shortcuts.remove')}
                  disabled={deletingID === item.id}
                  onClick={() => handleDelete(item)}
                >
                  x
                </button>
              ) : (
                <span className="absolute right-3 top-3 z-20 rounded-full border border-white/15 bg-ink/80 px-2 py-1 text-[10px] uppercase tracking-[0.1em] text-fog/60">
                  {t('shortcuts.defaultBadge')}
                </span>
              )}
              <a href={item.url} target="_blank" rel="noreferrer" className={`block ${canDelete ? 'pr-10' : 'pr-16'}`}>
                {item.icon_type === 'image' ? (
                  <div className="mb-4 h-16 w-full overflow-hidden rounded-2xl border border-white/15 bg-ink/40 px-3 py-2">
                    <img src={item.icon_value} alt={item.name} className="h-full w-full object-contain object-left" />
                  </div>
                ) : (
                  <div className="mb-4 h-14 w-14 overflow-hidden rounded-2xl border border-white/15 bg-ink/40">
                    <div className="grid h-full w-full place-items-center text-2xl">{item.icon_value}</div>
                  </div>
                )}
              <p className="text-base font-semibold text-fog">{item.name}</p>
              <p className="mt-1 text-sm text-fog/60 break-all">{item.description || item.url}</p>
              <p className="mt-4 text-xs uppercase tracking-[0.14em] text-glow/80">{t('shortcuts.openExternal')}</p>
              </a>
            </div>
          )
        })}
      </div>

      {loading && <p className="text-fog/60">{t('shortcuts.loading')}</p>}
      {!loading && items.length === 0 && <p className="text-fog/60">{t('shortcuts.empty')}</p>}
    </section>
  )
}
