import { request } from './http'

export const reviews = {
  listByCompany: (companyId) => request(`/reviews?company_id=${companyId}`),

  listByUser: (userId) => request(`/reviews?user_id=${userId}`),

  average: (targetType, targetId) =>
    request(`/reviews/average?target_type=${targetType}&target_id=${targetId}`),

  create: (data) => request('/reviews', { method: 'POST', body: data }),
}
