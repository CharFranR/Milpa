import { request } from './http'

export const inventory = {
  list: () => request('/inventory/'),

  upsert: (data) =>
    request('/inventory/', { method: 'PUT', body: data }),

  remove: (id) => request(`/inventory/${id}`, { method: 'DELETE' }),
}

export const recommendations = {
  availability: (supplierId, productName) =>
    request(
      `/recommendations/availability?supplier_id=${encodeURIComponent(supplierId)}&product_name=${encodeURIComponent(productName)}`,
    ),
}
