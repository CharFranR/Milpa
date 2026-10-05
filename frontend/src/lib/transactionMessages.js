const FRIENDLY = {
  'invalid transaction transition': 'Esta acción no corresponde al estado actual de la transacción.',
  'participant has already confirmed': 'Ya habías confirmado esta etapa.',
  'transaction is in a terminal state': 'La transacción ya está cerrada.',
  'reason is required': 'Indica el motivo de la cancelación.',
  'invalid match status': 'El match ya no permite esta operación.',
  'invalid request status': 'La solicitud ya no permite esta operación.',
  'invalid offer status': 'La oferta ya no permite esta operación.',
}

export function friendlyTransactionError(err) {
  const raw = typeof err === 'string' ? err : err?.message || ''
  const known = Object.keys(FRIENDLY).find((key) => raw.includes(key))
  if (known) return FRIENDLY[known]
  if (typeof err?.status === 'number' && err.status >= 500) {
    return 'Servicio no disponible, intenta de nuevo'
  }
  return raw || 'No se pudo completar la acción.'
}
