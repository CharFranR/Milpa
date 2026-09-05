import { useState, useEffect } from 'react'
import { offerings } from '../services/api'

export function useOfferings(companyId) {
  const [offeringsList, setOfferingsList] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  function fetchOfferings() {
    if (!companyId) { setLoading(false); return }
    setLoading(true)
    setError('')
    offerings.getByCompany(companyId)
      .then((data) => setOfferingsList(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || 'Error cargando productos'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchOfferings() }, [companyId])

  function createOffering(data) {
    return offerings.create(data).then((newOffering) => {
      setOfferingsList((prev) => [...prev, newOffering])
      return newOffering
    })
  }

  function updateOffering(id, data) {
    return offerings.update(id, data).then(() => {
      setOfferingsList((prev) => prev.map((o) => o.id === id ? { ...o, ...data } : o))
    })
  }

  return { offeringsList, loading, error, createOffering, updateOffering, refetch: fetchOfferings }
}
