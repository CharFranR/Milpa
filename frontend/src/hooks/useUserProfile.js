import { useState, useEffect } from 'react'
import { users } from '../services/api'
import { useAuth } from '../context/AuthContext'

function mergeProfile(prev, data) {
  const next = { ...prev, ...data }
  if ('address' in data) {
    next.address_line = data.address
    delete next.address
  }
  return next
}

export function useUserProfile(userId) {
  const { user: sessionUser, updateUser: updateSessionUser } = useAuth()
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
      setUser((prev) => mergeProfile(prev, data))
      updateSessionUser(mergeProfile(sessionUser, data))
    })
  }

  return { user, loading, error, updateUser }
}
