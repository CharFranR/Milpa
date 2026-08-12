const priceFormatter = new Intl.NumberFormat('es-CO', {
  style: 'currency',
  currency: 'COP',
  maximumFractionDigits: 0,
})

export function formatPrice(value) {
  return priceFormatter.format(value)
}

export function money(value) {
  return String(value).toLocaleString('es-CO')
}