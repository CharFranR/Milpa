import { request } from './http'

let listPromise = null

export const categories = {
  getAll: () => {
    if (!listPromise) {
      // dedupe concurrent callers only: clear as soon as the request settles,
      // otherwise the first load would be served forever and a category created
      // later (e.g. from AdminConfig) would never show up without a reload.
      listPromise = request('/categories').finally(() => {
        listPromise = null
      })
    }
    return listPromise
  },
}
