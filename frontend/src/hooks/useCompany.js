import { useState, useEffect } from 'react'
import { companies } from '../services/api'

export function useCompany(ownerId) {
  const [company, setCompany] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  function fetchCompany() {
    if (!ownerId) { setLoading(false); return }
    setLoading(true)
    setError('')
    companies.getByOwner(ownerId)
      .then((data) => {
        setCompany(Array.isArray(data) ? data[0] || null : data)
      })
      .catch((err) => {
        if (err.status === 404) {
          setCompany(null)
        } else {
          setError(err.message || 'Error cargando empresa')
        }
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchCompany() }, [ownerId])

  function createCompany(data) {
    return companies.create(data).then((newCompany) => {
      setCompany(newCompany)
      return newCompany
    })
  }

  function updateCompany(id, data) {
    return companies.update(id, data).then(() => {
      setCompany((prev) => ({ ...prev, ...data }))
    })
  }

  return { company, loading, error, createCompany, updateCompany, refetch: fetchCompany }
}
