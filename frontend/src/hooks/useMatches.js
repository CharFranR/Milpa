import { useCallback, useEffect, useState } from 'react'
import { matches } from '../services/matches'
import { transactions } from '../services/transactions'

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

export function useMatchIdsForOffers(items) {
  const matched = items.filter((item) => item.status === 1)
  const key = matched.map((item) => item.id).join('|')
  const [map, setMap] = useState({})

  useEffect(() => {
    const offers = items.filter((item) => item.status === 1)
    if (offers.length === 0) {
      setMap({})
      return undefined
    }
    let alive = true
    const requestIds = [...new Set(offers.map((offer) => offer.supply_request_id))]
    Promise.all(
      requestIds.map((requestId) =>
        transactions
          .listByRequest(requestId)
          .then((txs) =>
            Promise.all(txs.map((tx) => matches.getById(tx.match_id).catch(() => null))),
          )
          .catch(() => []),
      ),
    ).then((groups) => {
      const found = groups.flat().filter(Boolean)
      const pairs = offers.map((offer) => {
        const match = found.find((m) => m.supply_offer === offer.id)
        return [offer.id, match?.id || null]
      })
      if (alive) setMap(Object.fromEntries(pairs))
    })
    return () => {
      alive = false
    }
  }, [key, items])

  return map
}
