import { request } from './http'

export const supplyOffers = {
  listMine: () => request('/supply-offers/'),

  create: (data) =>
    request('/supply-offers/', { method: 'POST', body: data }),

  listByRequest: (requestId) =>
    request(`/supply-offers/requests/${requestId}`),

  getById: (id) => request(`/supply-offers/${id}`),

  update: (id, data) =>
    request(`/supply-offers/${id}`, { method: 'PATCH', body: data }),

  withdraw: (id) =>
    request(`/supply-offers/${id}/withdraw`, { method: 'POST' }),
}
