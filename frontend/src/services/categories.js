import { request } from './http'

let listPromise = null

export const categories = {
  getAll: () => {
    if (!listPromise) {
      listPromise = request('/categories').catch((error) => {
        listPromise = null
        throw error
      })
    }
    return listPromise
  },
}
