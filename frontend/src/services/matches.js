import { request } from './http'

export const matches = {
  like: (offerId) => request(`/matches/like/${offerId}`, { method: 'POST' }),

  pass: (offerId) => request(`/matches/pass/${offerId}`, { method: 'POST' }),

  listByRequest: (requestId) => request(`/matches/requests/${requestId}`),

  prioritized: (requestId) => request(`/matches/requests/${requestId}/prioritized`),

  getById: (matchId) => request(`/matches/${matchId}`),
}
