import { useCallback, useEffect, useState } from 'react'
import { supplyOffers } from '../services/supplyOffers'
import { supplyRequests } from '../services/supplyRequests'

export function useAvailableRequests() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    setLoading(true)
    setError('')
    supplyRequests
      .available()
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || 'Error al cargar las solicitudes disponibles.'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    reload()
  }, [reload])

  return { items, loading, error, reload }
}

export function useMyOffers() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    setLoading(true)
    setError('')
    supplyOffers
      .listMine()
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || 'Error al cargar tus ofertas.'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    reload()
  }, [reload])

  return { items, loading, error, reload }
}

export function useOffersByRequest(requestId) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(Boolean(requestId))
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    if (!requestId) return
    setLoading(true)
    setError('')
    supplyOffers
      .listByRequest(requestId)
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || 'Error al cargar las ofertas.'))
      .finally(() => setLoading(false))
  }, [requestId])

  useEffect(() => {
    if (!requestId) {
      setLoading(false)
      return
    }
    reload()
  }, [requestId, reload])

  return { items, loading, error, reload }
}
