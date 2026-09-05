import { getToken, clearSession } from '../lib/session'

const BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

async function request(path, options = {}) {
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

  const res = await fetch(url, config)

  if (res.status === 401) {
    clearSession()
    window.location.hash = '#/login'
    throw { message: 'Sesión expirada', status: 401, errors: {} }
  }

  const data = await res.json().catch(() => null)

  if (!res.ok) {
    throw {
      message: data?.message || `Error ${res.status}`,
      errors: data?.errors || {},
      status: res.status,
    }
  }

  return data
}

export const auth = {
  login: (email, password) =>
    request('/auth/login', { method: 'POST', body: { email, password } }),

  register: (data) =>
    request('/auth/register', { method: 'POST', body: data }),
}

export const inquiries = {
  create: (data) =>
    request('/inquiries', { method: 'POST', body: data }),

  getByProducer: () =>
    request('/inquiries'),

  updateStatus: (id, status) =>
    request(`/inquiries/${id}`, { method: 'PATCH', body: { status } }),
}

export default request
