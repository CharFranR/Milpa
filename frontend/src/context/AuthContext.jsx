import { createContext, startTransition, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { LOGOUT_EVENT, clearSession, getToken, getUser, setToken, setUser } from '../lib/session'
import { normalizeRole } from '../lib/roles'

const AuthContext = createContext(null)

function loadSessionUser() {
  const stored = getUser()
  if (!stored) return null
  const role = normalizeRole(stored.role)
  return role ? { ...stored, role } : null
}

function normalizeUserRole(nextUser) {
  if (!nextUser || typeof nextUser !== 'object' || nextUser.role === undefined) {
    return nextUser
  }
  const normalized = normalizeRole(nextUser.role)
  return normalized ? { ...nextUser, role: normalized } : nextUser
}

export function AuthProvider({ children }) {
  const [user, setAuthUser] = useState(loadSessionUser)
  const [token, setAuthToken] = useState(() => getToken())

  useEffect(() => {
    function handleLogout() {
      startTransition(() => {
        setAuthUser(null)
        setAuthToken(null)
      })
    }
    window.addEventListener(LOGOUT_EVENT, handleLogout)
    return () => window.removeEventListener(LOGOUT_EVENT, handleLogout)
  }, [])

  const login = useCallback((nextToken, nextUser) => {
    const normalizedUser = normalizeUserRole(nextUser)
    setToken(nextToken)
    setUser(normalizedUser)
    setAuthToken(nextToken)
    setAuthUser(normalizedUser)
  }, [])

  const logout = useCallback(() => {
    clearSession()
  }, [])

  const updateUser = useCallback((nextUser) => {
    const normalizedUser = normalizeUserRole(nextUser)
    setUser(normalizedUser)
    setAuthUser(normalizedUser)
  }, [])

  const value = useMemo(
    () => ({
      user,
      token,
      role: user ? normalizeRole(user.role) : null,
      isAuthenticated: !!token && !!user,
      login,
      logout,
      updateUser,
    }),
    [user, token, login, logout, updateUser],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth debe usarse dentro de <AuthProvider>')
  }
  return context
}
