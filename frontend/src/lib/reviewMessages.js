const FRIENDLY = {
  'rating must be between 1 and 5': 'La valoración debe estar entre 1 y 5 estrellas.',
  'transaction must be completed': 'Solo se puede reseñar una transacción completada.',
  'a review already exists': 'Ya enviaste una reseña para esta transacción.',
  'review target must be the other party': 'Solo puedes reseñar a la otra parte de la transacción.',
  'you cannot review yourself': 'No puedes reseñarte a ti mismo.',
  'target type must be company or user': 'Tipo de destinatario no válido.',
  'provide company_id or target_id': 'Faltan datos para guardar la reseña.',
  'target is required': 'Faltan datos para guardar la reseña.',
  'transaction is required': 'Faltan datos para guardar la reseña.',
}

export function friendlyReviewError(err) {
  const raw = typeof err === 'string' ? err : err?.message || ''
  const known = Object.keys(FRIENDLY).find((key) => raw.includes(key))
  if (known) return FRIENDLY[known]
  return raw || 'No se pudo guardar la reseña.'
}
