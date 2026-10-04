export const LIQUIDATION_STATUS_LABELS = ['Abierta', 'Cerrada', 'Expirada', 'Asignada']

export function hasExpired(liq, now = new Date()) {
  if (!liq?.expires_at) return false
  const expiry = new Date(liq.expires_at).getTime()
  if (Number.isNaN(expiry)) return false
  return expiry <= now.getTime()
}

export function isOpen(liq, now = new Date()) {
  return liq?.status === 0 && !hasExpired(liq, now)
}

export function statusLabel(liq, now = new Date()) {
  if (liq?.status === 0 && hasExpired(liq, now)) return 'Expirada'
  return LIQUIDATION_STATUS_LABELS[liq?.status] || 'Abierta'
}

export function remainingLabel(expiresAt, now = new Date()) {
  if (!expiresAt) return ''
  const diff = new Date(expiresAt).getTime() - now.getTime()
  if (Number.isNaN(diff) || diff <= 0) return 'Expirada'

  const minutes = Math.floor(diff / 60000)
  const days = Math.floor(minutes / 1440)
  const hours = Math.floor((minutes % 1440) / 60)
  const mins = minutes % 60

  if (days > 0) return `${days} d ${hours} h`
  if (hours > 0) return `${hours} h ${mins} min`
  return `${mins} min`
}

export function toLocalInputValue(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''

  const pad = (part) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(
    date.getHours(),
  )}:${pad(date.getMinutes())}`
}
