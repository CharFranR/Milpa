import { useState, useEffect } from 'react'
import { offerings } from '../services/api'

export function useFeaturedOfferings() {
  const [offeringsList, setOfferingsList] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    offerings.getFeatured()
      .then((data) => setOfferingsList(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || 'Error cargando productos'))
      .finally(() => setLoading(false))
  }, [])

  return { offeringsList, loading, error }
}
