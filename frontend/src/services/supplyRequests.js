import { request } from './http'

export const supplyRequests = {
  list: () => request('/supply-requests/'),

  create: (data) =>
    request('/supply-requests/', { method: 'POST', body: data }),

  getById: (id) => request(`/supply-requests/${id}`),

  available: () => request('/supply-requests/available'),

  update: (id, data) =>
    request(`/supply-requests/${id}`, { method: 'PATCH', body: data }),

  updateAmounts: (id, data) =>
    request(`/supply-requests/${id}/amounts`, { method: 'PATCH', body: data }),

  updateDeadlines: (id, data) =>
    request(`/supply-requests/${id}/deadlines`, { method: 'PATCH', body: data }),

  cancel: (id) =>
    request(`/supply-requests/${id}/cancel`, { method: 'POST' }),

  expire: (id) =>
    request(`/supply-requests/${id}/expire`, { method: 'POST' }),
}
