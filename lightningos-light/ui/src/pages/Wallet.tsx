import { useEffect, useState } from 'react'
import { createInvoice, getWalletSummary, payInvoice } from '../api'

const emptySummary = {
  balances: {
    onchain_sat: 0,
    lightning_sat: 0
  },
  activity: []
}

export default function Wallet() {
  const [summary, setSummary] = useState<any>(emptySummary)
  const [summaryError, setSummaryError] = useState('')
  const [summaryWarning, setSummaryWarning] = useState('')
  const [summaryLoading, setSummaryLoading] = useState(true)
  const [amount, setAmount] = useState('')
  const [memo, setMemo] = useState('')
  const [invoice, setInvoice] = useState('')
  const [paymentRequest, setPaymentRequest] = useState('')
  const [status, setStatus] = useState('')

  useEffect(() => {
    let mounted = true
    const load = async () => {
      setSummaryError('')
      setSummaryWarning('')
      try {
        const data = await getWalletSummary()
        if (!mounted) return
        setSummary(data || emptySummary)
        setSummaryWarning(data?.warning || '')
      } catch (err: any) {
        if (!mounted) return
        const message = err?.message || 'Wallet summary unavailable'
        setSummaryError(message)
      } finally {
        if (!mounted) return
        setSummaryLoading(false)
      }
    }
    load()
    const timer = setInterval(load, 15000)
    return () => {
      mounted = false
      clearInterval(timer)
    }
  }, [])

  const onchainBalance = summary?.balances?.onchain_sat ?? 0
  const lightningBalance = summary?.balances?.lightning_sat ?? 0
  const activity = summary?.activity ?? []

  const handleInvoice = async () => {
    setStatus('Creating invoice...')
    try {
      const res = await createInvoice({ amount_sat: Number(amount), memo })
      setInvoice(res.payment_request)
      setStatus('Invoice ready.')
    } catch {
      setStatus('Invoice failed.')
    }
  }

  const handlePay = async () => {
    setStatus('Paying invoice...')
    try {
      await payInvoice({ payment_request: paymentRequest })
      setStatus('Payment sent.')
    } catch {
      setStatus('Payment failed.')
    }
  }

  return (
    <section className="space-y-6">
      <div className="section-card">
        <h2 className="text-2xl font-semibold">Wallet</h2>
        <p className="text-fog/60">Manage Lightning and on-chain balances.</p>
        <div className="mt-4 grid gap-4 lg:grid-cols-2 text-sm">
          <div>
            <p className="text-fog/60">On-chain</p>
            <p className="text-xl">{onchainBalance} sat</p>
          </div>
          <div>
            <p className="text-fog/60">Lightning</p>
            <p className="text-xl">{lightningBalance} sat</p>
          </div>
        </div>
        {summaryLoading && !summaryError && (
          <p className="mt-4 text-sm text-fog/60">Fetching wallet balances...</p>
        )}
        {summaryWarning && !summaryError && (
          <p className="mt-4 text-sm text-brass">{summaryWarning}</p>
        )}
        {summaryError && (
          <p className="mt-4 text-sm text-ember">Wallet status: {summaryError}</p>
        )}
        {status && <p className="mt-4 text-sm text-brass">{status}</p>}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">Create invoice</h3>
          <input className="input-field" placeholder="Amount (sat)" value={amount} onChange={(e) => setAmount(e.target.value)} />
          <input className="input-field" placeholder="Memo" value={memo} onChange={(e) => setMemo(e.target.value)} />
          <button className="btn-primary" onClick={handleInvoice}>Generate invoice</button>
          {invoice && (
            <textarea className="input-field min-h-[120px]" value={invoice} readOnly />
          )}
        </div>

        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold">Pay invoice</h3>
          <textarea className="input-field min-h-[140px]" placeholder="Paste payment request" value={paymentRequest} onChange={(e) => setPaymentRequest(e.target.value)} />
          <button className="btn-primary" onClick={handlePay}>Pay invoice</button>
        </div>
      </div>

      <div className="section-card">
        <h3 className="text-lg font-semibold">Recent activity</h3>
        <div className="mt-4 space-y-2 text-sm">
          {summaryError ? (
            <p className="text-fog/60">Activity unavailable until LND is reachable.</p>
          ) : activity.length ? activity.map((item: any, idx: number) => (
            <div key={`${item.type}-${idx}`} className="flex items-center justify-between border-b border-white/10 pb-2">
              <span className="text-fog/70">{item.type} - {item.status}</span>
              <span>{item.amount_sat} sat</span>
            </div>
          )) : (
            <p className="text-fog/60">No recent activity yet.</p>
          )}
        </div>
      </div>
    </section>
  )
}
