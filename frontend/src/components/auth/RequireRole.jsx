import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { HOME_BY_ROLE } from '../../lib/routes'

export default function RequireRole({ role, roles, children }) {
  const location = useLocation()
  const { role: sessionRole, isAuthenticated } = useAuth()

  const allowed = roles || [role]
  if (!allowed.includes(sessionRole)) {
    const to = isAuthenticated && sessionRole && HOME_BY_ROLE[sessionRole]
      ? HOME_BY_ROLE[sessionRole]
      : '/login'
    return <Navigate to={to} replace state={{ from: location.pathname }} />
  }

  return children
}
