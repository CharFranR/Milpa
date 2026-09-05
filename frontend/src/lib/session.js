const SESSION_ROLE_KEY = 'milpa_session_role'

export function getSessionRole() {
  return localStorage.getItem(SESSION_ROLE_KEY)
}

export function setSessionRole(role) {
  localStorage.setItem(SESSION_ROLE_KEY, role)
}

export function clearSessionRole() {
  localStorage.removeItem(SESSION_ROLE_KEY)
}

export function hasBuyerSession() {
  return getSessionRole() === 'buyer'
}

export function hasProducerSession() {
  return getSessionRole() === 'producer'
}

export function hasAdminSession() {
  return getSessionRole() === 'admin'
}
