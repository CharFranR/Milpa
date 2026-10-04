import { useCallback, useEffect, useState } from 'react'
import { matches } from '../services/matches'

function useAsyncList(load, requestId, message) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(Boolean(requestId))
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    if (!requestId) return
    setLoading(true)
    setError('')
    load(requestId)
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || message))
      .finally(() => setLoading(false))
  }, [requestId, load, message])

  useEffect(() => {
    if (!requestId) {
      setLoading(false)
      return
    }
    reload()
  }, [requestId, reload])

  return { items, loading, error, reload }
}

export function usePrioritizedOffers(requestId) {
  return useAsyncList(
    matches.prioritized,
    requestId,
    'Error al cargar las ofertas recomendadas.',
  )
}

export function useRequestMatches(requestId) {
  return useAsyncList(matches.listByRequest, requestId, 'Error al cargar los matches.')
}
