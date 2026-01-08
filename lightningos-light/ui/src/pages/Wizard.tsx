import { useEffect, useState } from 'react'
import {
  createWalletSeed,
  initWallet,
  postBitcoinRemote,
  unlockWallet,
  getBitcoin
} from '../api'

export default function Wizard() {
  const [step, setStep] = useState(1)
  const [rpcUser, setRpcUser] = useState('')
  const [rpcPass, setRpcPass] = useState('')
  const [bitcoinHost, setBitcoinHost] = useState('')
  const [zmqBlock, setZmqBlock] = useState('')
  const [zmqTx, setZmqTx] = useState('')
  const [bitcoinInfo, setBitcoinInfo] = useState<any>(null)
  const [walletMode, setWalletMode] = useState<'create' | 'import'>('create')
  const [walletPassword, setWalletPassword] = useState('')
  const [walletPasswordConfirm, setWalletPasswordConfirm] = useState('')
  const [seedWords, setSeedWords] = useState<string[]>([])
  const [seedInput, setSeedInput] = useState('')
  const [ackSeed, setAckSeed] = useState(false)
  const [unlockPass, setUnlockPass] = useState('')
  const [status, setStatus] = useState('')

  useEffect(() => {
    getBitcoin().then((data: any) => {
      setBitcoinHost(data.rpchost)
      setZmqBlock(data.zmq_rawblock)
      setZmqTx(data.zmq_rawtx)
      setBitcoinInfo(data)
    }).catch(() => null)
  }, [])

  const next = () => setStep((prev) => Math.min(prev + 1, 4))

  const syncLabel = (info: any) => {
    if (!info || typeof info.verification_progress !== 'number') {
      return 'n/a'
    }
    return `${(info.verification_progress * 100).toFixed(2)}%`
  }

  const handleBitcoin = async () => {
    setStatus('Testing connection...')
    try {
      const res = await postBitcoinRemote({ rpcuser: rpcUser, rpcpass: rpcPass })
      const info = res?.info
      if (info) {
        setBitcoinInfo(info)
        setStatus(`RPC OK. ${info.chain || 'main'} @ ${info.blocks ?? 'n/a'} (${syncLabel(info)})`)
      } else {
        setStatus('Saved. RPC OK.')
      }
      next()
    } catch (err: any) {
      setStatus(err?.message || 'Failed to validate RPC. Check credentials.')
    }
  }

  const handleGenerateSeed = async () => {
    if ((walletPassword || walletPasswordConfirm) && walletPassword !== walletPasswordConfirm) {
      setStatus('Wallet password mismatch.')
      return
    }
    setStatus('Generating seed...')
    try {
      const res = await createWalletSeed()
      setSeedWords(res.seed_words || [])
      setStatus('Seed generated. Write it down.')
    } catch (err: any) {
      setStatus(err?.message || 'Seed generation failed.')
    }
  }

  const handleInitWallet = async () => {
    if (!walletPassword || walletPassword !== walletPasswordConfirm) {
      setStatus('Wallet password mismatch.')
      return
    }

    const words = walletMode === 'create' ? seedWords : seedInput.trim().split(/\s+/)
    if (words.length < 12) {
      setStatus('Seed words invalid.')
      return
    }
    if (walletMode === 'create' && !ackSeed) {
      setStatus('Confirm that you wrote down the seed.')
      return
    }

    setStatus('Initializing wallet...')
    try {
      await initWallet({ wallet_password: walletPassword, seed_words: words })
      setStatus('Wallet initialized.')
      next()
    } catch {
      setStatus('Wallet init failed.')
    }
  }

  const handleUnlock = async () => {
    if (!unlockPass) {
      setStatus('Enter wallet password.')
      return
    }
    setStatus('Unlocking...')
    try {
      await unlockWallet({ wallet_password: unlockPass })
      setStatus('Unlocked.')
      next()
    } catch {
      setStatus('Unlock failed.')
    }
  }

  return (
    <section className="space-y-6">
      <div className="section-card">
        <h2 className="text-2xl font-semibold">Welcome wizard</h2>
        <p className="text-fog/60 mt-2">Follow the guided setup for Bitcoin remote, LND wallet, and unlock.</p>
        {status && <p className="mt-4 text-sm text-brass">{status}</p>}
      </div>

      {step === 1 && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">Step 1: Connect to Bitcoin remote</h3>
          <div className="grid gap-4 lg:grid-cols-3 text-sm text-fog/60">
            <div>
              <p className="uppercase text-xs">RPC Host</p>
              <p>{bitcoinHost || 'bitcoin.br-ln.com:8085'}</p>
            </div>
            <div>
              <p className="uppercase text-xs">ZMQ Raw Block</p>
              <p>{zmqBlock || 'tcp://bitcoin.br-ln.com:28332'}</p>
            </div>
            <div>
              <p className="uppercase text-xs">ZMQ Raw Tx</p>
              <p>{zmqTx || 'tcp://bitcoin.br-ln.com:28333'}</p>
            </div>
          </div>
          <div className="grid gap-4 lg:grid-cols-3 text-sm text-fog/60">
            <div>
              <p className="uppercase text-xs">Chain</p>
              <p>{bitcoinInfo?.chain || 'n/a'}</p>
            </div>
            <div>
              <p className="uppercase text-xs">Blocks</p>
              <p>{bitcoinInfo?.blocks ?? 'n/a'}</p>
            </div>
            <div>
              <p className="uppercase text-xs">Sync</p>
              <p>{syncLabel(bitcoinInfo)}</p>
            </div>
          </div>
          <div className="grid gap-4 lg:grid-cols-2">
            <input className="input-field" placeholder="RPC user" value={rpcUser} onChange={(e) => setRpcUser(e.target.value)} />
            <input className="input-field" placeholder="RPC password" type="password" value={rpcPass} onChange={(e) => setRpcPass(e.target.value)} />
          </div>
          <button className="btn-primary" onClick={handleBitcoin}>Test and save</button>
        </div>
      )}

      {step === 2 && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">Step 2: LND wallet</h3>
          <div className="flex gap-3">
            <button className={walletMode === 'create' ? 'btn-primary' : 'btn-secondary'} onClick={() => setWalletMode('create')}>Create new</button>
            <button className={walletMode === 'import' ? 'btn-primary' : 'btn-secondary'} onClick={() => setWalletMode('import')}>Import existing</button>
          </div>
          <div className="grid gap-4 lg:grid-cols-2">
            <input className="input-field" placeholder="Wallet password" type="password" value={walletPassword} onChange={(e) => setWalletPassword(e.target.value)} />
            <input className="input-field" placeholder="Confirm password" type="password" value={walletPasswordConfirm} onChange={(e) => setWalletPasswordConfirm(e.target.value)} />
          </div>
          <p className="text-xs text-fog/60">Password is required to initialize and unlock the wallet. Seed generation does not require it.</p>

          {walletMode === 'create' ? (
            <div className="space-y-3">
              <button className="btn-secondary" onClick={handleGenerateSeed}>Generate seed</button>
              {seedWords.length > 0 && (
                <div className="bg-ink/50 rounded-2xl p-4 text-sm">
                  <p className="text-brass text-xs uppercase">Seed words (write once)</p>
                  <div className="mt-2 flex flex-wrap gap-2">
                    {seedWords.map((word: string, idx: number) => (
                      <span key={word + idx} className="px-2 py-1 rounded-xl bg-white/5 border border-white/10">{word}</span>
                    ))}
                  </div>
                  <label className="mt-4 flex items-center gap-2 text-xs text-fog/70">
                    <input type="checkbox" checked={ackSeed} onChange={(e) => setAckSeed(e.target.checked)} />
                    I wrote down the 24 words and understand they cannot be recovered.
                  </label>
                </div>
              )}
              <button className="btn-primary" onClick={handleInitWallet}>Initialize wallet</button>
            </div>
          ) : (
            <div className="space-y-3">
              <textarea className="input-field min-h-[120px]" placeholder="Paste 24 seed words" value={seedInput} onChange={(e) => setSeedInput(e.target.value)} />
              <button className="btn-primary" onClick={handleInitWallet}>Import and initialize</button>
            </div>
          )}
        </div>
      )}

      {step === 3 && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">Step 3: Unlock wallet</h3>
          <input className="input-field" placeholder="Wallet password" type="password" value={unlockPass} onChange={(e) => setUnlockPass(e.target.value)} />
          <button className="btn-primary" onClick={handleUnlock}>Unlock</button>
        </div>
      )}

      {step === 4 && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">Step 4: Finish</h3>
          <p className="text-fog/60">Setup complete. Jump to your dashboard.</p>
          <a className="btn-primary inline-flex" href="#dashboard">Go to Dashboard</a>
        </div>
      )}
    </section>
  )
}
