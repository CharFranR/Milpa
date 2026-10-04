import { request } from './http'

export const users = {
  getById: (id) =>
    request(`/users/${id}`),

  update: (id, data) =>
    request(`/users/${id}`, { method: 'PATCH', body: data }),
}
