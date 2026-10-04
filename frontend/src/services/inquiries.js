import { request } from './http'

export const inquiries = {
  create: (data) =>
    request('/inquiries', { method: 'POST', body: data }),

  getByUser: (userId) =>
    request(`/inquiries?user_id=${userId}`),

  getByCompany: (companyId) =>
    request(`/inquiries/company/${companyId}`),

  updateStatus: (id, status) =>
    request(`/inquiries/${id}`, { method: 'PATCH', body: { status } }),
}
