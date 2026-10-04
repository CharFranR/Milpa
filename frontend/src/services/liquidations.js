import { request } from './http'

export const liquidations = {
  getOpen: () => request('/liquidations/open'),
  getBySupplier: (supplierId) => request(`/liquidations/?supplier_id=${supplierId}`),
  create: (body) => request('/liquidations/', { method: 'POST', body }),
  update: (id, body) => request(`/liquidations/${id}`, { method: 'PATCH', body }),
  remove: (id) => request(`/liquidations/${id}`, { method: 'DELETE' }),
}
