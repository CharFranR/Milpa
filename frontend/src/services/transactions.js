import { request } from './http'

export const transactions = {
  getByMatch: (matchId) => request(`/transactions/matches/${matchId}`),
}
