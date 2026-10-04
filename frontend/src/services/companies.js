import { request } from './http'

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
