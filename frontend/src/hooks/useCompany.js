import { useState, useEffect } from 'react'
import { companies } from '../services/api'
import { getCompanyId, setCompanyId } from '../lib/session'

export function useCompany(ownerId) {
  const [company, setCompany] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  function fetchCompany() {
    if (!ownerId) { setLoading(false); return }

    const savedId = getCompanyId()
    if (savedId) {
      setLoading(true)
      setError('')
      companies.getById(savedId)
        .then((data) => setCompany(data))
        .catch(() => {
          setCompany(null)
          setCompanyId(null)
        })
        .finally(() => setLoading(false))
      return
    }

    setLoading(true)
    setError('')
    companies.getByOwner(ownerId)
      .then((data) => {
        const found = Array.isArray(data) ? data[0] || null : data
        setCompany(found)
        if (found?.id) setCompanyId(found.id)
      })
      .catch((err) => {
        if (err.status === 404 || err.status === 500) {
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
      if (newCompany?.id) setCompanyId(newCompany.id)
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
