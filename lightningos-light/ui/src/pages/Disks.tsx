import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getDisk } from '../api'

export default function Disks() {
  const { t } = useTranslation()
  const [disks, setDisks] = useState<any[]>([])

  useEffect(() => {
    getDisk()
      .then((data) => setDisks(Array.isArray(data) ? data : []))
      .catch(() => null)
  }, [])

  return (
    <section className="space-y-6">
      <div className="section-card">
        <h2 className="text-2xl font-semibold">{t('disks.title')}</h2>
        <p className="text-fog/60">{t('disks.subtitle')}</p>
      </div>

      <div className="section-card space-y-4">
        {disks.length ? (
          disks.map((disk) => (
            <div key={disk.device} className="border border-white/10 rounded-2xl p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-fog/60">{disk.device}</p>
                  <p className="text-lg font-semibold">{disk.type}</p>
                </div>
                <div className="text-sm text-fog/70">{t('disks.smartLabel', { status: disk.smart_status })}</div>
              </div>
              <div className="mt-3 grid gap-3 lg:grid-cols-3 text-sm">
                <div>
                  <p className="text-fog/60">{t('disks.wear')}</p>
                  <p>{disk.wear_percent_used}%</p>
                </div>
                <div>
                  <p className="text-fog/60">{t('disks.powerOnHours')}</p>
                  <p>{disk.power_on_hours}</p>
                </div>
                <div>
                  <p className="text-fog/60">{t('disks.daysLeft')}</p>
                  <p>{disk.days_left_estimate}</p>
                </div>
              </div>
              {disk.alerts?.length ? (
                <p className="mt-2 text-xs text-ember">{t('disks.alerts', { alerts: disk.alerts.join(', ') })}</p>
              ) : (
                <p className="mt-2 text-xs text-fog/50">{t('disks.noAlerts')}</p>
              )}
            </div>
          ))
        ) : (
          <p className="text-fog/60">{t('disks.noData')}</p>
        )}
      </div>
    </section>
  )
}
