const TOKEN_KEY = 'milpa_token'
const USER_KEY = 'milpa_user'
const COMPANY_KEY = 'milpa_company_id'

export const LOGOUT_EVENT = 'milpa:logout'

function notify() {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new Event(LOGOUT_EVENT))
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function getUser() {
  const data = localStorage.getItem(USER_KEY)
  if (!data) return null
  try { return JSON.parse(data) } catch { return null }
}

export function setUser(user) {
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function clearUser() {
  localStorage.removeItem(USER_KEY)
}

export function getCompanyId() {
  return localStorage.getItem(COMPANY_KEY)
}

export function setCompanyId(id) {
  if (id) {
    localStorage.setItem(COMPANY_KEY, id)
  } else {
    localStorage.removeItem(COMPANY_KEY)
  }
}

export function clearCompanyId() {
  localStorage.removeItem(COMPANY_KEY)
}

export function hasRole(role) {
  return getUser()?.role === role
}

export function isAuthenticated() {
  return !!getToken()
}

export function clearSession() {
  clearToken()
  clearUser()
  clearCompanyId()
  notify()
}

export function getSessionRole() {
  return getUser()?.role || null
}
