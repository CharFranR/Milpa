import { useEffect, useState } from 'react'
import { offerings } from '../services/api'

const DEBOUNCE_MS = 300

export function useSearch({ term = '', categoryId = '', maxPrice = null, sort = 'relevance', page = 1, pageSize = 6 } = {}) {
  const [results, setResults] = useState([])
  const [totalHits, setTotalHits] = useState(0)
  const [totalPages, setTotalPages] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError('')

    const timer = setTimeout(() => {
      const params = { term, sort, page, page_size: pageSize }
      if (categoryId) params.category_id = categoryId
      if (maxPrice !== null && maxPrice !== undefined) params.price_max = maxPrice

      offerings.search(params)
        .then((data) => {
          if (cancelled) return
          setResults(Array.isArray(data?.results) ? data.results : [])
          setTotalHits(data?.total_hits || 0)
          setTotalPages(data?.total_pages || 0)
        })
        .catch((err) => {
          if (cancelled) return
          setError(err.message || 'Error buscando productos')
        })
        .finally(() => {
          if (!cancelled) setLoading(false)
        })
    }, DEBOUNCE_MS)

    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [term, categoryId, maxPrice, sort, page, pageSize])

  return { results, totalHits, totalPages, loading, error }
}
