import { useEffect, useState } from 'react'
import { getBitcoin, getDisk, getLndStatus, getPostgres, getSystem, restartService } from '../api'

export default function Dashboard() {
  const [system, setSystem] = useState<any>(null)
  const [disk, setDisk] = useState<any[]>([])
  const [bitcoin, setBitcoin] = useState<any>(null)
  const [postgres, setPostgres] = useState<any>(null)
  const [lnd, setLnd] = useState<any>(null)
  const [status, setStatus] = useState('Loading...')

  const syncLabel = (info: any) => {
    if (!info || typeof info.verification_progress !== 'number') {
      return 'n/a'
    }
    return `${(info.verification_progress * 100).toFixed(2)}%`
  }

  useEffect(() => {
    let mounted = true
    const load = async () => {
      try {
        const [sys, disks, btc, pg, lndStatus] = await Promise.all([
          getSystem(),
          getDisk(),
          getBitcoin(),
          getPostgres(),
          getLndStatus()
        ])
        if (!mounted) return
        setSystem(sys)
        setDisk(Array.isArray(disks) ? disks : [])
        setBitcoin(btc)
        setPostgres(pg)
        setLnd(lndStatus)
        setStatus('OK')
      } catch {
        if (!mounted) return
        setStatus('Unavailable')
      }
    }
    load()
    const timer = setInterval(load, 15000)
    return () => {
      mounted = false
      clearInterval(timer)
    }
  }, [])

  const restart = async (service: string) => {
    await restartService({ service })
  }

  return (
    <section className="space-y-6">
      <div className="section-card">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-sm text-fog/60">System pulse</p>
            <h2 className="text-2xl font-semibold">Overall status: {status}</h2>
          </div>
          <div className="flex gap-2">
            <button className="btn-secondary" onClick={() => restart('lnd')}>Restart LND</button>
            <button className="btn-secondary" onClick={() => restart('lightningos-manager')}>Restart Manager</button>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="section-card">
          <h3 className="text-lg font-semibold">System</h3>
          {system ? (
            <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-fog/60">CPU load</p>
                <p>{system.cpu_load_1?.toFixed?.(2)} / {system.cpu_percent?.toFixed?.(1)}%</p>
              </div>
              <div>
                <p className="text-fog/60">RAM used</p>
                <p>{system.ram_used_mb} / {system.ram_total_mb} MB</p>
              </div>
              <div>
                <p className="text-fog/60">Uptime</p>
                <p>{Math.round(system.uptime_sec / 3600)} hours</p>
              </div>
              <div>
                <p className="text-fog/60">Temp</p>
                <p>{system.temperature_c?.toFixed?.(1)} C</p>
              </div>
            </div>
          ) : (
            <p className="text-fog/60 mt-4">Loading system info...</p>
          )}
        </div>

        <div className="section-card">
          <h3 className="text-lg font-semibold">Bitcoin Remote</h3>
          {bitcoin ? (
            <div className="mt-4 text-sm space-y-2">
              <div className="flex justify-between"><span>Host</span><span>{bitcoin.rpchost}</span></div>
              <div className="flex justify-between"><span>RPC</span><span>{bitcoin.rpc_ok ? 'OK' : 'FAIL'}</span></div>
              <div className="flex justify-between"><span>ZMQ Raw Block</span><span>{bitcoin.zmq_rawblock_ok ? 'OK' : 'FAIL'}</span></div>
              <div className="flex justify-between"><span>ZMQ Raw Tx</span><span>{bitcoin.zmq_rawtx_ok ? 'OK' : 'FAIL'}</span></div>
              {bitcoin.rpc_ok && (
                <>
                  <div className="flex justify-between"><span>Chain</span><span>{bitcoin.chain || 'n/a'}</span></div>
                  <div className="flex justify-between"><span>Blocks</span><span>{bitcoin.blocks ?? 'n/a'}</span></div>
                  <div className="flex justify-between"><span>Sync</span><span>{syncLabel(bitcoin)}</span></div>
                </>
              )}
            </div>
          ) : (
            <p className="text-fog/60 mt-4">Loading bitcoin remote status...</p>
          )}
        </div>

        <div className="section-card">
          <h3 className="text-lg font-semibold">Postgres</h3>
          {postgres ? (
            <div className="mt-4 text-sm space-y-2">
              <div className="flex justify-between"><span>Service</span><span>{postgres.service_active ? 'active' : 'inactive'}</span></div>
              <div className="flex justify-between"><span>DB size</span><span>{postgres.db_size_mb} MB</span></div>
              <div className="flex justify-between"><span>Connections</span><span>{postgres.connections}</span></div>
            </div>
          ) : (
            <p className="text-fog/60 mt-4">Loading postgres status...</p>
          )}
        </div>

        <div className="section-card">
          <h3 className="text-lg font-semibold">LND</h3>
          {lnd ? (
            <div className="mt-4 text-sm space-y-2">
              <div className="flex justify-between"><span>Wallet</span><span>{lnd.wallet_state}</span></div>
              <div className="flex justify-between"><span>Synced</span><span>{lnd.synced_to_chain ? 'chain' : 'not synced'} / {lnd.synced_to_graph ? 'graph' : 'graph pending'}</span></div>
              <div className="flex justify-between"><span>Channels</span><span>{lnd.channels.active} active / {lnd.channels.inactive} inactive</span></div>
              <div className="flex justify-between"><span>Balances</span><span>{lnd.balances.onchain_sat} sat on-chain / {lnd.balances.lightning_sat} sat LN</span></div>
            </div>
          ) : (
            <p className="text-fog/60 mt-4">Loading LND status...</p>
          )}
        </div>
      </div>

      <div className="section-card">
        <h3 className="text-lg font-semibold">Disks</h3>
        {disk.length ? (
          <div className="mt-4 grid gap-3">
            {disk.map((item) => (
              <div key={item.device} className="flex flex-col lg:flex-row lg:items-center lg:justify-between bg-ink/40 rounded-2xl p-4">
                <div>
                  <p className="text-sm text-fog/70">{item.device} ({item.type})</p>
                  <p className="text-xs text-fog/50">Power on hours: {item.power_on_hours}</p>
                </div>
                <div className="text-sm text-fog/80">
                  Wear: {item.wear_percent_used}% | Days left: {item.days_left_estimate}
                </div>
                <div className="text-xs text-fog/60">SMART: {item.smart_status}</div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-fog/60 mt-4">No disk data yet.</p>
        )}
      </div>
    </section>
  )
}
