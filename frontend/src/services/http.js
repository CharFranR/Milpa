import { getToken, clearSession } from '../lib/session'

export const BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

const FORBIDDEN_MESSAGE = 'No tienes permisos para realizar esta acción'
const UNAVAILABLE_MESSAGE = 'Servicio no disponible, intenta de nuevo'

export const RETURN_TO_KEY = 'milpa_return_to'

export function rememberReturnTo() {
  // Se guarda la ruta, no el hash: navigate() no acepta "#/algo".
  const path = window.location.hash.replace(/^#/, '')
  if (path && path !== '/login' && path !== '/register') {
    sessionStorage.setItem(RETURN_TO_KEY, path)
  }
}

export async function request(path, options = {}) {
  const url = `${BASE_URL}${path}`
  const token = getToken()

  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    ...options,
  }

  if (config.body && typeof config.body === 'object' && !(config.body instanceof FormData)) {
    config.body = JSON.stringify(config.body)
  }

  let res
  try {
    res = await fetch(url, config)
  } catch {
    throw { message: UNAVAILABLE_MESSAGE, status: 0, errors: {} }
  }

  const isAuthEndpoint = path.startsWith('/auth/')

  if (res.status === 401 && !isAuthEndpoint) {
    // El hash se escribe a mano (fuera del router), asi que el destino se
    // guarda aparte: si no, al volver a entrar el usuario cae en su panel y
    // pierde la pagina que estaba viendo.
    rememberReturnTo()
    clearSession()
    window.location.hash = '#/login'
    throw { message: 'Sesión expirada', status: 401, errors: {} }
  }

  const data = await res.json().catch(() => null)

  if (!res.ok) {
    if (res.status === 403 && String(data?.error || '').includes('suspended')) {
      clearSession()
      sessionStorage.setItem('milpa_notice', 'Tu cuenta está suspendida. Contacta al administrador.')
      window.location.hash = '#/login'
      throw { message: 'Tu cuenta está suspendida', status: 403, errors: {} }
    }
    throw buildError(res.status, data)
  }

  return data?.data !== undefined ? data.data : data
}

function buildError(status, data) {
  const errors = data?.errors || {}

  if (status === 403) {
    return { message: FORBIDDEN_MESSAGE, status, errors }
  }

  if (status >= 500) {
    return { message: UNAVAILABLE_MESSAGE, status, errors }
  }

  return { message: data?.error || data?.message || `Error ${status}`, status, errors }
}

export function chatUrl(conversationId) {
  const base = BASE_URL.startsWith('http')
    ? BASE_URL
    : `${window.location.origin}${BASE_URL}`

  const wsBase = base.replace(/^http/, 'ws')
  return `${wsBase}/ws/${conversationId}`
}

export default request
