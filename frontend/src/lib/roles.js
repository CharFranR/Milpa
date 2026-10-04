export const ROLE_MAP = { 0: 'pending', 1: 'buyer', 2: 'producer', 3: 'admin' }

const FRONTEND_ROLES = ['pending', 'buyer', 'producer', 'admin']

export function normalizeRole(role) {
  if (typeof role === 'number') return ROLE_MAP[role] ?? null
  return FRONTEND_ROLES.includes(role) ? role : null
}
