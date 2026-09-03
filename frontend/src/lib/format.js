const priceFormatter = new Intl.NumberFormat('es-NI', {
  style: 'currency',
  currency: 'NIO',
  maximumFractionDigits: 0,
})

export function formatPrice(value) {
  return priceFormatter.format(value)
}

export function money(value) {
  return String(value).toLocaleString('es-NI')
}