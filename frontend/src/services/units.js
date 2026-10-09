import { request } from './http'

let listPromise = null

export const units = {
  getAll: () => {
    if (!listPromise) {
      listPromise = request('/units-of-measure').finally(() => {
        listPromise = null
      })
    }
    return listPromise
  },
}