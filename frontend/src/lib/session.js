const TOKEN_KEY = 'milpa_token'
const USER_KEY = 'milpa_user'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || sessionStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
  sessionStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
  sessionStorage.removeItem(TOKEN_KEY)
}

export function getUser() {
  const data = localStorage.getItem(USER_KEY) || sessionStorage.getItem(USER_KEY)
  if (!data) return null
  try { return JSON.parse(data) } catch { return null }
}

export function setUser(user) {
  const data = JSON.stringify(user)
  localStorage.setItem(USER_KEY, data)
  sessionStorage.setItem(USER_KEY, data)
}

export function clearUser() {
  localStorage.removeItem(USER_KEY)
  sessionStorage.removeItem(USER_KEY)
}

export function hasRole(role) {
  const u = getUser()
  return u?.role === role
}

export function isAuthenticated() {
  return !!getToken()
}

export function clearSession() {
  clearToken()
  clearUser()
}

export function getSessionRole() {
  return getUser()?.role || null
}
