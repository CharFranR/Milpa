import { request } from './http'

export const images = {
  upload: async (file) => {
    const body = new FormData()
    body.append('image', file)

    const data = await request('/images/', { method: 'POST', body })
    return { path: data?.path || '' }
  },
}
