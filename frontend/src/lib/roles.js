// El backend alineó el enum con el brief (commit 551f07d en develop):
// 0 pending · 1 agricultor · 2 comprador minorista · 3 mayorista detallista
// 4 mayorista corporativo · 5 admin · 6 auditor.
// El frontend trabaja con roles lógicos: producer, minorista, mayorista, admin y auditor.
export const ROLE_MAP = {
  0: 'pending',
  1: 'producer',
  2: 'minorista',
  3: 'mayorista',
  4: 'mayorista',
  5: 'admin',
  6: 'auditor',
}

export const ROLE_LABELS = {
  pending: 'Pendiente',
  producer: 'Agricultor',
  minorista: 'Comprador Minorista',
  mayorista: 'Comprador Mayorista',
  admin: 'Administrador',
  auditor: 'Auditor',
}

const FRONTEND_ROLES = ['pending', 'producer', 'minorista', 'mayorista', 'admin', 'auditor']

export function normalizeRole(role) {
  if (typeof role === 'number') return ROLE_MAP[role] ?? null
  return FRONTEND_ROLES.includes(role) ? role : null
}

export function isBuyer(role) {
  return role === 'minorista' || role === 'mayorista'
}