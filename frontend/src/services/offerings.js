import { request } from './http'
import { FEATURED_COMPANY_IDS } from '../config/featuredCompanies'

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
