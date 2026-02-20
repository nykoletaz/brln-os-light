import { useEffect, useState, useCallback, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { getBoletoConfig, activateBoleto, getActivationStatus, createBoletoQuote, getBoletoStatus } from '../api'
import { getLocale } from '../i18n'

type BoletoConfig = {
  enabled: boolean
  activated: boolean
  feePercent: number
  provider: string
}

type ActivationData = {
  invoice: string
  paymentHash: string
  sats: number
  expiresAt: string
  message?: string
}

type BoletoQuote = {
  invoice: string
  paymentHash: string
  amountBrl: number
  baseSats: number
  feeSats: number
  feePercent: number
  totalSats: number
  btcRate: number
  bankName: string
  bankCode: string
  dueDate?: string | null
  expiresAt: string
}

type BoletoStatus = {
  status: string
  invoicePaid: boolean
  boletoPaid: boolean
  protocolo?: string
  amountBrl?: number
  totalSats?: number
  reason?: string
}

type Step = 'input' | 'confirm' | 'paying' | 'done' | 'error'

const fmt = (n: number, locale: string) =>
  new Intl.NumberFormat(locale, { style: 'currency', currency: 'BRL' }).format(n)

const fmtSats = (n: number, locale: string) =>
  new Intl.NumberFormat(locale).format(n) + ' sats'

export default function PayBoleto() {
  const { t } = useTranslation()
  const locale = getLocale()

  const [config, setConfig] = useState<BoletoConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [step, setStep] = useState<Step>('input')
  const [barcode, setBarcode] = useState('')
  const [error, setError] = useState('')
  const [quoting, setQuoting] = useState(false)
  const [quote, setQuote] = useState<BoletoQuote | null>(null)
  const [status, setStatus] = useState<BoletoStatus | null>(null)
  const [copied, setCopied] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const [history, setHistory] = useState<Array<{ paymentHash: string; amountBrl: number; totalSats: number; status: string; bankName: string; createdAt: string }>>([])

  // Activation state
  const [activating, setActivating] = useState(false)
  const [activation, setActivation] = useState<ActivationData | null>(null)
  const [activationCopied, setActivationCopied] = useState(false)
  const activationPollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Load config on mount
  useEffect(() => {
    getBoletoConfig()
      .then(setConfig)
      .catch(() => setConfig({ enabled: false, activated: false, feePercent: 6, provider: '' }))
      .finally(() => setLoading(false))
  }, [])

  // Cleanup poll on unmount
  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
      if (activationPollRef.current) clearInterval(activationPollRef.current)
    }
  }, [])

  // ─── Activation flow ────────────────────────────────────────────────

  const handleActivate = useCallback(async () => {
    setActivating(true)
    setError('')
    try {
      const data = await activateBoleto()
      setActivation(data)

      // Start polling for activation payment
      const poll = async () => {
        try {
          const s: any = await getActivationStatus(data.paymentHash)
          if (s.status === 'completed' && s.apiKey) {
            if (activationPollRef.current) clearInterval(activationPollRef.current)
            setActivation(null)
            setActivating(false)
            // Refresh config (key is now saved server-side)
            const cfg = await getBoletoConfig()
            setConfig(cfg)
          } else if (s.status === 'expired') {
            if (activationPollRef.current) clearInterval(activationPollRef.current)
            setActivation(null)
            setActivating(false)
            setError(t('boleto.activationExpired'))
          }
        } catch {
          // continue polling
        }
      }

      poll()
      activationPollRef.current = setInterval(poll, 4000)
    } catch (err: any) {
      setError(err.message || t('boleto.activationError'))
      setActivating(false)
    }
  }, [t])

  const copyActivationInvoice = useCallback(() => {
    if (!activation) return
    navigator.clipboard.writeText(activation.invoice).then(() => {
      setActivationCopied(true)
      setTimeout(() => setActivationCopied(false), 2000)
    })
  }, [activation])

  const handleQuote = useCallback(async () => {
    setError('')
    const clean = barcode.replace(/[\s.-]/g, '')
    if (clean.length < 44) {
      setError(t('boleto.errorMinDigits'))
      return
    }
    setQuoting(true)
    try {
      const q = await createBoletoQuote(clean)
      setQuote(q)
      setStep('confirm')
    } catch (err: any) {
      setError(err.message || t('boleto.errorQuote'))
    } finally {
      setQuoting(false)
    }
  }, [barcode, t])

  const handleConfirm = useCallback(() => {
    if (!quote) return
    setStep('paying')
    setCopied(false)

    // Save to history
    setHistory(prev => [{
      paymentHash: quote.paymentHash,
      amountBrl: quote.amountBrl,
      totalSats: quote.totalSats,
      status: 'pending',
      bankName: quote.bankName,
      createdAt: new Date().toISOString(),
    }, ...prev].slice(0, 20))

    // Start polling
    const poll = async () => {
      try {
        const s = await getBoletoStatus(quote.paymentHash)
        setStatus(s)
        if (s.status === 'completed') {
          setStep('done')
          setHistory(prev => prev.map(h =>
            h.paymentHash === quote.paymentHash ? { ...h, status: 'completed' } : h
          ))
          if (pollRef.current) clearInterval(pollRef.current)
        } else if (s.status === 'failed' || s.status === 'expired') {
          setStep('error')
          setError(s.reason || s.status)
          setHistory(prev => prev.map(h =>
            h.paymentHash === quote.paymentHash ? { ...h, status: s.status } : h
          ))
          if (pollRef.current) clearInterval(pollRef.current)
        }
      } catch {
        // continue polling
      }
    }

    poll()
    pollRef.current = setInterval(poll, 4000)
  }, [quote])

  const handleReset = useCallback(() => {
    setStep('input')
    setBarcode('')
    setQuote(null)
    setStatus(null)
    setError('')
    setCopied(false)
    if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
  }, [])

  const copyInvoice = useCallback(() => {
    if (!quote) return
    navigator.clipboard.writeText(quote.invoice).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }, [quote])

  if (loading) {
    return (
      <section className="space-y-6">
        <div className="section-card animate-pulse">
          <div className="h-6 bg-white/10 rounded w-48 mb-4" />
          <div className="h-4 bg-white/10 rounded w-72" />
        </div>
      </section>
    )
  }

  if (!config?.enabled) {
    return (
      <section className="space-y-6">
        <div className="section-card">
          <h2 className="text-xl font-bold text-fog mb-2">{t('boleto.title')}</h2>
          <p className="text-fog/60">{t('boleto.disabled')}</p>
        </div>
      </section>
    )
  }

  // ─── Activation screen ───────────────────────────────────────────────

  if (!config.activated) {
    return (
      <section className="space-y-6">
        <div className="section-card">
          <h2 className="text-xl font-bold text-fog mb-1">{t('boleto.title')}</h2>
          <p className="text-fog/60 text-sm">{t('boleto.subtitle')}</p>
        </div>

        <div className="section-card space-y-4">
          {!activation ? (
            <>
              <h3 className="text-lg font-semibold text-fog">{t('boleto.activationTitle')}</h3>
              <p className="text-fog/70 text-sm">{t('boleto.activationDesc', { sats: '1,000' })}</p>
              <ul className="text-fog/60 text-sm space-y-1 list-disc list-inside">
                <li>{t('boleto.activationStep1')}</li>
                <li>{t('boleto.activationStep2')}</li>
                <li>{t('boleto.activationStep3')}</li>
              </ul>
              {error && <p className="text-red-400 text-sm">{error}</p>}
              <button
                className="btn-primary"
                onClick={handleActivate}
                disabled={activating}
              >
                {activating ? t('boleto.activating') : t('boleto.activateBtn')}
              </button>
            </>
          ) : (
            <>
              <h3 className="text-lg font-semibold text-fog">{t('boleto.activationPayTitle')}</h3>
              <p className="text-fog/70 text-sm">{t('boleto.activationPayDesc', { sats: new Intl.NumberFormat(locale).format(activation.sats) })}</p>

              <div className="bg-white rounded-xl p-4 mx-auto w-fit">
                <img
                  src={`https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=${encodeURIComponent(activation.invoice)}`}
                  alt="QR Code"
                  className="w-[220px] h-[220px]"
                />
              </div>

              <div>
                <label className="block text-sm text-fog/70 mb-1">{t('boleto.invoiceLabel')}</label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    className="input-field font-mono text-xs"
                    value={activation.invoice}
                    readOnly
                  />
                  <button className="btn-secondary whitespace-nowrap" onClick={copyActivationInvoice}>
                    {activationCopied ? t('boleto.copied') : t('boleto.copyInvoice')}
                  </button>
                </div>
              </div>

              <div className="flex items-center gap-2 text-fog/60 text-sm">
                <div className="w-3 h-3 rounded-full bg-amber-400 animate-pulse" />
                {t('boleto.activationWaiting')}
              </div>
            </>
          )}
        </div>
      </section>
    )
  }

  return (
    <section className="space-y-6">
      {/* Header */}
      <div className="section-card">
        <h2 className="text-xl font-bold text-fog mb-1">{t('boleto.title')}</h2>
        <p className="text-fog/60 text-sm">{t('boleto.subtitle')}</p>
      </div>

      {/* Step: Input barcode */}
      {step === 'input' && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold text-fog">{t('boleto.inputTitle')}</h3>
          <div>
            <label className="block text-sm text-fog/70 mb-1">{t('boleto.barcodeLabel')}</label>
            <input
              type="text"
              className="input-field font-mono"
              placeholder={t('boleto.barcodePlaceholder')}
              value={barcode}
              onChange={e => setBarcode(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleQuote()}
              disabled={quoting}
            />
          </div>
          {error && <p className="text-ember text-sm">{error}</p>}
          <button
            className="btn-primary w-full"
            onClick={handleQuote}
            disabled={quoting || barcode.replace(/[\s.-]/g, '').length < 44}
          >
            {quoting ? t('boleto.quoting') : t('boleto.getQuote')}
          </button>

          <div className="border-t border-white/10 pt-3 mt-2">
            <p className="text-xs text-fog/50">
              {t('boleto.feeInfo', { percent: config.feePercent })}
            </p>
          </div>
        </div>
      )}

      {/* Step: Confirm quote */}
      {step === 'confirm' && quote && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold text-fog">{t('boleto.confirmTitle')}</h3>

          <div className="grid grid-cols-2 gap-3 text-sm">
            <div className="bg-ink/40 rounded-2xl p-3">
              <span className="text-fog/50 block">{t('boleto.bank')}</span>
              <span className="text-fog font-medium">{quote.bankName}</span>
            </div>
            <div className="bg-ink/40 rounded-2xl p-3">
              <span className="text-fog/50 block">{t('boleto.valueBrl')}</span>
              <span className="text-fog font-medium">{fmt(quote.amountBrl, locale)}</span>
            </div>
            <div className="bg-ink/40 rounded-2xl p-3">
              <span className="text-fog/50 block">{t('boleto.baseSats')}</span>
              <span className="text-fog font-medium">{fmtSats(quote.baseSats, locale)}</span>
            </div>
            <div className="bg-ink/40 rounded-2xl p-3">
              <span className="text-fog/50 block">{t('boleto.fee')}</span>
              <span className="text-brass font-medium">+{fmtSats(quote.feeSats, locale)} ({quote.feePercent}%)</span>
            </div>
          </div>

          <div className="bg-glow/10 border border-glow/30 rounded-2xl p-4 text-center">
            <span className="text-fog/60 text-sm block mb-1">{t('boleto.totalLabel')}</span>
            <span className="text-2xl font-bold text-glow">{fmtSats(quote.totalSats, locale)}</span>
          </div>

          <div className="text-xs text-fog/40 text-center">
            {t('boleto.rate')}: 1 BTC = {fmt(quote.btcRate, locale)}
            {quote.dueDate && (
              <> · {t('boleto.dueDate')}: {new Date(quote.dueDate).toLocaleDateString(locale)}</>
            )}
          </div>

          <div className="flex gap-3">
            <button className="btn-secondary flex-1" onClick={handleReset}>
              {t('boleto.cancel')}
            </button>
            <button className="btn-primary flex-1" onClick={handleConfirm}>
              {t('boleto.confirmPay')}
            </button>
          </div>
        </div>
      )}

      {/* Step: Paying — waiting for LN payment */}
      {step === 'paying' && quote && (
        <div className="section-card space-y-4">
          <h3 className="text-lg font-semibold text-fog">{t('boleto.payingTitle')}</h3>
          <p className="text-fog/60 text-sm">{t('boleto.payingSubtitle')}</p>

          {/* Invoice to pay */}
          <div className="bg-ink/40 rounded-2xl p-4 space-y-2">
            <span className="text-fog/50 text-xs block">{t('boleto.invoiceLabel')}</span>
            <div className="font-mono text-xs text-fog/80 break-all leading-relaxed max-h-24 overflow-y-auto">
              {quote.invoice}
            </div>
            <button
              className="btn-secondary text-xs py-1 px-3"
              onClick={copyInvoice}
            >
              {copied ? t('boleto.copied') : t('boleto.copyInvoice')}
            </button>
          </div>

          <div className="flex items-center gap-2 text-sm text-fog/60">
            <svg className="w-4 h-4 animate-spin text-glow" viewBox="0 0 24 24" fill="none">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            {status?.invoicePaid
              ? t('boleto.processingBoleto')
              : t('boleto.waitingPayment')}
          </div>

          <div className="text-center">
            <span className="text-2xl font-bold text-glow">{fmtSats(quote.totalSats, locale)}</span>
            <span className="text-fog/50 text-sm block">{fmt(quote.amountBrl, locale)}</span>
          </div>

          <button className="btn-secondary w-full text-sm" onClick={handleReset}>
            {t('boleto.cancel')}
          </button>
        </div>
      )}

      {/* Step: Done */}
      {step === 'done' && quote && status && (
        <div className="section-card space-y-4 text-center">
          <div className="w-16 h-16 mx-auto rounded-full bg-emerald-500/20 flex items-center justify-center">
            <svg className="w-8 h-8 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-emerald-400">{t('boleto.success')}</h3>
          <p className="text-fog/60 text-sm">{t('boleto.successDetail')}</p>

          <div className="grid grid-cols-2 gap-3 text-sm">
            <div className="bg-ink/40 rounded-2xl p-3">
              <span className="text-fog/50 block">{t('boleto.valueBrl')}</span>
              <span className="text-fog font-medium">{fmt(status.amountBrl || quote.amountBrl, locale)}</span>
            </div>
            <div className="bg-ink/40 rounded-2xl p-3">
              <span className="text-fog/50 block">{t('boleto.paidSats')}</span>
              <span className="text-fog font-medium">{fmtSats(status.totalSats || quote.totalSats, locale)}</span>
            </div>
          </div>

          {status.protocolo && (
            <div className="text-xs text-fog/40">
              {t('boleto.protocolo')}: <span className="font-mono">{status.protocolo}</span>
            </div>
          )}

          <button className="btn-primary w-full" onClick={handleReset}>
            {t('boleto.payAnother')}
          </button>
        </div>
      )}

      {/* Step: Error */}
      {step === 'error' && (
        <div className="section-card space-y-4 text-center">
          <div className="w-16 h-16 mx-auto rounded-full bg-ember/20 flex items-center justify-center">
            <svg className="w-8 h-8 text-ember" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-ember">{t('boleto.errorTitle')}</h3>
          <p className="text-fog/60 text-sm">
            {error || t('boleto.errorGeneric')}
          </p>
          <button className="btn-primary w-full" onClick={handleReset}>
            {t('boleto.tryAgain')}
          </button>
        </div>
      )}

      {/* History */}
      {history.length > 0 && (
        <div className="section-card">
          <h3 className="text-lg font-semibold text-fog mb-3">{t('boleto.history')}</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-fog/50 border-b border-white/10">
                  <th className="text-left py-2 pr-4">{t('boleto.bank')}</th>
                  <th className="text-right py-2 pr-4">{t('boleto.valueBrl')}</th>
                  <th className="text-right py-2 pr-4">Sats</th>
                  <th className="text-center py-2">{t('boleto.statusLabel')}</th>
                </tr>
              </thead>
              <tbody>
                {history.map(h => (
                  <tr key={h.paymentHash} className="border-b border-white/5">
                    <td className="py-2 pr-4 text-fog">{h.bankName}</td>
                    <td className="py-2 pr-4 text-right text-fog">{fmt(h.amountBrl, locale)}</td>
                    <td className="py-2 pr-4 text-right text-fog/70">{fmtSats(h.totalSats, locale)}</td>
                    <td className="py-2 text-center">
                      <span className={
                        h.status === 'completed' ? 'text-emerald-400' :
                        h.status === 'failed' || h.status === 'expired' ? 'text-ember' :
                        'text-brass'
                      }>
                        {t(`boleto.status.${h.status}`, h.status)}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </section>
  )
}
