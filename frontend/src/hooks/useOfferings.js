import { useState, useEffect } from 'react'
import { offerings } from '../services/api'

export function useOfferings(userId) {
  const [offeringsList, setOfferingsList] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  function fetchOfferings() {
    if (!userId) { setLoading(false); return }
    setLoading(true)
    setError('')
    offerings.getByUser(userId)
      .then((data) => setOfferingsList(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || 'Error cargando productos'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchOfferings() }, [userId])

  function createOffering(data) {
    return offerings.create(data).then((newOffering) => {
      setOfferingsList((prev) => [...prev, newOffering])
      return newOffering
    })
  }

  function mergeUpdated(id, updated, fallback) {
    setOfferingsList((prev) => prev.map((o) => o.id === id ? (updated?.id ? updated : { ...o, ...fallback }) : o))
    return updated
  }

  function updateOffering(id, data) {
    return offerings.update(id, data).then((updated) => mergeUpdated(id, updated, data))
  }

  function deactivateOffering(id) {
    return offerings.deactivate(id).then((updated) => mergeUpdated(id, updated, { is_active: false }))
  }

  function renewOffering(id, expiresAt) {
    return offerings.renew(id, expiresAt).then((updated) => mergeUpdated(id, updated, { expires_at: expiresAt, is_active: true }))
  }

  return {
    offeringsList,
    loading,
    error,
    createOffering,
    updateOffering,
    deactivateOffering,
    renewOffering,
    refetch: fetchOfferings,
  }
}
