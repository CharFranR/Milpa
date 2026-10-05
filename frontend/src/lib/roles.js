// El backend alineó el enum con el brief (commit 551f07d en develop):
// 0 pending · 1 agricultor · 2 comprador minorista · 3 mayorista detallista
// 4 mayorista corporativo · 5 admin · 6 auditor.
// El frontend trabaja con roles lógicos: productor, buyer (los tres tipos de
// comprador colapsan en uno), admin y auditor.
export const ROLE_MAP = {
  0: 'pending',
  1: 'producer',
  2: 'buyer',
  3: 'buyer',
  4: 'buyer',
  5: 'admin',
  6: 'auditor',
}

const FRONTEND_ROLES = ['pending', 'buyer', 'producer', 'admin', 'auditor']

export function normalizeRole(role) {
  if (typeof role === 'number') return ROLE_MAP[role] ?? null
  return FRONTEND_ROLES.includes(role) ? role : null
}
