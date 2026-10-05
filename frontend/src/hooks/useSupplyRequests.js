import { useCallback, useEffect, useState } from 'react'
import { supplyRequests } from '../services/supplyRequests'

export function useSupplyRequests() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    setLoading(true)
    setError('')
    supplyRequests
      .list()
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || 'Error al cargar tus solicitudes.'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    reload()
  }, [reload])

  return { items, loading, error, reload }
}

export function useSupplyRequest(id) {
  const [request, setRequest] = useState(null)
  const [loading, setLoading] = useState(Boolean(id))
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    if (!id) return
    setLoading(true)
    setError('')
    supplyRequests
      .getById(id)
      .then(setRequest)
      .catch((err) => setError(err.message || 'Error al cargar la solicitud.'))
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => {
    if (!id) {
      setLoading(false)
      return
    }
    reload()
  }, [id, reload])

  return { request, loading, error, reload, setRequest }
}
