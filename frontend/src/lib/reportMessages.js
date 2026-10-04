const FRIENDLY = {
  'reason must be between 10 and 500 characters': 'El motivo debe tener entre 10 y 500 caracteres.',
  'reason is required': 'Indica el motivo de la denuncia.',
  'target type must be': 'Tipo de destinatario no válido.',
  'a pending report already exists': 'Ya tienes una denuncia pendiente sobre este destinatario.',
  'you cannot report yourself': 'No puedes denunciarte a ti mismo.',
}

export function friendlyReportError(err) {
  const raw = typeof err === 'string' ? err : err?.message || ''
  const known = Object.keys(FRIENDLY).find((key) => raw.toLowerCase().includes(key))
  if (known) return FRIENDLY[known]
  return raw || 'No se pudo enviar la denuncia.'
}
