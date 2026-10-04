import { request } from './http'

export const auth = {
  login: (email, password) =>
    request('/auth/login', { method: 'POST', body: { email, password } }),

  register: (data) =>
    request('/auth/register', { method: 'POST', body: data }),
}
