import { useState, useEffect } from 'react'
import { users } from '../services/api'
import { setUser as setSessionUser } from '../lib/session'

export function useUserProfile(userId) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!userId) { setLoading(false); return }
    users.getById(userId)
      .then(setUser)
      .catch((err) => setError(err.message || 'Error cargando perfil'))
      .finally(() => setLoading(false))
  }, [userId])

  function updateUser(data) {
    return users.update(userId, data).then(() => {
      setUser((prev) => {
        const updated = { ...prev, ...data }
        setSessionUser(updated)
        return updated
      })
    })
  }

  return { user, loading, error, updateUser }
}
