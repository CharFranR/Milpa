import { getToken, clearSession } from '../lib/session'
import { FEATURED_COMPANY_IDS } from '../config/featuredCompanies'

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

  const isAuthEndpoint = path.startsWith('/auth/')

  if (res.status === 401 && !isAuthEndpoint) {
    clearSession()
    window.location.hash = '#/login'
    throw { message: 'Sesión expirada', status: 401, errors: {} }
  }

  const data = await res.json().catch(() => null)

  if (!res.ok) {
    throw {
      message: data?.error || data?.message || `Error ${res.status}`,
      errors: data?.errors || {},
      status: res.status,
    }
  }

  const payload = data?.data !== undefined ? data.data : data
  return payload
}

export const auth = {
  login: (email, password) =>
    request('/auth/login', { method: 'POST', body: { email, password } }),

  register: (data) =>
    request('/auth/register', { method: 'POST', body: data }),
}

export const companies = {
  getByOwner: (ownerId) =>
    request(`/companies?owner_id=${ownerId}`),

  getById: (id) =>
    request(`/companies/${id}`),

  create: (data) =>
    request('/companies', { method: 'POST', body: data }),

  update: (id, data) =>
    request(`/companies/${id}`, { method: 'PATCH', body: data }),
}

export const offerings = {
  getByCompany: (companyId) =>
    request(`/offerings?company_id=${companyId}`),

  getById: (id) =>
    request(`/offerings/${id}`),

  create: (data) =>
    request('/offerings', { method: 'POST', body: data }),

  update: (id, data) =>
    request(`/offerings/${id}`, { method: 'PATCH', body: data }),

  getFeatured: async () => {
    if (FEATURED_COMPANY_IDS.length === 0) return []
    const results = await Promise.all(
      FEATURED_COMPANY_IDS.map((id) =>
        request(`/offerings?company_id=${id}`).catch(() => [])
      )
    )
    return results.flat()
  },
}

export const users = {
  getById: (id) =>
    request(`/users/${id}`),

  update: (id, data) =>
    request(`/users/${id}`, { method: 'PATCH', body: data }),
}

export const categories = {
  getAll: () =>
    request('/categories'),
}

export const inquiries = {
  create: (data) =>
    request('/inquiries', { method: 'POST', body: data }),

  getByUser: (userId) =>
    request(`/inquiries?user_id=${userId}`),

  updateStatus: (id, status) =>
    request(`/inquiries/${id}`, { method: 'PATCH', body: { status } }),
}

export default request
