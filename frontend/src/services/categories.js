import { request } from './http'

export const categories = {
  getAll: () =>
    request('/categories'),
}
