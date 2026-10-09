export const MEASUREMENT_LABELS = ['Kg', 'Lb', 'Tn']

export const REQUEST_STATUS = {
  0: { label: 'Abierta', tone: 'green' },
  1: { label: 'Cancelada', tone: 'gray' },
  2: { label: 'Completada', tone: 'brand' },
  3: { label: 'Expirada', tone: 'amber' },
}

export const OFFER_STATUS = {
  0: { label: 'Activa', tone: 'brand' },
  1: { label: 'Matcheada', tone: 'green' },
  2: { label: 'Rechazada', tone: 'gray' },
  3: { label: 'Retirada', tone: 'amber' },
}

export const MATCH_STATUS = {
  0: { label: 'Activo', tone: 'green' },
  1: { label: 'Cancelado', tone: 'gray' },
}

export const TRANSACTION_STATUS = {
  0: { label: 'Matcheada', tone: 'brand' },
  1: { label: 'En progreso', tone: 'amber' },
  2: { label: 'Completada', tone: 'green' },
  3: { label: 'Cancelada', tone: 'gray' },
}

export const EMPTY_TIME = '0001-01-01T00:00:00Z'

export function requestStatus(status) {
  return REQUEST_STATUS[status] || { label: 'Sin estado', tone: 'gray' }
}

export function offerStatus(status) {
  return OFFER_STATUS[status] || { label: 'Sin estado', tone: 'gray' }
}

export function matchStatus(status) {
  return MATCH_STATUS[status] || { label: 'Sin estado', tone: 'gray' }
}

export function transactionStatus(status) {
  return TRANSACTION_STATUS[status] || { label: 'Sin estado', tone: 'gray' }
}

export function measurementLabel(unit) {
  return MEASUREMENT_LABELS[unit] ?? '—'
}

export function isZeroTime(value) {
  return !value || String(value).startsWith('0001-')
}

export function formatDeadline(value) {
  if (isZeroTime(value)) return 'Sin fecha'
  return new Date(value).toLocaleDateString('es-NI', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export function formatDateTime(value) {
  if (isZeroTime(value)) return 'Sin fecha'
  return new Date(value).toLocaleString('es-NI', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function toInputValue(value) {
  if (isZeroTime(value)) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (n) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}
