import { request } from './http'

export const reports = {
  create: (data) => request('/reports', { method: 'POST', body: data }),
}
