const base = ''

async function request(path: string, options?: RequestInit) {
  const res = await fetch(`${base}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options?.headers || {})
    }
  })
  if (!res.ok) {
    const text = await res.text()
    if (text) {
      try {
        const payload = JSON.parse(text)
        if (payload && typeof payload.error === 'string') {
          throw new Error(payload.error)
        }
      } catch {
        // fall through to raw text
      }
      throw new Error(text)
    }
    throw new Error('Request failed')
  }
  if (res.status === 204) return null
  return res.json()
}

export const getHealth = () => request('/api/health')
export const getSystem = () => request('/api/system')
export const getDisk = () => request('/api/disk')
export const getPostgres = () => request('/api/postgres')
export const getBitcoin = () => request('/api/bitcoin')
export const getLndStatus = () => request('/api/lnd/status')
export const getLndConfig = () => request('/api/lnd/config')

export const postBitcoinRemote = (payload: { rpcuser: string; rpcpass: string }) =>
  request('/api/wizard/bitcoin-remote', { method: 'POST', body: JSON.stringify(payload) })

export const createWalletSeed = (payload?: { seed_passphrase?: string; wallet_password?: string }) =>
  request('/api/wizard/lnd/create-wallet', { method: 'POST', body: JSON.stringify(payload ?? {}) })

export const initWallet = (payload: { wallet_password: string; seed_words: string[] }) =>
  request('/api/wizard/lnd/init-wallet', { method: 'POST', body: JSON.stringify(payload) })

export const unlockWallet = (payload: { wallet_password: string }) =>
  request('/api/wizard/lnd/unlock', { method: 'POST', body: JSON.stringify(payload) })

export const restartService = (payload: { service: string }) =>
  request('/api/actions/restart', { method: 'POST', body: JSON.stringify(payload) })

export const getLogs = (service: string, lines: number) =>
  request(`/api/logs?service=${service}&lines=${lines}`)

export const updateLndConfig = (payload: {
  alias: string
  min_channel_size_sat: number
  max_channel_size_sat: number
  apply_now: boolean
}) => request('/api/lnd/config', { method: 'POST', body: JSON.stringify(payload) })

export const updateLndRawConfig = (payload: { raw_user_conf: string; apply_now: boolean }) =>
  request('/api/lnd/config/raw', { method: 'POST', body: JSON.stringify(payload) })

export const getWalletSummary = () => request('/api/wallet/summary')
export const createInvoice = (payload: { amount_sat: number; memo: string }) =>
  request('/api/wallet/invoice', { method: 'POST', body: JSON.stringify(payload) })
export const payInvoice = (payload: { payment_request: string }) =>
  request('/api/wallet/pay', { method: 'POST', body: JSON.stringify(payload) })
