import { request } from './http'

const categoryBody = (data) => {
  const body = { ...data }
  const days = body.default_expiry_days
  if (days === undefined || days === null || days === '') {
    delete body.default_expiry_days
  } else {
    body.default_expiry_days = Number(days)
  }
  return body
}

export const admin = {
  listUsers: (params = {}) => {
    const searchParams = new URLSearchParams()
    if (params.page) searchParams.set('page', params.page)
    if (params.pageSize) searchParams.set('page_size', params.pageSize)
    const qs = searchParams.toString()
    return request(`/admin/users${qs ? '?' + qs : ''}`)
  },
  suspendUser: (id, action) => request(`/admin/users/${id}/suspend`, { method: 'PATCH', body: { action } }),
  setUserRole: (id, role) => request(`/admin/users/${id}/role`, { method: 'PATCH', body: { role } }),
  deleteOffering: (id) => request(`/admin/offerings/${id}`, { method: 'DELETE' }),
  listAuditLogs: (params = {}) => {
    const searchParams = new URLSearchParams()
    if (params.action) searchParams.set('action', params.action)
    if (params.actorId) searchParams.set('actor_id', params.actorId)
    if (params.targetType) searchParams.set('target_type', params.targetType)
    if (params.page) searchParams.set('page', params.page)
    if (params.pageSize) searchParams.set('page_size', params.pageSize)
    const qs = searchParams.toString()
    return request(`/admin/audit-logs${qs ? '?' + qs : ''}`)
  },
  createCategory: (data) => request('/admin/categories', { method: 'POST', body: categoryBody(data) }),
  updateCategory: (id, data) => request(`/admin/categories/${id}`, { method: 'PATCH', body: categoryBody(data) }),
  setCategoryStatus: (id, isActive) => request(`/admin/categories/${id}/status`, { method: 'PATCH', body: { is_active: isActive } }),
}