import { request } from './http'

export const transactions = {
  getByMatch: (matchId) => request(`/transactions/matches/${matchId}`),

  listByRequest: (requestId) => request(`/transactions/requests/${requestId}`),

  confirmStart: (transactionId) =>
    request(`/transactions/${transactionId}/confirm-start`, { method: 'POST' }),

  confirmDelivery: (transactionId) =>
    request(`/transactions/${transactionId}/confirm-delivery`, { method: 'POST' }),

  cancel: (transactionId, reason) =>
    request(`/transactions/${transactionId}/cancel`, { method: 'POST', body: { reason } }),
}
