import { request } from './http'
import { FEATURED_COMPANY_IDS } from '../config/featuredCompanies'

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
