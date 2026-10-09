import { request } from './http'

export const reports = {
  create: (data) => request('/reports', { method: 'POST', body: data }),
  list: (params = {}) => {
    const searchParams = new URLSearchParams()
    if (params.status) searchParams.set('status', params.status)
    if (params.targetType) searchParams.set('target_type', params.targetType)
    if (params.page) searchParams.set('page', params.page)
    if (params.pageSize) searchParams.set('page_size', params.pageSize)
    const qs = searchParams.toString()
    return request(`/reports${qs ? '?' + qs : ''}`)
  },
  resolve: (id, action) => request(`/reports/${id}/action`, { method: 'PATCH', body: { action } }),
}