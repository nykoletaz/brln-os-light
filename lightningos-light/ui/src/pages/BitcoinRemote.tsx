import { useEffect, useState } from 'react'
import { getBitcoin, postBitcoinRemote } from '../api'

export default function BitcoinRemote() {
  const [status, setStatus] = useState<any>(null)
  const [rpcUser, setRpcUser] = useState('')
  const [rpcPass, setRpcPass] = useState('')
  const [message, setMessage] = useState('')

  const syncLabel = (info: any) => {
    if (!info || typeof info.verification_progress !== 'number') {
      return 'n/a'
    }
    return `${(info.verification_progress * 100).toFixed(2)}%`
  }

  useEffect(() => {
    getBitcoin().then(setStatus).catch(() => null)
  }, [])

  const handleSave = async () => {
    setMessage('Saving...')
    try {
      await postBitcoinRemote({ rpcuser: rpcUser, rpcpass: rpcPass })
      setMessage('Saved. RPC OK.')
      const updated = await getBitcoin()
      setStatus(updated)
      setRpcPass('')
    } catch (err: any) {
      setMessage(err?.message || 'Validation failed. Check credentials.')
    }
  }

  return (
    <section className="space-y-6">
      <div className="section-card">
        <h2 className="text-2xl font-semibold">Bitcoin Remote</h2>
        <p className="text-fog/60">Default BRLN node, mainnet only.</p>
      </div>

      <div className="section-card space-y-4">
        <h3 className="text-lg font-semibold">Connection</h3>
        <div className="grid gap-4 lg:grid-cols-3 text-sm">
          <div>
            <p className="text-fog/60">RPC Host</p>
            <p>{status?.rpchost || 'bitcoin.br-ln.com:8085'}</p>
          </div>
          <div>
            <p className="text-fog/60">RPC Status</p>
            <p>{status?.rpc_ok ? 'OK' : 'FAIL'}</p>
          </div>
          <div>
            <p className="text-fog/60">ZMQ</p>
            <p>{status?.zmq_rawblock_ok && status?.zmq_rawtx_ok ? 'OK' : 'CHECK'}</p>
          </div>
        </div>
        <div className="grid gap-4 lg:grid-cols-3 text-sm">
          <div>
            <p className="text-fog/60">Chain</p>
            <p>{status?.chain || 'n/a'}</p>
          </div>
          <div>
            <p className="text-fog/60">Blocks</p>
            <p>{status?.blocks ?? 'n/a'}</p>
          </div>
          <div>
            <p className="text-fog/60">Sync</p>
            <p>{syncLabel(status)}</p>
          </div>
        </div>
      </div>

      <div className="section-card space-y-4">
        <h3 className="text-lg font-semibold">Update RPC credentials</h3>
        <div className="grid gap-4 lg:grid-cols-2">
          <input className="input-field" placeholder="RPC user" value={rpcUser} onChange={(e) => setRpcUser(e.target.value)} />
          <input className="input-field" placeholder="RPC password" type="password" value={rpcPass} onChange={(e) => setRpcPass(e.target.value)} />
        </div>
        <button className="btn-primary" onClick={handleSave}>Save</button>
        {message && <p className="text-sm text-brass">{message}</p>}
      </div>
    </section>
  )
}
