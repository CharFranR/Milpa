import { request } from './http'

export const offerings = {
  getByUser: (userId) =>
    request(`/offerings?user_id=${userId}`),

  getByCompany: (companyId) =>
    request(`/offerings?company_id=${companyId}`),

  getById: (id) =>
    request(`/offerings/${id}`),

  create: (data) =>
    request('/offerings', { method: 'POST', body: data }),

  update: (id, data) =>
    request(`/offerings/${id}`, { method: 'PATCH', body: data }),

  deactivate: (id) =>
    request(`/offerings/${id}/status`, { method: 'PATCH' }),

  renew: (id, expiresAt) =>
    request(`/offerings/${id}/renew`, { method: 'PATCH', body: { expires_at: expiresAt } }),

  search: (params = {}) => {
    const qs = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value === undefined || value === null || value === '') return
      qs.set(key, String(value))
    })
    const query = qs.toString()
    return request(`/search${query ? `?${query}` : ''}`)
  },
}
